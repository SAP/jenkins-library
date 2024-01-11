package codeql

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

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

type RepoInfo struct {
	ServerUrl string
	Repo      string
	CommitId  string
	Ref       string
	Owner     string
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

func BuildRepoReference(repository, analyzedRef string) (string, error) {
	ref := strings.Split(analyzedRef, "/")
	if len(ref) < 3 {
		return "", errors.New(fmt.Sprintf("Wrong analyzedRef format: %s", analyzedRef))
	}
	if strings.Contains(analyzedRef, "pull") {
		if len(ref) < 4 {
			return "", errors.New(fmt.Sprintf("Wrong analyzedRef format: %s", analyzedRef))
		}
		return fmt.Sprintf("%s/pull/%s", repository, ref[2]), nil
	}
	return fmt.Sprintf("%s/tree/%s", repository, ref[2]), nil
}

func CreateAndPersistToolRecord(utils piperutils.FileUtils, repoInfo RepoInfo, repoReference, repoUrl, modulePath string) (string, error) {
	toolRecord, err := createToolRecordCodeql(utils, repoInfo, repoReference, repoUrl, modulePath)
	if err != nil {
		return "", err
	}

	toolRecordFileName, err := persistToolRecord(toolRecord)
	if err != nil {
		return "", err
	}

	return toolRecordFileName, nil
}

func createToolRecordCodeql(utils piperutils.FileUtils, repoInfo RepoInfo, repoUrl, repoReference, modulePath string) (*toolrecord.Toolrecord, error) {
	record := toolrecord.New(utils, modulePath, "codeql", repoInfo.ServerUrl)

	if repoInfo.ServerUrl == "" {
		return record, errors.New("Repository not set")
	}

	if repoInfo.CommitId == "" || repoInfo.CommitId == "NA" {
		return record, errors.New("CommitId not set")
	}

	if repoInfo.Ref == "" {
		return record, errors.New("Analyzed Reference not set")
	}

	record.DisplayName = fmt.Sprintf("%s %s - %s %s", repoInfo.Owner, repoInfo.Repo, repoInfo.Ref, repoInfo.CommitId)
	record.DisplayURL = fmt.Sprintf("%s/security/code-scanning?query=is:open+ref:%s", repoUrl, repoInfo.Ref)

	err := record.AddKeyData("repository",
		fmt.Sprintf("%s/%s", repoInfo.Owner, repoInfo.Repo),
		fmt.Sprintf("%s %s", repoInfo.Owner, repoInfo.Repo),
		repoUrl)
	if err != nil {
		return record, err
	}

	err = record.AddKeyData("repositoryReference",
		repoInfo.Ref,
		fmt.Sprintf("%s - %s", repoInfo.Repo, repoInfo.Ref),
		repoReference)
	if err != nil {
		return record, err
	}

	err = record.AddKeyData("scanResult",
		fmt.Sprintf("%s/%s", repoInfo.Ref, repoInfo.CommitId),
		fmt.Sprintf("%s %s - %s %s", repoInfo.Owner, repoInfo.Repo, repoInfo.Ref, repoInfo.CommitId),
		fmt.Sprintf("%s/security/code-scanning?query=is:open+ref:%s", repoUrl, repoInfo.Ref))
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
