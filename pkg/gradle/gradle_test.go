package gradle

import (
	"github.com/SAP/jenkins-library/pkg/mock"

	"testing"

	"github.com/stretchr/testify/assert"
)

type MockUtils struct {
	shouldFail     bool
	requestedUrls  []string
	requestedFiles []string
	*mock.FilesMock
	*mock.ExecMockRunner
}

func NewMockUtils(downloadShouldFail bool) MockUtils {
	utils := MockUtils{
		shouldFail:     downloadShouldFail,
		FilesMock:      &mock.FilesMock{},
		ExecMockRunner: &mock.ExecMockRunner{},
	}
	return utils
}

func (f MockUtils) FileExists(filePath string) (bool, error) {
	switch filePath {
	case "build.gradle":
		return true, nil
	case "path/to/build.gradle":
		return true, nil
	}
	return false, nil
}

func TestExecute(t *testing.T) {
	t.Run("success - gradle build", func(t *testing.T) {
		utils := NewMockUtils(false)
		opts := ExecuteOptions{
			BuildGradlePath: "path/to",
			Task:            "build",
			CreateBOM:       false,
		}

		err := Execute(&opts, &utils)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(utils.Calls))
		assert.Equal(t, mock.ExecCall{Exec: "gradle", Params: []string{"build", "-p", "path/to"}}, utils.Calls[0])
	})

	t.Run("success - bom creation", func(t *testing.T) {
		utils := NewMockUtils(false)
		opts := ExecuteOptions{
			Task:      "build",
			CreateBOM: true,
		}

		err := Execute(&opts, &utils)
		assert.NoError(t, err)

		assert.Equal(t, 3, len(utils.Calls))
		assert.Equal(t, mock.ExecCall{Exec: "gradle", Params: []string{"tasks"}}, utils.Calls[0])
		assert.Equal(t, mock.ExecCall{Exec: "gradle", Params: []string{"--init-script", "cyclonedx.gradle", "cyclonedxBom"}}, utils.Calls[1])
		assert.Equal(t, mock.ExecCall{Exec: "gradle", Params: []string{"build"}}, utils.Calls[2])
	})
}
