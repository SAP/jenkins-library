//go:build unit

package cmd

import (
	"fmt"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/stretchr/testify/assert"
)

func TestRunIntegrationArtifactGetMplStatus(t *testing.T) {
	t.Parallel()

	t.Run("Successfully Test of Get Integration Flow MPL Status", func(t *testing.T) {
		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`
		config := integrationArtifactGetMplStatusOptions{
			APIServiceKey:     apiServiceKey,
			IntegrationFlowID: "flow1",
		}

		httpClient := httpMockCpis{CPIFunction: "IntegrationArtifactGetMplStatus", ResponseBody: ``, TestType: "Positive"}
		seOut := integrationArtifactGetMplStatusCommonPipelineEnvironment{}
		err := runIntegrationArtifactGetMplStatus(&config, nil, &httpClient, &seOut)

		if assert.NoError(t, err) {
			assert.EqualValues(t, seOut.custom.integrationFlowMplStatus, "COMPLETED")

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "https://demo/api/v1/MessageProcessingLogs?$filter=IntegrationArtifact/Id+eq+'flow1'+and+Status+ne+'DISCARDED'&$orderby=LogEnd+desc&$top=1", httpClient.URL)
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
		config := integrationArtifactGetMplStatusOptions{
			APIServiceKey:     apiServiceKey,
			IntegrationFlowID: "flow1",
		}

		httpClient := httpMockCpis{CPIFunction: "IntegrationArtifactGetMplStatus", ResponseBody: ``, TestType: "Negative"}

		seOut := integrationArtifactGetMplStatusCommonPipelineEnvironment{}
		err := runIntegrationArtifactGetMplStatus(&config, nil, &httpClient, &seOut)
		assert.EqualValues(t, seOut.custom.integrationFlowMplStatus, "")
		assert.EqualError(t, err, "HTTP GET request to https://demo/api/v1/MessageProcessingLogs?$filter=IntegrationArtifact/"+
			"Id+eq+'flow1'+and+Status+ne+'DISCARDED'&$orderby=LogEnd+desc&$top=1 failed with error: "+
			"Unable to get integration flow MPL status, Response Status code:400")
	})

	t.Run(" Integration flow message processing get Error message test", func(t *testing.T) {
		clientOptions := piperhttp.ClientOptions{}
		clientOptions.Token = fmt.Sprintf("Bearer %s", "Demo")
		httpClient := httpMockCpis{CPIFunction: "IntegrationArtifactGetMplStatusError", Options: clientOptions, ResponseBody: ``, TestType: "Negative"}
		seOut := integrationArtifactGetMplStatusCommonPipelineEnvironment{}
		message, err := getIntegrationArtifactMPLError(&seOut, "1000111", &httpClient, "demo")
		assert.NoError(t, err)
		assert.NotNil(t, message)
		assert.EqualValues(t, seOut.custom.integrationFlowMplError, "{\"message\": \"java.lang.IllegalStateException: No credentials for 'smtp' found\"}")
	})

}
