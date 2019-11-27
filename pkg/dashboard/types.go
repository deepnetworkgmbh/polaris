package dashboard

import (
	"github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/validator"
	"html/template"
	"strings"
)

// templateData is passed to the dashboard HTML template
type templateData struct {
	BasePath  string
	Config    config.Configuration
	AuditData validator.AuditData
	JSON      template.JS
}

// scanTemplateData is passed to the image-scan-details HTML template
type scanTemplateData struct {
	BasePath string
	Config   config.Configuration
	ImageTag string
	ScanResult string
	Description string
	UsedIn []imageUsage
	ScanTargets []imageScanTarget
}

type imageUsage struct {

}

type imageScanTarget struct {
	Name string
	VulnerabilitiesGroups []vulnerabilitiesGroup
}

type vulnerabilitiesGroup struct {
	Severity string
	Count int
	CVEs []cveDetails
}

type cveDetails struct {
	Id               string
	PackageName      string
	InstalledVersion string
	FixedVersion     string
	Title            string
	Description      string
	References       []string
}

type bySeverity []string

func (s bySeverity) Len() int {
	return len(s)
}
func (s bySeverity) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s bySeverity) Less(i, j int) bool {
	return severityWeigh(s[i]) < severityWeigh(s[j])
}

func severityWeigh(s string) int {
	switch strings.ToUpper(s) {
	case "CRITICAL":
		return 0
	case "HIGH":
		return 20
	case "MEDIUM":
		return 40
	case "LOW":
		return 60
	case "UNKNOWN":
		return 80
	default:
		return 1000
	}
}