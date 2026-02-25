//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunIntegrationArtifactUnDeploy(t *testing.T) {
	t.Parallel()

	t.Run("Successful undeploy of integration flow test", func(t *testing.T) {

		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`
		config := integrationArtifactUnDeployOptions{
			APIServiceKey:     apiServiceKey,
			IntegrationFlowID: "flow1",
		}
		httpClient := httpMockCpis{CPIFunction: "PositiveAndUnDeployIntegrationDesigntimeArtifact", ResponseBody: ``, TestType: "Positive"}

		// test
		err := runIntegrationArtifactUnDeploy(&config, nil, &httpClient)

		// assert
		assert.NoError(t, err)
	})

	t.Run("Failed undeploy of integration flow test", func(t *testing.T) {
		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`
		config := integrationArtifactUnDeployOptions{
			APIServiceKey:     apiServiceKey,
			IntegrationFlowID: "flow1",
		}

		httpClient := httpMockCpis{CPIFunction: "FailedIntegrationRuntimeArtifactUnDeployment", ResponseBody: ``, TestType: "Negative"}

		// test
		err := runIntegrationArtifactUnDeploy(&config, nil, &httpClient)

		// assert
		assert.EqualError(t, err, "integration flow undeployment failed, response Status code: 400")
	})
}
