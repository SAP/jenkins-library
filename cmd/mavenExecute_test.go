package cmd

import (
	"errors"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
)

type mavenMockUtils struct {
	shouldFail     bool
	requestedUrls  []string
	requestedFiles []string
	*mock.FilesMock
	*mock.ExecMockRunner
}

func (m *mavenMockUtils) DownloadFile(_, _ string, _ http.Header, _ []*http.Cookie) error {
	return errors.New("Test should not download files.")
}

func newMavenMockUtils() mavenMockUtils {
	utils := mavenMockUtils{
		shouldFail: false,
		FilesMock: &mock.FilesMock{},
		ExecMockRunner: &mock.ExecMockRunner{},
	}
	return utils
}

func TestMavenExecute(t *testing.T) {
	t.Run("mavenExecute should write output file", func(t *testing.T) {
		// init
		config := mavenExecuteOptions{
			Goals:                       []string{"goal"},
			LogSuccessfulMavenTransfers: true,
			ReturnStdout:                true,
		}

		mockRunner := mock.ExecMockRunner{}
		mockRunner.StdoutReturn = map[string]string{}
		mockRunner.StdoutReturn[""] = "test output"

		var outputFile string
		var output []byte

		oldWriteFile := writeFile
		writeFile = func(filename string, data []byte, perm os.FileMode) error {
			outputFile = filename
			output = data
			return nil
		}
		defer func() { writeFile = oldWriteFile }()

		// test
		err := runMavenExecute(config, &mockRunner)

		// assert
		expectedParams := []string{
			"--batch-mode", "goal",
		}

		assert.NoError(t, err)
		if assert.Equal(t, 1, len(mockRunner.Calls)) {
			assert.Equal(t, "mvn", mockRunner.Calls[0].Exec)
			assert.Equal(t, expectedParams, mockRunner.Calls[0].Params)
		}
		assert.Equal(t, "test output", string(output))
		assert.Equal(t, ".pipeline/maven_output.txt", outputFile)
	})

	t.Run("mavenExecute should NOT write output file", func(t *testing.T) {
		// init
		config := mavenExecuteOptions{
			Goals:                       []string{"goal"},
			LogSuccessfulMavenTransfers: true,
		}

		mockRunner := mock.ExecMockRunner{}
		mockRunner.StdoutReturn = map[string]string{}
		mockRunner.StdoutReturn[""] = "test output"

		var outputFile string
		var output []byte

		oldWriteFile := writeFile
		writeFile = func(filename string, data []byte, perm os.FileMode) error {
			outputFile = filename
			output = data
			return nil
		}
		defer func() { writeFile = oldWriteFile }()

		// test
		err := runMavenExecute(config, &mockRunner)

		// assert
		expectedParams := []string{
			"--batch-mode", "goal",
		}

		assert.NoError(t, err)
		if assert.Equal(t, 1, len(mockRunner.Calls)) {
			assert.Equal(t, "mvn", mockRunner.Calls[0].Exec)
			assert.Equal(t, expectedParams, mockRunner.Calls[0].Params)
		}
		assert.Equal(t, "", string(output))
		assert.Equal(t, "", outputFile)
	})
}
