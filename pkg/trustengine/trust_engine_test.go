package vault

import (
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/jarcoal/httpmock"
	"net/http"
	"net/url"
)

const testBaseURL = "https://www.project-piper.io/"
const testFullURL = "https://www.project-piper.io/test"

func TestTrustEngine(t *testing.T) {

	t.Run("Test getting Sonar token", func(t *testing.T) {
		t.Parallel()

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, testFullURL, httpmock.NewStringResponder(200, `OK`))

		client := &piperhttp.Client{}
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})

		url, _ := url.Parse(testBaseURL)
		GetTrustEngineSecret(url, "test", "123", client)
	})

}
