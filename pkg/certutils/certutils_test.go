//go:build unit
// +build unit

package certutils

import (
	"fmt"
	"net/http"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

const (
	caCertsFile = "/kaniko/ssl/certs/ca-certificates.crt"
)

func TestCertificateUpdate(t *testing.T) {
	certLinks := []string{"https://test-link-1.com/cert.crt", "https://test-link-2.com/cert.crt"}
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(http.MethodGet, "https://test-link-1.com/cert.crt", httpmock.NewStringResponder(200, "testCert"))
	httpmock.RegisterResponder(http.MethodGet, "https://test-link-2.com/cert.crt", httpmock.NewStringResponder(200, "testCert"))
	client := &piperhttp.Client{}
	client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})

	t.Run("success case", func(t *testing.T) {
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile(caCertsFile, []byte("initial cert\n"))

		err := CertificateUpdate(certLinks, client, fileUtils, caCertsFile)

		assert.NoError(t, err)
		result, err := fileUtils.FileRead(caCertsFile)
		assert.NoError(t, err)
		assert.Equal(t, "initial cert\ntestCert\ntestCert\n", string(result))
	})

	t.Run("error case - read certs", func(t *testing.T) {
		client := &piperhttp.Client{}
		fileUtils := &mock.FilesMock{}

		err := CertificateUpdate(certLinks, client, fileUtils, caCertsFile)
		assert.EqualError(t, err, "failed to load file '/kaniko/ssl/certs/ca-certificates.crt': could not read '/kaniko/ssl/certs/ca-certificates.crt'")
	})

	t.Run("error case - write certs", func(t *testing.T) {
		fileUtils := &mock.FilesMock{
			FileWriteErrors: map[string]error{
				caCertsFile: fmt.Errorf("write error"),
			},
		}
		fileUtils.AddFile(caCertsFile, []byte("initial cert\n"))

		err := CertificateUpdate(certLinks, client, fileUtils, caCertsFile)
		assert.EqualError(t, err, "failed to update file '/kaniko/ssl/certs/ca-certificates.crt': write error")
	})

	t.Run("error case - get cert via http", func(t *testing.T) {
		httpmock.RegisterResponder(http.MethodGet, "http://non-existing-url", httpmock.NewStringResponder(404, "not found"))

		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile(caCertsFile, []byte("initial cert\n"))

		err := CertificateUpdate([]string{"http://non-existing-url"}, client, fileUtils, caCertsFile)
		assert.Contains(t, err.Error(), "failed to load certificate from url: request to http://non-existing-url returned with response 404")
	})

}

func TestDownloadCACertbunde(t *testing.T) {
	certLinks := []string{"https://test-link-1.com/cert-1.crt", "https://test-link-2.com/cert-2.crt"}
	badCaseLink := "http://non-existing-url"

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(http.MethodGet, certLinks[0], httpmock.NewStringResponder(http.StatusOK, "testCert1"))
	httpmock.RegisterResponder(http.MethodGet, certLinks[1], httpmock.NewStringResponder(http.StatusOK, "testCert2"))
	httpmock.RegisterResponder(http.MethodGet, badCaseLink, httpmock.NewStringResponder(http.StatusNotFound, "not found"))

	client := &piperhttp.Client{}
	client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})

	testTable := []struct {
		name        string
		certsLinks  []string
		expected    string
		expectedErr string
	}{
		{
			name:       "good case",
			certsLinks: certLinks,
			expected:   "testCert1\ntestCert2\n",
		},
		{
			name: "no links",
		},
		{
			name:        "bad link",
			certsLinks:  []string{badCaseLink},
			expectedErr: fmt.Sprintf("failed to load certificate from url: request to %s returned with response 404", badCaseLink),
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			certs, err := CertificateDownload(testCase.certsLinks, client)
			if err != nil {
				assert.Contains(t, testCase.expectedErr, err.Error())
			}
			assert.Equal(t, testCase.expected, string(certs))
		})
	}
}
