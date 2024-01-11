//go:build unit
// +build unit

package tms

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
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
	filePath                 string
	fileFieldName            string
	fileContentString        string
	header                   http.Header
	isTechnicalErrorExpected bool
	formFields               map[string]string
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
	return &http.Response{StatusCode: um.httpStatusCode, Body: io.NopCloser(strings.NewReader(um.responseBody))}, httpError
}

func (um *uploaderMock) UploadFile(url, file, fieldName string, header http.Header, cookies []*http.Cookie, uploadType string) (*http.Response, error) {
	um.httpMethod = http.MethodPost
	um.urlCalled = url
	um.header = header
	return &http.Response{StatusCode: um.httpStatusCode, Body: io.NopCloser(bytes.NewReader([]byte(um.responseBody)))}, nil
}

func (um *uploaderMock) UploadRequest(method, url, file, fieldName string, header http.Header, cookies []*http.Cookie, uploadType string) (*http.Response, error) {
	um.httpMethod = http.MethodPost
	um.urlCalled = url
	um.header = header
	return &http.Response{StatusCode: um.httpStatusCode, Body: io.NopCloser(bytes.NewReader([]byte(um.responseBody)))}, nil
}

func (um *uploaderMock) Upload(uploadRequestData piperHttp.UploadRequestData) (*http.Response, error) {
	if um.isTechnicalErrorExpected {
		return nil, errors.New("Provoked technical error")
	}
	um.httpMethod = uploadRequestData.Method
	um.urlCalled = uploadRequestData.URL
	um.header = uploadRequestData.Header
	um.filePath = uploadRequestData.File
	um.fileFieldName = uploadRequestData.FileFieldName
	um.formFields = uploadRequestData.FormFields
	if uploadRequestData.FileContent != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(uploadRequestData.FileContent)
		um.fileContentString = buf.String()
	}
	var httpError error
	if um.httpStatusCode >= 300 {
		httpError = fmt.Errorf("http error %v", um.httpStatusCode)
	}
	return &http.Response{StatusCode: um.httpStatusCode, Body: io.NopCloser(strings.NewReader(um.responseBody))}, httpError
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

		assert.NoError(t, err, "Error occurred, but none expected")
		assert.Equal(t, "https://dummy.sap.com/oauth/token/?grant_type=client_credentials&response_type=token", uploaderMock.urlCalled, "Called url incorrect")
		assert.Equal(t, http.MethodPost, uploaderMock.httpMethod, "Http method incorrect")
		assert.Equal(t, []string{"application/x-www-form-urlencoded"}, uploaderMock.header[http.CanonicalHeaderKey("content-type")], "Content-Type header incorrect")
		assert.Equal(t, []string{"Basic dGVzdENsaWVudElkOnRlc3RDbGllbnRTZWNyZXQ="}, uploaderMock.header[http.CanonicalHeaderKey("authorization")], "Authorizatoin header incorrect")
		assert.Equal(t, "grant_type=password&password=testClientSecret&username=testClientId", uploaderMock.requestBody, "Request body incorrect")
		assert.Equal(t, "bearer testOAuthToken", token, "Obtained token incorrect")
	})

	t.Run("test error", func(t *testing.T) {
		uploaderMock := uploaderMock{responseBody: `Bad request provided`, httpStatusCode: http.StatusBadRequest}
		communicationInstance := CommunicationInstance{uaaUrl: "https://dummy.sap.com", clientId: "testClientId", clientSecret: "testClientSecret", httpClient: &uploaderMock, logger: logger, isVerbose: false}

		_, err := communicationInstance.getOAuthToken()

		assert.Error(t, err, "Error expected, but none occurred")
		assert.Equal(t, "https://dummy.sap.com/oauth/token/?grant_type=client_credentials&response_type=token", uploaderMock.urlCalled, "Called url incorrect")
		assert.Equal(t, http.MethodPost, uploaderMock.httpMethod, "Http method incorrect")
		assert.Equal(t, []string{"application/x-www-form-urlencoded"}, uploaderMock.header[http.CanonicalHeaderKey("content-type")], "Content-Type header incorrect")
		assert.Equal(t, []string{"Basic dGVzdENsaWVudElkOnRlc3RDbGllbnRTZWNyZXQ="}, uploaderMock.header[http.CanonicalHeaderKey("authorization")], "Authorizatoin header incorrect")
		assert.Equal(t, "grant_type=password&password=testClientSecret&username=testClientId", uploaderMock.requestBody, "Request body incorrect")
	})
}

func TestGetNodes(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/tms_test")
	t.Run("test success", func(t *testing.T) {
		getNodesResponse := `{"nodes": [{"id": 1,"name": "TEST_NODE"}]}`
		uploaderMock := uploaderMock{responseBody: getNodesResponse, httpStatusCode: http.StatusOK}
		communicationInstance := CommunicationInstance{tmsUrl: "https://tms.dummy.sap.com", httpClient: &uploaderMock, logger: logger, isVerbose: false}

		nodes, err := communicationInstance.GetNodes()

		assert.NoError(t, err, "Error occurred, but none expected")
		assert.Equal(t, "https://tms.dummy.sap.com/v2/nodes", uploaderMock.urlCalled, "Called url incorrect")
		assert.Equal(t, http.MethodGet, uploaderMock.httpMethod, "Http method incorrect")
		assert.Equal(t, []string{"application/json"}, uploaderMock.header[http.CanonicalHeaderKey("content-type")], "Content-Type header incorrect")
		assert.Equal(t, 1, len(nodes), "Length of nodes list incorrect")
		assert.Equal(t, int64(1), nodes[0].Id, "Id of node at position 0 in the list incorrect")
		assert.Equal(t, "TEST_NODE", nodes[0].Name, "Name of node at position 0 in the list incorrect")
	})

	t.Run("test error", func(t *testing.T) {
		uploaderMock := uploaderMock{responseBody: `Bad request provided`, httpStatusCode: http.StatusBadRequest}
		communicationInstance := CommunicationInstance{tmsUrl: "https://tms.dummy.sap.com", httpClient: &uploaderMock, logger: logger, isVerbose: false}

		_, err := communicationInstance.GetNodes()

		assert.Error(t, err, "Error expected, but none occurred")
		assert.Equal(t, "https://tms.dummy.sap.com/v2/nodes", uploaderMock.urlCalled, "Called url incorrect")
		assert.Equal(t, http.MethodGet, uploaderMock.httpMethod, "Http method incorrect")
		assert.Equal(t, []string{"application/json"}, uploaderMock.header[http.CanonicalHeaderKey("content-type")], "Content-Type header incorrect")

	})

}

func TestGetMtaExtDescriptor(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/tms_test")
	t.Run("test success", func(t *testing.T) {
		id := int64(777)
		mtaExtDescription := "This is a test description"
		mtaId := "test.mta.id"
		mtaExtId := "test.mta.id_ext"
		mtaVersion := "1.0.0"
		lastChangedAt := "2021-11-16T13:06:05.711Z"

		getMtaExtDescriptorResponse := fmt.Sprintf(`{"mtaExtDescriptors": [{"id": %v,"description": "%v","mtaId": "%v","mtaExtId": "%v","mtaVersion": "%v","lastChangedAt": "%v"}]}`, id, mtaExtDescription, mtaId, mtaExtId, mtaVersion, lastChangedAt)
		uploaderMock := uploaderMock{responseBody: getMtaExtDescriptorResponse, httpStatusCode: http.StatusOK}

		communicationInstance := CommunicationInstance{tmsUrl: "https://tms.dummy.sap.com", httpClient: &uploaderMock, logger: logger, isVerbose: false}

		nodeId := int64(111)
		mtaExtDescriptor, err := communicationInstance.GetMtaExtDescriptor(nodeId, mtaId, mtaVersion)

		assert.NoError(t, err, "Error occurred, but none expected")
		assert.Equal(t, fmt.Sprintf("https://tms.dummy.sap.com/v2/nodes/%v/mtaExtDescriptors?mtaId=%v&mtaVersion=%v", nodeId, mtaId, mtaVersion), uploaderMock.urlCalled, "Called url incorrect")
		assert.Equal(t, http.MethodGet, uploaderMock.httpMethod, "Http method incorrect")
		assert.Equal(t, []string{"application/json"}, uploaderMock.header[http.CanonicalHeaderKey("content-type")], "Content-Type header incorrect")
		assert.Equal(t, id, mtaExtDescriptor.Id, "MTA extension descriptor Id field incorrect")
		assert.Equal(t, mtaExtDescription, mtaExtDescriptor.Description, "MTA extension descriptor Description field incorrect")
		assert.Equal(t, mtaId, mtaExtDescriptor.MtaId, "MTA extension descriptor MtaId field incorrect")
		assert.Equal(t, mtaExtId, mtaExtDescriptor.MtaExtId, "MTA extension descriptor MtaExtId field incorrect")
		assert.Equal(t, mtaVersion, mtaExtDescriptor.MtaVersion, "MTA extension descriptor MtaVersion field incorrect")
		assert.Equal(t, lastChangedAt, mtaExtDescriptor.LastChangedAt, "MTA extension descriptor LastChangedAt field incorrect")
	})

	t.Run("test success, no MTA extension descriptor found", func(t *testing.T) {
		getMtaExtDescriptorResponse := `{"mtaExtDescriptors": []}`
		uploaderMock := uploaderMock{responseBody: getMtaExtDescriptorResponse, httpStatusCode: http.StatusOK}
		communicationInstance := CommunicationInstance{tmsUrl: "https://tms.dummy.sap.com", httpClient: &uploaderMock, logger: logger, isVerbose: false}

		nodeId := int64(111)
		mtaId := "test.mta.id"
		mtaVersion := "1.0.1"
		mtaExtDescriptor, err := communicationInstance.GetMtaExtDescriptor(nodeId, mtaId, mtaVersion)

		assert.NoError(t, err, "Error occurred, but none expected")
		assert.Equal(t, fmt.Sprintf("https://tms.dummy.sap.com/v2/nodes/%v/mtaExtDescriptors?mtaId=%v&mtaVersion=%v", nodeId, mtaId, mtaVersion), uploaderMock.urlCalled, "Called url incorrect")
		assert.Equal(t, http.MethodGet, uploaderMock.httpMethod, "Http method incorrect")
		assert.Equal(t, []string{"application/json"}, uploaderMock.header[http.CanonicalHeaderKey("content-type")], "Content-Type header incorrect")
		assert.Equal(t, MtaExtDescriptor{}, mtaExtDescriptor, "Initialized mtaExtDescriptor structure received, but a zero-valued expected")
	})

	t.Run("test error", func(t *testing.T) {
		uploaderMock := uploaderMock{responseBody: `Bad request provided`, httpStatusCode: http.StatusBadRequest}
		communicationInstance := CommunicationInstance{tmsUrl: "https://tms.dummy.sap.com", httpClient: &uploaderMock, logger: logger, isVerbose: false}

		nodeId := int64(111)
		mtaId := "test.mta.id"
		mtaVersion := "1.0.1"
		_, err := communicationInstance.GetMtaExtDescriptor(nodeId, mtaId, mtaVersion)

		assert.Error(t, err, "Error expected, but none occurred")
		assert.Equal(t, fmt.Sprintf("https://tms.dummy.sap.com/v2/nodes/%v/mtaExtDescriptors?mtaId=%v&mtaVersion=%v", nodeId, mtaId, mtaVersion), uploaderMock.urlCalled, "Called url incorrect")
		assert.Equal(t, http.MethodGet, uploaderMock.httpMethod, "Http method incorrect")
		assert.Equal(t, []string{"application/json"}, uploaderMock.header[http.CanonicalHeaderKey("content-type")], "Content-Type header incorrect")
	})

}

func TestUpdateMtaExtDescriptor(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/tms_test")
	t.Run("test success with trimming url slash in the end", func(t *testing.T) {
		idOfMtaExtDescriptor := int64(777)
		mtaExtDescription := "This is an updated description"
		mtaId := "fs-storage"
		mtaExtId := "fs-storage-ext"
		mtaVersion := "1.0.0"
		lastChangedAt := "2021-11-16T13:06:05.711Z"

		updateMtaExtDescriptorResponse := fmt.Sprintf(`{"id": %v,"description": "%v","mtaId": "%v","mtaExtId": "%v","mtaVersion": "%v","lastChangedAt": "%v"}`, idOfMtaExtDescriptor, mtaExtDescription, mtaId, mtaExtId, mtaVersion, lastChangedAt)
		uploaderMock := uploaderMock{responseBody: updateMtaExtDescriptorResponse, httpStatusCode: http.StatusOK}

		// the slash in the end of the url will be trimmed
		communicationInstance := CommunicationInstance{tmsUrl: "https://tms.dummy.sap.com/", httpClient: &uploaderMock, logger: logger, isVerbose: false}

		nodeId := int64(111)
		filePath := "./resources/cf_example.mtaext"
		namedUser := "testUser"
		mtaExtDescriptor, err := communicationInstance.UpdateMtaExtDescriptor(nodeId, idOfMtaExtDescriptor, filePath, mtaVersion, mtaExtDescription, namedUser)

		assert.NoError(t, err, "Error occurred, but none expected")
		assert.Equal(t, fmt.Sprintf("https://tms.dummy.sap.com/v2/nodes/%v/mtaExtDescriptors/%v", nodeId, idOfMtaExtDescriptor), uploaderMock.urlCalled, "Called url incorrect")
		assert.Equal(t, http.MethodPut, uploaderMock.httpMethod, "Http method incorrect")
		assert.Equal(t, []string{namedUser}, uploaderMock.header[http.CanonicalHeaderKey("tms-named-user")], "tms-named-user header incorrect")
		assert.Equal(t, filePath, uploaderMock.filePath, "File path incorrect")
		assert.Equal(t, "file", uploaderMock.fileFieldName, "File field name incorrect")
		assert.Equal(t, map[string]string{"mtaVersion": mtaVersion, "description": mtaExtDescription}, uploaderMock.formFields, "Form field(s) incorrect")

		fileHandle, _ := os.Open(filePath)
		defer fileHandle.Close()
		buf := new(bytes.Buffer)
		buf.ReadFrom(fileHandle)
		fileContentString := buf.String()
		assert.Equal(t, fileContentString, uploaderMock.fileContentString, "File content incorrect")

		assert.Equal(t, idOfMtaExtDescriptor, mtaExtDescriptor.Id, "MTA extension descriptor Id field incorrect")
		assert.Equal(t, mtaExtDescription, mtaExtDescriptor.Description, "MTA extension descriptor Description field incorrect")
		assert.Equal(t, mtaId, mtaExtDescriptor.MtaId, "MTA extension descriptor MtaId field incorrect")
		assert.Equal(t, mtaExtId, mtaExtDescriptor.MtaExtId, "MTA extension descriptor MtaExtId field incorrect")
		assert.Equal(t, mtaVersion, mtaExtDescriptor.MtaVersion, "MTA extension descriptor MtaVersion field incorrect")
		assert.Equal(t, lastChangedAt, mtaExtDescriptor.LastChangedAt, "MTA extension descriptor LastChangedAt field incorrect")
	})

	t.Run("test upload error", func(t *testing.T) {
		uploaderMock := uploaderMock{responseBody: `Bad request provided`, httpStatusCode: http.StatusBadRequest}
		communicationInstance := CommunicationInstance{tmsUrl: "https://tms.dummy.sap.com", httpClient: &uploaderMock, logger: logger, isVerbose: false}

		nodeId := int64(111)
		idOfMtaExtDescriptor := int64(777)
		filePath := "./resources/cf_example.mtaext"
		mtaVersion := "1.0.0"
		mtaExtDescription := "This is an updated description"
		namedUser := "testUser"
		_, err := communicationInstance.UpdateMtaExtDescriptor(nodeId, idOfMtaExtDescriptor, filePath, mtaVersion, mtaExtDescription, namedUser)

		assert.Error(t, err, "Error expected, but none occurred")
		assert.Equal(t, fmt.Sprintf("https://tms.dummy.sap.com/v2/nodes/%v/mtaExtDescriptors/%v", nodeId, idOfMtaExtDescriptor), uploaderMock.urlCalled, "Called url incorrect")
		assert.Equal(t, "http error 400", err.Error(), "Error text incorrect")
	})

	t.Run("test error on opening file", func(t *testing.T) {
		uploaderMock := uploaderMock{responseBody: `Some response`, httpStatusCode: http.StatusOK}
		communicationInstance := CommunicationInstance{tmsUrl: "https://tms.dummy.sap.com", httpClient: &uploaderMock, logger: logger, isVerbose: false}

		nodeId := int64(111)
		idOfMtaExtDescriptor := int64(777)
		filePath := "./resources/not_existing.mtaext"
		mtaVersion := "1.0.0"
		mtaExtDescription := "This is an updated description"
		namedUser := "testUser"
		_, err := communicationInstance.UpdateMtaExtDescriptor(nodeId, idOfMtaExtDescriptor, filePath, mtaVersion, mtaExtDescription, namedUser)

		assert.Error(t, err, "Error expected, but none occurred")
		assert.Contains(t, err.Error(), fmt.Sprintf("unable to locate file %v", filePath), "Error text does not contain expected string")
	})
}

func TestUploadMtaExtDescriptorToNode(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/tms_test")
	t.Run("test success with trimming url slash in the end", func(t *testing.T) {
		idOfMtaExtDescriptor := int64(777)
		mtaExtDescription := "This is a test description"
		mtaId := "fs-storage"
		mtaExtId := "fs-storage-ext"
		mtaVersion := "1.0.0"
		lastChangedAt := "2021-11-16T13:06:05.711Z"

		uploadMtaExtDescriptorResponse := fmt.Sprintf(`{"id": %v,"description": "%v","mtaId": "%v","mtaExtId": "%v","mtaVersion": "%v","lastChangedAt": "%v"}`, idOfMtaExtDescriptor, mtaExtDescription, mtaId, mtaExtId, mtaVersion, lastChangedAt)
		uploaderMock := uploaderMock{responseBody: uploadMtaExtDescriptorResponse, httpStatusCode: http.StatusCreated}

		// the slash in the end of the url will be trimmed
		communicationInstance := CommunicationInstance{tmsUrl: "https://tms.dummy.sap.com/", httpClient: &uploaderMock, logger: logger, isVerbose: false}

		nodeId := int64(111)
		filePath := "./resources/cf_example.mtaext"
		namedUser := "testUser"
		mtaExtDescriptor, err := communicationInstance.UploadMtaExtDescriptorToNode(nodeId, filePath, mtaVersion, mtaExtDescription, namedUser)

		assert.NoError(t, err, "Error occurred, but none expected")
		assert.Equal(t, fmt.Sprintf("https://tms.dummy.sap.com/v2/nodes/%v/mtaExtDescriptors", nodeId), uploaderMock.urlCalled, "Called url incorrect")
		assert.Equal(t, http.MethodPost, uploaderMock.httpMethod, "Http method incorrect")
		assert.Equal(t, []string{namedUser}, uploaderMock.header[http.CanonicalHeaderKey("tms-named-user")], "tms-named-user header incorrect")
		assert.Equal(t, filePath, uploaderMock.filePath, "File path incorrect")
		assert.Equal(t, "file", uploaderMock.fileFieldName, "File field name incorrect")
		assert.Equal(t, map[string]string{"mtaVersion": mtaVersion, "description": mtaExtDescription}, uploaderMock.formFields, "Form field(s) incorrect")

		fileHandle, _ := os.Open(filePath)
		defer fileHandle.Close()
		buf := new(bytes.Buffer)
		buf.ReadFrom(fileHandle)
		fileContentString := buf.String()
		assert.Equal(t, fileContentString, uploaderMock.fileContentString, "File content incorrect")

		assert.Equal(t, idOfMtaExtDescriptor, mtaExtDescriptor.Id, "MTA extension descriptor Id field incorrect")
		assert.Equal(t, mtaExtDescription, mtaExtDescriptor.Description, "MTA extension descriptor Description field incorrect")
		assert.Equal(t, mtaId, mtaExtDescriptor.MtaId, "MTA extension descriptor MtaId field incorrect")
		assert.Equal(t, mtaExtId, mtaExtDescriptor.MtaExtId, "MTA extension descriptor MtaExtId field incorrect")
		assert.Equal(t, mtaVersion, mtaExtDescriptor.MtaVersion, "MTA extension descriptor MtaVersion field incorrect")
		assert.Equal(t, lastChangedAt, mtaExtDescriptor.LastChangedAt, "MTA extension descriptor LastChangedAt field incorrect")
	})

	t.Run("test upload error", func(t *testing.T) {
		uploaderMock := uploaderMock{responseBody: `Bad request provided`, httpStatusCode: http.StatusBadRequest}
		communicationInstance := CommunicationInstance{tmsUrl: "https://tms.dummy.sap.com", httpClient: &uploaderMock, logger: logger, isVerbose: false}

		nodeId := int64(111)
		filePath := "./resources/cf_example.mtaext"
		mtaVersion := "1.0.0"
		mtaExtDescription := "This is a test description"
		namedUser := "testUser"
		_, err := communicationInstance.UploadMtaExtDescriptorToNode(nodeId, filePath, mtaVersion, mtaExtDescription, namedUser)

		assert.Error(t, err, "Error expected, but none occurred")
		assert.Equal(t, fmt.Sprintf("https://tms.dummy.sap.com/v2/nodes/%v/mtaExtDescriptors", nodeId), uploaderMock.urlCalled, "Called url incorrect")
		assert.Equal(t, "http error 400", err.Error(), "Error text incorrect")
	})

	t.Run("test error on opening file", func(t *testing.T) {
		uploaderMock := uploaderMock{responseBody: `Some response`, httpStatusCode: http.StatusOK}
		communicationInstance := CommunicationInstance{tmsUrl: "https://tms.dummy.sap.com", httpClient: &uploaderMock, logger: logger, isVerbose: false}

		nodeId := int64(111)
		filePath := "./resources/not_existing.mtaext"
		mtaVersion := "1.0.0"
		mtaExtDescription := "This is a test description"
		namedUser := "testUser"
		_, err := communicationInstance.UploadMtaExtDescriptorToNode(nodeId, filePath, mtaVersion, mtaExtDescription, namedUser)

		assert.Error(t, err, "Error expected, but none occurred")
		assert.Contains(t, err.Error(), fmt.Sprintf("unable to locate file %v", filePath), "Error text does not contain expected string")
	})
}

func TestUploadFile(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/tms_test")
	t.Run("test success with trimming url slash in the end", func(t *testing.T) {
		fileId := int64(333)
		fileName := "cf_example.mtar"

		uploadFileResponse := fmt.Sprintf(`{"fileId": %v,"fileName": "%v"}`, fileId, fileName)
		uploaderMock := uploaderMock{responseBody: uploadFileResponse, httpStatusCode: http.StatusCreated}

		// the slash in the end of the url will be trimmed
		communicationInstance := CommunicationInstance{tmsUrl: "https://tms.dummy.sap.com/", httpClient: &uploaderMock, logger: logger, isVerbose: false}

		filePath := "./resources/cf_example.mtar"
		namedUser := "testUser"
		fileInfo, err := communicationInstance.UploadFile(filePath, namedUser)

		assert.NoError(t, err, "Error occurred, but none expected")
		assert.Equal(t, "https://tms.dummy.sap.com/v2/files/upload", uploaderMock.urlCalled, "Called url incorrect")
		assert.Equal(t, http.MethodPost, uploaderMock.httpMethod, "Http method incorrect")
		assert.Equal(t, filePath, uploaderMock.filePath, "File path incorrect")
		assert.Equal(t, "file", uploaderMock.fileFieldName, "File field name incorrect")
		assert.Equal(t, map[string]string{"namedUser": namedUser}, uploaderMock.formFields, "Form field incorrect")

		fileHandle, _ := os.Open(filePath)
		defer fileHandle.Close()
		buf := new(bytes.Buffer)
		buf.ReadFrom(fileHandle)
		fileContentString := buf.String()
		assert.Equal(t, fileContentString, uploaderMock.fileContentString, "File content incorrect")

		assert.Equal(t, fileId, fileInfo.Id, "Id field of file info incorrect")
		assert.Equal(t, fileName, fileInfo.Name, "Name field of file info incorrect")
	})

	t.Run("test upload error", func(t *testing.T) {
		uploaderMock := uploaderMock{responseBody: `Bad request provided`, httpStatusCode: http.StatusBadRequest}
		communicationInstance := CommunicationInstance{tmsUrl: "https://tms.dummy.sap.com", httpClient: &uploaderMock, logger: logger, isVerbose: false}

		filePath := "./resources/cf_example.mtar"
		namedUser := "testUser"
		_, err := communicationInstance.UploadFile(filePath, namedUser)

		assert.Error(t, err, "Error expected, but none occurred")
		assert.Equal(t, "https://tms.dummy.sap.com/v2/files/upload", uploaderMock.urlCalled, "Called url incorrect")
		assert.Equal(t, "http error 400", err.Error(), "Error text incorrect")
	})

	t.Run("test error on opening file", func(t *testing.T) {
		uploaderMock := uploaderMock{responseBody: `Some response`, httpStatusCode: http.StatusOK}
		communicationInstance := CommunicationInstance{tmsUrl: "https://tms.dummy.sap.com", httpClient: &uploaderMock, logger: logger, isVerbose: false}

		filePath := "./resources/not_existing.mtar"
		namedUser := "testUser"
		_, err := communicationInstance.UploadFile(filePath, namedUser)

		assert.Error(t, err, "Error expected, but none occurred")
		assert.Contains(t, err.Error(), fmt.Sprintf("unable to locate file %v", filePath), "Error text does not contain expected string")
	})

	t.Run("test error due unexpected positive http status code", func(t *testing.T) {
		fileId := int64(333)
		fileName := "cf_example.mtar"

		uploadFileResponse := fmt.Sprintf(`{"fileId": %v,"fileName": "%v"}`, fileId, fileName)
		uploaderMock := uploaderMock{responseBody: uploadFileResponse, httpStatusCode: http.StatusOK}

		communicationInstance := CommunicationInstance{tmsUrl: "https://tms.dummy.sap.com", httpClient: &uploaderMock, logger: logger, isVerbose: false}

		filePath := "./resources/cf_example.mtar"
		namedUser := "testUser"
		_, err := communicationInstance.UploadFile(filePath, namedUser)

		assert.Error(t, err, "Error expected, but none occurred")
		assert.Equal(t, "https://tms.dummy.sap.com/v2/files/upload", uploaderMock.urlCalled, "Called url incorrect")
		assert.Equal(t, "unexpected positive HTTP status code 200, while it was expected 201", err.Error(), "Error text incorrect")
	})
}

func TestUploadFileToNode(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/tms_test")
	t.Run("test success", func(t *testing.T) {
		transportRequestId := int64(555)
		transportRequestDescription := "This is a test description"
		queueId := int64(123)
		nodeId := int64(456)
		nodeName := "TEST_NODE"
		queueEntryString := fmt.Sprintf(`{"queueId": %v,"nodeId": %v,"nodeName": "%v"}`, queueId, nodeId, nodeName)

		uploadFileToNodeResponse := fmt.Sprintf(`{"transportRequestId": %v,"transportRequestDescription": "%v","queueEntries": [%v]}`, transportRequestId, transportRequestDescription, queueEntryString)
		uploaderMock := uploaderMock{responseBody: uploadFileToNodeResponse, httpStatusCode: http.StatusOK}

		communicationInstance := CommunicationInstance{tmsUrl: "https://tms.dummy.sap.com", httpClient: &uploaderMock, logger: logger, isVerbose: false}

		fileInfo := FileInfo{Id: 111, Name: "test.mtar"}
		namedUser := "testUser"
		nodeUploadResponseEntity, err := communicationInstance.UploadFileToNode(fileInfo, nodeName, transportRequestDescription, namedUser)

		assert.NoError(t, err, "Error occurred, but none expected")
		assert.Equal(t, "https://tms.dummy.sap.com/v2/nodes/upload", uploaderMock.urlCalled, "Called url incorrect")
		assert.Equal(t, http.MethodPost, uploaderMock.httpMethod, "Http method incorrect")
		assert.Equal(t, []string{"application/json"}, uploaderMock.header[http.CanonicalHeaderKey("content-type")], "Content-Type header incorrect")

		entryString := fmt.Sprintf(`{"uri":"%v"}`, strconv.FormatInt(fileInfo.Id, 10))
		assert.Equal(t, fmt.Sprintf(`{"contentType":"MTA","storageType":"FILE","nodeName":"%v","description":"%v","namedUser":"%v","entries":[%v]}`, nodeName, transportRequestDescription, namedUser, entryString), uploaderMock.requestBody, "Request body incorrect")

		assert.Equal(t, transportRequestId, nodeUploadResponseEntity.TransportRequestId, "TransportRequestId field of node upload response incorrect")
		assert.Equal(t, transportRequestDescription, nodeUploadResponseEntity.TransportRequestDescription, "TransportRequestDescription field of node upload response incorrect")
		assert.Equal(t, 1, len(nodeUploadResponseEntity.QueueEntries), "Queue entries amount in node upload response incorrect")
		assert.Equal(t, queueId, nodeUploadResponseEntity.QueueEntries[0].Id, "Queue entry Id field incorrect")
		assert.Equal(t, nodeId, nodeUploadResponseEntity.QueueEntries[0].NodeId, "Queue entry NodeId field incorrect")
		assert.Equal(t, nodeName, nodeUploadResponseEntity.QueueEntries[0].NodeName, "Queue entry NodeName field incorrect")
	})

	t.Run("test error", func(t *testing.T) {
		uploaderMock := uploaderMock{responseBody: `Bad request provided`, httpStatusCode: http.StatusBadRequest}
		communicationInstance := CommunicationInstance{tmsUrl: "https://tms.dummy.sap.com", httpClient: &uploaderMock, logger: logger, isVerbose: false}

		fileInfo := FileInfo{Id: 111, Name: "test.mtar"}
		nodeName := "TEST_NODE"
		transportRequestDescription := "This is a test description"
		namedUser := "testUser"
		_, err := communicationInstance.UploadFileToNode(fileInfo, nodeName, transportRequestDescription, namedUser)

		assert.Error(t, err, "Error expected, but none occurred")
		assert.Equal(t, "https://tms.dummy.sap.com/v2/nodes/upload", uploaderMock.urlCalled, "Called url incorrect")
		assert.Equal(t, "http error 400", err.Error(), "Error text incorrect")
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

		// the slash in the end of the used url will be trimmed
		communicationInstance := CommunicationInstance{uaaUrl: "https://dummy.sap.com/", tmsUrl: "https://tms.dummy.sap.com/", httpClient: &uploaderMock, logger: logger, isVerbose: false}

		_, err := sendRequest(&communicationInstance, http.MethodGet, "/test", nil, nil, http.StatusOK, false)

		assert.NoError(t, err, "Error occurred, but none expected")
		assert.Equal(t, "https://tms.dummy.sap.com/test", uploaderMock.urlCalled, "Called url incorrect")
	})

	t.Run("test success with body values containing spaces", func(t *testing.T) {
		uploaderMock := uploaderMock{responseBody: `{"someKey": "someValue"}`, httpStatusCode: http.StatusOK}
		communicationInstance := CommunicationInstance{uaaUrl: "https://dummy.sap.com", tmsUrl: "https://tms.dummy.sap.com", httpClient: &uploaderMock, logger: logger, isVerbose: false}

		urlFormData := url.Values{
			"key1": {"value with spaces"},
		}
		_, err := sendRequest(&communicationInstance, http.MethodPost, "/test/?param1=value1", strings.NewReader(urlFormData.Encode()), nil, http.StatusOK, true)

		assert.NoError(t, err, "Error occurred, but none expected")
		assert.Equal(t, "https://dummy.sap.com/test/?param1=value1", uploaderMock.urlCalled, "Called url incorrect")
		assert.Equal(t, "key1=value+with+spaces", uploaderMock.requestBody, "Request body incorrect")
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
		communicationInstance, err := NewCommunicationInstance(&uploaderMock, "https://tms.dummy.sap.com", "https://dummy.sap.com", "testClientId", "testClientSecret", false, piperHttp.ClientOptions{})

		assert.NoError(t, err, "Error occurred, but none expected")
		assert.Equal(t, "https://dummy.sap.com", communicationInstance.uaaUrl, "uaaUrl field of communication instance incorrect")
		assert.Equal(t, "testClientId", communicationInstance.clientId, "clientId field of communication instance incorrect")
		assert.Equal(t, "testClientSecret", communicationInstance.clientSecret, "clientSecret field of communication instance incorrect")
		assert.Equal(t, false, communicationInstance.isVerbose, "isVerbose field of communication instance incorrect")
		assert.Equal(t, "bearer testOAuthToken", uploaderMock.token, "Obtained token incorrect")
	})

	t.Run("test error", func(t *testing.T) {
		uploaderMock := uploaderMock{responseBody: `Bad request provided`, httpStatusCode: http.StatusBadRequest}
		_, err := NewCommunicationInstance(&uploaderMock, "https://tms.dummy.sap.com", "https://dummy.sap.com", "testClientId", "testClientSecret", false, piperHttp.ClientOptions{})

		assert.Error(t, err, "Error expected, but none occurred")
		assert.Equal(t, "Error fetching OAuth token: http error 400", err.Error(), "Error text incorrect")
	})

}
