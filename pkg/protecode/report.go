package protecode

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/log"
)

type protecodeData struct {
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
func WriteReport(serverURL string, failOnSevereVulnerabilities bool, excludeCVEs string, scanReportFileName string, reportPath string, reportFileName string, result map[string]int, productID int, vulns []Vuln, writeToFile func(f string, d []byte, p os.FileMode) error) error {
	protecodeData := protecodeData{
		ServerURL:                   serverURL,
		FailOnSevereVulnerabilities: failOnSevereVulnerabilities,
		ExcludeCVEs:                 excludeCVEs,
		Target:                      scanReportFileName,
		Mandatory:                   true,
		ProductID:                   fmt.Sprintf("%v", productID),
		Count:                       fmt.Sprintf("%v", result["count"]),
		Cvss2GreaterOrEqualSeven:    fmt.Sprintf("%v", result["cvss2GreaterOrEqualSeven"]),
		Cvss3GreaterOrEqualSeven:    fmt.Sprintf("%v", result["cvss3GreaterOrEqualSeven"]),
		ExcludedVulnerabilities:     fmt.Sprintf("%v", result["excluded_vulnerabilities"]),
		TriagedVulnerabilities:      fmt.Sprintf("%v", result["triaged_vulnerabilities"]),
		HistoricalVulnerabilities:   fmt.Sprintf("%v", result["historical_vulnerabilities"]),
		Vulnerabilities:             vulns,
	}

	log.Entry().Infof("Protecode scan info, %v of which %v had a CVSS v2 score >= 7.0 and %v had a CVSS v3 score >= 7.0.\n %v vulnerabilities were excluded via configuration (%v) and %v vulnerabilities were triaged via the webUI.\nIn addition %v historical vulnerabilities were spotted. \n\n Vulnerabilities: %v",
		protecodeData.Count, protecodeData.Cvss2GreaterOrEqualSeven, protecodeData.Cvss3GreaterOrEqualSeven, protecodeData.ExcludedVulnerabilities, protecodeData.ExcludeCVEs, protecodeData.TriagedVulnerabilities, protecodeData.HistoricalVulnerabilities, protecodeData.Vulnerabilities)

	return writeJSON(reportPath, reportFileName, protecodeData, writeToFile)
}

func writeJSON(path, name string, data interface{}, writeToFile func(f string, d []byte, p os.FileMode) error) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return writeToFile(filepath.Join(path, name), jsonData, 0644)
}
