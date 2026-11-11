package contrast

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/toolrecord"
)

type ContrastAudit struct {
	ToolName                     string             `json:"toolName"`
	ApplicationUrl               string             `json:"applicationUrl"`
	ScanResults                  []ContrastFindings `json:"findings"`
	RouteTotalCount              *int               `json:"routeTotalCount,omitempty"`
	RouteDiscoveredCount         *int               `json:"routeDiscoveredCount,omitempty"`
	RouteExercisedCount          *int               `json:"routeExercisedCount,omitempty"`
	RouteCoveragePct             *float64           `json:"routeCoveragePct,omitempty"`
	RouteCoverageThreshold       int                `json:"routeCoverageThreshold"`
	AgentInactivityThresholdDays int                `json:"agentInactivityThresholdDays"`
	ConnectedServers             *int               `json:"connectedServers,omitempty"`
	Servers                      []ServerInfo       `json:"servers,omitempty"`
	SessionMetadataVersion       string             `json:"sessionMetadataVersion,omitempty"`
}

// ServerInfo is a summary of one server/agent instance connected to the application.
type ServerInfo struct {
	Name         string `json:"name"`
	Hostname     string `json:"hostname"`
	LastActivity string `json:"lastActivity"`
	Status       string `json:"status"`
}

// ToServerInfo converts a ServerResource API response into a ServerInfo report entry.
func (r ServerResource) ToServerInfo() ServerInfo {
	lastActivity := ""
	if r.LastActivity > 0 {
		lastActivity = time.UnixMilli(r.LastActivity).UTC().Format(time.RFC3339)
	}
	return ServerInfo{
		Name:         r.Name,
		Hostname:     r.Hostname,
		LastActivity: lastActivity,
		Status:       r.Status,
	}
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

func CreateAndPersistToolRecord(utils piperutils.FileUtils, appInfo *ApplicationInfo, modulePath string, exercisedRoutePercentage *float64) (string, error) {
	toolRecord, err := createToolRecordContrast(utils, appInfo, modulePath, exercisedRoutePercentage)
	if err != nil {
		return "", err
	}

	toolRecordFileName, err := persistToolRecord(toolRecord)
	if err != nil {
		return "", err
	}

	return toolRecordFileName, nil
}

func createToolRecordContrast(utils piperutils.FileUtils, appInfo *ApplicationInfo, modulePath string, exercisedRoutePercentage *float64) (*toolrecord.Toolrecord, error) {
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

	if exercisedRoutePercentage != nil {
		record.AddContext("exercisedRoutePercentage", *exercisedRoutePercentage)
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
