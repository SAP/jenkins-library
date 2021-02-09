package cmd

import (
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type integrationArtifactDownloadMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newIntegrationArtifactDownloadTestsUtils() integrationArtifactDownloadMockUtils {
	utils := integrationArtifactDownloadMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunIntegrationArtifactDownload(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		// init
		config := integrationArtifactDownloadOptions{
			Host:                   "https://roverpoc.it-accd002.cfapps.sap.hana.ondemand.com",
			DownloadPath:           "iflows",
			Username:               "sb-8ff0b149-c3e6-417e-ad27-21fa5a3349dd!b15187|it!b11463",
			Password:               "9f4e13b3-312f-4644-9607-7c21974cb0d6$a1sii7gpT3h_242UKSJLbKnV8wKyeQ6qCsQTxEmvDfE=",
			IntegrationFlowID:      "flow1",
			IntegrationFlowVersion: "1.0.24",
			OAuthTokenProviderURL:  "https://roverpoc.authentication.sap.hana.ondemand.com/oauth/token",
		}
		httpClient := &piperhttp.Client{}
		// test
		err := runIntegrationArtifactDownload(&config, nil, httpClient)
		// assert
		assert.NoError(t, err)
	})
}
