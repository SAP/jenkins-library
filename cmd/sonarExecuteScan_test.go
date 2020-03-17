package cmd

import (
	"net/http"
	"os"
	"path"
	"testing"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"
	FileUtils "github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type mockDownloader struct {
	shouldFail     bool
	requestedUrls  []string
	requestedFiles []string
}

func (m *mockDownloader) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	m.requestedUrls = append(m.requestedUrls, url)
	m.requestedFiles = append(m.requestedFiles, filename)
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
		workingDir, _ := os.Getwd()
		assert.Contains(t, sonar.Environment, "SONAR_SCANNER_OPTS=-Djavax.net.ssl.trustStore="+path.Join(workingDir, ".certificates", "cacerts"))
	})
}
