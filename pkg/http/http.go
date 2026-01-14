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
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
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
	transportProxy            *url.URL
	username                  string
	password                  string
	token                     string
	logger                    *logrus.Entry
	cookieJar                 http.CookieJar
	doLogRequestBodyOnDebug   bool
	doLogResponseBodyOnDebug  bool
	useDefaultTransport       bool
	trustedCerts              []string
	certificates              []tls.Certificate // contains one or more certificate chains to present to the other side of the connection (client-authentication)
	fileUtils                 piperutils.FileUtils
	httpClient                *http.Client
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
	TransportProxy            *url.URL
	Username                  string
	Password                  string
	Token                     string
	Logger                    *logrus.Entry
	CookieJar                 http.CookieJar
	DoLogRequestBodyOnDebug   bool
	DoLogResponseBodyOnDebug  bool
	UseDefaultTransport       bool
	TrustedCerts              []string          // defines the set of root certificate authorities that clients use when verifying server certificates
	Certificates              []tls.Certificate // contains one or more certificate chains to present to the other side of the connection (client-authentication)
}

// TransportWrapper is a wrapper for central round trip capabilities
type TransportWrapper struct {
	Transport                http.RoundTripper
	doLogRequestBodyOnDebug  bool
	doLogResponseBodyOnDebug bool
	username                 string
	password                 string
	token                    string
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
	UploadType  string
}

// Sender provides an interface to the piper http client for uid/pwd and token authenticated requests
type Sender interface {
	SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error)
	SetOptions(options ClientOptions)
}

// Uploader provides an interface to the piper http client for uid/pwd and token authenticated requests with upload capabilities
type Uploader interface {
	Sender
	UploadRequest(method, url, file, fieldName string, header http.Header, cookies []*http.Cookie, uploadType string) (*http.Response, error)
	UploadFile(url, file, fieldName string, header http.Header, cookies []*http.Cookie, uploadType string) (*http.Response, error)
	Upload(data UploadRequestData) (*http.Response, error)
}

// fileUtils lazy initializes the utils
func (c *Client) getFileUtils() piperutils.FileUtils {
	if c.fileUtils == nil {
		c.fileUtils = &piperutils.Files{}
	}

	return c.fileUtils
}

// UploadFile uploads a file's content as multipart-form POST request to the specified URL
func (c *Client) UploadFile(url, file, fileFieldName string, header http.Header, cookies []*http.Cookie, uploadType string) (*http.Response, error) {
	return c.UploadRequest(http.MethodPost, url, file, fileFieldName, header, cookies, uploadType)
}

// UploadRequest uploads a file's content as multipart-form with given http method request to the specified URL
func (c *Client) UploadRequest(method, url, file, fileFieldName string, header http.Header, cookies []*http.Cookie, uploadType string) (*http.Response, error) {
	fileHandle, err := c.getFileUtils().Open(file)

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
		UploadType:    uploadType,
	})
}

// Upload uploads a file's content as multipart-form or pure binary with given http method request to the specified URL
func (c *Client) Upload(data UploadRequestData) (*http.Response, error) {
	if data.Method != http.MethodPost && data.Method != http.MethodPut {
		return nil, errors.New(fmt.Sprintf("Http method %v is not allowed. Possible values are %v or %v", data.Method, http.MethodPost, http.MethodPut))
	}

	// Binary upload :: other options ("binary" or "form").
	if data.UploadType == "binary" {
		request, err := c.createRequest(data.Method, data.URL, data.FileContent, &data.Header, data.Cookies)
		if err != nil {
			c.logger.Debugf("New %v request to %v (binary upload)", data.Method, data.URL)
			return &http.Response{}, errors.Wrapf(err, "error creating %v request to %v (binary upload)", data.Method, data.URL)
		}
		request.Header.Add("Content-Type", "application/octet-stream")
		request.Header.Add("Connection", "Keep-Alive")

		return c.Send(request)

	} else { // For form upload

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

		_, err = piperutils.CopyData(fileWriter, data.FileContent)
		if err != nil {
			return &http.Response{}, errors.Wrapf(err, "unable to copy file content of %v into request body", data.File)
		}
		err = bodyWriter.Close()
		if err != nil {
			log.Entry().Warn("failed to close writer on request body")
		}

		request, err := c.createRequest(data.Method, data.URL, bodyBuffer, &data.Header, data.Cookies)
		if err != nil {
			c.logger.Debugf("new %v request to %v", data.Method, data.URL)
			return &http.Response{}, errors.Wrapf(err, "error creating %v request to %v", data.Method, data.URL)
		}

		startBoundary := strings.Index(bodyWriter.FormDataContentType(), "=") + 1
		boundary := bodyWriter.FormDataContentType()[startBoundary:]
		request.Header.Add("Content-Type", "multipart/form-data; boundary=\""+boundary+"\"")
		request.Header.Add("Connection", "Keep-Alive")

		return c.Send(request)
	}
}

// SendRequest sends a http request with a defined method
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

// Send sends a http request
func (c *Client) Send(request *http.Request) (*http.Response, error) {
	httpClient := c.initializeHttpClient()
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
	c.transportProxy = options.TransportProxy
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
	c.fileUtils = &piperutils.Files{}
	c.certificates = options.Certificates
}

// SetFileUtils can be used to overwrite the default file utils
func (c *Client) SetFileUtils(fileUtils piperutils.FileUtils) {
	c.fileUtils = fileUtils
}

// StandardClient returns a stdlib *http.Client which respects the custom settings.
func (c *Client) StandardClient() *http.Client {
	return c.initializeHttpClient()
}

func (c *Client) initializeHttpClient() *http.Client {
	if c.httpClient != nil {
		return c.httpClient
	}

	c.applyDefaults()

	var transport = &TransportWrapper{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: c.transportTimeout,
			}).DialContext,
			Proxy:                 http.ProxyURL(c.transportProxy),
			ResponseHeaderTimeout: c.transportTimeout,
			ExpectContinueTimeout: c.transportTimeout,
			TLSHandshakeTimeout:   c.transportTimeout,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: c.transportSkipVerification,
				Certificates:       c.certificates,
			},
		},
		doLogRequestBodyOnDebug:  c.doLogRequestBodyOnDebug,
		doLogResponseBodyOnDebug: c.doLogResponseBodyOnDebug,
		token:                    c.token,
		username:                 c.username,
		password:                 c.password,
	}

	if len(c.trustedCerts) > 0 && !c.useDefaultTransport && !c.transportSkipVerification {
		log.Entry().Debug("adding certs for tls to trust")
		err := c.configureTLSToTrustCertificates(transport)
		if err != nil {
			log.Entry().Infof("adding certs for tls config failed : %v, continuing with the existing tls config", err)
		}
	} else {
		log.Entry().Debug("no trusted certs found / using default transport / insecure skip set to true / : continuing with existing tls config")
	}

	if c.maxRetries > 0 {
		retryClient := retryablehttp.NewClient()
		retryClient.Logger = c.logger
		retryClient.HTTPClient.Timeout = c.maxRequestDuration
		retryClient.HTTPClient.Jar = c.cookieJar
		retryClient.RetryMax = c.maxRetries
		if !c.useDefaultTransport {
			retryClient.HTTPClient.Transport = transport
		} else {
			retryClient.HTTPClient.Transport = &TransportWrapper{
				Transport:                retryClient.HTTPClient.Transport,
				doLogRequestBodyOnDebug:  c.doLogRequestBodyOnDebug,
				doLogResponseBodyOnDebug: c.doLogResponseBodyOnDebug,
				token:                    c.token,
				username:                 c.username,
				password:                 c.password}
		}
		retryClient.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
			if err != nil && (strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "timed out") || strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "connection reset")) {
				// Assuming timeouts, resets, and similar could be retried
				return true, nil
			}
			return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
		}
		c.httpClient = retryClient.StandardClient()
	} else {
		c.httpClient = &http.Client{
			Timeout: c.maxRequestDuration,
			Jar:     c.cookieJar,
		}
		if !c.useDefaultTransport {
			c.httpClient.Transport = transport
		}
	}

	if c.transportSkipVerification {
		c.logger.Debugf("TLS verification disabled")
	}

	c.logger.Debugf("Transport timeout: %v, max request duration: %v", c.transportTimeout, c.maxRequestDuration)

	return c.httpClient
}

type contextKey struct {
	name string
}

var contextKeyRequestStart = &contextKey{"RequestStart"}
var authHeaderKey = "Authorization"

// RoundTrip is the core part of this module and implements http.RoundTripper.
// Executes HTTP requests with request/response logging.
func (t *TransportWrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := context.WithValue(req.Context(), contextKeyRequestStart, time.Now())
	req = req.WithContext(ctx)

	handleAuthentication(req, t.username, t.password, t.token)

	t.logRequest(req)

	resp, err := t.Transport.RoundTrip(req)

	t.logResponse(resp)

	return resp, err
}

func handleAuthentication(req *http.Request, username, password, token string) {
	// Handle authentication if not done already
	if (len(username) > 0 || len(password) > 0) && len(req.Header.Get(authHeaderKey)) == 0 {
		req.SetBasicAuth(username, password)
		log.Entry().Debug("Using Basic Authentication ****/****\n")
	}
	if len(token) > 0 && len(req.Header.Get(authHeaderKey)) == 0 {
		req.Header.Add(authHeaderKey, token)
		log.Entry().Debug("Using Token Authentication ****")
	}
}

func (t *TransportWrapper) logRequest(req *http.Request) {
	log.Entry().Debug("--------------------------------")
	log.Entry().Debugf("--> %v request to %v", req.Method, req.URL)
	log.Entry().Debugf("headers: %v", transformHeaders(req.Header))
	log.Entry().Debugf("cookies: %v", transformCookies(req.Cookies()))
	if t.doLogRequestBodyOnDebug && req.Header.Get("Content-Type") == "application/octet-stream" {
		// skip logging byte content as it's useless
	} else if t.doLogRequestBodyOnDebug && req.Body != nil {
		var buf bytes.Buffer
		tee := io.TeeReader(req.Body, &buf)
		log.Entry().Debugf("body: %v", transformBody(tee))
		req.Body = io.NopCloser(bytes.NewReader(buf.Bytes()))
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
		if t.doLogResponseBodyOnDebug && resp.Body != nil {
			var buf bytes.Buffer
			tee := io.TeeReader(resp.Body, &buf)
			log.Entry().Debugf("body: %v", transformBody(tee))
			resp.Body = io.NopCloser(bytes.NewReader(buf.Bytes()))
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
			//   2.) We cannot make assumptions about the structure of the auth
			//       header value since that depends on the type, e.g. several tokens
			//       where only some tokens define the secret
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

const maxLogBodyLength = 2 * 1024

func transformBody(body io.Reader) string {
	if body == nil {
		return ""
	}

	data, _ := io.ReadAll(body)
	if len(data) > maxLogBodyLength {
		return string(data[:maxLogBodyLength]) + "...(truncated)"
	}
	return string(data)
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

	handleAuthentication(request, c.username, c.password, c.token)

	for _, cookie := range cookies {
		request.AddCookie(cookie)
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

	return response, fmt.Errorf("request to %v returned with response %v", response.Request.URL, response.Status)
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
	if err != nil {
		return errors.Wrap(err, "failed to create trust store directory")
	}
	/* insecure := flag.Bool("insecure-ssl", false, "Accept/Ignore all server SSL certificates") */
	// Get the SystemCertPool, continue with an empty pool on error
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		log.Entry().Debugf("Caught error on store lookup %v", err)
	}

	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	*transport = TransportWrapper{
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
				Certificates:       c.certificates,
			},
		},
		doLogRequestBodyOnDebug:  c.doLogRequestBodyOnDebug,
		doLogResponseBodyOnDebug: c.doLogResponseBodyOnDebug,
		token:                    c.token,
		username:                 c.username,
		password:                 c.password,
	}

	for _, certificate := range c.trustedCerts {
		filename := path.Base(certificate)
		filename = strings.ReplaceAll(filename, " ", "")
		target := filepath.Join(trustStoreDir, filename)
		if exists, _ := c.getFileUtils().FileExists(target); !exists {
			log.Entry().WithField("source", certificate).WithField("target", target).Info("Downloading TLS certificate")
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
				defer response.Body.Close()
				parent := filepath.Dir(target)
				if len(parent) > 0 {
					if err = c.getFileUtils().MkdirAll(parent, 0777); err != nil {
						return err
					}
				}
				fileHandler, err := c.getFileUtils().Create(target)
				if err != nil {
					return errors.Wrapf(err, "unable to create file %v", filename)
				}
				defer fileHandler.Close()

				numWritten, err := io.Copy(fileHandler, response.Body)
				if err != nil {
					return errors.Wrapf(err, "unable to copy content from url to file %v", filename)
				}
				log.Entry().Debugf("wrote %v bytes from response body to file", numWritten)

				certs, err := os.ReadFile(target)
				if err != nil {
					return errors.Wrapf(err, "failed to read cert file %v", certificate)
				}
				// Append our cert to the system pool
				ok := rootCAs.AppendCertsFromPEM(certs)
				if !ok {
					return errors.Errorf("failed to append %v to root CA store", certificate)
				}
				log.Entry().Infof("%v appended to root CA successfully", certificate)
			} else {
				return errors.Wrapf(err, "Download of TLS certificate %v failed with status code %v", certificate, response.StatusCode)
			}
		} else {
			log.Entry().Debugf("existing certificate file %v found, appending it to rootCA", target)
			certs, err := os.ReadFile(target)
			if err != nil {
				return errors.Wrapf(err, "failed to read cert file %v", certificate)
			}
			// Append our cert to the system pool
			ok := rootCAs.AppendCertsFromPEM(certs)
			if !ok {
				return errors.Errorf("failed to append %v to root CA store", certificate)
			}
			log.Entry().Debugf("%v appended to root CA successfully", certificate)
		}

	}
	return nil
}

// TrustStoreDirectory default truststore location
const TrustStoreDirectory = ".pipeline/trustStore"

func getWorkingDirForTrustStore() (string, error) {
	fileUtils := &piperutils.Files{}
	if exists, _ := fileUtils.DirExists(TrustStoreDirectory); !exists {
		err := fileUtils.MkdirAll(TrustStoreDirectory, 0777)
		if err != nil {
			return "", errors.Wrap(err, "failed to create trust store directory")
		}
	}
	return TrustStoreDirectory, nil
}

// ParseHTTPResponseBodyXML parses an XML http response into a given interface
func ParseHTTPResponseBodyXML(resp *http.Response, response interface{}) error {
	if resp == nil {
		return errors.Errorf("cannot parse HTTP response with value <nil>")
	}

	bodyText, readErr := io.ReadAll(resp.Body)
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

	bodyText, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return errors.Wrapf(readErr, "HTTP response body could not be read")
	}

	marshalErr := json.Unmarshal(bodyText, &response)
	if marshalErr != nil {
		return errors.Wrapf(marshalErr, "HTTP response body could not be parsed as JSON: %v", string(bodyText))
	}

	return nil
}
