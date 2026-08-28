package cmd

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	checkmarxOne "github.com/SAP/jenkins-library/pkg/checkmarxone"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

type checkmarxOneSystemMock struct {
	response any
	influx   *checkmarxOneExecuteScanInflux
}

func (sys *checkmarxOneSystemMock) DownloadReport(reportID string) ([]byte, error) {
	return nil, nil
}

func (sys *checkmarxOneSystemMock) GetReportStatus(reportID string) (checkmarxOne.ReportStatus, error) {
	return checkmarxOne.ReportStatus{}, nil
}

func (sys *checkmarxOneSystemMock) RequestNewReport(scanID, projectID, branch, reportType string, engines []string) (string, error) {
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

func (sys *checkmarxOneSystemMock) GetScanMetadata(scan *checkmarxOne.Scan) (checkmarxOne.ScanMetadata, error) {
	return checkmarxOne.ScanMetadata{}, nil
}

func (sys *checkmarxOneSystemMock) GetScanSASTMetadata(scanID string) (checkmarxOne.ScanSASTMetadata, error) {
	return checkmarxOne.ScanSASTMetadata{}, nil
}

func (sys *checkmarxOneSystemMock) GetScanIACMetadata(scanID string) (checkmarxOne.ScanIACMetadata, error) {
	return checkmarxOne.ScanIACMetadata{}, nil
}

func (sys *checkmarxOneSystemMock) GetScanSASTMetadatas(scanID []string) ([]checkmarxOne.ScanSASTMetadata, error) {
	return []checkmarxOne.ScanSASTMetadata{}, nil
}

func (sys *checkmarxOneSystemMock) GetScanConfiguration(_, _ string) (map[string]string, error) {
	return map[string]string{}, nil
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

	return group, fmt.Errorf("No group matching %v", groupName)
}

func (sys *checkmarxOneSystemMock) GetIACPresetNameByID(_ string) (string, error) {
	return "my-iac-preset", nil
}

func (sys *checkmarxOneSystemMock) GetIACPresetIDByName(_ string) (string, error) {
	return "a-b-c-d", nil
}

func (sys *checkmarxOneSystemMock) GetIACFindingInfo(_ checkmarxOne.ScanResult) (checkmarxOne.IACFindingInfo, error) {
	return checkmarxOne.IACFindingInfo{
		Cwe: 0,
		URL: "Test",
	}, nil
}

func (sys *checkmarxOneSystemMock) LoadIACHelpLinks(_ string) error {
	return nil
}

func (sys *checkmarxOneSystemMock) GetGroupByID(groupID string) (checkmarxOne.Group, error) {
	return checkmarxOne.Group{}, nil
}

func (sys *checkmarxOneSystemMock) SetProjectBranch(projectID, branch string, allowOverride bool) error {
	return nil
}

func (sys *checkmarxOneSystemMock) SetProjectLanguageMode(projectID, languageMode string, allowOverride bool) error {
	return nil
}

func (sys *checkmarxOneSystemMock) SetProjectSASTPreset(projectID, presetName string, allowOverride bool) error {
	return nil
}

func (sys *checkmarxOneSystemMock) SetProjectIACPreset(projectID, presetName string, allowOverride bool) error {
	return nil
}

func (sys *checkmarxOneSystemMock) SetProjectSASTFileFilter(projectID, filter string, allowOverride bool) error {
	return nil
}

func (sys *checkmarxOneSystemMock) SetProjectIACFileFilter(projectID, filter string, allowOverride bool) error {
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

		options := checkmarxOneExecuteScanOptions{ProjectName: "ssba_notexist", VulnerabilityThresholdUnit: "absolute", FullScanCycle: "2", Incremental: true, FullScansScheduled: true, SastPreset: "CheckmarxDefault", GroupName: "TestGroup", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true, APIKey: "testAPIKey", ServerURL: "testURL", IamURL: "testIamURL", Tenant: "testTenant"}

		cx1sh := checkmarxOneExecuteScanHelper{nil, options, sys, nil, nil, nil, nil, nil, true, false, nil}

		_, err := cx1sh.GetProjectByName()

		assert.Contains(t, fmt.Sprint(err), "project not found")
	})
	t.Run("project name exists", func(t *testing.T) {
		t.Parallel()

		options := checkmarxOneExecuteScanOptions{ProjectName: "ssba-github", VulnerabilityThresholdUnit: "absolute", FullScanCycle: "2", Incremental: true, FullScansScheduled: true, SastPreset: "CheckmarxDefault", GroupName: "TestGroup", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true, APIKey: "testAPIKey", ServerURL: "testURL", IamURL: "testIamURL", Tenant: "testTenant"}

		cx1sh := checkmarxOneExecuteScanHelper{nil, options, sys, nil, nil, nil, nil, nil, true, false, nil}

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

		options := checkmarxOneExecuteScanOptions{ProjectName: "ssba", VulnerabilityThresholdUnit: "absolute", FullScanCycle: "2", Incremental: true, FullScansScheduled: true, SastPreset: "CheckmarxDefault" /*GroupName: "NotProvided",*/, VulnerabilityThresholdEnabled: true, GeneratePdfReport: true, APIKey: "testAPIKey", ServerURL: "testURL", IamURL: "testIamURL", Tenant: "testTenant"}

		cx1sh := checkmarxOneExecuteScanHelper{nil, options, sys, nil, nil, nil, nil, nil, true, false, nil}
		_, err := cx1sh.GetGroup()
		assert.Contains(t, fmt.Sprint(err), "No group name specified in configuration")
	})

	t.Run("group name not found", func(t *testing.T) {
		t.Parallel()

		options := checkmarxOneExecuteScanOptions{ProjectName: "ssba", VulnerabilityThresholdUnit: "absolute", FullScanCycle: "2", Incremental: true, FullScansScheduled: true, SastPreset: "CheckmarxDefault", GroupName: "GroupNotExist", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true, APIKey: "testAPIKey", ServerURL: "testURL", IamURL: "testIamURL", Tenant: "testTenant"}

		cx1sh := checkmarxOneExecuteScanHelper{nil, options, sys, nil, nil, nil, nil, nil, true, false, nil}

		_, err := cx1sh.GetGroup()
		assert.Contains(t, fmt.Sprint(err), "Failed to get Checkmarx One group by Name GroupNotExist: No group matching GroupNotExist")
	})

	t.Run("group name exists", func(t *testing.T) {
		t.Parallel()

		options := checkmarxOneExecuteScanOptions{ProjectName: "ssba-github", VulnerabilityThresholdUnit: "absolute", FullScanCycle: "2", Incremental: true, FullScansScheduled: true, SastPreset: "CheckmarxDefault", GroupName: "Group2", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true, APIKey: "testAPIKey", ServerURL: "testURL", IamURL: "testIamURL", Tenant: "testTenant"}

		cx1sh := checkmarxOneExecuteScanHelper{nil, options, sys, nil, nil, nil, nil, nil, true, false, nil}

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

		options := checkmarxOneExecuteScanOptions{ProjectName: "ssba", VulnerabilityThresholdUnit: "absolute", FullScanCycle: "2", Incremental: true, FullScansScheduled: true, SastPreset: "CheckmarxDefault" /*GroupName: "NotProvided",*/, VulnerabilityThresholdEnabled: true, GeneratePdfReport: true, APIKey: "testAPIKey", ServerURL: "testURL", IamURL: "testIamURL", Tenant: "testTenant"}

		cx1sh := checkmarxOneExecuteScanHelper{nil, options, sys, nil, nil, nil, nil, nil, true, false, nil}
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

		options := checkmarxOneExecuteScanOptions{ProjectName: "ssba", VulnerabilityThresholdUnit: "absolute", FullScanCycle: "2", Incremental: true, FullScansScheduled: true, SastPreset: "CheckmarxDefault" /*GroupName: "NotProvided",*/, VulnerabilityThresholdEnabled: true, GeneratePdfReport: true, APIKey: "testAPIKey", ServerURL: "testURL", IamURL: "testIamURL", Tenant: "testTenant", ProjectTags: `{"key3":"value3", "key2":"value5", "keywithoutvalue2":""}`}

		cx1sh := checkmarxOneExecuteScanHelper{nil, options, sys, nil, nil, &project, nil, nil, true, false, nil}
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

func TestCheckmarxOneZipFolder(t *testing.T) {
	t.Parallel()

	t.Run("output archive is not zipped into itself", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		err := os.WriteFile(filepath.Join(dir, "abcd.go"), []byte("abcd.go"), 0o700)
		assert.NoError(t, err)
		err = os.Mkdir(filepath.Join(dir, "somepath"), 0o700)
		assert.NoError(t, err)
		err = os.WriteFile(filepath.Join(dir, "somepath", "abcd.txt"), []byte("somepath/abcd.txt"), 0o700)
		assert.NoError(t, err)

		// the output archive lives inside the folder being zipped, exactly like workspace.zip
		zipFileName := filepath.Join(dir, "workspace.zip")
		zipFile, err := os.Create(zipFileName)
		assert.NoError(t, err)
		defer zipFile.Close()

		cx1sh := checkmarxOneExecuteScanHelper{}
		utils := newcheckmarxOneExecuteScanUtilsBundle(dir, nil)

		// no filter pattern - every file is a candidate, including the output archive itself
		err = cx1sh.zipFolder(dir, zipFile, []string{}, zipFileName, utils)
		assert.NoError(t, err)

		zipInfo, err := zipFile.Stat()
		assert.NoError(t, err)
		reader, err := zip.NewReader(zipFile, zipInfo.Size())
		assert.NoError(t, err)

		for _, f := range reader.File {
			assert.NotEqual(t, "workspace.zip", filepath.Base(f.Name), "the output archive must not be zipped into itself")
		}
		// only the two regular source files must be archived, never the output archive
		assert.Len(t, reader.File, 2)
	})
}

func testRawDetailedResults() map[string]any {
	return map[string]any{
		"Application":                     "80fdd5b7-9269-423e-a483-878ecb3c7ae8",
		"ApplicationFullPathOnReportDate": "SSBA",
		"Critical":                        map[string]int{"Issues": 2, "NotExploitable": 2},
		"DeepLink":                        "test/projects/ca6bbcca-c75a-4a07-8298-552ec400732a/overview?branch=new-branch2",
		"FilesScanned":                    10,
		"Group":                           "",
		"GroupFullPathOnReportDate":       "",
		"High":                            map[string]int{"Issues": 1, "NotExploitable": 1},
		"IACCritical":                     map[string]int{},
		"IACHigh":                         map[string]int{"Issues": 1, "NotFalsePositive": 1, "ToVerify": 1},
		"IACInformation":                  map[string]int{},
		"IACLow":                          map[string]int{"Issues": 1, "NotFalsePositive": 1, "ToVerify": 1},
		"IACLowPerQuery": map[string]map[string]int{
			"Healthcheck Instruction Missing": map[string]int{"Issues": 1, "NotFalsePositive": 1, "ToVerify": 1},
		},
		"IACMedium":             map[string]int{},
		"IACVersion":            "IAC: 2.1.20",
		"IacFilesScanned":       1,
		"IacLinesOfCodeScanned": 11,
		"IacPreset":             "all checks",
		"Information":           map[string]int{"Issues": 13, "NotExploitable": 1, "NotFalsePositive": 12, "ToVerify": 12},
		"InitiatorName":         "michael.kubiaczyk@checkmarx.com",
		"LinesOfCodeScanned":    158,
		"Low":                   map[string]int{"Issues": 2, "NotExploitable": 1, "NotFalsePositive": 1, "ToVerify": 1},
		"LowPerQuery": map[string]map[string]int{
			"Reflected_XSS":                          map[string]int{"Issues": 1, "NotFalsePositive": 1, "ToVerify": 1},
			"Spring_Missing_Content_Security_Policy": map[string]int{"Issues": 1, "NotExploitable": 1},
		},
		"Medium":             map[string]int{"Issues": 4, "NotExploitable": 1, "NotFalsePositive": 3, "ToVerify": 3},
		"Owner":              "Cx1 Gap: no project owner",
		"ProjectId":          "ca6bbcca-c75a-4a07-8298-552ec400732a",
		"ProjectName":        "piper-iac-sast-test1",
		"ReportCreationTime": "2026-08-11 13:38:56.647088 +0200 CEST m=+1.527286701",
		"SASTVersion":        "SAST: 9.7.6",
		"SastPreset":         "All",
		"ScanId":             "bdff84f5-f226-4591-86e0-f43511474f35",
		"ScanStart":          "2026-08-05T14:40:31.707039Z",
		"ScanTime":           "9.299223s",
		"ScanType":           "Incremental",
		"ToolVersion":        "CxOne: 3.64.0",
	}
}

func TestCheckmarxOneGitComment(t *testing.T) {
	detailedResults := testRawDetailedResults()

	mainconfig := checkmarxOneExecuteScanOptions{
		VulnerabilityThresholdEnabled:        true,
		VulnerabilityThresholdCritical:       100,
		VulnerabilityThresholdHigh:           100,
		VulnerabilityThresholdMedium:         100,
		VulnerabilityThresholdLow:            10,
		VulnerabilityThresholdLowPerQuery:    true,
		VulnerabilityThresholdLowPerQueryMax: 10,
		VulnerabilityThresholdResult:         "FAILURE",
		VulnerabilityThresholdUnit:           "percentage",
		IacVulnerabilityThresholdEnabled:     false,
	}

	t.Run("sast report with LowPerQuery enabled", func(t *testing.T) {
		var sast_status gitComment
		sastScanReportOverview := checkmarxOne.CreateJSONHeaderReport(&detailedResults, "sast")
		sast_status.Parse(sastScanReportOverview.Findings, &mainconfig)
		sastTable := sast_status.String()

		sastScan := fmt.Sprintf(`**SAST Scan type**: %s
**SAST Scan Preset**: %s
**SAST Results**
%s

`, strings.ToLower(sastScanReportOverview.ScanType), sastScanReportOverview.Preset, sastTable)

		t.Logf("Generated report: %s", sastScan)
		expectedReport := `**SAST Scan type**: incremental
**SAST Scan Preset**: All
**SAST Results**
Severity | Number of unaudited findings
--- | ---
:bangbang: Critical | :white_check_mark: 0
:red_circle: High | :white_check_mark: 0
:orange_circle: Medium | :x: 3
:yellow_circle: Low | :x: 1 Reflected_XSS (0 audited / 1 required) <br>:white_check_mark: 0 Spring_Missing_Content_Security_Policy (1 audited / 1 required) <br>

`
		assert.Equal(t, expectedReport, sastScan)
	})

	t.Run("iac report with LowPerQuery enabled", func(t *testing.T) {
		var iac_status gitComment
		iacScanReportOverview := checkmarxOne.CreateJSONHeaderReport(&detailedResults, "iac")
		iac_status.Parse(iacScanReportOverview.Findings, &mainconfig)
		iacTable := iac_status.String()

		iacScan := fmt.Sprintf(`**IAC Scan Preset**: %s
**IAC Results**
%s

`, iacScanReportOverview.Preset, iacTable)

		t.Logf("Generated report: %s", iacScan)

		expectedReport := `**IAC Scan Preset**: all checks
**IAC Results**
Severity | Number of unaudited findings
--- | ---
:bangbang: Critical | :white_check_mark: 0
:red_circle: High | :x: 1
:orange_circle: Medium | :white_check_mark: 0
:yellow_circle: Low | :x: 1 Healthcheck Instruction Missing (0 audited / 1 required) <br>

`
		assert.Equal(t, expectedReport, iacScan)
	})

	// remove lowPerQuery from config
	mainconfig.VulnerabilityThresholdLowPerQuery = false
	// remove lowPerQuery data from detailedResults
	delete(detailedResults, "LowPerQuery")
	delete(detailedResults, "IACLowPerQuery")

	t.Run("sast report with LowPerQuery disabled", func(t *testing.T) {
		var sast_status gitComment
		sastScanReportOverview := checkmarxOne.CreateJSONHeaderReport(&detailedResults, "sast")
		sast_status.Parse(sastScanReportOverview.Findings, &mainconfig)
		sastTable := sast_status.String()

		sastScan := fmt.Sprintf(`**SAST Scan type**: %s
**SAST Scan Preset**: %s
**SAST Results**
%s

`, strings.ToLower(sastScanReportOverview.ScanType), sastScanReportOverview.Preset, sastTable)

		t.Logf("Generated report: %s", sastScan)

		expectedReport := `**SAST Scan type**: incremental
**SAST Scan Preset**: All
**SAST Results**
Severity | Number of unaudited findings
--- | ---
:bangbang: Critical | :white_check_mark: 0
:red_circle: High | :white_check_mark: 0
:orange_circle: Medium | :x: 3
:yellow_circle: Low | :white_check_mark: 1

`

		assert.Equal(t, expectedReport, sastScan)
	})

	t.Run("iac report with LowPerQuery disabled", func(t *testing.T) {
		var iac_status gitComment
		iacScanReportOverview := checkmarxOne.CreateJSONHeaderReport(&detailedResults, "iac")
		iac_status.Parse(iacScanReportOverview.Findings, &mainconfig)
		iacTable := iac_status.String()

		iacScan := fmt.Sprintf(`**IAC Scan Preset**: %s
**IAC Results**
%s

`, iacScanReportOverview.Preset, iacTable)

		t.Logf("Generated report: %s", iacScan)

		expectedReport := `**IAC Scan Preset**: all checks
**IAC Results**
Severity | Number of unaudited findings
--- | ---
:bangbang: Critical | :white_check_mark: 0
:red_circle: High | :x: 1
:orange_circle: Medium | :white_check_mark: 0
:yellow_circle: Low | :x: 1

`

		assert.Equal(t, expectedReport, iacScan)
	})

	// remove all findings from report
	detailedResults["Critical"] = map[string]int{}
	detailedResults["High"] = map[string]int{}
	detailedResults["Medium"] = map[string]int{}
	detailedResults["Low"] = map[string]int{}
	detailedResults["Information"] = map[string]int{}
	detailedResults["IACCritical"] = map[string]int{}
	detailedResults["IACHigh"] = map[string]int{}
	detailedResults["IACMedium"] = map[string]int{}
	detailedResults["IACLow"] = map[string]int{}
	detailedResults["IACInformation"] = map[string]int{}
	t.Run("sast report with no findings LowPerQuery disabled", func(t *testing.T) {
		var sast_status gitComment
		sastScanReportOverview := checkmarxOne.CreateJSONHeaderReport(&detailedResults, "sast")
		sast_status.Parse(sastScanReportOverview.Findings, &mainconfig)
		sastTable := sast_status.String()

		sastScan := fmt.Sprintf(`**SAST Scan type**: %s
**SAST Scan Preset**: %s
**SAST Results**
%s

`, strings.ToLower(sastScanReportOverview.ScanType), sastScanReportOverview.Preset, sastTable)

		t.Logf("Generated report: %s", sastScan)

		expectedReport := `**SAST Scan type**: incremental
**SAST Scan Preset**: All
**SAST Results**
Severity | Number of unaudited findings
--- | ---
:bangbang: Critical | :white_check_mark: 0
:red_circle: High | :white_check_mark: 0
:orange_circle: Medium | :white_check_mark: 0
:yellow_circle: Low | :white_check_mark: 0

`

		assert.Equal(t, expectedReport, sastScan)
	})

	t.Run("iac report with no findings LowPerQuery disabled", func(t *testing.T) {
		var iac_status gitComment
		iacScanReportOverview := checkmarxOne.CreateJSONHeaderReport(&detailedResults, "iac")
		iac_status.Parse(iacScanReportOverview.Findings, &mainconfig)
		iacTable := iac_status.String()

		iacScan := fmt.Sprintf(`**IAC Scan Preset**: %s
**IAC Results**
%s

`, iacScanReportOverview.Preset, iacTable)

		t.Logf("Generated report: %s", iacScan)

		expectedReport := `**IAC Scan Preset**: all checks
**IAC Results**
Severity | Number of unaudited findings
--- | ---
:bangbang: Critical | :white_check_mark: 0
:red_circle: High | :white_check_mark: 0
:orange_circle: Medium | :white_check_mark: 0
:yellow_circle: Low | :white_check_mark: 0

`

		assert.Equal(t, expectedReport, iacScan)
	})

	// add lowPerQuery to config
	mainconfig.VulnerabilityThresholdLowPerQuery = true
	// add lowPerQuery data to detailedResults
	detailedResults["LowPerQuery"] = map[string]map[string]int{}
	detailedResults["IACLowPerQuery"] = map[string]map[string]int{}

	t.Run("sast report with no findings LowPerQuery enabled", func(t *testing.T) {
		var sast_status gitComment
		sastScanReportOverview := checkmarxOne.CreateJSONHeaderReport(&detailedResults, "sast")
		sast_status.Parse(sastScanReportOverview.Findings, &mainconfig)
		sastTable := sast_status.String()

		sastScan := fmt.Sprintf(`**SAST Scan type**: %s
**SAST Scan Preset**: %s
**SAST Results**
%s

`, strings.ToLower(sastScanReportOverview.ScanType), sastScanReportOverview.Preset, sastTable)

		t.Logf("Generated report: %s", sastScan)

		expectedReport := `**SAST Scan type**: incremental
**SAST Scan Preset**: All
**SAST Results**
Severity | Number of unaudited findings
--- | ---
:bangbang: Critical | :white_check_mark: 0
:red_circle: High | :white_check_mark: 0
:orange_circle: Medium | :white_check_mark: 0
:yellow_circle: Low | :white_check_mark: 0

`

		assert.Equal(t, expectedReport, sastScan)
	})

	t.Run("iac report with no findings LowPerQuery enabled", func(t *testing.T) {
		var iac_status gitComment
		iacScanReportOverview := checkmarxOne.CreateJSONHeaderReport(&detailedResults, "iac")
		iac_status.Parse(iacScanReportOverview.Findings, &mainconfig)
		iacTable := iac_status.String()

		iacScan := fmt.Sprintf(`**IAC Scan Preset**: %s
**IAC Results**
%s

`, iacScanReportOverview.Preset, iacTable)

		t.Logf("Generated report: %s", iacScan)

		expectedReport := `**IAC Scan Preset**: all checks
**IAC Results**
Severity | Number of unaudited findings
--- | ---
:bangbang: Critical | :white_check_mark: 0
:red_circle: High | :white_check_mark: 0
:orange_circle: Medium | :white_check_mark: 0
:yellow_circle: Low | :white_check_mark: 0

`

		assert.Equal(t, expectedReport, iacScan)
	})
}

func TestCheckmarxOneInflux(t *testing.T) {
	influx := checkmarxOneExecuteScanInflux{}
	cx1sh := checkmarxOneExecuteScanHelper{
		influx: &influx,
	}

	type measurement struct {
		Name  string
		Value any
	}
	referenceData := []measurement{
		{Name: "checkmarxOne", Value: false},
		{Name: "critical_issues", Value: 2},
		{Name: "critical_not_false_postive", Value: 0},
		{Name: "critical_not_exploitable", Value: 2},
		{Name: "critical_confirmed", Value: 0},
		{Name: "critical_urgent", Value: 0},
		{Name: "critical_proposed_not_exploitable", Value: 0},
		{Name: "critical_to_verify", Value: 0},
		{Name: "high_issues", Value: 2},
		{Name: "high_not_false_postive", Value: 1},
		{Name: "high_not_exploitable", Value: 1},
		{Name: "high_confirmed", Value: 0},
		{Name: "high_urgent", Value: 0},
		{Name: "high_proposed_not_exploitable", Value: 0},
		{Name: "high_to_verify", Value: 1},
		{Name: "medium_issues", Value: 4},
		{Name: "medium_not_false_postive", Value: 3},
		{Name: "medium_not_exploitable", Value: 1},
		{Name: "medium_confirmed", Value: 0},
		{Name: "medium_urgent", Value: 0},
		{Name: "medium_proposed_not_exploitable", Value: 0},
		{Name: "medium_to_verify", Value: 3},
		{Name: "low_issues", Value: 3},
		{Name: "low_not_false_postive", Value: 2},
		{Name: "low_not_exploitable", Value: 1},
		{Name: "low_confirmed", Value: 0},
		{Name: "low_urgent", Value: 0},
		{Name: "low_proposed_not_exploitable", Value: 0},
		{Name: "low_to_verify", Value: 2},
		{Name: "information_issues", Value: 13},
		{Name: "information_not_false_postive", Value: 12},
		{Name: "information_not_exploitable", Value: 1},
		{Name: "information_confirmed", Value: 0},
		{Name: "information_urgent", Value: 0},
		{Name: "information_proposed_not_exploitable", Value: 0},
		{Name: "information_to_verify", Value: 12},
		{Name: "lines_of_code_scanned", Value: 158},
		{Name: "files_scanned", Value: 10},
		{Name: "initiator_name", Value: "michael.kubiaczyk@checkmarx.com"},
		{Name: "owner", Value: "Cx1 Gap: no project owner"},
		{Name: "scan_id", Value: "bdff84f5-f226-4591-86e0-f43511474f35"},
		{Name: "project_id", Value: "ca6bbcca-c75a-4a07-8298-552ec400732a"},
		{Name: "projectName", Value: "piper-iac-sast-test1"},
		{Name: "group", Value: ""},
		{Name: "group_full_path_on_report_date", Value: ""},
		{Name: "scan_start", Value: "2026-08-05T14:40:31.707039Z"},
		{Name: "scan_time", Value: "9.299223s"},
		{Name: "tool_version", Value: "CxOne: 3.64.0, SAST: 9.7.6, IAC: 2.1.20"},
		{Name: "scan_type", Value: "Incremental"},
		{Name: "preset", Value: "All"},
		{Name: "iac_preset", Value: "all checks"},
		{Name: "deep_link", Value: "test/projects/ca6bbcca-c75a-4a07-8298-552ec400732a/overview?branch=new-branch2"},
		{Name: "report_creation_time", Value: "2026-08-11 13:38:56.647088 +0200 CEST m=+1.527286701"},
	}

	detailedResults := testRawDetailedResults()
	cx1sh.reportToInflux(&detailedResults)
	measurementContent := []measurement{
		{Name: "checkmarxOne", Value: influx.step_data.fields.checkmarxOne},
		{Name: "critical_issues", Value: influx.checkmarxOne_data.fields.critical_issues},
		{Name: "critical_not_false_postive", Value: influx.checkmarxOne_data.fields.critical_not_false_postive},
		{Name: "critical_not_exploitable", Value: influx.checkmarxOne_data.fields.critical_not_exploitable},
		{Name: "critical_confirmed", Value: influx.checkmarxOne_data.fields.critical_confirmed},
		{Name: "critical_urgent", Value: influx.checkmarxOne_data.fields.critical_urgent},
		{Name: "critical_proposed_not_exploitable", Value: influx.checkmarxOne_data.fields.critical_proposed_not_exploitable},
		{Name: "critical_to_verify", Value: influx.checkmarxOne_data.fields.critical_to_verify},
		{Name: "high_issues", Value: influx.checkmarxOne_data.fields.high_issues},
		{Name: "high_not_false_postive", Value: influx.checkmarxOne_data.fields.high_not_false_postive},
		{Name: "high_not_exploitable", Value: influx.checkmarxOne_data.fields.high_not_exploitable},
		{Name: "high_confirmed", Value: influx.checkmarxOne_data.fields.high_confirmed},
		{Name: "high_urgent", Value: influx.checkmarxOne_data.fields.high_urgent},
		{Name: "high_proposed_not_exploitable", Value: influx.checkmarxOne_data.fields.high_proposed_not_exploitable},
		{Name: "high_to_verify", Value: influx.checkmarxOne_data.fields.high_to_verify},
		{Name: "medium_issues", Value: influx.checkmarxOne_data.fields.medium_issues},
		{Name: "medium_not_false_postive", Value: influx.checkmarxOne_data.fields.medium_not_false_postive},
		{Name: "medium_not_exploitable", Value: influx.checkmarxOne_data.fields.medium_not_exploitable},
		{Name: "medium_confirmed", Value: influx.checkmarxOne_data.fields.medium_confirmed},
		{Name: "medium_urgent", Value: influx.checkmarxOne_data.fields.medium_urgent},
		{Name: "medium_proposed_not_exploitable", Value: influx.checkmarxOne_data.fields.medium_proposed_not_exploitable},
		{Name: "medium_to_verify", Value: influx.checkmarxOne_data.fields.medium_to_verify},
		{Name: "low_issues", Value: influx.checkmarxOne_data.fields.low_issues},
		{Name: "low_not_false_postive", Value: influx.checkmarxOne_data.fields.low_not_false_postive},
		{Name: "low_not_exploitable", Value: influx.checkmarxOne_data.fields.low_not_exploitable},
		{Name: "low_confirmed", Value: influx.checkmarxOne_data.fields.low_confirmed},
		{Name: "low_urgent", Value: influx.checkmarxOne_data.fields.low_urgent},
		{Name: "low_proposed_not_exploitable", Value: influx.checkmarxOne_data.fields.low_proposed_not_exploitable},
		{Name: "low_to_verify", Value: influx.checkmarxOne_data.fields.low_to_verify},
		{Name: "information_issues", Value: influx.checkmarxOne_data.fields.information_issues},
		{Name: "information_not_false_postive", Value: influx.checkmarxOne_data.fields.information_not_false_postive},
		{Name: "information_not_exploitable", Value: influx.checkmarxOne_data.fields.information_not_exploitable},
		{Name: "information_confirmed", Value: influx.checkmarxOne_data.fields.information_confirmed},
		{Name: "information_urgent", Value: influx.checkmarxOne_data.fields.information_urgent},
		{Name: "information_proposed_not_exploitable", Value: influx.checkmarxOne_data.fields.information_proposed_not_exploitable},
		{Name: "information_to_verify", Value: influx.checkmarxOne_data.fields.information_to_verify},
		{Name: "lines_of_code_scanned", Value: influx.checkmarxOne_data.fields.lines_of_code_scanned},
		{Name: "files_scanned", Value: influx.checkmarxOne_data.fields.files_scanned},
		{Name: "initiator_name", Value: influx.checkmarxOne_data.fields.initiator_name},
		{Name: "owner", Value: influx.checkmarxOne_data.fields.owner},
		{Name: "scan_id", Value: influx.checkmarxOne_data.fields.scan_id},
		{Name: "project_id", Value: influx.checkmarxOne_data.fields.project_id},
		{Name: "projectName", Value: influx.checkmarxOne_data.fields.projectName},
		{Name: "group", Value: influx.checkmarxOne_data.fields.group},
		{Name: "group_full_path_on_report_date", Value: influx.checkmarxOne_data.fields.group_full_path_on_report_date},
		{Name: "scan_start", Value: influx.checkmarxOne_data.fields.scan_start},
		{Name: "scan_time", Value: influx.checkmarxOne_data.fields.scan_time},
		{Name: "tool_version", Value: influx.checkmarxOne_data.fields.tool_version},
		{Name: "scan_type", Value: influx.checkmarxOne_data.fields.scan_type},
		{Name: "preset", Value: influx.checkmarxOne_data.fields.preset},
		{Name: "iac_preset", Value: influx.checkmarxOne_data.fields.iac_preset},
		{Name: "deep_link", Value: influx.checkmarxOne_data.fields.deep_link},
		{Name: "report_creation_time", Value: influx.checkmarxOne_data.fields.report_creation_time},
	}

	data, _ := json.Marshal(measurementContent)
	reference, _ := json.Marshal(referenceData)
	assert.Equal(t, reference, data)
}
