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

func TestSystemTrustReferenceArguments(t *testing.T) {
	params := []StepParameters{
		{Name: "assertionToken", ResourceRef: []ResourceReference{{Type: RefTypeSystemTrustSecret, Default: "btp-cf<scope>$(systemTrustScope)", Args: map[string]string{"ias_tenant_issuer_uri": "$(iasTenantIssuerUri)"}}}},
		{Name: "systemTrustScope"},
		{Name: "iasTenantIssuerUri"},
	}
	for _, tc := range []struct {
		name      string
		changes   map[string]any
		status    int
		response  string
		wantToken string
		wantCalls int
		wantScope string
	}{
		{name: "component", wantToken: "cf-assertion", wantCalls: 1, wantScope: "component"},
		{name: "workspace", changes: map[string]any{"systemTrustScope": "workspace"}, wantToken: "cf-assertion", wantCalls: 1, wantScope: "workspace"},
		{name: "preset", changes: map[string]any{"assertionToken": "preset"}, wantToken: "preset"},
		{name: "skipped", changes: map[string]any{skipSystemTrust: true}},
		{name: "missing issuer", changes: map[string]any{"iasTenantIssuerUri": nil}},
		{name: "empty issuer", changes: map[string]any{"iasTenantIssuerUri": ""}},
		{name: "non-string issuer", changes: map[string]any{"iasTenantIssuerUri": 42}},
		{name: "unresolved issuer", changes: map[string]any{"iasTenantIssuerUri": "$(absent)"}},
		{name: "missing scope", changes: map[string]any{"systemTrustScope": nil}},
		{name: "empty scope", changes: map[string]any{"systemTrustScope": ""}},
		{name: "non-string scope", changes: map[string]any{"systemTrustScope": true}},
		{name: "unresolved scope", changes: map[string]any{"systemTrustScope": "$(absent)"}},
		{name: "forbidden", status: 403, response: "forbidden", wantCalls: 1, wantScope: "component"},
		{name: "empty token", response: `{"btp-cf":""}`, wantCalls: 1, wantScope: "component"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			values := map[string]any{"assertionToken": "", "systemTrustScope": "component", "iasTenantIssuerUri": "https://ias.example", "username": "fallback-user", "password": "fallback-password"}
			for key, value := range tc.changes {
				if value == nil {
					delete(values, key)
				} else {
					values[key] = value
				}
			}
			httpmock.RegisterResponder(http.MethodPost, testFullURL, func(req *http.Request) (*http.Response, error) {
				var body []map[string]any
				assert.NoError(t, json.NewDecoder(req.Body).Decode(&body))
				assert.Equal(t, []map[string]any{{"system": "btp-cf", "scope": tc.wantScope, "args": map[string]any{"ias_tenant_issuer_uri": "https://ias.example"}}}, body)
				status, response := tc.status, tc.response
				if status == 0 {
					status = 200
				}
				if response == "" {
					response = `{"btp-cf":"cf-assertion"}`
				}
				resp := httpmock.NewStringResponse(status, response)
				resp.Request = req
				return resp, nil
			})
			client := &piperhttp.Client{}
			client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
			stepConfig := &StepConfig{Config: values}
			resolveAllSystemTrustReferences(stepConfig, params, systemtrust.Configuration{ServerURL: testServerURL, TokenEndPoint: testTokenEndPoint}, client)
			assert.Equal(t, tc.wantCalls, httpmock.GetTotalCallCount())
			assert.Equal(t, tc.wantToken, values["assertionToken"])
			assert.Equal(t, "fallback-user", values["username"])
			assert.Equal(t, "fallback-password", values["password"])
		})
	}
	t.Run("templated empty scope never uses pipeline default", func(t *testing.T) {
		_, _, err := resolveSystemTrustReference(ResourceReference{Default: "btp-cf<scope>$(scope)"}, map[string]any{"scope": ""})
		assert.ErrorContains(t, err, "nonempty system and scope")
	})
	t.Run("legacy literal scope parsing unchanged", func(t *testing.T) {
		ref, args, err := resolveSystemTrustReference(ResourceReference{Default: "sonar<scope>"}, nil)
		assert.NoError(t, err)
		assert.Equal(t, "sonar<scope>", ref)
		assert.Nil(t, args)
	})
}
