//go:build unit
// +build unit

package cmd

import (
	"errors"
	"net/http"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type mavenMockUtils struct {
	shouldFail bool
	*mock.FilesMock
	*mock.ExecMockRunner
}

func (m *mavenMockUtils) DownloadFile(_, _ string, _ http.Header, _ []*http.Cookie) error {
	return errors.New("Test should not download files.")
}

func newMavenMockUtils() mavenMockUtils {
	utils := mavenMockUtils{
		shouldFail:     false,
		FilesMock:      &mock.FilesMock{},
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

		mockUtils := newMavenMockUtils()
		mockUtils.StdoutReturn = map[string]string{}
		mockUtils.StdoutReturn[""] = "test output"

		// test
		err := runMavenExecute(config, &mockUtils)

		// assert
		expectedParams := []string{
			"--batch-mode", "goal",
		}

		assert.NoError(t, err)
		if assert.Equal(t, 1, len(mockUtils.Calls)) {
			assert.Equal(t, "mvn", mockUtils.Calls[0].Exec)
			assert.Equal(t, expectedParams, mockUtils.Calls[0].Params)
		}

		outputFileExists, _ := mockUtils.FileExists(".pipeline/maven_output.txt")
		assert.True(t, outputFileExists)

		output, _ := mockUtils.FileRead(".pipeline/maven_output.txt")

		assert.Equal(t, "test output", string(output))
	})

	t.Run("mavenExecute should NOT write output file", func(t *testing.T) {
		// init
		config := mavenExecuteOptions{
			Goals:                       []string{"goal"},
			LogSuccessfulMavenTransfers: true,
		}

		mockUtils := newMavenMockUtils()

		// test
		err := runMavenExecute(config, &mockUtils)

		// assert
		expectedParams := []string{
			"--batch-mode", "goal",
		}

		assert.NoError(t, err)
		if assert.Equal(t, 1, len(mockUtils.Calls)) {
			assert.Equal(t, "mvn", mockUtils.Calls[0].Exec)
			assert.Equal(t, expectedParams, mockUtils.Calls[0].Params)
		}

		outputFileExists, _ := mockUtils.FileExists(".pipeline/maven_output.txt")
		assert.False(t, outputFileExists)
	})
}
