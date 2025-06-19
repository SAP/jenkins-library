//go:build unit

package config

import (
	"fmt"
	"net/http"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/systemtrust"
	"github.com/jarcoal/httpmock"

	"github.com/stretchr/testify/assert"
)

const secretName = "sonar"
const secretNameInSystemTrust = "sonarSystemtrustSecretName"
const testServerURL = "https://www.project-piper.io"
const testTokenEndPoint = "tokens"
const testTokenQueryParamName = "systems"
const mockSonarToken = "mockSonarToken"

var testFullURL = fmt.Sprintf("%s/%s?%s=", testServerURL, testTokenEndPoint, testTokenQueryParamName)
var mockSingleTokenResponse = fmt.Sprintf("{\"sonar\": \"%s\"}", mockSonarToken)

func TestSystemTrustConfig(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(http.MethodGet, testFullURL+"sonar", httpmock.NewStringResponder(200, mockSingleTokenResponse))

	stepParams := []StepParameters{createStepParam(secretName, RefTypeSystemTrustSecret, secretNameInSystemTrust, secretName)}

	var systemTrustConfiguration = systemtrust.Configuration{
		Token:               "testToken",
		ServerURL:           testServerURL,
		TokenEndPoint:       testTokenEndPoint,
		TokenQueryParamName: testTokenQueryParamName,
	}
	client := &piperhttp.Client{}
	client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})

	t.Run("Load secret from System Trust - secret not set yet by Vault or config.yml", func(t *testing.T) {
		stepConfig := &StepConfig{Config: map[string]interface{}{
			secretName: "",
		}}

		resolveAllSystemTrustReferences(stepConfig, stepParams, systemTrustConfiguration, client)
		assert.Equal(t, mockSonarToken, stepConfig.Config[secretName])
	})

	t.Run("Load secret from System Trust - secret already by Vault or config.yml", func(t *testing.T) {
		stepConfig := &StepConfig{Config: map[string]interface{}{
			secretName: "aMockTokenFromVault",
		}}

		resolveAllSystemTrustReferences(stepConfig, stepParams, systemTrustConfiguration, client)
		assert.NotEqual(t, mockSonarToken, stepConfig.Config[secretName])
	})
}

func createStepParam(name, refType, systemTrustSecretNameProperty, defaultSecretNameName string) StepParameters {
	return StepParameters{
		Name:    name,
		Aliases: []Alias{},
		ResourceRef: []ResourceReference{
			{
				Type:    refType,
				Name:    systemTrustSecretNameProperty,
				Default: defaultSecretNameName,
			},
		},
	}
}
