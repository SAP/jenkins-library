package cmd

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/reporting"
	"github.com/SAP/jenkins-library/pkg/versioning"
	ws "github.com/SAP/jenkins-library/pkg/whitesource"
	"github.com/stretchr/testify/assert"
)

type whitesourceUtilsMock struct {
	*ws.ScanUtilsMock
	coordinates             versioning.Coordinates
	usedBuildTool           string
	usedBuildDescriptorFile string
	usedOptions             versioning.Options
}

func (w *whitesourceUtilsMock) GetArtifactCoordinates(buildTool, buildDescriptorFile string,
	options *versioning.Options) (versioning.Coordinates, error) {
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
	utils := newWhitesourceUtils(&config)

	assert.NotNil(t, utils.Client)
	assert.NotNil(t, utils.Command)
	assert.NotNil(t, utils.Files)
}

func TestRunWhitesourceExecuteScan(t *testing.T) {
	t.Parallel()
	t.Run("fails for invalid configured project token", func(t *testing.T) {
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
		err := runWhitesourceExecuteScan(&config, scan, utilsMock, systemMock, &cpe, &influx)
		// assert
		assert.EqualError(t, err, "failed to resolve and aggregate project name: failed to get project by token: no project with token 'no-such-project-token' found in Whitesource")
		assert.Equal(t, "", config.ProjectName)
		assert.Equal(t, "", scan.AggregateProjectName)
	})
	t.Run("retrieves aggregate project name by configured token", func(t *testing.T) {
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
		err := runWhitesourceExecuteScan(&config, scan, utilsMock, systemMock, &cpe, &influx)
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
	})
}

func TestCheckAndReportScanResults(t *testing.T) {
	t.Parallel()
	t.Run("no reports requested", func(t *testing.T) {
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
		_, err := checkAndReportScanResults(config, scan, utils, system, &influx)
		// assert
		assert.NoError(t, err)
		vPath := filepath.Join(ws.ReportsDirectory, "mock-project-vulnerability-report.txt")
		assert.False(t, utils.HasWrittenFile(vPath))
		rPath := filepath.Join(ws.ReportsDirectory, "mock-project-risk-report.pdf")
		assert.False(t, utils.HasWrittenFile(rPath))
	})
	t.Run("check vulnerabilities - invalid limit", func(t *testing.T) {
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
		_, err := checkAndReportScanResults(config, scan, utils, system, &influx)
		// assert
		assert.EqualError(t, err, "failed to parse parameter cvssSeverityLimit (invalid) as floating point number: strconv.ParseFloat: parsing \"invalid\": invalid syntax")
	})
	t.Run("check vulnerabilities - limit not hit", func(t *testing.T) {
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
		_, err := checkAndReportScanResults(config, scan, utils, system, &influx)
		// assert
		assert.NoError(t, err)
	})
	t.Run("check vulnerabilities - limit exceeded", func(t *testing.T) {
		// init
		config := &ScanOptions{
			ProductToken:            "mock-product-token",
			ProjectName:             "mock-project - 1",
			ProjectToken:            "mock-project-token",
			Version:                 "1",
			SecurityVulnerabilities: true,
			CvssSeverityLimit:       "4",
		}
		scan := newWhitesourceScan(config)
		utils := newWhitesourceUtilsMock()
		system := ws.NewSystemMock(time.Now().Format(ws.DateTimeLayout))
		influx := whitesourceExecuteScanInflux{}
		// test
		_, err := checkAndReportScanResults(config, scan, utils, system, &influx)
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
		config := ScanOptions{ProductName: "mock-product", Version: "1"}
		scan := newWhitesourceScan(&config)
		scan.AppendScannedProject("testProject1")
		systemMock := ws.NewSystemMock("ignored")
		systemMock.Alerts = []ws.Alert{}
		utilsMock := newWhitesourceUtilsMock()
		reportPaths := []piperutils.Path{
			{Target: filepath.Join("whitesource", "report1.pdf")},
			{Target: filepath.Join("whitesource", "report2.pdf")},
		}
		influx := whitesourceExecuteScanInflux{}

		path, err := checkPolicyViolations(&config, scan, systemMock, utilsMock, reportPaths, &influx)
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
		config := ScanOptions{}
		scan := newWhitesourceScan(&config)
		scan.AppendScannedProject("testProject1")
		systemMock := ws.NewSystemMock("ignored")
		systemMock.Alerts = []ws.Alert{}
		utilsMock := newWhitesourceUtilsMock()
		reportPaths := []piperutils.Path{}
		influx := whitesourceExecuteScanInflux{}

		path, err := checkPolicyViolations(&config, scan, systemMock, utilsMock, reportPaths, &influx)
		assert.NoError(t, err)

		fileContent, _ := utilsMock.FileRead(path.Target)
		content := string(fileContent)
		assert.Contains(t, content, `reports":[]`)
	})

	t.Run("error - policy violations", func(t *testing.T) {
		config := ScanOptions{}
		scan := newWhitesourceScan(&config)
		scan.AppendScannedProject("testProject1")
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

		path, err := checkPolicyViolations(&config, scan, systemMock, utilsMock, reportPaths, &influx)
		assert.Contains(t, fmt.Sprint(err), "2 policy violation(s) found")

		fileContent, _ := utilsMock.FileRead(path.Target)
		content := string(fileContent)
		assert.Contains(t, content, `"policyViolations":2`)
		assert.Contains(t, content, `"reports":["report1.pdf","report2.pdf"]`)
	})

	t.Run("error - get alerts", func(t *testing.T) {
		config := ScanOptions{}
		scan := newWhitesourceScan(&config)
		scan.AppendScannedProject("testProject1")
		systemMock := ws.NewSystemMock("ignored")
		systemMock.AlertError = fmt.Errorf("failed to read alerts")
		utilsMock := newWhitesourceUtilsMock()
		reportPaths := []piperutils.Path{}
		influx := whitesourceExecuteScanInflux{}

		_, err := checkPolicyViolations(&config, scan, systemMock, utilsMock, reportPaths, &influx)
		assert.Contains(t, fmt.Sprint(err), "failed to retrieve project policy alerts from WhiteSource")
	})

	t.Run("error - write file", func(t *testing.T) {
		config := ScanOptions{}
		scan := newWhitesourceScan(&config)
		scan.AppendScannedProject("testProject1")
		systemMock := ws.NewSystemMock("ignored")
		systemMock.Alerts = []ws.Alert{}
		utilsMock := newWhitesourceUtilsMock()
		utilsMock.FileWriteError = fmt.Errorf("failed to write file")
		reportPaths := []piperutils.Path{}
		influx := whitesourceExecuteScanInflux{}

		_, err := checkPolicyViolations(&config, scan, systemMock, utilsMock, reportPaths, &influx)
		assert.Contains(t, fmt.Sprint(err), "failed to write policy violation report:")
	})

	t.Run("failed to write json report", func(t *testing.T) {
		config := ScanOptions{ProductName: "mock-product", Version: "1"}
		scan := newWhitesourceScan(&config)
		scan.AppendScannedProject("testProject1")
		systemMock := ws.NewSystemMock("ignored")
		systemMock.Alerts = []ws.Alert{}
		utilsMock := newWhitesourceUtilsMock()
		utilsMock.FileWriteErrors = map[string]error{
			filepath.Join(reporting.StepReportDirectory, "whitesourceExecuteScan_ip_2d3120020f3f46393a54575a7f6f5675ad536721.json"): fmt.Errorf("write error"),
		}
		reportPaths := []piperutils.Path{}
		influx := whitesourceExecuteScanInflux{}

		_, err := checkPolicyViolations(&config, scan, systemMock, utilsMock, reportPaths, &influx)
		assert.Contains(t, fmt.Sprint(err), "failed to write json report")
	})
}

func TestCheckSecurityViolations(t *testing.T) {
	t.Parallel()

	t.Run("success - non-aggregated", func(t *testing.T) {
		config := ScanOptions{
			CvssSeverityLimit: "7",
		}
		scan := newWhitesourceScan(&config)
		scan.AppendScannedProject("testProject1")
		systemMock := ws.NewSystemMock("ignored")
		systemMock.Alerts = []ws.Alert{
			{Vulnerability: ws.Vulnerability{Name: "vul1", CVSS3Score: 6.0}},
		}
		utilsMock := newWhitesourceUtilsMock()
		influx := whitesourceExecuteScanInflux{}

		reportPaths, err := checkSecurityViolations(&config, scan, systemMock, utilsMock, &influx)
		assert.NoError(t, err)
		fileContent, err := utilsMock.FileRead(reportPaths[0].Target)
		assert.NoError(t, err)
		assert.True(t, len(fileContent) > 0)
	})

	t.Run("success - aggregated", func(t *testing.T) {
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

		reportPaths, err := checkSecurityViolations(&config, scan, systemMock, utilsMock, &influx)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(reportPaths))
	})

	t.Run("error - wrong limit", func(t *testing.T) {
		config := ScanOptions{CvssSeverityLimit: "x"}
		scan := newWhitesourceScan(&config)
		systemMock := ws.NewSystemMock("ignored")
		utilsMock := newWhitesourceUtilsMock()
		influx := whitesourceExecuteScanInflux{}

		_, err := checkSecurityViolations(&config, scan, systemMock, utilsMock, &influx)
		assert.Contains(t, fmt.Sprint(err), "failed to parse parameter cvssSeverityLimit")

	})

	t.Run("error - non-aggregated", func(t *testing.T) {
		config := ScanOptions{
			CvssSeverityLimit: "5",
		}
		scan := newWhitesourceScan(&config)
		scan.AppendScannedProject("testProject1")
		systemMock := ws.NewSystemMock("ignored")
		systemMock.Alerts = []ws.Alert{
			{Vulnerability: ws.Vulnerability{Name: "vul1", CVSS3Score: 6.0}},
		}
		utilsMock := newWhitesourceUtilsMock()
		influx := whitesourceExecuteScanInflux{}

		reportPaths, err := checkSecurityViolations(&config, scan, systemMock, utilsMock, &influx)
		assert.Contains(t, fmt.Sprint(err), "1 Open Source Software Security vulnerabilities")
		fileContent, err := utilsMock.FileRead(reportPaths[0].Target)
		assert.NoError(t, err)
		assert.True(t, len(fileContent) > 0)
	})

	t.Run("error - aggregated", func(t *testing.T) {
		config := ScanOptions{
			CvssSeverityLimit: "5",
			ProjectToken:      "theProjectToken",
		}
		scan := newWhitesourceScan(&config)
		systemMock := ws.NewSystemMock("ignored")
		systemMock.Alerts = []ws.Alert{
			{Vulnerability: ws.Vulnerability{Name: "vul1", CVSS3Score: 6.0}},
		}
		utilsMock := newWhitesourceUtilsMock()
		influx := whitesourceExecuteScanInflux{}

		reportPaths, err := checkSecurityViolations(&config, scan, systemMock, utilsMock, &influx)
		assert.Contains(t, fmt.Sprint(err), "1 Open Source Software Security vulnerabilities")
		assert.Equal(t, 0, len(reportPaths))
	})
}

func TestCheckProjectSecurityViolations(t *testing.T) {
	project := ws.Project{Name: "testProject - 1", Token: "testToken"}

	t.Run("success - no alerts", func(t *testing.T) {
		systemMock := ws.NewSystemMock("ignored")
		systemMock.Alerts = []ws.Alert{}
		influx := whitesourceExecuteScanInflux{}

		severeVulnerabilities, alerts, err := checkProjectSecurityViolations(7.0, project, systemMock, &influx)
		assert.NoError(t, err)
		assert.Equal(t, 0, severeVulnerabilities)
		assert.Equal(t, 0, len(alerts))
	})

	t.Run("error - some vulnerabilities", func(t *testing.T) {
		systemMock := ws.NewSystemMock("ignored")
		systemMock.Alerts = []ws.Alert{
			{Vulnerability: ws.Vulnerability{CVSS3Score: 7}},
			{Vulnerability: ws.Vulnerability{CVSS3Score: 6}},
		}
		influx := whitesourceExecuteScanInflux{}

		severeVulnerabilities, alerts, err := checkProjectSecurityViolations(7.0, project, systemMock, &influx)
		assert.Contains(t, fmt.Sprint(err), "1 Open Source Software Security vulnerabilities")
		assert.Equal(t, 1, severeVulnerabilities)
		assert.Equal(t, 2, len(alerts))
	})

	t.Run("error - WhiteSource failure", func(t *testing.T) {
		systemMock := ws.NewSystemMock("ignored")
		systemMock.AlertError = fmt.Errorf("failed to read alerts")
		influx := whitesourceExecuteScanInflux{}

		_, _, err := checkProjectSecurityViolations(7.0, project, systemMock, &influx)
		assert.Contains(t, fmt.Sprint(err), "failed to retrieve project alerts from WhiteSource")
	})

}

func TestCountSecurityVulnerabilities(t *testing.T) {
	t.Parallel()

	alerts := []ws.Alert{
		{Vulnerability: ws.Vulnerability{CVSS3Score: 7.1}},
		{Vulnerability: ws.Vulnerability{CVSS3Score: 7}},
		{Vulnerability: ws.Vulnerability{CVSS3Score: 6}},
	}

	severe, nonSevere := countSecurityVulnerabilities(&alerts, 7.0)
	assert.Equal(t, 2, severe)
	assert.Equal(t, 1, nonSevere)
}

func TestIsSevereVulnerability(t *testing.T) {
	tt := []struct {
		alert    ws.Alert
		limit    float64
		expected bool
	}{
		{alert: ws.Alert{Vulnerability: ws.Vulnerability{CVSS3Score: 0}}, limit: 0, expected: true},
		{alert: ws.Alert{Vulnerability: ws.Vulnerability{CVSS3Score: 6.9, Score: 6}}, limit: 7.0, expected: false},
		{alert: ws.Alert{Vulnerability: ws.Vulnerability{CVSS3Score: 7.0, Score: 6}}, limit: 7.0, expected: true},
		{alert: ws.Alert{Vulnerability: ws.Vulnerability{CVSS3Score: 7.1, Score: 6}}, limit: 7.0, expected: true},
		{alert: ws.Alert{Vulnerability: ws.Vulnerability{CVSS3Score: 6, Score: 6.9}}, limit: 7.0, expected: false},
		{alert: ws.Alert{Vulnerability: ws.Vulnerability{CVSS3Score: 6, Score: 7.0}}, limit: 7.0, expected: false},
		{alert: ws.Alert{Vulnerability: ws.Vulnerability{CVSS3Score: 6, Score: 7.1}}, limit: 7.0, expected: false},
		{alert: ws.Alert{Vulnerability: ws.Vulnerability{Score: 6.9}}, limit: 7.0, expected: false},
		{alert: ws.Alert{Vulnerability: ws.Vulnerability{Score: 7.0}}, limit: 7.0, expected: true},
		{alert: ws.Alert{Vulnerability: ws.Vulnerability{Score: 7.1}}, limit: 7.0, expected: true},
	}

	for i, test := range tt {
		assert.Equalf(t, test.expected, isSevereVulnerability(test.alert, test.limit), "run %v failed", i)
	}
}

func TestCreateCustomVulnerabilityReport(t *testing.T) {
	t.Parallel()

	t.Run("success case", func(t *testing.T) {
		config := &ScanOptions{}
		scan := newWhitesourceScan(config)
		scan.AppendScannedProject("testProject")
		alerts := []ws.Alert{
			{Library: ws.Library{Filename: "vul1"}, Vulnerability: ws.Vulnerability{CVSS3Score: 7.0, Score: 6}},
			{Library: ws.Library{Filename: "vul2"}, Vulnerability: ws.Vulnerability{CVSS3Score: 8.0, TopFix: ws.Fix{Message: "this is the top fix"}}},
			{Library: ws.Library{Filename: "vul3"}, Vulnerability: ws.Vulnerability{Score: 6}},
		}
		utilsMock := newWhitesourceUtilsMock()

		scanReport := createCustomVulnerabilityReport(config, scan, alerts, 7.0, utilsMock)

		assert.Equal(t, "WhiteSource Security Vulnerability Report", scanReport.Title)
		assert.Equal(t, 3, len(scanReport.DetailTable.Rows))

		// assert that library info is filled and sorting has been executed
		assert.Equal(t, "vul2", scanReport.DetailTable.Rows[0].Columns[5].Content)
		assert.Equal(t, "vul1", scanReport.DetailTable.Rows[1].Columns[5].Content)
		assert.Equal(t, "vul3", scanReport.DetailTable.Rows[2].Columns[5].Content)

		// assert that CVSS version identification has been done
		assert.Equal(t, "v3", scanReport.DetailTable.Rows[0].Columns[3].Content)
		assert.Equal(t, "v3", scanReport.DetailTable.Rows[1].Columns[3].Content)
		assert.Equal(t, "v2", scanReport.DetailTable.Rows[2].Columns[3].Content)

		// assert proper rating and styling of high prio issues
		assert.Equal(t, "8", scanReport.DetailTable.Rows[0].Columns[2].Content)
		assert.Equal(t, "7", scanReport.DetailTable.Rows[1].Columns[2].Content)
		assert.Equal(t, "6", scanReport.DetailTable.Rows[2].Columns[2].Content)
		assert.Equal(t, "red-cell", scanReport.DetailTable.Rows[0].Columns[2].Style.String())
		assert.Equal(t, "red-cell", scanReport.DetailTable.Rows[1].Columns[2].Style.String())
		assert.Equal(t, "yellow-cell", scanReport.DetailTable.Rows[2].Columns[2].Style.String())

		assert.Contains(t, scanReport.DetailTable.Rows[0].Columns[10].Content, "this is the top fix")

	})
}

func TestWriteCustomVulnerabilityReports(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		config := &ScanOptions{
			ProductName: "mock-product",
		}
		scan := &ws.Scan{ProductVersion: "1"}
		scan.AppendScannedProject("project1")
		scan.AppendScannedProject("project2")

		scanReport := reporting.ScanReport{}
		utilsMock := newWhitesourceUtilsMock()

		reportPaths, err := writeCustomVulnerabilityReports(config, scan, scanReport, utilsMock)

		assert.NoError(t, err)
		assert.Equal(t, 1, len(reportPaths))

		exists, err := utilsMock.FileExists(reportPaths[0].Target)
		assert.NoError(t, err)
		assert.True(t, exists)

		exists, err = utilsMock.FileExists(filepath.Join(reporting.StepReportDirectory, "whitesourceExecuteScan_oss_27322f16a39c10c852ba6639538140a03e08e93f.json"))
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("failed to write HTML report", func(t *testing.T) {
		config := &ScanOptions{
			ProductName: "mock-product",
		}
		scan := &ws.Scan{ProductVersion: "1"}
		scanReport := reporting.ScanReport{}
		utilsMock := newWhitesourceUtilsMock()
		utilsMock.FileWriteErrors = map[string]error{
			filepath.Join(ws.ReportsDirectory, "piper_whitesource_vulnerability_report.html"): fmt.Errorf("write error"),
		}

		_, err := writeCustomVulnerabilityReports(config, scan, scanReport, utilsMock)
		assert.Contains(t, fmt.Sprint(err), "failed to write html report")
	})

	t.Run("failed to write json report", func(t *testing.T) {
		config := &ScanOptions{
			ProductName: "mock-product",
		}
		scan := &ws.Scan{ProductVersion: "1"}
		scan.AppendScannedProject("project1")
		scanReport := reporting.ScanReport{}
		utilsMock := newWhitesourceUtilsMock()
		utilsMock.FileWriteErrors = map[string]error{
			filepath.Join(reporting.StepReportDirectory, "whitesourceExecuteScan_oss_e860d3a7cc8ca3261f065773404ba43e9a0b9d5b.json"): fmt.Errorf("write error"),
		}

		_, err := writeCustomVulnerabilityReports(config, scan, scanReport, utilsMock)
		assert.Contains(t, fmt.Sprint(err), "failed to write json report")
	})

}

func TestVulnerabilityScore(t *testing.T) {
	t.Parallel()

	tt := []struct {
		alert    ws.Alert
		expected float64
	}{
		{alert: ws.Alert{Vulnerability: ws.Vulnerability{CVSS3Score: 7.0, Score: 6}}, expected: 7.0},
		{alert: ws.Alert{Vulnerability: ws.Vulnerability{CVSS3Score: 7.0}}, expected: 7.0},
		{alert: ws.Alert{Vulnerability: ws.Vulnerability{Score: 6}}, expected: 6},
	}
	for i, test := range tt {
		assert.Equalf(t, test.expected, vulnerabilityScore(test.alert), "run %v failed", i)
	}
}

func TestAggregateVersionWideLibraries(t *testing.T) {
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
		err := aggregateVersionWideLibraries(config, utils, system)
		// assert
		resource := filepath.Join(ws.ReportsDirectory, "libraries-20100510-001542.csv")
		if assert.NoError(t, err) && assert.True(t, utils.HasWrittenFile(resource)) {
			contents, _ := utils.FileRead(resource)
			asString := string(contents)
			assert.Equal(t, "Library Name, Project Name\nmock-library, mock-project\n", asString)
			assert.NotEmpty(t, piperenv.GetParameter("", "whitesourceExecuteScan_reports.json"))
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
		assert.NotEmpty(t, piperenv.GetParameter("", "whitesourceExecuteScan_reports.json"))
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
