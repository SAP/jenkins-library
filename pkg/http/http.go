package http

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Client defines an http client object
type Client struct {
	timeout   time.Duration
	username  string
	password  string
	token     string
	logger    *logrus.Entry
	cookieJar http.CookieJar
}

// ClientOptions defines the options to be set on the client
type ClientOptions struct {
	Timeout   time.Duration
	Username  string
	Password  string
	Token     string
	Logger    *logrus.Entry
	CookieJar http.CookieJar
}

// Sender provides an interface to the piper http client for uid/pwd and token authenticated requests
type Sender interface {
	SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error)
	SetOptions(options ClientOptions)
}

// Uploader provides an interface to the piper http client for uid/pwd and token authenticated requests with upload capabilities
type Uploader interface {
	SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error)
	UploadRequest(method, url, file, fieldName string, header http.Header, cookies []*http.Cookie) (*http.Response, error)
	UploadFile(url, file, fieldName string, header http.Header, cookies []*http.Cookie) (*http.Response, error)
	SetOptions(options ClientOptions)
}

// UploadFile uploads a file's content as multipart-form POST request to the specified URL
func (c *Client) UploadFile(url, file, fieldName string, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	return c.UploadRequest(http.MethodPost, url, file, fieldName, header, cookies)
}

// UploadRequest uploads a file's content as multipart-form with given http method request to the specified URL
func (c *Client) UploadRequest(method, url, file, fieldName string, header http.Header, cookies []*http.Cookie) (*http.Response, error) {

	if method != http.MethodPost && method != http.MethodPut {
		return nil, errors.New(fmt.Sprintf("Http method %v is not allowed. Possible values are %v or %v", method, http.MethodPost, http.MethodPut))
	}

	httpClient := c.initialize()

	bodyBuffer := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuffer)

	fileWriter, err := bodyWriter.CreateFormFile(fieldName, file)
	if err != nil {
		return &http.Response{}, errors.Wrapf(err, "error creating form file %v for field %v", file, fieldName)
	}

	fileHandle, err := os.Open(file)
	if err != nil {
		return &http.Response{}, errors.Wrapf(err, "unable to locate file %v", file)
	}
	defer fileHandle.Close()

	_, err = io.Copy(fileWriter, fileHandle)
	if err != nil {
		return &http.Response{}, errors.Wrapf(err, "unable to copy file content of %v into request body", file)
	}
	err = bodyWriter.Close()

	request, err := c.createRequest(method, url, bodyBuffer, &header, cookies)
	if err != nil {
		c.logger.Debugf("New %v request to %v", method, url)
		return &http.Response{}, errors.Wrapf(err, "error creating %v request to %v", method, url)
	}

	startBoundary := strings.Index(bodyWriter.FormDataContentType(), "=") + 1
	boundary := bodyWriter.FormDataContentType()[startBoundary:]

	request.Header.Add("Content-Type", "multipart/form-data; boundary=\""+boundary+"\"")
	request.Header.Add("Connection", "Keep-Alive")

	response, err := httpClient.Do(request)
	if err != nil {
		return response, errors.Wrapf(err, "HTTP %v request to %v failed with error", method, url)
	}

	return c.handleResponse(response)
}

// SendRequest sends an http request with a defined method
func (c *Client) SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	httpClient := c.initialize()

	request, err := c.createRequest(method, url, body, &header, cookies)
	if err != nil {
		c.logger.Debugf("New %v request to %v", method, url)
		return &http.Response{}, errors.Wrapf(err, "error creating %v request to %v", method, url)
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return response, errors.Wrapf(err, "error opening %v", url)
	}

	return c.handleResponse(response)
}

// SetOptions sets options used for the http client
func (c *Client) SetOptions(options ClientOptions) {
	c.timeout = options.Timeout
	c.username = options.Username
	c.password = options.Password
	c.token = options.Token

	if options.Logger != nil {
		c.logger = options.Logger
	} else {
		c.logger = log.Entry().WithField("package", "SAP/jenkins-library/pkg/http")
	}
	c.cookieJar = options.CookieJar
}

func (c *Client) initialize() *http.Client {
	c.applyDefaults()
	c.logger = log.Entry().WithField("package", "SAP/jenkins-library/pkg/http")

	var httpClient = &http.Client{
		Timeout: c.timeout,
		Jar:     c.cookieJar,
	}
	c.logger.Debugf("Timeout set to %v", c.timeout)

	return httpClient
}

func (c *Client) createRequest(method, url string, body io.Reader, header *http.Header, cookies []*http.Cookie) (*http.Request, error) {
	c.logger.Debugf("New %v request to %v", method, url)
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return &http.Request{}, err
	}

	if header != nil {
		for name, headers := range *header {
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
		c.logger.Debug("Using Basic Authentication ****/****")
	}

	if len(c.token) > 0 {
		request.Header.Add("Authorization", c.token)
	}

	return request, nil
}

func (c *Client) handleResponse(response *http.Response) (*http.Response, error) {
	// 2xx codes do not create an error
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		return response, nil
	}

	switch response.StatusCode {
	case http.StatusUnauthorized:
		c.logger.WithField("HTTP Error", "401 (Unauthorized)").Error("Credentials invalid, please check your user credentials!")
	case http.StatusForbidden:
		c.logger.WithField("HTTP Error", "403 (Forbidden)").Error("Permission issue, please check your user permissions!")
	case http.StatusNotFound:
		c.logger.WithField("HTTP Error", "404 (Not Found)").Error("Requested resource could not be found")
	case http.StatusInternalServerError:
		c.logger.WithField("HTTP Error", "500 (Internal Server Error)").Error("Unknown error occured.")
	}

	return response, fmt.Errorf("Request to %v returned with HTTP Code %v", response.Request.URL, response.StatusCode)
}

func (c *Client) applyDefaults() {
	if c.timeout == 0 {
		c.timeout = time.Second * 10
	}
	if c.logger == nil {
		c.logger = log.Entry().WithField("package", "SAP/jenkins-library/pkg/http")
	}
}
