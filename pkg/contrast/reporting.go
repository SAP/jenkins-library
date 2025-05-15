package contrast

import (
	"encoding/json"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/toolrecord"
	"github.com/pkg/errors"
)

type ContrastAudit struct {
	ToolName       string             `json:"toolName"`
	ApplicationUrl string             `json:"applicationUrl"`
	ScanResults    []ContrastFindings `json:"findings"`
}

type ContrastFindings struct {
	ClassificationName string `json:"classificationName"`
	Total              int    `json:"total"`
	Audited            int    `json:"audited"`
}

type ApplicationInfo struct {
	Url    string
	Id     string
	Name   string
	Server string
}

func WriteJSONReport(jsonReport ContrastAudit, modulePath string) ([]piperutils.Path, error) {
	utils := piperutils.Files{}
	reportPaths := []piperutils.Path{}

	reportsDirectory := filepath.Join(modulePath, "contrast")
	jsonComplianceReportData := filepath.Join(reportsDirectory, "piper_contrast_report.json")
	if err := utils.MkdirAll(reportsDirectory, 0777); err != nil {
		return reportPaths, errors.Wrapf(err, "failed to create report directory")
	}

	file, _ := json.Marshal(jsonReport)
	if err := utils.FileWrite(jsonComplianceReportData, file, 0666); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return reportPaths, errors.Wrapf(err, "failed to write contrast json compliance report")
	}

	reportPaths = append(reportPaths, piperutils.Path{Name: "Contrast JSON Compliance Report", Target: jsonComplianceReportData})
	return reportPaths, nil
}

func CreateAndPersistToolRecord(utils piperutils.FileUtils, appInfo *ApplicationInfo, modulePath string) (string, error) {
	toolRecord, err := createToolRecordContrast(utils, appInfo, modulePath)
	if err != nil {
		return "", err
	}

	toolRecordFileName, err := persistToolRecord(toolRecord)
	if err != nil {
		return "", err
	}

	return toolRecordFileName, nil
}

func createToolRecordContrast(utils piperutils.FileUtils, appInfo *ApplicationInfo, modulePath string) (*toolrecord.Toolrecord, error) {
	record := toolrecord.New(utils, modulePath, "contrast", appInfo.Server)

	record.DisplayName = appInfo.Name
	record.DisplayURL = appInfo.Url

	err := record.AddKeyData("application",
		appInfo.Id,
		appInfo.Name,
		appInfo.Url)
	if err != nil {
		return record, err
	}

	return record, nil
}

func persistToolRecord(toolrecord *toolrecord.Toolrecord) (string, error) {
	err := toolrecord.Persist()
	if err != nil {
		return "", err
	}
	return toolrecord.GetFileName(), nil
}
