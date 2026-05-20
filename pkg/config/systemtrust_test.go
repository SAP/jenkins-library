//go:build unit
// +build unit

package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
const testTokenQueryParamName = "systems" // no longer used by the new implementation, but kept in config
const mockSonarToken = "mockSonarToken"
const testStagingServerURL = "https://staging.trust.tools.sap"

var testFullURL = fmt.Sprintf("%s/%s", testServerURL, testTokenEndPoint)
var mockSingleTokenResponse = fmt.Sprintf("{\"sonar\": \"%s\"}", mockSonarToken)

func TestSystemTrustConfig(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(http.MethodPost, testFullURL,
		func(req *http.Request) (*http.Response, error) {
			// verify request body matches new POST contract
			bodyBytes, err := io.ReadAll(req.Body)
			if err != nil {
				return httpmock.NewStringResponse(http.StatusBadRequest, "failed to read body"), nil
			}

			var got []map[string]string
			if err := json.Unmarshal(bodyBytes, &got); err != nil {
				return httpmock.NewStringResponse(http.StatusBadRequest, "invalid json body"), nil
			}

			// Expect: [{"system":"sonar","scope":"pipeline"}]
			if len(got) != 1 || got[0]["system"] != "sonar" || got[0]["scope"] != "pipeline" {
				return httpmock.NewStringResponse(http.StatusBadRequest, "unexpected request body"), nil
			}

			resp := httpmock.NewStringResponse(http.StatusOK, mockSingleTokenResponse)
			resp.Header.Set("Content-Type", "application/json")
			return resp, nil
		},
	)

	stepParams := []StepParameters{createStepParam(secretName, RefTypeSystemTrustSecret, secretNameInSystemTrust, secretName)}

	systemTrustConfiguration := systemtrust.Configuration{
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

// Optional helper if you prefer exact JSON matching instead of map-based checks above.
func mustCompactJSON(t *testing.T, s string) string {
	t.Helper()
	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(s)); err != nil {
		t.Fatalf("failed to compact json: %v", err)
	}
	return buf.String()
}

func TestSetSystemTrustConfiguration(t *testing.T) {
	hookConfig := map[string]interface{}{
		"systemtrust": map[string]interface{}{
			"serverURL":           testServerURL,
			"tokenEndPoint":       testTokenEndPoint,
			"tokenQueryParamName": testTokenQueryParamName,
		},
	}

	t.Run("Uses hook serverURL when no systemTrustURL in stepConfig", func(t *testing.T) {
		c := &Config{}
		c.systemTrustConfiguration.Token = "testToken"
		err := c.setSystemTrustConfiguration(hookConfig, map[string]interface{}{})
		assert.NoError(t, err)
		assert.Equal(t, testServerURL, c.systemTrustConfiguration.ServerURL)
	})

	t.Run("Overrides serverURL with user-provided systemTrustURL from stepConfig", func(t *testing.T) {
		c := &Config{}
		c.systemTrustConfiguration.Token = "testToken"
		err := c.setSystemTrustConfiguration(hookConfig, map[string]interface{}{
			"systemTrustURL": testStagingServerURL,
		})
		assert.NoError(t, err)
		assert.Equal(t, testStagingServerURL, c.systemTrustConfiguration.ServerURL)
	})

	t.Run("Does not override when systemTrustURL is empty string", func(t *testing.T) {
		c := &Config{}
		c.systemTrustConfiguration.Token = "testToken"
		err := c.setSystemTrustConfiguration(hookConfig, map[string]interface{}{
			"systemTrustURL": "",
		})
		assert.NoError(t, err)
		assert.Equal(t, testServerURL, c.systemTrustConfiguration.ServerURL)
	})

	t.Run("Returns error when no token set", func(t *testing.T) {
		c := &Config{}
		err := c.setSystemTrustConfiguration(hookConfig, map[string]interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no System Trust token found")
	})
}
