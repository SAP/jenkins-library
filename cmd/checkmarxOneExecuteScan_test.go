package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"testing"

	"github.com/stretchr/testify/assert"

	checkmarxOne "github.com/SAP/jenkins-library/pkg/checkmarxone"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/pkg/errors"
)

type checkmarxOneSystemMock struct {
	response interface{}
}

func (sys *checkmarxOneSystemMock) DownloadReport(reportID string) ([]byte, error) {
	return nil, nil
}

func (sys *checkmarxOneSystemMock) GetReportStatus(reportID string) (checkmarxOne.ReportStatus, error) {
	return checkmarxOne.ReportStatus{}, nil
}

func (sys *checkmarxOneSystemMock) RequestNewReport(scanID, projectID, branch, reportType string) (string, error) {
	return "", nil
}

func (sys *checkmarxOneSystemMock) CreateApplication(appname string) (checkmarxOne.Application, error) {
	return checkmarxOne.Application{}, nil
}

func (sys *checkmarxOneSystemMock) GetApplicationByName(appname string) (checkmarxOne.Application, error) {
	return checkmarxOne.Application{}, nil
}

func (sys *checkmarxOneSystemMock) GetApplicationByID(appname string) (checkmarxOne.Application, error) {
	return checkmarxOne.Application{}, nil
}

func (sys *checkmarxOneSystemMock) UpdateApplication(app *checkmarxOne.Application) error {
	return nil
}

func (sys *checkmarxOneSystemMock) GetScan(scanID string) (checkmarxOne.Scan, error) {
	return checkmarxOne.Scan{}, nil
}

func (sys *checkmarxOneSystemMock) GetScanMetadata(scanID string) (checkmarxOne.ScanMetadata, error) {
	return checkmarxOne.ScanMetadata{}, nil
}

func (sys *checkmarxOneSystemMock) GetScanResults(scanID string, limit uint64) ([]checkmarxOne.ScanResult, error) {
	return []checkmarxOne.ScanResult{}, nil
}

func (sys *checkmarxOneSystemMock) GetScanSummary(scanID string) (checkmarxOne.ScanSummary, error) {
	return checkmarxOne.ScanSummary{}, nil
}

func (sys *checkmarxOneSystemMock) GetResultsPredicates(SimilarityID int64, ProjectID string) ([]checkmarxOne.ResultsPredicates, error) {
	return []checkmarxOne.ResultsPredicates{}, nil
}

func (sys *checkmarxOneSystemMock) GetScanWorkflow(scanID string) ([]checkmarxOne.WorkflowLog, error) {
	return []checkmarxOne.WorkflowLog{}, nil
}

func (sys *checkmarxOneSystemMock) GetLastScans(projectID, branch string, limit int) ([]checkmarxOne.Scan, error) {
	return []checkmarxOne.Scan{}, nil
}

func (sys *checkmarxOneSystemMock) GetLastScansByStatus(projectID, branch string, limit int, status []string) ([]checkmarxOne.Scan, error) {
	return []checkmarxOne.Scan{}, nil
}

func (sys *checkmarxOneSystemMock) ScanProject(projectID, sourceUrl, branch, scanType string, settings []checkmarxOne.ScanConfiguration, tags map[string]string) (checkmarxOne.Scan, error) {
	return checkmarxOne.Scan{}, nil
}

func (sys *checkmarxOneSystemMock) ScanProjectZip(projectID, sourceUrl, branch string, settings []checkmarxOne.ScanConfiguration, tags map[string]string) (checkmarxOne.Scan, error) {
	return checkmarxOne.Scan{}, nil
}

func (sys *checkmarxOneSystemMock) ScanProjectGit(projectID, repoUrl, branch string, settings []checkmarxOne.ScanConfiguration, tags map[string]string) (checkmarxOne.Scan, error) {
	return checkmarxOne.Scan{}, nil
}

func (sys *checkmarxOneSystemMock) UploadProjectSourceCode(projectID string, zipFile string) (string, error) {
	return "", nil
}

func (sys *checkmarxOneSystemMock) CreateProject(projectName string, groupIDs []string) (checkmarxOne.Project, error) {
	return checkmarxOne.Project{}, nil
}

func (sys *checkmarxOneSystemMock) CreateProjectInApplication(projectName, applicationId string, groupIDs []string) (checkmarxOne.Project, error) {
	return checkmarxOne.Project{}, nil
}

func (sys *checkmarxOneSystemMock) GetPresets() ([]checkmarxOne.Preset, error) {
	return []checkmarxOne.Preset{}, nil
}

func (sys *checkmarxOneSystemMock) GetProjectByID(projectID string) (checkmarxOne.Project, error) {
	return checkmarxOne.Project{}, nil
}

func (sys *checkmarxOneSystemMock) GetProjectsByName(projectName string) ([]checkmarxOne.Project, error) {
	str := `[        
		{
			"id": "3cb99ae5-5245-4cf7-83aa-9b517b8c1c57",
			"name": "ssba-github",
			"createdAt": "2023-03-21T16:48:33.224554Z",
			"updatedAt": "2023-03-21T16:48:33.224554Z",
			"groups": [
				"af361bd1-e478-40f6-a4fb-d479828d5998"
			],
			"tags": {},
			"repoUrl": "",
			"mainBranch": "",
			"criticality": 3
		},
		{
			"id": "3cb99ae5-5245-4cf7-83aa-9b517b8c1c58",
			"name": "ssba-local",
			"createdAt": "2023-03-21T16:48:33.224554Z",
			"updatedAt": "2023-03-21T16:48:33.224554Z",
			"groups": [
				"af361bd1-e478-40f6-a4fb-d479828d5998"
			],
			"tags": {},
			"repoUrl": "",
			"mainBranch": "",
			"criticality": 3
		},
		{
			"id": "3cb99ae5-5245-4cf7-83aa-9b517b8c1c59",
			"name": "ssba-zip",
			"createdAt": "2023-03-21T16:48:33.224554Z",
			"updatedAt": "2023-03-21T16:48:33.224554Z",
			"groups": [
				"af361bd1-e478-40f6-a4fb-d479828d5998"
			],
			"tags": {},
			"repoUrl": "",
			"mainBranch": "",
			"criticality": 3
		}
	]`
	projects := []checkmarxOne.Project{}
	_ = json.Unmarshal([]byte(str), &projects)

	return projects, nil
}

func (sys *checkmarxOneSystemMock) GetProjectsByNameAndGroup(projectName, groupID string) ([]checkmarxOne.Project, error) {
	return []checkmarxOne.Project{}, nil
}

func (sys *checkmarxOneSystemMock) GetProjects() ([]checkmarxOne.Project, error) {
	return []checkmarxOne.Project{}, nil
}

func (sys *checkmarxOneSystemMock) GetQueries() ([]checkmarxOne.Query, error) {
	return []checkmarxOne.Query{}, nil
}

func (sys *checkmarxOneSystemMock) GetGroups() ([]checkmarxOne.Group, error) {
	str := `
	[
		{
			"id": "d857c923-cf53-48bc-bfe4-163f66ed7b39",
			"name": "Group1"
		},
		{
			"id": "a8009bce-c24f-4edc-a931-06eb91ace2f5",
			"name": "Group2"
		},
		{
			"id": "a9ef684c-a61b-4647-9c49-363efc3879d7",
			"name": "01100035870000224721"
		},
		{
			"id": "3078680e-d796-4607-8e96-0d658eff799a",
			"name": "Group3"
		}
	]
	`
	groups := []checkmarxOne.Group{}
	_ = json.Unmarshal([]byte(str), &groups)

	return groups, nil
}

func (sys *checkmarxOneSystemMock) GetGroupByName(groupName string) (checkmarxOne.Group, error) {
	groups, err := sys.GetGroups()
	var group checkmarxOne.Group
	if err != nil {
		return group, err
	}

	for _, g := range groups {
		if g.Name == groupName {
			return g, nil
		}
	}

	return group, errors.New(fmt.Sprintf("No group matching %v", groupName))
}

func (sys *checkmarxOneSystemMock) GetGroupByID(groupID string) (checkmarxOne.Group, error) {
	return checkmarxOne.Group{}, nil
}

func (sys *checkmarxOneSystemMock) SetProjectBranch(projectID, branch string, allowOverride bool) error {
	return nil
}

func (sys *checkmarxOneSystemMock) SetProjectPreset(projectID, presetName string, allowOverride bool) error {
	return nil
}

func (sys *checkmarxOneSystemMock) SetProjectLanguageMode(projectID, languageMode string, allowOverride bool) error {
	return nil
}

func (sys *checkmarxOneSystemMock) SetProjectFileFilter(projectID, filter string, allowOverride bool) error {
	return nil
}

func (sys *checkmarxOneSystemMock) GetProjectConfiguration(projectID string) ([]checkmarxOne.ProjectConfigurationSetting, error) {
	return []checkmarxOne.ProjectConfigurationSetting{}, nil
}

func (sys *checkmarxOneSystemMock) UpdateProjectConfiguration(projectID string, settings []checkmarxOne.ProjectConfigurationSetting) error {
	return nil
}

func (sys *checkmarxOneSystemMock) UpdateProject(project *checkmarxOne.Project) error {
	return nil
}

func (sys *checkmarxOneSystemMock) GetVersion() (checkmarxOne.VersionInfo, error) {
	return checkmarxOne.VersionInfo{}, nil
}

type checkmarxOneExecuteScanHelperMock struct {
	ctx     context.Context
	config  checkmarxOneExecuteScanOptions
	sys     *checkmarxOne.SystemInstance
	influx  *checkmarxOneExecuteScanInflux
	utils   checkmarxOneExecuteScanUtils
	Project *checkmarxOne.Project
	Group   *checkmarxOne.Group
	App     *checkmarxOne.Application
	reports []piperutils.Path
}

func TestGetProjectByName(t *testing.T) {
	t.Parallel()
	sys := &checkmarxOneSystemMock{}
	t.Run("project name not found", func(t *testing.T) {
		t.Parallel()

		options := checkmarxOneExecuteScanOptions{ProjectName: "ssba_notexist", VulnerabilityThresholdUnit: "absolute", FullScanCycle: "2", Incremental: true, FullScansScheduled: true, Preset: "CheckmarxDefault", GroupName: "TestGroup", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true, APIKey: "testAPIKey", ServerURL: "testURL", IamURL: "testIamURL", Tenant: "testTenant"}

		cx1sh := checkmarxOneExecuteScanHelper{nil, options, sys, nil, nil, nil, nil, nil, nil}

		_, err := cx1sh.GetProjectByName()

		assert.Contains(t, fmt.Sprint(err), "project not found")
	})
	t.Run("project name exists", func(t *testing.T) {
		t.Parallel()

		options := checkmarxOneExecuteScanOptions{ProjectName: "ssba-github", VulnerabilityThresholdUnit: "absolute", FullScanCycle: "2", Incremental: true, FullScansScheduled: true, Preset: "CheckmarxDefault", GroupName: "TestGroup", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true, APIKey: "testAPIKey", ServerURL: "testURL", IamURL: "testIamURL", Tenant: "testTenant"}

		cx1sh := checkmarxOneExecuteScanHelper{nil, options, sys, nil, nil, nil, nil, nil, nil}

		project, err := cx1sh.GetProjectByName()
		assert.NoError(t, err, "Error occurred but none expected")
		assert.Equal(t, project.ProjectID, "3cb99ae5-5245-4cf7-83aa-9b517b8c1c57")
		assert.Equal(t, project.Name, "ssba-github")
		assert.Equal(t, project.Groups[0], "af361bd1-e478-40f6-a4fb-d479828d5998")
	})
}

func TestGetGroup(t *testing.T) {
	t.Parallel()

	sys := &checkmarxOneSystemMock{}

	t.Run("group ID and group name is not provided", func(t *testing.T) {
		t.Parallel()

		options := checkmarxOneExecuteScanOptions{ProjectName: "ssba", VulnerabilityThresholdUnit: "absolute", FullScanCycle: "2", Incremental: true, FullScansScheduled: true, Preset: "CheckmarxDefault" /*GroupName: "NotProvided",*/, VulnerabilityThresholdEnabled: true, GeneratePdfReport: true, APIKey: "testAPIKey", ServerURL: "testURL", IamURL: "testIamURL", Tenant: "testTenant"}

		cx1sh := checkmarxOneExecuteScanHelper{nil, options, sys, nil, nil, nil, nil, nil, nil}
		_, err := cx1sh.GetGroup()
		assert.Contains(t, fmt.Sprint(err), "No group name specified in configuration")
	})

	t.Run("group name not found", func(t *testing.T) {
		t.Parallel()

		options := checkmarxOneExecuteScanOptions{ProjectName: "ssba", VulnerabilityThresholdUnit: "absolute", FullScanCycle: "2", Incremental: true, FullScansScheduled: true, Preset: "CheckmarxDefault", GroupName: "GroupNotExist", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true, APIKey: "testAPIKey", ServerURL: "testURL", IamURL: "testIamURL", Tenant: "testTenant"}

		cx1sh := checkmarxOneExecuteScanHelper{nil, options, sys, nil, nil, nil, nil, nil, nil}

		_, err := cx1sh.GetGroup()
		assert.Contains(t, fmt.Sprint(err), "Failed to get Checkmarx One group by Name GroupNotExist: No group matching GroupNotExist")
	})

	t.Run("group name exists", func(t *testing.T) {
		t.Parallel()

		options := checkmarxOneExecuteScanOptions{ProjectName: "ssba-github", VulnerabilityThresholdUnit: "absolute", FullScanCycle: "2", Incremental: true, FullScansScheduled: true, Preset: "CheckmarxDefault", GroupName: "Group2", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true, APIKey: "testAPIKey", ServerURL: "testURL", IamURL: "testIamURL", Tenant: "testTenant"}

		cx1sh := checkmarxOneExecuteScanHelper{nil, options, sys, nil, nil, nil, nil, nil, nil}

		group, err := cx1sh.GetGroup()
		assert.NoError(t, err, "Error occurred but none expected")
		assert.Equal(t, group.GroupID, "a8009bce-c24f-4edc-a931-06eb91ace2f5")
		assert.Equal(t, group.Name, "Group2")
	})
}

func TestUpdateProjectTags(t *testing.T) {
	t.Parallel()

	sys := &checkmarxOneSystemMock{}

	t.Run("project tags are not provided", func(t *testing.T) {
		t.Parallel()

		options := checkmarxOneExecuteScanOptions{ProjectName: "ssba", VulnerabilityThresholdUnit: "absolute", FullScanCycle: "2", Incremental: true, FullScansScheduled: true, Preset: "CheckmarxDefault" /*GroupName: "NotProvided",*/, VulnerabilityThresholdEnabled: true, GeneratePdfReport: true, APIKey: "testAPIKey", ServerURL: "testURL", IamURL: "testIamURL", Tenant: "testTenant"}

		cx1sh := checkmarxOneExecuteScanHelper{nil, options, sys, nil, nil, nil, nil, nil, nil}
		err := cx1sh.UpdateProjectTags()
		assert.NoError(t, err, "Error occurred but none expected")
	})

	t.Run("project tags are provided correctly", func(t *testing.T) {
		t.Parallel()

		projectJson := `{ "id": "702ba12b-ae61-48c0-9b6a-09b17666be32",
			"name": "test-apr24-piper",
			"tags": {
				"key1": "value1",
				"key2": "value2", 
				"keywithoutvalue1": ""
			},
			"groups": [],
			"criticality": 3,
			"mainBranch": "",
			"privatePackage": false
		}`
		var project checkmarxOne.Project
		_ = json.Unmarshal([]byte(projectJson), &project)

		options := checkmarxOneExecuteScanOptions{ProjectName: "ssba", VulnerabilityThresholdUnit: "absolute", FullScanCycle: "2", Incremental: true, FullScansScheduled: true, Preset: "CheckmarxDefault" /*GroupName: "NotProvided",*/, VulnerabilityThresholdEnabled: true, GeneratePdfReport: true, APIKey: "testAPIKey", ServerURL: "testURL", IamURL: "testIamURL", Tenant: "testTenant", ProjectTags: `{"key3":"value3", "key2":"value5", "keywithoutvalue2":""}`}

		cx1sh := checkmarxOneExecuteScanHelper{nil, options, sys, nil, nil, &project, nil, nil, nil}
		err := cx1sh.UpdateProjectTags()
		assert.NoError(t, err, "Error occurred but none expected")

		oldTagsJson := `{
			"key1": "value1",
			"key2": "value2", 
			"keywithoutvalue1": ""
		}`
		oldTags := make(map[string]string, 0)
		_ = json.Unmarshal([]byte(oldTagsJson), &oldTags)

		newTagsJson := `{"key3":"value3", "key2":"value5", "keywithoutvalue2":""}`
		newTags := make(map[string]string, 0)
		_ = json.Unmarshal([]byte(newTagsJson), &newTags)

		// merge new tags to the existing ones
		maps.Copy(oldTags, newTags)

		assert.Equal(t, project.Tags, oldTags) // project's tags must be merged
	})
}
