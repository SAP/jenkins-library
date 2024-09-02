package trustengine

import (
	"fmt"
	"github.com/jarcoal/httpmock"
	"net/http"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/stretchr/testify/assert"
)

const testBaseURL = "https://www.project-piper.io/tokens"
const mockSonarToken = "mockSonarToken"
const mockCumulusToken = "mockCumulusToken"
const errorMsg403 = "unauthorized to request token"

var mockSingleTokenResponse = fmt.Sprintf("{\"sonar\": \"%s\"}", mockSonarToken)
var mockTwoTokensResponse = fmt.Sprintf("{\"sonar\": \"%s\", \"cumulus\": \"%s\"}", mockSonarToken, mockCumulusToken)
var trustEngineConfiguration = Configuration{
	Token:     "testToken",
	ServerURL: testBaseURL,
}

func TestTrustEngine(t *testing.T) {

	t.Run("Get Sonar token - happy path", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, testBaseURL+"?systems=sonar", httpmock.NewStringResponder(200, mockSingleTokenResponse))

		client := &piperhttp.Client{}
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})

		token, err := GetToken("sonar", client, trustEngineConfiguration)
		assert.NoError(t, err)
		assert.Equal(t, mockSonarToken, token)
	})

	t.Run("Get multiple tokens - happy path", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, testBaseURL+"?systems=sonar,cumulus", httpmock.NewStringResponder(200, mockTwoTokensResponse))

		client := &piperhttp.Client{}
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})

		secrets, err := GetSecrets([]string{"sonar", "cumulus"}, client, trustEngineConfiguration)

		assert.NoError(t, err)
		assert.Len(t, secrets, 2)
		for _, s := range secrets {
			switch system := s.System; system {
			case "sonar":
				assert.Equal(t, mockSonarToken, s.Token)
			case "cumulus":
				assert.Equal(t, mockCumulusToken, s.Token)
			default:
				continue
			}
		}
	})

	t.Run("Get Sonar token - 403 error", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, testBaseURL+"?systems=sonar", httpmock.NewStringResponder(403, errorMsg403))

		client := &piperhttp.Client{}
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})

		_, err := GetToken("sonar", client, trustEngineConfiguration)
		assert.Error(t, err)
	})

}
