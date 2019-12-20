package checkmarx

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type senderMock struct {
	token          string
	httpMethod     string
	httpStatusCode int
	urlCalled      string
	requestBody    io.Reader
	responseBody   string
	header         http.Header
	logger         *logrus.Entry
}

func (sm *senderMock) SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	sm.httpMethod = method
	sm.urlCalled = url
	sm.header = header
	return &http.Response{StatusCode: sm.httpStatusCode, Body: ioutil.NopCloser(strings.NewReader(sm.responseBody))}, nil
}
func (sm *senderMock) UploadFile(url, file, fieldName string, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	sm.httpMethod = http.MethodPost
	sm.urlCalled = url
	sm.header = header
	return &http.Response{StatusCode: sm.httpStatusCode, Body: ioutil.NopCloser(bytes.NewReader([]byte(sm.responseBody)))}, nil
}
func (sm *senderMock) SetOptions(opts piperHttp.ClientOptions) {
	sm.token = opts.Token
}

func TestSendRequest(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"some": "test"}`, httpStatusCode: 200}
		sys := System{serverURL: "https://cx.wdf.sap.corp", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		_, err := sendRequest(&sys, "GET", "/test", nil, nil)

		assert.NoError(t, err, "Error occured but none expected")
		assert.Equal(t, "https://cx.wdf.sap.corp/CxRestAPI/test", myTestClient.urlCalled, "Called url incorrect")
	})

	t.Run("test error", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"some": "test"}`, httpStatusCode: 400}
		sys := System{serverURL: "https://cx.wdf.sap.corp", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)
		_, err := sendRequest(&sys, "GET", "/test", nil, nil)

		assert.Error(t, err, "Error expected but none occured")
		assert.Equal(t, "https://cx.wdf.sap.corp/CxRestAPI/test", myTestClient.urlCalled, "Called url incorrect")
	})
}

func TestGetOAuthToken(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"token_type":"Bearer","access_token":"abcd12345","expires_in":7045634}`, httpStatusCode: 200}
		sys := System{serverURL: "https://cx.wdf.sap.corp", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		token, err := sys.getOAuth2Token()

		assert.NoError(t, err, "Error occured but none expected")
		assert.Equal(t, "https://cx.wdf.sap.corp/CxRestAPI/auth/identity/connect/token", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "Bearer abcd12345", token, "Token incorrect")
	})

	t.Run("test authentication failure", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{}`, httpStatusCode: 400}
		sys := System{serverURL: "https://cx.wdf.sap.corp", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		_, err := sys.getOAuth2Token()

		assert.Error(t, err, "Error expected but none occured")
		assert.Equal(t, "https://cx.wdf.sap.corp/CxRestAPI/auth/identity/connect/token", myTestClient.urlCalled, "Called url incorrect")
	})
}

func TestGetTeams(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `[{"id":"1", "fullName":"Team1"}, {"id":"2", "fullName":"Team2"}, {"id":"3", "fullName":"Team3"}]`, httpStatusCode: 200}
		sys := System{serverURL: "https://cx.wdf.sap.corp", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		teams := sys.GetTeams()

		assert.Equal(t, "https://cx.wdf.sap.corp/CxRestAPI/auth/teams", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, 3, len(teams), "Token incorrect")
		assert.Equal(t, "Team1", teams[0].FullName, "Team name incorrect")
		assert.Equal(t, "Team2", teams[1].FullName, "Team name incorrect")
		assert.Equal(t, "Team3", teams[2].FullName, "Team name incorrect")
	})
}

func TestGetProjects(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `[{"id":"1", "teamId":"1", "name":"Project1"}, {"id":"2", "teamId":"2", "name":"Project2"}]`, httpStatusCode: 200}
		sys := System{serverURL: "https://cx.wdf.sap.corp", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		projects := sys.GetProjects()

		assert.Equal(t, "https://cx.wdf.sap.corp/CxRestAPI/projects", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, 2, len(projects), "Token incorrect")
		assert.Equal(t, "Project1", projects[0].Name, "Team name incorrect")
		assert.Equal(t, "Project2", projects[1].Name, "Team name incorrect")
	})
}

func TestCreateProject(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{httpStatusCode: 200}
		sys := System{serverURL: "https://cx.wdf.sap.corp", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		result := sys.CreateProject("TestProjectCreate", "4711")

		assert.Equal(t, true, result, "CreateProject call not successful")
		assert.Equal(t, "https://cx.wdf.sap.corp/CxRestAPI/projects", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "POST", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, "application/json", myTestClient.header.Get("Content-Type"), "Called url incorrect")
	})
}

func TestUploadProjectSourceCode(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{httpStatusCode: 204}
		sys := System{serverURL: "https://cx.wdf.sap.corp", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		result := sys.UploadProjectSourceCode(10415, "sources.zip")

		assert.Equal(t, true, result, "UploadProjectSourceCode call not successful")
		assert.Equal(t, "https://cx.wdf.sap.corp/CxRestAPI/projects/10415/sourceCode/attachments", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "POST", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, 2, len(myTestClient.header), "HTTP header incorrect")
		assert.Equal(t, "gzip,deflate", myTestClient.header.Get("Accept-Encoding"), "HTTP header incorrect")
		assert.Equal(t, "text/plain", myTestClient.header.Get("Accept"), "HTTP header incorrect")
	})
}
