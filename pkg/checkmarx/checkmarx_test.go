//go:build unit
// +build unit

package checkmarx

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
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
	errorExp       bool
}

func (sm *senderMock) SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	if sm.errorExp {
		return &http.Response{}, errors.New("Provoked technical error")
	}
	sm.httpMethod = method
	sm.urlCalled = url
	sm.header = header
	if body != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(body)
		sm.requestBody = buf.String()
	}
	var httpError error
	if sm.httpStatusCode > 399 {
		httpError = fmt.Errorf("http error %v", sm.httpStatusCode)
	}
	return &http.Response{StatusCode: sm.httpStatusCode, Body: io.NopCloser(strings.NewReader(sm.responseBody))}, httpError
}
func (sm *senderMock) UploadFile(url, file, fieldName string, header http.Header, cookies []*http.Cookie, uploadType string) (*http.Response, error) {
	sm.httpMethod = http.MethodPost
	sm.urlCalled = url
	sm.header = header
	return &http.Response{StatusCode: sm.httpStatusCode, Body: io.NopCloser(bytes.NewReader([]byte(sm.responseBody)))}, nil
}
func (sm *senderMock) UploadRequest(method, url, file, fieldName string, header http.Header, cookies []*http.Cookie, uploadType string) (*http.Response, error) {
	sm.httpMethod = http.MethodPost
	sm.urlCalled = url
	sm.header = header
	return &http.Response{StatusCode: sm.httpStatusCode, Body: io.NopCloser(bytes.NewReader([]byte(sm.responseBody)))}, nil
}
func (sm *senderMock) Upload(_ piperHttp.UploadRequestData) (*http.Response, error) {
	return &http.Response{}, fmt.Errorf("not implemented")
}
func (sm *senderMock) SetOptions(opts piperHttp.ClientOptions) {
	sm.token = opts.Token
}

func TestSendRequest(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"some": "test"}`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx.server.com", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		_, err := sendRequest(&sys, "GET", "/test", nil, nil)

		assert.NoError(t, err, "Error occurred but none expected")
		assert.Equal(t, "https://cx.server.com/cxrestapi/test", myTestClient.urlCalled, "Called url incorrect")
	})

	t.Run("test error", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"some": "test"}`, httpStatusCode: 400}
		sys := SystemInstance{serverURL: "https://cx.server.com", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)
		_, err := sendRequest(&sys, "GET", "/test", nil, nil)

		assert.Error(t, err, "Error expected but none occurred")
		assert.Equal(t, "https://cx.server.com/cxrestapi/test", myTestClient.urlCalled, "Called url incorrect")
	})

	t.Run("test technical error", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"some": "test"}`, httpStatusCode: 400}
		sys := SystemInstance{serverURL: "https://cx.server.com", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)
		_, err := sendRequest(&sys, "error", "/test", nil, nil)

		assert.Error(t, err, "Error expected but none occurred")
	})
}

func TestSendRequestInternal(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx_test")
	opts := piperHttp.ClientOptions{}

	t.Run("test accepted error", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"some": "test"}`, httpStatusCode: 404}
		sys := SystemInstance{serverURL: "https://cx.server.com", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)
		_, err := sendRequestInternal(&sys, "GET", "/test", nil, nil, []int{404})

		assert.NoError(t, err, "No error expected but error occurred")
	})
}

func TestGetOAuthToken(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"token_type":"Bearer","access_token":"abcd12345","expires_in":7045634}`, httpStatusCode: 200}
		sys, _ := NewSystemInstance(&myTestClient, "https://cx.server.com", "test", "user")
		myTestClient.SetOptions(opts)

		token, err := sys.getOAuth2Token()

		assert.NoError(t, err, "Error occurred but none expected")
		assert.Equal(t, "https://cx.server.com/cxrestapi/auth/identity/connect/token", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "Bearer abcd12345", token, "Token incorrect")
		assert.Equal(t, "client_id=resource_owner_client&client_secret=014DF517-39D1-4453-B7B3-9930C563627C&grant_type=password&password=user&scope=sast_rest_api&username=test", myTestClient.requestBody, "Request body incorrect")
	})

	t.Run("test authentication failure", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{}`, httpStatusCode: 400}
		sys := SystemInstance{serverURL: "https://cx.server.com", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		_, err := sys.getOAuth2Token()

		assert.Error(t, err, "Error expected but none occurred")
		assert.Equal(t, "https://cx.server.com/cxrestapi/auth/identity/connect/token", myTestClient.urlCalled, "Called url incorrect")
	})

	t.Run("test new system", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"token_type":"Bearer","access_token":"abcd12345","expires_in":7045634}`, httpStatusCode: 200}
		_, err := NewSystemInstance(&myTestClient, "https://cx.server.com", "test", "user")

		assert.NoError(t, err, "Error occurred but none expected")
		assert.Equal(t, "https://cx.server.com/cxrestapi/auth/identity/connect/token", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "Bearer abcd12345", myTestClient.token, "Token incorrect")
	})

	t.Run("test technical error", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{}`, httpStatusCode: 400}
		sys := SystemInstance{serverURL: "https://cx.server.com", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)
		myTestClient.errorExp = true

		_, err := sys.getOAuth2Token()

		assert.Error(t, err, "Error expected but none occurred")
	})
}

func TestGetTeams(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `[{"id":"1", "fullName":"Team1"}, {"id":2, "fullName":"Team2"}, {"id":3, "fullName":"Team3"}, {"id":4, "fullName":"/Team/4"}]`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx.server.com", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		teams := sys.GetTeams()

		assert.Equal(t, "https://cx.server.com/cxrestapi/auth/teams", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, 4, len(teams), "Number of Teams incorrect")
		assert.Equal(t, "Team1", teams[0].FullName, "Team name 1 incorrect")
		assert.Equal(t, "Team2", teams[1].FullName, "Team name 2 incorrect")
		assert.Equal(t, "Team3", teams[2].FullName, "Team name 3 incorrect")
		assert.Equal(t, "/Team/4", teams[3].FullName, "Team name 4 incorrect")

		t.Run("test filter teams by name", func(t *testing.T) {
			team2, _ := sys.FilterTeamByName(teams, "Team2")
			assert.Equal(t, "Team2", team2.FullName, "Team name incorrect")
			assert.Equal(t, json.RawMessage([]byte(strconv.Itoa(2))), team2.ID, "Team id incorrect")
		})

		t.Run("test filter teams by name with backslash/forward slash", func(t *testing.T) {
			team4, _ := sys.FilterTeamByName(teams, "\\Team\\4")
			assert.Equal(t, "/Team/4", team4.FullName, "Team name incorrect")
			assert.Equal(t, json.RawMessage([]byte(strconv.Itoa(4))), team4.ID, "Team id incorrect")
		})

		t.Run("test Filter teams by ID", func(t *testing.T) {
			team1 := sys.FilterTeamByID(teams, json.RawMessage(`"1"`))
			assert.Equal(t, "Team1", team1.FullName, "Team name incorrect")
			assert.Equal(t, json.RawMessage(`"1"`), team1.ID, "Team id incorrect")
		})

		t.Run("test Filter teams by numeric ID", func(t *testing.T) {
			team3 := sys.FilterTeamByID(teams, json.RawMessage(`3`))
			assert.Equal(t, "Team3", team3.FullName, "Team name incorrect")
			assert.Equal(t, json.RawMessage(`3`), team3.ID, "Team id incorrect")
		})

		t.Run("test fail Filter teams by name", func(t *testing.T) {
			team, err := sys.FilterTeamByName(teams, "Team")
			assert.Equal(t, "", team.FullName, "Team name incorrect")
			assert.Contains(t, fmt.Sprint(err), "Failed to find team")
		})
	})

	t.Run("test technical error", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `[{"id":"1", "fullName":"Team1"}, {"id":"2", "fullName":"Team2"}, {"id":"3", "fullName":"Team3"}]`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx.server.com", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)
		myTestClient.errorExp = true

		teams := sys.GetTeams()

		assert.Equal(t, 0, len(teams), "Error expected but none occurred")
	})
}

func TestGetProjects(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `[{"id":"1", "teamId":"1", "name":"Project1"}, {"id":"2", "teamId":"2", "name":"Project2"}]`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx.server.com", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		projects, err := sys.GetProjects()

		assert.NoError(t, err)
		assert.Equal(t, "https://cx.server.com/cxrestapi/projects", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, 2, len(projects), "Number of Projects incorrect")
		assert.Equal(t, "Project1", projects[0].Name, "Project name 1 incorrect")
		assert.Equal(t, "Project2", projects[1].Name, "Project name 2 incorrect")

		t.Run("test Filter projects by name", func(t *testing.T) {
			project1 := sys.FilterProjectByName(projects, "Project1")
			assert.Equal(t, "Project1", project1.Name, "Project name incorrect")
			assert.Equal(t, "1", project1.TeamID, "Project teamId incorrect")
		})

		t.Run("test fail Filter projects by name", func(t *testing.T) {
			project := sys.FilterProjectByName(projects, "Project5")
			assert.Equal(t, "", project.Name, "Project name incorrect")
		})
	})

	t.Run("test technical error", func(t *testing.T) {
		myTestClient := senderMock{httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx.server.com", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)
		myTestClient.errorExp = true

		_, err := sys.GetProjects()

		assert.Contains(t, fmt.Sprint(err), "Provoked technical error")
	})
}

func TestCreateProject(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"id": 16}`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx.server.com", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		result, err := sys.CreateProject("TestProjectCreate", "4711")

		assert.NoError(t, err, "CreateProject call not successful")
		assert.Equal(t, 16, result.ID, "Wrong project ID")
		assert.Equal(t, "https://cx.server.com/cxrestapi/projects", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "POST", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, "application/json", myTestClient.header.Get("Content-Type"), "Called url incorrect")
		assert.Equal(t, `{"isPublic":true,"name":"TestProjectCreate","owningTeam":"4711"}`, myTestClient.requestBody, "Request body incorrect")
	})

	t.Run("test technical error", func(t *testing.T) {
		myTestClient := senderMock{httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx.server.com", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)
		myTestClient.errorExp = true

		_, err := sys.CreateProject("Test", "13")

		assert.Contains(t, fmt.Sprint(err), "", "expected a different error")
	})
}

func TestUploadProjectSourceCode(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{httpStatusCode: 204}
		sys := SystemInstance{serverURL: "https://cx.server.com", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		err := sys.UploadProjectSourceCode(10415, "sources.zip")

		assert.NoError(t, err, "UploadProjectSourceCode call not successful")
		assert.Equal(t, "https://cx.server.com/cxrestapi/projects/10415/sourceCode/attachments", myTestClient.urlCalled, "Called url incorrect")
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
		sys := SystemInstance{serverURL: "https://cx.server.com", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		err := sys.UpdateProjectExcludeSettings(10457, "some,test,a/b/c", "*.go")

		assert.NoError(t, err, "UpdateProjectExcludeSettings call not successful")
		assert.Equal(t, "https://cx.server.com/cxrestapi/projects/10457/sourceCode/excludeSettings", myTestClient.urlCalled, "Called url incorrect")
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
		myTestClient := senderMock{responseBody: `[{"id":1, "name":"Preset1", "ownerName":"Team1", "link":{"rel":"rel", "uri":"https://1234"}}, {"id":2, "name":"Preset2", "ownerName":"Team1", "link":{"rel":"re2l", "uri":"https://12347"}}]`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx.server.com", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		presets := sys.GetPresets()

		assert.Equal(t, "https://cx.server.com/cxrestapi/sast/presets", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, 2, len(presets), "Number of Presets incorrect")
		assert.Equal(t, "Preset1", presets[0].Name, "Preset name incorrect")
		assert.Equal(t, "https://1234", presets[0].Link.URI, "Preset name incorrect")
		assert.Equal(t, "Preset2", presets[1].Name, "Preset name incorrect")

		t.Run("test Filter preset by name", func(t *testing.T) {
			preset2 := sys.FilterPresetByName(presets, "Preset2")
			assert.Equal(t, "Preset2", preset2.Name, "Preset name incorrect")
			assert.Equal(t, "Team1", preset2.OwnerName, "Preset ownerName incorrect")
		})
		t.Run("test fail Filter preset by name", func(t *testing.T) {
			preset := sys.FilterPresetByName(presets, "Preset5")
			assert.Equal(t, "", preset.Name, "Preset name incorrect")
		})
		t.Run("test Filter preset by ID", func(t *testing.T) {
			preset2 := sys.FilterPresetByID(presets, 2)
			assert.Equal(t, "Preset2", preset2.Name, "Preset ID incorrect")
			assert.Equal(t, "Team1", preset2.OwnerName, "Preset ownerName incorrect")
		})
		t.Run("test fail Filter preset by ID", func(t *testing.T) {
			preset := sys.FilterPresetByID(presets, 15)
			assert.Equal(t, "", preset.Name, "Preset ID incorrect")
		})
	})
}

func TestUpdateProjectConfiguration(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{httpStatusCode: 204}
		sys := SystemInstance{serverURL: "https://cx.server.com", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		err := sys.UpdateProjectConfiguration(12, 15, "1")

		assert.NoError(t, err, "UpdateProjectConfiguration call not successful")
		assert.Equal(t, "https://cx.server.com/cxrestapi/sast/scanSettings", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "POST", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, `{"engineConfigurationId":1,"presetId":15,"projectId":12}`, myTestClient.requestBody, "Request body incorrect")
	})
}

func TestScanProject(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"id":1, "link":{"rel":"rel", "uri":"https://scan1234"}}`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx.server.com", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		scan, err := sys.ScanProject(10745, false, false, false)

		assert.NoError(t, err, "ScanProject call not successful")
		assert.Equal(t, "https://cx.server.com/cxrestapi/sast/scans", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "POST", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, 1, scan.ID, "Scan ID incorrect")
		assert.Equal(t, "https://scan1234", scan.Link.URI, "Scan link URI incorrect")
	})
}

func TestGetScans(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `[
			{
			  "id": 1000000,
			  "project": {
				"id": 1,
				"name": "Project 1 (CxTechDocs)"
			  },
			  "status": {
				"id": 7,
				"name": "Finished"
			  },
			  "isIncremental": false
			},
			{
				"id": 1000001,
				"project": {
				  "id": 2,
				  "name": "Project 2 (CxTechDocs)"
				},
				"status": {
				  "id": 7,
				  "name": "Finished"
				},
				"isIncremental": true
			  }
		  ]`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx.server.com", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		scans, err := sys.GetScans(10745)

		assert.NoError(t, err, "ScanProject call not successful")
		assert.Equal(t, "https://cx.server.com/cxrestapi/sast/scans?last=20&projectId=10745", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "GET", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, 2, len(scans), "Incorrect number of scans")
		assert.Equal(t, true, scans[1].IsIncremental, "Scan link URI incorrect")
	})
}

func TestGetScanStatusAndDetail(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"status":{"id":1,"name":"SUCCESS", "details":{"stage": "1 of 15", "step": "One"}}}`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx.server.com", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		result, detail := sys.GetScanStatusAndDetail(10745)

		assert.Equal(t, "https://cx.server.com/cxrestapi/sast/scans/10745", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "GET", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, "SUCCESS", result, "Request body incorrect")
		assert.Equal(t, "One", detail.Step, "Detail step incorrect")
		assert.Equal(t, "1 of 15", detail.Stage, "Detail stage incorrect")
	})
}

func TestGetResults(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"highSeverity":5, "mediumSeverity":4, "lowSeverity":20, "infoSeverity":10}`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx.server.com", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		result := sys.GetResults(10745)

		assert.Equal(t, "https://cx.server.com/cxrestapi/sast/scans/10745/resultsStatistics", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "GET", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, 5, result.High, "High findings incorrect")
		assert.Equal(t, 4, result.Medium, "Medium findings incorrect")
		assert.Equal(t, 20, result.Low, "Low findings incorrect")
		assert.Equal(t, 10, result.Info, "Info findings incorrect")
	})
}

func TestRequestNewReport(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{
			"reportId": 6,
			"links": {
			  "report": {
				"rel": "content",
				"uri": "/reports/sastScan/6"
			  },
			  "status": {
				"rel": "status",
				"uri": "/reports/sastScan/6/status"
			  }
			}
		  }`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx.server.com", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		result, err := sys.RequestNewReport(10745, "XML")

		assert.NoError(t, err, "Result status incorrect")
		assert.Equal(t, "https://cx.server.com/cxrestapi/reports/sastScan", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, `{"comment":"Scan report triggered by Piper","reportType":"XML","scanId":10745}`, myTestClient.requestBody, "Request body incorrect")
		assert.Equal(t, "POST", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, 6, result.ReportID, "Report ID incorrect")
	})
}

func TestGetReportStatus(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{
			"link": {
			  "rel": "content",
			  "uri": "/reports/sastScan/51"
			},
			"contentType": "application/xml",
			"status": {
			  "id": 2,
			  "value": "Created"
			}
		  }`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx.server.com", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		result, err := sys.GetReportStatus(6)

		assert.NoError(t, err, "error occured but none expected")
		assert.Equal(t, "https://cx.server.com/cxrestapi/reports/sastScan/6/status", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "GET", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, 2, result.Status.ID, "Status ID incorrect")
		assert.Equal(t, "Created", result.Status.Value, "Status incorrect")
	})
}

func TestDownloadReport(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: "abc", httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx.server.com", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		result, err := sys.DownloadReport(6)
		assert.NoError(t, err, "DownloadReport returned unexpected error")
		assert.Equal(t, "https://cx.server.com/cxrestapi/reports/sastScan/6", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "GET", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, []byte("abc"), result, "Result incorrect")
	})
}

func TestGetProjectByID(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"id": 209, "teamID": "Test", "name":"Project1_PR-18"}`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx.server.com", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		result, err := sys.GetProjectByID(815)
		assert.NoError(t, err, "GetProjectByID returned unexpected error")
		assert.Equal(t, "https://cx.server.com/cxrestapi/projects/815", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "GET", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, 209, result.ID, "Result incorrect")
	})
}

func TestGetProjectByName(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `[{"id": 209, "teamID": "Test", "name":"Project1_PR-18"}]`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx.server.com", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		result, err := sys.GetProjectsByNameAndTeam("Project1_PR-18", "Test")
		assert.NoError(t, err, "error occured but none expected")
		assert.Equal(t, 1, len(result), "GetProjectByName returned unexpected error")
		assert.Equal(t, "https://cx.server.com/cxrestapi/projects?projectName=Project1_PR-18&teamId=Test", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "GET", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, "Project1_PR-18", result[0].Name, "Result incorrect")
	})
}

func TestGetShortDescription(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"shortDescription":"This is a dummy short description."}`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx.server.com", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		shortDescription, err := sys.GetShortDescription(11037, 1)

		assert.NoError(t, err)
		assert.Equal(t, "https://cx.server.com/cxrestapi/sast/scans/11037/results/1/shortDescription", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "GET", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, "This is a dummy short description.", shortDescription.Text, "Description incorrect")
	})
}
