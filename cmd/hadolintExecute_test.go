package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/hadolint/mocks"
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
			On("FileExists", config.ConfigurationFile).Return(false, nil)
		clientMock.
			On("SetOptions", mock.Anything)

		// test
		err := runHadolint(config, hadolintUtils{
			HadolintPiperFileUtils: fileMock,
			HadolintClient:         clientMock,
			hadolintRunner:         runnerMock,
		})
		// assert
		assert.NoError(t, err)
		assert.NotEmpty(t, runnerMock.Calls)
		assert.Equal(t, 1, len(runnerMock.Calls))
		assert.Equal(t, "hadolint", runnerMock.Calls[0].Exec)
		assert.Contains(t, runnerMock.Calls[0].Params, "--format")
		assert.Contains(t, runnerMock.Calls[0].Params, "checkstyle")
		assert.NotContains(t, runnerMock.Calls[0].Params, "--config")
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
			On("FileExists", config.ConfigurationFile).Return(true, nil)
		// test
		err := runHadolint(config, hadolintUtils{
			HadolintPiperFileUtils: fileMock,
			HadolintClient:         clientMock,
			hadolintRunner:         runnerMock,
		})
		// assert
		assert.NoError(t, err)
		assert.NotEmpty(t, runnerMock.Calls)
		assert.Equal(t, 1, len(runnerMock.Calls))
		assert.Equal(t, "hadolint", runnerMock.Calls[0].Exec)
		assert.Contains(t, runnerMock.Calls[0].Params, "--format")
		assert.Contains(t, runnerMock.Calls[0].Params, "checkstyle")
		assert.Contains(t, runnerMock.Calls[0].Params, "--config")
		fileMock.AssertExpectations(t)
		clientMock.AssertExpectations(t)
	})
}
