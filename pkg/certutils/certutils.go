package certutils

import (
	"io/ioutil"
	"net/http"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/pkg/errors"
)

func CertificateUpdate(certLinks []string, httpClient piperhttp.Sender, fileUtils piperutils.FileUtils, caCertsFile string) error {
	caCerts, err := fileUtils.FileRead(caCertsFile)
	if err != nil {
		return errors.Wrapf(err, "failed to load file '%v'", caCertsFile)
	}

	for _, link := range certLinks {
		response, err := httpClient.SendRequest(http.MethodGet, link, nil, nil, nil)
		if err != nil {
			return errors.Wrap(err, "failed to load certificate from url")
		}

		content, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return errors.Wrap(err, "error reading response")
		}
		_ = response.Body.Close()
		content = append(content, []byte("\n")...)
		caCerts = append(caCerts, content...)
	}
	err = fileUtils.FileWrite(caCertsFile, caCerts, 0644)
	if err != nil {
		return errors.Wrapf(err, "failed to update file '%v'", caCertsFile)
	}
	return nil
}
