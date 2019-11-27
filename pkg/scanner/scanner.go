package scanner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Scanner base struct
type Scanner struct {
	ScannerURL string
}

// ImageScanResult contains details about all the found vulnerabilities
type ImageScanResult struct {
	Image       string            `json:"image"`
	ScanResult  string            `json:"scanResult"`
	Description string            `json:"description"`
	Targets     []TrivyScanTarget `json:"targets"`
}

type TrivyScanTarget struct {
	Target          string                     `json:"Target"`
	Vulnerabilities []VulnerabilityDescription `json:"Vulnerabilities"`
}

type VulnerabilityDescription struct {
	CVE              string   `json:"VulnerabilityID"`
	Package          string   `json:"PkgName"`
	InstalledVersion string   `json:"InstalledVersion"`
	FixedVersion     string   `json:"FixedVersion"`
	Title            string   `json:"Title"`
	Description      string   `json:"Description"`
	Severity         string   `json:"Severity"`
	References       []string `json:"References"`
}

// ImageScanResultSummary contains vulnerabilities summary
type ImageScanResultSummary struct {
	Image       string                 `json:"image"`
	ScanResult  string                 `json:"scanResult"`
	Description string                 `json:"description"`
	Counters    []VulnerabilityCounter `json:"counters"`
}

// VulnerabilityCounter represents amount of issues with specified severity
type VulnerabilityCounter struct {
	Severity string `json:"severity"`
	Count    int    `json:"count"`
}

func (s *ImageScanResultSummary) GetSeverity() string {
	switch s.ScanResult {
	case "Succeeded":
		if len(s.Counters) == 0 {
			return "Success"
		} else {
			for _, counter := range s.Counters {
				if counter.Severity == "CRITICAL" || counter.Severity == "HIGH" {
					return "Error"
				}
			}

			return "Warning"
		}
	default:
		return "NoData"
	}
}

func (s *ImageScanResultSummary) GetScansMessage() string {
	switch s.ScanResult {
	case "Succeeded":
		if len(s.Counters) == 0 {
			return "The image passed vulnerabilities check"
		} else {
			message := ""

			for _, counter := range s.Counters {
				message += fmt.Sprintf("%v %v, ", counter.Count, counter.Severity)
			}

			return fmt.Sprintf("Image has %v known vulnerabilities", message[:len(message)-2])
		}
	default:
		return "No vulnerabilities check data"
	}
}

// NewScanner returns a new scanner instance.
func NewScanner(url string) *Scanner {
	return &Scanner{ScannerURL: url}
}

// Scan all images
func (s *Scanner) Scan(images []string) {
	bytesRepresentation, err := json.Marshal(images)
	scansURL := fmt.Sprintf("%s/scan/images", s.ScannerURL)

	resp, err := http.Post(scansURL, "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		logrus.Errorf("Error requesting image scan %v", err)
		return
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Fatal(err)
	}
	bodyString := string(bodyBytes)
	logrus.Info(bodyString)
}

// Get returns detailed single image scan result
func (s *Scanner) Get(image string) (scanResult ImageScanResult, err error) {
	scanResultURL := fmt.Sprintf("%s/scan-results/trivy/%s", s.ScannerURL, url.QueryEscape(image))

	resp, err := http.Get(scanResultURL)
	if err != nil {
		logrus.Errorf("Error requesting image scan result %v", err)
		return
	}

	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&scanResult)

	return
}

// GetAll returns scan result summary for an array of images
func (s *Scanner) GetAll(images []string) (scanResults []ImageScanResultSummary, err error) {
	getURL := fmt.Sprintf("%s/scan-results/trivy?", s.ScannerURL)

	for _, image := range images {
		getURL = fmt.Sprintf("%s&images=%s", getURL, url.QueryEscape(image))
	}

	scanRequestsURL := getURL[:len(getURL)-1]
	resp, err := http.Get(scanRequestsURL)
	if err != nil {
		logrus.Errorf("Error requesting images scan results %v", err)
		return
	}

	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&scanResults)
	return
}
