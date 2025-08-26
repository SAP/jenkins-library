package cmd

import (
	"bytes"
	"fmt"
	"io"
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
		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`

		config := integrationArtifactDeployOptions{
			APIServiceKey:     apiServiceKey,
			IntegrationFlowID: "flow1",
		}

		httpClient := httpMockCpis{CPIFunction: "", ResponseBody: ``, TestType: "PositiveAndDeployIntegrationDesigntimeArtifactResBody"}

		err := runIntegrationArtifactDeploy(&config, nil, &httpClient)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "https://demo/api/v1/BuildAndDeployStatus(TaskId='')", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})
		}
	})

	t.Run("Trigger Failure for Integration Flow Deployment", func(t *testing.T) {

		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`

		config := integrationArtifactDeployOptions{
			APIServiceKey:     apiServiceKey,
			IntegrationFlowID: "flow1",
		}

		httpClient := httpMockCpis{CPIFunction: "FailIntegrationDesigntimeArtifactDeployment", ResponseBody: ``, TestType: "Negative"}

		err := runIntegrationArtifactDeploy(&config, nil, &httpClient)

		if assert.Error(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "https://demo/api/v1/DeployIntegrationDesigntimeArtifact?Id='flow1'&Version='Active'", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "POST", httpClient.Method)
			})
		}
	})

	t.Run("Failed Integration Flow Deploy Test", func(t *testing.T) {

		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`

		config := integrationArtifactDeployOptions{
			APIServiceKey:     apiServiceKey,
			IntegrationFlowID: "flow1",
		}

		httpClient := httpMockCpis{CPIFunction: "", ResponseBody: ``, TestType: "NegativeAndDeployIntegrationDesigntimeArtifactResBody"}

		err := runIntegrationArtifactDeploy(&config, nil, &httpClient)

		assert.EqualError(t, err, "{\"message\": \"java.lang.IllegalStateException: No credentials for 'smtp' found\"}")
	})

	t.Run("Successfull GetIntegrationArtifactDeployStatus Test", func(t *testing.T) {
		clientOptions := piperhttp.ClientOptions{}
		clientOptions.Token = fmt.Sprintf("Bearer %s", "Demo")
		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`

		config := integrationArtifactDeployOptions{
			APIServiceKey:     apiServiceKey,
			IntegrationFlowID: "flow1",
		}

		httpClient := httpMockCpis{CPIFunction: "GetIntegrationArtifactDeployStatus", Options: clientOptions, ResponseBody: ``, TestType: "PositiveAndDeployIntegrationDesigntimeArtifactResBody"}

		resp, err := getIntegrationArtifactDeployStatus(&config, &httpClient, "https://demo", "9094d6cd-3683-4a99-794f-834ed30fcb01")

		assert.Equal(t, "STARTED", resp)

		assert.NoError(t, err)
	})

	t.Run("Successfull GetIntegrationArtifactDeployError Test", func(t *testing.T) {
		clientOptions := piperhttp.ClientOptions{}
		clientOptions.Token = fmt.Sprintf("Bearer %s", "Demo")
		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`

		config := integrationArtifactDeployOptions{
			APIServiceKey:     apiServiceKey,
			IntegrationFlowID: "flow1",
		}

		httpClient := httpMockCpis{CPIFunction: "GetIntegrationArtifactDeployErrorDetails", Options: clientOptions, ResponseBody: ``, TestType: "PositiveAndGetDeployedIntegrationDesigntimeArtifactErrorResBody"}

		resp, err := getIntegrationArtifactDeployError(&config, &httpClient, "https://demo")

		assert.Equal(t, "{\"message\": \"java.lang.IllegalStateException: No credentials for 'smtp' found\"}", resp)

		assert.NoError(t, err)
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
		_, err := io.ReadAll(r)

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
			Body:       io.NopCloser(bytes.NewReader([]byte(c.ResponseBody))),
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
