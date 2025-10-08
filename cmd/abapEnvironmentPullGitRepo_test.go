package cmd

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/piperutils"

	"github.com/stretchr/testify/assert"
)

var executionLogStringPull string
var logResultErrorPull string

func init() {
	executionLog := abaputils.LogProtocolResults{
		Results: []abaputils.LogProtocol{
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
	executionLogStringPull = string(executionLogResponse)
	logResultErrorPull = `{"d": { "sc_name": "/DMO/SWC", "status": "S", "to_Log_Overview": { "results": [ { "log_index": 1, "log_name": "Main Import", "type_of_found_issues": "Error", "timestamp": "/Date(1644332299000+0000)/", "to_Log_Protocol": { "results": [ { "log_index": 1, "index_no": "1", "log_name": "", "type": "Info", "descr": "Main import", "timestamp": null, "criticality": 0 } ] } } ] } } }`
}

func TestPullStep(t *testing.T) {
	t.Run("Run Step Successful", func(t *testing.T) {

		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		config := abapEnvironmentPullGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			RepositoryNames:   []string{"testRepo1"},
			LogOutput:         "STANDARD",
		}

		logResultSuccess := `{"d": { "sc_name": "/DMO/SWC", "status": "S", "to_Log_Overview": { "results": [ { "log_index": 1, "log_name": "Main Import", "type_of_found_issues": "Success", "timestamp": "/Date(1644332299000+0000)/", "to_Log_Protocol": { "results": [ { "log_index": 1, "index_no": "1", "log_name": "", "type": "Info", "descr": "Main import", "timestamp": null, "criticality": 0 } ] } } ] } } }`
		client := &abaputils.ClientMock{
			BodyList: []string{
				`{"d" : ` + executionLogStringPull + `}`,
				logResultSuccess,
				`{"d" : { "status" : "S" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "R" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		apiManager = &abaputils.SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Nanosecond, Force0510: true}

		var reports []piperutils.Path
		logOutputManager := abaputils.LogOutputManager{
			LogOutput:    config.LogOutput,
			PiperStep:    "pull",
			FileNameStep: "pull",
			StepReports:  reports,
		}

		err := runAbapEnvironmentPullGitRepo(&config, &autils, apiManager, &logOutputManager)
		assert.NoError(t, err, "Did not expect error")
		assert.Equal(t, 0, len(client.BodyList), "Not all requests were done")
	})

	t.Run("Run Step Failure", func(t *testing.T) {
		expectedErrorMessage := "Checking configuration failed: You have not specified any repository configuration to be pulled into the ABAP Environment System. Please make sure that you specified the repositories that should be pulled either in a dedicated file or via the parameter 'repositoryNames'. For more information please read the User documentation"

		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		receivedURI := "example.com/Entity"
		tokenExpected := "myToken"

		client := &abaputils.ClientMock{
			Body:       `{"d" : { "__metadata" : { "uri" : "` + receivedURI + `" } } }`,
			Token:      tokenExpected,
			StatusCode: 200,
		}

		config := abapEnvironmentPullGitRepoOptions{}

		var reports []piperutils.Path
		logOutputManager := abaputils.LogOutputManager{
			LogOutput:    "STANDARD",
			PiperStep:    "pull",
			FileNameStep: "pull",
			StepReports:  reports,
		}

		apiManager = &abaputils.SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Nanosecond, Force0510: true}
		err := runAbapEnvironmentPullGitRepo(&config, &autils, apiManager, &logOutputManager)
		assert.Equal(t, expectedErrorMessage, err.Error(), "Different error message expected")
	})

	t.Run("Success case: pull repos from file config", func(t *testing.T) {
		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		receivedURI := "example.com/Branches"
		client := &abaputils.ClientMock{
			Body:       `{"d" : { "__metadata" : { "uri" : "` + receivedURI + `" } } }`,
			Token:      "myToken",
			StatusCode: 200,
		}

		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir

		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		manifestFileString := `
repositories:
- name: 'testRepo'
  branch: 'testBranch'
  commitID: '9caede7f31028cd52333eb496434275687fefb47'
- name: 'testRepo2'
  branch: 'testBranch2'
- name: 'testRepo3'
  branch: 'testBranch3'`

		err := os.WriteFile("repositoriesTest.yml", []byte(manifestFileString), 0644)

		config := abapEnvironmentPullGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			Repositories:      "repositoriesTest.yml",
			LogOutput:         "STANDARD",
		}

		var reports []piperutils.Path
		logOutputManager := abaputils.LogOutputManager{
			LogOutput:    config.LogOutput,
			PiperStep:    "pull",
			FileNameStep: "pull",
			StepReports:  reports,
		}

		apiManager = &abaputils.SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Nanosecond, Force0510: true}
		err = runAbapEnvironmentPullGitRepo(&config, &autils, apiManager, &logOutputManager)
		assert.NoError(t, err)
	})

	t.Run("Status Error", func(t *testing.T) {
		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		body := `---
repositories:
- name: /DMO/REPO_A
  tag: v-1.0.1-build-0001
  branch: branchA
  version: 1.0.1
  commitID: ABCD1234
`
		file, _ := os.Create("filename.yaml")
		_, err := file.Write([]byte(body))
		assert.NoError(t, err)

		config := abapEnvironmentPullGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			Repositories:      "filename.yaml",
			LogOutput:         "STANDARD",
		}
		client := &abaputils.ClientMock{
			BodyList: []string{
				`{"d" : ` + executionLogStringPull + `}`,
				logResultErrorPull,
				`{"d" : { "EntitySets" : [ "LogOverviews" ] } }`,
				`{"d" : { "status" : "E" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "R" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		var reports []piperutils.Path
		logOutputManager := abaputils.LogOutputManager{
			LogOutput:    config.LogOutput,
			PiperStep:    "pull",
			FileNameStep: "pull",
			StepReports:  reports,
		}

		apiManager = &abaputils.SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Nanosecond, Force0510: true}
		err = runAbapEnvironmentPullGitRepo(&config, &autils, apiManager, &logOutputManager)
		if assert.Error(t, err, "Expected error") {
			assert.Equal(t, "Pull of the repository / software component '/DMO/REPO_A', commit 'ABCD1234' failed on the ABAP system", err.Error(), "Expected different error message")
		}
	})

	t.Run("Status Error, Ignore Commit", func(t *testing.T) {
		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		body := `---
repositories:
- name: /DMO/REPO_A
  tag: v-1.0.1-build-0001
  branch: branchA
  version: 1.0.1
  commitID: ABCD1234
`
		file, _ := os.Create("filename.yaml")
		_, err := file.Write([]byte(body))
		assert.NoError(t, err)

		config := abapEnvironmentPullGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			Repositories:      "filename.yaml",
			IgnoreCommit:      true,
			LogOutput:         "STANDARD",
		}
		client := &abaputils.ClientMock{
			BodyList: []string{
				`{"d" : ` + executionLogStringPull + `}`,
				logResultErrorPull,
				`{"d" : { "EntitySets" : [ "LogOverviews" ] } }`,
				`{"d" : { "status" : "E" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "R" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		var reports []piperutils.Path
		logOutputManager := abaputils.LogOutputManager{
			LogOutput:    config.LogOutput,
			PiperStep:    "pull",
			FileNameStep: "pull",
			StepReports:  reports,
		}

		apiManager = &abaputils.SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Nanosecond, Force0510: true}
		err = runAbapEnvironmentPullGitRepo(&config, &autils, apiManager, &logOutputManager)
		if assert.Error(t, err, "Expected error") {
			assert.Equal(t, "Pull of the repository / software component '/DMO/REPO_A', tag 'v-1.0.1-build-0001' failed on the ABAP system", err.Error(), "Expected different error message")
		}
	})

	t.Run("Status Error, With Commit", func(t *testing.T) {
		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		config := abapEnvironmentPullGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			RepositoryName:    "/DMO/SWC",
			CommitID:          "123456",
			IgnoreCommit:      false,
			LogOutput:         "STANDARD",
		}
		client := &abaputils.ClientMock{
			BodyList: []string{
				`{"d" : ` + executionLogStringPull + `}`,
				logResultErrorPull,
				`{"d" : { "EntitySets" : [ "LogOverviews" ] } }`,
				`{"d" : { "status" : "E" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "R" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		var reports []piperutils.Path
		logOutputManager := abaputils.LogOutputManager{
			LogOutput:    config.LogOutput,
			PiperStep:    "pull",
			FileNameStep: "pull",
			StepReports:  reports,
		}

		apiManager = &abaputils.SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Nanosecond, Force0510: true}
		err := runAbapEnvironmentPullGitRepo(&config, &autils, apiManager, &logOutputManager)
		if assert.Error(t, err, "Expected error") {
			assert.Equal(t, "Pull of the repository / software component '/DMO/SWC', commit '123456' failed on the ABAP system", err.Error(), "Expected different error message")
		}
	})

	t.Run("Status Error, RepositoryName without commit", func(t *testing.T) {
		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		config := abapEnvironmentPullGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			RepositoryName:    "/DMO/SWC",
			IgnoreCommit:      false,
			LogOutput:         "STANDARD",
		}
		client := &abaputils.ClientMock{
			BodyList: []string{
				`{"d" : ` + executionLogStringPull + `}`,
				logResultErrorPull,
				`{"d" : { "EntitySets" : [ "LogOverviews" ] } }`,
				`{"d" : { "status" : "E" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "R" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		var reports []piperutils.Path
		logOutputManager := abaputils.LogOutputManager{
			LogOutput:    config.LogOutput,
			PiperStep:    "pull",
			FileNameStep: "pull",
			StepReports:  reports,
		}

		apiManager = &abaputils.SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Nanosecond, Force0510: true}
		err := runAbapEnvironmentPullGitRepo(&config, &autils, apiManager, &logOutputManager)
		if assert.Error(t, err, "Expected error") {
			assert.Equal(t, "Pull of the repository / software component '/DMO/SWC' failed on the ABAP system", err.Error(), "Expected different error message")
		}
	})

	t.Run("Failure case: pull repos from empty file config", func(t *testing.T) {
		expectedErrorMessage := "Error in config file repositoriesTest.yml, AddonDescriptor doesn't contain any repositories"

		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		receivedURI := "example.com/Branches"
		client := &abaputils.ClientMock{
			Body:       `{"d" : { "__metadata" : { "uri" : "` + receivedURI + `" } } }`,
			Token:      "myToken",
			StatusCode: 200,
		}

		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		manifestFileString := ``

		manifestFileStringBody := []byte(manifestFileString)
		err := os.WriteFile("repositoriesTest.yml", manifestFileStringBody, 0644)

		config := abapEnvironmentPullGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			Repositories:      "repositoriesTest.yml",
			LogOutput:         "STANDARD",
		}

		var reports []piperutils.Path
		logOutputManager := abaputils.LogOutputManager{
			LogOutput:    config.LogOutput,
			PiperStep:    "pull",
			FileNameStep: "pull",
			StepReports:  reports,
		}

		apiManager = &abaputils.SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Nanosecond, Force0510: true}
		err = runAbapEnvironmentPullGitRepo(&config, &autils, apiManager, &logOutputManager)
		assert.EqualError(t, err, expectedErrorMessage)
	})

	t.Run("Failure case: pull repos from wrong file config", func(t *testing.T) {
		expectedErrorMessage := "Could not unmarshal repositoriesTest.yml"

		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		pollIntervall := abaputils.AUtilsMock{}
		defer pollIntervall.Cleanup()

		receivedURI := "example.com/Branches"
		client := &abaputils.ClientMock{
			Body:       `{"d" : { "__metadata" : { "uri" : "` + receivedURI + `" } } }`,
			Token:      "myToken",
			StatusCode: 200,
		}

		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		manifestFileString := `
- repo: 'testRepo'
- repo: 'testRepo2'`

		manifestFileStringBody := []byte(manifestFileString)
		err := os.WriteFile("repositoriesTest.yml", manifestFileStringBody, 0644)

		config := abapEnvironmentPullGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			Repositories:      "repositoriesTest.yml",
			LogOutput:         "STANDARD",
		}

		var reports []piperutils.Path
		logOutputManager := abaputils.LogOutputManager{
			LogOutput:    config.LogOutput,
			PiperStep:    "pull",
			FileNameStep: "pull",
			StepReports:  reports,
		}

		apiManager = &abaputils.SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Nanosecond, Force0510: true}
		err = runAbapEnvironmentPullGitRepo(&config, &autils, apiManager, &logOutputManager)
		assert.EqualError(t, err, expectedErrorMessage)
	})
}

func TestPullConfigChecker(t *testing.T) {
	t.Run("Success case: check config file", func(t *testing.T) {
		config := abapEnvironmentPullGitRepoOptions{
			Repositories: "test.file",
		}
		err := checkPullRepositoryConfiguration(config)
		assert.NoError(t, err)
	})
	t.Run("Success case: check config", func(t *testing.T) {
		config := abapEnvironmentPullGitRepoOptions{
			RepositoryNames: []string{"testRepo", "testRepo2"},
		}
		err := checkPullRepositoryConfiguration(config)
		assert.NoError(t, err)
	})
	t.Run("Failure case: empty config", func(t *testing.T) {
		errorMessage := "Checking configuration failed: You have not specified any repository configuration to be pulled into the ABAP Environment System. Please make sure that you specified the repositories that should be pulled either in a dedicated file or via the parameter 'repositoryNames'. For more information please read the User documentation"
		config := abapEnvironmentPullGitRepoOptions{}
		err := checkPullRepositoryConfiguration(config)
		assert.Equal(t, errorMessage, err.Error(), "Different error message expected")
	})
	t.Run("Failure case: config overload", func(t *testing.T) {
		errorMessage := "Checking configuration failed: Only one of the paramters `RepositoryName`,`RepositoryNames` or `Repositories` may be configured at the same time"
		config := abapEnvironmentPullGitRepoOptions{
			RepositoryNames: []string{"testRepo", "testRepo2"},
			RepositoryName:  "Test",
			CommitID:        "123456",
		}
		err := checkPullRepositoryConfiguration(config)
		assert.Equal(t, errorMessage, err.Error(), "Different error message expected")
	})
}

func TestHelpFunctions(t *testing.T) {
	t.Run("Ignore Commit", func(t *testing.T) {
		repo1 := abaputils.Repository{
			Name:     "Repo1",
			CommitID: "ABCD1234",
		}
		repo2 := abaputils.Repository{
			Name: "Repo2",
		}

		repoList := []abaputils.Repository{repo1, repo2}

		handleIgnoreCommit(repoList, true)

		assert.Equal(t, "", repoList[0].CommitID, "Expected emtpy CommitID")
		assert.Equal(t, "", repoList[1].CommitID, "Expected emtpy CommitID")
	})
	t.Run("Not Ignore Commit", func(t *testing.T) {
		repo1 := abaputils.Repository{
			Name:     "Repo1",
			CommitID: "ABCD1234",
		}
		repo2 := abaputils.Repository{
			Name: "Repo2",
		}

		repoList := []abaputils.Repository{repo1, repo2}

		handleIgnoreCommit(repoList, false)

		assert.Equal(t, "ABCD1234", repoList[0].CommitID, "Expected CommitID")
		assert.Equal(t, "", repoList[1].CommitID, "Expected emtpy CommitID")
	})
}
