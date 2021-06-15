package cmd

import (
	"github.com/SAP/jenkins-library/pkg/cpi"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type integrationArtifactGetMplStatusMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newIntegrationArtifactGetMplStatusTestsUtils() integrationArtifactGetMplStatusMockUtils {
	utils := integrationArtifactGetMplStatusMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunIntegrationArtifactGetMplStatus(t *testing.T) {
	t.Parallel()

	t.Run("Successfully Test of Get Integration Flow MPL Status", func(t *testing.T) {
		config := integrationArtifactGetMplStatusOptions{
			Host:                  "https://demo",
			OAuthTokenProviderURL: "https://demo/oauth/token",
			Username:              "demouser",
			Password:              "******",
			IntegrationFlowID:     "flow1",
			Platform:              "cf",
		}

		httpClient := cpi.HttpMockCpis{CPIFunction: "IntegrationArtifactGetMplStatus", ResponseBody: ``, TestType: "Positive"}
		seOut := integrationArtifactGetMplStatusCommonPipelineEnvironment{}
		err := runIntegrationArtifactGetMplStatus(&config, nil, &httpClient, &seOut)
		assert.EqualValues(t, seOut.custom.iFlowMplStatus, "COMPLETED")

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "https://demo/api/v1/MessageProcessingLogs?$filter=IntegrationArtifact/Id+eq+'flow1'&$orderby=LogEnd+desc&$top=1", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})
		}

	})

	t.Run("Failed Test of Get Integration Flow MPL Status", func(t *testing.T) {
		config := integrationArtifactGetMplStatusOptions{
			Host:                  "https://demo",
			OAuthTokenProviderURL: "https://demo/oauth/token",
			Username:              "demouser",
			Password:              "******",
			IntegrationFlowID:     "flow1",
			Platform:              "cf",
		}

		httpClient := cpi.HttpMockCpis{CPIFunction: "IntegrationArtifactGetMplStatus", ResponseBody: ``, TestType: "Negative"}

		seOut := integrationArtifactGetMplStatusCommonPipelineEnvironment{}
		err := runIntegrationArtifactGetMplStatus(&config, nil, &httpClient, &seOut)
		assert.EqualValues(t, seOut.custom.iFlowMplStatus, "")
		assert.EqualError(t, err, "HTTP GET request to https://demo/api/v1/MessageProcessingLogs?$filter=IntegrationArtifact/"+
			"Id+eq+'flow1'&$orderby=LogEnd+desc&$top=1 failed with error: "+
			"Unable to get integration flow MPL status, Response Status code:400")
	})
}
