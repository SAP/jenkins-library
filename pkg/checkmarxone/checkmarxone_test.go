package checkmarxOne

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxOne_test")
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
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxOne_test")
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
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxOne_test")
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
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxOne_test")
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

func TestGetScanMetadata(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxOne_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"scanId":"03d66397-36df-40b5-8976-f38bcce695a7","projectId":"eac4dc3b-4bbf-4d04-87e5-3b3cedae38fb","loc":158,"fileCount":39,"isIncremental":false,"isIncrementalCanceled":false,"queryPreset":"Checkmarx Default"}`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		scanmeta, err := sys.GetScanMetadata("03d66397-36df-40b5-8976-f38bcce695a7")
		assert.NoError(t, err, "Error occurred but none expected")

		assert.Equal(t, "03d66397-36df-40b5-8976-f38bcce695a7", scanmeta.ScanID, "ScanID is incorrect")
		assert.Equal(t, "eac4dc3b-4bbf-4d04-87e5-3b3cedae38fb", scanmeta.ProjectID, "ProjectID is incorrect")
		assert.Equal(t, 158, scanmeta.LOC, "LOC is incorrect")
		assert.Equal(t, 39, scanmeta.FileCount, "FileCount is incorrect")
		assert.Equal(t, false, scanmeta.IsIncremental, "IsIncremental is incorrect")
		assert.Equal(t, false, scanmeta.IsIncrementalCanceled, "IsIncrementalCanceled is incorrect")
		assert.Equal(t, "Checkmarx Default", scanmeta.PresetName, "PresetName is incorrect")
	})

	t.Run("test technical error", func(t *testing.T) {
		myTestClient := senderMock{httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)
		myTestClient.errorExp = true

		_, err := sys.GetScanMetadata("03d66397-36df-40b5-8976-f38bcce695a7")
		assert.Contains(t, fmt.Sprint(err), "Provoked technical error")
	})
}

func TestGetScan(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxOne_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"id":"7343f9f5-7633-40d5-b000-0a7a3c2c432e","status":"Completed","statusDetails":[{"name":"general","status":"Completed","details":""},{"name":"sast","status":"Completed","details":"","loc":2148}],"branch":"master","createdAt":"2023-03-31T08:35:56.412514Z","updatedAt":"2023-03-31T08:36:53.526569Z","projectId":"e7a7704c-4bfe-4054-9137-d32c156ca641","projectName":"fullScanCycle","userAgent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:105.0) Gecko/20100101 Firefox/105.0","initiator":"user@sap.com","tags":{},"metadata":{"id":"7343f9f5-7633-40d5-b000-0a7a3c2c432e","type":"upload","Handler":{"UploadHandler":{"branch":"master","upload_url":"https://cx1.server.com/storage/st-gcp-9k90xv-uploads/b68ee5ba-3657-424f-9b68-05452300d5d7/271b80e3-b0d4-4be6-9f66-9469126b624f?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=ast%2F20230331%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20230331T083556Z\u0026X-Amz-Expires=86400\u0026X-Amz-Signature=94d74276d93945c37243f7ccec3d1e30b15d4d6ec79a869d3d9e46622fd89acd\u0026X-Amz-SignedHeaders=host"}},"configs":[{"type":"sast","value":{"presetName":"Checkmarx Default","incremental":"true","languageMode":"primary"}}],"project":{"id":"e7a7704c-4bfe-4054-9137-d32c156ca641"},"created_at":{"nanos":387074846,"seconds":1680251756}},"engines":["sast"],"sourceType":"zip","sourceOrigin":"Mozilla"}`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		scan, err := sys.GetScan("7343f9f5-7633-40d5-b000-0a7a3c2c432e")
		assert.NoError(t, err, "Error occurred but none expected")
		assert.Equal(t, "7343f9f5-7633-40d5-b000-0a7a3c2c432e", scan.ScanID, "ScanID is incorrect")
		assert.Equal(t, "master", scan.Branch, "Branch is incorrect")
		assert.Equal(t, 2, len(scan.StatusDetails), "StatusDetails is incorrect")
	})

	t.Run("test technical error", func(t *testing.T) {
		myTestClient := senderMock{httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)
		myTestClient.errorExp = true

		_, err := sys.GetScan("7343f9f5-7633-40d5-b000-0a7a3c2c432e")
		assert.Contains(t, fmt.Sprint(err), "Provoked technical error")
	})
}

func TestGetApplicationByName(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxOne_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"totalCount":6,"filteredTotalCount":6,"applications":[{"id":"8cf83fcf-ac61-4e32-b988-47cde3cc818c","name":"test_dev2","description":"","criticality":3,"rules":[],"projectIds":[],"tags":{},"createdAt":"2023-04-06T13:57:00.082719Z","updatedAt":"2023-04-06T13:57:00.082719Z"},{"id":"dee8573b-c58e-4945-a97c-a66884380093","name":"test_dev1","description":"","criticality":3,"rules":[],"projectIds":[],"tags":{},"createdAt":"2023-04-06T13:44:32.212065Z","updatedAt":"2023-04-06T13:44:32.212065Z"},{"id":"0ff00c77-b7e6-4d27-bd88-9e14520e06e6","name":"test_dev","description":"","criticality":3,"rules":[],"projectIds":[],"tags":{},"createdAt":"2023-04-06T13:24:36.459375Z","updatedAt":"2023-04-06T13:24:36.459375Z"},{"id":"5d482cfc-27ae-43e1-ba45-68d557df8423","name":"SSBA","description":"","criticality":3,"rules":[{"id":"e00a5b13-93d0-4128-8c32-9d6a46db85b0","type":"project.name.in","value":"ssba-zip;ssba-git;cx_cli_ssba_test"}],"projectIds":["2d75e828-6db9-4cfa-87e7-b953ad59ea25","f00a9d02-b552-4461-835a-c701e30957d8","f61cf5f0-fa91-4563-b87b-8154a4fd2408"],"tags":{},"createdAt":"2023-03-15T13:44:31.831175Z","updatedAt":"2023-03-15T13:44:31.831175Z"},{"id":"68f2f996-e7eb-495e-8829-8996241eb84e","name":"test_1","description":"","criticality":3,"rules":[{"id":"3a08b06e-a76a-4a48-bcde-1b43b9890f31","type":"project.name.in","value":"OAuth-CLI-test;test-piper-1;cx_cli_ssba_test"}],"projectIds":["2d75e828-6db9-4cfa-87e7-b953ad59ea25","db82605a-26e4-4693-a59c-ec1d584840d0","31c44a7c-0c68-492a-9921-052d336e5d5a"],"tags":{"TEST_APP":""},"createdAt":"2023-02-20T13:12:02.927562Z","updatedAt":"2023-02-20T13:12:02.927562Z"},{"id":"095dced0-60b0-4dd6-b1e8-0063fa04eaa7","name":"TEST","description":"","criticality":3,"rules":[{"id":"fc02a324-0706-4522-a89f-e24bcbf76cf7","type":"project.tag.key.exists","value":"test"}],"projectIds":["db82605a-26e4-4693-a59c-ec1d584840d0"],"tags":{"TEST_APP":""},"createdAt":"2023-01-12T13:22:38.222789Z","updatedAt":"2023-01-12T13:22:38.222789Z"}]}`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		apps, err := sys.GetApplicationsByName("test", 10)
		assert.NoError(t, err, "Error occurred but none expected")
		assert.Equal(t, 6, len(apps), "TotalCount is incorrect")

		app1, _ := sys.GetApplicationByName("test_dev2")
		assert.Equal(t, "8cf83fcf-ac61-4e32-b988-47cde3cc818c", app1.ApplicationID, "ApplicationID is incorrect")

		_, err = sys.GetApplicationByName("ssba")
		assert.Contains(t, fmt.Sprint(err), "no application found named ssba")
	})

	t.Run("test technical error", func(t *testing.T) {
		myTestClient := senderMock{httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)
		myTestClient.errorExp = true

		_, err := sys.GetApplicationsByName("test", 10)
		assert.Contains(t, fmt.Sprint(err), "Provoked technical error")
	})
}

func TestUpdateProject(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxOne_test")
	opts := piperHttp.ClientOptions{}

	requestJson := `{ "id": "702ba12b-ae61-48c0-9b6a-09b17666be32",
		"name": "test-apr24-piper",
		"tags": {
			"\"key1\"": "\"value1\"",
			"\"keywithoutvalue\"": "\"\""
		},
		"criticality": 3,
		"mainBranch": "",
		"privatePackage": false
	}`
	var project Project
	_ = json.Unmarshal([]byte(requestJson), &project)

	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: ``, httpStatusCode: 204}
		serverURL := "https://cx1.server.com"
		sys := SystemInstance{serverURL: serverURL, iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		err := sys.UpdateProject(&project)
		assert.NoError(t, err, "Error occurred but none expected")
		assert.Equal(t, serverURL+"/api/projects/"+project.ProjectID, myTestClient.urlCalled, "Called url incorrect")
		assert.Equal(t, "PUT", myTestClient.httpMethod, "HTTP method incorrect")
		var body Project
		_ = json.Unmarshal([]byte(myTestClient.requestBody), &body)
		assert.Equal(t, project, body, "Request body incorrect")

	})

	t.Run("test technical error", func(t *testing.T) {
		myTestClient := senderMock{httpStatusCode: 403}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)
		myTestClient.errorExp = true

		err := sys.UpdateProject(&project)
		assert.Contains(t, fmt.Sprint(err), "Provoked technical error")
	})
}
