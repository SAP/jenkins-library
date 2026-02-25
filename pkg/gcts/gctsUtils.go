package gcts

import (
	"net/http/cookiejar"
	"net/url"

	"github.com/SAP/jenkins-library/pkg/http"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
)

func NewHttpClientOptions(username, password, proxy string, skipSSLVerification bool) (http.ClientOptions, error) {

	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return piperhttp.ClientOptions{}, errors.Wrap(err, "creating a cookie jar failed")
	}

	options := piperhttp.ClientOptions{
		CookieJar:                 cookieJar,
		Username:                  username,
		Password:                  password,
		MaxRetries:                -1,
		TransportSkipVerification: skipSSLVerification,
	}

	// Add proxy support if configured
	if proxy != "" {
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			return piperhttp.ClientOptions{}, errors.Wrap(err, "parsing proxy-url failed")
		}
		options.TransportProxy = proxyURL
		log.Entry().Infof("Using proxy: %v", proxy)
	}

	return options, nil
}
