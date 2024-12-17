//go:build unit
// +build unit

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

const secretName = "sonar"
const secretNameInTrustEngine = "sonarTrustengineSecretName"
const testServerURL = "https://www.project-piper.io"
const testTokenEndPoint = "tokens"
const testTokenQueryParamName = "systems"
const mockSonarToken = "mockSonarToken"

var testFullURL = fmt.Sprintf("%s/%s?%s=", testServerURL, testTokenEndPoint, testTokenQueryParamName)
var mockSingleTokenResponse = fmt.Sprintf("{\"sonar\": \"%s\"}", mockSonarToken)

func TestTrustEngineConfig(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(http.MethodGet, testFullURL+"sonar", httpmock.NewStringResponder(200, mockSingleTokenResponse))

	stepParams := []StepParameters{createStepParam(secretName, RefTypeTrustengineSecret, secretNameInTrustEngine, secretName)}

	var trustEngineConfiguration = trustengine.Configuration{
		Token:               "testToken",
		ServerURL:           testServerURL,
		TokenEndPoint:       testTokenEndPoint,
		TokenQueryParamName: testTokenQueryParamName,
	}
	client := &piperhttp.Client{}
	client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})

	t.Run("Load secret from Trust Engine - secret not set yet by Vault or config.yml", func(t *testing.T) {
		stepConfig := &StepConfig{Config: map[string]interface{}{
			secretName: "",
		}}

		resolveAllTrustEngineReferences(stepConfig, stepParams, trustEngineConfiguration, client)
		assert.Equal(t, mockSonarToken, stepConfig.Config[secretName])
	})

	t.Run("Load secret from Trust Engine - secret already by Vault or config.yml", func(t *testing.T) {
		stepConfig := &StepConfig{Config: map[string]interface{}{
			secretName: "aMockTokenFromVault",
		}}

		resolveAllTrustEngineReferences(stepConfig, stepParams, trustEngineConfiguration, client)
		assert.NotEqual(t, mockSonarToken, stepConfig.Config[secretName])
	})
}

func createStepParam(name, refType, trustengineSecretNameProperty, defaultSecretNameName string) StepParameters {
	return StepParameters{
		Name:    name,
		Aliases: []Alias{},
		ResourceRef: []ResourceReference{
			{
				Type:    refType,
				Name:    trustengineSecretNameProperty,
				Default: defaultSecretNameName,
			},
		},
	}
}
