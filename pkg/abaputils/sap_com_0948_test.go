//go:build unit
// +build unit

package abaputils

import (
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var conTest0948 ConnectionDetailsHTTP
var repoTest0948 Repository

func init() {

	conTest0948.User = "CC_USER"
	conTest0948.Password = "123abc"
	conTest0948.URL = "https://example.com"

	repoTest0948.Name = "/DMO/REPO"
	repoTest0948.Branch = "main"

	repoTest0948.IsByog = false
	repoTest0948.ByogAuthMethod = "token"
	repoTest0948.ByogUsername = "byogUser"
	repoTest0948.ByogPassword = "byogToken"

}

func TestRetry0948(t *testing.T) {
	t.Run("Test retry success", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{ "status" : "R", "UUID" : "GUID" }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : { "lang" : "de", "value" : "Software component lifecycle activities in progress. Try again later..."} } }`,
				`{ }`,
			},
			Token:      "myToken",
			StatusCode: 200,
			ErrorList: []error{
				nil,
				errors.New("HTTP 400"),
				nil,
			},
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond}

		api, err := apiManager.GetAPI(conTest0948, repoTest0948)
		api.setSleepTimeConfig(time.Nanosecond, 120*time.Nanosecond)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0948{}, api.(*SAP_COM_0948), "API has wrong type")

		errAction := api.(*SAP_COM_0948).triggerRequest(ConnectionDetailsHTTP{User: "CC_USER", Password: "abc123", URL: "https://example.com/path"}, []byte("{}"))
		assert.NoError(t, errAction)
		assert.Equal(t, "GUID", api.getUUID(), "API does not contain correct UUID")

	})

	t.Run("Test retry not allowed", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{ "status" : "R", "UUID" : "GUID" }`,
				`{"error" : { "code" : "A4C_A2G/224", "message" : "Error Text" } }`,
				`{ }`,
			},
			Token:      "myToken",
			StatusCode: 200,
			ErrorList: []error{
				nil,
				errors.New("HTTP 400"),
				nil,
			},
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond}

		api, err := apiManager.GetAPI(conTest0948, repoTest0948)
		api.setSleepTimeConfig(time.Nanosecond, 20*time.Nanosecond)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0948{}, api.(*SAP_COM_0948), "API has wrong type")

		errAction := api.(*SAP_COM_0948).triggerRequest(ConnectionDetailsHTTP{User: "CC_USER", Password: "abc123", URL: "https://example.com/path"}, []byte("{}"))
		assert.ErrorContains(t, errAction, "HTTP 400: A4C_A2G/224 - Error Text")
		assert.Empty(t, api.getUUID(), "API does not contain correct UUID")

	})

	t.Run("Test retry maxSleepTime", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{"error" : { "code" : "A4C_A2G/228", "message" : "Error Text" } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : "Error Text" } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : "Error Text" } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : "Error Text" } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : "Error Text" } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : "Error Text" } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : "Error Text" } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : "Error Text" } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : "Error Text" } }`,
				`{ }`,
			},
			Token:      "myToken",
			StatusCode: 200,
			ErrorList: []error{
				errors.New("HTTP 400"),
				errors.New("HTTP 400"),
				errors.New("HTTP 400"),
				errors.New("HTTP 400"),
				errors.New("HTTP 400"),
				errors.New("HTTP 400"),
				errors.New("HTTP 400"),
				errors.New("HTTP 400"),
				errors.New("HTTP 400"),
				nil,
			},
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond}

		api, err := apiManager.GetAPI(conTest0948, repoTest0948)
		api.setSleepTimeConfig(time.Nanosecond, 20*time.Nanosecond)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0948{}, api.(*SAP_COM_0948), "API has wrong type")

		api.(*SAP_COM_0948).maxRetries = 5

		errAction := api.(*SAP_COM_0948).triggerRequest(ConnectionDetailsHTTP{User: "CC_USER", Password: "abc123", URL: "https://example.com/path"}, []byte("{}"))
		assert.ErrorContains(t, errAction, "HTTP 400: A4C_A2G/228 - Error Text")
		assert.Empty(t, api.getUUID(), "API does not contain correct UUID")

		assert.Equal(t, 6, len(client.BodyList), "Expected maxSleepTime to limit requests")
	})

	t.Run("Test retry maxRetries", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{"error" : { "code" : "A4C_A2G/228", "message" : "Error Text" } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : "Error Text" } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : "Error Text" } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : "Error Text" } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : "Error Text" } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : "Error Text" } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : "Error Text" } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : "Error Text" } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : "Error Text" } }`,
				`{ }`,
			},
			Token:      "myToken",
			StatusCode: 200,
			ErrorList: []error{
				errors.New("HTTP 400"),
				errors.New("HTTP 400"),
				errors.New("HTTP 400"),
				errors.New("HTTP 400"),
				errors.New("HTTP 400"),
				errors.New("HTTP 400"),
				errors.New("HTTP 400"),
				errors.New("HTTP 400"),
				errors.New("HTTP 400"),
				nil,
			},
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond}

		api, err := apiManager.GetAPI(conTest0948, repoTest0948)
		api.setSleepTimeConfig(time.Nanosecond, 999*time.Nanosecond)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0948{}, api.(*SAP_COM_0948), "API has wrong type")

		api.(*SAP_COM_0948).maxRetries = 3

		errAction := api.(*SAP_COM_0948).triggerRequest(ConnectionDetailsHTTP{User: "CC_USER", Password: "abc123", URL: "https://example.com/path"}, []byte("{}"))
		assert.ErrorContains(t, errAction, "HTTP 400: A4C_A2G/228 - Error Text")
		assert.Empty(t, api.getUUID(), "API does not contain correct UUID")

		assert.Equal(t, 5, len(client.BodyList), "Expected maxRetries to limit requests")
	})

}
func TestClone0948(t *testing.T) {
	t.Run("Test Clone Success", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{ "status" : "R", "UUID" : "GUID" }`,
				`{ }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond}

		api, err := apiManager.GetAPI(conTest0948, repoTest0948)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0948{}, api.(*SAP_COM_0948), "API has wrong type")

		errClone := api.Clone()
		assert.NoError(t, errClone)
		assert.Equal(t, "GUID", api.getUUID(), "API does not contain correct UUID")
	})

	t.Run("Test Clone Failure", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{ "d" : {} }`,
				`{ "d" : {} }`,
				`{ }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond}

		api, err := apiManager.GetAPI(conTest0948, repoTest0948)
		api.setSleepTimeConfig(time.Nanosecond, 120*time.Nanosecond)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0948{}, api.(*SAP_COM_0948), "API has wrong type")

		errClone := api.Clone()
		assert.ErrorContains(t, errClone, "Request to ABAP System not successful")
		assert.Empty(t, api.getUUID(), "API does not contain correct UUID")
	})

	t.Run("Test Clone Retry", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{ "status" : "R", "UUID" : "GUID" }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : { "lang" : "de", "value" : "Software component lifecycle activities in progress. Try again later..."} } }`,
				`{ }`,
			},
			Token:      "myToken",
			StatusCode: 200,
			ErrorList: []error{
				nil,
				errors.New("HTTP 400"),
				nil,
			},
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond}

		api, err := apiManager.GetAPI(conTest0948, repoTest0948)
		api.setSleepTimeConfig(time.Nanosecond, 120*time.Nanosecond)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0948{}, api.(*SAP_COM_0948), "API has wrong type")

		errClone := api.Clone()
		assert.NoError(t, errClone)
		assert.Equal(t, "GUID", api.getUUID(), "API does not contain correct UUID")
	})

	t.Run("Test Clone Body Success", func(t *testing.T) {

		cloneBody, _ := repoTest0948.GetCloneRequestBody()
		assert.Equal(t, "{\"branch_name\":\"main\"}", string([]byte(cloneBody)), "Clone body is not correct")
	})

	t.Run("Test Clone Body Failure", func(t *testing.T) {

		repoTest0948.Branch = "wrongBranch"

		cloneBody, _ := repoTest0948.GetCloneRequestBody()
		assert.NotEqual(t, "{\"branch_name\":\"main\"}", string([]byte(cloneBody)), "Clone body should not match")

		repoTest0948.Branch = "main"

	})

	t.Run("Test Clone Body BYOG Success", func(t *testing.T) {

		repoTest0948.IsByog = true

		cloneBody, _ := repoTest0948.GetCloneRequestBody()
		assert.Equal(t, "{\"branch_name\":\"main\", \"auth_method\":\"token\", \"username\":\"byogUser\", \"password\":\"byogToken\"}", string([]byte(cloneBody)), "Clone body for byog parameter is not correct")

		repoTest0948.IsByog = false
	})

	t.Run("Test Clone Body BYOG Failure", func(t *testing.T) {

		repoTest0948.ByogPassword = "wrongToken"
		repoTest0948.IsByog = true

		cloneBody, _ := repoTest0948.GetCloneRequestBody()
		assert.NotEqual(t, "{\"branch_name\":\"main\", \"auth_method\":\"token\", \"username\":\"byogUser\", \"password\":\"byogToken\"}", string([]byte(cloneBody)), "Clone body for byog parameter should not match")

		repoTest0948.ByogPassword = "byogToken"
		repoTest0948.IsByog = false
	})

}

func TestPull0948(t *testing.T) {
	t.Run("Test Pull Success", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{ "status" : "R", "UUID" : "GUID" }`,
				`{ }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond}

		api, err := apiManager.GetAPI(conTest0948, repoTest0948)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0948{}, api.(*SAP_COM_0948), "API has wrong type")

		errPull := api.Pull()
		assert.NoError(t, errPull)
		assert.Equal(t, "GUID", api.getUUID(), "API does not contain correct UUID")
	})

	t.Run("Test Pull Failure", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{ "d" : {} }`,
				`{ }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond}

		api, err := apiManager.GetAPI(conTest0948, repoTest0948)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0948{}, api.(*SAP_COM_0948), "API has wrong type")

		errPull := api.Pull()
		assert.ErrorContains(t, errPull, "Request to ABAP System not successful")
		assert.Empty(t, api.getUUID(), "API does not contain correct UUID")
	})
}

func TestCheckout0948(t *testing.T) {
	t.Run("Test Checkout Success", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{ "status" : "R", "UUID" : "GUID" }`,
				`{ }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond}

		api, err := apiManager.GetAPI(conTest0948, repoTest0948)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0948{}, api.(*SAP_COM_0948), "API has wrong type")

		errCheckout := api.CheckoutBranch()
		assert.NoError(t, errCheckout)
		assert.Equal(t, "GUID", api.getUUID(), "API does not contain correct UUID")
	})

	t.Run("Test Checkout Failure", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{ "d" : {} }`,
				`{ }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond}

		api, err := apiManager.GetAPI(conTest0948, repoTest0948)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0948{}, api.(*SAP_COM_0948), "API has wrong type")

		errCheckoput := api.CheckoutBranch()
		assert.ErrorContains(t, errCheckoput, "Request to ABAP System not successful")
		assert.Empty(t, api.getUUID(), "API does not contain correct UUID")
	})

	t.Run("Test Checkout encodes branch slashes", func(t *testing.T) {
		client := &ClientMock{
			BodyList:   []string{`{ "status":"R","UUID":"GUID" }`, `{}`},
			StatusCode: 200,
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond}

		repo := Repository{
			Name:   "/DMO/REPO/",
			Branch: "feature/a-b_c.d/path",
		}

		api, err := apiManager.GetAPI(conTest0948, repo)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0948{}, api.(*SAP_COM_0948))

		errCheckout := api.CheckoutBranch()
		assert.NoError(t, errCheckout)
		assert.Equal(t, "GUID", api.getUUID())

		// Assert the actual requested URL has encoded slashes in the branch segment
		if assert.NotEmpty(t, client.RequestedURLs, "mock did not capture any request URLs") {
			got := client.RequestedURLs[len(client.RequestedURLs)-1]
			assert.Contains(t, got, "/Branches/%2FDMO%2FREPO%2F/feature%2Fa-b_c.d%2Fpath/SAP__self.checkout_branch")
			assert.NotContains(t, got, "%252F", "must not double-encode")
		}
	})
}

func TestGetRepo0948(t *testing.T) {
	t.Run("Test GetRepo Success", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{ "sc_name" : "testRepo1", "avail_on_inst" : true, "active_branch": "testBranch1" }`,
				`{"d" : [] }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond}

		api, err := apiManager.GetAPI(conTest0948, repoTest0948)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0948{}, api.(*SAP_COM_0948), "API has wrong type")

		cloned, activeBranch, errAction, _ := api.GetRepository()
		assert.True(t, cloned)
		assert.Equal(t, "testBranch1", activeBranch)
		assert.NoError(t, errAction)
	})
}

func TestCreateTag0948(t *testing.T) {
	t.Run("Test Tag Success", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{ "status" : "R", "UUID" : "GUID" }`,
				`{ }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond}

		api, err := apiManager.GetAPI(conTest0948, repoTest0948)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0948{}, api.(*SAP_COM_0948), "API has wrong type")

		errCreateTag := api.CreateTag(Tag{TagName: "myTag", TagDescription: "descr"})
		assert.NoError(t, errCreateTag)
		assert.Equal(t, "GUID", api.getUUID(), "API does not contain correct UUID")
	})

	t.Run("Test Tag Failure", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{ "d" : {} }`,
				`{ }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond}

		api, err := apiManager.GetAPI(conTest0948, repoTest0948)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0948{}, api.(*SAP_COM_0948), "API has wrong type")

		errCreateTag := api.CreateTag(Tag{TagName: "myTag", TagDescription: "descr"})
		assert.ErrorContains(t, errCreateTag, "Request to ABAP System not successful")
		assert.Empty(t, api.getUUID(), "API does not contain correct UUID")
	})

	t.Run("Test Tag Empty", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{ "d" : {} }`,
				`{ }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond}

		api, err := apiManager.GetAPI(conTest0948, repoTest0948)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0948{}, api.(*SAP_COM_0948), "API has wrong type")

		errCreateTag := api.CreateTag(Tag{})
		assert.ErrorContains(t, errCreateTag, "No Tag provided")
		assert.Empty(t, api.getUUID(), "API does not contain correct UUID")
	})
}

func TestSleepTime0948(t *testing.T) {
	t.Run("Test Sleep Time", func(t *testing.T) {

		api := SAP_COM_0948{
			retryMaxSleepTime:  120 * time.Nanosecond,
			retryBaseSleepUnit: 1 * time.Nanosecond,
		}

		expectedResults := make([]time.Duration, 12)
		expectedResults[0] = 0
		expectedResults[1] = 1
		expectedResults[2] = 1
		expectedResults[3] = 2
		expectedResults[4] = 3
		expectedResults[5] = 5
		expectedResults[6] = 8
		expectedResults[7] = 13
		expectedResults[8] = 21
		expectedResults[9] = 34
		expectedResults[10] = 55
		expectedResults[11] = 89
		results := make([]time.Duration, 12)
		var err error

		for i := 0; i <= 11; i++ {

			results[i], err = api.getSleepTime(i)
			assert.NoError(t, err)
		}
		assert.ElementsMatch(t, expectedResults, results)

		_, err = api.getSleepTime(-10)
		assert.Error(t, err)

		_, err = api.getSleepTime(12)
		assert.ErrorContains(t, err, "Exceeded max sleep time")
	})
}

func TestTimeConverter0948(t *testing.T) {

	api := SAP_COM_0948{}

	t.Run("Test example time", func(t *testing.T) {
		inputDate := "2024-05-02T09:25:40Z"
		expectedDate := "2024-05-02 09:25:40 +0000 UTC"
		result := api.ConvertTime(inputDate)
		assert.Equal(t, expectedDate, result.String(), "Dates do not match after conversion")
	})
	t.Run("Test Unix time", func(t *testing.T) {
		inputDate := "2023-12-24T16:19:29.000Z"
		expectedDate := "2023-12-24 16:19:29 +0000 UTC"
		result := api.ConvertTime(inputDate)
		assert.Equal(t, expectedDate, result.String(), "Dates do not match after conversion")
	})
	t.Run("Test unexpected format", func(t *testing.T) {
		inputDate := "2024-05-02T09:254:40Z"
		expectedDate := "1970-01-01 00:00:00 +0000 UTC"
		result := api.ConvertTime(inputDate)
		assert.Equal(t, expectedDate, result.String(), "Dates do not match after conversion")
	})
}

func TestGetExecutionLog(t *testing.T) {
	t.Run("Test Get Executionlog Success", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{ "value" : [{"index_no":1,"timestamp":"2021-08-23T12:00:00.000Z","type":"Success", "descr":"First log entry"}]}`,
				``,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond}

		api, _ := apiManager.GetAPI(conTest0948, Repository{Name: "/DMO/REPO"})

		results, errAction := api.GetExecutionLog()
		assert.NoError(t, errAction)
		assert.NotEmpty(t, results)
		assert.Equal(t, "First log entry", results.Value[0].Descr)
	})
}

func TestGetLogArchive(t *testing.T) {
	t.Run("Test Get Log Archive Success", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{ zip content from log archive endpoint }`,
				``,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		apiManager := &SoftwareComponentApiManager{Client: client}

		api, _ := apiManager.GetAPI(conTest0948, Repository{Name: "/DMO/REPO"})

		results, errAction := api.GetLogArchive()
		assert.NoError(t, errAction)
		assert.NotEmpty(t, results)
	})
}

func Test_PathEncoding_GitAllowed_Branches(t *testing.T) {
	tests := []struct {
		name    string
		branch  string
		wantSeg string // exact expected segment with leading slash
	}{
		//unchanged
		{"plain", "main", "/main"},
		{"hyphen", "feature-123", "/feature-123"},
		{"underscore", "release_candidate_1", "/release_candidate_1"},
		{"dot", "release-1.2.3", "/release-1.2.3"},
		{"mix", "hotfix-1_2.3", "/hotfix-1_2.3"},

		// slashes in branch names are valid in Git but must be escaped in URL segment
		{"single slash", "release/sprint-72", "/release%2Fsprint-72"},
		{"multiple slashes", "feature/a-b_c.d/path", "/feature%2Fa-b_c.d%2Fpath"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			api := &SAP_COM_0948{
				repository: Repository{
					Name:   "dummy", // irrelevant here
					Branch: tc.branch,
				},
			}
			got := api.getBranchNameForPath()
			assert.Equal(t, tc.wantSeg, got)
			assert.NotContains(t, got, "%252F", "must not double-encode slashes")
		})
	}
}

func Test_PathEncoding_GitAllowed_Repositories(t *testing.T) {
	tests := []struct {
		name    string
		repo    string
		wantSeg string
	}{
		{"simple", "repo", "/repo"},
		{"hyphen", "repo-name", "/repo-name"},
		{"underscore", "repo_name", "/repo_name"},
		{"dot", "repo.name", "/repo.name"},
		{"mix", "Repo.Name-1_2", "/Repo.Name-1_2"},
		{"slashes", "/DMO/REPO/", "/%2FDMO%2FREPO%2F"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			api := &SAP_COM_0948{
				repository: Repository{
					Name:   tc.repo,
					Branch: "main",
				},
			}
			got := api.getRepoNameForPath()
			assert.Equal(t, tc.wantSeg, got)
			assert.NotContains(t, got, "%252F", "must not double-encode anything")
		})
	}
}
