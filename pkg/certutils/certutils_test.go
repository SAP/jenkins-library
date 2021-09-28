package certutils

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

const (
	caCertsFile = "/kaniko/ssl/certs/ca-certificates.crt"
)

func TestCertificateUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte("testCert"))
	}))
	// Close the server when test finishes
	defer server.Close()
	certLinks := []string{server.URL, server.URL}

	t.Run("success case", func(t *testing.T) {
		client := &piperhttp.Client{}
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
		client := &piperhttp.Client{}
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
		client := &piperhttp.Client{}

		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile(caCertsFile, []byte("initial cert\n"))

		err := CertificateUpdate([]string{"http://non-existing-url"}, client, fileUtils, caCertsFile)
		assert.Contains(t, err.Error(), "failed to load certificate from url: HTTP GET request to http://non-existing-url failed: Get \"http://non-existing-url\": dial tcp: lookup non-existing-url")
	})

}
