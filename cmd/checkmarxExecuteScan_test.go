package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
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
	response      interface{}
	isIncremental bool
	isPublic      bool
	forceScan     bool
}

func (sys *systemMock) GetPresetByName(presets []checkmarx.Preset, presetName string) checkmarx.Preset {
	return checkmarx.Preset{ID: 10050, Name: "SAP_JS_Default", OwnerName: "16"}
}
func (sys *systemMock) GetPresetByID(presets []checkmarx.Preset, presetID int) checkmarx.Preset {
	return checkmarx.Preset{ID: 10048, Name: "SAP_Default", OwnerName: "16"}
}
func (sys *systemMock) GetProjectByName(projects []checkmarx.Project, projectName string) checkmarx.Project {
	return checkmarx.Project{ID: 1, Name: "Test", TeamID: "16", IsPublic: false}
}
func (sys *systemMock) GetProjectByID(projectID int) (bool, checkmarx.Project) {
	return true, checkmarx.Project{ID: 19, Name: "Test_PR-19", TeamID: "16", IsPublic: false}
}
func (sys *systemMock) GetTeamByName(teams []checkmarx.Team, teamName string) checkmarx.Team {
	return checkmarx.Team{ID: "16", FullName: "OpenSource/Cracks/16"}
}
func (sys *systemMock) GetTeamByID(teams []checkmarx.Team, teamID string) checkmarx.Team {
	return checkmarx.Team{ID: "15", FullName: "OpenSource/Cracks/15"}
}
func (sys *systemMock) DownloadReport(reportID int) (bool, []byte) {
	return true, sys.response.([]byte)
}
func (sys *systemMock) GetReportStatus(reportID int) checkmarx.ReportStatusResponse {
	return checkmarx.ReportStatusResponse{Status: checkmarx.ReportStatus{ID: 2, Value: "Created"}}
}
func (sys *systemMock) RequestNewReport(scanID int, reportType string) (bool, checkmarx.Report) {
	return true, checkmarx.Report{ReportID: 17}
}
func (sys *systemMock) GetResults(scanID int) checkmarx.ResultsStatistics {
	return checkmarx.ResultsStatistics{}
}
func (sys *systemMock) GetScans(projectID int) (bool, []checkmarx.ScanStatus) {
	return true, []checkmarx.ScanStatus{checkmarx.ScanStatus{IsIncremental: true}, checkmarx.ScanStatus{IsIncremental: true}, checkmarx.ScanStatus{IsIncremental: true}, checkmarx.ScanStatus{IsIncremental: false}}
}
func (sys *systemMock) GetScanStatus(scanID int) string {
	return "Finished"
}
func (sys *systemMock) ScanProject(projectID int, isIncrementalV, isPublicV, forceScanV bool) (bool, checkmarx.Scan) {
	sys.isIncremental = isIncrementalV
	sys.isPublic = isPublicV
	sys.forceScan = forceScanV
	return true, checkmarx.Scan{ID: 16}
}
func (sys *systemMock) UpdateProjectConfiguration(projectID int, presetID int, engineConfigurationID string) bool {
	return true
}
func (sys *systemMock) UpdateProjectExcludeSettings(projectID int, excludeFolders string, excludeFiles string) bool {
	return true
}
func (sys *systemMock) UploadProjectSourceCode(projectID int, zipFile string) bool {
	return true
}
func (sys *systemMock) CreateProject(projectName string, teamID string) (bool, checkmarx.ProjectCreateResult) {
	return true, checkmarx.ProjectCreateResult{ID: 20}
}
func (sys *systemMock) CreateBranch(projectID int, branchName string) int {
	return 18
}
func (sys *systemMock) GetPresets() []checkmarx.Preset {
	return []checkmarx.Preset{checkmarx.Preset{ID: 10078, Name: "SAP Java Default", OwnerName: "16"}, checkmarx.Preset{ID: 10048, Name: "SAP JS Default", OwnerName: "16"}}
}
func (sys *systemMock) GetProjects(teamID string) []checkmarx.Project {
	return []checkmarx.Project{checkmarx.Project{ID: 15, Name: "OtherTest", TeamID: teamID}, checkmarx.Project{ID: 1, Name: "Test", TeamID: teamID}}
}
func (sys *systemMock) GetTeams() []checkmarx.Team {
	return []checkmarx.Team{checkmarx.Team{ID: "16", FullName: "OpenSource/Cracks/16"}, checkmarx.Team{ID: "15", FullName: "OpenSource/Cracks/15"}}
}

type systemMockForExistingProject struct {
	response      interface{}
	isIncremental bool
	isPublic      bool
	forceScan     bool
}

func (sys *systemMockForExistingProject) GetPresetByName(presets []checkmarx.Preset, presetName string) checkmarx.Preset {
	return checkmarx.Preset{ID: 10050, Name: "SAP_JS_Default", OwnerName: "16"}
}
func (sys *systemMockForExistingProject) GetPresetByID(presets []checkmarx.Preset, presetID int) checkmarx.Preset {
	return checkmarx.Preset{ID: 10048, Name: "SAP_Default", OwnerName: "16"}
}
func (sys *systemMockForExistingProject) GetProjectByName(projects []checkmarx.Project, projectName string) checkmarx.Project {
	return checkmarx.Project{ID: 1, Name: "TestExisting", TeamID: "16", IsPublic: false}
}
func (sys *systemMockForExistingProject) GetProjectByID(projectID int) (bool, checkmarx.Project) {
	return false, checkmarx.Project{}
}
func (sys *systemMockForExistingProject) GetTeamByName(teams []checkmarx.Team, teamName string) checkmarx.Team {
	return checkmarx.Team{ID: "16", FullName: "OpenSource/Cracks/16"}
}
func (sys *systemMockForExistingProject) GetTeamByID(teams []checkmarx.Team, teamID string) checkmarx.Team {
	return checkmarx.Team{ID: "15", FullName: "OpenSource/Cracks/15"}
}
func (sys *systemMockForExistingProject) DownloadReport(reportID int) (bool, []byte) {
	return true, sys.response.([]byte)
}
func (sys *systemMockForExistingProject) GetReportStatus(reportID int) checkmarx.ReportStatusResponse {
	return checkmarx.ReportStatusResponse{Status: checkmarx.ReportStatus{ID: 2, Value: "Created"}}
}
func (sys *systemMockForExistingProject) RequestNewReport(scanID int, reportType string) (bool, checkmarx.Report) {
	return true, checkmarx.Report{ReportID: 17}
}
func (sys *systemMockForExistingProject) GetResults(scanID int) checkmarx.ResultsStatistics {
	return checkmarx.ResultsStatistics{}
}
func (sys *systemMockForExistingProject) GetScans(projectID int) (bool, []checkmarx.ScanStatus) {
	return true, []checkmarx.ScanStatus{checkmarx.ScanStatus{IsIncremental: true}, checkmarx.ScanStatus{IsIncremental: true}, checkmarx.ScanStatus{IsIncremental: true}, checkmarx.ScanStatus{IsIncremental: false}}
}
func (sys *systemMockForExistingProject) GetScanStatus(scanID int) string {
	return "Finished"
}
func (sys *systemMockForExistingProject) ScanProject(projectID int, isIncrementalV, isPublicV, forceScanV bool) (bool, checkmarx.Scan) {
	sys.isIncremental = isIncrementalV
	sys.isPublic = isPublicV
	sys.forceScan = forceScanV
	return true, checkmarx.Scan{ID: 16}
}
func (sys *systemMockForExistingProject) UpdateProjectConfiguration(projectID int, presetID int, engineConfigurationID string) bool {
	return true
}
func (sys *systemMockForExistingProject) UpdateProjectExcludeSettings(projectID int, excludeFolders string, excludeFiles string) bool {
	return true
}
func (sys *systemMockForExistingProject) UploadProjectSourceCode(projectID int, zipFile string) bool {
	return true
}
func (sys *systemMockForExistingProject) CreateProject(projectName string, teamID string) (bool, checkmarx.ProjectCreateResult) {
	return false, checkmarx.ProjectCreateResult{}
}
func (sys *systemMockForExistingProject) CreateBranch(projectID int, branchName string) int {
	return 0
}
func (sys *systemMockForExistingProject) GetPresets() []checkmarx.Preset {
	return []checkmarx.Preset{checkmarx.Preset{ID: 10078, Name: "SAP Java Default", OwnerName: "16"}, checkmarx.Preset{ID: 10048, Name: "SAP JS Default", OwnerName: "16"}}
}
func (sys *systemMockForExistingProject) GetProjects(teamID string) []checkmarx.Project {
	return []checkmarx.Project{checkmarx.Project{ID: 1, Name: "TestExisting", TeamID: teamID}}
}
func (sys *systemMockForExistingProject) GetTeams() []checkmarx.Team {
	return []checkmarx.Team{checkmarx.Team{ID: "16", FullName: "OpenSource/Cracks/16"}, checkmarx.Team{ID: "15", FullName: "OpenSource/Cracks/15"}}
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
		dir, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatal("Failed to create temporary directory")
		}
		// clean up tmp dir
		defer os.RemoveAll(dir)

		ioutil.WriteFile(filepath.Join(dir, "checkmarxExecuteScan.go"), []byte("cxnjckdsjhkjfz8izr"), 0700)
		ioutil.WriteFile(filepath.Join(dir, "somepath", "checkmarxExecuteScan.txt"), []byte("0987654321"), 0700)
		ioutil.WriteFile(filepath.Join(dir, "checkmarxExecuteScan_test.go"), []byte("123456789"), 0700)

		var zipFileMock bytes.Buffer
		zipFolder(dir, &zipFileMock, []string{"!checkmarxExecuteScan_test.go", "**/*.txt", "**/checkmarxExecuteScan.go"})

		got := zipFileMock.Len()
		want := 136

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
		result := getDetailedResults(sys, 2635)
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
	options := checkmarxExecuteScanOptions{CheckmarxProject: "TestExisting", VulnerabilityThresholdUnit: "absolute", FullScanCycle: "2", Incremental: true, FullScansScheduled: true, Preset: "10048", CheckmarxGroupID: "16", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true}
	workspace, err := ioutil.TempDir("", "workspace")
	if err != nil {
		t.Fatal("Failed to create temporary workspace directory")
	}
	// clean up tmp dir
	defer os.RemoveAll(workspace)

	influx := checkmarxExecuteScanInflux{}

	err = runScan(options, sys, workspace, &influx)
	assert.NoError(t, err, "Unexpected error detected")
	assert.Equal(t, false, sys.isIncremental, "isIncremental has wrong value")
	assert.Equal(t, false, sys.isPublic, "isPublic has wrong value")
	assert.Equal(t, false, sys.forceScan, "forceScan has wrong value")
}

func TestRunScanWOtherCycle(t *testing.T) {
	sys := &systemMock{response: []byte(`<?xml version="1.0" encoding="utf-8"?><CxXMLResults />`)}
	options := checkmarxExecuteScanOptions{VulnerabilityThresholdUnit: "percentage", FullScanCycle: "3", Incremental: true, FullScansScheduled: true, Preset: "SAP_JS_Default", CheckmarxGroupID: "16", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true}
	workspace, err := ioutil.TempDir("", "workspace")
	if err != nil {
		t.Fatal("Failed to create temporary workspace directory")
	}
	// clean up tmp dir
	defer os.RemoveAll(workspace)

	influx := checkmarxExecuteScanInflux{}

	err = runScan(options, sys, workspace, &influx)
	assert.NoError(t, err, "Unexpected error detected")
	assert.Equal(t, true, sys.isIncremental, "isIncremental has wrong value")
	assert.Equal(t, false, sys.isPublic, "isPublic has wrong value")
	assert.Equal(t, false, sys.forceScan, "forceScan has wrong value")
}

func TestRunScanForPullRequest(t *testing.T) {
	sys := &systemMock{response: []byte(`<?xml version="1.0" encoding="utf-8"?><CxXMLResults />`)}
	options := checkmarxExecuteScanOptions{PullRequestName: "Test_PR-19", VulnerabilityThresholdUnit: "percentage", FullScanCycle: "3", Incremental: true, FullScansScheduled: true, Preset: "SAP_JS_Default", CheckmarxGroupID: "16", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true}
	workspace, err := ioutil.TempDir("", "workspace")
	if err != nil {
		t.Fatal("Failed to create temporary workspace directory")
	}
	// clean up tmp dir
	defer os.RemoveAll(workspace)

	influx := checkmarxExecuteScanInflux{}

	err = runScan(options, sys, workspace, &influx)
	assert.NoError(t, err, "Unexpected error detected")
	assert.Equal(t, true, sys.isIncremental, "isIncremental has wrong value")
	assert.Equal(t, false, sys.isPublic, "isPublic has wrong value")
	assert.Equal(t, false, sys.forceScan, "forceScan has wrong value")
}

func TestRunScanForPullRequestProjectNew(t *testing.T) {
	sys := &systemMock{response: []byte(`<?xml version="1.0" encoding="utf-8"?><CxXMLResults />`)}
	options := checkmarxExecuteScanOptions{PullRequestName: "OtherTest_PR-17", VulnerabilityThresholdUnit: "percentage", FullScanCycle: "3", Incremental: true, FullScansScheduled: true, Preset: "SAP_JS_Default", TeamName: "OpenSource/Cracks/15", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true}
	workspace, err := ioutil.TempDir("", "workspace")
	if err != nil {
		t.Fatal("Failed to create temporary workspace directory")
	}
	// clean up tmp dir
	defer os.RemoveAll(workspace)

	influx := checkmarxExecuteScanInflux{}

	err = runScan(options, sys, workspace, &influx)
	assert.NoError(t, err, "Unexpected error detected")
	assert.Equal(t, true, sys.isIncremental, "isIncremental has wrong value")
	assert.Equal(t, false, sys.isPublic, "isPublic has wrong value")
	assert.Equal(t, false, sys.forceScan, "forceScan has wrong value")
}

func TestRunScanHighViolationPercentage(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
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
		options := checkmarxExecuteScanOptions{VulnerabilityThresholdUnit: "percentage", VulnerabilityThresholdResult: "FAILURE", VulnerabilityThresholdHigh: "100", FullScanCycle: "10", FullScansScheduled: true, Preset: "10048", CheckmarxGroupID: "16", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true}
		workspace, err := ioutil.TempDir("", "workspace")
		if err != nil {
			t.Fatal("Failed to create temporary workspace directory")
		}
		// clean up tmp dir
		defer os.RemoveAll(workspace)

		influx := checkmarxExecuteScanInflux{}

		runScan(options, sys, workspace, &influx)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestRunScanHighViolationPercentage")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestRunScanHighViolationAbsolute(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
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
		options := checkmarxExecuteScanOptions{VulnerabilityThresholdUnit: "absolute", VulnerabilityThresholdResult: "FAILURE", VulnerabilityThresholdLow: "1", FullScanCycle: "10", FullScansScheduled: true, Preset: "10048", CheckmarxGroupID: "16", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true}
		workspace, err := ioutil.TempDir("", "workspace")
		if err != nil {
			t.Fatal("Failed to create temporary workspace directory")
		}
		// clean up tmp dir
		defer os.RemoveAll(workspace)

		influx := checkmarxExecuteScanInflux{}

		runScan(options, sys, workspace, &influx)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestRunScanHighViolationAbsolute")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}
