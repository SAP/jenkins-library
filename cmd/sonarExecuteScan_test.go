package cmd

import (
	"net/http"
	"os"
	"os/exec"
	"path"
	"testing"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"
	FileUtils "github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func (m *mockDownloader) SetOptions(options piperHttp.ClientOptions) {
	return
}

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
		assert.Equal(t, path.Join(dest, expectSrc), src)
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

func TestRunSonar(t *testing.T) {
	mockRunner := mock.ExecMockRunner{}
	mockClient := mockDownloader{shouldFail: false}

	t.Run("default", func(t *testing.T) {
		// init
		sonar = sonarSettings{
			Binary:      "sonar-scanner",
			Environment: []string{},
			Options:     []string{},
		}
		options := sonarExecuteScanOptions{
			CustomTLSCertificateLinks: "",
			Token:                     "secret-ABC",
			Host:                      "https://sonar.sap.com",
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
		err := runSonar(options, &mockClient, &mockRunner)
		// assert
		assert.NoError(t, err)
		assert.Contains(t, sonar.Options, "-Dsonar.projectVersion=1.2.3")
		assert.Contains(t, sonar.Options, "-Dsonar.organization=SAP")
		assert.Contains(t, sonar.Environment, "SONAR_HOST_URL=https://sonar.sap.com")
		assert.Contains(t, sonar.Environment, "SONAR_AUTH_TOKEN=secret-ABC")
		assert.Contains(t, sonar.Environment, "SONAR_SCANNER_OPTS=-Djavax.net.ssl.trustStore="+path.Join(getWorkingDir(), ".certificates", "cacerts"))
	})
}

func TestSonarHandlePullRequest(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		// init
		sonar = sonarSettings{
			Binary:      "sonar-scanner",
			Environment: []string{},
			Options:     []string{},
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
		assert.Contains(t, sonar.Options, "sonar.pullrequest.key=123")
		assert.Contains(t, sonar.Options, "sonar.pullrequest.provider=github")
		assert.Contains(t, sonar.Options, "sonar.pullrequest.base=master")
		assert.Contains(t, sonar.Options, "sonar.pullrequest.branch=feat/bogus")
		assert.Contains(t, sonar.Options, "sonar.pullrequest.github.repository=SAP/jenkins-library")
	})
	t.Run("unsupported scm provider", func(t *testing.T) {
		// init
		sonar = sonarSettings{
			Binary:      "sonar-scanner",
			Environment: []string{},
			Options:     []string{},
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
			Binary:      "sonar-scanner",
			Environment: []string{},
			Options:     []string{},
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
		assert.Contains(t, sonar.Options, "sonar.analysis.mode=preview")
		assert.Contains(t, sonar.Options, "sonar.github.pullRequest=123")
		assert.Contains(t, sonar.Options, "sonar.github.oauth=some-token")
		assert.Contains(t, sonar.Options, "sonar.github.repository=SAP/jenkins-library")
		assert.Contains(t, sonar.Options, "sonar.github.disableInlineComments=true")
	})
}

func TestSonarLoadScanner(t *testing.T) {
	mockClient := mockDownloader{shouldFail: false}

	t.Run("use preinstalled sonar-scanner", func(t *testing.T) {
		// init
		ignore := ""
		sonar = sonarSettings{
			Binary:      "local-sonar-scanner",
			Environment: []string{},
			Options:     []string{},
		}
		execLookPath = mockExecLookPath
		defer func() { execLookPath = exec.LookPath }()
		// test
		err := loadSonarScanner(ignore, &mockClient)
		// assert
		assert.NoError(t, err)
		assert.Equal(t, "local-sonar-scanner", sonar.Binary)
	})

	t.Run("use downloaded sonar-scanner", func(t *testing.T) {
		// init
		url := "https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/sonar-scanner-cli-4.3.0.2102-linux.zip"
		sonar = sonarSettings{
			Binary:      "sonar-scanner",
			Environment: []string{},
			Options:     []string{},
		}
		execLookPath = mockExecLookPath
		fileUtilsUnzip = mockFileUtilsUnzip(t, "sonar-scanner-cli-4.3.0.2102-linux.zip")
		osRename = mockOsRename(t, "sonar-scanner-4.3.0.2102-linux", ".sonar-scanner")
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
		assert.Regexp(t, "sonar-scanner-cli-4.3.0.2102-linux.zip$", mockClient.requestedFile[0])
		assert.Equal(t, path.Join(getWorkingDir(), ".sonar-scanner", "bin", "sonar-scanner"), sonar.Binary)
	})
}

func TestSonarLoadCertificates(t *testing.T) {
	mockRunner := mock.ExecMockRunner{}
	mockClient := mockDownloader{shouldFail: false}

	t.Run("use local trust store", func(t *testing.T) {
		// init
		sonar = sonarSettings{
			Binary:      "sonar-scanner",
			Environment: []string{},
			Options:     []string{},
		}
		fileUtilsExists = mockFileUtilsExists(true)
		defer func() { fileUtilsExists = FileUtils.FileExists }()
		// test
		err := loadCertificates("", &mockClient, &mockRunner)
		// assert
		assert.NoError(t, err)
		assert.Contains(t, sonar.Environment, "SONAR_SCANNER_OPTS=-Djavax.net.ssl.trustStore="+path.Join(getWorkingDir(), ".certificates", "cacerts"))
	})

	t.Run("use local trust store with downloaded certificates", func(t *testing.T) {
		// init
		sonar = sonarSettings{
			Binary:      "sonar-scanner",
			Environment: []string{},
			Options:     []string{},
		}
		fileUtilsExists = mockFileUtilsExists(false)
		os.Setenv("PIPER_SONAR_LOAD_CERTIFICATES", "true")
		require.Equal(t, "true", os.Getenv("PIPER_SONAR_LOAD_CERTIFICATES"), "PIPER_SONAR_LOAD_CERTIFICATES must be set")
		defer func() {
			fileUtilsExists = FileUtils.FileExists
			os.Unsetenv("PIPER_SONAR_LOAD_CERTIFICATES")
		}()
		// test
		err := loadCertificates("https://sap.com/custom-1.crt,https://sap.com/custom-2.crt", &mockClient, &mockRunner)
		// assert
		assert.NoError(t, err)
		assert.Equal(t, "https://sap.com/custom-1.crt", mockClient.requestedURL[0])
		assert.Equal(t, "https://sap.com/custom-2.crt", mockClient.requestedURL[1])
		assert.Regexp(t, "custom-1.crt$", mockClient.requestedFile[0])
		assert.Regexp(t, "custom-2.crt$", mockClient.requestedFile[1])
		assert.Contains(t, sonar.Environment, "SONAR_SCANNER_OPTS=-Djavax.net.ssl.trustStore="+path.Join(getWorkingDir(), ".certificates", "cacerts"))
	})

	t.Run("use local trust store with downloaded certificates - deactivated", func(t *testing.T) {
		// init
		sonar = sonarSettings{
			Binary:      "sonar-scanner",
			Environment: []string{},
			Options:     []string{},
		}
		fileUtilsExists = mockFileUtilsExists(false)
		require.Empty(t, os.Getenv("PIPER_SONAR_LOAD_CERTIFICATES"), "PIPER_SONAR_LOAD_CERTIFICATES must not be set")
		defer func() { fileUtilsExists = FileUtils.FileExists }()
		// test
		err := loadCertificates("any-certificate-url", &mockClient, &mockRunner)
		// assert
		assert.NoError(t, err)
		assert.NotContains(t, sonar.Environment, "SONAR_SCANNER_OPTS=-Djavax.net.ssl.trustStore="+path.Join(getWorkingDir(), ".certificates", "cacerts"))
	})

	t.Run("use no trust store", func(t *testing.T) {
		// init
		sonar = sonarSettings{
			Binary:      "sonar-scanner",
			Environment: []string{},
			Options:     []string{},
		}
		fileUtilsExists = mockFileUtilsExists(false)
		os.Setenv("PIPER_SONAR_LOAD_CERTIFICATES", "true")
		require.Equal(t, "true", os.Getenv("PIPER_SONAR_LOAD_CERTIFICATES"), "PIPER_SONAR_LOAD_CERTIFICATES must be set")
		defer func() {
			fileUtilsExists = FileUtils.FileExists
			os.Unsetenv("PIPER_SONAR_LOAD_CERTIFICATES")
		}()
		// test
		err := loadCertificates("", &mockClient, &mockRunner)
		// assert
		assert.NoError(t, err)
		assert.Empty(t, sonar.Environment)
	})
}
