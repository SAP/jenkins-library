package maven

import (
	"fmt"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"os"
	"testing"
)

func TestSettings(t *testing.T) {

	httpClient := httpMock{StatusCode: 200}

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

	t.Run("Invalid settimgs file type", func(t *testing.T) {

		fileUtils := fileUtilsMock{}

		err := GetSettingsFile(-1, "/dev/null", &fileUtils, &httpClient)

		assert.NotNil(t, err)
		assert.Equal(t, "Invalid SettingsFileType", err.Error())
	})

	t.Run("Retrieve global settings file", func(t *testing.T) {

		fileUtils := fileUtilsMock{existingFiles: map[string]string{"/opt/sap/maven/global-settings.xml": ""}}

		err := GetSettingsFile(GlobalSettingsFile, "/opt/sap/maven/global-settings.xml", &fileUtils, &httpClient)

		assert.Nil(t, err)
		assert.Equal(t, "/usr/share/maven/conf/settings.xml", fileUtils.copiedFiles["/opt/sap/maven/global-settings.xml"])
	})

	t.Run("Retrieve project settings file", func(t *testing.T) {

		fileUtils := fileUtilsMock{existingFiles: map[string]string{"/opt/sap/maven/project-settings.xml": ""}}

		err := GetSettingsFile(ProjectSettingsFile, "/opt/sap/maven/project-settings.xml", &fileUtils, &httpClient)

		assert.Nil(t, err)
		assert.Equal(t, "/home/me/.m2/settings.xml", fileUtils.copiedFiles["/opt/sap/maven/project-settings.xml"])
	})

	t.Run("Retrieve global settings file via http", func(t *testing.T) {

		fileUtils := fileUtilsMock{}

		err := GetSettingsFile(GlobalSettingsFile, "https://example.org/maven/global-settings.xml", &fileUtils, &httpClient)

		assert.Nil(t, err)
		_, ok := fileUtils.writtenFiles["/usr/share/maven/conf/settings.xml"]
		assert.True(t, ok)
	})

	t.Run("Retrieve settings file via http with http code not found", func(t *testing.T) {

		fileUtils := fileUtilsMock{}

		err := GetSettingsFile(GlobalSettingsFile, "https://example.org/maven/global-settings.xml", &fileUtils, &httpClient)

		assert.Nil(t, err)
		_, ok := fileUtils.writtenFiles["/usr/share/maven/conf/settings.xml"]
		assert.True(t, ok)
	})

	t.Run("Retrieve project settings file via http", func(t *testing.T) {

		fileUtils := fileUtilsMock{}

		err := GetSettingsFile(ProjectSettingsFile, "https://example.org/maven/project-settings.xml", &fileUtils, &httpClient)

		assert.Nil(t, err)
		_, ok := fileUtils.writtenFiles["/home/me/.m2/settings.xml"]
		assert.True(t, ok)
	})

	t.Run("Retrieve project settings file via http invalid protocol", func(t *testing.T) {

		defer func() {
			httpClient.StatusCode = 200
		}()

		fileUtils := fileUtilsMock{}
		httpClient.StatusCode = 404

		err := GetSettingsFile(ProjectSettingsFile, "https://example.org/maven/project-settings.xml", &fileUtils, &httpClient)

		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "Got 404 reponse from download attempt")
	})

	t.Run("Retrieve project settings file - file not found", func(t *testing.T) {

		fileUtils := fileUtilsMock{}

		err := GetSettingsFile(ProjectSettingsFile, "/opt/sap/maven/project-settings.xml", &fileUtils, &httpClient)

		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "File \"/opt/sap/maven/project-settings.xml\" not found")
	})
}

type httpMock struct {
	StatusCode int
}

type httpMockResponse struct {
	StatusCode int
}

func (h *httpMockResponse) Close() error {
	return nil
}

func (h *httpMockResponse) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (h *httpMock) SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {

	res := http.Response{StatusCode: h.StatusCode}
	res.Body = &httpMockResponse{}
	return &res, nil
}

func (h *httpMock) SetOptions(options piperhttp.ClientOptions) {

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
		return 0, fmt.Errorf("File '%s' does not exist", src)
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
