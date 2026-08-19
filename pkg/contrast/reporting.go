package contrast

import (
	"fmt"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/toolrecord"
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

// SaveReportFile saves report data to the contrast reports directory
func SaveReportFile(utils piperutils.FileUtils, fileName, displayName string, data []byte) ([]piperutils.Path, error) {
	reportsDirectory := filepath.Join("./", "contrast")
	reportPath := filepath.Join(reportsDirectory, fileName)

	if err := utils.MkdirAll(reportsDirectory, 0777); err != nil {
		return nil, fmt.Errorf("failed to create contrast directory: %w", err)
	}

	if err := utils.FileWrite(reportPath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write %s file: %w", fileName, err)
	}

	log.Entry().Infof("Report saved to %s", reportPath)
	return []piperutils.Path{{Name: displayName, Target: reportPath}}, nil
}
