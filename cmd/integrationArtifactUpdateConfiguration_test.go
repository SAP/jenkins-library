package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunIntegrationArtifactUpdateConfiguration(t *testing.T) {
	t.Parallel()

	t.Run("Successfully update of Integration Flow configuration parameter test", func(t *testing.T) {
		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`
		config := integrationArtifactUpdateConfigurationOptions{
			APIServiceKey:          apiServiceKey,
			IntegrationFlowID:      "flow1",
			IntegrationFlowVersion: "1.0.1",
			ParameterKey:           "myheader",
			ParameterValue:         "def",
		}

		httpClient := httpMockCpis{CPIFunction: "IntegrationArtifactUpdateConfiguration", ResponseBody: ``, TestType: "Positive", Method: "PUT", URL: "https://demo/api/v1/IntegrationDesigntimeArtifacts(Id='flow1',Version='1.0.1')"}

		err := runIntegrationArtifactUpdateConfiguration(&config, nil, &httpClient)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "https://demo/api/v1/IntegrationDesigntimeArtifacts(Id='flow1',Version='1.0.1')/$links/Configurations('myheader')", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "PUT", httpClient.Method)
			})
		}

	})

	t.Run("Failed case of Integration Flow configuration parameter Test", func(t *testing.T) {
		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`
		config := integrationArtifactUpdateConfigurationOptions{
			APIServiceKey:          apiServiceKey,
			IntegrationFlowID:      "flow1",
			IntegrationFlowVersion: "1.0.1",
			ParameterKey:           "myheader",
			ParameterValue:         "def",
		}

		httpClient := httpMockCpis{CPIFunction: "IntegrationArtifactUpdateConfiguration", ResponseBody: ``, TestType: "Negative"}

		err := runIntegrationArtifactUpdateConfiguration(&config, nil, &httpClient)
		assert.EqualError(t, err, "HTTP \"PUT\" request to \"https://demo/api/v1/IntegrationDesigntimeArtifacts(Id='flow1',Version='1.0.1')/$links/Configurations('myheader')\" failed with error: Not found - either wrong version for the given Id or wrong parameter key")
	})

	t.Run("Failed case of Integration Flow configuration parameter test with error body", func(t *testing.T) {
		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`
		config := integrationArtifactUpdateConfigurationOptions{
			APIServiceKey:          apiServiceKey,
			IntegrationFlowID:      "flow1",
			IntegrationFlowVersion: "1.0.1",
			ParameterKey:           "myheader",
			ParameterValue:         "def",
		}

		httpClient := httpMockCpis{CPIFunction: "IntegrationArtifactUpdateConfiguration", ResponseBody: ``, TestType: "Negative_With_ResponseBody"}

		err := runIntegrationArtifactUpdateConfiguration(&config, nil, &httpClient)
		assert.EqualError(t, err, "Failed to update the integration flow configuration parameter, Response Status code: 400")
	})
}
