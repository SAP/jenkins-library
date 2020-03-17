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
)

type mockDownloader struct {
	shouldFail    bool
	requestedURL  string
	requestedFile string
}

func (m *mockDownloader) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	m.requestedURL = url
	m.requestedFile = filename
	if m.shouldFail {
		return errors.New("something happened")
	}
	return nil
}

func (m *mockDownloader) SetOptions(options piperHttp.ClientOptions) {
	return
}

func mockFileUtilsExists(filename string) (bool, error) {
	return true, nil
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

func mockOsRemove(t *testing.T) func(string) error {
	return func(filename string) error {
		//TODO: tempDir name not known here
		//assert.Equal(t, "", filename)
		return nil
	}
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
		handlePullRequest(options)
		// assert
		assert.Contains(t, sonar.Options, "sonar.pullrequest.key=123")
		assert.Contains(t, sonar.Options, "sonar.pullrequest.provider=GitHub")
		assert.Contains(t, sonar.Options, "sonar.pullrequest.base=master")
		assert.Contains(t, sonar.Options, "sonar.pullrequest.branch=feat/bogus")
		assert.Contains(t, sonar.Options, "sonar.pullrequest.github.repository=SAP/jenkins-library")
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
		loadSonarScanner(ignore, &mockClient)
		// assert
		assert.Equal(t, "local-sonar-scanner", sonar.Binary)
	})

	t.Run("download sonar-scanner from url", func(t *testing.T) {
		// init
		url := "https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/sonar-scanner-cli-4.3.0.2102-linux.zip"
		sonar = sonarSettings{
			Binary:      "sonar-scanner",
			Environment: []string{},
			Options:     []string{},
		}
		execLookPath = mockExecLookPath
		defer func() { execLookPath = exec.LookPath }()
		fileUtilsUnzip = mockFileUtilsUnzip(t, "sonar-scanner-cli-4.3.0.2102-linux.zip")
		defer func() { fileUtilsUnzip = FileUtils.Unzip }()
		osRename = mockOsRename(t, "sonar-scanner-4.3.0.2102-linux", ".sonar-scanner")
		defer func() { osRename = os.Rename }()
		osRemove = mockOsRemove(t)
		defer func() { osRemove = os.Remove }()
		// test
		loadSonarScanner(url, &mockClient)
		// assert
		assert.Equal(t, url, mockClient.requestedURL)
		assert.Regexp(t, "sonar-scanner-cli-4.3.0.2102-linux.zip$", mockClient.requestedFile)
		assert.Equal(t, path.Join(getWorkingDir(), ".sonar-scanner", "bin", "sonar-scanner"), sonar.Binary)
	})
}

func TestSonarLoadCertificates(t *testing.T) {
	mockRunner := mock.ExecMockRunner{}
	mockClient := mockDownloader{shouldFail: false}

	t.Run("use custom trust store", func(t *testing.T) {
		// init
		sonar = sonarSettings{
			Binary:      "sonar-scanner",
			Environment: []string{},
			Options:     []string{},
		}
		fileUtilsExists = mockFileUtilsExists
		defer func() { fileUtilsExists = FileUtils.FileExists }()
		// test
		loadCertificates(&mockRunner, "", &mockClient)
		// assert
		assert.Contains(t, sonar.Environment, "SONAR_SCANNER_OPTS=-Djavax.net.ssl.trustStore="+path.Join(getWorkingDir(), ".certificates", "cacerts"))
	})
}
