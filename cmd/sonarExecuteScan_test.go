package cmd

import (
	"fmt"
	"github.com/bmatcuk/doublestar"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"
	FileUtils "github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//TODO: extract to mock package
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
	require.NoError(t, os.MkdirAll(filepath.Join(workingDir, ".scannerwork"), 0755))
	require.NoError(t, ioutil.WriteFile(filepath.Join(workingDir, ".scannerwork", "report-task.txt"), []byte("projectKey=piper-test\nserverUrl=https://sonarcloud.io\nserverVersion=8.0.0.12345\ndashboardUrl=https://sonarcloud.io/dashboard/index/piper-test\nceTaskId=AXERR2JBbm9IiM5TEST\nceTaskUrl=https://sonarcloud.io/api/ce/task?id=AXERR2JBbm9IiMTEST"), 0755))
	require.FileExists(t, filepath.Join(workingDir, ".scannerwork", "report-task.txt"))
}

func TestRunSonar(t *testing.T) {
	mockRunner := mock.ExecMockRunner{}
	mockClient := mockDownloader{shouldFail: false}

	t.Run("default", func(t *testing.T) {
		// init
		tmpFolder, err := ioutil.TempDir(".", "test-sonar-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpFolder)
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
			ServerURL:                 "https://sonar.sap.com",
			Organization:              "SAP",
			ProjectVersion:            "1.2.3",
		}
		fileUtilsExists = mockFileUtilsExists(true)
		os.Setenv("PIPER_SONAR_LOAD_CERTIFICATES", "true")
		require.Equal(t, "true", os.Getenv("PIPER_SONAR_LOAD_CERTIFICATES"), "PIPER_SONAR_LOAD_CERTIFICATES must be set")
		defer func() {
			fileUtilsExists = FileUtils.FileExists
			os.Unsetenv("PIPER_SONAR_LOAD_CERTIFICATES")
		}()
		// test
		err = runSonar(options, &mockClient, &mockRunner)
		// assert
		assert.NoError(t, err)
		assert.Contains(t, sonar.options, "-Dsonar.projectVersion=1.2.3")
		assert.Contains(t, sonar.options, "-Dsonar.organization=SAP")
		assert.Contains(t, sonar.environment, "SONAR_HOST_URL=https://sonar.sap.com")
		assert.Contains(t, sonar.environment, "SONAR_TOKEN=secret-ABC")
		assert.Contains(t, sonar.environment, "SONAR_SCANNER_OPTS=-Djavax.net.ssl.trustStore="+filepath.Join(getWorkingDir(), ".certificates", "cacerts")+" -Djavax.net.ssl.trustStorePassword=changeit")
		assert.FileExists(t, filepath.Join(sonar.workingDir, "sonarExecuteScan_reports.json"))
		assert.FileExists(t, filepath.Join(sonar.workingDir, "sonarExecuteScan_links.json"))
	})
	t.Run("with custom options", func(t *testing.T) {
		// init
		tmpFolder, err := ioutil.TempDir(".", "test-sonar-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpFolder)
		createTaskReportFile(t, tmpFolder)

		sonar = sonarSettings{
			workingDir:  tmpFolder,
			binary:      "sonar-scanner",
			environment: []string{},
			options:     []string{},
		}
		options := sonarExecuteScanOptions{
			Options: []string{"-Dsonar.projectKey=piper"},
		}
		fileUtilsExists = mockFileUtilsExists(true)
		defer func() {
			fileUtilsExists = FileUtils.FileExists
		}()
		// test
		err = runSonar(options, &mockClient, &mockRunner)
		// assert
		assert.NoError(t, err)
		assert.Contains(t, sonar.options, "-Dsonar.projectKey=piper")
	})
	t.Run("with binaries option", func(t *testing.T) {
		// init
		tmpFolder, err := ioutil.TempDir(".", "test-sonar-")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tmpFolder) }()
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
			fileUtilsExists = FileUtils.FileExists
			doublestarGlob = doublestar.Glob
			osStat = os.Stat
		}()
		options := sonarExecuteScanOptions{
			InferJavaBinaries: true,
		}
		// test
		err = runSonar(options, &mockClient, &mockRunner)
		// assert
		assert.NoError(t, err)
		assert.Contains(t, sonar.options, fmt.Sprintf("-Dsonar.java.binaries=%s,%s,%s",
			filepath.Join("target", "classes"),
			filepath.Join("target", "test-classes"),
			filepath.Join("application", "target", "classes")))
	})
	t.Run("with binaries option already given", func(t *testing.T) {
		// init
		tmpFolder, err := ioutil.TempDir(".", "test-sonar-")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tmpFolder) }()
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
			fileUtilsExists = FileUtils.FileExists
			doublestarGlob = doublestar.Glob
			osStat = os.Stat
		}()
		options := sonarExecuteScanOptions{
			Options:           []string{"-Dsonar.java.binaries=user/provided"},
			InferJavaBinaries: true,
		}
		// test
		err = runSonar(options, &mockClient, &mockRunner)
		// assert
		assert.NoError(t, err)
		assert.NotContains(t, sonar.options, fmt.Sprintf("-Dsonar.java.binaries=%s",
			filepath.Join("target", "classes")))
		assert.Contains(t, sonar.options, "-Dsonar.java.binaries=user/provided")
	})
	t.Run("projectKey, coverageExclusions, m2Path", func(t *testing.T) {
		// init
		tmpFolder, err := ioutil.TempDir(".", "test-sonar-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpFolder)
		createTaskReportFile(t, tmpFolder)

		sonar = sonarSettings{
			workingDir:  tmpFolder,
			binary:      "sonar-scanner",
			environment: []string{},
			options:     []string{},
		}
		options := sonarExecuteScanOptions{
			ProjectKey:         "mock-project-key",
			M2Path:             "my/custom/m2", // assumed to be resolved via alias from mavenExecute
			InferJavaLibraries: true,
			CoverageExclusions: []string{"one", "**/two", "three**"},
		}
		fileUtilsExists = mockFileUtilsExists(true)
		defer func() {
			fileUtilsExists = FileUtils.FileExists
		}()
		// test
		err = runSonar(options, &mockClient, &mockRunner)
		// assert
		assert.NoError(t, err)
		assert.Contains(t, sonar.options, "-Dsonar.projectKey=mock-project-key")
		assert.Contains(t, sonar.options, fmt.Sprintf("-Dsonar.java.libraries=%s",
			filepath.Join("my/custom/m2", "**")))
		assert.Contains(t, sonar.options, "-Dsonar.coverage.exclusions=one,**/two,three**")
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
		assert.Contains(t, sonar.options, "sonar.pullrequest.key=123")
		assert.Contains(t, sonar.options, "sonar.pullrequest.provider=github")
		assert.Contains(t, sonar.options, "sonar.pullrequest.base=master")
		assert.Contains(t, sonar.options, "sonar.pullrequest.branch=feat/bogus")
		assert.Contains(t, sonar.options, "sonar.pullrequest.github.repository=SAP/jenkins-library")
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
		url := "https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/sonar-scanner-cli-4.4.0.2170-linux.zip"
		sonar = sonarSettings{
			binary:      "sonar-scanner",
			environment: []string{},
			options:     []string{},
		}
		execLookPath = mockExecLookPath
		fileUtilsUnzip = mockFileUtilsUnzip(t, "sonar-scanner-cli-4.4.0.2170-linux.zip")
		osRename = mockOsRename(t, "sonar-scanner-4.4.0.2170-linux", ".sonar-scanner")
		defer func() {
			execLookPath = exec.LookPath
			fileUtilsUnzip = FileUtils.Unzip
			osRename = os.Rename
		}()
		// test
		err := loadSonarScanner(url, &mockClient)
		// assert
		assert.NoError(t, err)
		assert.Equal(t, url, mockClient.requestedURL[0])
		assert.Regexp(t, "sonar-scanner-cli-4.4.0.2170-linux.zip$", mockClient.requestedFile[0])
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
		defer func() { fileUtilsExists = FileUtils.FileExists }()
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
		os.Setenv("PIPER_SONAR_LOAD_CERTIFICATES", "true")
		require.Equal(t, "true", os.Getenv("PIPER_SONAR_LOAD_CERTIFICATES"), "PIPER_SONAR_LOAD_CERTIFICATES must be set")
		defer func() {
			fileUtilsExists = FileUtils.FileExists
			os.Unsetenv("PIPER_SONAR_LOAD_CERTIFICATES")
		}()
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

	t.Run("use local trust store with downloaded certificates - deactivated", func(t *testing.T) {
		// init
		sonar = sonarSettings{
			binary:      "sonar-scanner",
			environment: []string{},
			options:     []string{},
		}
		fileUtilsExists = mockFileUtilsExists(false)
		require.Empty(t, os.Getenv("PIPER_SONAR_LOAD_CERTIFICATES"), "PIPER_SONAR_LOAD_CERTIFICATES must not be set")
		defer func() { fileUtilsExists = FileUtils.FileExists }()
		// test
		err := loadCertificates([]string{"any-certificate-url"}, &mockClient, &mockRunner)
		// assert
		assert.NoError(t, err)
		assert.NotContains(t, sonar.environment, "SONAR_SCANNER_OPTS=-Djavax.net.ssl.trustStore="+filepath.Join(getWorkingDir(), ".certificates", "cacerts")+" -Djavax.net.ssl.trustStorePassword=changeit")
	})

	t.Run("use no trust store", func(t *testing.T) {
		// init
		sonar = sonarSettings{
			binary:      "sonar-scanner",
			environment: []string{},
			options:     []string{},
		}
		fileUtilsExists = mockFileUtilsExists(false)
		os.Setenv("PIPER_SONAR_LOAD_CERTIFICATES", "true")
		require.Equal(t, "true", os.Getenv("PIPER_SONAR_LOAD_CERTIFICATES"), "PIPER_SONAR_LOAD_CERTIFICATES must be set")
		defer func() {
			fileUtilsExists = FileUtils.FileExists
			os.Unsetenv("PIPER_SONAR_LOAD_CERTIFICATES")
		}()
		// test
		err := loadCertificates([]string{}, &mockClient, &mockRunner)
		// assert
		assert.NoError(t, err)
		assert.Empty(t, sonar.environment)
	})
}
