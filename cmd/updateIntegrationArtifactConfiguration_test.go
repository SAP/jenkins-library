package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type updateIntegrationArtifactConfigurationMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newUpdateIntegrationArtifactConfigurationTestsUtils() updateIntegrationArtifactConfigurationMockUtils {
	utils := updateIntegrationArtifactConfigurationMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunUpdateIntegrationArtifactConfiguration(t *testing.T) {
	t.Parallel()

	t.Run("Successfull update of Integration Flow configuration parameter test", func(t *testing.T) {
		config := updateIntegrationArtifactConfigurationOptions{
			Host:                   "https://demo",
			OAuthTokenProviderURL:  "https://demo/oauth/token",
			Username:               "demouser",
			Password:               "******",
			IntegrationFlowID:      "flow1",
			IntegrationFlowVersion: "1.0.1",
			ParameterKey:           "myheader",
			ParameterValue:         "def",
		}

		httpClient := httpMockCpis{CPIFunction: "UpdateIntegrationArtifactConfiguration", ResponseBody: ``, TestType: "Positive"}

		err := runUpdateIntegrationArtifactConfiguration(&config, nil, &httpClient)
		// assert
		assert.NoError(t, err)
	})

	t.Run("Failed case of Integration Flow configuration parameter Test", func(t *testing.T) {
		config := updateIntegrationArtifactConfigurationOptions{
			Host:                   "https://demo",
			OAuthTokenProviderURL:  "https://demo/oauth/token",
			Username:               "demouser",
			Password:               "******",
			IntegrationFlowID:      "flow1",
			IntegrationFlowVersion: "1.0.1",
			ParameterKey:           "myheader",
			ParameterValue:         "def",
		}

		httpClient := httpMockCpis{CPIFunction: "UpdateIntegrationArtifactConfiguration", ResponseBody: ``, TestType: "Negative", Method: "PUT", URL: "https://demo/api/v1/IntegrationDesigntimeArtifacts(Id='flow1',Version='1.0.1')"}

		err := runUpdateIntegrationArtifactConfiguration(&config, nil, &httpClient)
		// assert
		assert.EqualError(t, err, "HTTP PUT request to https://demo/api/v1/IntegrationDesigntimeArtifacts(Id='flow1',Version='1.0.1')/$links/Configurations('myheader') failed with error: Not found - either wrong version for the given Id or wrong parameter key")
	})
}
