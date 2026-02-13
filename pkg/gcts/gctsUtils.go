package gcts

import (
	"fmt"
	"net/http/cookiejar"
	"net/url"

	"github.com/SAP/jenkins-library/pkg/http"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
)

func NewHttpClientOptions(username, password, proxy string, skipSSLVerification bool) (http.ClientOptions, error) {

	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return piperhttp.ClientOptions{}, fmt.Errorf("creating a cookie jar failed: %w", err)
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
			return piperhttp.ClientOptions{}, fmt.Errorf("parsing proxy-url failed: %w", err)
		}
		options.TransportProxy = proxyURL
		log.Entry().Infof("Using proxy: %v", proxy)
	}

	return options, nil
}
