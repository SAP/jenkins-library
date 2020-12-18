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
			Username:               "demouser",
			Password:               "******",
			IntegrationFlowID:      "flow1",
			IntegrationFlowVersion: "1.0.1",
			Platform:               "cf",
		}

		httpClient := httpMockGcts{StatusCode: 202,
			Header:       map[string][]string{"Authorization": {"dummyBearerToken"}},
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
			Username:               "demouser",
			Password:               "******",
			IntegrationFlowID:      "flow1",
			IntegrationFlowVersion: "1.0.1",
			Platform:               "cf",
		}

		httpClient := httpMockGcts{StatusCode: 500,
			Header: map[string][]string{"Authorization": {"dummyBearerToken"}},
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
