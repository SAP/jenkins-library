package gradle

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/SAP/jenkins-library/pkg/mock"

	"testing"

	"github.com/stretchr/testify/assert"
)

type MockUtils struct {
	existingFiles []string
	writtenFiles  []string
	removedFiles  []string
	*mock.FilesMock
	*mock.ExecMockRunner
}

func (m *MockUtils) FileExists(filePath string) (bool, error) {
	for _, filename := range m.existingFiles {
		if filename == filePath {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockUtils) FileWrite(path string, content []byte, perm os.FileMode) error {
	m.writtenFiles = append(m.writtenFiles, path)
	return nil
}

func (m *MockUtils) FileRemove(path string) error {
	m.removedFiles = append(m.removedFiles, path)
	return nil
}

func (m *MockUtils) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	return fmt.Errorf("not implemented")
}

func TestExecute(t *testing.T) {
	t.Run("success - gradle build", func(t *testing.T) {
		utils := &MockUtils{
			FilesMock:      &mock.FilesMock{},
			ExecMockRunner: &mock.ExecMockRunner{},
			existingFiles:  []string{"path/to/build.gradle"},
		}
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

	t.Run("failed - gradle build", func(t *testing.T) {
		utils := &MockUtils{
			FilesMock: &mock.FilesMock{},
			ExecMockRunner: &mock.ExecMockRunner{
				ShouldFailOnCommand: map[string]error{"gradle build -p path/to": errors.New("failed to build")},
			},
			existingFiles: []string{"path/to/build.gradle"},
		}
		opts := ExecuteOptions{
			BuildGradlePath: "path/to",
			Task:            "build",
			CreateBOM:       false,
		}

		err := Execute(&opts, utils)
		assert.Contains(t, err.Error(), "failed to build")
	})

	t.Run("success - bom creation", func(t *testing.T) {
		utils := &MockUtils{
			FilesMock:      &mock.FilesMock{},
			ExecMockRunner: &mock.ExecMockRunner{},
			existingFiles:  []string{"build.gradle"},
		}
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

	t.Run("failed - bom creation", func(t *testing.T) {
		utils := &MockUtils{
			FilesMock: &mock.FilesMock{},
			ExecMockRunner: &mock.ExecMockRunner{
				ShouldFailOnCommand: map[string]error{"gradle --init-script cyclonedx.gradle cyclonedxBom": errors.New("failed to create BOM")},
			},
			existingFiles: []string{"build.gradle"},
		}
		opts := ExecuteOptions{
			Task:      "build",
			CreateBOM: true,
		}

		err := Execute(&opts, utils)
		assert.Contains(t, err.Error(), "failed to create BOM")
	})

	t.Run("success - publish to staging repository", func(t *testing.T) {
		utils := &MockUtils{
			FilesMock:      &mock.FilesMock{},
			ExecMockRunner: &mock.ExecMockRunner{},
			existingFiles:  []string{"build.gradle"},
		}
		opts := ExecuteOptions{
			Task:               "build",
			Publish:            true,
			RepositoryURL:      "url",
			RepositoryPassword: "password",
			RepositoryUsername: "username",
			ArtifactVersion:    "1.1.0",
		}

		err := Execute(&opts, utils)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(utils.Calls))
		assert.Equal(t, mock.ExecCall{Exec: "gradle", Params: []string{"build"}}, utils.Calls[0])
		assert.Equal(t, mock.ExecCall{Exec: "gradle", Params: []string{"--init-script", "maven-publish.gradle", "--info", "publish"}}, utils.Calls[1])
		assert.Equal(t, []string{"maven-publish.gradle"}, utils.writtenFiles)
		assert.Equal(t, []string{"maven-publish.gradle"}, utils.removedFiles)
	})

	t.Run("failed - publish to staging repository", func(t *testing.T) {
		utils := &MockUtils{
			FilesMock: &mock.FilesMock{},
			ExecMockRunner: &mock.ExecMockRunner{
				ShouldFailOnCommand: map[string]error{"gradle --init-script maven-publish.gradle --info publish": errors.New("failed to publish artifacts")},
			},
			existingFiles: []string{"build.gradle"},
		}
		opts := ExecuteOptions{
			Task:               "build",
			Publish:            true,
			RepositoryURL:      "url",
			RepositoryPassword: "password",
			RepositoryUsername: "username",
			ArtifactVersion:    "1.1.0",
		}

		err := Execute(&opts, utils)
		assert.Contains(t, err.Error(), "failed to publish artifacts")
	})
}
