package sonar

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const reportCodeCheckFileName = "codecheck.json"
const reportCombinedFileName = "sonarscan.json"
const reportHotSpotFileName = "hotspot.json"

// ReportCodeCheckData is representing the data of the step report JSON
type ReportCodeCheckData struct {
	ServerURL      string            `json:"serverUrl"`
	ProjectKey     string            `json:"projectKey"`
	TaskID         string            `json:"taskId"`
	ChangeID       string            `json:"changeID,omitempty"`
	BranchName     string            `json:"branchName,omitempty"`
	Organization   string            `json:"organization,omitempty"`
	NumberOfIssues *Issues           `json:"numberOfIssues"`
	ScanResults    []Severity        `json:"scanResults"`
	Coverage       *SonarCoverage    `json:"coverage,omitempty"`
	LinesOfCode    *SonarLinesOfCode `json:"linesOfCode,omitempty"`
}

// ReportCodeCheckData is representing the data of the step report JSON
type ReportHotSpotData struct {
	ServerURL        string            `json:"serverUrl"`
	ProjectKey       string            `json:"projectKey"`
	TaskID           string            `json:"taskId"`
	ChangeID         string            `json:"changeID,omitempty"`
	BranchName       string            `json:"branchName,omitempty"`
	Organization     string            `json:"organization,omitempty"`
	SecurityHotspots []SecurityHotspot `json:"securityHotspots"`
}

type ReportCombinedData struct {
	NumberOfIssues   *Issues           `json:"numberOfIssues"`
	ScanResults      []Severity        `json:"scanResults"`
	SecurityHotspots []SecurityHotspot `json:"securityHotspots"`
}

// HotSpot Security Issues
type SecurityHotspot struct {
	Priority string `json:"priority"`
	Hotspots int    `json:"hotspots"`
}

// Issues ...
type Issues struct {
	Blocker  int `json:"blocker"`
	Critical int `json:"critical"`
	Major    int `json:"major"`
	Minor    int `json:"minor"`
	Info     int `json:"info"`
}

type Severity struct {
	SeverityType string `json:"severity"`
	IssueType    string `json:"error_type,omitempty"`
	IssueCount   int    `json:"issues,omitempty"`
}

// WriteReport ...
func WriteCodeCheckReport(data ReportCodeCheckData, reportPath string, writeToFile func(f string, d []byte, p os.FileMode) error) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return writeToFile(filepath.Join(reportPath, reportCodeCheckFileName), jsonData, 0644)
}

func WriteHotSpotReport(data ReportHotSpotData, reportPath string, writeToFile func(f string, d []byte, p os.FileMode) error) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return writeToFile(filepath.Join(reportPath, reportHotSpotFileName), jsonData, 0644)
}

func WriteCombinedReport(data ReportCombinedData, reportPath string, writeToFile func(f string, d []byte, p os.FileMode) error) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return writeToFile(filepath.Join(reportPath, reportCombinedFileName), jsonData, 0644)
}
