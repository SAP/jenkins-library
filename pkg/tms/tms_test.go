package tms

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/stretchr/testify/assert"
)

type uploaderMock struct {
	token                    string
	httpMethod               string
	httpStatusCode           int
	urlCalled                string
	requestBody              string
	responseBody             string
	header                   http.Header
	isTechnicalErrorExpected bool
}

func (um *uploaderMock) SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	if um.isTechnicalErrorExpected {
		return nil, errors.New("Provoked technical error")
	}
	um.httpMethod = method
	um.urlCalled = url
	um.header = header
	if body != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(body)
		um.requestBody = buf.String()
	}
	var httpError error
	if um.httpStatusCode >= 300 {
		httpError = fmt.Errorf("http error %v", um.httpStatusCode)
	}
	return &http.Response{StatusCode: um.httpStatusCode, Body: ioutil.NopCloser(strings.NewReader(um.responseBody))}, httpError
}

func (um *uploaderMock) UploadFile(url, file, fieldName string, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	um.httpMethod = http.MethodPost
	um.urlCalled = url
	um.header = header
	return &http.Response{StatusCode: um.httpStatusCode, Body: ioutil.NopCloser(bytes.NewReader([]byte(um.responseBody)))}, nil
}

func (um *uploaderMock) UploadRequest(method, url, file, fieldName string, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	um.httpMethod = http.MethodPost
	um.urlCalled = url
	um.header = header
	return &http.Response{StatusCode: um.httpStatusCode, Body: ioutil.NopCloser(bytes.NewReader([]byte(um.responseBody)))}, nil
}

func (um *uploaderMock) Upload(_ piperHttp.UploadRequestData) (*http.Response, error) {
	return &http.Response{}, fmt.Errorf("not implemented")
}

func (um *uploaderMock) SetOptions(options piperHttp.ClientOptions) {
	um.token = options.Token
}

func TestGetOAuthToken(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/tms_test")
	t.Run("test success", func(t *testing.T) {
		uploaderMock := uploaderMock{responseBody: `{"token_type":"bearer","access_token":"testOAuthToken","expires_in":54321}`, httpStatusCode: http.StatusOK}
		communicationInstance := CommunicationInstance{uaaUrl: "https://dummy.sap.com", clientId: "testClientId", clientSecret: "testClientSecret", httpClient: &uploaderMock, logger: logger, isVerbose: false}

		token, err := communicationInstance.getOAuthToken()

		// TODO: how to check that certain messages were printed into the log?
		assert.NoError(t, err, "Error occurred, but none expected")
		assert.Equal(t, "https://dummy.sap.com/oauth/token/?grant_type=client_credentials&response_type=token", uploaderMock.urlCalled, "Called url incorrect")
		assert.Equal(t, http.MethodPost, uploaderMock.httpMethod, "Http method incorrect")
		assert.Equal(t, []string{"application/x-www-form-urlencoded"}, uploaderMock.header[http.CanonicalHeaderKey("content-type")], "Content-Type header incorrect")
		assert.Equal(t, []string{"Basic dGVzdENsaWVudElkOnRlc3RDbGllbnRTZWNyZXQ="}, uploaderMock.header[http.CanonicalHeaderKey("authorization")], "Authorizatoin header incorrect")
		assert.Equal(t, "grant_type=password&password=testClientSecret&username=testClientId", uploaderMock.requestBody, "Request body incorrect")
		assert.Equal(t, "testOAuthToken", token, "Token incorrect")
	})

	t.Run("test error", func(t *testing.T) {
		uploaderMock := uploaderMock{responseBody: `Bad request provided`, httpStatusCode: http.StatusBadRequest}
		communicationInstance := CommunicationInstance{uaaUrl: "https://dummy.sap.com", clientId: "testClientId", clientSecret: "testClientSecret", httpClient: &uploaderMock, logger: logger, isVerbose: false}

		_, err := communicationInstance.getOAuthToken()

		// TODO: how to check that certain messages were printed into the log?
		assert.Error(t, err, "Error expected, but none occurred")
		assert.Equal(t, "https://dummy.sap.com/oauth/token/?grant_type=client_credentials&response_type=token", uploaderMock.urlCalled, "Called url incorrect")
		assert.Equal(t, http.MethodPost, uploaderMock.httpMethod, "Http method incorrect")
		assert.Equal(t, []string{"application/x-www-form-urlencoded"}, uploaderMock.header[http.CanonicalHeaderKey("content-type")], "Content-Type header incorrect")
		assert.Equal(t, []string{"Basic dGVzdENsaWVudElkOnRlc3RDbGllbnRTZWNyZXQ="}, uploaderMock.header[http.CanonicalHeaderKey("authorization")], "Authorizatoin header incorrect")
		assert.Equal(t, "grant_type=password&password=testClientSecret&username=testClientId", uploaderMock.requestBody, "Request body incorrect")
	})
}

func TestSendRequest(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/tms_test")
	t.Run("test success against uaa", func(t *testing.T) {
		uploaderMock := uploaderMock{responseBody: `{"someKey": "someValue"}`, httpStatusCode: http.StatusOK}
		communicationInstance := CommunicationInstance{uaaUrl: "https://dummy.sap.com", tmsUrl: "https://tms.dummy.sap.com", httpClient: &uploaderMock, logger: logger, isVerbose: false}

		urlFormData := url.Values{
			"key1": {"value1"},
		}
		header := http.Header{}
		header.Add("Authorization", "Basic dGVzdENsaWVudElkOnRlc3RDbGllbnRTZWNyZXQ=")
		data, err := sendRequest(&communicationInstance, http.MethodPost, "/test/?param1=value1", strings.NewReader(urlFormData.Encode()), header, http.StatusOK, true)

		assert.NoError(t, err, "Error occurred, but none expected")
		assert.Equal(t, "https://dummy.sap.com/test/?param1=value1", uploaderMock.urlCalled, "Called url incorrect")
		assert.Equal(t, http.MethodPost, uploaderMock.httpMethod, "Http method incorrect")
		assert.Equal(t, 1, len(uploaderMock.header), "Length of headers map incorrect")
		assert.Equal(t, []string{"Basic dGVzdENsaWVudElkOnRlc3RDbGllbnRTZWNyZXQ="}, uploaderMock.header[http.CanonicalHeaderKey("authorization")], "Authorizatoin header incorrect")
		assert.Equal(t, "key1=value1", uploaderMock.requestBody, "Request body incorrect")
		assert.Equal(t, []byte(uploaderMock.responseBody), data, "Response body incorrect")
	})

	t.Run("test success against tms", func(t *testing.T) {
		uploaderMock := uploaderMock{responseBody: `{"someKey": "someValue"}`, httpStatusCode: http.StatusOK}
		communicationInstance := CommunicationInstance{uaaUrl: "https://dummy.sap.com", tmsUrl: "https://tms.dummy.sap.com", httpClient: &uploaderMock, logger: logger, isVerbose: false}

		_, err := sendRequest(&communicationInstance, http.MethodGet, "/test", nil, nil, http.StatusOK, false)

		assert.NoError(t, err, "Error occurred, but none expected")
		assert.Equal(t, "https://tms.dummy.sap.com/test", uploaderMock.urlCalled, "Called url incorrect")
	})

	t.Run("test success with trimming url slash in the end", func(t *testing.T) {
		uploaderMock := uploaderMock{responseBody: `{"someKey": "someValue"}`, httpStatusCode: http.StatusOK}
		communicationInstance := CommunicationInstance{uaaUrl: "https://dummy.sap.com/", tmsUrl: "https://tms.dummy.sap.com/", httpClient: &uploaderMock, logger: logger, isVerbose: false}

		_, err := sendRequest(&communicationInstance, http.MethodGet, "/test", nil, nil, http.StatusOK, false)

		assert.NoError(t, err, "Error occurred, but none expected")
		assert.Equal(t, "https://tms.dummy.sap.com/test", uploaderMock.urlCalled, "Called url incorrect")
	})

	t.Run("test error", func(t *testing.T) {
		uploaderMock := uploaderMock{responseBody: `{"someKey": "someValue"}`, httpStatusCode: http.StatusBadRequest}
		communicationInstance := CommunicationInstance{uaaUrl: "https://dummy.sap.com", tmsUrl: "https://tms.dummy.sap.com", httpClient: &uploaderMock, logger: logger, isVerbose: false}

		_, err := sendRequest(&communicationInstance, http.MethodGet, "/test", nil, nil, http.StatusOK, false)

		assert.Error(t, err, "Error expected, but none occurred")
		assert.Equal(t, "https://tms.dummy.sap.com/test", uploaderMock.urlCalled, "Called url incorrect")
		assert.Equal(t, "http error 400", err.Error(), "Error text incorrect")
	})

	t.Run("test error due unexpected positive http status code", func(t *testing.T) {
		uploaderMock := uploaderMock{responseBody: `{"someKey": "someValue"}`, httpStatusCode: http.StatusCreated}
		communicationInstance := CommunicationInstance{uaaUrl: "https://dummy.sap.com", tmsUrl: "https://tms.dummy.sap.com", httpClient: &uploaderMock, logger: logger, isVerbose: false}

		_, err := sendRequest(&communicationInstance, http.MethodPost, "/test", nil, nil, http.StatusOK, false)

		assert.Error(t, err, "Error expected, but none occurred")
		assert.Equal(t, "https://tms.dummy.sap.com/test", uploaderMock.urlCalled, "Called url incorrect")
		assert.Equal(t, "unexpected positive HTTP status code 201, while it was expected 200", err.Error(), "Error text incorrect")
	})

	t.Run("test technical error", func(t *testing.T) {
		uploaderMock := uploaderMock{isTechnicalErrorExpected: true}
		communicationInstance := CommunicationInstance{uaaUrl: "https://dummy.sap.com", tmsUrl: "https://tms.dummy.sap.com", httpClient: &uploaderMock, logger: logger, isVerbose: false}

		data, err := sendRequest(&communicationInstance, http.MethodGet, "/test", nil, nil, http.StatusOK, false)

		assert.Error(t, err, "Error expected, but none occurred")
		assert.Nil(t, data, "Nil result expected, but was not")
		assert.Equal(t, "Provoked technical error", err.Error(), "Error text incorrect")
	})

}

func TestNewCommunicationInstance(t *testing.T) {
	t.Run("test success", func(t *testing.T) {
		uploaderMock := uploaderMock{responseBody: `{"token_type":"bearer","access_token":"testOAuthToken","expires_in":54321}`, httpStatusCode: http.StatusOK}
		communicationInstance, err := NewCommunicationInstance(&uploaderMock, "https://dummy.sap.com", "testClientId", "testClientSecret", false)

		assert.NoError(t, err, "Error occurred, but none expected")
		assert.Equal(t, "https://dummy.sap.com", communicationInstance.uaaUrl, "uaaUrl field of communication instance incorrect")
		assert.Equal(t, "testClientId", communicationInstance.clientId, "clientId field of communication instance incorrect")
		assert.Equal(t, "testClientSecret", communicationInstance.clientSecret, "clientSecret field of communication instance incorrect")
		assert.Equal(t, false, communicationInstance.isVerbose, "isVerbose field of communication instance incorrect")
		assert.Equal(t, "testOAuthToken", uploaderMock.token, "Obtained token incorrect")
	})

	t.Run("test error", func(t *testing.T) {
		uploaderMock := uploaderMock{responseBody: `Bad request provided`, httpStatusCode: http.StatusBadRequest}
		_, err := NewCommunicationInstance(&uploaderMock, "https://dummy.sap.com", "testClientId", "testClientSecret", false)

		assert.Error(t, err, "Error expected, but none occurred")
		assert.Equal(t, "Error fetching OAuth token: http error 400", err.Error(), "Error text incorrect")
	})

}
