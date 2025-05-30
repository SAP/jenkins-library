package sonar

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const reportCodeCheckFileName = "sonarscan.json"
const reportHotSpotFileName = "hotspot.json"

// ReportData is representing the data of the step report JSON
type ReportData struct {
	ServerURL             string                 `json:"serverUrl"`
	ProjectKey            string                 `json:"projectKey"`
	TaskID                string                 `json:"taskId"`
	ChangeID              string                 `json:"changeID,omitempty"`
	BranchName            string                 `json:"branchName,omitempty"`
	Organization          string                 `json:"organization,omitempty"`
	NumberOfIssues        *Issues                `json:"numberOfIssues,omitempty"`
	Errors                []Severity             `json:"errors,omitempty"`
	Coverage              *SonarCoverage         `json:"coverage,omitempty"`
	LinesOfCode           *SonarLinesOfCode      `json:"linesOfCode,omitempty"`
	HotSpotSecurityIssues []HotSpotSecurityIssue `json:"hotSpotSecurityIssues,omitempty"`
}

// HotSpot Security Issues
type HotSpotSecurityIssue struct {
	IssueType string `json:"type"`
	Count     int    `json:"count"`
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
func WriteCodeCheckReport(data ReportData, reportPath string, writeToFile func(f string, d []byte, p os.FileMode) error) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return writeToFile(filepath.Join(reportPath, reportCodeCheckFileName), jsonData, 0644)
}

func WriteHotSpotReport(data ReportData, reportPath string, writeToFile func(f string, d []byte, p os.FileMode) error) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return writeToFile(filepath.Join(reportPath, reportHotSpotFileName), jsonData, 0644)
}
