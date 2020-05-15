package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/motemen/go-nuts/roundtime"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Client defines an http client object
type Client struct {
	maxRequestDuration       time.Duration
	transportTimeout         time.Duration
	username                 string
	password                 string
	token                    string
	logger                   *logrus.Entry
	cookieJar                http.CookieJar
	doLogRequestBodyOnDebug  bool
	doLogResponseBodyOnDebug bool
}

// ClientOptions defines the options to be set on the client
type ClientOptions struct {
	// MaxRequestDuration has a default value of "0", meaning "no maximum
	// request duration". If it is greater than 0, an overall, hard timeout
	// for the request will be enforced. This should only be used if the
	// length of the request bodies is known.
	MaxRequestDuration time.Duration
	// TransportTimeout defaults to 10 seconds, if not specified. It is
	// used for the transport layer and duration of handshakes and such.
	TransportTimeout         time.Duration
	Username                 string
	Password                 string
	Token                    string
	Logger                   *logrus.Entry
	CookieJar                http.CookieJar
	DoLogRequestBodyOnDebug  bool
	DoLogResponseBodyOnDebug bool
}

// TransportWrapper is a wrapper for central logging capabilities
type TransportWrapper struct {
	Transport                http.RoundTripper
	doLogRequestBodyOnDebug  bool
	doLogResponseBodyOnDebug bool
}

// UploadRequestData encapsulates the parameters for calling uploader.Upload()
type UploadRequestData struct {
	// Method is the HTTP method used for the request. Must be one of http.MethodPost or http.MethodPut.
	Method string
	// URL for the request
	URL string
	// File path to be stored in the created form field.
	File string
	// Form field name under which the file name will be stored.
	FileFieldName string
	// Additional form fields which will be added to the request if not nil.
	FormFields map[string]string
	// Reader from which the file contents will be read.
	FileContent io.Reader
	Header      http.Header
	Cookies     []*http.Cookie
}

// Sender provides an interface to the piper http client for uid/pwd and token authenticated requests
type Sender interface {
	SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error)
	SetOptions(options ClientOptions)
}

// Uploader provides an interface to the piper http client for uid/pwd and token authenticated requests with upload capabilities
type Uploader interface {
	Sender
	UploadRequest(method, url, file, fieldName string, header http.Header, cookies []*http.Cookie) (*http.Response, error)
	UploadFile(url, file, fieldName string, header http.Header, cookies []*http.Cookie) (*http.Response, error)
	Upload(data UploadRequestData) (*http.Response, error)
}

// UploadFile uploads a file's content as multipart-form POST request to the specified URL
func (c *Client) UploadFile(url, file, fileFieldName string, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	return c.UploadRequest(http.MethodPost, url, file, fileFieldName, header, cookies)
}

// UploadRequest uploads a file's content as multipart-form with given http method request to the specified URL
func (c *Client) UploadRequest(method, url, file, fileFieldName string, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	fileHandle, err := os.Open(file)
	if err != nil {
		return &http.Response{}, errors.Wrapf(err, "unable to locate file %v", file)
	}
	defer fileHandle.Close()
	return c.Upload(UploadRequestData{
		Method:        method,
		URL:           url,
		File:          file,
		FileFieldName: fileFieldName,
		FileContent:   fileHandle,
		Header:        header,
		Cookies:       cookies,
	})
}

// Upload uploads a file's content as multipart-form with given http method request to the specified URL
func (c *Client) Upload(data UploadRequestData) (*http.Response, error) {
	if data.Method != http.MethodPost && data.Method != http.MethodPut {
		return nil, errors.New(fmt.Sprintf("Http method %v is not allowed. Possible values are %v or %v", data.Method, http.MethodPost, http.MethodPut))
	}

	httpClient := c.initialize()

	bodyBuffer := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuffer)

	if data.FormFields != nil {
		for fieldName, fieldValue := range data.FormFields {
			err := bodyWriter.WriteField(fieldName, fieldValue)
			if err != nil {
				return &http.Response{}, errors.Wrapf(err, "error writing form field %v with value %v", fieldName, fieldValue)
			}
		}
	}

	fileWriter, err := bodyWriter.CreateFormFile(data.FileFieldName, data.File)
	if err != nil {
		return &http.Response{}, errors.Wrapf(err, "error creating form file %v for field %v", data.File, data.FileFieldName)
	}

	_, err = io.Copy(fileWriter, data.FileContent)
	if err != nil {
		return &http.Response{}, errors.Wrapf(err, "unable to copy file content of %v into request body", data.File)
	}
	err = bodyWriter.Close()

	request, err := c.createRequest(data.Method, data.URL, bodyBuffer, &data.Header, data.Cookies)
	if err != nil {
		c.logger.Debugf("New %v request to %v", data.Method, data.URL)
		return &http.Response{}, errors.Wrapf(err, "error creating %v request to %v", data.Method, data.URL)
	}

	startBoundary := strings.Index(bodyWriter.FormDataContentType(), "=") + 1
	boundary := bodyWriter.FormDataContentType()[startBoundary:]

	request.Header.Add("Content-Type", "multipart/form-data; boundary=\""+boundary+"\"")
	request.Header.Add("Connection", "Keep-Alive")

	response, err := httpClient.Do(request)
	if err != nil {
		return response, errors.Wrapf(err, "HTTP %v request to %v failed with error", data.Method, data.URL)
	}

	return c.handleResponse(response)
}

// SendRequest sends an http request with a defined method
func (c *Client) SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	httpClient := c.initialize()

	request, err := c.createRequest(method, url, body, &header, cookies)
	if err != nil {
		return &http.Response{}, errors.Wrapf(err, "error creating %v request to %v", method, url)
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return response, errors.Wrapf(err, "error calling %v", url)
	}

	return c.handleResponse(response)
}

// SetOptions sets options used for the http client
func (c *Client) SetOptions(options ClientOptions) {
	c.doLogRequestBodyOnDebug = options.DoLogRequestBodyOnDebug
	c.doLogResponseBodyOnDebug = options.DoLogResponseBodyOnDebug
	c.transportTimeout = options.TransportTimeout
	c.maxRequestDuration = options.MaxRequestDuration
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

	var transport = &TransportWrapper{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: c.transportTimeout,
			}).DialContext,
			ResponseHeaderTimeout: c.transportTimeout,
			ExpectContinueTimeout: c.transportTimeout,
			TLSHandshakeTimeout:   c.transportTimeout,
		},
		doLogRequestBodyOnDebug:  c.doLogRequestBodyOnDebug,
		doLogResponseBodyOnDebug: c.doLogResponseBodyOnDebug,
	}
	var httpClient = &http.Client{
		Timeout:   c.maxRequestDuration,
		Transport: transport,
		Jar:       c.cookieJar,
	}
	c.logger.Debugf("Transport timeout: %v, max request duration: %v", c.transportTimeout, c.maxRequestDuration)

	return httpClient
}

type contextKey struct {
	name string
}

var contextKeyRequestStart = &contextKey{"RequestStart"}

// RoundTrip is the core part of this module and implements http.RoundTripper.
// Executes HTTP request with request/response logging.
func (t *TransportWrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := context.WithValue(req.Context(), contextKeyRequestStart, time.Now())
	req = req.WithContext(ctx)

	t.logRequest(req)
	resp, err := t.Transport.RoundTrip(req)
	t.logResponse(resp)

	return resp, err
}

func (t *TransportWrapper) logRequest(req *http.Request) {
	log.Entry().Debug("--------------------------------")
	log.Entry().Debugf("--> %v request to %v", req.Method, req.URL)
	log.Entry().Debugf("headers: %v", req.Header)
	log.Entry().Debugf("cookies: %v", transformCookies(req.Cookies()))
	if t.doLogRequestBodyOnDebug {
		log.Entry().Debugf("body: %v", transformBody(req.Body))
	}
	log.Entry().Debug("--------------------------------")
}

func (t *TransportWrapper) logResponse(resp *http.Response) {
	if resp != nil {
		ctx := resp.Request.Context()
		if start, ok := ctx.Value(contextKeyRequestStart).(time.Time); ok {
			log.Entry().Debugf("<-- response %v %v (%v)", resp.StatusCode, resp.Request.URL, roundtime.Duration(time.Now().Sub(start), 2))
		} else {
			log.Entry().Debugf("<-- response %v %v", resp.StatusCode, resp.Request.URL)
		}
		if t.doLogResponseBodyOnDebug {
			log.Entry().Debugf("body: %v", transformBody(resp.Body))
		}
	} else {
		log.Entry().Debug("response <nil>")
	}
	log.Entry().Debug("--------------------------------")
}

func transformCookies(cookies []*http.Cookie) string {
	result := ""
	for _, c := range cookies {
		result = fmt.Sprintf("%v %v", result, c.String())
	}
	return result
}

func transformBody(body io.ReadCloser) string {
	if body == nil {
		return ""
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(body)
	return buf.String()
}

func (c *Client) createRequest(method, url string, body io.Reader, header *http.Header, cookies []*http.Cookie) (*http.Request, error) {
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
		c.logger.WithField("HTTP Error", "500 (Internal Server Error)").Error("Unknown error occurred.")
	}

	return response, fmt.Errorf("Request to %v returned with response %v", response.Request.URL, response.Status)
}

func (c *Client) applyDefaults() {
	if c.transportTimeout == 0 {
		c.transportTimeout = 10 * time.Second
	}
	if c.logger == nil {
		c.logger = log.Entry().WithField("package", "SAP/jenkins-library/pkg/http")
	}
}
