//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/cmd/mocks"
	piperMocks "github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRunHadolintExecute(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		// init
		fileMock := &mocks.HadolintPiperFileUtils{}
		clientMock := &mocks.HadolintClient{}
		runnerMock := &piperMocks.ExecMockRunner{}
		config := hadolintExecuteOptions{
			DockerFile:        "./Dockerfile",   // default
			ConfigurationFile: ".hadolint.yaml", // default
		}

		fileMock.
			On("FileExists", config.ConfigurationFile).Return(false, nil).
			On("Glob", config.DockerFile).Return([]string{config.DockerFile}, nil).
			On("WriteFile", "hadolintExecute_reports.json", mock.Anything, mock.Anything).Return(nil).
			On("WriteFile", "hadolintExecute_links.json", mock.Anything, mock.Anything).Return(nil)

		// test
		err := runHadolint(config, hadolintUtils{
			HadolintPiperFileUtils: fileMock,
			HadolintClient:         clientMock,
			hadolintRunner:         runnerMock,
		})
		// assert
		assert.NoError(t, err)
		if assert.Len(t, runnerMock.Calls, 1) {
			assert.Equal(t, "hadolint", runnerMock.Calls[0].Exec)
			assert.Contains(t, runnerMock.Calls[0].Params, config.DockerFile)
			assert.Contains(t, runnerMock.Calls[0].Params, "--format")
			assert.Contains(t, runnerMock.Calls[0].Params, "checkstyle")
			assert.NotContains(t, runnerMock.Calls[0].Params, "--config")
			assert.NotContains(t, runnerMock.Calls[0].Params, config.ConfigurationFile)
		}
		// assert that mocks are called as previously defined
		fileMock.AssertExpectations(t)
		clientMock.AssertExpectations(t)
	})

	t.Run("with remote config", func(t *testing.T) {
		// init
		fileMock := &mocks.HadolintPiperFileUtils{}
		clientMock := &mocks.HadolintClient{}
		runnerMock := &piperMocks.ExecMockRunner{}
		config := hadolintExecuteOptions{
			DockerFile:        "./Dockerfile",   // default
			ConfigurationFile: ".hadolint.yaml", // default
			ConfigurationURL:  "https://myconfig",
		}

		clientMock.
			On("SetOptions", mock.Anything).
			On("DownloadFile", config.ConfigurationURL, config.ConfigurationFile, mock.Anything, mock.Anything).Return(nil)
		fileMock.
			// checks if config exists before downloading
			On("FileExists", config.ConfigurationFile).Return(false, nil).Once().
			// checks again but config is now downloaded
			On("FileExists", config.ConfigurationFile).Return(true, nil).
			On("Glob", config.DockerFile).Return([]string{config.DockerFile}, nil).
			On("WriteFile", "hadolintExecute_reports.json", mock.Anything, mock.Anything).Return(nil).
			On("WriteFile", "hadolintExecute_links.json", mock.Anything, mock.Anything).Return(nil)

		// test
		err := runHadolint(config, hadolintUtils{
			HadolintPiperFileUtils: fileMock,
			HadolintClient:         clientMock,
			hadolintRunner:         runnerMock,
		})
		// assert
		assert.NoError(t, err)
		if assert.Len(t, runnerMock.Calls, 1) {
			assert.Equal(t, "hadolint", runnerMock.Calls[0].Exec)
			assert.Contains(t, runnerMock.Calls[0].Params, config.DockerFile)
			assert.Contains(t, runnerMock.Calls[0].Params, "--format")
			assert.Contains(t, runnerMock.Calls[0].Params, "checkstyle")
			assert.Contains(t, runnerMock.Calls[0].Params, "--config")
			assert.Contains(t, runnerMock.Calls[0].Params, config.ConfigurationFile)
		}
		// assert that mocks are called as previously defined
		fileMock.AssertExpectations(t)
		clientMock.AssertExpectations(t)
	})

	t.Run("with glob pattern", func(t *testing.T) {
		// init
		fileMock := &mocks.HadolintPiperFileUtils{}
		clientMock := &mocks.HadolintClient{}
		runnerMock := &piperMocks.ExecMockRunner{}
		config := hadolintExecuteOptions{
			DockerFile:        "images/**/Dockerfile",
			ConfigurationFile: ".hadolint.yaml",
		}
		matchedFiles := []string{
			"images/a/Dockerfile",
			"images/b/Dockerfile",
		}

		fileMock.
			On("FileExists", config.ConfigurationFile).Return(false, nil).
			On("Glob", config.DockerFile).Return(matchedFiles, nil).
			On("WriteFile", "hadolintExecute_reports.json", mock.Anything, mock.Anything).Return(nil).
			On("WriteFile", "hadolintExecute_links.json", mock.Anything, mock.Anything).Return(nil)

		// test
		err := runHadolint(config, hadolintUtils{
			HadolintPiperFileUtils: fileMock,
			HadolintClient:         clientMock,
			hadolintRunner:         runnerMock,
		})
		// assert
		assert.NoError(t, err)
		if assert.Len(t, runnerMock.Calls, 1) {
			assert.Equal(t, "hadolint", runnerMock.Calls[0].Exec)
			assert.Contains(t, runnerMock.Calls[0].Params, "images/a/Dockerfile")
			assert.Contains(t, runnerMock.Calls[0].Params, "images/b/Dockerfile")
			assert.Contains(t, runnerMock.Calls[0].Params, "--format")
			assert.Contains(t, runnerMock.Calls[0].Params, "checkstyle")
		}
		fileMock.AssertExpectations(t)
		clientMock.AssertExpectations(t)
	})

	t.Run("no matching files", func(t *testing.T) {
		// init
		fileMock := &mocks.HadolintPiperFileUtils{}
		clientMock := &mocks.HadolintClient{}
		runnerMock := &piperMocks.ExecMockRunner{}
		config := hadolintExecuteOptions{
			DockerFile:        "images/**/Dockerfile",
			ConfigurationFile: ".hadolint.yaml",
		}

		fileMock.
			On("FileExists", config.ConfigurationFile).Return(false, nil).
			On("Glob", config.DockerFile).Return([]string{}, nil)

		// test
		err := runHadolint(config, hadolintUtils{
			HadolintPiperFileUtils: fileMock,
			HadolintClient:         clientMock,
			hadolintRunner:         runnerMock,
		})
		// assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no Dockerfiles found")
		assert.Empty(t, runnerMock.Calls)
		fileMock.AssertExpectations(t)
		clientMock.AssertExpectations(t)
	})
}
