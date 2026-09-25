//go:build unit
// +build unit

package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"testing"

	"github.com/SAP/jenkins-library/pkg/config/mocks"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/systemtrust"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestSystemTrustPreferredOverVault(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	const vaultPath = "team1"
	stepParams := []StepParameters{{
		Name: secretName,
		ResourceRef: []ResourceReference{
			{Type: "vaultSecret", Name: "sonarVaultSecretName", Default: secretName},
			{Type: RefTypeSystemTrustSecret, Name: secretNameInSystemTrust, Default: secretName},
		},
	}}
	systemTrustConfig := systemtrust.Configuration{
		Token:               "testToken",
		ServerURL:           testServerURL,
		TokenEndPoint:       testTokenEndPoint,
		TokenQueryParamName: testTokenQueryParamName,
	}

	t.Run("System Trust value is not overwritten by Vault", func(t *testing.T) {
		httpmock.RegisterResponder(http.MethodPost, testFullURL, httpmock.NewStringResponder(http.StatusOK, mockSingleTokenResponse))

		vaultMock := &mocks.VaultClient{}
		stepConfig := &StepConfig{Config: map[string]interface{}{
			"vaultPath": vaultPath,
			secretName:  "",
		}}
		client := &piperhttp.Client{}
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})

		resolveAllSystemTrustReferences(stepConfig, stepParams, systemTrustConfig, client)
		resolveAllVaultReferences(stepConfig, vaultMock, stepParams)

		assert.Equal(t, mockSonarToken, stepConfig.Config[secretName])
		vaultMock.AssertNotCalled(t, "GetKvSecret", path.Join(vaultPath, secretName))
	})

	t.Run("Vault remains the fallback when System Trust fails", func(t *testing.T) {
		httpmock.RegisterResponder(http.MethodPost, testFullURL, httpmock.NewStringResponder(http.StatusForbidden, "forbidden"))

		vaultMock := &mocks.VaultClient{}
		vaultMock.On("GetKvSecret", path.Join(vaultPath, secretName)).Return(map[string]string{secretName: "vaultToken"}, nil)
		stepConfig := &StepConfig{Config: map[string]interface{}{
			"vaultPath": vaultPath,
			secretName:  "",
		}}
		client := &piperhttp.Client{}
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})

		resolveAllSystemTrustReferences(stepConfig, stepParams, systemTrustConfig, client)
		resolveAllVaultReferences(stepConfig, vaultMock, stepParams)

		assert.Equal(t, "vaultToken", stepConfig.Config[secretName])
		vaultMock.AssertExpectations(t)
	})

	t.Run("Vault is used when System Trust is skipped", func(t *testing.T) {
		httpmock.ZeroCallCounters()
		httpmock.RegisterResponder(http.MethodPost, testFullURL, httpmock.NewStringResponder(http.StatusOK, mockSingleTokenResponse))

		vaultMock := &mocks.VaultClient{}
		vaultMock.On("GetKvSecret", path.Join(vaultPath, secretName)).Return(map[string]string{secretName: "vaultToken"}, nil)
		stepConfig := &StepConfig{Config: map[string]interface{}{
			"vaultPath":     vaultPath,
			skipSystemTrust: true,
			secretName:      "presetToken",
		}}
		client := &piperhttp.Client{}
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})

		resolveAllSystemTrustReferences(stepConfig, stepParams, systemTrustConfig, client)
		resolveAllVaultReferences(stepConfig, vaultMock, stepParams)

		assert.Equal(t, "vaultToken", stepConfig.Config[secretName])
		assert.Equal(t, 0, httpmock.GetTotalCallCount())
		vaultMock.AssertExpectations(t)
	})

	t.Run("Preset value remains when System Trust is skipped and Vault overwrite is disabled", func(t *testing.T) {
		httpmock.ZeroCallCounters()

		vaultMock := &mocks.VaultClient{}
		stepConfig := &StepConfig{Config: map[string]interface{}{
			"vaultPath":           vaultPath,
			skipSystemTrust:       true,
			vaultDisableOverwrite: true,
			secretName:            "presetToken",
		}}

		resolveAllSystemTrustReferences(stepConfig, stepParams, systemTrustConfig, &piperhttp.Client{})
		resolveAllVaultReferences(stepConfig, vaultMock, stepParams)

		assert.Equal(t, "presetToken", stepConfig.Config[secretName])
		assert.Equal(t, 0, httpmock.GetTotalCallCount())
		vaultMock.AssertNotCalled(t, "GetKvSecret", path.Join(vaultPath, secretName))
	})
	t.Run("Vault fallback when System Trust returns an empty token", func(t *testing.T) {
		httpmock.RegisterResponder(http.MethodPost, testFullURL, httpmock.NewStringResponder(http.StatusOK, `{"sonar":""}`))

		vaultMock := &mocks.VaultClient{}
		vaultMock.On("GetKvSecret", path.Join(vaultPath, secretName)).Return(map[string]string{secretName: "vaultToken"}, nil)
		stepConfig := &StepConfig{Config: map[string]interface{}{
			"vaultPath": vaultPath,
			secretName:  "",
		}}
		client := &piperhttp.Client{}
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})

		resolveAllSystemTrustReferences(stepConfig, stepParams, systemTrustConfig, client)
		resolveAllVaultReferences(stepConfig, vaultMock, stepParams)

		assert.Equal(t, "vaultToken", stepConfig.Config[secretName])
		vaultMock.AssertExpectations(t)
	})
}

const secretName = "sonar"
const secretNameInSystemTrust = "sonarSystemtrustSecretName"
const testServerURL = "https://www.project-piper.io"
const testTokenEndPoint = "tokens"
const testTokenQueryParamName = "systems" // no longer used by the new implementation, but kept in config
const mockSonarToken = "mockSonarToken"

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
	t.Run("Load secret from System Trust - parameter is absent", func(t *testing.T) {
		stepConfig := &StepConfig{Config: map[string]interface{}{}}

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
