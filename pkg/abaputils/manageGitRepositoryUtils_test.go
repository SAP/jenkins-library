//go:build unit
// +build unit

package abaputils

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var executionLogString string

func init() {
	executionLog := LogProtocolResults{
		Count: fmt.Sprint(math.Round(numberOfEntriesPerPage * 1.5)),
		Results: []LogProtocol{
			{
				ProtocolLine:  1,
				OverviewIndex: 1,
				Type:          "LogEntry",
				Description:   "S",
				Timestamp:     "/Date(1644332299000+0000)/",
			},
		},
	}

	executionLogResponse, _ := json.Marshal(executionLog)
	executionLogString = string(executionLogResponse)
}

func TestPollEntity(t *testing.T) {

	t.Run("Test poll entity - success case", func(t *testing.T) {

		logResultSuccess := fmt.Sprintf(`{"d": { "sc_name": "/DMO/SWC", "status": "S", "to_Log_Overview": { "results": [ { "log_index": 1, "log_name": "Main Import", "type_of_found_issues": "Success", "timestamp": "/Date(1644332299000+0000)/", "to_Log_Protocol": { "results": [ { "log_index": 1, "index_no": "1", "log_name": "", "type": "Info", "descr": "Main import", "timestamp": null, "criticality": 0 } ] } } ] } } }`)
		client := &ClientMock{
			BodyList: []string{
				`{"d" : ` + executionLogString + `}`,
				`{"d" : ` + executionLogString + `}`,
				logResultSuccess,
				`{"d" : { "status" : "S" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "Q" } }`,
				`{}`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		con := ConnectionDetailsHTTP{
			User:       "MY_USER",
			Password:   "MY_PW",
			URL:        "https://api.endpoint.com/Entity/",
			XCsrfToken: "MY_TOKEN",
		}

		swcManager := SoftwareComponentApiManager{Client: client, Force0510: true}
		repo := Repository{Name: "testRepo1"}
		api, _ := swcManager.GetAPI(con, repo)

		status, _ := PollEntity(api, 0)
		assert.Equal(t, "S", status)
		assert.Equal(t, 0, len(client.BodyList), "Not all requests were done")
	})

	t.Run("Test poll entity - error case", func(t *testing.T) {
		logResultError := fmt.Sprintf(`{"d": { "sc_name": "/DMO/SWC", "status": "S", "to_Log_Overview": { "results": [ { "log_index": 1, "log_name": "Main Import", "type_of_found_issues": "Error", "timestamp": "/Date(1644332299000+0000)/", "to_Log_Protocol": { "results": [ { "log_index": 1, "index_no": "1", "log_name": "", "type": "Info", "descr": "Main import", "timestamp": null, "criticality": 0 } ] } } ] } } }`)
		client := &ClientMock{
			BodyList: []string{
				`{"d" : ` + executionLogString + `}`,
				`{"d" : ` + executionLogString + `}`,
				logResultError,
				`{"d" : { "status" : "E" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "Q" } }`,
				`{}`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		con := ConnectionDetailsHTTP{
			User:       "MY_USER",
			Password:   "MY_PW",
			URL:        "https://api.endpoint.com/Entity/",
			XCsrfToken: "MY_TOKEN",
		}

		swcManager := SoftwareComponentApiManager{Client: client, Force0510: true}
		repo := Repository{Name: "testRepo1"}
		api, _ := swcManager.GetAPI(con, repo)

		status, _ := PollEntity(api, 0)
		assert.Equal(t, "E", status)
		assert.Equal(t, 0, len(client.BodyList), "Not all requests were done")
	})
}

func TestGetRepositories(t *testing.T) {
	t.Run("Get Repositories from config - success case", func(t *testing.T) {
		expectedRepositoryList := []Repository{{
			Name:   "testRepo",
			Branch: "testBranch",
		}, {
			Name:   "testRepo2",
			Branch: "testBranch2",
		}, {
			Name:   "testRepo3",
			Branch: "testBranch3",
		}, {
			Name:   "testRepository",
			Branch: "testBranch",
		}, {
			Name: "testRepository",
		}}

		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)

		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		manifestFileString := `
repositories:
- name: 'testRepo'
  branch: 'testBranch'
- name: 'testRepo2'
  branch: 'testBranch2'
- name: 'testRepo3'
  branch: 'testBranch3'`

		err := os.WriteFile("repositoriesTest.yml", []byte(manifestFileString), 0644)

		config := RepositoriesConfig{
			BranchName:      "testBranch",
			RepositoryName:  "testRepository",
			RepositoryNames: []string{"testRepository"},
			Repositories:    "repositoriesTest.yml",
		}

		repositories, err := GetRepositories(&config, true)

		assert.Equal(t, expectedRepositoryList, repositories)
		assert.NoError(t, err)
	})
	t.Run("Get Repositories from file config - failure case", func(t *testing.T) {
		expectedRepositoryList := []Repository([]Repository{})
		expectedErrorMessage := "Error in config file repositoriesTest.yml, AddonDescriptor doesn't contain any repositories"

		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)

		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		manifestFileString := `
repositories:
- repo: 'testRepo'
- repo: 'testRepo2'`

		err := os.WriteFile("repositoriesTest.yml", []byte(manifestFileString), 0644)

		config := RepositoriesConfig{
			Repositories: "repositoriesTest.yml",
		}

		repositories, err := GetRepositories(&config, false)

		assert.Equal(t, expectedRepositoryList, repositories)
		assert.EqualError(t, err, expectedErrorMessage)
	})
	t.Run("Get Repositories from config - failure case", func(t *testing.T) {
		expectedRepositoryList := []Repository([]Repository{})
		expectedErrorMessage := "Error in config file repositoriesTest.yml, AddonDescriptor doesn't contain any repositories"

		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)

		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		manifestFileString := `
repositories:
- repo: 'testRepo'
- repo: 'testRepo2'`

		err := os.WriteFile("repositoriesTest.yml", []byte(manifestFileString), 0644)

		config := RepositoriesConfig{
			Repositories: "repositoriesTest.yml",
		}

		repositories, err := GetRepositories(&config, false)

		assert.Equal(t, expectedRepositoryList, repositories)
		assert.EqualError(t, err, expectedErrorMessage)
	})
	t.Run("Get Repositories from empty config - failure case", func(t *testing.T) {
		expectedRepositoryList := []Repository([]Repository{})
		expectedErrorMessage := "Failed to read repository configuration: You have not specified any repository configuration. Please make sure that you have correctly specified it. For more information please read the User documentation"

		config := RepositoriesConfig{}

		repositories, err := GetRepositories(&config, false)

		assert.Equal(t, expectedRepositoryList, repositories)
		assert.EqualError(t, err, expectedErrorMessage)
	})
}

func TestCreateLogStrings(t *testing.T) {
	t.Run("Clone LogString Tag and Commit", func(t *testing.T) {
		repo := Repository{
			Name:     "/DMO/REPO",
			Branch:   "main",
			CommitID: "1234567",
			Tag:      "myTag",
		}
		logString := repo.GetCloneLogString()
		assert.Equal(t, "repository / software component '/DMO/REPO', branch 'main', commit '1234567'", logString, "Expected different string")
	})
	t.Run("Clone LogString Tag", func(t *testing.T) {
		repo := Repository{
			Name:   "/DMO/REPO",
			Branch: "main",
			Tag:    "myTag",
		}
		logString := repo.GetCloneLogString()
		assert.Equal(t, "repository / software component '/DMO/REPO', branch 'main', tag 'myTag'", logString, "Expected different string")
	})
	t.Run("Pull LogString Tag and Commit", func(t *testing.T) {
		repo := Repository{
			Name:     "/DMO/REPO",
			Branch:   "main",
			CommitID: "1234567",
			Tag:      "myTag",
		}
		logString := repo.GetPullLogString()
		assert.Equal(t, "repository / software component '/DMO/REPO', commit '1234567'", logString, "Expected different string")
	})
	t.Run("Pull LogString Tag", func(t *testing.T) {
		repo := Repository{
			Name:   "/DMO/REPO",
			Branch: "main",
			Tag:    "myTag",
		}
		logString := repo.GetPullLogString()
		assert.Equal(t, "repository / software component '/DMO/REPO', tag 'myTag'", logString, "Expected different string")
	})
}

func TestCreateRequestBodies(t *testing.T) {
	t.Run("Clone Body Tag and Commit", func(t *testing.T) {
		repo := Repository{
			Name:     "/DMO/REPO",
			Branch:   "main",
			CommitID: "1234567",
			Tag:      "myTag",
		}
		body := repo.GetCloneRequestBody()
		assert.Equal(t, `{"branch_name":"main", "commit_id":"1234567"}`, body, "Expected different body")
	})
	t.Run("Clone Body Tag", func(t *testing.T) {
		repo := Repository{
			Name:   "/DMO/REPO",
			Branch: "main",
			Tag:    "myTag",
		}
		body := repo.GetCloneRequestBodyWithSWC()
		assert.Equal(t, `{"sc_name":"/DMO/REPO", "branch_name":"main", "tag_name":"myTag"}`, body, "Expected different body")
	})
	t.Run("Pull Body Tag and Commit", func(t *testing.T) {
		repo := Repository{
			Name:     "/DMO/REPO",
			Branch:   "main",
			CommitID: "1234567",
			Tag:      "myTag",
		}
		body := repo.GetPullRequestBody()
		assert.Equal(t, `{"sc_name":"/DMO/REPO", "commit_id":"1234567"}`, body, "Expected different body")
	})
	t.Run("Pull Body Tag", func(t *testing.T) {
		repo := Repository{
			Name:   "/DMO/REPO",
			Branch: "main",
			Tag:    "myTag",
		}
		body := repo.GetPullRequestBody()
		assert.Equal(t, `{"sc_name":"/DMO/REPO", "tag_name":"myTag"}`, body, "Expected different body")
	})
}

func TestExecutionLogOutput(t *testing.T) {
	t.Run("Test execution log output", func(t *testing.T) {

		executionLogValue := []ExecutionLogValue{
			{IndexNo: 1, Type: "Success", Descr: "Something went well", Timestamp: "/Date(1644332299000+0000)/"},
			{IndexNo: 2, Type: "Error", Descr: "Something went wrong", Timestamp: "/Date(1644332299000+0000)/"},
		}
		executionLog := ExecutionLog{
			Value: executionLogValue,
		}
		printExecutionLogs(executionLog)

	})
}
