package config

import (
	"fmt"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/trustengine"
	"github.com/jarcoal/httpmock"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const secretName = "token"
const secretNameInTrustEngine = "sonar"
const testBaseURL = "https://www.project-piper.io/tokens"
const mockSonarToken = "mockSonarToken"

var mockSingleTokenResponse = fmt.Sprintf("{\"sonar\": \"%s\"}", mockSonarToken)

func TestTrustEngineConfig(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(http.MethodGet, testBaseURL+"?systems=sonar", httpmock.NewStringResponder(200, mockSingleTokenResponse))

	stepParams := []StepParameters{stepParam(secretName, "trustengineSecret", secretNameInTrustEngine, secretName)}

	trustEngineConfiguration := trustengine.Configuration{
		Token:     "mockToken",
		ServerURL: testBaseURL,
	}
	client := &piperhttp.Client{}
	client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})

	t.Run("Load secret from Trust Engine - secret not set yet by Vault or config.yml", func(t *testing.T) {
		stepConfig := &StepConfig{Config: map[string]interface{}{
			secretName: "",
		}}

		ResolveAllTrustEngineReferences(stepConfig, stepParams, trustEngineConfiguration, client)
		assert.Equal(t, mockSonarToken, stepConfig.Config[secretName])
	})

	t.Run("Load secret from Trust Engine - secret already by Vault or config.yml", func(t *testing.T) {
		stepConfig := &StepConfig{Config: map[string]interface{}{
			secretName: "aMockTokenFromVault",
		}}

		ResolveAllTrustEngineReferences(stepConfig, stepParams, trustEngineConfiguration, client)
		assert.NotEqual(t, mockSonarToken, stepConfig.Config[secretName])
	})
}

func stepParam(name, refType, vaultSecretNameProperty, defaultSecretNameName string) StepParameters {
	return StepParameters{
		Name:    name,
		Aliases: []Alias{},
		ResourceRef: []ResourceReference{
			{
				Type:    refType,
				Name:    vaultSecretNameProperty,
				Default: defaultSecretNameName,
			},
		},
	}
}
