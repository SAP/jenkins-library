package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/checkmarx"
	"github.com/stretchr/testify/assert"
)

type fileInfo struct {
	nam     string      // base name of the file
	siz     int64       // length in bytes for regular files; system-dependent for others
	mod     os.FileMode // file mode bits
	modtime time.Time   // modification time
	dir     bool        // abbreviation for Mode().IsDir()
	syss    interface{} // underlying data source (can return nil)
}

func (fi fileInfo) IsDir() bool {
	return fi.dir
}
func (fi fileInfo) Name() string {
	return fi.nam
}
func (fi fileInfo) Size() int64 {
	return fi.siz
}
func (fi fileInfo) ModTime() time.Time {
	return fi.modtime
}
func (fi fileInfo) Mode() os.FileMode {
	return fi.mod
}
func (fi fileInfo) Sys() interface{} {
	return fi.syss
}

type systemMock struct {
	response                         interface{}
	isIncremental                    bool
	isPublic                         bool
	forceScan                        bool
	createProject                    bool
	previousPName                    string
	getPresetsCalled                 bool
	updateProjectConfigurationCalled bool
}

func (sys *systemMock) FilterPresetByName(presets []checkmarx.Preset, presetName string) checkmarx.Preset {
	if presetName == "CX_Default" {
		return checkmarx.Preset{ID: 16, Name: "CX_Default", OwnerName: "16"}
	}
	return checkmarx.Preset{ID: 10050, Name: "SAP_JS_Default", OwnerName: "16"}
}
func (sys *systemMock) FilterPresetByID(presets []checkmarx.Preset, presetID int) checkmarx.Preset {
	return checkmarx.Preset{ID: 10048, Name: "SAP_Default", OwnerName: "16"}
}
func (sys *systemMock) FilterProjectByName(projects []checkmarx.Project, projectName string) checkmarx.Project {
	return checkmarx.Project{ID: 1, Name: "Test", TeamID: "16", IsPublic: true}
}
func (sys *systemMock) GetProjectByID(projectID int) (checkmarx.Project, error) {
	if projectID == 17 {
		return checkmarx.Project{ID: 17, Name: "Test_PR-17", TeamID: "16", IsPublic: true}, nil
	}
	return checkmarx.Project{ID: 19, Name: "Test_PR-19", TeamID: "16", IsPublic: true}, nil
}

func (sys *systemMock) GetProjectsByNameAndTeam(projectName, teamID string) ([]checkmarx.Project, error) {
	if !sys.createProject || sys.previousPName == projectName {
		if strings.Contains(projectName, "PR-17") {
			return []checkmarx.Project{{ID: 17, Name: projectName, TeamID: teamID, IsPublic: true}}, nil
		}
		return []checkmarx.Project{{ID: 19, Name: projectName, TeamID: teamID, IsPublic: true}}, nil
	}
	if projectName == "Test" {
		return []checkmarx.Project{{ID: 1, Name: projectName, TeamID: teamID, IsPublic: true}}, nil
	}
	sys.previousPName = projectName
	return []checkmarx.Project{}, fmt.Errorf("no project error")
}
func (sys *systemMock) FilterTeamByName(teams []checkmarx.Team, teamName string) checkmarx.Team {
	if teamName == "OpenSource/Cracks/16" {
		return checkmarx.Team{ID: json.RawMessage(`"16"`), FullName: "OpenSource/Cracks/16"}
	}
	return checkmarx.Team{ID: json.RawMessage(`15`), FullName: "OpenSource/Cracks/15"}
}
func (sys *systemMock) FilterTeamByID(teams []checkmarx.Team, teamID json.RawMessage) checkmarx.Team {
	teamIDBytes, _ := teamID.MarshalJSON()
	if bytes.Compare(teamIDBytes, []byte(`"16"`)) == 0 {
		return checkmarx.Team{ID: json.RawMessage(`"16"`), FullName: "OpenSource/Cracks/16"}
	}
	return checkmarx.Team{ID: json.RawMessage(`15`), FullName: "OpenSource/Cracks/15"}
}
func (sys *systemMock) DownloadReport(reportID int) ([]byte, error) {
	return sys.response.([]byte), nil
}
func (sys *systemMock) GetReportStatus(reportID int) (checkmarx.ReportStatusResponse, error) {
	return checkmarx.ReportStatusResponse{Status: checkmarx.ReportStatus{ID: 2, Value: "Created"}}, nil
}
func (sys *systemMock) RequestNewReport(scanID int, reportType string) (checkmarx.Report, error) {
	return checkmarx.Report{ReportID: 17}, nil
}
func (sys *systemMock) GetResults(scanID int) checkmarx.ResultsStatistics {
	return checkmarx.ResultsStatistics{}
}
func (sys *systemMock) GetScans(projectID int) ([]checkmarx.ScanStatus, error) {
	return []checkmarx.ScanStatus{{IsIncremental: true}, {IsIncremental: true}, {IsIncremental: true}, {IsIncremental: false}}, nil
}
func (sys *systemMock) GetScanStatusAndDetail(scanID int) (string, checkmarx.ScanStatusDetail) {
	return "Finished", checkmarx.ScanStatusDetail{Stage: "Step 1 of 25", Step: "Scan something"}
}
func (sys *systemMock) ScanProject(projectID int, isIncrementalV, isPublicV, forceScanV bool) (checkmarx.Scan, error) {
	sys.isIncremental = isIncrementalV
	sys.isPublic = isPublicV
	sys.forceScan = forceScanV
	return checkmarx.Scan{ID: 16}, nil
}
func (sys *systemMock) UpdateProjectConfiguration(projectID int, presetID int, engineConfigurationID string) error {
	sys.updateProjectConfigurationCalled = true
	return nil
}
func (sys *systemMock) UpdateProjectExcludeSettings(projectID int, excludeFolders string, excludeFiles string) error {
	return nil
}
func (sys *systemMock) UploadProjectSourceCode(projectID int, zipFile string) error {
	return nil
}
func (sys *systemMock) CreateProject(projectName string, teamID string) (checkmarx.ProjectCreateResult, error) {
	return checkmarx.ProjectCreateResult{ID: 20}, nil
}
func (sys *systemMock) CreateBranch(projectID int, branchName string) int {
	return 18
}
func (sys *systemMock) GetPresets() []checkmarx.Preset {
	sys.getPresetsCalled = true
	return []checkmarx.Preset{{ID: 10078, Name: "SAP Java Default", OwnerName: "16"}, {ID: 10048, Name: "SAP JS Default", OwnerName: "16"}, {ID: 16, Name: "CX_Default", OwnerName: "16"}}
}
func (sys *systemMock) GetProjects() ([]checkmarx.Project, error) {
	return []checkmarx.Project{{ID: 15, Name: "OtherTest", TeamID: "16"}, {ID: 1, Name: "Test", TeamID: "16"}}, nil
}
func (sys *systemMock) GetTeams() []checkmarx.Team {
	return []checkmarx.Team{{ID: json.RawMessage(`"16"`), FullName: "OpenSource/Cracks/16"}, {ID: json.RawMessage(`15`), FullName: "OpenSource/Cracks/15"}}
}

type systemMockForExistingProject struct {
	response          interface{}
	isIncremental     bool
	isPublic          bool
	forceScan         bool
	scanProjectCalled bool
}

func (sys *systemMockForExistingProject) FilterPresetByName(presets []checkmarx.Preset, presetName string) checkmarx.Preset {
	return checkmarx.Preset{ID: 10050, Name: "SAP_JS_Default", OwnerName: "16"}
}
func (sys *systemMockForExistingProject) FilterPresetByID(presets []checkmarx.Preset, presetID int) checkmarx.Preset {
	return checkmarx.Preset{ID: 10048, Name: "SAP_Default", OwnerName: "16"}
}
func (sys *systemMockForExistingProject) FilterProjectByName(projects []checkmarx.Project, projectName string) checkmarx.Project {
	return checkmarx.Project{ID: 1, Name: "TestExisting", TeamID: "16", IsPublic: true}
}
func (sys *systemMockForExistingProject) GetProjectByID(projectID int) (checkmarx.Project, error) {
	return checkmarx.Project{}, nil
}
func (sys *systemMockForExistingProject) GetProjectsByNameAndTeam(projectName, teamID string) ([]checkmarx.Project, error) {
	return []checkmarx.Project{{ID: 19, Name: projectName, TeamID: teamID, IsPublic: true}}, nil
}
func (sys *systemMockForExistingProject) FilterTeamByName(teams []checkmarx.Team, teamName string) checkmarx.Team {
	return checkmarx.Team{ID: json.RawMessage(`"16"`), FullName: "OpenSource/Cracks/16"}
}
func (sys *systemMockForExistingProject) FilterTeamByID(teams []checkmarx.Team, teamID json.RawMessage) checkmarx.Team {
	return checkmarx.Team{ID: json.RawMessage(`"15"`), FullName: "OpenSource/Cracks/15"}
}
func (sys *systemMockForExistingProject) DownloadReport(reportID int) ([]byte, error) {
	return sys.response.([]byte), nil
}
func (sys *systemMockForExistingProject) GetReportStatus(reportID int) (checkmarx.ReportStatusResponse, error) {
	return checkmarx.ReportStatusResponse{Status: checkmarx.ReportStatus{ID: 2, Value: "Created"}}, nil
}
func (sys *systemMockForExistingProject) RequestNewReport(scanID int, reportType string) (checkmarx.Report, error) {
	return checkmarx.Report{ReportID: 17}, nil
}
func (sys *systemMockForExistingProject) GetResults(scanID int) checkmarx.ResultsStatistics {
	return checkmarx.ResultsStatistics{}
}
func (sys *systemMockForExistingProject) GetScans(projectID int) ([]checkmarx.ScanStatus, error) {
	return []checkmarx.ScanStatus{{IsIncremental: true}, {IsIncremental: true}, {IsIncremental: true}, {IsIncremental: false}}, nil
}
func (sys *systemMockForExistingProject) GetScanStatusAndDetail(scanID int) (string, checkmarx.ScanStatusDetail) {
	return "Finished", checkmarx.ScanStatusDetail{Stage: "", Step: ""}
}
func (sys *systemMockForExistingProject) ScanProject(projectID int, isIncrementalV, isPublicV, forceScanV bool) (checkmarx.Scan, error) {
	sys.scanProjectCalled = true
	sys.isIncremental = isIncrementalV
	sys.isPublic = isPublicV
	sys.forceScan = forceScanV
	return checkmarx.Scan{ID: 16}, nil
}
func (sys *systemMockForExistingProject) UpdateProjectConfiguration(projectID int, presetID int, engineConfigurationID string) error {
	return nil
}
func (sys *systemMockForExistingProject) UpdateProjectExcludeSettings(projectID int, excludeFolders string, excludeFiles string) error {
	return nil
}
func (sys *systemMockForExistingProject) UploadProjectSourceCode(projectID int, zipFile string) error {
	return nil
}
func (sys *systemMockForExistingProject) CreateProject(projectName string, teamID string) (checkmarx.ProjectCreateResult, error) {
	return checkmarx.ProjectCreateResult{}, fmt.Errorf("create project error")
}
func (sys *systemMockForExistingProject) CreateBranch(projectID int, branchName string) int {
	return 0
}
func (sys *systemMockForExistingProject) GetPresets() []checkmarx.Preset {
	return []checkmarx.Preset{{ID: 10078, Name: "SAP_Java_Default", OwnerName: "16"}, {ID: 10048, Name: "SAP_JS_Default", OwnerName: "16"}}
}
func (sys *systemMockForExistingProject) GetProjects() ([]checkmarx.Project, error) {
	return []checkmarx.Project{{ID: 1, Name: "TestExisting", TeamID: "16"}}, nil
}
func (sys *systemMockForExistingProject) GetTeams() []checkmarx.Team {
	return []checkmarx.Team{{ID: json.RawMessage(`"16"`), FullName: "OpenSource/Cracks/16"}, {ID: json.RawMessage(`"15"`), FullName: "OpenSource/Cracks/15"}}
}

func TestFilterFileGlob(t *testing.T) {
	tt := []struct {
		input    string
		fInfo    fileInfo
		expected bool
	}{
		{input: "somepath/node_modules/someOther/some.file", fInfo: fileInfo{}, expected: true},
		{input: "somepath/non_modules/someOther/some.go", fInfo: fileInfo{}, expected: false},
		{input: ".xmake/someOther/some.go", fInfo: fileInfo{}, expected: true},
		{input: "another/vendor/some.html", fInfo: fileInfo{}, expected: false},
		{input: "another/vendor/some.pdf", fInfo: fileInfo{}, expected: true},
		{input: "another/vendor/some.test", fInfo: fileInfo{}, expected: true},
		{input: "some.test", fInfo: fileInfo{}, expected: false},
		{input: "a/b/c", fInfo: fileInfo{dir: true}, expected: false},
	}

	for k, v := range tt {
		assert.Equal(t, v.expected, filterFileGlob([]string{"!**/node_modules/**", "!**/.xmake/**", "!**/*_test.go", "!**/vendor/**/*.go", "**/*.go", "**/*.html", "*.test"}, v.input, v.fInfo), fmt.Sprintf("wrong long name for run %v", k))
	}
}

func TestZipFolder(t *testing.T) {

	t.Run("zip files", func(t *testing.T) {
		dir, err := ioutil.TempDir("", "test zip files")
		if err != nil {
			t.Fatal("Failed to create temporary directory")
		}
		// clean up tmp dir
		defer os.RemoveAll(dir)

		ioutil.WriteFile(filepath.Join(dir, "abcd.go"), []byte{byte(1), byte(2), byte(3)}, 0700)
		ioutil.WriteFile(filepath.Join(dir, "somepath", "abcd.txt"), []byte{byte(1), byte(2), byte(3)}, 0700)
		ioutil.WriteFile(filepath.Join(dir, "abcd_test.go"), []byte{byte(1), byte(2), byte(3)}, 0700)

		var zipFileMock bytes.Buffer
		zipFolder(dir, &zipFileMock, []string{"!abc_test.go", "**/abcd.txt", "**/abc.go"})

		got := zipFileMock.Len()
		want := 164

		if got != want {
			t.Errorf("Zipping test failed expected %v but got %v", want, got)
		}
	})
}

func TestGetDetailedResults(t *testing.T) {

	t.Run("success case", func(t *testing.T) {
		sys := &systemMock{response: []byte(`<?xml version="1.0" encoding="utf-8"?>
		<CxXMLResults InitiatorName="admin" Owner="admin" ScanId="1000005" ProjectId="2" ProjectName="Project 1" TeamFullPathOnReportDate="CxServer" DeepLink="http://WIN2K12-TEMP/CxWebClient/ViewerMain.aspx?scanid=1000005&amp;projectid=2" ScanStart="Sunday, December 3, 2017 4:50:34 PM" Preset="Checkmarx Default" ScanTime="00h:03m:18s" LinesOfCodeScanned="6838" FilesScanned="34" ReportCreationTime="Sunday, December 3, 2017 6:13:45 PM" Team="CxServer" CheckmarxVersion="8.6.0" ScanComments="" ScanType="Incremental" SourceOrigin="LocalPath" Visibility="Public">
		<Query id="430" categories="PCI DSS v3.2;PCI DSS (3.2) - 6.5.1 - Injection flaws - particularly SQL injection,OWASP Top 10 2013;A1-Injection,FISMA 2014;System And Information Integrity,NIST SP 800-53;SI-10 Information Input Validation (P1),OWASP Top 10 2017;A1-Injection" cweId="89" name="SQL_Injection" group="CSharp_High_Risk" Severity="High" Language="CSharp" LanguageHash="1363215419077432" LanguageChangeDate="2017-12-03T00:00:00.0000000" SeverityIndex="3" QueryPath="CSharp\Cx\CSharp High Risk\SQL Injection Version:0" QueryVersionCode="430">
			<Result NodeId="10000050002" FileName="bookstore/Login.cs" Status="Recurrent" Line="179" Column="103" FalsePositive="False" Severity="High" AssignToUser="" state="0" Remark="" DeepLink="http://WIN2K12-TEMP/CxWebClient/ViewerMain.aspx?scanid=1000005&amp;projectid=2&amp;pathid=2" SeverityIndex="3">
				<Path ResultId="1000005" PathId="2" SimilarityId="1765812516"/>
			</Result>
			<Result NodeId="10000050003" FileName="bookstore/Login.cs" Status="Recurrent" Line="180" Column="10" FalsePositive="False" Severity="High" AssignToUser="" state="1" Remark="" DeepLink="http://WIN2K12-TEMP/CxWebClient/ViewerMain.aspx?scanid=1000005&amp;projectid=2&amp;pathid=2" SeverityIndex="3">
				<Path ResultId="1000005" PathId="2" SimilarityId="1765812516"/>
			</Result>
			<Result NodeId="10000050004" FileName="bookstore/Login.cs" Status="Recurrent" Line="181" Column="190" FalsePositive="True" Severity="Medium" AssignToUser="" state="2" Remark="" DeepLink="http://WIN2K12-TEMP/CxWebClient/ViewerMain.aspx?scanid=1000005&amp;projectid=2&amp;pathid=2" SeverityIndex="2">
				<Path ResultId="1000005" PathId="2" SimilarityId="1765812516"/>
			</Result>
			<Result NodeId="10000050005" FileName="bookstore/Login.cs" Status="Recurrent" Line="181" Column="190" FalsePositive="True" Severity="Low" AssignToUser="" state="3" Remark="" DeepLink="http://WIN2K12-TEMP/CxWebClient/ViewerMain.aspx?scanid=1000005&amp;projectid=2&amp;pathid=2" SeverityIndex="2">
				<Path ResultId="1000005" PathId="2" SimilarityId="1765812516"/>
			</Result>
			<Result NodeId="10000050006" FileName="bookstore/Login.cs" Status="Recurrent" Line="181" Column="190" FalsePositive="True" Severity="Low" AssignToUser="" state="4" Remark="" DeepLink="http://WIN2K12-TEMP/CxWebClient/ViewerMain.aspx?scanid=1000005&amp;projectid=2&amp;pathid=2" SeverityIndex="2">
				<Path ResultId="1000005" PathId="2" SimilarityId="1765812516"/>
			</Result>
		</Query>
		</CxXMLResults>`)}
		dir, err := ioutil.TempDir("", "test detailed results")
		if err != nil {
			t.Fatal("Failed to create temporary directory")
		}
		// clean up tmp dir
		defer os.RemoveAll(dir)
		result, err := getDetailedResults(sys, filepath.Join(dir, "abc.xml"), 2635)
		assert.NoError(t, err, "error occured but none expected")
		assert.Equal(t, "2", result["ProjectId"], "Project ID incorrect")
		assert.Equal(t, "Project 1", result["ProjectName"], "Project name incorrect")
		assert.Equal(t, 2, result["High"].(map[string]int)["Issues"], "Number of High issues incorrect")
		assert.Equal(t, 2, result["High"].(map[string]int)["NotFalsePositive"], "Number of High NotFalsePositive issues incorrect")
		assert.Equal(t, 1, result["Medium"].(map[string]int)["Issues"], "Number of Medium issues incorrect")
		assert.Equal(t, 0, result["Medium"].(map[string]int)["NotFalsePositive"], "Number of Medium NotFalsePositive issues incorrect")
	})
}

func TestRunScan(t *testing.T) {
	sys := &systemMockForExistingProject{response: []byte(`<?xml version="1.0" encoding="utf-8"?><CxXMLResults />`)}
	options := checkmarxExecuteScanOptions{ProjectName: "TestExisting", VulnerabilityThresholdUnit: "absolute", FullScanCycle: "2", Incremental: true, FullScansScheduled: true, Preset: "10048", TeamID: "16", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true}
	workspace, err := ioutil.TempDir("", "workspace1")
	if err != nil {
		t.Fatal("Failed to create temporary workspace directory")
	}
	// clean up tmp dir
	defer os.RemoveAll(workspace)

	influx := checkmarxExecuteScanInflux{}

	err = runScan(options, sys, workspace, &influx)
	assert.NoError(t, err, "error occured but none expected")
	assert.Equal(t, false, sys.isIncremental, "isIncremental has wrong value")
	assert.Equal(t, true, sys.isPublic, "isPublic has wrong value")
	assert.Equal(t, true, sys.forceScan, "forceScan has wrong value")
	assert.Equal(t, true, sys.scanProjectCalled, "ScanProject was not invoked")
}

func TestSetPresetForProjectWithIDProvided(t *testing.T) {
	sys := &systemMock{}
	err := setPresetForProject(sys, 12345, 16, "testProject", "CX_Default", "")
	assert.NoError(t, err, "error occured but none expected")
	assert.Equal(t, false, sys.getPresetsCalled, "GetPresets was called")
	assert.Equal(t, true, sys.updateProjectConfigurationCalled, "UpdateProjectConfiguration was not called")
}

func TestSetPresetForProjectWithNameProvided(t *testing.T) {
	sys := &systemMock{}
	presetID, _ := strconv.Atoi("CX_Default")
	err := setPresetForProject(sys, 12345, presetID, "testProject", "CX_Default", "")
	assert.NoError(t, err, "error occured but none expected")
	assert.Equal(t, true, sys.getPresetsCalled, "GetPresets was not called")
	assert.Equal(t, true, sys.updateProjectConfigurationCalled, "UpdateProjectConfiguration was not called")
}

func TestVerifyOnly(t *testing.T) {
	sys := &systemMockForExistingProject{response: []byte(`<?xml version="1.0" encoding="utf-8"?><CxXMLResults />`)}
	options := checkmarxExecuteScanOptions{VerifyOnly: true, ProjectName: "TestExisting", VulnerabilityThresholdUnit: "absolute", FullScanCycle: "2", Incremental: true, FullScansScheduled: true, Preset: "10048", TeamName: "OpenSource/Cracks/15", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true}
	workspace, err := ioutil.TempDir("", "workspace1")
	if err != nil {
		t.Fatal("Failed to create temporary workspace directory")
	}
	// clean up tmp dir
	defer os.RemoveAll(workspace)

	influx := checkmarxExecuteScanInflux{}

	err = runScan(options, sys, workspace, &influx)
	assert.NoError(t, err, "error occured but none expected")
	assert.Equal(t, false, sys.scanProjectCalled, "ScanProject was invoked but shouldn't")
}

func TestRunScanWOtherCycle(t *testing.T) {
	sys := &systemMock{response: []byte(`<?xml version="1.0" encoding="utf-8"?><CxXMLResults />`), createProject: true}
	options := checkmarxExecuteScanOptions{VulnerabilityThresholdUnit: "percentage", FullScanCycle: "3", Incremental: true, FullScansScheduled: true, Preset: "SAP_JS_Default", TeamID: "16", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true}
	workspace, err := ioutil.TempDir("", "workspace2")
	if err != nil {
		t.Fatal("Failed to create temporary workspace directory")
	}
	// clean up tmp dir
	defer os.RemoveAll(workspace)

	influx := checkmarxExecuteScanInflux{}

	err = runScan(options, sys, workspace, &influx)
	assert.NoError(t, err, "error occured but none expected")
	assert.Equal(t, true, sys.isIncremental, "isIncremental has wrong value")
	assert.Equal(t, true, sys.isPublic, "isPublic has wrong value")
	assert.Equal(t, true, sys.forceScan, "forceScan has wrong value")
}

func TestRunScanForPullRequest(t *testing.T) {
	sys := &systemMock{response: []byte(`<?xml version="1.0" encoding="utf-8"?><CxXMLResults />`)}
	options := checkmarxExecuteScanOptions{PullRequestName: "PR-19", ProjectName: "Test", VulnerabilityThresholdUnit: "percentage", FullScanCycle: "3", Incremental: true, FullScansScheduled: true, Preset: "SAP_JS_Default", TeamID: "16", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true, AvoidDuplicateProjectScans: false}
	workspace, err := ioutil.TempDir("", "workspace3")
	if err != nil {
		t.Fatal("Failed to create temporary workspace directory")
	}
	// clean up tmp dir
	defer os.RemoveAll(workspace)

	influx := checkmarxExecuteScanInflux{}

	err = runScan(options, sys, workspace, &influx)
	assert.Equal(t, true, sys.isIncremental, "isIncremental has wrong value")
	assert.Equal(t, true, sys.isPublic, "isPublic has wrong value")
	assert.Equal(t, true, sys.forceScan, "forceScan has wrong value")
}

func TestRunScanForPullRequestProjectNew(t *testing.T) {
	sys := &systemMock{response: []byte(`<?xml version="1.0" encoding="utf-8"?><CxXMLResults />`), createProject: true}
	options := checkmarxExecuteScanOptions{PullRequestName: "PR-17", ProjectName: "Test", AvoidDuplicateProjectScans: true, VulnerabilityThresholdUnit: "percentage", FullScanCycle: "3", Incremental: true, FullScansScheduled: true, Preset: "10048", TeamName: "OpenSource/Cracks/15", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true}
	workspace, err := ioutil.TempDir("", "workspace4")
	if err != nil {
		t.Fatal("Failed to create temporary workspace directory")
	}
	// clean up tmp dir
	defer os.RemoveAll(workspace)

	influx := checkmarxExecuteScanInflux{}

	err = runScan(options, sys, workspace, &influx)
	assert.NoError(t, err, "Unexpected error caught")
	assert.Equal(t, true, sys.isIncremental, "isIncremental has wrong value")
	assert.Equal(t, true, sys.isPublic, "isPublic has wrong value")
	assert.Equal(t, false, sys.forceScan, "forceScan has wrong value")
}

func TestRunScanHighViolationPercentage(t *testing.T) {
	sys := &systemMock{response: []byte(`<?xml version="1.0" encoding="utf-8"?>
	<CxXMLResults InitiatorName="admin" Owner="admin" ScanId="1000005" ProjectId="2" ProjectName="Project 1" TeamFullPathOnReportDate="CxServer" DeepLink="http://WIN2K12-TEMP/CxWebClient/ViewerMain.aspx?scanid=1000005&amp;projectid=2" ScanStart="Sunday, December 3, 2017 4:50:34 PM" Preset="Checkmarx Default" ScanTime="00h:03m:18s" LinesOfCodeScanned="6838" FilesScanned="34" ReportCreationTime="Sunday, December 3, 2017 6:13:45 PM" Team="CxServer" CheckmarxVersion="8.6.0" ScanComments="" ScanType="Incremental" SourceOrigin="LocalPath" Visibility="Public">
	<Query id="430" categories="PCI DSS v3.2;PCI DSS (3.2) - 6.5.1 - Injection flaws - particularly SQL injection,OWASP Top 10 2013;A1-Injection,FISMA 2014;System And Information Integrity,NIST SP 800-53;SI-10 Information Input Validation (P1),OWASP Top 10 2017;A1-Injection" cweId="89" name="SQL_Injection" group="CSharp_High_Risk" Severity="High" Language="CSharp" LanguageHash="1363215419077432" LanguageChangeDate="2017-12-03T00:00:00.0000000" SeverityIndex="3" QueryPath="CSharp\Cx\CSharp High Risk\SQL Injection Version:0" QueryVersionCode="430">
		<Result NodeId="10000050002" FileName="bookstore/Login.cs" Status="Recurrent" Line="179" Column="103" FalsePositive="False" Severity="High" AssignToUser="" state="0" Remark="" DeepLink="http://WIN2K12-TEMP/CxWebClient/ViewerMain.aspx?scanid=1000005&amp;projectid=2&amp;pathid=2" SeverityIndex="3">
			<Path ResultId="1000005" PathId="2" SimilarityId="1765812516"/>
		</Result>
		<Result NodeId="10000050003" FileName="bookstore/Login.cs" Status="Recurrent" Line="180" Column="10" FalsePositive="False" Severity="High" AssignToUser="" state="0" Remark="" DeepLink="http://WIN2K12-TEMP/CxWebClient/ViewerMain.aspx?scanid=1000005&amp;projectid=2&amp;pathid=2" SeverityIndex="3">
			<Path ResultId="1000005" PathId="2" SimilarityId="1765812516"/>
		</Result>
		<Result NodeId="10000050004" FileName="bookstore/Login.cs" Status="Recurrent" Line="181" Column="190" FalsePositive="True" Severity="Medium" AssignToUser="" state="0" Remark="" DeepLink="http://WIN2K12-TEMP/CxWebClient/ViewerMain.aspx?scanid=1000005&amp;projectid=2&amp;pathid=2" SeverityIndex="2">
			<Path ResultId="1000005" PathId="2" SimilarityId="1765812516"/>
		</Result>
		<Result NodeId="10000050005" FileName="bookstore/Login.cs" Status="Recurrent" Line="181" Column="190" FalsePositive="True" Severity="Low" AssignToUser="" state="0" Remark="" DeepLink="http://WIN2K12-TEMP/CxWebClient/ViewerMain.aspx?scanid=1000005&amp;projectid=2&amp;pathid=2" SeverityIndex="2">
			<Path ResultId="1000005" PathId="2" SimilarityId="1765812516"/>
		</Result>
		<Result NodeId="10000050006" FileName="bookstore/Login.cs" Status="Recurrent" Line="181" Column="190" FalsePositive="True" Severity="Low" AssignToUser="" state="0" Remark="" DeepLink="http://WIN2K12-TEMP/CxWebClient/ViewerMain.aspx?scanid=1000005&amp;projectid=2&amp;pathid=2" SeverityIndex="2">
			<Path ResultId="1000005" PathId="2" SimilarityId="1765812516"/>
		</Result>
	</Query>
	</CxXMLResults>`)}
	options := checkmarxExecuteScanOptions{VulnerabilityThresholdUnit: "percentage", VulnerabilityThresholdResult: "FAILURE", VulnerabilityThresholdHigh: 100, FullScanCycle: "10", FullScansScheduled: true, Preset: "10048", TeamID: "16", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true}
	workspace, err := ioutil.TempDir("", "workspace5")
	if err != nil {
		t.Fatal("Failed to create temporary workspace directory")
	}
	// clean up tmp dir
	defer os.RemoveAll(workspace)

	influx := checkmarxExecuteScanInflux{}

	err = runScan(options, sys, workspace, &influx)
	assert.Contains(t, fmt.Sprint(err), "the project is not compliant", "Expected different error")
}

func TestRunScanHighViolationAbsolute(t *testing.T) {
	sys := &systemMock{response: []byte(`<?xml version="1.0" encoding="utf-8"?>
		<CxXMLResults InitiatorName="admin" Owner="admin" ScanId="1000005" ProjectId="2" ProjectName="Project 1" TeamFullPathOnReportDate="CxServer" DeepLink="http://WIN2K12-TEMP/CxWebClient/ViewerMain.aspx?scanid=1000005&amp;projectid=2" ScanStart="Sunday, December 3, 2017 4:50:34 PM" Preset="Checkmarx Default" ScanTime="00h:03m:18s" LinesOfCodeScanned="6838" FilesScanned="34" ReportCreationTime="Sunday, December 3, 2017 6:13:45 PM" Team="CxServer" CheckmarxVersion="8.6.0" ScanComments="" ScanType="Incremental" SourceOrigin="LocalPath" Visibility="Public">
		<Query id="430" categories="PCI DSS v3.2;PCI DSS (3.2) - 6.5.1 - Injection flaws - particularly SQL injection,OWASP Top 10 2013;A1-Injection,FISMA 2014;System And Information Integrity,NIST SP 800-53;SI-10 Information Input Validation (P1),OWASP Top 10 2017;A1-Injection" cweId="89" name="SQL_Injection" group="CSharp_High_Risk" Severity="High" Language="CSharp" LanguageHash="1363215419077432" LanguageChangeDate="2017-12-03T00:00:00.0000000" SeverityIndex="3" QueryPath="CSharp\Cx\CSharp High Risk\SQL Injection Version:0" QueryVersionCode="430">
			<Result NodeId="10000050002" FileName="bookstore/Login.cs" Status="Recurrent" Line="179" Column="103" FalsePositive="True" Severity="High" AssignToUser="" state="0" Remark="" DeepLink="http://WIN2K12-TEMP/CxWebClient/ViewerMain.aspx?scanid=1000005&amp;projectid=2&amp;pathid=2" SeverityIndex="3">
				<Path ResultId="1000005" PathId="2" SimilarityId="1765812516"/>
			</Result>
			<Result NodeId="10000050003" FileName="bookstore/Login.cs" Status="Recurrent" Line="180" Column="10" FalsePositive="True" Severity="High" AssignToUser="" state="0" Remark="" DeepLink="http://WIN2K12-TEMP/CxWebClient/ViewerMain.aspx?scanid=1000005&amp;projectid=2&amp;pathid=2" SeverityIndex="3">
				<Path ResultId="1000005" PathId="2" SimilarityId="1765812516"/>
			</Result>
			<Result NodeId="10000050004" FileName="bookstore/Login.cs" Status="Recurrent" Line="181" Column="190" FalsePositive="True" Severity="Medium" AssignToUser="" state="0" Remark="" DeepLink="http://WIN2K12-TEMP/CxWebClient/ViewerMain.aspx?scanid=1000005&amp;projectid=2&amp;pathid=2" SeverityIndex="2">
				<Path ResultId="1000005" PathId="2" SimilarityId="1765812516"/>
			</Result>
			<Result NodeId="10000050005" FileName="bookstore/Login.cs" Status="Recurrent" Line="181" Column="190" FalsePositive="False" Severity="Low" AssignToUser="" state="0" Remark="" DeepLink="http://WIN2K12-TEMP/CxWebClient/ViewerMain.aspx?scanid=1000005&amp;projectid=2&amp;pathid=2" SeverityIndex="2">
				<Path ResultId="1000005" PathId="2" SimilarityId="1765812516"/>
			</Result>
			<Result NodeId="10000050006" FileName="bookstore/Login.cs" Status="Recurrent" Line="181" Column="190" FalsePositive="False" Severity="Low" AssignToUser="" state="0" Remark="" DeepLink="http://WIN2K12-TEMP/CxWebClient/ViewerMain.aspx?scanid=1000005&amp;projectid=2&amp;pathid=2" SeverityIndex="2">
				<Path ResultId="1000005" PathId="2" SimilarityId="1765812516"/>
			</Result>
		</Query>
		</CxXMLResults>`)}
	options := checkmarxExecuteScanOptions{VulnerabilityThresholdUnit: "absolute", VulnerabilityThresholdResult: "FAILURE", VulnerabilityThresholdLow: 1, FullScanCycle: "10", FullScansScheduled: true, Preset: "10048", TeamID: "16", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true}
	workspace, err := ioutil.TempDir("", "workspace6")
	if err != nil {
		t.Fatal("Failed to create temporary workspace directory")
	}
	// clean up tmp dir
	defer os.RemoveAll(workspace)

	influx := checkmarxExecuteScanInflux{}

	err = runScan(options, sys, workspace, &influx)
	assert.Contains(t, fmt.Sprint(err), "the project is not compliant", "Expected different error")
}

func TestEnforceThresholds(t *testing.T) {
	results := map[string]interface{}{}
	results["High"] = map[string]int{}
	results["Medium"] = map[string]int{}
	results["Low"] = map[string]int{}

	results["High"].(map[string]int)["NotFalsePositive"] = 10
	results["Medium"].(map[string]int)["NotFalsePositive"] = 10
	results["Low"].(map[string]int)["NotFalsePositive"] = 10
	results["Low"].(map[string]int)["NotExploitable"] = 0
	results["Low"].(map[string]int)["Confirmed"] = 0

	results["High"].(map[string]int)["Issues"] = 10
	results["Medium"].(map[string]int)["Issues"] = 10
	results["Low"].(map[string]int)["Issues"] = 10

	t.Run("percentage high violation", func(t *testing.T) {
		options := checkmarxExecuteScanOptions{VulnerabilityThresholdUnit: "percentage", VulnerabilityThresholdHigh: 100, VulnerabilityThresholdEnabled: true}
		insecure := enforceThresholds(options, results)

		assert.Equal(t, true, insecure, "Expected results to be insecure but where not")
	})

	t.Run("absolute high violation", func(t *testing.T) {
		options := checkmarxExecuteScanOptions{VulnerabilityThresholdUnit: "absolute", VulnerabilityThresholdHigh: 5, VulnerabilityThresholdEnabled: true}
		insecure := enforceThresholds(options, results)

		assert.Equal(t, true, insecure, "Expected results to be insecure but where not")
	})

	t.Run("percentage medium violation", func(t *testing.T) {
		options := checkmarxExecuteScanOptions{VulnerabilityThresholdUnit: "percentage", VulnerabilityThresholdMedium: 100, VulnerabilityThresholdEnabled: true}
		insecure := enforceThresholds(options, results)

		assert.Equal(t, true, insecure, "Expected results to be insecure but where not")
	})

	t.Run("absolute medium violation", func(t *testing.T) {
		options := checkmarxExecuteScanOptions{VulnerabilityThresholdUnit: "absolute", VulnerabilityThresholdMedium: 5, VulnerabilityThresholdEnabled: true}
		insecure := enforceThresholds(options, results)

		assert.Equal(t, true, insecure, "Expected results to be insecure but where not")
	})

	t.Run("percentage low violation", func(t *testing.T) {
		options := checkmarxExecuteScanOptions{VulnerabilityThresholdUnit: "percentage", VulnerabilityThresholdLow: 100, VulnerabilityThresholdEnabled: true}
		insecure := enforceThresholds(options, results)

		assert.Equal(t, true, insecure, "Expected results to be insecure but where not")
	})

	t.Run("absolute low violation", func(t *testing.T) {
		options := checkmarxExecuteScanOptions{VulnerabilityThresholdUnit: "absolute", VulnerabilityThresholdLow: 5, VulnerabilityThresholdEnabled: true}
		insecure := enforceThresholds(options, results)

		assert.Equal(t, true, insecure, "Expected results to be insecure but where not")
	})

	t.Run("percentage no violation", func(t *testing.T) {
		options := checkmarxExecuteScanOptions{VulnerabilityThresholdUnit: "percentage", VulnerabilityThresholdLow: 0, VulnerabilityThresholdEnabled: true}
		insecure := enforceThresholds(options, results)

		assert.Equal(t, false, insecure, "Expected results to be insecure but where not")
	})

	t.Run("absolute no violation", func(t *testing.T) {
		options := checkmarxExecuteScanOptions{VulnerabilityThresholdUnit: "absolute", VulnerabilityThresholdLow: 15, VulnerabilityThresholdMedium: 15, VulnerabilityThresholdHigh: 15, VulnerabilityThresholdEnabled: true}
		insecure := enforceThresholds(options, results)

		assert.Equal(t, false, insecure, "Expected results to be insecure but where not")
	})
}

func TestLoadPreset(t *testing.T) {
	sys := &systemMock{}

	t.Run("resolve via name", func(t *testing.T) {
		preset, err := loadPreset(sys, "SAP_JS_Default")
		assert.NoError(t, err, "Expected success but failed")
		assert.Equal(t, "SAP_JS_Default", preset.Name, "Expected result but got none")
	})

	t.Run("error case", func(t *testing.T) {
		preset, err := loadPreset(sys, "")
		assert.Contains(t, fmt.Sprint(err), "preset SAP_JS_Default not found", "Expected different error")
		assert.Equal(t, 0, preset.ID, "Expected result but got none")
	})
}
