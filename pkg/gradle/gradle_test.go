package gradle

import (
	"fmt"
	"net/http"
	"os"

	"github.com/SAP/jenkins-library/pkg/mock"

	"testing"

	"github.com/stretchr/testify/assert"
)

type MockUtils struct {
	writtenFiles []string
	removedFiles []string
	*mock.FilesMock
	*mock.ExecMockRunner
}

func NewMockUtils(downloadShouldFail bool) *MockUtils {
	utils := MockUtils{
		FilesMock:      &mock.FilesMock{},
		ExecMockRunner: &mock.ExecMockRunner{},
	}
	return &utils
}

func (f *MockUtils) FileExists(filePath string) (bool, error) {
	switch filePath {
	case "build.gradle":
		return true, nil
	case "path/to/build.gradle":
		return true, nil
	}
	return false, nil
}

func (f *MockUtils) FileWrite(path string, content []byte, perm os.FileMode) error {
	f.writtenFiles = append(f.writtenFiles, path)
	return nil
}

func (f *MockUtils) FileRemove(path string) error {
	f.removedFiles = append(f.removedFiles, path)
	return nil
}

func (g *MockUtils) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	return fmt.Errorf("not implemented")
}

func TestExecute(t *testing.T) {
	t.Run("success - gradle build", func(t *testing.T) {
		utils := NewMockUtils(false)
		opts := ExecuteOptions{
			BuildGradlePath: "path/to",
			Task:            "build",
			CreateBOM:       false,
		}

		err := Execute(&opts, utils)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(utils.Calls))
		assert.Equal(t, mock.ExecCall{Exec: "gradle", Params: []string{"build", "-p", "path/to"}}, utils.Calls[0])
		assert.Equal(t, []string(nil), utils.writtenFiles)
		assert.Equal(t, []string(nil), utils.removedFiles)
	})

	t.Run("success - bom creation", func(t *testing.T) {
		utils := NewMockUtils(false)
		opts := ExecuteOptions{
			Task:      "build",
			CreateBOM: true,
		}

		err := Execute(&opts, utils)
		assert.NoError(t, err)

		assert.Equal(t, 3, len(utils.Calls))
		assert.Equal(t, mock.ExecCall{Exec: "gradle", Params: []string{"tasks"}}, utils.Calls[0])
		assert.Equal(t, mock.ExecCall{Exec: "gradle", Params: []string{"--init-script", "cyclonedx.gradle", "cyclonedxBom"}}, utils.Calls[1])
		assert.Equal(t, mock.ExecCall{Exec: "gradle", Params: []string{"build"}}, utils.Calls[2])
		assert.Equal(t, []string{"cyclonedx.gradle"}, utils.writtenFiles)
		assert.Equal(t, []string{"cyclonedx.gradle"}, utils.removedFiles)
	})
}
