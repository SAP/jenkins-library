package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type integrationArtifactGetServiceEndpointMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newIntegrationArtifactGetServiceEndpointTestsUtils() integrationArtifactGetServiceEndpointMockUtils {
	utils := integrationArtifactGetServiceEndpointMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunIntegrationArtifactGetServiceEndpoint(t *testing.T) {
	t.Parallel()

	t.Run("Successfully Test of Get Integration Flow Service Endpoint", func(t *testing.T) {
		serviceKey := `{
			"url": "https://demo",
			"uaa": {
				"clientid": "demouser",
				"clientsecret": "******",
				"url": "https://demo/oauth/token"
			}
		}`
		config := integrationArtifactGetServiceEndpointOptions{
			ServiceKey:        serviceKey,
			IntegrationFlowID: "CPI_IFlow_Call_using_Cert",
			Platform:          "cf",
		}

		httpClient := httpMockCpis{CPIFunction: "IntegrationArtifactGetServiceEndpoint", ResponseBody: ``, TestType: "PositiveAndGetetIntegrationArtifactGetServiceResBody"}
		seOut := integrationArtifactGetServiceEndpointCommonPipelineEnvironment{}
		err := runIntegrationArtifactGetServiceEndpoint(&config, nil, &httpClient, &seOut)
		assert.EqualValues(t, seOut.custom.iFlowServiceEndpoint, "https://demo.cfapps.sap.hana.ondemand.com/http/testwithcert")

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
		serviceKey := `{
			"url": "https://demo",
			"uaa": {
				"clientid": "demouser",
				"clientsecret": "******",
				"url": "https://demo/oauth/token"
			}
		}`
		config := integrationArtifactGetServiceEndpointOptions{
			ServiceKey:        serviceKey,
			IntegrationFlowID: "CPI_IFlow_Call_using_Cert",
			Platform:          "cf",
		}

		httpClient := httpMockCpis{CPIFunction: "IntegrationArtifactGetServiceEndpoint", ResponseBody: ``, TestType: "Negative"}

		seOut := integrationArtifactGetServiceEndpointCommonPipelineEnvironment{}
		err := runIntegrationArtifactGetServiceEndpoint(&config, nil, &httpClient, &seOut)
		assert.EqualValues(t, seOut.custom.iFlowServiceEndpoint, "")
		assert.EqualError(t, err, "HTTP GET request to https://demo/api/v1/ServiceEndpoints?$expand=EntryPoints failed with error: Unable to get integration flow service endpoint, Response Status code:400")
	})

}
