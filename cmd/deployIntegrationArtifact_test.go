package cmd

import (
	"testing"

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
		// init
		config := deployIntegrationArtifactOptions{
			Host:                   "https://demo",
			OAuthTokenProviderURL:  "https://demo/oauth/token",
			Username:               "sb-8f9-c3e6-417e-ad27-21fa5a3349dd!b15187|it!b11463",
			Password:               "9f43-312f-4644-9607-7c21974cb01sii7gpT3h_242UKSJLbKnV8wKyeQ6qCsQTxEmvDfE=",
			IntegrationFlowID:      "flow1",
			IntegrationFlowVersion: "1.0.1",
			Platform:               "cf",
		}

		httpClient := httpMockGcts{StatusCode: 202,
			Header:       map[string][]string{"Authorization": {"eyJhbGciOiJSUzI1NiIsImprdSI6Imh0dHBzOi8vY3Bpc3VpdGUtZXVyb3BlLTA4LmF1dGhlbnRp"}},
			ResponseBody: ``}

		err := runDeployIntegrationArtifact(&config, nil, &httpClient)
		// assert
		assert.NoError(t, err)
	})

	t.Run("Failed case of Integration Flow Deploy Test", func(t *testing.T) {
		// init
		config := deployIntegrationArtifactOptions{
			Host:                   "https://demo",
			OAuthTokenProviderURL:  "https://demo/oauth/token",
			Username:               "sb-8f9-c3e6-417e-ad27-21fa5a3349dd!b15187|it!b11463",
			Password:               "9f43-312f-4644-9607-7c21974cb01sii7gpT3h_242UKSJLbKnV8wKyeQ6qCsQTxEmvDfE=",
			IntegrationFlowID:      "flow1",
			IntegrationFlowVersion: "1.0.1",
			Platform:               "cf",
		}

		httpClient := httpMockGcts{StatusCode: 500,
			Header: map[string][]string{"Authorization": {"eyJhbGciOiJSUzI1NiIsImprdSI6Imh0dHBzOi8vY3Bpc3VpdGUtZXVyb3BlLTA4LmF1dGhlbnRp"}},
			ResponseBody: `{
				"code": "Internal Server Error",
				"message": {
				   "@lang": "en",
				   "#text": "Cannot deploy artifact with Id 'flow1'!"
				}
			 }`}

		err := runDeployIntegrationArtifact(&config, nil, &httpClient)
		// assert
		assert.Error(t, err)
	})

}
