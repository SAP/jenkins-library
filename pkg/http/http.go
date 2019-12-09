package http

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
)

// Client defines an http client object
type Client struct {
	timeout  time.Duration
	username string
	password string
	token    string
}

// ClientOptions defines the options to be set on the client
type ClientOptions struct {
	Timeout  time.Duration
	Username string
	Password string
	Token    string
}

// Sender provides an interface to the piper http client for uid/pwd authenticated requests
type Sender interface {
	SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error)
	SetOptions(options ClientOptions)
}

// SendRequest sends an http request with a defined method
func (c *Client) SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {

	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/http")

	c.applyDefaults()

	var httpClient = &http.Client{
		Timeout: c.timeout,
	}

	logger.Debugf("Timeout set to %v", c.timeout)

	request, err := http.NewRequest(method, url, body)
	logger.Debugf("New %v request to %v", method, url)
	if err != nil {
		return &http.Response{}, errors.Wrapf(err, "error creating %v request to %v", method, url)
	}

	if header != nil {
		for name, headers := range header {
			for _, h := range headers {
				request.Header.Add(name, h)
			}
		}
	}

	if cookies != nil {
		for _, cookie := range cookies {
			request.AddCookie(cookie)
		}
	}

	if len(c.username) > 0 && len(c.password) > 0 {
		request.SetBasicAuth(c.username, c.password)
		logger.Debug("Using Basic Authentication ****/****")
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return response, errors.Wrapf(err, "error opening %v", url)
	}

	// 2xx codes do not create an error
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		return response, nil
	}

	switch response.StatusCode {
	case 401:
		logger.WithField("HTTP Error", "401 (Unauthorized)").Error("Credentials invalid, please check your user credentials!")
	case 403:
		logger.WithField("HTTP Error", "403 (Forbidden)").Error("Permission issue, please check your user permissions!")
	case 404:
		logger.WithField("HTTP Error", "404 (Not Found)").Error("Requested resource could not be found")
	case 500:
		logger.WithField("HTTP Error", "500 (Internal Server Error)").Error("Unknown error occured.")
	}

	return response, fmt.Errorf("Request to %v returned with HTTP Code %v", url, response.StatusCode)
}

// SetOptions sets options used for the http client
func (c *Client) SetOptions(options ClientOptions) {
	c.timeout = options.Timeout
	c.username = options.Username
	c.password = options.Password
	c.token = options.Token
}

func (c *Client) applyDefaults() {
	if c.timeout == 0 {
		c.timeout = time.Second * 10
	}
}
