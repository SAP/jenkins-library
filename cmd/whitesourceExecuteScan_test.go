package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/format"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/reporting"
	"github.com/SAP/jenkins-library/pkg/versioning"
	ws "github.com/SAP/jenkins-library/pkg/whitesource"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/google/go-github/v68/github"
)

type whitesourceUtilsMock struct {
	*ws.ScanUtilsMock
	coordinates             versioning.Coordinates
	usedBuildTool           string
	usedBuildDescriptorFile string
	usedOptions             versioning.Options
}

func (w *whitesourceUtilsMock) GetArtifactCoordinates(buildTool, buildDescriptorFile string,
	options *versioning.Options,
) (versioning.Coordinates, error) {
	w.usedBuildTool = buildTool
	w.usedBuildDescriptorFile = buildDescriptorFile
	w.usedOptions = *options
	return w.coordinates, nil
}

const wsTimeNow = "2010-05-10 00:15:42"

func (w *whitesourceUtilsMock) Now() time.Time {
	now, _ := time.Parse("2006-01-02 15:04:05", wsTimeNow)
	return now
}

func (w *whitesourceUtilsMock) GetIssueService() *github.IssuesService {
	return nil
}

func (w *whitesourceUtilsMock) GetSearchService() *github.SearchService {
	return nil
}

func newWhitesourceUtilsMock() *whitesourceUtilsMock {
	return &whitesourceUtilsMock{
		ScanUtilsMock: &ws.ScanUtilsMock{
			FilesMock:      &mock.FilesMock{},
			ExecMockRunner: &mock.ExecMockRunner{},
		},
		coordinates: versioning.Coordinates{
			GroupID:    "mock-group-id",
			ArtifactID: "mock-artifact-id",
			Version:    "1.0.42",
		},
	}
}

func TestNewWhitesourceUtils(t *testing.T) {
	t.Parallel()
	config := ScanOptions{}
	utils := newWhitesourceUtils(&config, &github.Client{Issues: &github.IssuesService{}, Search: &github.SearchService{}})

	assert.NotNil(t, utils.Client)
	assert.NotNil(t, utils.Command)
	assert.NotNil(t, utils.Files)
}

func TestRunWhitesourceExecuteScan(t *testing.T) {
	t.Parallel()
	t.Run("fails for invalid configured project token", func(t *testing.T) {
		ctx := context.Background()
		// init
		config := ScanOptions{
			BuildDescriptorFile: "my-mta.yml",
			VersioningModel:     "major",
			ProductName:         "mock-product",
			ProjectToken:        "no-such-project-token",
			AgentDownloadURL:    "https://whitesource.com/agent.jar",
			AgentFileName:       "ua.jar",
		}
		utilsMock := newWhitesourceUtilsMock()
		utilsMock.AddFile("wss-generated-file.config", []byte("key=value"))
		systemMock := ws.NewSystemMock("ignored")
		scan := newWhitesourceScan(&config)
		cpe := whitesourceExecuteScanCommonPipelineEnvironment{}
		influx := whitesourceExecuteScanInflux{}
		// test
		err := runWhitesourceExecuteScan(ctx, &config, scan, utilsMock, systemMock, &cpe, &influx)
		// assert
		assert.EqualError(t, err, "failed to resolve and aggregate project name: failed to get project by token: no project with token 'no-such-project-token' found in Whitesource")
		assert.Equal(t, "", config.ProjectName)
		assert.Equal(t, "", scan.AggregateProjectName)
	})
	t.Run("retrieves aggregate project name by configured token", func(t *testing.T) {
		ctx := context.Background()
		// init
		config := ScanOptions{
			BuildDescriptorFile:       "my-mta.yml",
			VersioningModel:           "major",
			AgentDownloadURL:          "https://whitesource.com/agent.jar",
			VulnerabilityReportFormat: "pdf",
			Reporting:                 true,
			AgentFileName:             "ua.jar",
			ProductName:               "mock-product",
			ProjectToken:              "mock-project-token",
		}
		utilsMock := newWhitesourceUtilsMock()
		utilsMock.AddFile("wss-generated-file.config", []byte("key=value"))
		lastUpdatedDate := time.Now().Format(ws.DateTimeLayout)
		systemMock := ws.NewSystemMock(lastUpdatedDate)
		systemMock.Alerts = []ws.Alert{}
		scan := newWhitesourceScan(&config)
		cpe := whitesourceExecuteScanCommonPipelineEnvironment{}
		influx := whitesourceExecuteScanInflux{}
		// test
		err := runWhitesourceExecuteScan(ctx, &config, scan, utilsMock, systemMock, &cpe, &influx)
		// assert
		assert.NoError(t, err)
		// Retrieved project name is stored in scan.AggregateProjectName, but not in config.ProjectName
		// in order to differentiate between aggregate-project scanning and multi-project scanning.
		assert.Equal(t, "", config.ProjectName)
		assert.Equal(t, "mock-project", scan.AggregateProjectName)
		if assert.Len(t, utilsMock.DownloadedFiles, 1) {
			assert.Equal(t, ws.DownloadedFile{
				SourceURL: "https://whitesource.com/agent.jar",
				FilePath:  "ua.jar",
			}, utilsMock.DownloadedFiles[0])
		}
		if assert.Len(t, cpe.custom.whitesourceProjectNames, 1) {
			assert.Equal(t, []string{"mock-project - 1"}, cpe.custom.whitesourceProjectNames)
		}
		assert.True(t, utilsMock.HasWrittenFile(filepath.Join(ws.ReportsDirectory, "mock-project - 1-vulnerability-report.pdf")))
		assert.True(t, utilsMock.HasWrittenFile(filepath.Join(ws.ReportsDirectory, "mock-project - 1-vulnerability-report.pdf")))
		assert.Equal(t, 3, len(utilsMock.ExecMockRunner.Calls), "no InstallCommand must be executed")
	})
	t.Run("executes the InstallCommand prior to the scan", func(t *testing.T) {
		ctx := context.Background()
		// init
		config := ScanOptions{
			BuildDescriptorFile:       "my-mta.yml",
			VersioningModel:           "major",
			AgentDownloadURL:          "https://whitesource.com/agent.jar",
			VulnerabilityReportFormat: "pdf",
			Reporting:                 true,
			AgentFileName:             "ua.jar",
			ProductName:               "mock-product",
			ProjectToken:              "mock-project-token",
			InstallCommand:            "echo hello world",
		}
		utilsMock := newWhitesourceUtilsMock()
		utilsMock.AddFile("wss-generated-file.config", []byte("key=value"))
		lastUpdatedDate := time.Now().Format(ws.DateTimeLayout)
		systemMock := ws.NewSystemMock(lastUpdatedDate)
		systemMock.Alerts = []ws.Alert{}
		scan := newWhitesourceScan(&config)
		cpe := whitesourceExecuteScanCommonPipelineEnvironment{}
		influx := whitesourceExecuteScanInflux{}
		// test
		err := runWhitesourceExecuteScan(ctx, &config, scan, utilsMock, systemMock, &cpe, &influx)
		// assert
		assert.NoError(t, err)
		assert.Equal(t, 4, len(utilsMock.ExecMockRunner.Calls), "InstallCommand not executed")
		assert.Equal(t, mock.ExecCall{Exec: "echo", Params: []string{"hello", "world"}}, utilsMock.ExecMockRunner.Calls[0], "run command/params of InstallCommand incorrect")
	})
	t.Run("fails if the InstallCommand fails", func(t *testing.T) {
		ctx := context.Background()
		// init
		config := ScanOptions{
			BuildDescriptorFile:       "my-mta.yml",
			VersioningModel:           "major",
			AgentDownloadURL:          "https://whitesource.com/agent.jar",
			VulnerabilityReportFormat: "pdf",
			Reporting:                 true,
			AgentFileName:             "ua.jar",
			ProductName:               "mock-product",
			ProjectToken:              "mock-project-token",
			InstallCommand:            "echo this-will-fail",
		}
		utilsMock := newWhitesourceUtilsMock()
		utilsMock.AddFile("wss-generated-file.config", []byte("key=value"))
		lastUpdatedDate := time.Now().Format(ws.DateTimeLayout)
		systemMock := ws.NewSystemMock(lastUpdatedDate)
		systemMock.Alerts = []ws.Alert{}
		scan := newWhitesourceScan(&config)
		cpe := whitesourceExecuteScanCommonPipelineEnvironment{}
		influx := whitesourceExecuteScanInflux{}
		utilsMock.ExecMockRunner.ShouldFailOnCommand = map[string]error{
			"echo this-will-fail": errors.New("error case"),
		}
		// test
		err := runWhitesourceExecuteScan(ctx, &config, scan, utilsMock, systemMock, &cpe, &influx)
		// assert
		assert.EqualError(t, err, "failed to execute WhiteSource scan: failed to execute Scan: failed to execute install command: echo this-will-fail: error case")
	})
}

func TestCheckAndReportScanResults(t *testing.T) {
	t.Parallel()
	t.Run("no reports requested", func(t *testing.T) {
		ctx := context.Background()
		// init
		config := &ScanOptions{
			ProductToken: "mock-product-token",
			ProjectToken: "mock-project-token",
			Version:      "1",
		}
		scan := newWhitesourceScan(config)
		utils := newWhitesourceUtilsMock()
		system := ws.NewSystemMock(time.Now().Format(ws.DateTimeLayout))
		influx := whitesourceExecuteScanInflux{}
		// test
		_, err := checkAndReportScanResults(ctx, config, scan, utils, system, &influx)
		// assert
		assert.NoError(t, err)
		vPath := filepath.Join(ws.ReportsDirectory, "mock-project-vulnerability-report.txt")
		assert.False(t, utils.HasWrittenFile(vPath))
		rPath := filepath.Join(ws.ReportsDirectory, "mock-project-risk-report.pdf")
		assert.False(t, utils.HasWrittenFile(rPath))
	})
	t.Run("check vulnerabilities - invalid limit", func(t *testing.T) {
		ctx := context.Background()
		// init
		config := &ScanOptions{
			SecurityVulnerabilities: true,
			CvssSeverityLimit:       "invalid",
		}
		scan := newWhitesourceScan(config)
		utils := newWhitesourceUtilsMock()
		system := ws.NewSystemMock(time.Now().Format(ws.DateTimeLayout))
		influx := whitesourceExecuteScanInflux{}
		// test
		_, err := checkAndReportScanResults(ctx, config, scan, utils, system, &influx)
		// assert
		assert.EqualError(t, err, "failed to parse parameter cvssSeverityLimit (invalid) as floating point number: strconv.ParseFloat: parsing \"invalid\": invalid syntax")
	})
	t.Run("check vulnerabilities - limit not hit", func(t *testing.T) {
		ctx := context.Background()
		// init
		config := &ScanOptions{
			ProductToken:            "mock-product-token",
			ProjectToken:            "mock-project-token",
			Version:                 "1",
			SecurityVulnerabilities: true,
			CvssSeverityLimit:       "6.0",
		}
		scan := newWhitesourceScan(config)
		utils := newWhitesourceUtilsMock()
		system := ws.NewSystemMock(time.Now().Format(ws.DateTimeLayout))
		influx := whitesourceExecuteScanInflux{}
		// test
		_, err := checkAndReportScanResults(ctx, config, scan, utils, system, &influx)
		// assert
		assert.NoError(t, err)
	})
	t.Run("check vulnerabilities - limit exceeded", func(t *testing.T) {
		ctx := context.Background()
		// init
		config := &ScanOptions{
			ProductToken:                "mock-product-token",
			ProjectName:                 "mock-project - 1",
			ProjectToken:                "mock-project-token",
			Version:                     "1",
			SecurityVulnerabilities:     true,
			CvssSeverityLimit:           "4",
			FailOnSevereVulnerabilities: true,
		}
		scan := newWhitesourceScan(config)
		utils := newWhitesourceUtilsMock()
		system := ws.NewSystemMock(time.Now().Format(ws.DateTimeLayout))
		influx := whitesourceExecuteScanInflux{}
		// test
		_, err := checkAndReportScanResults(ctx, config, scan, utils, system, &influx)
		// assert
		assert.EqualError(t, err, "1 Open Source Software Security vulnerabilities with CVSS score greater or equal to 4.0 detected in project mock-project - 1")
	})
}

func TestResolveProjectIdentifiers(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		// init
		config := ScanOptions{
			BuildTool:           "mta",
			BuildDescriptorFile: "my-mta.yml",
			VersioningModel:     "major",
			ProductName:         "mock-product",
			M2Path:              "m2/path",
			ProjectSettingsFile: "project-settings.xml",
			GlobalSettingsFile:  "global-settings.xml",
		}
		utilsMock := newWhitesourceUtilsMock()
		systemMock := ws.NewSystemMock("ignored")
		scan := newWhitesourceScan(&config)
		// test
		err := resolveProjectIdentifiers(&config, scan, utilsMock, systemMock)
		// assert
		if assert.NoError(t, err) {
			assert.Equal(t, "mock-group-id-mock-artifact-id", scan.AggregateProjectName)
			assert.Equal(t, "1", config.Version)
			assert.Equal(t, "mock-product-token", config.ProductToken)
			assert.Equal(t, "mta", utilsMock.usedBuildTool)
			assert.Equal(t, "my-mta.yml", utilsMock.usedBuildDescriptorFile)
			assert.Equal(t, "project-settings.xml", utilsMock.usedOptions.ProjectSettingsFile)
			assert.Equal(t, "global-settings.xml", utilsMock.usedOptions.GlobalSettingsFile)
			assert.Equal(t, "m2/path", utilsMock.usedOptions.M2Path)
		}
	})
	t.Run("success - with version from default", func(t *testing.T) {
		// init
		config := ScanOptions{
			BuildTool:           "mta",
			BuildDescriptorFile: "my-mta.yml",
			Version:             "1.2.3-20200101",
			VersioningModel:     "major",
			ProductName:         "mock-product",
			M2Path:              "m2/path",
			ProjectSettingsFile: "project-settings.xml",
			GlobalSettingsFile:  "global-settings.xml",
		}
		utilsMock := newWhitesourceUtilsMock()
		systemMock := ws.NewSystemMock("ignored")
		scan := newWhitesourceScan(&config)
		// test
		err := resolveProjectIdentifiers(&config, scan, utilsMock, systemMock)
		// assert
		if assert.NoError(t, err) {
			assert.Equal(t, "mock-group-id-mock-artifact-id", scan.AggregateProjectName)
			assert.Equal(t, "1", config.Version)
			assert.Equal(t, "mock-product-token", config.ProductToken)
			assert.Equal(t, "mta", utilsMock.usedBuildTool)
			assert.Equal(t, "my-mta.yml", utilsMock.usedBuildDescriptorFile)
			assert.Equal(t, "project-settings.xml", utilsMock.usedOptions.ProjectSettingsFile)
			assert.Equal(t, "global-settings.xml", utilsMock.usedOptions.GlobalSettingsFile)
			assert.Equal(t, "m2/path", utilsMock.usedOptions.M2Path)
		}
	})
	t.Run("success - with custom scan version", func(t *testing.T) {
		// init
		config := ScanOptions{
			BuildTool:           "mta",
			BuildDescriptorFile: "my-mta.yml",
			CustomScanVersion:   "2.3.4",
			VersioningModel:     "major",
			ProductName:         "mock-product",
			M2Path:              "m2/path",
			ProjectSettingsFile: "project-settings.xml",
			GlobalSettingsFile:  "global-settings.xml",
		}
		utilsMock := newWhitesourceUtilsMock()
		systemMock := ws.NewSystemMock("ignored")
		scan := newWhitesourceScan(&config)
		// test
		err := resolveProjectIdentifiers(&config, scan, utilsMock, systemMock)
		// assert
		if assert.NoError(t, err) {
			assert.Equal(t, "mock-group-id-mock-artifact-id", scan.AggregateProjectName)
			assert.Equal(t, "2.3.4", config.Version)
			assert.Equal(t, "mock-product-token", config.ProductToken)
			assert.Equal(t, "mta", utilsMock.usedBuildTool)
			assert.Equal(t, "my-mta.yml", utilsMock.usedBuildDescriptorFile)
			assert.Equal(t, "project-settings.xml", utilsMock.usedOptions.ProjectSettingsFile)
			assert.Equal(t, "global-settings.xml", utilsMock.usedOptions.GlobalSettingsFile)
			assert.Equal(t, "m2/path", utilsMock.usedOptions.M2Path)
		}
	})
	t.Run("success - with custom scan version (projectName is filled)", func(t *testing.T) {
		// init
		config := ScanOptions{
			BuildTool:         "mta",
			CustomScanVersion: "latest",
			VersioningModel:   "major",
			ProductName:       "mock-product",
			ProjectName:       "mock-project",
			Version:           "0.0.1",
		}
		utilsMock := newWhitesourceUtilsMock()
		systemMock := ws.NewSystemMock("ignored")
		scan := newWhitesourceScan(&config)
		// test
		err := resolveProjectIdentifiers(&config, scan, utilsMock, systemMock)
		// assert
		if assert.NoError(t, err) {
			assert.Equal(t, "mock-project", scan.AggregateProjectName)
			assert.Equal(t, "latest", config.Version)
			assert.Equal(t, "mock-product-token", config.ProductToken)
		}
	})
	t.Run("success - with version from default (projectName is filled)", func(t *testing.T) {
		// init
		config := ScanOptions{
			BuildTool:       "mta",
			VersioningModel: "major-minor",
			ProductName:     "mock-product",
			ProjectName:     "mock-project",
			Version:         "1.2.3",
		}
		utilsMock := newWhitesourceUtilsMock()
		systemMock := ws.NewSystemMock("ignored")
		scan := newWhitesourceScan(&config)
		// test
		err := resolveProjectIdentifiers(&config, scan, utilsMock, systemMock)
		// assert
		if assert.NoError(t, err) {
			assert.Equal(t, "mock-project", scan.AggregateProjectName)
			assert.Equal(t, "1.2", config.Version)
			assert.Equal(t, "mock-product-token", config.ProductToken)
		}
	})
	t.Run("retrieves token for configured project name", func(t *testing.T) {
		// init
		config := ScanOptions{
			BuildTool:           "mta",
			BuildDescriptorFile: "my-mta.yml",
			VersioningModel:     "major",
			ProductName:         "mock-product",
			ProjectName:         "mock-project",
		}
		utilsMock := newWhitesourceUtilsMock()
		systemMock := ws.NewSystemMock("ignored")
		scan := newWhitesourceScan(&config)
		// test
		err := resolveProjectIdentifiers(&config, scan, utilsMock, systemMock)
		// assert
		if assert.NoError(t, err) {
			assert.Equal(t, "mock-project", scan.AggregateProjectName)
			assert.Equal(t, "1", config.Version)
			assert.Equal(t, "mock-product-token", config.ProductToken)
			assert.Equal(t, "mta", utilsMock.usedBuildTool)
			assert.Equal(t, "my-mta.yml", utilsMock.usedBuildDescriptorFile)
			assert.Equal(t, "mock-project-token", config.ProjectToken)
		}
	})
	t.Run("product not found", func(t *testing.T) {
		// init
		config := ScanOptions{
			BuildTool:       "mta",
			VersioningModel: "major",
			ProductName:     "does-not-exist",
		}
		utilsMock := newWhitesourceUtilsMock()
		systemMock := ws.NewSystemMock("ignored")
		scan := newWhitesourceScan(&config)
		// test
		err := resolveProjectIdentifiers(&config, scan, utilsMock, systemMock)
		// assert
		assert.EqualError(t, err, "error resolving product token: failed to get product by name: no product with name 'does-not-exist' found in Whitesource")
	})
	t.Run("product not found, created from pipeline", func(t *testing.T) {
		// init
		config := ScanOptions{
			BuildTool:                            "mta",
			CreateProductFromPipeline:            true,
			EmailAddressesOfInitialProductAdmins: []string{"user1@domain.org", "user2@domain.org"},
			VersioningModel:                      "major",
			ProductName:                          "created-by-pipeline",
		}
		utilsMock := newWhitesourceUtilsMock()
		systemMock := ws.NewSystemMock("ignored")
		scan := newWhitesourceScan(&config)
		// test
		err := resolveProjectIdentifiers(&config, scan, utilsMock, systemMock)
		// assert
		assert.NoError(t, err)
		assert.Len(t, systemMock.Products, 2)
		assert.Equal(t, "created-by-pipeline", systemMock.Products[1].Name)
		assert.Equal(t, "mock-product-token-1", config.ProductToken)
	})
}

func TestCheckPolicyViolations(t *testing.T) {
	t.Parallel()

	t.Run("success - no violations", func(t *testing.T) {
		ctx := context.Background()
		config := ScanOptions{ProductName: "mock-product", Version: "1"}
		scan := newWhitesourceScan(&config)
		if err := scan.AppendScannedProject("testProject1"); err != nil {
			t.Fail()
		}
		systemMock := ws.NewSystemMock("ignored")
		systemMock.Alerts = []ws.Alert{}
		utilsMock := newWhitesourceUtilsMock()
		reportPaths := []piperutils.Path{
			{Target: filepath.Join("whitesource", "report1.pdf")},
			{Target: filepath.Join("whitesource", "report2.pdf")},
		}
		influx := whitesourceExecuteScanInflux{}

		path, err := checkPolicyViolations(ctx, &config, scan, systemMock, utilsMock, reportPaths, &influx)
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(ws.ReportsDirectory, "whitesource-ip.json"), path.Target)

		fileContent, _ := utilsMock.FileRead(path.Target)
		content := string(fileContent)
		assert.Contains(t, content, `"policyViolations":0`)
		assert.Contains(t, content, `"reports":["report1.pdf","report2.pdf"]`)

		exists, err := utilsMock.FileExists(filepath.Join(reporting.StepReportDirectory, "whitesourceExecuteScan_ip_2d3120020f3f46393a54575a7f6f5675ad536721.json"))
		assert.True(t, exists)
	})

	t.Run("success - no reports", func(t *testing.T) {
		ctx := context.Background()
		config := ScanOptions{}
		scan := newWhitesourceScan(&config)
		if err := scan.AppendScannedProject("testProject1"); err != nil {
			t.Fail()
		}
		systemMock := ws.NewSystemMock("ignored")
		systemMock.Alerts = []ws.Alert{}
		utilsMock := newWhitesourceUtilsMock()
		reportPaths := []piperutils.Path{}
		influx := whitesourceExecuteScanInflux{}

		path, err := checkPolicyViolations(ctx, &config, scan, systemMock, utilsMock, reportPaths, &influx)
		assert.NoError(t, err)

		fileContent, _ := utilsMock.FileRead(path.Target)
		content := string(fileContent)
		assert.Contains(t, content, `reports":[]`)
	})

	t.Run("error - policy violations", func(t *testing.T) {
		ctx := context.Background()
		config := ScanOptions{FailOnSevereVulnerabilities: true}
		scan := newWhitesourceScan(&config)
		if err := scan.AppendScannedProject("testProject1"); err != nil {
			t.Fail()
		}
		systemMock := ws.NewSystemMock("ignored")
		systemMock.Alerts = []ws.Alert{
			{Vulnerability: ws.Vulnerability{Name: "policyVul1"}},
			{Vulnerability: ws.Vulnerability{Name: "policyVul2"}},
		}
		utilsMock := newWhitesourceUtilsMock()
		reportPaths := []piperutils.Path{
			{Target: "report1.pdf"},
			{Target: "report2.pdf"},
		}
		influx := whitesourceExecuteScanInflux{}

		path, err := checkPolicyViolations(ctx, &config, scan, systemMock, utilsMock, reportPaths, &influx)
		assert.Contains(t, fmt.Sprint(err), "2 policy violation(s) found")

		fileContent, _ := utilsMock.FileRead(path.Target)
		content := string(fileContent)
		assert.Contains(t, content, `"policyViolations":2`)
		assert.Contains(t, content, `"reports":["report1.pdf","report2.pdf"]`)
	})

	t.Run("error - get alerts", func(t *testing.T) {
		ctx := context.Background()
		config := ScanOptions{}
		scan := newWhitesourceScan(&config)
		if err := scan.AppendScannedProject("testProject1"); err != nil {
			t.Fail()
		}
		systemMock := ws.NewSystemMock("ignored")
		systemMock.AlertError = fmt.Errorf("failed to read alerts")
		utilsMock := newWhitesourceUtilsMock()
		reportPaths := []piperutils.Path{}
		influx := whitesourceExecuteScanInflux{}

		_, err := checkPolicyViolations(ctx, &config, scan, systemMock, utilsMock, reportPaths, &influx)
		assert.Contains(t, fmt.Sprint(err), "failed to retrieve project policy alerts from WhiteSource")
	})

	t.Run("error - write file", func(t *testing.T) {
		ctx := context.Background()
		config := ScanOptions{}
		scan := newWhitesourceScan(&config)
		if err := scan.AppendScannedProject("testProject1"); err != nil {
			t.Fail()
		}
		systemMock := ws.NewSystemMock("ignored")
		systemMock.Alerts = []ws.Alert{}
		utilsMock := newWhitesourceUtilsMock()
		utilsMock.FileWriteError = fmt.Errorf("failed to write file")
		reportPaths := []piperutils.Path{}
		influx := whitesourceExecuteScanInflux{}

		_, err := checkPolicyViolations(ctx, &config, scan, systemMock, utilsMock, reportPaths, &influx)
		assert.Contains(t, fmt.Sprint(err), "failed to write policy violation report:")
	})

	t.Run("failed to write json report", func(t *testing.T) {
		ctx := context.Background()
		config := ScanOptions{ProductName: "mock-product", Version: "1"}
		scan := newWhitesourceScan(&config)
		if err := scan.AppendScannedProject("testProject1"); err != nil {
			t.Fail()
		}
		systemMock := ws.NewSystemMock("ignored")
		systemMock.Alerts = []ws.Alert{}
		utilsMock := newWhitesourceUtilsMock()
		utilsMock.FileWriteErrors = map[string]error{
			filepath.Join(reporting.StepReportDirectory, "whitesourceExecuteScan_ip_2d3120020f3f46393a54575a7f6f5675ad536721.json"): fmt.Errorf("write error"),
		}
		reportPaths := []piperutils.Path{}
		influx := whitesourceExecuteScanInflux{}

		_, err := checkPolicyViolations(ctx, &config, scan, systemMock, utilsMock, reportPaths, &influx)
		assert.Contains(t, fmt.Sprint(err), "failed to write json report")
	})
}

func TestCheckSecurityViolations(t *testing.T) {
	t.Parallel()

	t.Run("success - non-aggregated", func(t *testing.T) {
		ctx := context.Background()
		config := ScanOptions{
			CvssSeverityLimit: "7",
		}
		scan := newWhitesourceScan(&config)
		if err := scan.AppendScannedProject("testProject1"); err != nil {
			t.Fail()
		}
		systemMock := ws.NewSystemMock("ignored")
		systemMock.Alerts = []ws.Alert{
			{Vulnerability: ws.Vulnerability{Name: "vul1", CVSS3Score: 6.0}},
		}
		utilsMock := newWhitesourceUtilsMock()
		influx := whitesourceExecuteScanInflux{}

		reportPaths, err := checkSecurityViolations(ctx, &config, scan, systemMock, utilsMock, &influx)
		assert.NoError(t, err)
		fileContent, err := utilsMock.FileRead(reportPaths[0].Target)
		assert.NoError(t, err)
		assert.True(t, len(fileContent) > 0)
	})

	t.Run("success - aggregated", func(t *testing.T) {
		ctx := context.Background()
		config := ScanOptions{
			CvssSeverityLimit: "7",
			ProjectToken:      "theProjectToken",
		}
		scan := newWhitesourceScan(&config)
		systemMock := ws.NewSystemMock("ignored")
		systemMock.Alerts = []ws.Alert{
			{Vulnerability: ws.Vulnerability{Name: "vul1", CVSS3Score: 6.0}},
		}
		utilsMock := newWhitesourceUtilsMock()
		influx := whitesourceExecuteScanInflux{}

		reportPaths, err := checkSecurityViolations(ctx, &config, scan, systemMock, utilsMock, &influx)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(reportPaths))
	})

	t.Run("error - wrong limit", func(t *testing.T) {
		ctx := context.Background()
		config := ScanOptions{CvssSeverityLimit: "x"}
		scan := newWhitesourceScan(&config)
		systemMock := ws.NewSystemMock("ignored")
		utilsMock := newWhitesourceUtilsMock()
		influx := whitesourceExecuteScanInflux{}

		_, err := checkSecurityViolations(ctx, &config, scan, systemMock, utilsMock, &influx)
		assert.Contains(t, fmt.Sprint(err), "failed to parse parameter cvssSeverityLimit")
	})

	t.Run("error - non-aggregated", func(t *testing.T) {
		ctx := context.Background()
		config := ScanOptions{
			CvssSeverityLimit:           "5",
			FailOnSevereVulnerabilities: true,
		}
		scan := newWhitesourceScan(&config)
		if err := scan.AppendScannedProject("testProject1"); err != nil {
			t.Fail()
		}
		systemMock := ws.NewSystemMock("ignored")
		systemMock.Alerts = []ws.Alert{
			{Vulnerability: ws.Vulnerability{Name: "vul1", CVSS3Score: 6.0}},
		}
		utilsMock := newWhitesourceUtilsMock()
		influx := whitesourceExecuteScanInflux{}

		reportPaths, err := checkSecurityViolations(ctx, &config, scan, systemMock, utilsMock, &influx)
		assert.Contains(t, fmt.Sprint(err), "1 Open Source Software Security vulnerabilities")
		fileContent, err := utilsMock.FileRead(reportPaths[0].Target)
		assert.NoError(t, err)
		assert.True(t, len(fileContent) > 0)
	})

	t.Run("error - aggregated", func(t *testing.T) {
		ctx := context.Background()
		config := ScanOptions{
			CvssSeverityLimit:           "5",
			ProjectToken:                "theProjectToken",
			FailOnSevereVulnerabilities: true,
		}
		scan := newWhitesourceScan(&config)
		systemMock := ws.NewSystemMock("ignored")
		systemMock.Alerts = []ws.Alert{
			{Vulnerability: ws.Vulnerability{Name: "vul1", CVSS3Score: 6.0}},
		}
		utilsMock := newWhitesourceUtilsMock()
		influx := whitesourceExecuteScanInflux{}

		reportPaths, err := checkSecurityViolations(ctx, &config, scan, systemMock, utilsMock, &influx)
		assert.Contains(t, fmt.Sprint(err), "1 Open Source Software Security vulnerabilities")
		assert.Equal(t, 3, len(reportPaths))
	})
}

func TestCheckProjectSecurityViolations(t *testing.T) {
	project := ws.Project{Name: "testProject - 1", Token: "testToken"}

	t.Run("success - no alerts", func(t *testing.T) {
		systemMock := ws.NewSystemMock("ignored")
		systemMock.Alerts = []ws.Alert{}
		influx := whitesourceExecuteScanInflux{}

		severeVulnerabilities, alerts, assessedAlerts, err := checkProjectSecurityViolations(&ScanOptions{FailOnSevereVulnerabilities: true}, 7.0, project, systemMock, &[]format.Assessment{}, &influx)
		assert.NoError(t, err)
		assert.Equal(t, 0, severeVulnerabilities)
		assert.Equal(t, 0, len(alerts))
		assert.Equal(t, 0, len(assessedAlerts))
	})

	t.Run("error - some vulnerabilities", func(t *testing.T) {
		systemMock := ws.NewSystemMock("ignored")
		systemMock.Alerts = []ws.Alert{
			{Vulnerability: ws.Vulnerability{CVSS3Score: 7, Name: "CVE-2025-001"}, Library: ws.Library{KeyID: 42, Name: "test", GroupID: "com.sap", ArtifactID: "test", Version: "1.2.3", LibType: "Java"}},
			{Vulnerability: ws.Vulnerability{CVSS3Score: 6, Name: "CVE-2025-002"}, Library: ws.Library{KeyID: 42, Name: "test", GroupID: "com.sap", ArtifactID: "test", Version: "1.2.3", LibType: "Java"}},
		}
		influx := whitesourceExecuteScanInflux{}

		severeVulnerabilities, alerts, assessedAlerts, err := checkProjectSecurityViolations(&ScanOptions{FailOnSevereVulnerabilities: true}, 7.0, project, systemMock, &[]format.Assessment{}, &influx)
		assert.Contains(t, fmt.Sprint(err), "1 Open Source Software Security vulnerabilities")
		assert.Equal(t, 1, severeVulnerabilities)
		assert.Equal(t, 2, len(alerts))
		assert.Equal(t, 0, len(assessedAlerts))
	})

	t.Run("success - assessed vulnerabilities", func(t *testing.T) {
		systemMock := ws.NewSystemMock("ignored")
		systemMock.Alerts = []ws.Alert{
			{Vulnerability: ws.Vulnerability{CVSS3Score: 7.8, Name: "CVE-2025-001"}, Library: ws.Library{KeyID: 42, Name: "test", GroupID: "com.sap", ArtifactID: "test", Version: "1.2.3", LibType: "Java"}},
			{Vulnerability: ws.Vulnerability{CVSS3Score: 6, Name: "CVE-2025-002"}, Library: ws.Library{KeyID: 42, Name: "test", GroupID: "com.sap", ArtifactID: "test", Version: "1.2.3", LibType: "Java"}},
		}
		influx := whitesourceExecuteScanInflux{}

		severeVulnerabilities, alerts, assessedAlerts, err := checkProjectSecurityViolations(&ScanOptions{FailOnSevereVulnerabilities: true}, 7.0, project, systemMock, &[]format.Assessment{{Vulnerability: "CVE-2025-001", Purls: []format.Purl{{Purl: "pkg:/maven/com.sap/test@1.2.3"}}}, {Vulnerability: "CVE-2025-002", Purls: []format.Purl{{Purl: "pkg:/maven/com.sap/test@1.2.3"}}}}, &influx)
		assert.NoError(t, err)
		assert.Equal(t, 0, severeVulnerabilities)
		assert.Equal(t, 0, len(alerts))
		assert.Equal(t, 2, len(assessedAlerts))
	})

	t.Run("error - WhiteSource failure", func(t *testing.T) {
		systemMock := ws.NewSystemMock("ignored")
		systemMock.AlertError = fmt.Errorf("failed to read alerts")
		influx := whitesourceExecuteScanInflux{}

		_, _, _, err := checkProjectSecurityViolations(&ScanOptions{FailOnSevereVulnerabilities: true}, 7.0, project, systemMock, &[]format.Assessment{}, &influx)
		assert.Contains(t, fmt.Sprint(err), "failed to retrieve project alerts from WhiteSource")
	})
}

func TestAggregateVersionWideLibraries(t *testing.T) {
	t.Parallel()
	t.Run("happy path", func(t *testing.T) {
		// init
		config := &ScanOptions{
			FailOnSevereVulnerabilities: true,
			ProductToken:                "mock-product-token",
			Version:                     "1",
		}
		utils := newWhitesourceUtilsMock()
		system := ws.NewSystemMock("2010-05-30 00:15:00 +0100")
		// test
		err := aggregateVersionWideLibraries(config, utils, system)
		// assert
		resource := filepath.Join(ws.ReportsDirectory, "libraries-20100510-001542.csv")
		if assert.NoError(t, err) && assert.True(t, utils.HasWrittenFile(resource)) {
			contents, _ := utils.FileRead(resource)
			asString := string(contents)
			assert.Equal(t, "Library Name, Project Name\nmock-library, mock-project\n", asString)
			c, _ := utils.ReadFile("/whitesourceExecuteScan_reports.json")
			assert.NotEmpty(t, c)
		}
	})
}

func TestAggregateVersionWideVulnerabilities(t *testing.T) {
	t.Parallel()
	t.Run("happy path", func(t *testing.T) {
		// init
		config := &ScanOptions{
			ProductToken: "mock-product-token",
			Version:      "1",
		}
		utils := newWhitesourceUtilsMock()
		system := ws.NewSystemMock("2010-05-30 00:15:00 +0100")
		// test
		err := aggregateVersionWideVulnerabilities(config, utils, system)
		// assert
		resource := filepath.Join(ws.ReportsDirectory, "project-names-aggregated.txt")
		assert.NoError(t, err)
		if assert.True(t, utils.HasWrittenFile(resource)) {
			contents, _ := utils.FileRead(resource)
			asString := string(contents)
			assert.Equal(t, "mock-project - 1\n", asString)
		}
		reportSheet := filepath.Join(ws.ReportsDirectory, "vulnerabilities-20100510-001542.xlsx")
		sheetContents, err := utils.FileRead(reportSheet)
		assert.NoError(t, err)
		assert.NotEmpty(t, sheetContents)
		c, _ := utils.ReadFile("whitesourceExecuteScan_reports.json")
		assert.NotEmpty(t, c)
	})
}

func TestPersistScannedProjects(t *testing.T) {
	t.Parallel()
	t.Run("write 1 scanned projects", func(t *testing.T) {
		// init
		cpe := whitesourceExecuteScanCommonPipelineEnvironment{}
		config := &ScanOptions{Version: "1"}
		scan := newWhitesourceScan(config)
		_ = scan.AppendScannedProject("project")
		// test
		persistScannedProjects(config, scan, &cpe)
		// assert
		assert.Equal(t, []string{"project - 1"}, cpe.custom.whitesourceProjectNames)
	})
	t.Run("write 2 scanned projects", func(t *testing.T) {
		// init
		cpe := whitesourceExecuteScanCommonPipelineEnvironment{}
		config := &ScanOptions{Version: "1"}
		scan := newWhitesourceScan(config)
		_ = scan.AppendScannedProject("project-app")
		_ = scan.AppendScannedProject("project-db")
		// test
		persistScannedProjects(config, scan, &cpe)
		// assert
		assert.Equal(t, []string{"project-app - 1", "project-db - 1"}, cpe.custom.whitesourceProjectNames)
	})
	t.Run("write no projects", func(t *testing.T) {
		// init
		cpe := whitesourceExecuteScanCommonPipelineEnvironment{}
		config := &ScanOptions{Version: "1"}
		scan := newWhitesourceScan(config)
		// test
		persistScannedProjects(config, scan, &cpe)
		// assert
		assert.Equal(t, []string{}, cpe.custom.whitesourceProjectNames)
	})
	t.Run("write aggregated project", func(t *testing.T) {
		// init
		cpe := whitesourceExecuteScanCommonPipelineEnvironment{}
		config := &ScanOptions{ProjectName: "project", Version: "1"}
		scan := newWhitesourceScan(config)
		// test
		persistScannedProjects(config, scan, &cpe)
		// assert
		assert.Equal(t, []string{"project - 1"}, cpe.custom.whitesourceProjectNames)
	})
}

func TestBuildToolFiles(t *testing.T) {
	t.Parallel()
	t.Run("buildTool = dub", func(t *testing.T) {
		err := validationBuildDescriptorFile("dub", "/home/mta.yaml")
		assert.ErrorContains(t, err, "extension of buildDescriptorFile must be in '*.json'")
		err = validationBuildDescriptorFile("dub", "/home/dub.json")
		assert.NoError(t, err)
	})
	t.Run("buildTool = gradle", func(t *testing.T) {
		err := validationBuildDescriptorFile("gradle", "/home/go.mod")
		assert.ErrorContains(t, err, "extension of buildDescriptorFile must be in '*.properties'")
		err = validationBuildDescriptorFile("gradle", "/home/gradle.properties")
		assert.NoError(t, err)
	})
	t.Run("buildTool = golang", func(t *testing.T) {
		err := validationBuildDescriptorFile("golang", "/home/go.json")
		assert.ErrorContains(t, err, "buildDescriptorFile must be one of  [\"go.mod\",\"VERSION\", \"version.txt\"]")
		err = validationBuildDescriptorFile("golang", "/home/go.mod")
		assert.NoError(t, err)
		err = validationBuildDescriptorFile("golang", "/home/VERSION")
		assert.NoError(t, err)
		err = validationBuildDescriptorFile("golang", "/home/version.txt")
		assert.NoError(t, err)
	})
	t.Run("buildTool = maven", func(t *testing.T) {
		err := validationBuildDescriptorFile("maven", "/home/go.mod")
		assert.ErrorContains(t, err, "extension of buildDescriptorFile must be in '*.xml'")
		err = validationBuildDescriptorFile("maven", "/home/pom.xml")
		assert.NoError(t, err)
	})
	t.Run("buildTool = mta", func(t *testing.T) {
		err := validationBuildDescriptorFile("mta", "/home/go.mod")
		assert.ErrorContains(t, err, "extension of buildDescriptorFile must be in '*.yaml'")
		err = validationBuildDescriptorFile("mta", "/home/mta.yaml")
		assert.NoError(t, err)
	})
	t.Run("buildTool = npm", func(t *testing.T) {
		err := validationBuildDescriptorFile("npm", "/home/go.mod")
		assert.ErrorContains(t, err, "extension of buildDescriptorFile must be in '*.json'")
		err = validationBuildDescriptorFile("npm", "/home/package.json")
		assert.NoError(t, err)
	})
	t.Run("buildTool = yarn", func(t *testing.T) {
		err := validationBuildDescriptorFile("yarn", "/home/go.mod")
		assert.ErrorContains(t, err, "extension of buildDescriptorFile must be in '*.json'")
		err = validationBuildDescriptorFile("yarn", "/home/package.json")
		assert.NoError(t, err)
	})
	t.Run("buildTool = pip", func(t *testing.T) {
		err := validationBuildDescriptorFile("pip", "/home/go.mod")
		assert.ErrorContains(t, err, "buildDescriptorFile must be one of  [\"setup.py\",\"version.txt\", \"VERSION\"]")
		err = validationBuildDescriptorFile("pip", "/home/setup.py")
		assert.NoError(t, err)
		err = validationBuildDescriptorFile("pip", "/home/version.txt")
		assert.NoError(t, err)
		err = validationBuildDescriptorFile("pip", "/home/VERSION")
		assert.NoError(t, err)
	})
	t.Run("buildTool = sbt", func(t *testing.T) {
		err := validationBuildDescriptorFile("sbt", "/home/go.mod")
		assert.ErrorContains(t, err, "extension of buildDescriptorFile must be in '*.json'")
		err = validationBuildDescriptorFile("sbt", "/home/sbtDescriptor.json")
		assert.NoError(t, err)
		err = validationBuildDescriptorFile("sbt", "/home/build.sbt")
		assert.NoError(t, err)
	})
}
