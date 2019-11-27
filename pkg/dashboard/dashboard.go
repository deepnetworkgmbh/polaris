// Copyright 2019 FairwindsOps Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dashboard

import (
	"bytes"
	"encoding/json"
	"github.com/fairwindsops/polaris/pkg/scanner"
	"html/template"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"

	"github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/kube"
	"github.com/fairwindsops/polaris/pkg/validator"
	packr "github.com/gobuffalo/packr/v2"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gitlab.com/golang-commonmark/markdown"
)

const (
	// MainTemplateName is the main template
	MainTemplateName = "main.gohtml"
	// HeadTemplateName contains styles and meta info
	HeadTemplateName = "head.gohtml"
	// NavbarTemplateName contains the navbar
	NavbarTemplateName = "navbar.gohtml"
	// PreambleTemplateName contains an empty preamble that can be overridden
	PreambleTemplateName = "preamble.gohtml"
	// DashboardTemplateName contains the content of the dashboard
	DashboardTemplateName = "dashboard.gohtml"
	// FooterTemplateName contains the footer
	FooterTemplateName = "footer.gohtml"
	// CheckDetailsTemplateName is a page for rendering details about a given check
	CheckDetailsTemplateName = "check-details.gohtml"
	// ImageScanDetailsTemplateName is a page for rendering a single image scan details
	ImageScanDetailsTemplateName = "image-scan-details.gohtml"
)

var (
	templateBox = (*packr.Box)(nil)
	assetBox    = (*packr.Box)(nil)
	markdownBox = (*packr.Box)(nil)
)

// GetAssetBox returns a binary-friendly set of assets packaged from disk
func GetAssetBox() *packr.Box {
	if assetBox == (*packr.Box)(nil) {
		assetBox = packr.New("Assets", "assets")
	}
	return assetBox
}

// GetTemplateBox returns a binary-friendly set of templates for rendering the dash
func GetTemplateBox() *packr.Box {
	if templateBox == (*packr.Box)(nil) {
		templateBox = packr.New("Templates", "templates")
	}
	return templateBox
}

// GetMarkdownBox returns a binary-friendly set of markdown files with error details
func GetMarkdownBox() *packr.Box {
	if markdownBox == (*packr.Box)(nil) {
		markdownBox = packr.New("Markdown", "../../docs/check-documentation")
	}
	return markdownBox
}

// GetBaseTemplate puts together the dashboard template. Individual pieces can be overridden before rendering.
func GetBaseTemplate(name string) (*template.Template, error) {
	tmpl := template.New(name).Funcs(template.FuncMap{
		"getWarningWidth":         getWarningWidth,
		"getSuccessWidth":         getSuccessWidth,
		"getClusterWeatherIcon":   getClusterWeatherIcon,
		"getScanWeatherIcon":      getScanWeatherIcon,
		"getWeatherText":          getWeatherText,
		"getClusterGrade":         getClusterGrade,
		"getScansGrade":           getScansGrade,
		"getIcon":                 getIcon,
		"getCategoryLink":         getCategoryLink,
		"getHelpLink":             getHelpLink,
		"getCategoryInfo":         getCategoryInfo,
		"getAllControllerResults": getAllControllerResults,
	})

	templateFileNames := []string{
		DashboardTemplateName,
		HeadTemplateName,
		NavbarTemplateName,
		PreambleTemplateName,
		FooterTemplateName,
		MainTemplateName,
	}
	return parseTemplateFiles(tmpl, templateFileNames)
}

func parseTemplateFiles(tmpl *template.Template, templateFileNames []string) (*template.Template, error) {
	templateBox := GetTemplateBox()
	for _, fname := range templateFileNames {
		templateFile, err := templateBox.Find(fname)
		if err != nil {
			return nil, err
		}

		tmpl, err = tmpl.Parse(string(templateFile))
		if err != nil {
			return nil, err
		}
	}
	return tmpl, nil
}

func writeTemplate(tmpl *template.Template, data *templateData, w http.ResponseWriter) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	buf.WriteTo(w)
}

func getConfigForQuery(base config.Configuration, query url.Values) config.Configuration {
	c := base
	exemptions := query.Get("disallowExemptions")
	if exemptions == "false" {
		c.DisallowExemptions = false
	}
	if exemptions == "true" {
		c.DisallowExemptions = true
	}
	return c
}

// GetRouter returns a mux router serving all routes necessary for the dashboard
func GetRouter(c config.Configuration, auditPath string, port int, basePath string, auditData *validator.AuditData) *mux.Router {
	router := mux.NewRouter().PathPrefix(basePath).Subrouter()
	kubeScanner := scanner.NewScanner(c.Images.ScannerUrl)
	fileServer := http.FileServer(GetAssetBox())

	router.PathPrefix("/static/").Handler(http.StripPrefix(path.Join(basePath, "/static/"), fileServer))

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	router.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		favicon, err := GetAssetBox().Find("favicon-32x32.png")
		if err != nil {
			logrus.Errorf("Error getting favicon: %v", err)
			http.Error(w, "Error getting favicon", http.StatusInternalServerError)
			return
		}
		w.Write(favicon)
	})

	router.HandleFunc("/results.json", func(w http.ResponseWriter, r *http.Request) {
		adjustedConf := getConfigForQuery(c, r.URL.Query())
		if auditData == nil {
			k, err := kube.CreateResourceProvider(auditPath)
			if err != nil {
				logrus.Errorf("Error fetching Kubernetes resources %v", err)
				http.Error(w, "Error fetching Kubernetes resources", http.StatusInternalServerError)
				return
			}

			k.FilterByNamespace(adjustedConf.NamespacesToScan...)
			auditDataObj, err := validator.RunAudit(adjustedConf, k)
			if err != nil {
				http.Error(w, "Error Fetching Deployments", http.StatusInternalServerError)
				return
			}
			auditData = &auditDataObj
		}

		JSONHandler(w, r, auditData)
	})

	router.HandleFunc("/details/{category}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		category := vars["category"]
		category = strings.Replace(category, ".md", "", -1)
		DetailsHandler(w, r, category, basePath)
	})

	router.HandleFunc("/image/{imageTag:.*}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		imageTag := vars["imageTag"]

		decodedValue, err := url.QueryUnescape(imageTag)
		if err != nil {
			logrus.Error(err, "Failed to unescape", imageTag)
			return
		}

		scanResult, err := kubeScanner.Get(decodedValue)
		if err != nil {
			logrus.Error(err, "Failed to get image scan details", imageTag)
			return
		}

		imageScanDetailsHandler(w, r, c, basePath, &scanResult)
	})

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && r.URL.Path != basePath {
			http.NotFound(w, r)
			return
		}
		adjustedConf := getConfigForQuery(c, r.URL.Query())

		if auditData == nil {
			k, err := kube.CreateResourceProvider(auditPath)
			if err != nil {
				logrus.Errorf("Error fetching Kubernetes resources %v", err)
				http.Error(w, "Error fetching Kubernetes resources", http.StatusInternalServerError)
				return
			}

			k.FilterByNamespace(adjustedConf.NamespacesToScan...)
			auditData, err := validator.RunAudit(adjustedConf, k)
			if err != nil {
				logrus.Errorf("Error getting audit data: %v", err)
				http.Error(w, "Error running audit", 500)
				return
			}
			MainHandler(w, r, adjustedConf, auditData, basePath)
		} else {
			MainHandler(w, r, adjustedConf, *auditData, basePath)
		}

	})
	return router
}

func imageScanDetailsHandler(w http.ResponseWriter, r *http.Request, c config.Configuration, basePath string, scan *scanner.ImageScanResult) {
	templateFileNames := []string{
		HeadTemplateName,
		NavbarTemplateName,
		ImageScanDetailsTemplateName,
		FooterTemplateName,
	}
	tmpl := template.New("image-scan-details")
	tmpl, err := parseTemplateFiles(tmpl, templateFileNames)
	if err != nil {
		logrus.Printf("Error getting template data %v", err)
		http.Error(w, "Error getting template data", 500)
		return
	}

	data := scanTemplateData{
		BasePath:   basePath,
		Config:     c,
		ImageTag:   scan.Image,
		ScanResult: scan.ScanResult,
		Description:scan.Description,
		UsedIn: []imageUsage{},
		ScanTargets: []imageScanTarget{},
	}

	for _, target := range scan.Targets{
		targetModel := imageScanTarget{
			Name:target.Target,
			VulnerabilitiesGroups: []vulnerabilitiesGroup{},
		}

		cveDict := make(map[string][]cveDetails)

		for _, cve := range target.Vulnerabilities {
			cveDict[cve.Severity] = append(cveDict[cve.Severity], cveDetails{
				Id:               cve.CVE,
				PackageName:      cve.Package,
				InstalledVersion: cve.InstalledVersion,
				FixedVersion:     cve.FixedVersion,
				Title:            cve.Title,
				Description:      cve.Description,
				References:       cve.References,
			})
		}

		keys := make([]string, 0, len(cveDict))
		for k := range cveDict {
			keys = append(keys, k)
		}
		sort.Sort(bySeverity(keys))

		for _, k := range keys {
			targetModel.VulnerabilitiesGroups = append(targetModel.VulnerabilitiesGroups, vulnerabilitiesGroup{
				Severity:k,
				Count: len(cveDict[k]),
				CVEs:cveDict[k],
			})
		}

		data.ScanTargets = append(data.ScanTargets, targetModel)
	}

	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	buf.WriteTo(w)
}

// MainHandler gets template data and renders the dashboard with it.
func MainHandler(w http.ResponseWriter, r *http.Request, c config.Configuration, auditData validator.AuditData, basePath string) {
	jsonData, err := json.Marshal(auditData)

	if err != nil {
		http.Error(w, "Error serializing audit data", 500)
		return
	}

	data := templateData{
		BasePath:  basePath,
		AuditData: auditData,
		JSON:      template.JS(jsonData),
		Config:    c,
	}
	tmpl, err := GetBaseTemplate("main")
	if err != nil {
		logrus.Printf("Error getting template data %v", err)
		http.Error(w, "Error getting template data", 500)
		return
	}
	writeTemplate(tmpl, &data, w)
}

// JSONHandler gets template data and renders json with it.
func JSONHandler(w http.ResponseWriter, r *http.Request, auditData interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(auditData)
}

// DetailsHandler returns details for a given error type
func DetailsHandler(w http.ResponseWriter, r *http.Request, category string, basePath string) {
	box := GetMarkdownBox()
	contents, err := box.Find(category + ".md")
	if err != nil {
		http.Error(w, "Error details not found for category "+category, http.StatusNotFound)
		return
	}
	md := markdown.New(markdown.XHTMLOutput(true))
	detailsHTML := "{{ define \"details\" }}" + md.RenderToString(contents) + "{{ end }}"

	templateFileNames := []string{
		HeadTemplateName,
		NavbarTemplateName,
		CheckDetailsTemplateName,
		FooterTemplateName,
	}
	tmpl := template.New("check-details")
	tmpl, err = parseTemplateFiles(tmpl, templateFileNames)
	if err != nil {
		logrus.Printf("Error getting template data %v", err)
		http.Error(w, "Error getting template data", 500)
		return
	}
	tmpl.Parse(detailsHTML)
	data := templateData{
		BasePath: basePath,
	}
	writeTemplate(tmpl, &data, w)
}
