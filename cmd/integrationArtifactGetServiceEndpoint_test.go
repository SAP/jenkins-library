//go:build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunIntegrationArtifactGetServiceEndpoint(t *testing.T) {
	t.Parallel()

	t.Run("Successfully Test of Get Integration Flow Service Endpoint", func(t *testing.T) {
		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`
		config := integrationArtifactGetServiceEndpointOptions{
			APIServiceKey:     apiServiceKey,
			IntegrationFlowID: "CPI_IFlow_Call_using_Cert",
		}

		httpClient := httpMockCpis{CPIFunction: "IntegrationArtifactGetServiceEndpoint", ResponseBody: ``, TestType: "PositiveAndGetetIntegrationArtifactGetServiceResBody"}
		seOut := integrationArtifactGetServiceEndpointCommonPipelineEnvironment{}
		err := runIntegrationArtifactGetServiceEndpoint(&config, nil, &httpClient, &seOut)
		assert.EqualValues(t, seOut.custom.integrationFlowServiceEndpoint, "https://demo.cfapps.sap.hana.ondemand.com/http/testwithcert")

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "https://demo/api/v1/ServiceEndpoints?$expand=EntryPoints", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})
		}

	})

	t.Run("Failed Test of Get Integration Flow MPL Status", func(t *testing.T) {
		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`
		config := integrationArtifactGetServiceEndpointOptions{
			APIServiceKey:     apiServiceKey,
			IntegrationFlowID: "CPI_IFlow_Call_using_Cert",
		}

		httpClient := httpMockCpis{CPIFunction: "IntegrationArtifactGetServiceEndpoint", ResponseBody: ``, TestType: "Negative"}

		seOut := integrationArtifactGetServiceEndpointCommonPipelineEnvironment{}
		err := runIntegrationArtifactGetServiceEndpoint(&config, nil, &httpClient, &seOut)
		assert.EqualValues(t, seOut.custom.integrationFlowServiceEndpoint, "")
		assert.EqualError(t, err, "HTTP GET request to https://demo/api/v1/ServiceEndpoints?$expand=EntryPoints failed with error: Unable to get integration flow service endpoint, Response Status code:400")
	})

}
