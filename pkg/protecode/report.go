package protecode

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/log"
)

//ReportData is representing the data of the step report JSON
type ReportData struct {
	Target                      string `json:"target,omitempty"`
	Mandatory                   bool   `json:"mandatory,omitempty"`
	ProductID                   string `json:"productID,omitempty"`
	ServerURL                   string `json:"serverUrl,omitempty"`
	FailOnSevereVulnerabilities bool   `json:"failOnSevereVulnerabilities,omitempty"`
	ExcludeCVEs                 string `json:"excludeCVEs,omitempty"`
	Count                       string `json:"count,omitempty"`
	Cvss2GreaterOrEqualSeven    string `json:"cvss2GreaterOrEqualSeven,omitempty"`
	Cvss3GreaterOrEqualSeven    string `json:"cvss3GreaterOrEqualSeven,omitempty"`
	ExcludedVulnerabilities     string `json:"excludedVulnerabilities,omitempty"`
	TriagedVulnerabilities      string `json:"triagedVulnerabilities,omitempty"`
	HistoricalVulnerabilities   string `json:"historicalVulnerabilities,omitempty"`
	Vulnerabilities             []Vuln `json:"Vulnerabilities,omitempty"`
}

// WriteReport ...
func WriteReport(data ReportData, reportPath string, reportFileName string, result map[string]int, writeToFile func(f string, d []byte, p os.FileMode) error) error {
	data.Mandatory = true
	data.Count = fmt.Sprintf("%v", result["count"])
	data.Cvss2GreaterOrEqualSeven = fmt.Sprintf("%v", result["cvss2GreaterOrEqualSeven"])
	data.Cvss3GreaterOrEqualSeven = fmt.Sprintf("%v", result["cvss3GreaterOrEqualSeven"])
	data.ExcludedVulnerabilities = fmt.Sprintf("%v", result["excluded_vulnerabilities"])
	data.TriagedVulnerabilities = fmt.Sprintf("%v", result["triaged_vulnerabilities"])
	data.HistoricalVulnerabilities = fmt.Sprintf("%v", result["historical_vulnerabilities"])

	log.Entry().Infof("Protecode scan info, %v of which %v had a CVSS v2 score >= 7.0 and %v had a CVSS v3 score >= 7.0.\n %v vulnerabilities were excluded via configuration (%v) and %v vulnerabilities were triaged via the webUI.\nIn addition %v historical vulnerabilities were spotted. \n\n Vulnerabilities: %v",
		data.Count, data.Cvss2GreaterOrEqualSeven, data.Cvss3GreaterOrEqualSeven,
		data.ExcludedVulnerabilities, data.ExcludeCVEs, data.TriagedVulnerabilities,
		data.HistoricalVulnerabilities, data.Vulnerabilities)
	return writeJSON(reportPath, reportFileName, data, writeToFile)
}

func writeJSON(path, name string, data interface{}, writeToFile func(f string, d []byte, p os.FileMode) error) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return writeToFile(filepath.Join(path, name), jsonData, 0644)
}
