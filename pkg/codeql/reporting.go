package codeql

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/toolrecord"
	"github.com/pkg/errors"
)

type CodeqlAudit struct {
	ToolName               string           `json:"toolName"`
	RepositoryUrl          string           `json:"repositoryUrl"`
	RepositoryReferenceUrl string           `json:"repositoryReferenceUrl"` //URL of PR or Branch where scan was performed
	CodeScanningLink       string           `json:"codeScanningLink"`
	QuerySuite             string           `json:"querySuite"`
	ScanResults            []CodeqlFindings `json:"findings"`
}

type CodeqlFindings struct {
	ClassificationName string `json:"classificationName"`
	Total              int    `json:"total"`
	Audited            int    `json:"audited"`
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

func CreateAndPersistToolRecord(utils piperutils.FileUtils, repoInfo *RepoInfo, modulePath string) (string, error) {
	toolRecord, err := createToolRecordCodeql(utils, repoInfo, modulePath)
	if err != nil {
		return "", err
	}

	toolRecordFileName, err := persistToolRecord(toolRecord)
	if err != nil {
		return "", err
	}

	return toolRecordFileName, nil
}

func createToolRecordCodeql(utils piperutils.FileUtils, repoInfo *RepoInfo, modulePath string) (*toolrecord.Toolrecord, error) {
	record := toolrecord.New(utils, modulePath, "codeql", repoInfo.ServerUrl)

	if repoInfo.ServerUrl == "" {
		return record, errors.New("Repository not set")
	}

	if repoInfo.CommitId == "" || repoInfo.CommitId == "NA" {
		return record, errors.New("CommitId not set")
	}

	if repoInfo.AnalyzedRef == "" {
		return record, errors.New("Analyzed Reference not set")
	}

	record.DisplayName = fmt.Sprintf("%s %s - %s %s", repoInfo.Owner, repoInfo.Repo, repoInfo.AnalyzedRef, repoInfo.CommitId)
	record.DisplayURL = repoInfo.ScanUrl

	err := record.AddKeyData("repository",
		fmt.Sprintf("%s/%s", repoInfo.Owner, repoInfo.Repo),
		fmt.Sprintf("%s %s", repoInfo.Owner, repoInfo.Repo),
		repoInfo.FullUrl)
	if err != nil {
		return record, err
	}

	err = record.AddKeyData("repositoryReference",
		repoInfo.AnalyzedRef,
		fmt.Sprintf("%s - %s", repoInfo.Repo, repoInfo.AnalyzedRef),
		repoInfo.FullRef)
	if err != nil {
		return record, err
	}

	err = record.AddKeyData("scanResult",
		fmt.Sprintf("%s/%s", repoInfo.AnalyzedRef, repoInfo.CommitId),
		fmt.Sprintf("%s %s - %s %s", repoInfo.Owner, repoInfo.Repo, repoInfo.AnalyzedRef, repoInfo.CommitId),
		repoInfo.ScanUrl)
	if err != nil {
		return record, err
	}

	return record, nil
}

func persistToolRecord(toolRecord *toolrecord.Toolrecord) (string, error) {
	err := toolRecord.Persist()
	if err != nil {
		return "", err
	}
	return toolRecord.GetFileName(), nil
}
