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

// TODO: similar test for IAC
func TestGetScanSASTMetadata(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxOne_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"scanId":"03d66397-36df-40b5-8976-f38bcce695a7","projectId":"eac4dc3b-4bbf-4d04-87e5-3b3cedae38fb","loc":158,"fileCount":39,"isIncremental":false,"isIncrementalCanceled":false,"queryPreset":"Checkmarx Default"}`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		scanmeta, err := sys.GetScanSASTMetadata("03d66397-36df-40b5-8976-f38bcce695a7")
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

		_, err := sys.GetScanSASTMetadata("03d66397-36df-40b5-8976-f38bcce695a7")
		assert.Contains(t, fmt.Sprint(err), "Provoked technical error")
	})
}
func TestGetScanIACMetadata(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxOne_test")
	opts := piperHttp.ClientOptions{}
	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `{"scanId":"b2f0ad4e-414c-45a1-8e49-9b2af97be7b2","projectId":"7553c43d-d371-4061-9173-594f713bce31","loc":9,"kicsLoc":11,"fileCount":1}`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		scanmeta, err := sys.GetScanIACMetadata("b2f0ad4e-414c-45a1-8e49-9b2af97be7b2")
		assert.NoError(t, err, "Error occurred but none expected")

		assert.Equal(t, "b2f0ad4e-414c-45a1-8e49-9b2af97be7b2", scanmeta.ScanID, "ScanID is incorrect")
		assert.Equal(t, "7553c43d-d371-4061-9173-594f713bce31", scanmeta.ProjectID, "ProjectID is incorrect")
		assert.Equal(t, 9, scanmeta.LOC, "LOC is incorrect")
		assert.Equal(t, 11, scanmeta.IACLOC, "IAC LOC is incorrect")
		assert.Equal(t, 1, scanmeta.FileCount, "FileCount is incorrect")
		assert.Equal(t, IACDefaultBlankPreset, scanmeta.PresetName, "PresetName is incorrect")
	})

	t.Run("test technical error", func(t *testing.T) {
		myTestClient := senderMock{httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)
		myTestClient.errorExp = true

		_, err := sys.GetScanIACMetadata("03d66397-36df-40b5-8976-f38bcce695a7")
		assert.Contains(t, fmt.Sprint(err), "Provoked technical error")
	})
}

func TestGetIACFindingInfo(t *testing.T) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxOne_test")
	opts := piperHttp.ClientOptions{}
	result := ScanResult{
		Data: ScanResultData{
			QueryID:   scanresultQueryID{Value: "b03a748a-542d-44f4-bb86-9199ab4fd2d5"},
			Platform:  "Dockerfile",
			QueryName: "Healthcheck Instruction Missing",
		},
	}

	t.Run("test success", func(t *testing.T) {
		myTestClient := senderMock{responseBody: `[{"title":"Checkmarx predefined","key":"cx","children":[{"title":"common","key":"Cx-common","children":[{"isLeaf":true,"title":"Last User Is 'root'","key":"67fd0c4a-68cf-46d7-8c41-bc9fba7e40ae","data":{"custom":false,"cwe":250,"severity":"HIGH","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/67fd0c4a-68cf-46d7-8c41-bc9fba7e40ae"}},{"isLeaf":true,"title":"Missing User Instruction","key":"fd54f200-402c-4333-a5a4-36ef6709af2f","data":{"custom":false,"cwe":250,"severity":"HIGH","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/fd54f200-402c-4333-a5a4-36ef6709af2f"}},{"isLeaf":true,"title":"Add Instead of Copy","key":"9513a694-aa0d-41d8-be61-3271e056f36b","data":{"custom":false,"cwe":610,"severity":"MEDIUM","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/9513a694-aa0d-41d8-be61-3271e056f36b"}},{"isLeaf":true,"title":"Apt Get Install Pin Version Not Defined","key":"965a08d7-ef86-4f14-8792-4a3b2098937e","data":{"custom":false,"cwe":1357,"severity":"MEDIUM","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/965a08d7-ef86-4f14-8792-4a3b2098937e"}},{"isLeaf":true,"title":"Changing Default Shell Using RUN Command","key":"8a301064-c291-4b20-adcb-403fe7fd95fd","data":{"custom":false,"cwe":710,"severity":"MEDIUM","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/8a301064-c291-4b20-adcb-403fe7fd95fd"}},{"isLeaf":true,"title":"Gem Install Without Version","key":"22cd11f7-9c6c-4f6e-84c0-02058120b341","data":{"custom":false,"cwe":1357,"severity":"MEDIUM","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/22cd11f7-9c6c-4f6e-84c0-02058120b341"}},{"isLeaf":true,"title":"Image Version Not Explicit","key":"9efb0b2d-89c9-41a3-91ca-dcc0aec911fd","data":{"custom":false,"cwe":1357,"severity":"MEDIUM","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/9efb0b2d-89c9-41a3-91ca-dcc0aec911fd"}},{"isLeaf":true,"title":"Image Version Using 'latest'","key":"f45ea400-6bbe-4501-9fc7-1c3d75c32067","data":{"custom":false,"cwe":1357,"severity":"MEDIUM","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/f45ea400-6bbe-4501-9fc7-1c3d75c32067"}},{"isLeaf":true,"title":"Missing Version Specification In dnf install","key":"93d88cf7-f078-46a8-8ddc-178e03aeacf1","data":{"custom":false,"cwe":1357,"severity":"MEDIUM","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/93d88cf7-f078-46a8-8ddc-178e03aeacf1"}},{"isLeaf":true,"title":"Missing Zypper Non-interactive Switch","key":"45e1fca5-f90e-465d-825f-c2cb63fa3944","data":{"custom":false,"cwe":710,"severity":"MEDIUM","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/45e1fca5-f90e-465d-825f-c2cb63fa3944"}},{"isLeaf":true,"title":"NPM Install Command Without Pinned Version","key":"e36d8880-3f78-4546-b9a1-12f0745ca0d5","data":{"custom":false,"cwe":1357,"severity":"MEDIUM","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/e36d8880-3f78-4546-b9a1-12f0745ca0d5"}},{"isLeaf":true,"title":"Not Using JSON In CMD And ENTRYPOINT Arguments","key":"b86987e1-6397-4619-81d5-8807f2387c79","data":{"custom":false,"cwe":573,"severity":"MEDIUM","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/b86987e1-6397-4619-81d5-8807f2387c79"}},{"isLeaf":true,"title":"Run Using Sudo","key":"8ada6e80-0ade-439e-b176-0b28f6bce35a","data":{"custom":false,"cwe":440,"severity":"MEDIUM","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/8ada6e80-0ade-439e-b176-0b28f6bce35a"}},{"isLeaf":true,"title":"Unpinned Package Version in Apk Add","key":"d3499f6d-1651-41bb-a9a7-de925fea487b","data":{"custom":false,"cwe":1357,"severity":"MEDIUM","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/d3499f6d-1651-41bb-a9a7-de925fea487b"}},{"isLeaf":true,"title":"Unpinned Package Version in Pip Install","key":"02d9c71f-3ee8-4986-9c27-1a20d0d19bfc","data":{"custom":false,"cwe":1357,"severity":"MEDIUM","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/02d9c71f-3ee8-4986-9c27-1a20d0d19bfc"}},{"isLeaf":true,"title":"Yum install Without Version","key":"6452c424-1d92-4deb-bb18-a03e95d579c4","data":{"custom":false,"cwe":1357,"severity":"MEDIUM","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/6452c424-1d92-4deb-bb18-a03e95d579c4"}},{"isLeaf":true,"title":"APT-GET Missing Flags To Avoid Manual Input","key":"77783205-c4ca-4f80-bb80-c777f267c547","data":{"custom":false,"cwe":710,"severity":"LOW","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/77783205-c4ca-4f80-bb80-c777f267c547"}},{"isLeaf":true,"title":"COPY '--from' References Current FROM Alias","key":"cdddb86f-95f6-4fc4-b5a1-483d9afceb2b","data":{"custom":false,"cwe":706,"severity":"LOW","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/cdddb86f-95f6-4fc4-b5a1-483d9afceb2b"}},{"isLeaf":true,"title":"Chown Flag Exists","key":"aa93e17f-b6db-4162-9334-c70334e7ac28","data":{"custom":false,"cwe":282,"severity":"LOW","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/aa93e17f-b6db-4162-9334-c70334e7ac28"}},{"isLeaf":true,"title":"Copy With More Than Two Arguments Not Ending With Slash","key":"6db6e0c2-32a3-4a2e-93b5-72c35f4119db","data":{"custom":false,"cwe":628,"severity":"LOW","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/6db6e0c2-32a3-4a2e-93b5-72c35f4119db"}},{"isLeaf":true,"title":"Curl or Wget Instead of Add","key":"4b410d24-1cbe-4430-a632-62c9a931cf1c","data":{"custom":false,"cwe":610,"severity":"LOW","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/4b410d24-1cbe-4430-a632-62c9a931cf1c"}},{"isLeaf":true,"title":"Exposing Port 22 (SSH)","key":"5907595b-5b6d-4142-b173-dbb0e73fbff8","data":{"custom":false,"cwe":710,"severity":"LOW","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/5907595b-5b6d-4142-b173-dbb0e73fbff8"}},{"isLeaf":true,"title":"Healthcheck Instruction Missing","key":"b03a748a-542d-44f4-bb86-9199ab4fd2d5","data":{"custom":false,"cwe":710,"severity":"LOW","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/b03a748a-542d-44f4-bb86-9199ab4fd2d5"}},{"isLeaf":true,"title":"MAINTAINER Instruction Being Used","key":"99614418-f82b-4852-a9ae-5051402b741c","data":{"custom":false,"cwe":710,"severity":"LOW","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/99614418-f82b-4852-a9ae-5051402b741c"}},{"isLeaf":true,"title":"Missing Dnf Clean All","key":"295acb63-9246-4b21-b441-7c1f1fb62dc0","data":{"custom":false,"cwe":459,"severity":"LOW","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/295acb63-9246-4b21-b441-7c1f1fb62dc0"}},{"isLeaf":true,"title":"Missing Flag From Dnf Install","key":"7ebd323c-31b7-4e5b-b26f-de5e9e477af8","data":{"custom":false,"cwe":710,"severity":"LOW","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/7ebd323c-31b7-4e5b-b26f-de5e9e477af8"}},{"isLeaf":true,"title":"Missing Zypper Clean","key":"38300d1a-feb2-4a48-936a-d1ef1cd24313","data":{"custom":false,"cwe":459,"severity":"LOW","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/38300d1a-feb2-4a48-936a-d1ef1cd24313"}},{"isLeaf":true,"title":"Multiple CMD Instructions Listed","key":"41c195f4-fc31-4a5c-8a1b-90605538d49f","data":{"custom":false,"cwe":1041,"severity":"LOW","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/41c195f4-fc31-4a5c-8a1b-90605538d49f"}},{"isLeaf":true,"title":"Multiple ENTRYPOINT Instructions Listed","key":"6938958b-3f1a-451c-909b-baeee14bdc97","data":{"custom":false,"cwe":1041,"severity":"LOW","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/6938958b-3f1a-451c-909b-baeee14bdc97"}},{"isLeaf":true,"title":"Multiple RUN, ADD, COPY, Instructions Listed","key":"0008c003-79aa-42d8-95b8-1c2fe37dbfe6","data":{"custom":false,"cwe":710,"severity":"LOW","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/0008c003-79aa-42d8-95b8-1c2fe37dbfe6"}},{"isLeaf":true,"title":"Pip install Keeping Cached Packages","key":"f2f903fb-b977-461e-98d7-b3e2185c6118","data":{"custom":false,"cwe":459,"severity":"LOW","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/f2f903fb-b977-461e-98d7-b3e2185c6118"}},{"isLeaf":true,"title":"RUN Instruction Using 'cd' Instead of WORKDIR","key":"f4a6bcd3-e231-4acf-993c-aa027be50d2e","data":{"custom":false,"cwe":710,"severity":"LOW","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/f4a6bcd3-e231-4acf-993c-aa027be50d2e"}},{"isLeaf":true,"title":"Run Using 'wget' and 'curl'","key":"fc775e75-fcfb-4c98-b2f2-910c5858b359","data":{"custom":false,"cwe":1041,"severity":"LOW","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/fc775e75-fcfb-4c98-b2f2-910c5858b359"}},{"isLeaf":true,"title":"Run Using apt","key":"b84a0b47-2e99-4c9f-8933-98bcabe2b94d","data":{"custom":false,"cwe":758,"severity":"LOW","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/b84a0b47-2e99-4c9f-8933-98bcabe2b94d"}},{"isLeaf":true,"title":"Same Alias In Different Froms","key":"f2daed12-c802-49cd-afed-fe41d0b82fed","data":{"custom":false,"cwe":694,"severity":"LOW","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/f2daed12-c802-49cd-afed-fe41d0b82fed"}},{"isLeaf":true,"title":"Shell Running A Pipe Without Pipefail Flag","key":"efbf148a-67e9-42d2-ac47-02fa1c0d0b22","data":{"custom":false,"cwe":710,"severity":"LOW","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/efbf148a-67e9-42d2-ac47-02fa1c0d0b22"}},{"isLeaf":true,"title":"Update Instruction Alone","key":"9bae49be-0aa3-4de5-bab2-4c3a069e40cd","data":{"custom":false,"cwe":710,"severity":"LOW","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/9bae49be-0aa3-4de5-bab2-4c3a069e40cd"}},{"isLeaf":true,"title":"Using Unnamed Build Stages","key":"68a51e22-ae5a-4d48-8e87-b01a323605c9","data":{"custom":false,"cwe":710,"severity":"LOW","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/68a51e22-ae5a-4d48-8e87-b01a323605c9"}},{"isLeaf":true,"title":"WORKDIR Path Not Absolute","key":"6b376af8-cfe8-49ab-a08d-f32de23661a4","data":{"custom":false,"cwe":665,"severity":"LOW","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/6b376af8-cfe8-49ab-a08d-f32de23661a4"}},{"isLeaf":true,"title":"Yum Clean All Missing","key":"00481784-25aa-4a55-8633-3136dfcf4f37","data":{"custom":false,"cwe":459,"severity":"LOW","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/00481784-25aa-4a55-8633-3136dfcf4f37"}},{"isLeaf":true,"title":"Yum Install Allows Manual Input","key":"6e19193a-8753-436d-8a09-76dcff91bb03","data":{"custom":false,"cwe":710,"severity":"LOW","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/6e19193a-8753-436d-8a09-76dcff91bb03"}},{"isLeaf":true,"title":"Zypper Install Without Version","key":"562952e4-0348-4dea-9826-44f3a2c6117b","data":{"custom":false,"cwe":1357,"severity":"LOW","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/562952e4-0348-4dea-9826-44f3a2c6117b"}},{"isLeaf":true,"title":"APT-GET Not Avoiding Additional Packages","key":"7384dfb2-fcd1-4fbf-91cd-6c44c318c33c","data":{"custom":false,"cwe":710,"severity":"INFO","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/7384dfb2-fcd1-4fbf-91cd-6c44c318c33c"}},{"isLeaf":true,"title":"Apk Add Using Local Cache Path","key":"ae9c56a6-3ed1-4ac0-9b54-31267f51151d","data":{"custom":false,"cwe":459,"severity":"INFO","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/ae9c56a6-3ed1-4ac0-9b54-31267f51151d"}},{"isLeaf":true,"title":"Apt Get Install Lists Were Not Deleted","key":"df746b39-6564-4fed-bf85-e9c44382303c","data":{"custom":false,"cwe":459,"severity":"INFO","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/df746b39-6564-4fed-bf85-e9c44382303c"}},{"isLeaf":true,"title":"Run Utilities And POSIX Commands","key":"9b6b0f38-92a2-41f9-b881-3a1083d99f1b","data":{"custom":false,"cwe":710,"severity":"INFO","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/9b6b0f38-92a2-41f9-b881-3a1083d99f1b"}},{"isLeaf":true,"title":"UNIX Ports Out Of Range","key":"71bf8cf8-f0a1-42fa-b9d2-d10525e0a38e","data":{"custom":false,"cwe":682,"severity":"INFO","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/71bf8cf8-f0a1-42fa-b9d2-d10525e0a38e"}},{"isLeaf":true,"title":"Using Platform Flag with FROM Command","key":"b16e8501-ef3c-44e1-a543-a093238099c9","data":{"custom":false,"cwe":695,"severity":"INFO","url":"https://docs.kics.io/2.1.20/queries/dockerfile-queries/b16e8501-ef3c-44e1-a543-a093238099c9"}}]}]}]`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger, iacQueryCache: make(map[string]IACFindingInfo)}
		myTestClient.SetOptions(opts)

		info, err := sys.GetIACFindingInfo(result)
		assert.NoError(t, err, "Error occurred but none expected")

		assert.Equal(t, 710, info.Cwe, "CWE is incorrect")
		assert.Equal(t, "https://docs.kics.io/2.1.20/queries/dockerfile-queries/b03a748a-542d-44f4-bb86-9199ab4fd2d5", info.URL, "URL is incorrect")
	})

	t.Run("test technical error", func(t *testing.T) {
		myTestClient := senderMock{httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx1.server.com", iamURL: "https://cx1iam.server.com", tenant: "tenant", client: &myTestClient, logger: logger, iacQueryCache: make(map[string]IACFindingInfo)}
		myTestClient.SetOptions(opts)
		myTestClient.errorExp = true

		_, err := sys.GetIACFindingInfo(result)
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
