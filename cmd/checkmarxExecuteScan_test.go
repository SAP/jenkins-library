//go:build unit
// +build unit

package cmd

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/SAP/jenkins-library/pkg/checkmarx"
	"github.com/stretchr/testify/assert"

	"github.com/google/go-github/v68/github"
)

type fileInfo struct {
	nam     string      // base name of the file
	siz     int64       // length in bytes for regular files; system-dependent for others
	mod     os.FileMode // file mode bits
	modtime time.Time   // modification time
	dir     bool        // abbreviation for Mode().IsDir()
	syss    interface{} // underlying data source (can return nil)
}

func (f *fileInfo) IsDir() bool {
	return f.dir
}

func (f *fileInfo) Name() string {
	return f.nam
}

func (f *fileInfo) Size() int64 {
	return f.siz
}

func (f *fileInfo) ModTime() time.Time {
	return f.modtime
}

func (f *fileInfo) Mode() os.FileMode {
	return f.mod
}

func (f *fileInfo) Sys() interface{} {
	return f.syss
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

func (sys *systemMock) FilterPresetByName(_ []checkmarx.Preset, presetName string) checkmarx.Preset {
	if presetName == "CX_Default" {
		return checkmarx.Preset{ID: 16, Name: "CX_Default", OwnerName: "16"}
	}
	return checkmarx.Preset{ID: 10050, Name: "SAP_JS_Default", OwnerName: "16"}
}

func (sys *systemMock) FilterPresetByID([]checkmarx.Preset, int) checkmarx.Preset {
	return checkmarx.Preset{ID: 10048, Name: "SAP_Default", OwnerName: "16"}
}

func (sys *systemMock) FilterProjectByName([]checkmarx.Project, string) checkmarx.Project {
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
	if strings.EqualFold(projectName, "Test") {
		return []checkmarx.Project{{ID: 1, Name: projectName, TeamID: teamID, IsPublic: true}}, nil
	}
	sys.previousPName = projectName
	return []checkmarx.Project{}, fmt.Errorf("no project error")
}

func (sys *systemMock) FilterTeamByName(_ []checkmarx.Team, teamName string) (checkmarx.Team, error) {
	if teamName == "OpenSource/Cracks/16" {
		return checkmarx.Team{ID: json.RawMessage(`"16"`), FullName: "OpenSource/Cracks/16"}, nil
	}
	return checkmarx.Team{ID: json.RawMessage(`15`), FullName: "OpenSource/Cracks/15"}, nil
}

func (sys *systemMock) FilterTeamByID(_ []checkmarx.Team, teamID json.RawMessage) checkmarx.Team {
	teamIDBytes, _ := teamID.MarshalJSON()
	if bytes.Equal(teamIDBytes, []byte(`"16"`)) {
		return checkmarx.Team{ID: json.RawMessage(`"16"`), FullName: "OpenSource/Cracks/16"}
	}
	return checkmarx.Team{ID: json.RawMessage(`15`), FullName: "OpenSource/Cracks/15"}
}

func (sys *systemMock) DownloadReport(int) ([]byte, error) {
	return sys.response.([]byte), nil
}

func (sys *systemMock) GetReportStatus(int) (checkmarx.ReportStatusResponse, error) {
	return checkmarx.ReportStatusResponse{Status: checkmarx.ReportStatus{ID: 2, Value: "Created"}}, nil
}

func (sys *systemMock) RequestNewReport(int, string) (checkmarx.Report, error) {
	return checkmarx.Report{ReportID: 17}, nil
}

func (sys *systemMock) GetResults(int) checkmarx.ResultsStatistics {
	return checkmarx.ResultsStatistics{}
}

func (sys *systemMock) GetScans(int) ([]checkmarx.ScanStatus, error) {
	return []checkmarx.ScanStatus{{IsIncremental: true}, {IsIncremental: true}, {IsIncremental: true}, {IsIncremental: false}}, nil
}

func (sys *systemMock) GetScanStatusAndDetail(int) (string, checkmarx.ScanStatusDetail) {
	return "Finished", checkmarx.ScanStatusDetail{Stage: "Step 1 of 25", Step: "Scan something"}
}

func (sys *systemMock) ScanProject(_ int, isIncrementalV, isPublicV, forceScanV bool) (checkmarx.Scan, error) {
	sys.isIncremental = isIncrementalV
	sys.isPublic = isPublicV
	sys.forceScan = forceScanV
	return checkmarx.Scan{ID: 16}, nil
}

func (sys *systemMock) UpdateProjectConfiguration(int, int, string) error {
	sys.updateProjectConfigurationCalled = true
	return nil
}

func (sys *systemMock) UpdateProjectExcludeSettings(int, string, string) error {
	return nil
}

func (sys *systemMock) UploadProjectSourceCode(int, string) error {
	return nil
}

func (sys *systemMock) CreateProject(string, string) (checkmarx.ProjectCreateResult, error) {
	return checkmarx.ProjectCreateResult{ID: 20}, nil
}

func (sys *systemMock) CreateBranch(int, string) int {
	return 18
}

func (sys *systemMock) GetShortDescription(int, int) (checkmarx.ShortDescription, error) {
	return checkmarx.ShortDescription{Text: "dummyText"}, nil
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

func (sys *systemMockForExistingProject) FilterPresetByName([]checkmarx.Preset, string) checkmarx.Preset {
	return checkmarx.Preset{ID: 10050, Name: "SAP_JS_Default", OwnerName: "16"}
}

func (sys *systemMockForExistingProject) FilterPresetByID([]checkmarx.Preset, int) checkmarx.Preset {
	return checkmarx.Preset{ID: 10048, Name: "SAP_Default", OwnerName: "16"}
}

func (sys *systemMockForExistingProject) FilterProjectByName([]checkmarx.Project, string) checkmarx.Project {
	return checkmarx.Project{ID: 1, Name: "TestExisting", TeamID: "16", IsPublic: true}
}

func (sys *systemMockForExistingProject) GetProjectByID(int) (checkmarx.Project, error) {
	return checkmarx.Project{}, nil
}

func (sys *systemMockForExistingProject) GetProjectsByNameAndTeam(projectName, teamID string) ([]checkmarx.Project, error) {
	return []checkmarx.Project{{ID: 19, Name: projectName, TeamID: teamID, IsPublic: true}}, nil
}

func (sys *systemMockForExistingProject) FilterTeamByName([]checkmarx.Team, string) (checkmarx.Team, error) {
	return checkmarx.Team{ID: json.RawMessage(`"16"`), FullName: "OpenSource/Cracks/16"}, nil
}

func (sys *systemMockForExistingProject) FilterTeamByID([]checkmarx.Team, json.RawMessage) checkmarx.Team {
	return checkmarx.Team{ID: json.RawMessage(`"15"`), FullName: "OpenSource/Cracks/15"}
}

func (sys *systemMockForExistingProject) DownloadReport(int) ([]byte, error) {
	return sys.response.([]byte), nil
}

func (sys *systemMockForExistingProject) GetReportStatus(int) (checkmarx.ReportStatusResponse, error) {
	return checkmarx.ReportStatusResponse{Status: checkmarx.ReportStatus{ID: 2, Value: "Created"}}, nil
}

func (sys *systemMockForExistingProject) RequestNewReport(int, string) (checkmarx.Report, error) {
	return checkmarx.Report{ReportID: 17}, nil
}

func (sys *systemMockForExistingProject) GetResults(int) checkmarx.ResultsStatistics {
	return checkmarx.ResultsStatistics{}
}

func (sys *systemMockForExistingProject) GetScans(int) ([]checkmarx.ScanStatus, error) {
	return []checkmarx.ScanStatus{{IsIncremental: true}, {IsIncremental: true}, {IsIncremental: true}, {IsIncremental: false}}, nil
}

func (sys *systemMockForExistingProject) GetShortDescription(int, int) (checkmarx.ShortDescription, error) {
	return checkmarx.ShortDescription{Text: "dummyText"}, nil
}

func (sys *systemMockForExistingProject) GetScanStatusAndDetail(int) (string, checkmarx.ScanStatusDetail) {
	return "Finished", checkmarx.ScanStatusDetail{Stage: "", Step: ""}
}

func (sys *systemMockForExistingProject) ScanProject(_ int, isIncrementalV, isPublicV, forceScanV bool) (checkmarx.Scan, error) {
	sys.scanProjectCalled = true
	sys.isIncremental = isIncrementalV
	sys.isPublic = isPublicV
	sys.forceScan = forceScanV
	return checkmarx.Scan{ID: 16}, nil
}

func (sys *systemMockForExistingProject) UpdateProjectConfiguration(int, int, string) error {
	return nil
}

func (sys *systemMockForExistingProject) UpdateProjectExcludeSettings(int, string, string) error {
	return nil
}

func (sys *systemMockForExistingProject) UploadProjectSourceCode(int, string) error {
	return nil
}

func (sys *systemMockForExistingProject) CreateProject(string, string) (checkmarx.ProjectCreateResult, error) {
	return checkmarx.ProjectCreateResult{}, fmt.Errorf("create project error")
}

func (sys *systemMockForExistingProject) CreateBranch(int, string) int {
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

type checkmarxExecuteScanUtilsMock struct {
	errorOnFileInfoHeader bool
	errorOnStat           bool
	errorOnOpen           bool
	errorOnWriteFile      bool
	errorOnPathMatch      bool
	workspace             string
}

func newCheckmarxExecuteScanUtilsMock() *checkmarxExecuteScanUtilsMock {
	return &checkmarxExecuteScanUtilsMock{}
}

func (c *checkmarxExecuteScanUtilsMock) GetWorkspace() string {
	return c.workspace
}

func (c *checkmarxExecuteScanUtilsMock) PathMatch(pattern, name string) (bool, error) {
	if c.errorOnPathMatch {
		return false, fmt.Errorf("error on PathMatch")
	}
	return doublestar.PathMatch(pattern, name)
}

func (c *checkmarxExecuteScanUtilsMock) WriteFile(filename string, data []byte, perm os.FileMode) error {
	if c.errorOnWriteFile {
		return fmt.Errorf("error on WriteFile")
	}
	return os.WriteFile(filename, data, perm)
}

func (c *checkmarxExecuteScanUtilsMock) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (c *checkmarxExecuteScanUtilsMock) FileInfoHeader(fi os.FileInfo) (*zip.FileHeader, error) {
	if c.errorOnFileInfoHeader {
		return nil, fmt.Errorf("error on FileInfoHeader")
	}
	return zip.FileInfoHeader(fi)
}

func (c *checkmarxExecuteScanUtilsMock) Stat(name string) (os.FileInfo, error) {
	if c.errorOnStat {
		return nil, fmt.Errorf("error on Stat")
	}
	return os.Stat(name)
}

func (c *checkmarxExecuteScanUtilsMock) Open(name string) (*os.File, error) {
	if c.errorOnOpen {
		return nil, fmt.Errorf("error on Open")
	}
	return os.Open(name)
}

func (c *checkmarxExecuteScanUtilsMock) GetIssueService() *github.IssuesService {
	return nil
}

func (c *checkmarxExecuteScanUtilsMock) GetSearchService() *github.SearchService {
	return nil
}

func TestFilterFileGlob(t *testing.T) {
	t.Parallel()
	tt := []struct {
		input    string
		fInfo    fileInfo
		expected bool
	}{
		{input: filepath.Join("somepath", "node_modules", "someOther", "some.file"), fInfo: fileInfo{}, expected: true},
		{input: filepath.Join("somepath", "non_modules", "someOther", "some.go"), fInfo: fileInfo{}, expected: false},
		{input: filepath.Join(".xmake", "someOther", "some.go"), fInfo: fileInfo{}, expected: true},
		{input: filepath.Join("another", "vendor", "some.html"), fInfo: fileInfo{}, expected: false},
		{input: filepath.Join("another", "vendor", "some.pdf"), fInfo: fileInfo{}, expected: true},
		{input: filepath.Join("another", "vendor", "some.test"), fInfo: fileInfo{}, expected: true},
		{input: filepath.Join("some.test"), fInfo: fileInfo{}, expected: false},
		{input: filepath.Join("a", "b", "c"), fInfo: fileInfo{dir: true}, expected: false},
	}

	for k, v := range tt {
		result, err := isFileNotMatchingPattern([]string{"!**/node_modules/**", "!**/.xmake/**", "!**/*_test.go", "!**/vendor/**/*.go", "**/*.go", "**/*.html", "*.test"}, v.input, &v.fInfo, newCheckmarxExecuteScanUtilsMock())
		assert.Equal(t, v.expected, result, fmt.Sprintf("wrong result for run %v", k))
		assert.NoError(t, err, "no error expected in run %v", k)
	}
}

func TestFilterFileGlob_errorOnPathMatch(t *testing.T) {
	t.Parallel()

	utilsMock := newCheckmarxExecuteScanUtilsMock()
	utilsMock.errorOnPathMatch = true

	result, err := isFileNotMatchingPattern([]string{"!**/node_modules/**", "!**/.xmake/**", "!**/*_test.go", "!**/vendor/**/*.go", "**/*.go", "**/*.html", "*.test"}, filepath.Join("a", "b", "c"), &fileInfo{}, utilsMock)
	assert.Equal(t, false, result, "wrong result")
	assert.EqualError(t, err, "Pattern **/node_modules/** could not get executed: error on PathMatch")
}

func TestZipFolder(t *testing.T) {
	t.Parallel()

	t.Run("zip files successfully", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		err := os.WriteFile(filepath.Join(dir, "abcd.go"), []byte("abcd.go"), 0o700)
		assert.NoError(t, err)
		err = os.WriteFile(filepath.Join(dir, "abcd.yaml"), []byte("abcd.yaml"), 0o700)
		assert.NoError(t, err)
		err = os.Mkdir(filepath.Join(dir, "somepath"), 0o700)
		assert.NoError(t, err)
		err = os.WriteFile(filepath.Join(dir, "somepath", "abcd.txt"), []byte("somepath/abcd.txt"), 0o700)
		assert.NoError(t, err)
		err = os.WriteFile(filepath.Join(dir, "abcd_test.go"), []byte("abcd_test.go"), 0o700)
		assert.NoError(t, err)
		err = os.WriteFile(filepath.Join(dir, "abc_test.go"), []byte("abc_test.go"), 0o700)
		assert.NoError(t, err)

		var zipFileMock bytes.Buffer
		err = zipFolder(dir, &zipFileMock, []string{"!**/abc_test.txt", "!**/abcd.yaml", "**/abcd.*"}, newCheckmarxExecuteScanUtilsMock())
		assert.NoError(t, err)

		zipString := zipFileMock.String()

		// assert.Equal(t, 724, zipFileMock.Len(), "Expected length of 724, but got %v", zipFileMock.Len())
		assert.True(t, strings.Contains(zipString, "abcd.go"), "Expected 'abcd.go' contained")
		assert.True(t, strings.Contains(zipString, filepath.Join("somepath", "abcd.txt")), "Expected 'somepath/abcd.txt' contained")
		assert.False(t, strings.Contains(zipString, "abcd_test.go"), "Not expected 'abcd_test.go' contained")
		assert.False(t, strings.Contains(zipString, "abcd.yaml"), "Not expected 'abcd.yaml' contained")
		assert.False(t, strings.Contains(zipString, "abc_test.go"), "Not expected 'abc_test.go' contained")
	})

	t.Run("error on query file info header", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		err := os.WriteFile(filepath.Join(dir, "abcd.go"), []byte("abcd.go"), 0o700)
		assert.NoError(t, err)
		err = os.Mkdir(filepath.Join(dir, "somepath"), 0o700)
		assert.NoError(t, err)
		err = os.WriteFile(filepath.Join(dir, "somepath", "abcd.txt"), []byte("somepath/abcd.txt"), 0o700)
		assert.NoError(t, err)
		err = os.WriteFile(filepath.Join(dir, "abcd_test.go"), []byte("abcd_test.go"), 0o700)
		assert.NoError(t, err)
		err = os.WriteFile(filepath.Join(dir, "abc_test.go"), []byte("abc_test.go"), 0o700)
		assert.NoError(t, err)

		var zipFileMock bytes.Buffer
		mock := newCheckmarxExecuteScanUtilsMock()
		mock.errorOnFileInfoHeader = true
		err = zipFolder(dir, &zipFileMock, []string{"!**/abc_test.txt", "**/abcd.*"}, mock)

		assert.EqualError(t, err, "error on FileInfoHeader")
	})

	t.Run("error on os stat", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		err := os.WriteFile(filepath.Join(dir, "abcd.go"), []byte("abcd.go"), 0o700)
		assert.NoError(t, err)
		err = os.Mkdir(filepath.Join(dir, "somepath"), 0o700)
		assert.NoError(t, err)
		err = os.WriteFile(filepath.Join(dir, "somepath", "abcd.txt"), []byte("somepath/abcd.txt"), 0o700)
		assert.NoError(t, err)
		err = os.WriteFile(filepath.Join(dir, "abcd_test.go"), []byte("abcd_test.go"), 0o700)
		assert.NoError(t, err)
		err = os.WriteFile(filepath.Join(dir, "abc_test.go"), []byte("abc_test.go"), 0o700)
		assert.NoError(t, err)

		var zipFileMock bytes.Buffer
		mock := newCheckmarxExecuteScanUtilsMock()
		mock.errorOnStat = true
		err = zipFolder(dir, &zipFileMock, []string{"!**/abc_test.txt", "**/abcd.*"}, mock)

		assert.NoError(t, err)
	})

	t.Run("error on os Open", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		err := os.WriteFile(filepath.Join(dir, "abcd.go"), []byte("abcd.go"), 0o700)
		assert.NoError(t, err)
		err = os.Mkdir(filepath.Join(dir, "somepath"), 0o700)
		assert.NoError(t, err)
		err = os.WriteFile(filepath.Join(dir, "somepath", "abcd.txt"), []byte("somepath/abcd.txt"), 0o700)
		assert.NoError(t, err)
		err = os.WriteFile(filepath.Join(dir, "abcd_test.go"), []byte("abcd_test.go"), 0o700)
		assert.NoError(t, err)
		err = os.WriteFile(filepath.Join(dir, "abc_test.go"), []byte("abc_test.go"), 0o700)
		assert.NoError(t, err)

		var zipFileMock bytes.Buffer
		mock := newCheckmarxExecuteScanUtilsMock()
		mock.errorOnOpen = true
		err = zipFolder(dir, &zipFileMock, []string{"!**/abc_test.txt", "**/abcd.*"}, mock)

		assert.EqualError(t, err, "error on Open")
	})
}

func TestGetDetailedResults(t *testing.T) {
	t.Parallel()

	t.Run("success case", func(t *testing.T) {
		t.Parallel()
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

		<Query id="430" categories="PCI DSS v3.2;PCI DSS (3.2) - 6.5.1 - Injection flaws - particularly SQL injection,OWASP Top 10 2013;A1-Injection,FISMA 2014;System And Information Integrity,NIST SP 800-53;SI-10 Information Input Validation (P1),OWASP Top 10 2017;A1-Injection" cweId="89" name="Code_Injection" group="CSharp_High_Risk" Severity="High" Language="CSharp" LanguageHash="1363215419077432" LanguageChangeDate="2017-12-03T00:00:00.0000000" SeverityIndex="3" QueryPath="CSharp\Cx\CSharp High Risk\SQL Injection Version:0" QueryVersionCode="430">
			<Result NodeId="10000050005" FileName="bookstore/Login.cs" Status="Recurrent" Line="181" Column="190" FalsePositive="True" Severity="Low" AssignToUser="" state="1" Remark="" DeepLink="http://WIN2K12-TEMP/CxWebClient/ViewerMain.aspx?scanid=1000005&amp;projectid=2&amp;pathid=2" SeverityIndex="2">
				<Path ResultId="1000005" PathId="2" SimilarityId="1765812516"/>
			</Result>
			<Result NodeId="10000050006" FileName="bookstore/Login.cs" Status="Recurrent" Line="181" Column="190" FalsePositive="True" Severity="Low" AssignToUser="" state="2" Remark="" DeepLink="http://WIN2K12-TEMP/CxWebClient/ViewerMain.aspx?scanid=1000005&amp;projectid=2&amp;pathid=2" SeverityIndex="2">
				<Path ResultId="1000005" PathId="2" SimilarityId="1765812516"/>
			</Result>
		</Query>
		</CxXMLResults>`)}
		dir := t.TempDir()
		options := checkmarxExecuteScanOptions{VulnerabilityThresholdUnit: "percentage", VulnerabilityThresholdHigh: 100, VulnerabilityThresholdEnabled: true, VulnerabilityThresholdLowPerQuery: true, VulnerabilityThresholdLowPerQueryMax: 10, VulnerabilityThresholdLow: 10}
		result, err := getDetailedResults(options, sys, filepath.Join(dir, "abc.xml"), 2635, newCheckmarxExecuteScanUtilsMock())
		assert.NoError(t, err, "error occurred but none expected")
		assert.Equal(t, "2", result["ProjectId"], "Project ID incorrect")
		assert.Equal(t, "Project 1", result["ProjectName"], "Project name incorrect")
		assert.Equal(t, 2, result["High"].(map[string]int)["Issues"], "Number of High issues incorrect")
		assert.Equal(t, 2, result["High"].(map[string]int)["NotFalsePositive"], "Number of High NotFalsePositive issues incorrect")
		assert.Equal(t, 1, result["Medium"].(map[string]int)["Issues"], "Number of Medium issues incorrect")
		assert.Equal(t, 0, result["Medium"].(map[string]int)["NotFalsePositive"], "Number of Medium NotFalsePositive issues incorrect")
		assert.Equal(t, 4, result["Low"].(map[string]int)["Issues"], "Number of Low issues incorrect")
		assert.Equal(t, 0, result["Low"].(map[string]int)["NotFalsePositive"], "Number of Low NotFalsePositive issues incorrect")
		assert.Equal(t, 2, result["LowPerQuery"].(map[string]map[string]int)["SQL_Injection"]["Issues"], "Number of Low issues (per query 1) incorrect")
		assert.Equal(t, 1, result["LowPerQuery"].(map[string]map[string]int)["SQL_Injection"]["Urgent"], "Number of Low issues (per query 1) incorrect")
		assert.Equal(t, 1, result["LowPerQuery"].(map[string]map[string]int)["SQL_Injection"]["ProposedNotExploitable"], "Number of Low issues (per query 1) incorrect")
		assert.Equal(t, 2, result["LowPerQuery"].(map[string]map[string]int)["Code_Injection"]["Issues"], "Number of Low issues (per query 2) incorrect")
		assert.Equal(t, 1, result["LowPerQuery"].(map[string]map[string]int)["Code_Injection"]["Confirmed"], "Number of Low issues (per query 2) incorrect")
		assert.Equal(t, 1, result["LowPerQuery"].(map[string]map[string]int)["Code_Injection"]["NotExploitable"], "Number of Low issues (per query 2) incorrect")
	})

	t.Run("error on write file", func(t *testing.T) {
		t.Parallel()
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
		dir := t.TempDir()
		utils := newCheckmarxExecuteScanUtilsMock()
		utils.errorOnWriteFile = true
		options := checkmarxExecuteScanOptions{VulnerabilityThresholdUnit: "percentage", VulnerabilityThresholdHigh: 100, VulnerabilityThresholdEnabled: true, VulnerabilityThresholdLowPerQuery: true, VulnerabilityThresholdLowPerQueryMax: 10, VulnerabilityThresholdLow: 10}
		_, err := getDetailedResults(options, sys, filepath.Join(dir, "abc.xml"), 2635, utils)
		assert.EqualError(t, err, "failed to write file: error on WriteFile")
	})
}

func TestRunScan(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	sys := &systemMockForExistingProject{response: []byte(`<?xml version="1.0" encoding="utf-8"?><CxXMLResults />`)}
	options := checkmarxExecuteScanOptions{ProjectName: "TestExisting", VulnerabilityThresholdUnit: "absolute", FullScanCycle: "2", Incremental: true, FullScansScheduled: true, Preset: "10048", TeamID: "16", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true}
	workspace := t.TempDir()
	err := os.WriteFile(filepath.Join(workspace, "abcd.go"), []byte("abcd.go"), 0o700)
	assert.NoError(t, err)
	options.FilterPattern = "**/abcd.go"

	influx := checkmarxExecuteScanInflux{}

	utilsMock := newCheckmarxExecuteScanUtilsMock()
	utilsMock.workspace = workspace

	err = runScan(ctx, options, sys, &influx, utilsMock)
	assert.NoError(t, err, "error occurred but none expected")
	assert.Equal(t, false, sys.isIncremental, "isIncremental has wrong value")
	assert.Equal(t, true, sys.isPublic, "isPublic has wrong value")
	assert.Equal(t, true, sys.forceScan, "forceScan has wrong value")
	assert.Equal(t, true, sys.scanProjectCalled, "ScanProject was not invoked")
}

func TestRunScan_nonNumeralPreset(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	sys := &systemMockForExistingProject{response: []byte(`<?xml version="1.0" encoding="utf-8"?><CxXMLResults />`)}
	options := checkmarxExecuteScanOptions{ProjectName: "TestExisting", VulnerabilityThresholdUnit: "absolute", FullScanCycle: "2", Incremental: true, FullScansScheduled: true, Preset: "SAP_JS_Default", TeamID: "16", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true}
	workspace := t.TempDir()
	err := os.WriteFile(filepath.Join(workspace, "abcd.go"), []byte("abcd.go"), 0o700)
	assert.NoError(t, err)
	options.FilterPattern = "**/abcd.go"

	influx := checkmarxExecuteScanInflux{}

	utilsMock := newCheckmarxExecuteScanUtilsMock()
	utilsMock.workspace = workspace

	err = runScan(ctx, options, sys, &influx, utilsMock)
	assert.NoError(t, err, "error occurred but none expected")
}

func TestRunOptimizedScan(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	sys := &systemMockForExistingProject{response: []byte(`<?xml version="1.0" encoding="utf-8"?><CxXMLResults />`)}
	options := checkmarxExecuteScanOptions{IsOptimizedAndScheduled: true, ProjectName: "TestExisting", VulnerabilityThresholdUnit: "absolute", FullScanCycle: "1", Incremental: true, FullScansScheduled: true, Preset: "10048", TeamID: "16", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true}
	workspace := t.TempDir()
	err := os.WriteFile(filepath.Join(workspace, "abcd.go"), []byte("abcd.go"), 0o700)
	assert.NoError(t, err)
	options.FilterPattern = "**/abcd.go"

	influx := checkmarxExecuteScanInflux{}

	utilsMock := newCheckmarxExecuteScanUtilsMock()
	utilsMock.workspace = workspace

	err = runScan(ctx, options, sys, &influx, utilsMock)
	assert.NoError(t, err, "error occurred but none expected")
	assert.Equal(t, false, sys.isIncremental, "isIncremental has wrong value")
	assert.Equal(t, true, sys.isPublic, "isPublic has wrong value")
	assert.Equal(t, true, sys.forceScan, "forceScan has wrong value")
	assert.Equal(t, true, sys.scanProjectCalled, "ScanProject was not invoked")
}

func TestSetPresetForProjectWithIDProvided(t *testing.T) {
	t.Parallel()

	sys := &systemMock{}
	err := setPresetForProject(sys, 12345, 16, "testProject", "CX_Default", "")
	assert.NoError(t, err, "error occurred but none expected")
	assert.Equal(t, false, sys.getPresetsCalled, "GetPresets was called")
	assert.Equal(t, true, sys.updateProjectConfigurationCalled, "UpdateProjectConfiguration was not called")
}

func TestSetPresetForProjectWithNameProvided(t *testing.T) {
	t.Parallel()

	sys := &systemMock{}
	presetID, _ := strconv.Atoi("CX_Default")
	err := setPresetForProject(sys, 12345, presetID, "testProject", "CX_Default", "")
	assert.NoError(t, err, "error occurred but none expected")
	assert.Equal(t, true, sys.getPresetsCalled, "GetPresets was not called")
	assert.Equal(t, true, sys.updateProjectConfigurationCalled, "UpdateProjectConfiguration was not called")
}

func TestVerifyOnly(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	sys := &systemMockForExistingProject{response: []byte(`<?xml version="1.0" encoding="utf-8"?><CxXMLResults />`)}
	options := checkmarxExecuteScanOptions{VerifyOnly: true, ProjectName: "TestExisting", VulnerabilityThresholdUnit: "absolute", FullScanCycle: "2", Incremental: true, FullScansScheduled: true, Preset: "10048", TeamName: "OpenSource/Cracks/15", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true}
	workspace := t.TempDir()

	influx := checkmarxExecuteScanInflux{}

	utilsMock := newCheckmarxExecuteScanUtilsMock()
	utilsMock.workspace = workspace

	err := runScan(ctx, options, sys, &influx, utilsMock)
	assert.NoError(t, err, "error occurred but none expected")
	assert.Equal(t, false, sys.scanProjectCalled, "ScanProject was invoked but shouldn't")
}

func TestVerifyOnly_errorOnWriteFileDoesNotBlock(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	sys := &systemMockForExistingProject{response: []byte(`<?xml version="1.0" encoding="utf-8"?><CxXMLResults />`)}
	options := checkmarxExecuteScanOptions{VerifyOnly: true, ProjectName: "TestExisting", VulnerabilityThresholdUnit: "absolute", FullScanCycle: "2", Incremental: true, FullScansScheduled: true, Preset: "10048", TeamName: "OpenSource/Cracks/15", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true}
	workspace := t.TempDir()

	influx := checkmarxExecuteScanInflux{}

	utilsMock := newCheckmarxExecuteScanUtilsMock()
	utilsMock.workspace = workspace
	utilsMock.errorOnWriteFile = true

	err := runScan(ctx, options, sys, &influx, utilsMock)
	assert.EqualError(t, err, "scan, upload, and result validation returned an error: project TestExisting not compliant: failed to get detailed results: failed to write file: error on WriteFile")
}

func TestRunScanWOtherCycle(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	sys := &systemMock{response: []byte(`<?xml version="1.0" encoding="utf-8"?><CxXMLResults />`), createProject: true}
	options := checkmarxExecuteScanOptions{ProjectName: "test", VulnerabilityThresholdUnit: "percentage", FullScanCycle: "3", Incremental: true, FullScansScheduled: true, Preset: "123", TeamID: "16", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true}
	workspace := t.TempDir()
	err := os.WriteFile(filepath.Join(workspace, "abcd.go"), []byte("abcd.go"), 0o700)
	assert.NoError(t, err)
	options.FilterPattern = "**/abcd.go"

	influx := checkmarxExecuteScanInflux{}

	utilsMock := newCheckmarxExecuteScanUtilsMock()
	utilsMock.workspace = workspace

	err = runScan(ctx, options, sys, &influx, utilsMock)
	assert.NoError(t, err, "error occurred but none expected")
	assert.Equal(t, true, sys.isIncremental, "isIncremental has wrong value")
	assert.Equal(t, true, sys.isPublic, "isPublic has wrong value")
	assert.Equal(t, true, sys.forceScan, "forceScan has wrong value")
}

func TestRunScanErrorInZip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	sys := &systemMock{response: []byte(`<?xml version="1.0" encoding="utf-8"?><CxXMLResults />`), createProject: true}
	options := checkmarxExecuteScanOptions{ProjectName: "test", VulnerabilityThresholdUnit: "percentage", FullScanCycle: "3", Incremental: true, FullScansScheduled: true, Preset: "123", TeamID: "16", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true}
	workspace := t.TempDir()

	influx := checkmarxExecuteScanInflux{}

	utilsMock := newCheckmarxExecuteScanUtilsMock()
	utilsMock.workspace = workspace
	utilsMock.errorOnFileInfoHeader = true

	err := runScan(ctx, options, sys, &influx, utilsMock)
	assert.EqualError(t, err, "scan, upload, and result validation returned an error: failed to zip workspace files: failed to compact folder: error on FileInfoHeader")
}

func TestRunScanForPullRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	sys := &systemMock{response: []byte(`<?xml version="1.0" encoding="utf-8"?><CxXMLResults />`)}
	options := checkmarxExecuteScanOptions{PullRequestName: "PR-19", ProjectName: "Test", VulnerabilityThresholdUnit: "percentage", FullScanCycle: "3", Incremental: true, FullScansScheduled: true, Preset: "123", TeamID: "16", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true, AvoidDuplicateProjectScans: false}
	workspace := t.TempDir()
	err := os.WriteFile(filepath.Join(workspace, "abcd.go"), []byte("abcd.go"), 0o700)
	assert.NoError(t, err)
	options.FilterPattern = "**/abcd.go"

	influx := checkmarxExecuteScanInflux{}

	utilsMock := newCheckmarxExecuteScanUtilsMock()
	utilsMock.workspace = workspace

	err = runScan(ctx, options, sys, &influx, utilsMock)
	assert.Equal(t, true, sys.isIncremental, "isIncremental has wrong value")
	assert.Equal(t, true, sys.isPublic, "isPublic has wrong value")
	assert.Equal(t, true, sys.forceScan, "forceScan has wrong value")
}

func TestRunScanForPullRequestProjectNew(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	sys := &systemMock{response: []byte(`<?xml version="1.0" encoding="utf-8"?><CxXMLResults />`), createProject: true}
	options := checkmarxExecuteScanOptions{PullRequestName: "PR-17", ProjectName: "Test", AvoidDuplicateProjectScans: true, VulnerabilityThresholdUnit: "percentage", FullScanCycle: "3", Incremental: true, FullScansScheduled: true, Preset: "10048", TeamName: "OpenSource/Cracks/15", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true}
	workspace := t.TempDir()
	err := os.WriteFile(filepath.Join(workspace, "abcd.go"), []byte("abcd.go"), 0o700)
	assert.NoError(t, err)
	options.FilterPattern = "**/abcd.go"

	influx := checkmarxExecuteScanInflux{}

	utilsMock := newCheckmarxExecuteScanUtilsMock()
	utilsMock.workspace = workspace

	err = runScan(ctx, options, sys, &influx, utilsMock)
	assert.NoError(t, err, "Unexpected error caught")
	assert.Equal(t, true, sys.isIncremental, "isIncremental has wrong value")
	assert.Equal(t, true, sys.isPublic, "isPublic has wrong value")
	assert.Equal(t, false, sys.forceScan, "forceScan has wrong value")
}

func TestRunScanForPullRequestProjectNew_nonNumeralPreset(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	sys := &systemMock{response: []byte(`<?xml version="1.0" encoding="utf-8"?><CxXMLResults />`), createProject: true}
	options := checkmarxExecuteScanOptions{PullRequestName: "PR-17", ProjectName: "Test", AvoidDuplicateProjectScans: true, VulnerabilityThresholdUnit: "percentage", FullScanCycle: "3", Incremental: true, FullScansScheduled: true, Preset: "SAP_JS_Default", TeamName: "OpenSource/Cracks/15", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true}
	workspace := t.TempDir()
	err := os.WriteFile(filepath.Join(workspace, "abcd.go"), []byte("abcd.go"), 0o700)
	assert.NoError(t, err)
	options.FilterPattern = "**/abcd.go"

	influx := checkmarxExecuteScanInflux{}

	utilsMock := newCheckmarxExecuteScanUtilsMock()
	utilsMock.workspace = workspace

	err = runScan(ctx, options, sys, &influx, utilsMock)
	assert.NoError(t, err, "error occurred but none expected")
}

func TestRunScanHighViolationPercentage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

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
	options := checkmarxExecuteScanOptions{ProjectName: "test", VulnerabilityThresholdUnit: "percentage", VulnerabilityThresholdResult: "FAILURE", VulnerabilityThresholdHigh: 100, FullScanCycle: "10", FullScansScheduled: true, Preset: "10048", TeamID: "16", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true}
	workspace := t.TempDir()
	err := os.WriteFile(filepath.Join(workspace, "abcd.go"), []byte("abcd.go"), 0o700)
	assert.NoError(t, err)
	options.FilterPattern = "**/abcd.go"

	influx := checkmarxExecuteScanInflux{}

	utilsMock := newCheckmarxExecuteScanUtilsMock()
	utilsMock.workspace = workspace

	err = runScan(ctx, options, sys, &influx, utilsMock)
	assert.Contains(t, fmt.Sprint(err), "the project is not compliant", "Expected different error")
}

func TestRunScanHighViolationAbsolute(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

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
	options := checkmarxExecuteScanOptions{ProjectName: "test", VulnerabilityThresholdUnit: "absolute", VulnerabilityThresholdResult: "FAILURE", VulnerabilityThresholdLow: 1, FullScanCycle: "10", FullScansScheduled: true, Preset: "10048", TeamID: "16", VulnerabilityThresholdEnabled: true, GeneratePdfReport: true}
	workspace := t.TempDir()
	err := os.WriteFile(filepath.Join(workspace, "abcd.go"), []byte("abcd.go"), 0o700)
	assert.NoError(t, err)
	options.FilterPattern = "**/abcd.go"

	influx := checkmarxExecuteScanInflux{}

	utilsMock := newCheckmarxExecuteScanUtilsMock()
	utilsMock.workspace = workspace

	err = runScan(ctx, options, sys, &influx, utilsMock)
	assert.Contains(t, fmt.Sprint(err), "the project is not compliant", "Expected different error")
}

func TestEnforceThresholds(t *testing.T) {
	t.Parallel()

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

	lowPerQuery := map[string]map[string]int{}
	submap := map[string]int{}
	submap["Issues"] = 8
	submap["Confirmed"] = 1
	submap["NotExploitable"] = 0
	lowPerQuery["Low_Query_Name_1"] = submap
	submap = map[string]int{}
	submap["Issues"] = 100
	submap["Confirmed"] = 5
	submap["NotExploitable"] = 5
	lowPerQuery["Low_Query_Name_2"] = submap
	results["LowPerQuery"] = lowPerQuery

	t.Run("percentage low violation per query", func(t *testing.T) {
		t.Parallel()

		options := checkmarxExecuteScanOptions{VulnerabilityThresholdUnit: "percentage", VulnerabilityThresholdHigh: 0, VulnerabilityThresholdMedium: 0, VulnerabilityThresholdLow: 20, VulnerabilityThresholdEnabled: true, VulnerabilityThresholdLowPerQuery: true, VulnerabilityThresholdLowPerQueryMax: 10}
		insecure, _, _ := enforceThresholds(options, results)

		assert.Equal(t, true, insecure, "Expected results to be insecure but where not")
	})

	t.Run("percentage low no violation per query", func(t *testing.T) {
		t.Parallel()

		options := checkmarxExecuteScanOptions{VulnerabilityThresholdUnit: "percentage", VulnerabilityThresholdHigh: 0, VulnerabilityThresholdMedium: 0, VulnerabilityThresholdLow: 10, VulnerabilityThresholdEnabled: true, VulnerabilityThresholdLowPerQuery: true, VulnerabilityThresholdLowPerQueryMax: 10}
		insecure, _, _ := enforceThresholds(options, results)

		assert.Equal(t, false, insecure, "Expected results to be insecure but where not")
	})

	t.Run("percentage high violation", func(t *testing.T) {
		t.Parallel()

		options := checkmarxExecuteScanOptions{VulnerabilityThresholdUnit: "percentage", VulnerabilityThresholdHigh: 100, VulnerabilityThresholdEnabled: true}
		insecure, insecureResults, neutralResults := enforceThresholds(options, results)

		assert.Equal(t, true, insecure, "Expected results to be insecure but where not")
		assert.Equal(t, 1, len(insecureResults), fmt.Sprintf("Unexpected number of results: %v", insecureResults))
		assert.Equal(t, 2, len(neutralResults), fmt.Sprintf("Unexpected number of results: %v", neutralResults))
	})

	t.Run("absolute high violation", func(t *testing.T) {
		t.Parallel()

		options := checkmarxExecuteScanOptions{VulnerabilityThresholdUnit: "absolute", VulnerabilityThresholdHigh: 5, VulnerabilityThresholdEnabled: true}
		insecure, insecureResults, neutralResults := enforceThresholds(options, results)

		assert.Equal(t, true, insecure, "Expected results to be insecure but where not")
		assert.Equal(t, 3, len(insecureResults), fmt.Sprintf("Unexpected number of results: %v", insecureResults))
		assert.Equal(t, 0, len(neutralResults), fmt.Sprintf("Unexpected number of results: %v", neutralResults))
	})

	t.Run("percentage medium violation", func(t *testing.T) {
		t.Parallel()

		options := checkmarxExecuteScanOptions{VulnerabilityThresholdUnit: "percentage", VulnerabilityThresholdMedium: 100, VulnerabilityThresholdEnabled: true}
		insecure, insecureResults, neutralResults := enforceThresholds(options, results)

		assert.Equal(t, true, insecure, "Expected results to be insecure but where not")
		assert.Equal(t, 1, len(insecureResults), fmt.Sprintf("Unexpected number of results: %v", insecureResults))
		assert.Equal(t, 2, len(neutralResults), fmt.Sprintf("Unexpected number of results: %v", neutralResults))
	})

	t.Run("absolute medium violation", func(t *testing.T) {
		t.Parallel()

		options := checkmarxExecuteScanOptions{VulnerabilityThresholdUnit: "absolute", VulnerabilityThresholdMedium: 5, VulnerabilityThresholdEnabled: true}
		insecure, insecureResults, neutralResults := enforceThresholds(options, results)

		assert.Equal(t, true, insecure, "Expected results to be insecure but where not")
		assert.Equal(t, 3, len(insecureResults), fmt.Sprintf("Unexpected number of results: %v", insecureResults))
		assert.Equal(t, 0, len(neutralResults), fmt.Sprintf("Unexpected number of results: %v", neutralResults))
	})

	t.Run("percentage low violation", func(t *testing.T) {
		t.Parallel()

		options := checkmarxExecuteScanOptions{VulnerabilityThresholdUnit: "percentage", VulnerabilityThresholdLow: 100, VulnerabilityThresholdEnabled: true}
		insecure, insecureResults, neutralResults := enforceThresholds(options, results)

		assert.Equal(t, true, insecure, "Expected results to be insecure but where not")
		assert.Equal(t, 1, len(insecureResults), fmt.Sprintf("Unexpected number of results: %v", insecureResults))
		assert.Equal(t, 2, len(neutralResults), fmt.Sprintf("Unexpected number of results: %v", neutralResults))
	})

	t.Run("absolute low violation", func(t *testing.T) {
		t.Parallel()

		options := checkmarxExecuteScanOptions{VulnerabilityThresholdUnit: "absolute", VulnerabilityThresholdLow: 5, VulnerabilityThresholdEnabled: true}
		insecure, insecureResults, neutralResults := enforceThresholds(options, results)

		assert.Equal(t, true, insecure, "Expected results to be insecure but where not")
		assert.Equal(t, 3, len(insecureResults), fmt.Sprintf("Unexpected number of results: %v", insecureResults))
		assert.Equal(t, 0, len(neutralResults), fmt.Sprintf("Unexpected number of results: %v", neutralResults))
	})

	t.Run("percentage no violation", func(t *testing.T) {
		t.Parallel()

		options := checkmarxExecuteScanOptions{VulnerabilityThresholdUnit: "percentage", VulnerabilityThresholdLow: 0, VulnerabilityThresholdEnabled: true}
		insecure, insecureResults, neutralResults := enforceThresholds(options, results)

		assert.Equal(t, false, insecure, "Expected results to be insecure but where not")
		assert.Equal(t, 0, len(insecureResults), fmt.Sprintf("Unexpected number of results: %v", insecureResults))
		assert.Equal(t, 3, len(neutralResults), fmt.Sprintf("Unexpected number of results: %v", neutralResults))
	})

	t.Run("absolute no violation", func(t *testing.T) {
		t.Parallel()

		options := checkmarxExecuteScanOptions{VulnerabilityThresholdUnit: "absolute", VulnerabilityThresholdLow: 15, VulnerabilityThresholdMedium: 15, VulnerabilityThresholdHigh: 15, VulnerabilityThresholdEnabled: true}
		insecure, insecureResults, neutralResults := enforceThresholds(options, results)

		assert.Equal(t, false, insecure, "Expected results to be insecure but where not")
		assert.Equal(t, 0, len(insecureResults), fmt.Sprintf("Unexpected number of results: %v", insecureResults))
		assert.Equal(t, 3, len(neutralResults), fmt.Sprintf("Unexpected number of results: %v", neutralResults))
	})
}

func TestLoadPreset(t *testing.T) {
	t.Parallel()

	sys := &systemMock{}

	t.Run("resolve via name", func(t *testing.T) {
		t.Parallel()

		preset, err := loadPreset(sys, "SAP_JS_Default")
		assert.NoError(t, err, "Expected success but failed")
		assert.Equal(t, "SAP_JS_Default", preset.Name, "Expected result but got none")
	})

	t.Run("error case", func(t *testing.T) {
		t.Parallel()

		preset, err := loadPreset(sys, "")
		assert.Contains(t, fmt.Sprint(err), "preset SAP_JS_Default not found", "Expected different error")
		assert.Equal(t, 0, preset.ID, "Expected result but got none")
	})
}

func TestPreventInfiniteLoop(t *testing.T) {
	t.Parallel()

	utils := checkmarxExecuteScanUtilsBundle{
		workspace: "abc",
	}

	assert.Equal(t, "abc", utils.GetWorkspace(), "Wrong workspace has been loaded")
}
