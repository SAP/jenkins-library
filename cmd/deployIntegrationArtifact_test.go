package cmd

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	cpi "github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type deployIntegrationArtifactMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newDeployIntegrationArtifactTestsUtils() deployIntegrationArtifactMockUtils {
	utils := deployIntegrationArtifactMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunDeployIntegrationArtifact(t *testing.T) {
	t.Parallel()

	t.Run("Successfull Integration Flow Deploy Test", func(t *testing.T) {

		config := deployIntegrationArtifactOptions{
			Host:                   "https://demo",
			OAuthTokenProviderURL:  "https://demo/oauth/token",
			Username:               "demouser",
			Password:               "******",
			IntegrationFlowID:      "flow1",
			IntegrationFlowVersion: "1.0.1",
			Platform:               "cf",
		}

		httpClient := httpMockCpis{CPIFunction: "DeployIntegrationDesigntimeArtifact", ResponseBody: ``, TestType: "Positive"}

		err := runDeployIntegrationArtifact(&config, nil, &httpClient)
		// assert
		assert.NoError(t, err)
	})

	t.Run("Failed case of Integration Flow Deploy Test", func(t *testing.T) {
		config := deployIntegrationArtifactOptions{
			Host:                   "https://demo",
			OAuthTokenProviderURL:  "https://demo/oauth/token",
			Username:               "demouser",
			Password:               "******",
			IntegrationFlowID:      "flow1",
			IntegrationFlowVersion: "1.0.1",
			Platform:               "cf",
		}

		httpClient := httpMockCpis{CPIFunction: "DeployIntegrationDesigntimeArtifact", ResponseBody: ``, TestType: "Negative"}

		err := runDeployIntegrationArtifact(&config, nil, &httpClient)
		// assert
		assert.EqualError(t, err, "Integration Flow deployment failed, Response Status code: 500")
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
	return cpi.GetCPIFunctionMockResponse(c.CPIFunction, c.TestType)
}
