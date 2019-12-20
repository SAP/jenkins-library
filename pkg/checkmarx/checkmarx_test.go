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
	requestBody    string
	responseBody   string
	header         http.Header
	logger         *logrus.Entry
}

func (sm *senderMock) SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	sm.httpMethod = method
	sm.urlCalled = url
	sm.header = header
	if body != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(body)
		sm.requestBody = buf.String()
	}
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
		sys, _ := NewSystem(&myTestClient, "https://cx.wdf.sap.corp", "test", "user")
		myTestClient.SetOptions(opts)

		token, err := sys.getOAuth2Token()

		assert.NoError(t, err, "Error occured but none expected")
		assert.Equal(t, "https://cx.wdf.sap.corp/CxRestAPI/auth/identity/connect/token", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "Bearer abcd12345", token, "Token incorrect")
		assert.Equal(t, "client_id=resource_owner_client&client_secret=014DF517-39D1-4453-B7B3-9930C563627C&grant_type=password&password=user&scope=sast_rest_api&username=test", myTestClient.requestBody, "Request body incorrect")
	})

	t.Run("test authentication failure", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{}`, httpStatusCode: 400}
		sys := System{serverURL: "https://cx.wdf.sap.corp", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		_, err := sys.getOAuth2Token()

		assert.Error(t, err, "Error expected but none occured")
		assert.Equal(t, "https://cx.wdf.sap.corp/CxRestAPI/auth/identity/connect/token", myTestClient.urlCalled, "Called url incorrect")
	})

	t.Run("test new system", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"token_type":"Bearer","access_token":"abcd12345","expires_in":7045634}`, httpStatusCode: 200}
		_, err := NewSystem(&myTestClient, "https://cx.wdf.sap.corp", "test", "user")

		assert.NoError(t, err, "Error occured but none expected")
		assert.Equal(t, "https://cx.wdf.sap.corp/CxRestAPI/auth/identity/connect/token", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "Bearer abcd12345", myTestClient.token, "Token incorrect")
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
		assert.Equal(t, 3, len(teams), "Number of Teams incorrect")
		assert.Equal(t, "Team1", teams[0].FullName, "Team name 1 incorrect")
		assert.Equal(t, "Team2", teams[1].FullName, "Team name 2 incorrect")
		assert.Equal(t, "Team3", teams[2].FullName, "Team name 3 incorrect")

		t.Run("test get teams by name", func(t *testing.T) {
			team2 := sys.GetTeamByName(teams, "Team2")
			assert.Equal(t, "Team2", team2.FullName, "Team name incorrect")
			assert.Equal(t, "2", team2.ID, "Team id incorrect")
		})
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
		assert.Equal(t, 2, len(projects), "Number of Projects incorrect")
		assert.Equal(t, "Project1", projects[0].Name, "Project name 1 incorrect")
		assert.Equal(t, "Project2", projects[1].Name, "Project name 2 incorrect")

		t.Run("test get projects by name", func(t *testing.T) {
			project1 := sys.GetProjectByName(projects, "Project1")
			assert.Equal(t, "Project1", project1.Name, "Project name incorrect")
			assert.Equal(t, "1", project1.TeamID, "Project teamId incorrect")
		})
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
		assert.Equal(t, `{"isPublic":true,"name":"TestProjectCreate","owningTeam":"4711"}`, myTestClient.requestBody, "Request body incorrect")
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

func TestUpdateProjectExcludeSettings(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{httpStatusCode: 204}
		sys := System{serverURL: "https://cx.wdf.sap.corp", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		result := sys.UpdateProjectExcludeSettings(10457, "some,test,a/b/c", "*.go")

		assert.Equal(t, true, result, "UpdateProjectExcludeSettings call not successful")
		assert.Equal(t, "https://cx.wdf.sap.corp/CxRestAPI/projects/10457/sourceCode/excludeSettings", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "PUT", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, 1, len(myTestClient.header), "HTTP header incorrect")
		assert.Equal(t, "application/json", myTestClient.header.Get("Content-Type"), "HTTP header incorrect")
		assert.Equal(t, `{"excludeFilesPattern":"*.go","excludeFoldersPattern":"some,test,a/b/c"}`, myTestClient.requestBody, "Request body incorrect")
	})
}

func TestGetPresets(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `[{"id":"1", "name":"Preset1", "ownerName":"Team1", "link":{"rel":"rel", "uri":"https://1234"}}, {"id":"2", "name":"Preset2", "ownerName":"Team1", "link":{"rel":"re2l", "uri":"https://12347"}}]`, httpStatusCode: 200}
		sys := System{serverURL: "https://cx.wdf.sap.corp", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		presets := sys.GetPresets()

		assert.Equal(t, "https://cx.wdf.sap.corp/CxRestAPI/sast/presets", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, 2, len(presets), "Number of Presets incorrect")
		assert.Equal(t, "Preset1", presets[0].Name, "Preset name incorrect")
		assert.Equal(t, "https://1234", presets[0].Link.URI, "Preset name incorrect")
		assert.Equal(t, "Preset2", presets[1].Name, "Preset name incorrect")

		t.Run("test get preset by name", func(t *testing.T) {
			preset2 := sys.GetPresetByName(presets, "Preset2")
			assert.Equal(t, "Preset2", preset2.Name, "Preset name incorrect")
			assert.Equal(t, "Team1", preset2.OwnerName, "Preset ownerName incorrect")
		})
		t.Run("test fail get preset by name", func(t *testing.T) {
			preset := sys.GetPresetByName(presets, "Preset5")
			assert.Equal(t, "", preset.Name, "Preset name incorrect")
		})
	})
}

func TestUpdateProjectConfiguration(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{httpStatusCode: 204}
		sys := System{serverURL: "https://cx.wdf.sap.corp", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		result := sys.UpdateProjectConfiguration(12, 15, "1")

		assert.Equal(t, true, result, "UpdateProjectConfiguration call not successful")
		assert.Equal(t, "https://cx.wdf.sap.corp/CxRestAPI/sast/scanSettings", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "POST", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, `{"engineConfigurationId":1,"presetId":15,"projectId":12}`, myTestClient.requestBody, "Request body incorrect")
	})
}

func TestScanProject(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"id":1, "link":{"rel":"rel", "uri":"https://scan1234"}}`, httpStatusCode: 200}
		sys := System{serverURL: "https://cx.wdf.sap.corp", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		result, scan := sys.ScanProject(10745)

		assert.Equal(t, true, result, "ScanProject call not successful")
		assert.Equal(t, "https://cx.wdf.sap.corp/CxRestAPI/sast/scans", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "POST", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, 1, scan.ID, "Scan ID incorrect")
		assert.Equal(t, "https://scan1234", scan.Link.URI, "Scan link URI incorrect")
	})
}

func TestGetScanStatus(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"status":{"id":1,"name":"SUCCESS"}}`, httpStatusCode: 200}
		sys := System{serverURL: "https://cx.wdf.sap.corp", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		result := sys.GetScanStatus(10745)

		assert.Equal(t, "https://cx.wdf.sap.corp/CxRestAPI/sast/scans/10745", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "GET", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, "SUCCESS", result, "Request body incorrect")
	})
}

func TestGetResults(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"highSeverity":5, "mediumSeverity":4, "lowSeverity":20, "infoSeverity":10}`, httpStatusCode: 200}
		sys := System{serverURL: "https://cx.wdf.sap.corp", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		result := sys.GetResults(10745)

		assert.Equal(t, "https://cx.wdf.sap.corp/CxRestAPI/sast/scans/10745/resultsStatistics", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "GET", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, 5, result.High, "High findings incorrect")
		assert.Equal(t, 4, result.Medium, "Medium findings incorrect")
		assert.Equal(t, 20, result.Low, "Low findings incorrect")
		assert.Equal(t, 10, result.Info, "Info findings incorrect")
	})
}
