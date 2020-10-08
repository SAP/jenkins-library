package cmd

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
)

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
		}

		client := &abaputils.ClientMock{
			BodyList: []string{
				`{"d" : { "status" : "S" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "R" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		err := runAbapEnvironmentPullGitRepo(&config, nil, &autils, client)
		assert.NoError(t, err, "Did not expect error")
	})
	t.Run("Run Step Failure", func(t *testing.T) {
		expectedErrorMessage := "Something failed during the pull of the repositories: Checking configuration failed: You have not specified any repository configuration to be pulled into the ABAP Environment System. Please make sure that you specified the repositories that should be pulled either in a dedicated file or via the parameter 'repositoryNames'. For more information please read the User documentation"

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
		err := runAbapEnvironmentPullGitRepo(&config, nil, &autils, client)
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

		dir, err := ioutil.TempDir("", "test pull repos")
		if err != nil {
			t.Fatal("Failed to create temporary directory")
		}
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir

		defer func() {
			_ = os.Chdir(oldCWD)
			_ = os.RemoveAll(dir)
		}()

		manifestFileString := `
repositories:
- name: 'testRepo'
  branch: 'testBranch'
- name: 'testRepo2'
  branch: 'testBranch2'
- name: 'testRepo3'
  branch: 'testBranch3'`

		err = ioutil.WriteFile("repositoriesTest.yml", []byte(manifestFileString), 0644)

		config := abapEnvironmentPullGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			Repositories:      "repositoriesTest.yml",
		}
		err = runAbapEnvironmentPullGitRepo(&config, nil, &autils, client)
		assert.NoError(t, err)
	})
	t.Run("Failure case: pull repos from empty file config", func(t *testing.T) {
		expectedErrorMessage := "Something failed during the pull of the repositories: Error in config file repositoriesTest.yml, AddonDescriptor doesn't contain any repositories"

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

		dir, err := ioutil.TempDir("", "test pull repos")
		if err != nil {
			t.Fatal("Failed to create temporary directory")
		}
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
			_ = os.RemoveAll(dir)
		}()

		manifestFileString := ``

		manifestFileStringBody := []byte(manifestFileString)
		err = ioutil.WriteFile("repositoriesTest.yml", manifestFileStringBody, 0644)

		config := abapEnvironmentPullGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			Repositories:      "repositoriesTest.yml",
		}
		err = runAbapEnvironmentPullGitRepo(&config, nil, &autils, client)
		assert.EqualError(t, err, expectedErrorMessage)
	})
	t.Run("Failure case: pull repos from wrong file config", func(t *testing.T) {
		expectedErrorMessage := "Something failed during the pull of the repositories: Could not unmarshal repositoriesTest.yml"

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

		dir, err := ioutil.TempDir("", "test pull repos")
		if err != nil {
			t.Fatal("Failed to create temporary directory")
		}
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
			_ = os.RemoveAll(dir)
		}()

		manifestFileString := `
- repo: 'testRepo'
- repo: 'testRepo2'`

		manifestFileStringBody := []byte(manifestFileString)
		err = ioutil.WriteFile("repositoriesTest.yml", manifestFileStringBody, 0644)

		config := abapEnvironmentPullGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			Repositories:      "repositoriesTest.yml",
		}
		err = runAbapEnvironmentPullGitRepo(&config, nil, &autils, client)
		assert.EqualError(t, err, expectedErrorMessage)
	})
}

func TestTriggerPull(t *testing.T) {

	t.Run("Test trigger pull: success case", func(t *testing.T) {

		receivedURI := "example.com/Entity"
		uriExpected := receivedURI + "?$expand=to_Execution_log,to_Transport_log"
		tokenExpected := "myToken"

		client := &abaputils.ClientMock{
			Body:       `{"d" : { "__metadata" : { "uri" : "` + receivedURI + `" } } }`,
			Token:      tokenExpected,
			StatusCode: 200,
		}
		config := abapEnvironmentPullGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			RepositoryNames:   []string{"testRepo1", "testRepo2"},
		}

		con := abaputils.ConnectionDetailsHTTP{
			User:     "MY_USER",
			Password: "MY_PW",
			URL:      "https://api.endpoint.com/Entity/",
		}
		entityConnection, err := triggerPull(config.RepositoryNames[0], con, client)
		assert.Nil(t, err)
		assert.Equal(t, uriExpected, entityConnection.URL)
		assert.Equal(t, tokenExpected, entityConnection.XCsrfToken)
	})

	t.Run("Test trigger pull: ABAP Error", func(t *testing.T) {

		errorMessage := "ABAP Error Message"
		errorCode := "ERROR/001"
		HTTPErrorMessage := "HTTP Error Message"
		combinedErrorMessage := "HTTP Error Message: ERROR/001 - ABAP Error Message"

		client := &abaputils.ClientMock{
			Body:       `{"error" : { "code" : "` + errorCode + `", "message" : { "lang" : "en", "value" : "` + errorMessage + `" } } }`,
			Token:      "myToken",
			StatusCode: 400,
			Error:      errors.New(HTTPErrorMessage),
		}
		config := abapEnvironmentPullGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			RepositoryNames:   []string{"testRepo1", "testRepo2"},
		}

		con := abaputils.ConnectionDetailsHTTP{
			User:     "MY_USER",
			Password: "MY_PW",
			URL:      "https://api.endpoint.com/Entity/",
		}
		_, err := triggerPull(config.RepositoryNames[0], con, client)
		assert.Equal(t, combinedErrorMessage, err.Error(), "Different error message expected")
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
}
