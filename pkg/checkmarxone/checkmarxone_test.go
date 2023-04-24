package checkmarxOne

import (
	"bytes"
	"errors"
	"fmt"
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
	return &http.Response{StatusCode: sm.httpStatusCode, Body: ioutil.NopCloser(strings.NewReader(sm.responseBody))}, httpError
}
func (sm *senderMock) UploadFile(url, file, fieldName string, header http.Header, cookies []*http.Cookie, uploadType string) (*http.Response, error) {
	sm.httpMethod = http.MethodPost
	sm.urlCalled = url
	sm.header = header
	return &http.Response{StatusCode: sm.httpStatusCode, Body: ioutil.NopCloser(bytes.NewReader([]byte(sm.responseBody)))}, nil
}
func (sm *senderMock) UploadRequest(method, url, file, fieldName string, header http.Header, cookies []*http.Cookie, uploadType string) (*http.Response, error) {
	sm.httpMethod = http.MethodPost
	sm.urlCalled = url
	sm.header = header
	return &http.Response{StatusCode: sm.httpStatusCode, Body: ioutil.NopCloser(bytes.NewReader([]byte(sm.responseBody)))}, nil
}
func (sm *senderMock) Upload(_ piperHttp.UploadRequestData) (*http.Response, error) {
	return &http.Response{}, fmt.Errorf("not implemented")
}
func (sm *senderMock) SetOptions(opts piperHttp.ClientOptions) {
	sm.token = opts.Token
}

func TestSendRequest(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxone_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"some": "test"}`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		_, err := sendRequest(&sys, "GET", "/test", nil, nil, []int{})

		assert.NoError(t, err, "Error occurred but none expected")
		assert.Equal(t, "https://cx1.server.com/api/test", myTestClient.urlCalled, "Called url incorrect")
	})

	t.Run("test error", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"some": "test"}`, httpStatusCode: 400}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)
		_, err := sendRequest(&sys, "GET", "/test", nil, nil, []int{})

		assert.Error(t, err, "Error expected but none occurred")
		assert.Equal(t, "https://cx1.server.com/api/test", myTestClient.urlCalled, "Called url incorrect")
	})

	t.Run("test technical error", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"some": "test"}`, httpStatusCode: 400}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)
		_, err := sendRequest(&sys, "error", "/test", nil, nil, []int{})

		assert.Error(t, err, "Error expected but none occurred")
	})
}

func TestSendRequestInternal(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxone_test")
	opts := piperHttp.ClientOptions{}

	t.Run("test accepted error", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"some": "test"}`, httpStatusCode: 404}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)
		_, err := sendRequestInternal(&sys, "GET", "/test", nil, nil, []int{404})

		assert.NoError(t, err, "No error expected but error occurred")
	})
}

func TestGetOAuthToken(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxone_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"token_type":"Bearer","access_token":"abcd12345","expires_in":7045634}`, httpStatusCode: 200}
		sys, _ := NewSystemInstance(&myTestClient, "https://cx1.server.com", "https://cx1iam.server.com", "tenant", "", "client", "secret")
		myTestClient.SetOptions(opts)

		token, err := sys.getOAuth2Token()

		assert.NoError(t, err, "Error occurred but none expected")
		assert.Equal(t, "https://cx1iam.server.com/auth/realms/tenant/protocol/openid-connect/token", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "Bearer abcd12345", token, "Token incorrect")
		assert.Equal(t, "client_id=client&client_secret=secret&grant_type=client_credentials", myTestClient.requestBody, "Request body incorrect")
	})

	t.Run("test authentication failure", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{}`, httpStatusCode: 400}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		_, err := sys.getOAuth2Token()

		assert.Error(t, err, "Error expected but none occurred")
		assert.Equal(t, "https://cx1iam.server.com/auth/realms/tenant/protocol/openid-connect/token", myTestClient.urlCalled, "Called url incorrect")
	})

	t.Run("test new system", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"token_type":"Bearer","access_token":"abcd12345","expires_in":7045634}`, httpStatusCode: 200}
		_, err := NewSystemInstance(&myTestClient, "https://cx1.server.com", "https://cx1iam.server.com", "tenant", "", "client", "secret")

		assert.NoError(t, err, "Error occurred but none expected")
		assert.Equal(t, "https://cx1iam.server.com/auth/realms/tenant/protocol/openid-connect/token", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "Bearer abcd12345", myTestClient.token, "Token incorrect")
	})

	t.Run("test technical error", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{}`, httpStatusCode: 400}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)
		myTestClient.errorExp = true

		_, err := sys.getOAuth2Token()

		assert.Error(t, err, "Error expected but none occurred")
	})
}

func TestGetGroups(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxone_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `[{"id":"be82031b-a75c-4fc0-894b-fff4deab2854","name":"Group1","path":"/Group1","subGroups":[]},{"id":"b368988c-b124-4151-b507-c8fcad501165","name":"Group2","path":"/Group2","subGroups":[]}]`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		groups, err := sys.GetGroups()
		assert.NoError(t, err, "Error occurred but none expected")

		assert.Equal(t, "https://cx1iam.server.com/auth/realms/tenant/pip/groups", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, 2, len(groups), "Number of Groups incorrect")
		assert.Equal(t, "Group1", groups[0].Name, "Group name 1 incorrect")
		assert.Equal(t, "Group2", groups[1].Name, "Group name 2 incorrect")

		t.Run("test filter groups by name", func(t *testing.T) {
			group2, _ := sys.GetGroupByName("Group2")
			assert.Equal(t, "Group2", group2.Name, "Group name incorrect")
			assert.Equal(t, "b368988c-b124-4151-b507-c8fcad501165", group2.GroupID, "Group id incorrect")
		})

		t.Run("test Filter groups by ID", func(t *testing.T) {
			group1, _ := sys.GetGroupByID("be82031b-a75c-4fc0-894b-fff4deab2854")
			assert.Equal(t, "Group1", group1.Name, "Group name incorrect")
			assert.Equal(t, "be82031b-a75c-4fc0-894b-fff4deab2854", group1.GroupID, "Group id incorrect")
		})

		t.Run("test fail Filter groups by name", func(t *testing.T) {
			group, err := sys.GetGroupByName("Group")
			assert.Equal(t, "", group.Name, "Group name incorrect")
			assert.Contains(t, fmt.Sprint(err), "No group matching")
		})
	})

	t.Run("test technical error", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `[{"id":"1", "fullName":"Group1"}, {"id":"2", "fullName":"Group2"}, {"id":"3", "fullName":"Group3"}]`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)
		myTestClient.errorExp = true

		groups, _ := sys.GetGroups()

		assert.Equal(t, 0, len(groups), "Error expected but none occurred")
	})
}

func TestGetProjects(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxone_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"totalCount":2,"filteredTotalCount":2,"projects":[{"id":"872b63f8-7a78-434e-9bc8-9dc81449e6a8","name":"Project1","createdAt":"2022-12-15T06:24:07.148202Z","updatedAt":"2022-12-15T06:24:07.148202Z","groups":[],"tags":{},"repoUrl":"","mainBranch":"","criticality":3,"privatePackage":false},{"id":"872b63f8-7a78-434e-9bc8-9dc81449e6a9","name":"Project2","createdAt":"2022-12-16T06:24:07.148202Z","updatedAt":"2022-12-16T06:24:07.148202Z","groups":[],"tags":{},"repoUrl":"","mainBranch":"","criticality":3,"privatePackage":false}]}`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		projects, err := sys.GetProjects()

		assert.NoError(t, err)
		assert.Equal(t, "https://cx1.server.com/api/projects/", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, 2, len(projects), "Number of Projects incorrect")
		assert.Equal(t, "Project1", projects[0].Name, "Project name 1 incorrect")
		assert.Equal(t, "Project2", projects[1].Name, "Project name 2 incorrect")

		t.Run("test Filter projects by name", func(t *testing.T) {
			projects, _ := sys.GetProjectsByName("Project")
			assert.Equal(t, "Project1", projects[0].Name, "Project name incorrect")
			assert.Equal(t, "872b63f8-7a78-434e-9bc8-9dc81449e6a8", projects[0].ProjectID, "Project groupId incorrect")
		})
	})

	t.Run("test technical error", func(t *testing.T) {
		myTestClient := senderMock{httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)
		myTestClient.errorExp = true

		_, err := sys.GetProjects()

		assert.Contains(t, fmt.Sprint(err), "error")
	})
}

func TestCreateProject(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxone_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"id":"06b1e9d5-5773-4bc9-9f15-9b84f2b967b4","name":"TestProjectCreate","createdAt":"2023-04-06T13:31:32.915146717Z","updatedAt":"2023-04-06T13:31:32.915146717Z","groups":["e2f958a6-fcef-4b18-aa27-ae2e24930ab2"],"tags":{},"repoUrl":"","mainBranch":"","origin":"Postmans","criticality":3,"privatePackage":false}`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		result, err := sys.CreateProject("TestProjectCreate", []string{"e2f958a6-fcef-4b18-aa27-ae2e24930ab2"})

		assert.NoError(t, err, "CreateProject call not successful")
		assert.Equal(t, "06b1e9d5-5773-4bc9-9f15-9b84f2b967b4", result.ProjectID, "Wrong project ID")
		assert.Equal(t, "https://cx1.server.com/api/projects", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "POST", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, "application/json", myTestClient.header.Get("Content-Type"), "Called url incorrect")
		assert.Equal(t, `{"criticality":3,"groups":["e2f958a6-fcef-4b18-aa27-ae2e24930ab2"],"name":"TestProjectCreate","origin":"GolangScript"}`, myTestClient.requestBody, "Request body incorrect")
	})

	t.Run("test technical error", func(t *testing.T) {
		myTestClient := senderMock{httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)
		myTestClient.errorExp = true

		_, err := sys.CreateProject("Test", []string{"13"})

		assert.Contains(t, fmt.Sprint(err), "", "expected a different error")
	})
}

/*
func TestUploadProjectSourceCode(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxone_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test upload zip success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{ "url": "https://cx1.server.com/storage/location" }`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		uploadurl, err := sys.UploadProjectSourceCode("123", "sources.zip")

		assert.NoError(t, err, "UploadProjectSourceCode call not successful")
		assert.Equal(t, "https://cx1.server.com/api/uploads", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "PUT", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, "https://cx1.server.com/storage/location", uploadurl)
		assert.Equal(t, 2, len(myTestClient.header), "HTTP header incorrect")
	})
}*/

/*

func TestUpdateProjectExcludeSettings(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxone_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{httpStatusCode: 204}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		err := sys.UpdateProjectExcludeSettings(10457, "some,test,a/b/c", "*.go")

		assert.NoError(t, err, "UpdateProjectExcludeSettings call not successful")
		assert.Equal(t, "https://cx1.server.com/cxrestapi/projects/10457/sourceCode/excludeSettings", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "PUT", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, 1, len(myTestClient.header), "HTTP header incorrect")
		assert.Equal(t, "application/json", myTestClient.header.Get("Content-Type"), "HTTP header incorrect")
		assert.Equal(t, `{"excludeFilesPattern":"*.go","excludeFoldersPattern":"some,test,a/b/c"}`, myTestClient.requestBody, "Request body incorrect")
	})
}
*/
/*
func TestGetPresets(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxone_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"totalCount":3,"presets":[{"id":100028,"name":"ASA Premium"},{"id":1,"name":"All"},{"id":9,"name":"Android"}]}`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		presets, _ := sys.GetPresets()

		assert.Equal(t, "https://cx1.server.com/api/presets", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, 3, len(presets), "Number of Presets incorrect")
		assert.Equal(t, "ASA Premium", presets[0].Name, "Preset name incorrect")
		assert.Equal(t, 9, presets[2].PresetID, "Preset ID incorrect")
	})
} */

/*
func TestUpdateProjectConfiguration(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxone_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{httpStatusCode: 204}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		err := sys.UpdateProjectConfiguration(12, 15, "1")

		assert.NoError(t, err, "UpdateProjectConfiguration call not successful")
		assert.Equal(t, "https://cx1.server.com/cxrestapi/sast/scanSettings", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "POST", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, `{"engineConfigurationId":1,"presetId":15,"projectId":12}`, myTestClient.requestBody, "Request body incorrect")
	})
}*/

func TestUpdateProjectConfiguration(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxone_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test update preset", func(t *testing.T) {
		myTestClient := senderMock{httpStatusCode: 204}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		err := sys.SetProjectPreset("123", "All", true)

		assert.NoError(t, err, "SetProjectPreset call not successful")
		assert.Equal(t, "https://cx1.server.com/api/configuration/project?project-id=123", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "PATCH", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, `[{"key":"scan.config.sast.presetName","name":"","category":"","originLevel":"","value":"All","valuetype":"","valuetypeparams":"","allowOverride":true}]`, myTestClient.requestBody, "Request body incorrect")
	})
}

func TestScanProject(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxone_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"id":1, "link":{"rel":"rel", "uri":"https://scan1234"}}`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		scan, err := sys.ScanProject(10745, false, false, false)

		assert.NoError(t, err, "ScanProject call not successful")
		assert.Equal(t, "https://cx1.server.com/cxrestapi/sast/scans", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "POST", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, 1, scan.ID, "Scan ID incorrect")
		assert.Equal(t, "https://scan1234", scan.Link.URI, "Scan link URI incorrect")
	})
}

/*
func TestGetScans(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxone_test")
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
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		scans, err := sys.GetScans(10745)

		assert.NoError(t, err, "ScanProject call not successful")
		assert.Equal(t, "https://cx1.server.com/cxrestapi/sast/scans?last=20&projectId=10745", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "GET", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, 2, len(scans), "Incorrect number of scans")
		assert.Equal(t, true, scans[1].IsIncremental, "Scan link URI incorrect")
	})
}

func TestGetScanStatusAndDetail(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxone_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"status":{"id":1,"name":"SUCCESS", "details":{"stage": "1 of 15", "step": "One"}}}`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		result, detail := sys.GetScanStatusAndDetail(10745)

		assert.Equal(t, "https://cx1.server.com/cxrestapi/sast/scans/10745", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "GET", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, "SUCCESS", result, "Request body incorrect")
		assert.Equal(t, "One", detail.Step, "Detail step incorrect")
		assert.Equal(t, "1 of 15", detail.Stage, "Detail stage incorrect")
	})
}

func TestGetResults(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxone_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"highSeverity":5, "mediumSeverity":4, "lowSeverity":20, "infoSeverity":10}`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		result := sys.GetResults(10745)

		assert.Equal(t, "https://cx1.server.com/cxrestapi/sast/scans/10745/resultsStatistics", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "GET", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, 5, result.High, "High findings incorrect")
		assert.Equal(t, 4, result.Medium, "Medium findings incorrect")
		assert.Equal(t, 20, result.Low, "Low findings incorrect")
		assert.Equal(t, 10, result.Info, "Info findings incorrect")
	})
}

func TestRequestNewReport(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxone_test")
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
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		result, err := sys.RequestNewReport(10745, "XML")

		assert.NoError(t, err, "Result status incorrect")
		assert.Equal(t, "https://cx1.server.com/cxrestapi/reports/sastScan", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, `{"comment":"Scan report triggered by Piper","reportType":"XML","scanId":10745}`, myTestClient.requestBody, "Request body incorrect")
		assert.Equal(t, "POST", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, 6, result.ReportID, "Report ID incorrect")
	})
}

func TestGetReportStatus(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxone_test")
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
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		result, err := sys.GetReportStatus(6)

		assert.NoError(t, err, "error occured but none expected")
		assert.Equal(t, "https://cx1.server.com/cxrestapi/reports/sastScan/6/status", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "GET", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, 2, result.Status.ID, "Status ID incorrect")
		assert.Equal(t, "Created", result.Status.Value, "Status incorrect")
	})
}

func TestDownloadReport(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxone_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: "abc", httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		result, err := sys.DownloadReport(6)
		assert.NoError(t, err, "DownloadReport returned unexpected error")
		assert.Equal(t, "https://cx1.server.com/cxrestapi/reports/sastScan/6", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "GET", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, []byte("abc"), result, "Result incorrect")
	})
}

func TestCreateBranch(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxone_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"id": 13, "link": {}}`, httpStatusCode: 201}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		result := sys.CreateBranch(6, "PR-17")
		assert.Equal(t, "https://cx1.server.com/cxrestapi/projects/6/branch", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "POST", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, `{"name":"PR-17"}`, myTestClient.requestBody, "Request body incorrect")
		assert.Equal(t, 13, result, "result incorrect")
	})
}

func TestGetProjectByID(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxone_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"id": 209, "groupID": "Test", "name":"Project1_PR-18"}`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		result, err := sys.GetProjectByID(815)
		assert.NoError(t, err, "GetProjectByID returned unexpected error")
		assert.Equal(t, "https://cx1.server.com/cxrestapi/projects/815", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "GET", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, 209, result.ID, "Result incorrect")
	})
}

func TestGetProjectByName(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxone_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `[{"id": 209, "groupID": "Test", "name":"Project1_PR-18"}]`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		result, err := sys.GetProjectsByNameAndGroup("Project1_PR-18", "Test")
		assert.NoError(t, err, "error occured but none expected")
		assert.Equal(t, 1, len(result), "GetProjectByName returned unexpected error")
		assert.Equal(t, "https://cx1.server.com/cxrestapi/projects?projectName=Project1_PR-18&groupId=Test", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "GET", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, "Project1_PR-18", result[0].Name, "Result incorrect")
	})
}

func TestGetShortDescription(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxone_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"shortDescription":"This is a dummy short description."}`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		shortDescription, err := sys.GetShortDescription(11037, 1)

		assert.NoError(t, err)
		assert.Equal(t, "https://cx1.server.com/cxrestapi/sast/scans/11037/results/1/shortDescription", myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "GET", myTestClient.httpMethod, "HTTP method incorrect")
		assert.Equal(t, "This is a dummy short description.", shortDescription.Text, "Description incorrect")
	})
}
*/
