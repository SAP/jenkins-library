package maven

import (
	"fmt"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
)

func TestSettings(t *testing.T) {

	defer func() {
		getenv = os.Getenv
	}()

	getenv = func(name string) string {
		if name == "M2_HOME" {
			return "/usr/share/maven"
		} else if name == "HOME" {
			return "/home/me"
		}
		return ""
	}

	t.Run("Settings file source location not provided", func(t *testing.T) {

		httpClient := httpMock{}
		fileUtils := fileUtilsMock{}

		err := downloadAndCopySettingsFile("", "foo", &fileUtils, &httpClient)

		assert.EqualError(t, err, "Settings file source location not provided")
	})

	t.Run("Settings file destination location not provided", func(t *testing.T) {

		httpClient := httpMock{}
		fileUtils := fileUtilsMock{}

		err := downloadAndCopySettingsFile("/opt/sap/maven/global-settings.xml", "", &fileUtils, &httpClient)

		assert.EqualError(t, err, "Settings file destination location not provided")
	})

	t.Run("Retrieve settings files", func(t *testing.T) {

		httpClient := httpMock{}
		fileUtils := fileUtilsMock{existingFiles: map[string]string{
			"/opt/sap/maven/global-settings.xml":  "",
			"/opt/sap/maven/project-settings.xml": "",
		}}

		err := DownloadAndCopySettingsFiles("/opt/sap/maven/global-settings.xml", "/opt/sap/maven/project-settings.xml", &fileUtils, &httpClient)

		if assert.NoError(t, err) {
			assert.Equal(t, "/usr/share/maven/conf/settings.xml", fileUtils.copiedFiles["/opt/sap/maven/global-settings.xml"])
			assert.Equal(t, "/home/me/.m2/settings.xml", fileUtils.copiedFiles["/opt/sap/maven/project-settings.xml"])
		}

		assert.Empty(t, httpClient.downloadedFiles)
	})

	t.Run("Retrieve settings file via http", func(t *testing.T) {

		httpClient := httpMock{}
		fileUtils := fileUtilsMock{}

		err := downloadAndCopySettingsFile("https://example.org/maven/global-settings.xml", "/usr/share/maven/conf/settings.xml", &fileUtils, &httpClient)

		if assert.NoError(t, err) {
			assert.Equal(t, "/usr/share/maven/conf/settings.xml", httpClient.downloadedFiles["https://example.org/maven/global-settings.xml"])
		}
	})

	t.Run("Retrieve settings file via http - received error from downloader", func(t *testing.T) {

		httpClient := httpMock{expectedError: fmt.Errorf("Download failed")}
		fileUtils := fileUtilsMock{}

		err := downloadAndCopySettingsFile("https://example.org/maven/global-settings.xml", "/usr/share/maven/conf/settings.xml", &fileUtils, &httpClient)

		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "failed to download maven settings from URL")
		}
	})

	t.Run("Retrieve project settings file - file not found", func(t *testing.T) {

		httpClient := httpMock{}
		fileUtils := fileUtilsMock{}

		err := downloadAndCopySettingsFile("/opt/sap/maven/project-settings.xml", "/home/me/.m2/settings.xml", &fileUtils, &httpClient)

		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "Source file '/opt/sap/maven/project-settings.xml' does not exist")
		}
	})
}

type httpMock struct {
	expectedError   error
	downloadedFiles map[string]string // src, dest
}

func (c *httpMock) SetOptions(options piperhttp.ClientOptions) {
}

func (c *httpMock) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {

	if c.expectedError != nil {
		return c.expectedError
	}

	if c.downloadedFiles == nil {
		c.downloadedFiles = make(map[string]string)
	}
	c.downloadedFiles[url] = filename
	return nil
}

type fileUtilsMock struct {
	existingFiles map[string]string
	writtenFiles  map[string]string
	copiedFiles   map[string]string
}

func (f *fileUtilsMock) FileExists(path string) (bool, error) {

	if _, ok := f.existingFiles[path]; ok {
		return true, nil
	}
	return false, nil
}

func (f *fileUtilsMock) Copy(src, dest string) (int64, error) {

	exists, err := f.FileExists(src)

	if err != nil {
		return 0, err
	}

	if !exists {
		return 0, fmt.Errorf("Source file '"+src+"' does not exist", src)
	}

	if f.copiedFiles == nil {
		f.copiedFiles = make(map[string]string)
	}
	f.copiedFiles[src] = dest

	return 0, nil
}

func (f *fileUtilsMock) FileRead(path string) ([]byte, error) {
	return []byte(f.existingFiles[path]), nil
}

func (f *fileUtilsMock) FileWrite(path string, content []byte, perm os.FileMode) error {

	if f.writtenFiles == nil {
		f.writtenFiles = make(map[string]string)
	}

	if _, ok := f.writtenFiles[path]; ok {
		delete(f.writtenFiles, path)
	}
	f.writtenFiles[path] = string(content)
	return nil
}

func (f *fileUtilsMock) MkdirAll(path string, perm os.FileMode) error {
	return nil
}

func (f *fileUtilsMock) Chmod(path string, mode os.FileMode) error {
	return fmt.Errorf("not implemented. func is only present in order to fullfil the interface contract. Needs to be ajusted in case it gets used.")
}

func (f *fileUtilsMock) Abs(path string) (string, error) {
	return "", fmt.Errorf("not implemented. func is only present in order to fullfil the interface contract. Needs to be ajusted in case it gets used.")
}

func (f *fileUtilsMock) Glob(pattern string) (matches []string, err error) {
	return nil, fmt.Errorf("not implemented. func is only present in order to fullfil the interface contract. Needs to be ajusted in case it gets used.")
}
