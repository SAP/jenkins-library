package cmd

import (
	"fmt"
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
			Username:               "sb-8255b149-c3e6-417e-ad27-21fa5a3349dd!b15187|it!b11463",
			Password:               "911e13b3-312f-4644-9607-7c21974cb0d6$a1sii7gpT3h_242UKSJLbKnV8wKyeQ6qCsQTxEmvDfE=",
			IntegrationFlowID:      "flow1",
			IntegrationFlowVersion: "1.0.1",
			Platform:               "cf",
		}

		httpClient := httpMockGcts{StatusCode: 202,
			Header:       map[string][]string{"Authorization": {"eyJhbGciOiJSUzI1NiIsImprdSI6Imh0dHBzOi8vY3Bpc3VpdGUtZXVyb3BlLTA4LmF1dGhlbnRp"}},
			ResponseBody: ``}

		err := runDeployIntegrationArtifact(&config, nil, &httpClient)
		fmt.Printf("%s.\n", err)

		// assert
		assert.NoError(t, err)
	})

}
