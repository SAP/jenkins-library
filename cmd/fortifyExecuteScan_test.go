//go:build unit

package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/mock"

	"github.com/SAP/jenkins-library/pkg/fortify"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/versioning"

	"github.com/google/go-github/v68/github"
	"github.com/stretchr/testify/assert"

	"github.com/piper-validation/fortify-client-go/models"
)

const author string = "johnDoe178"

type fortifyTestUtilsBundle struct {
	*execRunnerMock
	*mock.FilesMock
	getArtifactShouldFail bool
}

func (f *fortifyTestUtilsBundle) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	panic("not expected to be called in tests")
}

func (f *fortifyTestUtilsBundle) GetArtifact(buildTool, buildDescriptorFile string, options *versioning.Options) (versioning.Artifact, error) {
	if f.getArtifactShouldFail {
		return nil, fmt.Errorf("build tool '%v' not supported", buildTool)
	}
	return artifactMock{Coordinates: newCoordinatesMock()}, nil
}

func (f *fortifyTestUtilsBundle) GetIssueService() *github.IssuesService {
	return nil
}

func (cf *fortifyTestUtilsBundle) GetSearchService() *github.SearchService {
	return nil
}

func newFortifyTestUtilsBundle() fortifyTestUtilsBundle {
	utilsBundle := fortifyTestUtilsBundle{
		execRunnerMock: &execRunnerMock{},
		FilesMock:      &mock.FilesMock{},
	}
	return utilsBundle
}
func mockExecinPath(exec string) (string, error) {
	executable_list := []string{"fortifyupdate", "sourceanalyzer"}
	for _, exec := range executable_list {
		if exec == "fortifyupdate" || exec == "sourceanalyzer" {
			return "/" + exec, nil
		} else {
			err_string := fmt.Sprintf("ERROR , command not found: %s. Please configure a supported docker image or install Fortify SCA on the system.", exec)
			return "", errors.New(err_string)
		}
	}
	return "", nil
}

func failMockExecinPathfortifyupdate(exec string) (string, error) {
	if exec == "fortifyupdate" {
		return "", errors.New("Command not found: fortifyupdate. Please configure a supported docker image or install Fortify SCA on the system.")
	}
	return "/fortifyupdate", nil
}
func failMockExecinPathsourceanalyzer(exec string) (string, error) {
	if exec == "sourceanalyzer" {
		return "", errors.New("Command not found: sourceanalyzer. Please configure a supported docker image or install Fortify SCA on the system.")
	}
	return "/sourceanalyzer", nil
}

type artifactMock struct {
	Coordinates versioning.Coordinates
}

func newCoordinatesMock() versioning.Coordinates {
	return versioning.Coordinates{
		GroupID:    "a",
		ArtifactID: "b",
		Version:    "1.0.0",
	}
}

func (a artifactMock) VersioningScheme() string {
	return "full"
}

func (a artifactMock) GetVersion() (string, error) {
	return a.Coordinates.Version, nil
}

func (a artifactMock) SetVersion(v string) error {
	a.Coordinates.Version = v
	return nil
}

func (a artifactMock) GetCoordinates() (versioning.Coordinates, error) {
	return a.Coordinates, nil
}

type fortifyMock struct {
	Successive                       bool
	getArtifactsOfProjectVersionIdx  int
	getArtifactsOfProjectVersionTime time.Time
}

func (f *fortifyMock) GetProjectByName(name string, autoCreate bool, projectVersion string) (*models.Project, error) {
	return &models.Project{Name: &name, ID: 64}, nil
}

func (f *fortifyMock) GetProjectVersionDetailsByProjectIDAndVersionName(id int64, name string, autoCreate bool, projectName string) (*models.ProjectVersion, error) {
	return &models.ProjectVersion{ID: id, Name: &name, Project: &models.Project{Name: &projectName}}, nil
}

func (f *fortifyMock) GetProjectVersionAttributesByProjectVersionID(id int64) ([]*models.Attribute, error) {
	return []*models.Attribute{}, nil
}

func (f *fortifyMock) SetProjectVersionAttributesByProjectVersionID(id int64, attributes []*models.Attribute) ([]*models.Attribute, error) {
	return attributes, nil
}

func (f *fortifyMock) CreateProjectVersionIfNotExist(projectName, projectVersionName, description string) (*models.ProjectVersion, error) {
	return &models.ProjectVersion{ID: 4711, Name: &projectVersionName, Project: &models.Project{Name: &projectName}}, nil
}

func (f *fortifyMock) LookupOrCreateProjectVersionDetailsForPullRequest(projectID int64, masterProjectVersion *models.ProjectVersion, pullRequestName string) (*models.ProjectVersion, error) {
	return &models.ProjectVersion{ID: 4712, Name: &pullRequestName, Project: masterProjectVersion.Project}, nil
}

func (f *fortifyMock) CreateProjectVersion(version *models.ProjectVersion) (*models.ProjectVersion, error) {
	return version, nil
}

func (f *fortifyMock) ProjectVersionCopyFromPartial(sourceID, targetID int64) error {
	return nil
}

func (f *fortifyMock) ProjectVersionCopyCurrentState(sourceID, targetID int64) error {
	return nil
}

func (f *fortifyMock) ProjectVersionCopyPermissions(sourceID, targetID int64) error {
	return nil
}

func (f *fortifyMock) CommitProjectVersion(id int64) (*models.ProjectVersion, error) {
	name := "Committed"
	return &models.ProjectVersion{ID: id, Name: &name}, nil
}

func (f *fortifyMock) MergeProjectVersionStateOfPRIntoMaster(downloadEndpoint, uploadEndpoint string, masterProjectID, masterProjectVersionID int64, pullRequestName string) error {
	return nil
}

func (f *fortifyMock) GetArtifactsOfProjectVersion(id int64) ([]*models.Artifact, error) {
	switch id {
	case 4711:
		return []*models.Artifact{{
			Status:     "PROCESSED",
			UploadDate: toFortifyTime(time.Now()),
		}}, nil
	case 4712:
		return []*models.Artifact{{
			Status:     "ERROR_PROCESSING",
			UploadDate: toFortifyTime(time.Now()),
		}}, nil
	case 4713:
		return []*models.Artifact{{
			Status:     "REQUIRE_AUTH",
			UploadDate: toFortifyTime(time.Now()),
		}}, nil
	case 4714:
		return []*models.Artifact{{
			Status:     "PROCESSING",
			UploadDate: toFortifyTime(time.Now()),
		}}, nil
	case 4715:
		return []*models.Artifact{{
			Status: "PROCESSED",
			Embed: &models.EmbeddedScans{
				Scans: []*models.Scan{{BuildLabel: "/commit/test"}},
			},
			UploadDate: toFortifyTime(time.Now()),
		}}, nil
	case 4716:
		var status string
		if f.getArtifactsOfProjectVersionIdx == 0 {
			f.getArtifactsOfProjectVersionTime = time.Now().Add(-2 * time.Minute)
		}
		if f.getArtifactsOfProjectVersionIdx < 2 {
			status = "PROCESSING"
		} else {
			f.getArtifactsOfProjectVersionTime = time.Now()
			status = "PROCESSED"
		}
		f.getArtifactsOfProjectVersionIdx++

		return []*models.Artifact{{
			Status:     status,
			UploadDate: toFortifyTime(f.getArtifactsOfProjectVersionTime),
		}}, nil
	case 4718:
		return []*models.Artifact{
			{
				Status:     "PROCESSED",
				UploadDate: toFortifyTime(time.Now()),
			},
			{
				Status:     "ERROR_PROCESSING",
				UploadDate: toFortifyTime(time.Now().Add(-2 * time.Minute)),
			},
		}, nil

	default:
		return []*models.Artifact{}, nil
	}
}

func (f *fortifyMock) GetFilterSetOfProjectVersionByTitle(id int64, title string) (*models.FilterSet, error) {
	return &models.FilterSet{}, nil
}

func (f *fortifyMock) GetIssueFilterSelectorOfProjectVersionByName(id int64, names []string, options []string) (*models.IssueFilterSelectorSet, error) {
	return &models.IssueFilterSelectorSet{}, nil
}

func (f *fortifyMock) GetFilterSetByDisplayName(issueFilterSelectorSet *models.IssueFilterSelectorSet, name string) *models.IssueFilterSelector {
	if issueFilterSelectorSet.FilterBySet != nil {
		for _, filter := range issueFilterSelectorSet.FilterBySet {
			if filter.DisplayName == name {
				return filter
			}
		}
	}
	return &models.IssueFilterSelector{DisplayName: name}
}

func (f *fortifyMock) GetProjectIssuesByIDAndFilterSetGroupedBySelector(id int64, filter, filterSetGUID string, issueFilterSelectorSet *models.IssueFilterSelectorSet) ([]*models.ProjectVersionIssueGroup, error) {
	if filter == "ET1:abcd" {
		group := "HTTP Verb tampering"
		total := int32(4)
		audited := int32(3)
		group2 := "Password in code"
		total2 := int32(4)
		audited2 := int32(4)
		group3 := "Memory leak"
		total3 := int32(5)
		audited3 := int32(4)
		return []*models.ProjectVersionIssueGroup{
			{ID: &group, TotalCount: &total, AuditedCount: &audited},
			{ID: &group2, TotalCount: &total2, AuditedCount: &audited2},
			{ID: &group3, TotalCount: &total3, AuditedCount: &audited3},
		}, nil
	}
	if issueFilterSelectorSet != nil && issueFilterSelectorSet.FilterBySet != nil && len(issueFilterSelectorSet.FilterBySet) > 0 && issueFilterSelectorSet.FilterBySet[0].GUID == "3" {
		groupName := "Suspicious"
		groupName2 := "Exploitable"
		group := "3"
		total := int32(4)
		audited := int32(0)
		group2 := "4"
		total2 := int32(5)
		audited2 := int32(0)
		return []*models.ProjectVersionIssueGroup{
			{ID: &group, CleanName: &groupName, TotalCount: &total, AuditedCount: &audited},
			{ID: &group2, CleanName: &groupName2, TotalCount: &total2, AuditedCount: &audited2},
		}, nil
	}
	group := "Audit All"
	total := int32(15)
	audited := int32(12)
	group2 := "Corporate Security Requirements"
	total2 := int32(20)
	audited2 := int32(11)
	group3 := "Spot Checks of Each Category"
	total3 := int32(5)
	audited3 := int32(4)
	return []*models.ProjectVersionIssueGroup{
		{ID: &group, CleanName: &group, TotalCount: &total, AuditedCount: &audited},
		{ID: &group2, CleanName: &group2, TotalCount: &total2, AuditedCount: &audited2},
		{ID: &group3, CleanName: &group3, TotalCount: &total3, AuditedCount: &audited3},
	}, nil
}

func (f *fortifyMock) ReduceIssueFilterSelectorSet(issueFilterSelectorSet *models.IssueFilterSelectorSet, names []string, options []string) *models.IssueFilterSelectorSet {
	return issueFilterSelectorSet
}

func (f *fortifyMock) GetIssueStatisticsOfProjectVersion(id int64) ([]*models.IssueStatistics, error) {
	suppressed := int32(6)
	return []*models.IssueStatistics{{SuppressedCount: &suppressed}}, nil
}

func (f *fortifyMock) GenerateQGateReport(projectID, projectVersionID, reportTemplateID int64, projectName, projectVersionName, reportFormat string) (*models.SavedReport, error) {
	if !f.Successive {
		f.Successive = true
		return &models.SavedReport{Status: "PROCESSING"}, nil
	}
	f.Successive = false
	return &models.SavedReport{Status: "PROCESS_COMPLETE"}, nil
}

func (f *fortifyMock) GetReportDetails(id int64) (*models.SavedReport, error) {
	return &models.SavedReport{Status: "PROCESS_COMPLETE"}, nil
}

func (f *fortifyMock) GetAllIssueDetails(projectVersionId int64) ([]*models.ProjectVersionIssue, error) {
	exploitable := "Exploitable"
	friority := "High"
	hascomments := true
	return []*models.ProjectVersionIssue{{ID: 1111, Audited: true, PrimaryTag: &exploitable, HasComments: &hascomments, Friority: &friority}, {ID: 1112, Audited: true, PrimaryTag: &exploitable, HasComments: &hascomments, Friority: &friority}}, nil
}

func (f *fortifyMock) GetIssueDetails(projectVersionId int64, issueInstanceId string) ([]*models.ProjectVersionIssue, error) {
	exploitable := "Exploitable"
	friority := "High"
	hascomments := true
	return []*models.ProjectVersionIssue{{ID: 1111, Audited: true, PrimaryTag: &exploitable, HasComments: &hascomments, Friority: &friority}}, nil
}

func (f *fortifyMock) GetIssueComments(parentId int64) ([]*models.IssueAuditComment, error) {
	comment := "Dummy"
	return []*models.IssueAuditComment{{Comment: &comment}}, nil
}

func (f *fortifyMock) UploadResultFile(endpoint, file string, projectVersionID int64) error {
	return nil
}

func (f *fortifyMock) DownloadReportFile(endpoint string, reportID int64) ([]byte, error) {
	return []byte("abcd"), nil
}

func (f *fortifyMock) DownloadResultFile(endpoint string, projectVersionID int64) ([]byte, error) {
	return []byte("defg"), nil
}

type pullRequestServiceMock struct{}

func (prService pullRequestServiceMock) ListPullRequestsWithCommit(ctx context.Context, owner, repo, sha string, opts *github.ListOptions) ([]*github.PullRequest, *github.Response, error) {
	authorString := author
	user := github.User{Login: &authorString}
	if owner == "A" {
		result := 17
		return []*github.PullRequest{{Number: &result, User: &user}}, &github.Response{}, nil
	} else if owner == "C" {
		return []*github.PullRequest{{User: &user}}, &github.Response{}, errors.New("Test error")
	} else if owner == "E" {
		return []*github.PullRequest{{User: nil}}, &github.Response{}, errors.New("Test error")
	}
	return []*github.PullRequest{}, &github.Response{}, nil
}

type execRunnerMock struct {
	numExecutions int
	current       *execution
	executions    []*execution
}

type execution struct {
	dirValue   string
	envValue   []string
	outWriter  io.Writer
	errWriter  io.Writer
	executable string
	parameters []string
}

func (er *execRunnerMock) newExecution() *execution {
	newExecution := &execution{}
	er.executions = append(er.executions, newExecution)
	return newExecution
}

func (er *execRunnerMock) currentExecution() *execution {
	if nil == er.current {
		er.numExecutions = 0
		er.current = er.newExecution()
	}
	return er.current
}

func (er *execRunnerMock) SetDir(d string) {
	er.currentExecution().dirValue = d
}

func (er *execRunnerMock) SetEnv(e []string) {
	er.currentExecution().envValue = e
}

func (er *execRunnerMock) Stdout(out io.Writer) {
	er.currentExecution().outWriter = out
}

func (er *execRunnerMock) Stderr(err io.Writer) {
	er.currentExecution().errWriter = err
}

func (er *execRunnerMock) RunExecutable(e string, p ...string) error {
	er.numExecutions++
	er.currentExecution().executable = e
	if len(p) > 0 && slices.Contains(p, "--failTranslate") {
		return errors.New("Translate failed")
	}
	er.currentExecution().parameters = p
	classpathPip := "/usr/lib/python35.zip;/usr/lib/python3.5;/usr/lib/python3.5/plat-x86_64-linux-gnu;/usr/lib/python3.5/lib-dynload;/home/piper/.local/lib/python3.5/site-packages;/usr/local/lib/python3.5/dist-packages;/usr/lib/python3/dist-packages;./lib"
	classpathMaven := "some.jar;someother.jar"
	if e == "python2" {
		if p[1] == "invalid" {
			return errors.New("Invalid command")
		}
		_, err := er.currentExecution().outWriter.Write([]byte(classpathPip))
		if err != nil {
			return err
		}
	} else if e == "mvn" {
		path := strings.ReplaceAll(p[2], "-Dmdep.outputFile=", "")
		err := os.WriteFile(path, []byte(classpathMaven), 0o644)
		if err != nil {
			return err
		}
	}
	er.current = er.newExecution()
	return nil
}

func TestDetermineArtifact(t *testing.T) {
	t.Run("Cannot get artifact without build tool", func(t *testing.T) {
		utilsMock := newFortifyTestUtilsBundle()
		utilsMock.getArtifactShouldFail = true

		_, err := determineArtifact(fortifyExecuteScanOptions{}, &utilsMock)
		assert.EqualError(t, err, "Unable to get artifact from descriptor : build tool '' not supported")
	})
}

func TestFailFortifyexecinPath(t *testing.T) {
	t.Run("Testing if fortifyupdate in $PATH or not", func(t *testing.T) {
		ff := fortifyMock{}
		ctx := context.Background()
		utils := newFortifyTestUtilsBundle()
		influx := fortifyExecuteScanInflux{}
		auditStatus := map[string]string{}
		execInPath = failMockExecinPathfortifyupdate
		config := fortifyExecuteScanOptions{SpotCheckMinimum: 4, MustAuditIssueGroups: "Audit All, Corporate Security Requirements", SpotAuditIssueGroups: "Spot Checks of Each Category"}
		_, err := runFortifyScan(ctx, config, &ff, &utils, nil, &influx, auditStatus)
		assert.EqualError(t, err, "Command not found: fortifyupdate. Please configure a supported docker image or install Fortify SCA on the system.")

	})
	t.Run("Testing if sourceanalyzer in $PATH or not", func(t *testing.T) {
		ff := fortifyMock{}
		ctx := context.Background()
		utils := newFortifyTestUtilsBundle()
		influx := fortifyExecuteScanInflux{}
		auditStatus := map[string]string{}
		execInPath = failMockExecinPathsourceanalyzer
		config := fortifyExecuteScanOptions{SpotCheckMinimum: 4, MustAuditIssueGroups: "Audit All, Corporate Security Requirements", SpotAuditIssueGroups: "Spot Checks of Each Category"}
		_, err := runFortifyScan(ctx, config, &ff, &utils, nil, &influx, auditStatus)
		assert.EqualError(t, err, "Command not found: sourceanalyzer. Please configure a supported docker image or install Fortify SCA on the system.")

	})
}

func TestExecutions(t *testing.T) {
	type parameterTestData struct {
		nameOfRun             string
		config                fortifyExecuteScanOptions
		expectedReportsLength int
		expectedReports       []string
	}

	testData := []parameterTestData{
		{
			nameOfRun:             "golang scan and verify",
			config:                fortifyExecuteScanOptions{BuildTool: "golang", BuildDescriptorFile: "go.mod"},
			expectedReportsLength: 2,
			expectedReports:       []string{"target/fortify-scan.*", "target/*.fpr"},
		},
		{
			nameOfRun:             "golang verify only",
			config:                fortifyExecuteScanOptions{BuildTool: "golang", BuildDescriptorFile: "go.mod", VerifyOnly: true},
			expectedReportsLength: 0,
		},
		{
			nameOfRun:             "maven scan and verify",
			config:                fortifyExecuteScanOptions{BuildTool: "maven", BuildDescriptorFile: "pom.xml", UpdateRulePack: true, Reporting: true, UploadResults: true},
			expectedReportsLength: 2,
			expectedReports:       []string{"target/fortify-scan.*", "target/*.fpr"},
		},
	}

	for _, data := range testData {
		t.Run(data.nameOfRun, func(t *testing.T) {
			ctx := context.Background()
			ff := fortifyMock{}
			utils := newFortifyTestUtilsBundle()
			influx := fortifyExecuteScanInflux{}
			auditStatus := map[string]string{}
			execInPath = mockExecinPath
			reports, _ := runFortifyScan(ctx, data.config, &ff, &utils, nil, &influx, auditStatus)
			if len(data.expectedReports) != data.expectedReportsLength {
				assert.Fail(t, fmt.Sprintf("Wrong number of reports detected, expected %v, actual %v", data.expectedReportsLength, len(data.expectedReports)))
			}
			if len(data.expectedReports) > 0 {
				for _, expectedPath := range data.expectedReports {
					found := false
					for _, actualPath := range reports {
						if actualPath.Target == expectedPath {
							found = true
						}
					}
					if !found {
						assert.Failf(t, "Expected path %s not found", expectedPath)
					}
				}
			}
		})
	}
}

func TestAnalyseSuspiciousExploitable(t *testing.T) {
	config := fortifyExecuteScanOptions{SpotCheckMinimum: 4, MustAuditIssueGroups: "Audit All, Corporate Security Requirements", SpotAuditIssueGroups: "Spot Checks of Each Category"}
	ff := fortifyMock{}
	influx := fortifyExecuteScanInflux{}
	name := "test"
	selectorGUID := "3"
	selectorName := "Analysis"
	selectorEntityType := "CUSTOMTAG"
	projectVersion := models.ProjectVersion{ID: 4711, Name: &name}
	auditStatus := map[string]string{}
	selectorSet := models.IssueFilterSelectorSet{
		FilterBySet: []*models.IssueFilterSelector{
			{
				GUID:        selectorGUID,
				DisplayName: selectorName,
				EntityType:  selectorEntityType,
			},
		},
		GroupBySet: []*models.IssueSelector{
			{
				GUID:        &selectorGUID,
				DisplayName: &selectorName,
				EntityType:  &selectorEntityType,
			},
		},
	}
	issues, groups := analyseSuspiciousExploitable(config, &ff, &projectVersion, &models.FilterSet{}, &selectorSet, &influx, auditStatus)
	assert.Equal(t, 9, issues)
	assert.Equal(t, 2, len(groups))

	assert.Equal(t, 4, influx.fortify_data.fields.suspicious)
	assert.Equal(t, 5, influx.fortify_data.fields.exploitable)
	assert.Equal(t, 6, influx.fortify_data.fields.suppressed)
}

func TestAnalyseUnauditedIssues(t *testing.T) {
	config := fortifyExecuteScanOptions{SpotCheckMinimumUnit: "number", SpotCheckMinimum: 4, MustAuditIssueGroups: "Audit All, Corporate Security Requirements", SpotAuditIssueGroups: "Spot Checks of Each Category"}
	ff := fortifyMock{}
	influx := fortifyExecuteScanInflux{}
	name := "test"
	projectVersion := models.ProjectVersion{ID: 4711, Name: &name}
	auditStatus := map[string]string{}
	selectorSet := models.IssueFilterSelectorSet{
		FilterBySet: []*models.IssueFilterSelector{
			{
				GUID:        "1",
				DisplayName: "Folder",
				EntityType:  "ET1",
				SelectorOptions: []*models.SelectorOption{
					{
						Value: "abcd",
					},
				},
			},
			{
				GUID:        "2",
				DisplayName: "Category",
				EntityType:  "ET2",
				SelectorOptions: []*models.SelectorOption{
					{
						Value: "abcd",
					},
				},
			},
			{
				GUID:        "3",
				DisplayName: "Analysis",
				EntityType:  "ET3",
				SelectorOptions: []*models.SelectorOption{
					{
						Value: "abcd",
					},
				},
			},
		},
	}

	spotChecksCountByCategory := []fortify.SpotChecksAuditCount{}
	issues, groups, err := analyseUnauditedIssues(config, &ff, &projectVersion, &models.FilterSet{}, &selectorSet, &influx, auditStatus, &spotChecksCountByCategory)
	assert.NoError(t, err)
	assert.Equal(t, 13, issues)
	assert.Equal(t, 3, len(groups))

	assert.Equal(t, 15, influx.fortify_data.fields.auditAllTotal)
	assert.Equal(t, 12, influx.fortify_data.fields.auditAllAudited)
	assert.Equal(t, 20, influx.fortify_data.fields.corporateTotal)
	assert.Equal(t, 11, influx.fortify_data.fields.corporateAudited)
	assert.Equal(t, 13, influx.fortify_data.fields.spotChecksTotal)
	assert.Equal(t, 11, influx.fortify_data.fields.spotChecksAudited)
	assert.Equal(t, 1, influx.fortify_data.fields.spotChecksGap)
	assert.Equal(t, 3, len(spotChecksCountByCategory))
}

func TestAnalyseUnauditedIssuesWithWrongConfig(t *testing.T) {
	config := fortifyExecuteScanOptions{SpotCheckMinimumUnit: "float"}
	spotChecksCountByCategory := []fortify.SpotChecksAuditCount{}
	ff := fortifyMock{}
	auditStatus := map[string]string{}
	_, _, err := analyseUnauditedIssues(config, &ff, &models.ProjectVersion{}, &models.FilterSet{}, &models.IssueFilterSelectorSet{}, &fortifyExecuteScanInflux{}, auditStatus, &spotChecksCountByCategory)
	assert.Error(t, err)
	assert.Equal(t, "Invalid spotCheckMinimumUnit. Please set it as 'percentage' or 'number'.", err.Error())
}

func TestTriggerFortifyScan(t *testing.T) {
	t.Run("maven", func(t *testing.T) {
		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		utils := newFortifyTestUtilsBundle()
		config := fortifyExecuteScanOptions{
			BuildTool:                "maven",
			AutodetectClasspath:      true,
			BuildDescriptorFile:      "./pom.xml",
			AdditionalScanParameters: []string{"-Dtest=property"},
			Memory:                   "-Xmx4G -Xms2G",
			Src:                      []string{"**/*.xml", "**/*.html", "**/*.jsp", "**/*.js", "src/main/resources/**/*", "src/main/java/**/*"},
		}
		triggerFortifyScan(config, &utils, "test", "testLabel", "my.group-myartifact")

		assert.Equal(t, 3, utils.numExecutions)

		assert.Equal(t, "mvn", utils.executions[0].executable)
		assert.Equal(t, []string{"--file", "./pom.xml", "-Dmdep.outputFile=fortify-execute-scan-cp.txt", "-Dfortify", "-DincludeScope=compile", "-DskipTests", "-Dmaven.javadoc.skip=true", "--fail-at-end", "-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn", "--batch-mode", "dependency:build-classpath", "package"}, utils.executions[0].parameters)

		assert.Equal(t, "sourceanalyzer", utils.executions[1].executable)
		assert.True(t, reflect.DeepEqual([]string{"-verbose", "-64", "-b", "test", "-Xmx4G", "-Xms2G", "-cp", "some.jar;someother.jar", "-exclude", "**/src/test/**/*", "**/*.xml", "**/*.html", "**/*.jsp", "**/*.js", "src/main/resources/**/*", "src/main/java/**/*"}, utils.executions[1].parameters) || reflect.DeepEqual([]string{"-verbose", "-64", "-b", "test", "-Xmx4G", "-Xms2G", "-cp", "some.jar;someother.jar", "-exclude", "**/src/test/**/*", "**/*.xml", "**/*.html", "**/*.jsp", "**/*.js", "src/main/resources/**/*", "src/main/java/**/*"}, utils.executions[1].parameters))

		assert.Equal(t, "sourceanalyzer", utils.executions[2].executable)
		assert.Equal(t, []string{"-verbose", "-64", "-b", "test", "-scan", "-Xmx4G", "-Xms2G", "-Dtest=property", "-build-label", "testLabel", "-build-project", "my.group-myartifact", "-logfile", "target/fortify-scan.log", "-f", "target/result.fpr"}, utils.executions[2].parameters)
	})

	t.Run("pip", func(t *testing.T) {
		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		utils := newFortifyTestUtilsBundle()
		config := fortifyExecuteScanOptions{BuildTool: "pip", PythonVersion: "python2", AutodetectClasspath: true, BuildDescriptorFile: "./setup.py", PythonRequirementsFile: "./requirements.txt", PythonInstallCommand: "pip2 install --user", Memory: "-Xmx4G -Xms2G"}
		triggerFortifyScan(config, &utils, "test", "testLabel", "")

		assert.Equal(t, 5, utils.numExecutions)

		assert.Equal(t, "python2", utils.executions[0].executable)
		separator := getSeparator()
		template := fmt.Sprintf("import sys;p=sys.path;p.remove('');print('%v'.join(p))", separator)
		assert.Equal(t, []string{"-c", template}, utils.executions[0].parameters)

		assert.Equal(t, "pip2", utils.executions[1].executable)
		assert.Equal(t, []string{"install", "--user", "-r", "./requirements.txt", ""}, utils.executions[1].parameters)

		assert.Equal(t, "pip2", utils.executions[2].executable)
		assert.Equal(t, []string{"install", "--user"}, utils.executions[2].parameters)

		assert.Equal(t, "sourceanalyzer", utils.executions[3].executable)
		assert.Equal(t, []string{"-verbose", "-64", "-b", "test", "-Xmx4G", "-Xms2G", "-python-path", "/usr/lib/python35.zip;/usr/lib/python3.5;/usr/lib/python3.5/plat-x86_64-linux-gnu;/usr/lib/python3.5/lib-dynload;/home/piper/.local/lib/python3.5/site-packages;/usr/local/lib/python3.5/dist-packages;/usr/lib/python3/dist-packages;./lib", "-python-version", "2", "-exclude", fmt.Sprintf("./**/tests/**/*%s./**/setup.py", separator), "./**/*"}, utils.executions[3].parameters)

		assert.Equal(t, "sourceanalyzer", utils.executions[4].executable)
		assert.Equal(t, []string{"-verbose", "-64", "-b", "test", "-scan", "-Xmx4G", "-Xms2G", "-build-label", "testLabel", "-logfile", "target/fortify-scan.log", "-f", "target/result.fpr"}, utils.executions[4].parameters)
	})

	t.Run("invalid buildTool", func(t *testing.T) {
		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		utils := newFortifyTestUtilsBundle()
		config := fortifyExecuteScanOptions{
			BuildTool:           "docker",
			AutodetectClasspath: true,
		}
		err := triggerFortifyScan(config, &utils, "test", "testLabel", "my.group-myartifact")

		assert.Error(t, err)
		assert.Equal(t, "buildTool 'docker' is not supported by this step", err.Error())
	})
}

func TestGetMinSpotChecksPerCategory(t *testing.T) {
	testExpectedGetMinSpotChecksPerCategory := func(spotChecksMinUnit string, spotChecksMax int, spotChecksMin int, issuesPerCategory int, spotChecksMinCalculatedExpected int) {
		testName := fmt.Sprintf("Test GetMinSpotChecksPerCategory for SpotCheckMinimumUnit: %v, SpotCheckMaximum: %v, SpotCheckMinimum: %v, issuesPerCategory: %v", spotChecksMinUnit, spotChecksMax, spotChecksMin, issuesPerCategory)
		t.Run(testName, func(t *testing.T) {
			config := fortifyExecuteScanOptions{SpotCheckMinimumUnit: spotChecksMinUnit, SpotCheckMaximum: spotChecksMax, SpotCheckMinimum: spotChecksMin}
			spotCheckMin := getMinSpotChecksPerCategory(config, issuesPerCategory)
			assert.Equal(t, spotChecksMinCalculatedExpected, spotCheckMin)
		})
	}

	testExpectedGetMinSpotChecksPerCategory("percentage", 0, 1, 10, 1)
	testExpectedGetMinSpotChecksPerCategory("percentage", 10, 10, 3, 1)
	testExpectedGetMinSpotChecksPerCategory("percentage", 10, 10, 8, 1)
	testExpectedGetMinSpotChecksPerCategory("percentage", 10, 10, 10, 1)
	testExpectedGetMinSpotChecksPerCategory("percentage", 10, 10, 24, 3)
	testExpectedGetMinSpotChecksPerCategory("percentage", 10, 10, 26, 3)
	testExpectedGetMinSpotChecksPerCategory("percentage", 10, 10, 100, 10)
	testExpectedGetMinSpotChecksPerCategory("percentage", 10, 10, 200, 10)
	testExpectedGetMinSpotChecksPerCategory("percentage", 10, 50, 10, 5)
	testExpectedGetMinSpotChecksPerCategory("percentage", 0, 50, 100, 50)
	testExpectedGetMinSpotChecksPerCategory("percentage", -10, 50, 100, 50)

	testExpectedGetMinSpotChecksPerCategory("number", 0, 1, 10, 1)
	testExpectedGetMinSpotChecksPerCategory("number", 5, 10, 100, 5)
}

func TestGenerateAndDownloadQGateReport(t *testing.T) {
	ffMock := fortifyMock{Successive: false}
	config := fortifyExecuteScanOptions{ReportTemplateID: 18, ReportType: "PDF"}
	name := "test"
	projectVersion := models.ProjectVersion{ID: 4711, Name: &name}
	project := models.Project{ID: 815, Name: &name}
	projectVersion.Project = &project

	t.Run("success", func(t *testing.T) {
		data, err := generateAndDownloadQGateReport(config, &ffMock, &project, &projectVersion)
		assert.NoError(t, err)
		assert.Equal(t, []byte("abcd"), data)
	})
}

var (
	defaultPollingDelay   = 10 * time.Second
	defaultPollingTimeout = 0 * time.Minute
)

func verifyScanResultsFinishedUploadingDefaults(config fortifyExecuteScanOptions, sys fortify.System, projectVersionID int64) error {
	return verifyScanResultsFinishedUploading(config, sys, projectVersionID, "", &models.FilterSet{},
		defaultPollingDelay, defaultPollingTimeout)
}

func TestVerifyScanResultsFinishedUploading(t *testing.T) {
	t.Parallel()

	t.Run("error no recent upload detected", func(t *testing.T) {
		ffMock := fortifyMock{}
		config := fortifyExecuteScanOptions{DeltaMinutes: -1}
		err := verifyScanResultsFinishedUploadingDefaults(config, &ffMock, 4711)
		assert.EqualError(t, err, "no recent upload detected on Project Version")
	})

	config := fortifyExecuteScanOptions{DeltaMinutes: 20}
	t.Run("success", func(t *testing.T) {
		ffMock := fortifyMock{}
		err := verifyScanResultsFinishedUploadingDefaults(config, &ffMock, 4711)
		assert.NoError(t, err)
	})

	t.Run("error processing", func(t *testing.T) {
		ffMock := fortifyMock{}
		err := verifyScanResultsFinishedUploadingDefaults(config, &ffMock, 4712)
		assert.EqualError(t, err, "There are artifacts that failed processing for Project Version 4712\n/html/ssc/index.jsp#!/version/4712/artifacts?filterSet=")
	})

	t.Run("error required auth", func(t *testing.T) {
		ffMock := fortifyMock{}
		err := verifyScanResultsFinishedUploadingDefaults(config, &ffMock, 4713)
		assert.EqualError(t, err, "There are artifacts that require manual approval for Project Version 4713, please visit Fortify SSC and approve them for processing\n/html/ssc/index.jsp#!/version/4713/artifacts?filterSet=")
	})

	t.Run("error polling timeout", func(t *testing.T) {
		ffMock := fortifyMock{}
		err := verifyScanResultsFinishedUploadingDefaults(config, &ffMock, 4714)
		assert.EqualError(t, err, "terminating after 0s since artifact for Project Version 4714 is still in status PROCESSING")
	})

	t.Run("success build label", func(t *testing.T) {
		ffMock := fortifyMock{}
		err := verifyScanResultsFinishedUploading(config, &ffMock, 4715, "/commit/test", &models.FilterSet{},
			10*time.Second, time.Duration(config.PollingMinutes)*time.Minute)
		assert.NoError(t, err)
	})

	t.Run("failure after polling", func(t *testing.T) {
		config := fortifyExecuteScanOptions{DeltaMinutes: 1}
		ffMock := fortifyMock{}
		const pollingDelay = 1 * time.Second
		const timeout = 1 * time.Second
		err := verifyScanResultsFinishedUploading(config, &ffMock, 4716, "", &models.FilterSet{}, pollingDelay, timeout)
		assert.EqualError(t, err, "terminating after 1s since artifact for Project Version 4716 is still in status PROCESSING")
	})

	t.Run("success after polling", func(t *testing.T) {
		config := fortifyExecuteScanOptions{DeltaMinutes: 1}
		ffMock := fortifyMock{}
		const pollingDelay = 500 * time.Millisecond
		const timeout = 1 * time.Second
		err := verifyScanResultsFinishedUploading(config, &ffMock, 4716, "", &models.FilterSet{}, pollingDelay, timeout)
		assert.NoError(t, err)
	})

	t.Run("error no artifacts", func(t *testing.T) {
		ffMock := fortifyMock{}
		err := verifyScanResultsFinishedUploadingDefaults(config, &ffMock, 4717)
		assert.EqualError(t, err, "no uploaded artifacts for assessment detected for project version with ID 4717")
	})

	t.Run("warn old artifacts have errors", func(t *testing.T) {
		ffMock := fortifyMock{}

		logBuffer := new(bytes.Buffer)
		logOutput := log.Entry().Logger.Out
		log.Entry().Logger.Out = logBuffer
		defer func() { log.Entry().Logger.Out = logOutput }()

		err := verifyScanResultsFinishedUploadingDefaults(config, &ffMock, 4718)
		assert.NoError(t, err)
		assert.Contains(t, logBuffer.String(), "Previous uploads detected that failed processing")
	})
}

func TestCalculateTimeDifferenceToLastUpload(t *testing.T) {
	diffSeconds := calculateTimeDifferenceToLastUpload(models.Iso8601MilliDateTime(time.Now().UTC()), 1234)

	assert.Equal(t, true, diffSeconds < 1)
}

func TestExecuteTemplatedCommand(t *testing.T) {
	utils := newFortifyTestUtilsBundle()
	template := []string{"{{.Executable}}", "-c", "{{.Param}}"}
	context := map[string]string{"Executable": "test.cmd", "Param": "abcd"}
	executeTemplatedCommand(&utils, template, context)

	assert.Equal(t, "test.cmd", utils.executions[0].executable)
	assert.Equal(t, []string{"-c", "abcd"}, utils.executions[0].parameters)
}

func TestDeterminePullRequestMerge(t *testing.T) {
	config := fortifyExecuteScanOptions{CommitMessage: "Merge pull request #2462 from branch f-test", PullRequestMessageRegex: `(?m).*Merge pull request #(\d+) from.*`, PullRequestMessageRegexGroup: 1}

	t.Run("success", func(t *testing.T) {
		match, authorString := determinePullRequestMerge(config)
		assert.Equal(t, "2462", match, "Expected different result")
		assert.Equal(t, "", authorString, "Expected different result")
	})

	t.Run("no match", func(t *testing.T) {
		config.CommitMessage = "Some test commit"
		match, authorString := determinePullRequestMerge(config)
		assert.Equal(t, "0", match, "Expected different result")
		assert.Equal(t, "", authorString, "Expected different result")
	})
}

func TestDeterminePullRequestMergeGithub(t *testing.T) {
	prServiceMock := pullRequestServiceMock{}

	t.Run("success", func(t *testing.T) {
		match, authorString, err := determinePullRequestMergeGithub(nil, fortifyExecuteScanOptions{Owner: "A"}, prServiceMock)
		assert.NoError(t, err)
		assert.Equal(t, "17", match, "Expected different result")
		assert.Equal(t, author, authorString, "Expected different result")
	})

	t.Run("no match", func(t *testing.T) {
		match, authorString, err := determinePullRequestMergeGithub(nil, fortifyExecuteScanOptions{Owner: "B"}, prServiceMock)
		assert.NoError(t, err)
		assert.Equal(t, "0", match, "Expected different result")
		assert.Equal(t, "", authorString, "Expected different result")
	})

	t.Run("error", func(t *testing.T) {
		match, authorString, err := determinePullRequestMergeGithub(nil, fortifyExecuteScanOptions{Owner: "E"}, prServiceMock)
		assert.EqualError(t, err, "Test error")
		assert.Equal(t, "0", match, "Expected different result")
		assert.Equal(t, "", authorString, "Expected different result")
	})
}

func TestTranslateProject(t *testing.T) {
	t.Run("python", func(t *testing.T) {
		utils := newFortifyTestUtilsBundle()
		config := fortifyExecuteScanOptions{BuildTool: "pip", Memory: "-Xmx4G", Translate: `[{"pythonPath":"./some/path","src":"./**/*","exclude":"./tests/**/*"}]`}
		translateProject(&config, &utils, "/commit/7267658798797", "")
		assert.Equal(t, "sourceanalyzer", utils.executions[0].executable, "Expected different executable")
		assert.Equal(t, []string{"-verbose", "-64", "-b", "/commit/7267658798797", "-Xmx4G", "-python-path", "./some/path", "-exclude", "./tests/**/*", "./**/*"}, utils.executions[0].parameters, "Expected different parameters")
	})

	t.Run("asp", func(t *testing.T) {
		utils := newFortifyTestUtilsBundle()
		config := fortifyExecuteScanOptions{BuildTool: "windows", Memory: "-Xmx6G", Translate: `[{"aspnetcore":"true","dotNetCoreVersion":"3.5","exclude":"./tests/**/*","libDirs":"tmp/","src":"./**/*"}]`}
		translateProject(&config, &utils, "/commit/7267658798797", "")
		assert.Equal(t, "sourceanalyzer", utils.executions[0].executable, "Expected different executable")
		assert.Equal(t, []string{"-verbose", "-64", "-b", "/commit/7267658798797", "-Xmx6G", "-aspnetcore", "-dotnet-core-version", "3.5", "-libdirs", "tmp/", "-exclude", "./tests/**/*", "./**/*"}, utils.executions[0].parameters, "Expected different parameters")
	})

	t.Run("java", func(t *testing.T) {
		utils := newFortifyTestUtilsBundle()
		config := fortifyExecuteScanOptions{BuildTool: "maven", Memory: "-Xmx2G", Translate: `[{"classpath":"./classes/*.jar","extdirs":"tmp/","jdk":"1.8.0-21","source":"1.8","sourcepath":"src/ext/","src":"./**/*"}]`}
		translateProject(&config, &utils, "/commit/7267658798797", "")
		assert.Equal(t, "sourceanalyzer", utils.executions[0].executable, "Expected different executable")
		assert.Equal(t, []string{"-verbose", "-64", "-b", "/commit/7267658798797", "-Xmx2G", "-cp", "./classes/*.jar", "-extdirs", "tmp/", "-source", "1.8", "-jdk", "1.8.0-21", "-sourcepath", "src/ext/", "./**/*"}, utils.executions[0].parameters, "Expected different parameters")
	})

	t.Run("auto classpath", func(t *testing.T) {
		utils := newFortifyTestUtilsBundle()
		config := fortifyExecuteScanOptions{BuildTool: "maven", Memory: "-Xmx2G", Translate: `[{"classpath":"./classes/*.jar", "extdirs":"tmp/","jdk":"1.8.0-21","source":"1.8","sourcepath":"src/ext/","src":"./**/*"}]`}
		translateProject(&config, &utils, "/commit/7267658798797", "./WEB-INF/lib/*.jar")
		assert.Equal(t, "sourceanalyzer", utils.executions[0].executable, "Expected different executable")
		assert.Equal(t, []string{"-verbose", "-64", "-b", "/commit/7267658798797", "-Xmx2G", "-cp", "./WEB-INF/lib/*.jar", "-extdirs", "tmp/", "-source", "1.8", "-jdk", "1.8.0-21", "-sourcepath", "src/ext/", "./**/*"}, utils.executions[0].parameters, "Expected different parameters")
	})

	t.Run("failure propagated", func(t *testing.T) {
		utils := newFortifyTestUtilsBundle()
		config := fortifyExecuteScanOptions{BuildTool: "maven", Memory: "-Xmx2G", Translate: `[{"classpath":"./classes/*.jar", "extdirs":"tmp/","jdk":"1.8.0-21","source":"1.8","sourcepath":"src/ext/","src":"./**/*"}]`}
		err := translateProject(&config, &utils, "--failTranslate", "./WEB-INF/lib/*.jar")
		assert.Error(t, err)
		assert.Equal(t, "failed to execute sourceanalyzer translate command with options [-verbose -64 -b --failTranslate -Xmx2G -cp ./WEB-INF/lib/*.jar -extdirs tmp/ -source 1.8 -jdk 1.8.0-21 -sourcepath src/ext/ ./**/*]: Translate failed", err.Error())
	})
}

func TestScanProject(t *testing.T) {
	config := fortifyExecuteScanOptions{Memory: "-Xmx4G"}

	t.Run("normal", func(t *testing.T) {
		utils := newFortifyTestUtilsBundle()
		scanProject(&config, &utils, "/commit/7267658798797", "label", "my.group-myartifact")
		assert.Equal(t, "sourceanalyzer", utils.executions[0].executable, "Expected different executable")
		assert.Equal(t, []string{"-verbose", "-64", "-b", "/commit/7267658798797", "-scan", "-Xmx4G", "-build-label", "label", "-build-project", "my.group-myartifact", "-logfile", "target/fortify-scan.log", "-f", "target/result.fpr"}, utils.executions[0].parameters, "Expected different parameters")
	})

	t.Run("quick", func(t *testing.T) {
		utils := newFortifyTestUtilsBundle()
		config.QuickScan = true
		scanProject(&config, &utils, "/commit/7267658798797", "", "")
		assert.Equal(t, "sourceanalyzer", utils.executions[0].executable, "Expected different executable")
		assert.Equal(t, []string{"-verbose", "-64", "-b", "/commit/7267658798797", "-scan", "-Xmx4G", "-quick", "-logfile", "target/fortify-scan.log", "-f", "target/result.fpr"}, utils.executions[0].parameters, "Expected different parameters")
	})
}

func TestAutoresolveClasspath(t *testing.T) {
	t.Run("success pip", func(t *testing.T) {
		utils := newFortifyTestUtilsBundle()
		dir := t.TempDir()
		file := filepath.Join(dir, "cp.txt")

		result, err := autoresolvePipClasspath("python2", []string{"-c", "import sys;p=sys.path;p.remove('');print(';'.join(p))"}, file, &utils)
		assert.NoError(t, err)
		assert.Equal(t, "python2", utils.executions[0].executable, "Expected different executable")
		assert.Equal(t, []string{"-c", "import sys;p=sys.path;p.remove('');print(';'.join(p))"}, utils.executions[0].parameters, "Expected different parameters")
		assert.Equal(t, "/usr/lib/python35.zip;/usr/lib/python3.5;/usr/lib/python3.5/plat-x86_64-linux-gnu;/usr/lib/python3.5/lib-dynload;/home/piper/.local/lib/python3.5/site-packages;/usr/local/lib/python3.5/dist-packages;/usr/lib/python3/dist-packages;./lib", result, "Expected different result")
	})

	t.Run("error pip file", func(t *testing.T) {
		utils := newFortifyTestUtilsBundle()

		_, err := autoresolvePipClasspath("python2", []string{"-c", "import sys;p=sys.path;p.remove('');print(';'.join(p))"}, "../.", &utils)
		assert.Error(t, err)
	})

	t.Run("error pip command", func(t *testing.T) {
		utils := newFortifyTestUtilsBundle()
		dir := t.TempDir()
		file := filepath.Join(dir, "cp.txt")

		_, err := autoresolvePipClasspath("python2", []string{"-c", "invalid"}, file, &utils)
		assert.Error(t, err)
		assert.Equal(t, "failed to run classpath autodetection command python2 with parameters [-c invalid]: Invalid command", err.Error())
	})

	t.Run("success maven", func(t *testing.T) {
		utils := newFortifyTestUtilsBundle()
		dir := t.TempDir()
		file := filepath.Join(dir, "cp.txt")

		result, err := autoresolveMavenClasspath(fortifyExecuteScanOptions{BuildDescriptorFile: "pom.xml"}, file, &utils)
		assert.NoError(t, err)
		assert.Equal(t, "mvn", utils.executions[0].executable, "Expected different executable")
		assert.Equal(t, []string{"--file", "pom.xml", fmt.Sprintf("-Dmdep.outputFile=%v", file), "-Dfortify", "-DincludeScope=compile", "-DskipTests", "-Dmaven.javadoc.skip=true", "--fail-at-end", "-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn", "--batch-mode", "dependency:build-classpath", "package"}, utils.executions[0].parameters, "Expected different parameters")
		assert.Equal(t, "some.jar;someother.jar", result, "Expected different result")
	})
}

func TestPopulateMavenTranslate(t *testing.T) {
	t.Run("src without translate", func(t *testing.T) {
		config := fortifyExecuteScanOptions{Src: []string{"./**/*"}}
		translate, err := populateMavenGradleTranslate(&config, "")
		assert.NoError(t, err)
		assert.Equal(t, `[{"classpath":"","exclude":"**/src/test/**/*","src":"./**/*"}]`, translate)
	})

	t.Run("exclude without translate", func(t *testing.T) {
		config := fortifyExecuteScanOptions{Exclude: []string{"./**/*"}}
		translate, err := populateMavenGradleTranslate(&config, "")
		assert.NoError(t, err)
		assert.Equal(t, `[{"classpath":"","exclude":"./**/*","src":"**/*.xml:**/*.html:**/*.jsp:**/*.js:**/src/main/resources/**/*:**/src/main/java/**/*:**/src/gen/java/cds/**/*:**/target/main/java/**/*:**/target/main/resources/**/*:**/target/generated-sources/**/*"}]`, translate)
	})

	t.Run("with translate", func(t *testing.T) {
		config := fortifyExecuteScanOptions{Translate: `[{"classpath":""}]`, Src: []string{"./**/*"}, Exclude: []string{"./**/*"}}
		translate, err := populateMavenGradleTranslate(&config, "ignored/path")
		assert.NoError(t, err)
		assert.Equal(t, `[{"classpath":""}]`, translate)
	})
}

func TestPopulatePipTranslate(t *testing.T) {
	t.Run("PythonAdditionalPath without translate", func(t *testing.T) {
		config := fortifyExecuteScanOptions{PythonVersion: "python2", PythonAdditionalPath: []string{"./lib", "."}}
		translate, err := populatePipTranslate(&config, "")
		separator := getSeparator()
		expected := fmt.Sprintf(`[{"exclude":"./**/tests/**/*%v./**/setup.py","pythonPath":"%v./lib%v.","pythonVersion":"2","src":"./**/*"}]`,
			separator, separator, separator)
		assert.NoError(t, err)
		assert.Equal(t, expected, translate)
	})

	t.Run("Invalid python version", func(t *testing.T) {
		config := fortifyExecuteScanOptions{PythonVersion: "python4", PythonAdditionalPath: []string{"./lib", "."}}
		_, err := populatePipTranslate(&config, "")
		assert.Error(t, err)
	})

	t.Run("Src without translate", func(t *testing.T) {
		config := fortifyExecuteScanOptions{PythonVersion: "python3", Src: []string{"./**/*.py"}}
		translate, err := populatePipTranslate(&config, "")
		separator := getSeparator()
		expected := fmt.Sprintf(
			`[{"exclude":"./**/tests/**/*%v./**/setup.py","pythonPath":"%v","pythonVersion":"3","src":"./**/*.py"}]`,
			separator, separator)
		assert.NoError(t, err)
		assert.Equal(t, expected, translate)
	})

	t.Run("Exclude without translate", func(t *testing.T) {
		config := fortifyExecuteScanOptions{PythonVersion: "python3", Exclude: []string{"./**/tests/**/*"}}
		translate, err := populatePipTranslate(&config, "")
		separator := getSeparator()
		expected := fmt.Sprintf(
			`[{"exclude":"./**/tests/**/*","pythonPath":"%v","pythonVersion":"3","src":"./**/*"}]`,
			separator)
		assert.NoError(t, err)
		assert.Equal(t, expected, translate)
	})

	t.Run("with translate", func(t *testing.T) {
		config := fortifyExecuteScanOptions{
			Translate:            `[{"pythonPath":""}]`,
			Src:                  []string{"./**/*"},
			PythonAdditionalPath: []string{"./lib", "."},
		}
		translate, err := populatePipTranslate(&config, "ignored/path")
		assert.NoError(t, err)
		assert.Equal(t, `[{"pythonPath":""}]`, translate, "Expected different parameters")
	})
}

func TestRemoveDuplicates(t *testing.T) {
	testData := []struct {
		name      string
		input     string
		expected  string
		separator string
	}{
		{"empty", "", "", "x"},
		{"no duplicates", ":a::b::", "a:b", ":"},
		{"duplicates", "::a:b:a:b::a", "a:b", ":"},
		{"long separator", "..a.b....ab..a.b", "a.b..ab", ".."},
		{"no separator", "abc", "abc", ""},
	}
	for _, data := range testData {
		t.Run(data.name, func(t *testing.T) {
			assert.Equal(t, data.expected, removeDuplicates(data.input, data.separator))
		})
	}
}

func toFortifyTime(time time.Time) models.Iso8601MilliDateTime {
	return models.Iso8601MilliDateTime(time.UTC())
}

func TestGetProxyParams(t *testing.T) {
	t.Run("Valid Proxy URL", func(t *testing.T) {
		proxyPort, proxyHost := getProxyParams("http://testproxy.com:8080")
		assert.Equal(t, "8080", proxyPort)
		assert.Equal(t, "testproxy.com", proxyHost)
	})

	t.Run("Invalid Proxy URL", func(t *testing.T) {
		proxyPort, proxyHost := getProxyParams("testproxy.com:8080")
		assert.Equal(t, "", proxyPort)
		assert.Equal(t, "", proxyHost)
	})
}
