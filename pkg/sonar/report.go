package sonar

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const reportFileName = "sonarscan.json"

// ReportData is representing the data of the step report JSON
type ReportData struct {
	ServerURL      string            `json:"serverUrl"`
	ProjectKey     string            `json:"projectKey"`
	TaskID         string            `json:"taskId"`
	ChangeID       string            `json:"changeID,omitempty"`
	BranchName     string            `json:"branchName,omitempty"`
	Organization   string            `json:"organization,omitempty"`
	NumberOfIssues Issues            `json:"numberOfIssues"`
	Errors         []Severity        `json:"errors"`
	Coverage       *SonarCoverage    `json:"coverage,omitempty"`
	LinesOfCode    *SonarLinesOfCode `json:"linesOfCode,omitempty"`
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
	Count        int    `json:"issues"`
}

// WriteReport ...
func WriteReport(data ReportData, reportPath string, writeToFile func(f string, d []byte, p os.FileMode) error) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return writeToFile(filepath.Join(reportPath, reportFileName), jsonData, 0644)
}
