package codeql

import (
	"encoding/json"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/pkg/errors"
)

type CodeqlAudit struct {
	ToolName               string         `json:"toolName"`
	RepositoryUrl          string         `json:"repositoryUrl"`
	RepositoryReferenceUrl string         `json:"repositoryReferenceUrl"` //URL of PR or Branch where scan was performed
	CodeScanningLink       string         `json:"codeScanningLink"`
	ScanResults            CodeqlScanning `json:"scanResults"`
}

type CodeqlScanning struct {
	Total   int `json:"total"`
	Audited int `json:"audited"`
}

func WriteJSONReport(jsonReport CodeqlAudit, modulePath string) ([]piperutils.Path, error) {
	utils := piperutils.Files{}
	reportPaths := []piperutils.Path{}

	reportsDirectory := filepath.Join(modulePath, "codeql")
	jsonComplianceReportPath := filepath.Join(reportsDirectory, "piper_codeql_report.json")
	if err := utils.MkdirAll(reportsDirectory, 0777); err != nil {
		return reportPaths, errors.Wrapf(err, "failed to create report directory")
	}

	file, _ := json.Marshal(jsonReport)
	if err := utils.FileWrite(jsonComplianceReportPath, file, 0666); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return reportPaths, errors.Wrapf(err, "failed to write codeql json compliance report")
	}

	reportPaths = append(reportPaths, piperutils.Path{Name: "Codeql JSON Compliance Report", Target: jsonComplianceReportPath})

	return reportPaths, nil
}
