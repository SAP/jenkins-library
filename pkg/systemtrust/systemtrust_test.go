//go:build unit
// +build unit

package systemtrust

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

const testServerURL = "https://www.project-piper.io"
const testTokenEndPoint = "tokens"
const mockSonarToken = "mockSonarToken"
const mockblackduckToken = "mockblackduckToken"
const errorMsg403 = "unauthorized to request token"

var testFullURL = fmt.Sprintf("%s/%s", testServerURL, testTokenEndPoint)

var mockSingleTokenResponse = fmt.Sprintf("{\"sonar\": \"%s\"}", mockSonarToken)
var mockTwoTokensResponse = fmt.Sprintf("{\"sonar\": \"%s\", \"blackduck\": \"%s\"}", mockSonarToken, mockblackduckToken)

var systemTrustConfiguration = Configuration{
	Token:               "testToken",
	ServerURL:           testServerURL,
	TokenEndPoint:       testTokenEndPoint,
	TokenQueryParamName: "systems", // no longer used by implementation, but kept for compatibility
}

func TestSystemTrust(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	t.Run("Get Sonar token - happy path (POST + JSON body)", func(t *testing.T) {
		httpmock.RegisterResponder(http.MethodPost, testFullURL, func(req *http.Request) (*http.Response, error) {
			defer req.Body.Close()

			bodyBytes, err := io.ReadAll(req.Body)
			assert.NoError(t, err)

			var got []tokenRequest
			err = json.Unmarshal(bodyBytes, &got)
			assert.NoError(t, err)

			// Expect exactly one request: system=sonar, scope=defaultScope
			if assert.Len(t, got, 1) {
				assert.Equal(t, "sonar", got[0].System)
				assert.Equal(t, defaultScope, got[0].Scope)
			}

			return httpmock.NewStringResponse(200, mockSingleTokenResponse), nil
		})

		client := &piperhttp.Client{}
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})

		token, err := GetToken("sonar", client, systemTrustConfiguration)
		assert.NoError(t, err)
		assert.Equal(t, mockSonarToken, token)
	})

	t.Run("Get multiple tokens - happy path (POST + JSON array body)", func(t *testing.T) {
		httpmock.RegisterResponder(http.MethodPost, testFullURL, func(req *http.Request) (*http.Response, error) {
			defer req.Body.Close()

			bodyBytes, err := io.ReadAll(req.Body)
			assert.NoError(t, err)

			var got []tokenRequest
			err = json.Unmarshal(bodyBytes, &got)
			assert.NoError(t, err)

			// Expect two requests in any order
			assert.Len(t, got, 2)

			seen := map[string]string{}
			for _, r := range got {
				seen[r.System] = r.Scope
			}
			assert.Equal(t, defaultScope, seen["sonar"])
			assert.Equal(t, defaultScope, seen["blackduck"])

			return httpmock.NewStringResponse(200, mockTwoTokensResponse), nil
		})

		client := &piperhttp.Client{}
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})

		secrets, err := getSecrets(client, systemTrustConfiguration,
			refNameToTokenBody("sonar"),
			refNameToTokenBody("blackduck"),
		)

		assert.NoError(t, err)
		assert.Len(t, secrets, 2)

		for _, s := range secrets {
			switch s.System {
			case "sonar":
				assert.Equal(t, mockSonarToken, s.Token)
			case "blackduck":
				assert.Equal(t, mockblackduckToken, s.Token)
			}
		}
	})

	t.Run("refNameToTokenBody parses <scope> marker", func(t *testing.T) {
		req := refNameToTokenBody("github-app<scope>pipeline-ghas")
		assert.Equal(t, "github-app", req.System)
		assert.Equal(t, "pipeline-ghas", req.Scope)

		req2 := refNameToTokenBody("sonar")
		assert.Equal(t, "sonar", req2.System)
		assert.Equal(t, defaultScope, req2.Scope)
	})

	t.Run("Get Sonar token - 403 error (POST)", func(t *testing.T) {
		httpmock.RegisterResponder(http.MethodPost, testFullURL, httpmock.NewStringResponder(403, errorMsg403))

		client := &piperhttp.Client{}
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})

		_, err := GetToken("sonar", client, systemTrustConfiguration)
		assert.Error(t, err)
	})
}
