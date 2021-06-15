package cmd

import (
	"fmt"
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

		httpClient := cpi.HttpMockCpis{CPIFunction: "", ResponseBody: ``, TestType: "PositiveAndDeployIntegrationDesigntimeArtifactResBody"}

		err := runIntegrationArtifactDeploy(&config, nil, &httpClient)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "https://demo/api/v1/IntegrationRuntimeArtifacts('flow1')", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})
		}
	})

	t.Run("Trigger Failure for Integration Flow Deployment", func(t *testing.T) {

		config := integrationArtifactDeployOptions{
			Host:                   "https://demo",
			OAuthTokenProviderURL:  "https://demo/oauth/token",
			Username:               "demouser",
			Password:               "******",
			IntegrationFlowID:      "flow1",
			IntegrationFlowVersion: "1.0.1",
			Platform:               "cf",
		}

		httpClient := cpi.HttpMockCpis{CPIFunction: "FailIntegrationDesigntimeArtifactDeployment", ResponseBody: ``, TestType: "Negative"}

		err := runIntegrationArtifactDeploy(&config, nil, &httpClient)

		if assert.Error(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "https://demo/api/v1/DeployIntegrationDesigntimeArtifact?Id='flow1'&Version='1.0.1'", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "POST", httpClient.Method)
			})
		}
	})

	t.Run("Failed Integration Flow Deploy Test", func(t *testing.T) {

		config := integrationArtifactDeployOptions{
			Host:                   "https://demo",
			OAuthTokenProviderURL:  "https://demo/oauth/token",
			Username:               "demouser",
			Password:               "******",
			IntegrationFlowID:      "flow1",
			IntegrationFlowVersion: "1.0.1",
			Platform:               "cf",
		}

		httpClient := cpi.HttpMockCpis{CPIFunction: "", ResponseBody: ``, TestType: "NegativeAndDeployIntegrationDesigntimeArtifactResBody"}

		err := runIntegrationArtifactDeploy(&config, nil, &httpClient)

		assert.EqualError(t, err, "{\"message\": \"java.lang.IllegalStateException: No credentials for 'smtp' found\"}")
	})

	t.Run("Successfull GetIntegrationArtifactDeployStatus Test", func(t *testing.T) {
		clientOptions := piperhttp.ClientOptions{}
		clientOptions.Token = fmt.Sprintf("Bearer %s", "Demo")
		config := integrationArtifactDeployOptions{
			Host:                   "https://demo",
			OAuthTokenProviderURL:  "https://demo/oauth/token",
			Username:               "demouser",
			Password:               "******",
			IntegrationFlowID:      "flow1",
			IntegrationFlowVersion: "1.0.1",
			Platform:               "cf",
		}

		httpClient := cpi.HttpMockCpis{CPIFunction: "GetIntegrationArtifactDeployStatus", Options: clientOptions, ResponseBody: ``, TestType: "PositiveAndDeployIntegrationDesigntimeArtifactResBody"}

		resp, err := getIntegrationArtifactDeployStatus(&config, &httpClient)

		assert.Equal(t, "STARTED", resp)

		assert.NoError(t, err)
	})

	t.Run("Successfull GetIntegrationArtifactDeployError Test", func(t *testing.T) {
		clientOptions := piperhttp.ClientOptions{}
		clientOptions.Token = fmt.Sprintf("Bearer %s", "Demo")
		config := integrationArtifactDeployOptions{
			Host:                   "https://demo",
			OAuthTokenProviderURL:  "https://demo/oauth/token",
			Username:               "demouser",
			Password:               "******",
			IntegrationFlowID:      "flow1",
			IntegrationFlowVersion: "1.0.1",
			Platform:               "cf",
		}

		httpClient := cpi.HttpMockCpis{CPIFunction: "GetIntegrationArtifactDeployErrorDetails", Options: clientOptions, ResponseBody: ``, TestType: "PositiveAndGetDeployedIntegrationDesigntimeArtifactErrorResBody"}

		resp, err := getIntegrationArtifactDeployError(&config, &httpClient)

		assert.Equal(t, "{\"message\": \"java.lang.IllegalStateException: No credentials for 'smtp' found\"}", resp)

		assert.NoError(t, err)
	})
}
