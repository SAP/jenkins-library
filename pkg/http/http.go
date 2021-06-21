package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/reporting"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/motemen/go-nuts/roundtime"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Client defines an http client object
type Client struct {
	maxRequestDuration        time.Duration
	maxRetries                int
	transportTimeout          time.Duration
	transportSkipVerification bool
	username                  string
	password                  string
	token                     string
	logger                    *logrus.Entry
	cookieJar                 http.CookieJar
	doLogRequestBodyOnDebug   bool
	doLogResponseBodyOnDebug  bool
	useDefaultTransport       bool
	trustedCerts              []string
}

// ClientOptions defines the options to be set on the client
type ClientOptions struct {
	// MaxRequestDuration has a default value of "0", meaning "no maximum
	// request duration". If it is greater than 0, an overall, hard timeout
	// for the request will be enforced. This should only be used if the
	// length of the request bodies is known.
	MaxRequestDuration time.Duration
	MaxRetries         int
	// TransportTimeout defaults to 3 minutes, if not specified. It is
	// used for the transport layer and duration of handshakes and such.
	TransportTimeout          time.Duration
	TransportSkipVerification bool
	Username                  string
	Password                  string
	Token                     string
	Logger                    *logrus.Entry
	CookieJar                 http.CookieJar
	DoLogRequestBodyOnDebug   bool
	DoLogResponseBodyOnDebug  bool
	UseDefaultTransport       bool
	TrustedCerts              []string
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

	return c.Send(request)
}

// SendRequest sends an http request with a defined method
//
// On error, any Response can be ignored and the Response.Body
// does not need to be closed.
func (c *Client) SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	request, err := c.createRequest(method, url, body, &header, cookies)
	if err != nil {
		return &http.Response{}, errors.Wrapf(err, "error creating %v request to %v", method, url)
	}

	return c.Send(request)
}

// Send sends an http request
func (c *Client) Send(request *http.Request) (*http.Response, error) {
	httpClient := c.initialize()
	response, err := httpClient.Do(request)
	if err != nil {
		return response, errors.Wrapf(err, "HTTP %v request to %v failed", request.Method, request.URL)
	}
	return c.handleResponse(response, request.URL.String())
}

// SetOptions sets options used for the http client
func (c *Client) SetOptions(options ClientOptions) {
	c.doLogRequestBodyOnDebug = options.DoLogRequestBodyOnDebug
	c.doLogResponseBodyOnDebug = options.DoLogResponseBodyOnDebug
	c.useDefaultTransport = options.UseDefaultTransport
	c.transportTimeout = options.TransportTimeout
	c.transportSkipVerification = options.TransportSkipVerification
	c.maxRequestDuration = options.MaxRequestDuration
	c.username = options.Username
	c.password = options.Password
	c.token = options.Token
	if options.MaxRetries < 0 {
		c.maxRetries = 0
	} else if options.MaxRetries == 0 {
		c.maxRetries = 15
	} else {
		c.maxRetries = options.MaxRetries
	}

	if options.Logger != nil {
		c.logger = options.Logger
	} else {
		c.logger = log.Entry().WithField("package", "SAP/jenkins-library/pkg/http")
	}
	c.cookieJar = options.CookieJar
	c.trustedCerts = options.TrustedCerts
}

// StandardClient returns a stdlib *http.Client which respects the custom settings.
func (c *Client) StandardClient() *http.Client {
	return c.initialize()
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
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: c.transportSkipVerification,
			},
		},
		doLogRequestBodyOnDebug:  c.doLogRequestBodyOnDebug,
		doLogResponseBodyOnDebug: c.doLogResponseBodyOnDebug,
	}

	if (len(c.trustedCerts)) > 0 {
		log.Entry().Info("Anil test : adding certs")
		c.configureTLSToTrustCertificates(transport)
	} else {
		log.Entry().Info("Anil test : not adding certs")
	}

	var httpClient *http.Client
	if c.maxRetries > 0 {
		retryClient := retryablehttp.NewClient()
		localLogger := log.Entry()
		localLogger.Level = logrus.DebugLevel
		retryClient.Logger = localLogger
		retryClient.HTTPClient.Timeout = c.maxRequestDuration
		retryClient.HTTPClient.Jar = c.cookieJar
		retryClient.RetryMax = c.maxRetries
		if !c.useDefaultTransport {
			retryClient.HTTPClient.Transport = transport
		}
		retryClient.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
			if err != nil && (strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "timed out") || strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "connection reset")) {
				// Assuming timeouts, resets, and similar could be retried
				return true, nil
			}
			return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
		}
		httpClient = retryClient.StandardClient()
	} else {
		httpClient = &http.Client{}
		httpClient.Timeout = c.maxRequestDuration
		httpClient.Jar = c.cookieJar
		if !c.useDefaultTransport {
			httpClient.Transport = transport
		}
	}

	if c.transportSkipVerification {
		c.logger.Debugf("TLS verification disabled")
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
	log.Entry().Debugf("headers: %v", transformHeaders(req.Header))
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

func transformHeaders(header http.Header) http.Header {
	var h http.Header = map[string][]string{}
	for name, value := range header {
		if name == "Authorization" {
			for _, v := range value {
				// The format of the Authorization header value is: <type> <cred>.
				// We don't register the full string since only the part after
				// the first token is the secret in the narrower sense (applies at
				// least for basic auth)
				log.RegisterSecret(strings.Join(strings.Split(v, " ")[1:], " "))
			}
			// Since
			//   1.) The auth header type itself might serve as a vector for an
			//       intrusion
			//   2.) We cannot make assumtions about the structure of the auth
			//       header value since that depends on the type, e.g. several tokens
			//       where only some of the tokens define the secret
			// we hide the full auth header value anyway in order to be on the
			// save side.
			value = []string{"<set>"}
		}
		h[name] = value
	}
	return h
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

	if len(c.username) > 0 {
		request.SetBasicAuth(c.username, c.password)
		c.logger.Debug("Using Basic Authentication ****/****")
	}

	if len(c.token) > 0 {
		request.Header.Add("Authorization", c.token)
	}

	return request, nil
}

func (c *Client) handleResponse(response *http.Response, url string) (*http.Response, error) {
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
		c.logger.WithField("HTTP Error", "404 (Not Found)").Errorf("Requested resource ('%s') could not be found", url)
	case http.StatusInternalServerError:
		c.logger.WithField("HTTP Error", "500 (Internal Server Error)").Error("Unknown error occurred.")
	}

	return response, fmt.Errorf("Request to %v returned with response %v", response.Request.URL, response.Status)
}

func (c *Client) applyDefaults() {
	if c.transportTimeout == 0 {
		c.transportTimeout = 3 * time.Minute
	}
	if c.logger == nil {
		c.logger = log.Entry().WithField("package", "SAP/jenkins-library/pkg/http")
	}
}

func (c *Client) configureTLSToTrustCertificates(transport *TransportWrapper) error {

	trustStoreDir, err := getWorkingDirForTrustStore()
	fileUtils := &piperutils.Files{}
	if err != nil {
		return errors.Wrap(err, "failed to create trust store directory")
	}

	for _, certificate := range c.trustedCerts {
		filename := path.Base(certificate) // decode?
		target := filepath.Join(trustStoreDir, filename)
		if exists, _ := fileUtils.FileExists(target); !exists {
			log.Entry().WithField("source", certificate).WithField("target", target).Info("Downloading TLS certificate")
			// download certificate
			request, err := http.NewRequest("GET", certificate, nil)
			if err != nil {
				return err
			}

			httpClient := &http.Client{}
			httpClient.Timeout = c.maxRequestDuration
			httpClient.Jar = c.cookieJar
			if !c.useDefaultTransport {
				httpClient.Transport = transport
			}
			response, err := httpClient.Do(request)
			if err != nil {
				return errors.Wrapf(err, "HTTP %v request to %v failed", request.Method, request.URL)
			}

			if response.StatusCode >= 200 && response.StatusCode < 300 {
				// Get the SystemCertPool, continue with an empty pool on error
				rootCAs, _ := x509.SystemCertPool()
				if rootCAs == nil {
					rootCAs = x509.NewCertPool()
				}

				certs, err := ioutil.ReadFile(target)
				if err != nil {
					return errors.Wrapf(err, "Failed to append %q to RootCAs: %v", target, err)
				}

				// Append our cert to the system pool
				if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
					log.Entry().Info("No certs appended, using system certs only")
				}

				transport = &TransportWrapper{
					Transport: &http.Transport{
						DialContext: (&net.Dialer{
							Timeout: c.transportTimeout,
						}).DialContext,
						ResponseHeaderTimeout: c.transportTimeout,
						ExpectContinueTimeout: c.transportTimeout,
						TLSHandshakeTimeout:   c.transportTimeout,
						TLSClientConfig: &tls.Config{
							InsecureSkipVerify: false,
							RootCAs:            rootCAs,
						},
					},
					doLogRequestBodyOnDebug:  c.doLogRequestBodyOnDebug,
					doLogResponseBodyOnDebug: c.doLogResponseBodyOnDebug,
				}

				log.Entry().Infof("%v appended to root CA", certificate)

			} else {
				return errors.Wrapf(err, "Download of TLS certificate %v failed with status code %v", certificate, response.StatusCode)
			}
		} else {
			log.Entry().Infof("skipped %v append to root CA it exists", certificate)
		}

	}
	return nil
}

func getWorkingDirForTrustStore() (string, error) {
	fileUtils := &piperutils.Files{}
	if exists, _ := fileUtils.DirExists(reporting.StepReportDirectory); !exists {
		err := fileUtils.MkdirAll(".pipeline/trustStore", 0777)
		if err != nil {
			return "", errors.Wrap(err, "failed to create trust store directory")
		}
	}
	return ".pipeline/trustStore", nil
}

// ParseHTTPResponseBodyXML parses a XML http response into a given interface
func ParseHTTPResponseBodyXML(resp *http.Response, response interface{}) error {
	if resp == nil {
		return errors.Errorf("cannot parse HTTP response with value <nil>")
	}

	bodyText, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return errors.Wrap(readErr, "HTTP response body could not be read")
	}

	marshalErr := xml.Unmarshal(bodyText, &response)
	if marshalErr != nil {
		return errors.Wrapf(marshalErr, "HTTP response body could not be parsed as XML: %v", string(bodyText))
	}

	return nil
}

// ParseHTTPResponseBodyJSON parses a JSON http response into a given interface
func ParseHTTPResponseBodyJSON(resp *http.Response, response interface{}) error {
	if resp == nil {
		return errors.Errorf("cannot parse HTTP response with value <nil>")
	}

	bodyText, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return errors.Wrapf(readErr, "HTTP response body could not be read")
	}

	marshalErr := json.Unmarshal(bodyText, &response)
	if marshalErr != nil {
		return errors.Wrapf(marshalErr, "HTTP response body could not be parsed as JSON: %v", string(bodyText))
	}

	return nil
}
