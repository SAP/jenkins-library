package certutils

import (
	"fmt"
	"io"
	"net/http"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

// CertificateUpdate adds certificates to the given truststore
func CertificateUpdate(certLinks []string, httpClient piperhttp.Sender, fileUtils piperutils.FileUtils, caCertsFile string) error {
	// TODO this implementation doesn't work on non-linux machines, is not failsafe and should be implemented differently

	if len(certLinks) == 0 {
		return nil
	}

	caCerts, err := fileUtils.FileRead(caCertsFile)
	if err != nil {
		return fmt.Errorf("failed to load file '%v': %w", caCertsFile, err)
	}

	byteCerts, err := CertificateDownload(certLinks, httpClient)
	if err != nil {
		return err
	}

	caCerts = append(caCerts, byteCerts...)

	err = fileUtils.FileWrite(caCertsFile, caCerts, 0644)
	if err != nil {
		return fmt.Errorf("failed to update file '%v': %w", caCertsFile, err)
	}
	return nil
}

// CertificateDownload downloads certificates and returns them as a byte slice
func CertificateDownload(certLinks []string, client piperhttp.Sender) ([]byte, error) {
	if len(certLinks) == 0 {
		return nil, nil
	}

	var certs []byte
	for _, certLink := range certLinks {
		log.Entry().Debugf("Downloading CA certificate from URL: %s", certLink)
		response, err := client.SendRequest(http.MethodGet, certLink, nil, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to load certificate from url: %w", err)
		}

		content, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}
		_ = response.Body.Close()
		content = append(content, []byte("\n")...)
		certs = append(certs, content...)
	}

	return certs, nil
}
