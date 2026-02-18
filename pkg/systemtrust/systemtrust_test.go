//go:build unit

package systemtrust

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/stretchr/testify/assert"
)

const testServerURL = "https://www.project-piper.io"
const testTokenEndPoint = "tokens"
const testTokenQueryParamName = "systems"
const mockSonarToken = "mockSonarToken"
const mockblackduckToken = "mockblackduckToken"
const errorMsg403 = "unauthorized to request token"

var testFullURL = fmt.Sprintf("%s/%s?%s=", testServerURL, testTokenEndPoint, testTokenQueryParamName)
var mockSingleTokenResponse = fmt.Sprintf("{\"sonar\": \"%s\"}", mockSonarToken)
var mockTwoTokensResponse = fmt.Sprintf("{\"sonar\": \"%s\", \"blackduck\": \"%s\"}", mockSonarToken, mockblackduckToken)
var systemTrustConfiguration = Configuration{
	Token:               "testToken",
	ServerURL:           testServerURL,
	TokenEndPoint:       testTokenEndPoint,
	TokenQueryParamName: testTokenQueryParamName,
}

func TestSystemTrust(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	t.Run("Get Sonar token - happy path", func(t *testing.T) {
		httpmock.RegisterResponder(http.MethodGet, testFullURL+"sonar", httpmock.NewStringResponder(200, mockSingleTokenResponse))

		client := &piperhttp.Client{}
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})

		token, err := GetToken("sonar", client, systemTrustConfiguration)
		assert.NoError(t, err)
		assert.Equal(t, mockSonarToken, token)
	})

	t.Run("Get multiple tokens - happy path", func(t *testing.T) {
		httpmock.RegisterResponder(http.MethodGet, testFullURL+"sonar,blackduck", httpmock.NewStringResponder(200, mockTwoTokensResponse))

		client := &piperhttp.Client{}
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})

		secrets, err := getSecrets([]string{"sonar", "blackduck"}, client, systemTrustConfiguration)

		assert.NoError(t, err)
		assert.Len(t, secrets, 2)
		for _, s := range secrets {
			switch system := s.System; system {
			case "sonar":
				assert.Equal(t, mockSonarToken, s.Token)
			case "blackduck":
				assert.Equal(t, mockblackduckToken, s.Token)
			default:
				continue
			}
		}
	})

	t.Run("Get Sonar token - 403 error", func(t *testing.T) {
		httpmock.RegisterResponder(http.MethodGet, testFullURL+"sonar", httpmock.NewStringResponder(403, errorMsg403))

		client := &piperhttp.Client{}
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})

		_, err := GetToken("sonar", client, systemTrustConfiguration)
		assert.Error(t, err)
	})

}
