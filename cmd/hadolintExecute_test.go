package cmd

import (
	"net/http"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	piperMocks "github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type hadolintExecuteScanMockUtils struct {
	*piperMocks.ExecMockRunner
	// *mock.FilesMock
	*hadolintMockClient
}

func newHadolintExecuteScanTestsUtils() hadolintExecuteScanMockUtils {
	utils := hadolintExecuteScanMockUtils{
		ExecMockRunner:     &piperMocks.ExecMockRunner{},
		hadolintMockClient: &hadolintMockClient{},
	}
	return utils
}

type hadolintMockClient struct {
	requestedURL  []string
	requestedFile []string
}

func (c *hadolintMockClient) SetOptions(opts piperhttp.ClientOptions) {}

func (c *hadolintMockClient) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	c.requestedURL = append(c.requestedURL, url)
	c.requestedFile = append(c.requestedFile, filename)
	return nil
}

func TestRunHadolintExecute(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		// init
		config := hadolintExecuteOptions{
			DockerFile: "./Dockerfile",
		}
		utils := newHadolintExecuteScanTestsUtils()
		// test
		err := runHadolint(config, &utils, &utils)
		// assert
		assert.NoError(t, err)
		assert.NotEmpty(t, utils.Calls)
		assert.Equal(t, 1, len(utils.Calls))
		assert.Equal(t, "hadolint", utils.Calls[0].Exec)
		assert.Contains(t, utils.Calls[0].Params, "--format")
		assert.Contains(t, utils.Calls[0].Params, "checkstyle")
		assert.NotContains(t, utils.Calls[0].Params, "--config")
		assert.Empty(t, utils.requestedURL)
		assert.Empty(t, utils.requestedFile)
	})

	t.Run("with remote config", func(t *testing.T) {
		// init
		config := hadolintExecuteOptions{
			DockerFile:        "./Dockerfile",
			ConfigurationFile: ".hadolint.yaml",
			ConfigurationURL:  "https://myconfig",
		}
		utils := newHadolintExecuteScanTestsUtils()
		// test
		err := runHadolint(config, &utils, &utils)
		// assert
		assert.NoError(t, err)
		assert.NotEmpty(t, utils.Calls)
		assert.Equal(t, 1, len(utils.Calls))
		assert.Equal(t, "hadolint", utils.Calls[0].Exec)
		assert.Contains(t, utils.Calls[0].Params, "--format")
		assert.Contains(t, utils.Calls[0].Params, "checkstyle")
		// hasFile not mocked yet
		// assert.Contains(t, utils.Calls[0].Params, "--config")
		assert.Contains(t, utils.requestedURL, "https://myconfig")
		assert.Contains(t, utils.requestedFile, ".hadolint.yaml")
	})
}
