//go:build unit
// +build unit

package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/bmatcuk/doublestar"
	"github.com/jarcoal/httpmock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	SonarUtils "github.com/SAP/jenkins-library/pkg/sonar"
)

// TODO: extract to mock package
type mockDownloader struct {
	shouldFail    bool
	requestedURL  []string
	requestedFile []string
}

func (m *mockDownloader) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	m.requestedURL = append(m.requestedURL, url)
	m.requestedFile = append(m.requestedFile, filename)
	if m.shouldFail {
		return errors.New("something happened")
	}
	return nil
}

func (m *mockDownloader) SetOptions(options piperHttp.ClientOptions) {}

func mockFileUtilsExists(exists bool) func(string) (bool, error) {
	return func(filename string) (bool, error) {
		if exists {
			return true, nil
		}
		return false, errors.New("something happened")
	}
}

func mockExecLookPath(executable string) (string, error) {
	if executable == "local-sonar-scanner" {
		return "/usr/bin/sonar-scanner", nil
	}
	return "", errors.New("something happened")
}

func mockFileUtilsUnzip(t *testing.T, expectSrc string) func(string, string) ([]string, error) {
	return func(src, dest string) ([]string, error) {
		assert.Equal(t, filepath.Join(dest, expectSrc), src)
		return []string{}, nil
	}
}

func mockOsRename(t *testing.T, expectOld, expectNew string) func(string, string) error {
	return func(old, new string) error {
		assert.Regexp(t, expectOld, old)
		assert.Equal(t, expectNew, new)
		return nil
	}
}

func mockOsStat(exists map[string]bool) func(name string) (os.FileInfo, error) {
	return func(name string) (os.FileInfo, error) {
		_, exists := exists[name]
		if exists {
			// Exploits the fact that FileInfo result from os.Stat() is ignored anyway
			return nil, nil
		}
		return nil, errors.New("something happened")
	}
}

func mockGlob(matchesForPatterns map[string][]string) func(pattern string) ([]string, error) {
	return func(pattern string) ([]string, error) {
		matches, exists := matchesForPatterns[pattern]
		if exists {
			return matches, nil
		}
		return nil, errors.New("something happened")
	}
}

func createTaskReportFile(t *testing.T, workingDir string) {
	require.NoError(t, os.MkdirAll(filepath.Join(workingDir, ".scannerwork"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(workingDir, ".scannerwork", "report-task.txt"), []byte(taskReportContent), 0o755))
	require.FileExists(t, filepath.Join(workingDir, ".scannerwork", "report-task.txt"))
}

const sonarServerURL = "https://sonarcloud.io"

const taskReportContent = `
projectKey=piper-test
serverUrl=` + sonarServerURL + `
serverVersion=8.0.0.12345
dashboardUrl=` + sonarServerURL + `/dashboard/index/piper-test
ceTaskId=AXERR2JBbm9IiM5TEST
ceTaskUrl=` + sonarServerURL + `/api/ce/task?id=AXERR2JBbm9IiMTEST
`

const measuresComponentResponse = `
{
	"component": {
	  "key": "com.sap.piper.test",
	  "name": "com.sap.piper.test",
	  "qualifier": "TRK",
	  "measures": [
		{
		  "metric": "line_coverage",
		  "value": "80.4",
		  "bestValue": false
		},
		{
		  "metric": "branch_coverage",
		  "value": "81.0",
		  "bestValue": false
		},
		{
		  "metric": "coverage",
		  "value": "80.7",
		  "bestValue": false
		},
		{
		  "metric": "extra_valie",
		  "value": "42.7",
		  "bestValue": false
		}
	  ]
	}
  }
`

func TestRunSonar(t *testing.T) {
	mockRunner := mock.ExecMockRunner{}
	mockDownloadClient := mockDownloader{shouldFail: false}
	apiClient := &piperHttp.Client{}
	apiClient.SetOptions(piperHttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
	// mock SonarQube API calls
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	// add response handler
	httpmock.RegisterResponder(http.MethodGet, sonarServerURL+"/api/"+SonarUtils.EndpointCeTask+"", httpmock.NewStringResponder(http.StatusOK, `{ "task": { "componentId": "AXERR2JBbm9IiM5TEST", "status": "SUCCESS" }}`))
	httpmock.RegisterResponder(http.MethodGet, sonarServerURL+"/api/"+SonarUtils.EndpointIssuesSearch+"", httpmock.NewStringResponder(http.StatusOK, `{ "total": 0 }`))
	httpmock.RegisterResponder(http.MethodGet, sonarServerURL+"/api/"+SonarUtils.EndpointMeasuresComponent+"", httpmock.NewStringResponder(http.StatusOK, measuresComponentResponse))

	t.Run("default", func(t *testing.T) {
		// init
		tmpFolder := t.TempDir()
		createTaskReportFile(t, tmpFolder)

		sonar = sonarSettings{
			workingDir:  tmpFolder,
			binary:      "sonar-scanner",
			environment: []string{},
			options:     []string{},
		}
		options := sonarExecuteScanOptions{
			CustomTLSCertificateLinks: []string{},
			Token:                     "secret-ABC",
			ServerURL:                 sonarServerURL,
			Organization:              "SAP",
			Version:                   "1.2.3",
			VersioningModel:           "major",
			PullRequestProvider:       "GitHub",
		}
		fileUtilsExists = mockFileUtilsExists(true)
		os.Setenv("SONAR_SCANNER_OPTS", "-Xmx42m")
		defer os.Setenv("SONAR_SCANNER_OPTS", "")
		// test
		err := runSonar(options, &mockDownloadClient, &mockRunner, apiClient, &mock.FilesMock{}, &sonarExecuteScanInflux{})
		assert.NoError(t, err)
		// load sonarscan report file
		reportFile, err := os.ReadFile(filepath.Join(tmpFolder, "sonarscan.json"))
		assert.NoError(t, err)
		var reportData SonarUtils.ReportData
		err = json.Unmarshal(reportFile, &reportData)
		assert.NoError(t, err)
		// assert
		assert.NotNil(t, reportData.Errors)
		assert.Contains(t, sonar.options, "-Dsonar.projectVersion=1")
		assert.Contains(t, sonar.options, "-Dsonar.organization=SAP")
		assert.Contains(t, sonar.environment, "SONAR_HOST_URL="+sonarServerURL)
		assert.Contains(t, sonar.environment, "SONAR_TOKEN=secret-ABC")
		assert.Contains(t, sonar.environment, "SONAR_SCANNER_OPTS=-Xmx42m -Djavax.net.ssl.trustStore="+filepath.Join(getWorkingDir(), ".certificates", "cacerts")+" -Djavax.net.ssl.trustStorePassword=changeit")
	})
	t.Run("with custom options", func(t *testing.T) {
		// init
		tmpFolder := t.TempDir()
		createTaskReportFile(t, tmpFolder)

		sonar = sonarSettings{
			workingDir:  tmpFolder,
			binary:      "sonar-scanner",
			environment: []string{},
			options:     []string{},
		}
		options := sonarExecuteScanOptions{
			Options:             []string{"-Dsonar.projectKey=piper"},
			PullRequestProvider: "GitHub",
		}
		fileUtilsExists = mockFileUtilsExists(true)
		defer func() {
			fileUtilsExists = piperutils.FileExists
		}()
		// test
		err := runSonar(options, &mockDownloadClient, &mockRunner, apiClient, &mock.FilesMock{}, &sonarExecuteScanInflux{})
		// assert
		assert.NoError(t, err)
		assert.Contains(t, sonar.options, "-Dsonar.projectKey=piper")
	})
	t.Run("with binaries option", func(t *testing.T) {
		// init
		tmpFolder := t.TempDir()
		createTaskReportFile(t, tmpFolder)

		sonar = sonarSettings{
			workingDir:  tmpFolder,
			binary:      "sonar-scanner",
			environment: []string{},
			options:     []string{},
		}
		fileUtilsExists = mockFileUtilsExists(true)

		globMatches := make(map[string][]string)
		globMatches[pomXMLPattern] = []string{"pom.xml", "application/pom.xml"}
		doublestarGlob = mockGlob(globMatches)

		existsMap := make(map[string]bool)
		existsMap[filepath.Join("target", "classes")] = true
		existsMap[filepath.Join("target", "test-classes")] = true
		existsMap[filepath.Join("application", "target", "classes")] = true
		osStat = mockOsStat(existsMap)

		defer func() {
			fileUtilsExists = piperutils.FileExists
			doublestarGlob = doublestar.Glob
			osStat = os.Stat
		}()
		options := sonarExecuteScanOptions{
			InferJavaBinaries:   true,
			PullRequestProvider: "GitHub",
		}
		// test
		err := runSonar(options, &mockDownloadClient, &mockRunner, apiClient, &mock.FilesMock{}, &sonarExecuteScanInflux{})
		// assert
		assert.NoError(t, err)
		assert.Contains(t, sonar.options, fmt.Sprintf("-Dsonar.java.binaries=%s,%s,%s",
			filepath.Join("target", "classes"),
			filepath.Join("target", "test-classes"),
			filepath.Join("application", "target", "classes")))
	})
	t.Run("with binaries option already given", func(t *testing.T) {
		// init
		tmpFolder := t.TempDir()
		createTaskReportFile(t, tmpFolder)

		sonar = sonarSettings{
			workingDir:  tmpFolder,
			binary:      "sonar-scanner",
			environment: []string{},
			options:     []string{},
		}
		fileUtilsExists = mockFileUtilsExists(true)

		globMatches := make(map[string][]string)
		globMatches[pomXMLPattern] = []string{"pom.xml"}
		doublestarGlob = mockGlob(globMatches)

		existsMap := make(map[string]bool)
		existsMap[filepath.Join("target", "classes")] = true
		osStat = mockOsStat(existsMap)

		defer func() {
			fileUtilsExists = piperutils.FileExists
			doublestarGlob = doublestar.Glob
			osStat = os.Stat
		}()
		options := sonarExecuteScanOptions{
			Options:             []string{"-Dsonar.java.binaries=user/provided"},
			InferJavaBinaries:   true,
			PullRequestProvider: "GitHub",
		}
		// test
		err := runSonar(options, &mockDownloadClient, &mockRunner, apiClient, &mock.FilesMock{}, &sonarExecuteScanInflux{})
		// assert
		assert.NoError(t, err)
		assert.NotContains(t, sonar.options, fmt.Sprintf("-Dsonar.java.binaries=%s",
			filepath.Join("target", "classes")))
		assert.Contains(t, sonar.options, "-Dsonar.java.binaries=user/provided")
	})
	t.Run("projectKey, coverageExclusions, m2Path, verbose", func(t *testing.T) {
		// init
		tmpFolder := t.TempDir()
		createTaskReportFile(t, tmpFolder)

		sonar = sonarSettings{
			workingDir:  tmpFolder,
			binary:      "sonar-scanner",
			environment: []string{},
			options:     []string{},
		}
		options := sonarExecuteScanOptions{
			ProjectKey:          "mock-project-key",
			M2Path:              "my/custom/m2", // assumed to be resolved via alias from mavenExecute
			InferJavaLibraries:  true,
			CoverageExclusions:  []string{"one", "**/two", "three**"},
			PullRequestProvider: "GitHub",
		}
		GeneralConfig.Verbose = true
		defer func() { GeneralConfig.Verbose = false }()
		fileUtilsExists = mockFileUtilsExists(true)
		defer func() {
			fileUtilsExists = piperutils.FileExists
		}()
		// test
		err := runSonar(options, &mockDownloadClient, &mockRunner, apiClient, &mock.FilesMock{}, &sonarExecuteScanInflux{})
		// assert
		assert.NoError(t, err)
		assert.Contains(t, sonar.options, "-Dsonar.projectKey=mock-project-key")
		assert.Contains(t, sonar.options, fmt.Sprintf("-Dsonar.java.libraries=%s",
			filepath.Join("my/custom/m2", "**")))
		assert.Contains(t, sonar.options, "-Dsonar.coverage.exclusions=one,**/two,three**")
		assert.Contains(t, sonar.options, "-Dsonar.verbose=true")
	})
}

func TestSonarHandlePullRequest(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		// init
		sonar = sonarSettings{
			binary:      "sonar-scanner",
			environment: []string{},
			options:     []string{},
		}
		options := sonarExecuteScanOptions{
			ChangeID:            "123",
			PullRequestProvider: "GitHub",
			ChangeBranch:        "feat/bogus",
			ChangeTarget:        "master",
			Owner:               "SAP",
			Repository:          "jenkins-library",
		}
		// test
		err := handlePullRequest(options)
		// assert
		assert.NoError(t, err)
		//assert.Contains(t, sonar.options, "sonar.pullrequest.key=123")
		//assert.Contains(t, sonar.options, "sonar.pullrequest.provider=github")
		//assert.Contains(t, sonar.options, "sonar.pullrequest.base=master")
		//assert.Contains(t, sonar.options, "sonar.pullrequest.branch=feat/bogus")
		//assert.Contains(t, sonar.options, "sonar.pullrequest.github.repository=SAP/jenkins-library")
	})
	t.Run("unsupported scm provider", func(t *testing.T) {
		// init
		sonar = sonarSettings{
			binary:      "sonar-scanner",
			environment: []string{},
			options:     []string{},
		}
		options := sonarExecuteScanOptions{
			ChangeID:            "123",
			PullRequestProvider: "Gerrit",
		}
		// test
		err := handlePullRequest(options)
		// assert
		assert.Error(t, err)
		assert.Equal(t, "Pull-Request provider 'gerrit' is not supported!", err.Error())
	})
	t.Run("legacy", func(t *testing.T) {
		// init
		sonar = sonarSettings{
			binary:      "sonar-scanner",
			environment: []string{},
			options:     []string{},
		}
		options := sonarExecuteScanOptions{
			LegacyPRHandling:      true,
			ChangeID:              "123",
			Owner:                 "SAP",
			Repository:            "jenkins-library",
			GithubToken:           "some-token",
			DisableInlineComments: true,
		}
		// test
		err := handlePullRequest(options)
		// assert
		assert.NoError(t, err)
		assert.Contains(t, sonar.options, "sonar.analysis.mode=preview")
		assert.Contains(t, sonar.options, "sonar.github.pullRequest=123")
		assert.Contains(t, sonar.options, "sonar.github.oauth=some-token")
		assert.Contains(t, sonar.options, "sonar.github.repository=SAP/jenkins-library")
		assert.Contains(t, sonar.options, "sonar.github.disableInlineComments=true")
	})
}

func TestSonarLoadScanner(t *testing.T) {
	mockClient := mockDownloader{shouldFail: false}

	t.Run("use preinstalled sonar-scanner", func(t *testing.T) {
		// init
		ignore := ""
		sonar = sonarSettings{
			binary:      "local-sonar-scanner",
			environment: []string{},
			options:     []string{},
		}
		execLookPath = mockExecLookPath
		defer func() { execLookPath = exec.LookPath }()
		// test
		err := loadSonarScanner(ignore, &mockClient)
		// assert
		assert.NoError(t, err)
		assert.Equal(t, "local-sonar-scanner", sonar.binary)
	})

	t.Run("use downloaded sonar-scanner", func(t *testing.T) {
		// init
		url := "https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/sonar-scanner-cli-4.6.2.2472-linux.zip"
		sonar = sonarSettings{
			binary:      "sonar-scanner",
			environment: []string{},
			options:     []string{},
		}
		execLookPath = mockExecLookPath
		fileUtilsUnzip = mockFileUtilsUnzip(t, "sonar-scanner-cli-4.6.2.2472-linux.zip")
		osRename = mockOsRename(t, "sonar-scanner-4.6.2.2472-linux", ".sonar-scanner")
		defer func() {
			execLookPath = exec.LookPath
			fileUtilsUnzip = piperutils.Unzip
			osRename = os.Rename
		}()
		// test
		err := loadSonarScanner(url, &mockClient)
		// assert
		assert.NoError(t, err)
		assert.Equal(t, url, mockClient.requestedURL[0])
		assert.Regexp(t, "sonar-scanner-cli-4.6.2.2472-linux.zip$", mockClient.requestedFile[0])
		assert.Equal(t, filepath.Join(getWorkingDir(), ".sonar-scanner", "bin", "sonar-scanner"), sonar.binary)
	})
}

func TestSonarLoadCertificates(t *testing.T) {
	mockRunner := mock.ExecMockRunner{}
	mockClient := mockDownloader{shouldFail: false}

	t.Run("use local trust store", func(t *testing.T) {
		// init
		sonar = sonarSettings{
			binary:      "sonar-scanner",
			environment: []string{},
			options:     []string{},
		}
		fileUtilsExists = mockFileUtilsExists(true)
		defer func() { fileUtilsExists = piperutils.FileExists }()
		defer os.Setenv("SONAR_SCANNER_OPTS", "")
		// test
		err := loadCertificates([]string{}, &mockClient, &mockRunner)
		// assert
		assert.NoError(t, err)
		assert.Contains(t, sonar.environment, "SONAR_SCANNER_OPTS=-Djavax.net.ssl.trustStore="+filepath.Join(getWorkingDir(), ".certificates", "cacerts")+" -Djavax.net.ssl.trustStorePassword=changeit")
	})

	t.Run("use local trust store with downloaded certificates", func(t *testing.T) {
		// init
		sonar = sonarSettings{
			binary:      "sonar-scanner",
			environment: []string{},
			options:     []string{},
		}
		fileUtilsExists = mockFileUtilsExists(false)
		defer os.Setenv("SONAR_SCANNER_OPTS", "")
		// test
		err := loadCertificates([]string{"https://sap.com/custom-1.crt", "https://sap.com/custom-2.crt"}, &mockClient, &mockRunner)
		// assert
		assert.NoError(t, err)
		assert.Equal(t, "https://sap.com/custom-1.crt", mockClient.requestedURL[0])
		assert.Equal(t, "https://sap.com/custom-2.crt", mockClient.requestedURL[1])
		assert.Regexp(t, "custom-1.crt$", mockClient.requestedFile[0])
		assert.Regexp(t, "custom-2.crt$", mockClient.requestedFile[1])
		assert.Contains(t, sonar.environment, "SONAR_SCANNER_OPTS=-Djavax.net.ssl.trustStore="+filepath.Join(getWorkingDir(), ".certificates", "cacerts")+" -Djavax.net.ssl.trustStorePassword=changeit")
	})

	t.Run("use no trust store", func(t *testing.T) {
		// init
		sonar = sonarSettings{
			binary:      "sonar-scanner",
			environment: []string{},
			options:     []string{},
		}
		fileUtilsExists = mockFileUtilsExists(false)
		// test
		err := loadCertificates([]string{}, &mockClient, &mockRunner)
		// assert
		assert.NoError(t, err)
		assert.Empty(t, sonar.environment)
	})
}
