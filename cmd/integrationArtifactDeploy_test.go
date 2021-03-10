package cmd

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type integrationArtifactDeployMockUtils struct {
	*mock.ExecMockRunner
}

func newIntegrationArtifactDeployTestsUtils() integrationArtifactDeployMockUtils {
	utils := integrationArtifactDeployMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
	}
	return utils
}

func TestRunIntegrationArtifactDeploy(t *testing.T) {
	t.Parallel()

	t.Run("Successfull Integration Flow Deploy Test", func(t *testing.T) {

		config := integrationArtifactDeployOptions{
			Host:                   "https://demo",
			OAuthTokenProviderURL:  "https://demo/oauth/token",
			Username:               "demouser",
			Password:               "******",
			IntegrationFlowID:      "flow1",
			IntegrationFlowVersion: "1.0.1",
			Platform:               "cf",
		}

		httpClient := httpMockCpis{CPIFunction: "IntegrationArtifactDeploy", ResponseBody: ``, TestType: "Positive"}

		err := runIntegrationArtifactDeploy(&config, nil, &httpClient)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "https://demo/api/v1/DeployIntegrationDesigntimeArtifact?Id='flow1'&Version='1.0.1'", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "POST", httpClient.Method)
			})
		}
	})

	t.Run("Failed case of Integration Flow Deploy Test", func(t *testing.T) {
		config := integrationArtifactDeployOptions{
			Host:                   "https://demo",
			OAuthTokenProviderURL:  "https://demo/oauth/token",
			Username:               "demouser",
			Password:               "******",
			IntegrationFlowID:      "flow1",
			IntegrationFlowVersion: "1.0.1",
			Platform:               "cf",
		}

		httpClient := httpMockCpis{CPIFunction: "IntegrationArtifactDeploy", ResponseBody: ``, TestType: "Negative"}

		err := runIntegrationArtifactDeploy(&config, nil, &httpClient)

		assert.EqualError(t, err, "HTTP POST request to https://demo/api/v1/DeployIntegrationDesigntimeArtifact?Id='flow1'&Version='1.0.1' failed with error: Internal Server Error")
	})

}

type httpMockCpis struct {
	Method       string
	URL          string
	Header       map[string][]string
	ResponseBody string
	Options      piperhttp.ClientOptions
	StatusCode   int
	CPIFunction  string
	TestType     string
}

func (c *httpMockCpis) SetOptions(options piperhttp.ClientOptions) {
	c.Options = options
}

func (c *httpMockCpis) SendRequest(method string, url string, r io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {

	c.Method = method
	c.URL = url

	if r != nil {
		_, err := ioutil.ReadAll(r)

		if err != nil {
			return nil, err
		}
	}

	if c.Options.Token == "" {
		c.ResponseBody = "{\r\n\t\t\t\"access_token\": \"demotoken\",\r\n\t\t\t\"token_type\": \"Bearer\",\r\n\t\t\t\"expires_in\": 3600,\r\n\t\t\t\"scope\": \"\"\r\n\t\t}"
		c.StatusCode = 200
		res := http.Response{
			StatusCode: c.StatusCode,
			Header:     c.Header,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(c.ResponseBody))),
		}
		return &res, nil
	}
	if c.CPIFunction == "" {
		c.CPIFunction = cpi.GetCPIFunctionNameByURLCheck(url, method, c.TestType)
		resp, error := cpi.GetCPIFunctionMockResponse(c.CPIFunction, c.TestType)
		c.CPIFunction = ""
		return resp, error
	}

	return cpi.GetCPIFunctionMockResponse(c.CPIFunction, c.TestType)
}
