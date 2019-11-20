package scanner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
)

// Scanner base struct
type Scanner struct {
	Results    map[string]ImageScanResult
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

	var result map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&result)
	logrus.Print(result)
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
	logrus.Print(scanResult)

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
	logrus.Print(scanResults)

	return
}
