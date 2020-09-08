package cmd

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestCheckoutBranchStep(t *testing.T) {
	t.Run("Run Step Successful", func(t *testing.T) {

		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		config := abapEnvironmentCheckoutBranchOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			RepositoryName:    "testRepo1",
			BranchName:        "testBranch",
		}

		client := &abaputils.ClientMock{
			BodyList: []string{
				`{"d" : { "status" : "S" } }`,
				`{"d" : { "status" : "S" } }`,
				`{"d" : { "status" : "S" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		err := runAbapEnvironmentCheckoutBranch(&config, nil, &autils, client)
		assert.NoError(t, err, "Did not expect error")
	})
}

func TestTriggerCheckout(t *testing.T) {

	t.Run("Test trigger checkout: success case", func(t *testing.T) {

		// given
		receivedURI := "example.com/Branches"
		uriExpected := receivedURI + "?$expand=to_Execution_log,to_Transport_log"
		tokenExpected := "myToken"

		client := &abaputils.ClientMock{
			Body:       `{"d" : { "__metadata" : { "uri" : "` + receivedURI + `" } } }`,
			Token:      tokenExpected,
			StatusCode: 200,
		}
		config := abapEnvironmentCheckoutBranchOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			RepositoryName:    "testRepo1",
			BranchName:        "feature-unit-test",
		}
		con := abaputils.ConnectionDetailsHTTP{
			User:     "MY_USER",
			Password: "MY_PW",
			URL:      "https://api.endpoint.com/Branches",
		}
		// when
		entityConnection, err := triggerCheckout(config.RepositoryName, config.BranchName, con, client)

		// then
		assert.NoError(t, err)
		assert.Equal(t, uriExpected, entityConnection.URL)
		assert.Equal(t, tokenExpected, entityConnection.XCsrfToken)
	})

	t.Run("Test trigger checkout: ABAP Error case", func(t *testing.T) {

		// given
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
		config := abapEnvironmentCheckoutBranchOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			RepositoryName:    "testRepo1",
			BranchName:        "feature-unit-test",
		}
		con := abaputils.ConnectionDetailsHTTP{
			User:     "MY_USER",
			Password: "MY_PW",
			URL:      "https://api.endpoint.com/Branches",
		}

		// when
		_, err := triggerCheckout(config.RepositoryName, config.BranchName, con, client)

		// then
		assert.Equal(t, combinedErrorMessage, err.Error(), "Different error message expected")
	})
}
func TestCheckoutBranchConfig(t *testing.T) {
	t.Run("Success case: checkout Branches from file config", func(t *testing.T) {
		pollIntervall := abaputils.AUtilsMock{}
		defer pollIntervall.Cleanup()

		receivedURI := "example.com/Branches"
		client := &abaputils.ClientMock{
			Body:       `{"d" : { "__metadata" : { "uri" : "` + receivedURI + `" } } }`,
			Token:      "myToken",
			StatusCode: 200,
		}

		dir, err := ioutil.TempDir("", "test checkout branches")
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
- name: 'testRepo'
  branch: 'testBranch'
- name: 'testRepo2'
  branch: 'testBranch2'
- name: 'testRepo3'
  branch: 'testBranch3'`
		manifestFile2String := `
- name: 'testRepo4'
  branch: 'testBranch4'
- name: 'testRepo5'
  branch: 'testBranch5'
- name: 'testRepo6'
  branch: 'testBranch6'`

		err = ioutil.WriteFile("repositoriesTest.yml", []byte(manifestFileString), 0644)
		err = ioutil.WriteFile("repositoriesTest2.yml", []byte(manifestFile2String), 0644)

		config := abapEnvironmentCheckoutBranchOptions{
			RepositoryNamesFiles: []string{"repositoriesTest.yml", "repositoriesTest2.yml"},
		}
		con := abaputils.ConnectionDetailsHTTP{
			User:     "MY_USER",
			Password: "MY_PW",
			URL:      "https://api.endpoint.com/Branches",
		}
		err = checkoutBranchesFromFileConfig(config.RepositoryNamesFiles, con, client, pollIntervall.GetPollIntervall())
		assert.NoError(t, err)
	})
	t.Run("Failure case: checkout Branches from empty file config", func(t *testing.T) {
		pollIntervall := abaputils.AUtilsMock{}
		defer pollIntervall.Cleanup()

		expectedErrorMessage := "Failed to parse repository configuration file: Empty or wrong configuration file. Please make sure that you have correctly specified the branches in the repositories to be checked out"
		receivedURI := "example.com/Branches"
		client := &abaputils.ClientMock{
			Body:       `{"d" : { "__metadata" : { "uri" : "` + receivedURI + `" } } }`,
			Token:      "myToken",
			StatusCode: 200,
		}

		dir, err := ioutil.TempDir("", "test checkout branches")
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

		config := abapEnvironmentCheckoutBranchOptions{
			RepositoryNamesFiles: []string{"repositoriesTest.yml"},
		}
		con := abaputils.ConnectionDetailsHTTP{
			User:     "MY_USER",
			Password: "MY_PW",
			URL:      "https://api.endpoint.com/Branches",
		}
		err = checkoutBranchesFromFileConfig(config.RepositoryNamesFiles, con, client, pollIntervall.GetPollIntervall())
		assert.Equal(t, expectedErrorMessage, err.Error(), "Different error message expected")
	})
	t.Run("Failure case: checkout Branches from wrong file config", func(t *testing.T) {
		pollIntervall := abaputils.AUtilsMock{}
		defer pollIntervall.Cleanup()

		expectedErrorMessage := "Failed to read repository configuration file: Eror in configuration file, most likely you have entered empty or wrong configuration values. Please make sure that you have correctly specified the branches in the repositories to be checked out"
		receivedURI := "example.com/Branches"
		client := &abaputils.ClientMock{
			Body:       `{"d" : { "__metadata" : { "uri" : "` + receivedURI + `" } } }`,
			Token:      "myToken",
			StatusCode: 200,
		}

		dir, err := ioutil.TempDir("", "test checkout branches")
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

		config := abapEnvironmentCheckoutBranchOptions{
			RepositoryNamesFiles: []string{"repositoriesTest.yml"},
		}
		con := abaputils.ConnectionDetailsHTTP{
			User:     "MY_USER",
			Password: "MY_PW",
			URL:      "https://api.endpoint.com/Branches",
		}
		err = checkoutBranchesFromFileConfig(config.RepositoryNamesFiles, con, client, pollIntervall.GetPollIntervall())
		assert.Equal(t, expectedErrorMessage, err.Error(), "Different error message expected")
	})
	t.Run("Success case: checkout Branch from config", func(t *testing.T) {
		pollIntervall := abaputils.AUtilsMock{}
		defer pollIntervall.Cleanup()

		receivedURI := "example.com/Branches"
		client := &abaputils.ClientMock{
			Body:       `{"d" : { "__metadata" : { "uri" : "` + receivedURI + `" } } }`,
			Token:      "myToken",
			StatusCode: 200,
		}

		config := abapEnvironmentCheckoutBranchOptions{
			RepositoryName: "testRepo1",
			BranchName:     "feature-unit-test",
		}
		con := abaputils.ConnectionDetailsHTTP{
			User:     "MY_USER",
			Password: "MY_PW",
			URL:      "https://api.endpoint.com/Branches",
		}
		err := checkoutBranchFromConfig(&config, con, client, pollIntervall.GetPollIntervall())
		assert.NoError(t, err)
	})
	t.Run("Failure case: checkout Branch with non-existent config", func(t *testing.T) {
		expectedErrorMessage := "Checkout of  for software component  failed on the ABAP System: Failed to trigger checkout: Repository and Branch Configuration is empty. Please make sure that you have specified the correct values"
		pollIntervall := abaputils.AUtilsMock{}
		defer pollIntervall.Cleanup()

		receivedURI := "example.com/Branches"
		client := &abaputils.ClientMock{
			Body:       `{"d" : { "__metadata" : { "uri" : "` + receivedURI + `" } } }`,
			Token:      "myToken",
			StatusCode: 200,
		}

		config := abapEnvironmentCheckoutBranchOptions{}
		con := abaputils.ConnectionDetailsHTTP{
			User:     "MY_USER",
			Password: "MY_PW",
			URL:      "https://api.endpoint.com/Branches",
		}
		err := checkoutBranchFromConfig(&config, con, client, pollIntervall.GetPollIntervall())
		assert.Equal(t, expectedErrorMessage, err.Error(), "Different error message expected")
	})
}
func TestCheckoutConfigChecker(t *testing.T) {
	t.Run("Success case: check config", func(t *testing.T) {
		config := abapEnvironmentCheckoutBranchOptions{
			RepositoryName: "testRepo1",
			BranchName:     "feature-unit-test",
		}
		err := checkCheckoutBranchRepositoryConfiguration(config)
		assert.NoError(t, err)
	})
	t.Run("Success case: check file config", func(t *testing.T) {
		config := abapEnvironmentCheckoutBranchOptions{
			RepositoryNamesFiles: []string{"test.file", "test2.file"},
			BranchName:           "feature-unit-test",
		}
		err := checkCheckoutBranchRepositoryConfiguration(config)
		assert.NoError(t, err)
	})
	t.Run("Failure case: check empty config", func(t *testing.T) {
		expectedErrorMessage := "Checking configuration failed: You have not specified any repository configuration to be pulled into the ABAP Environment System. Please make sure that you specified the repositories with their branches that should be pulled either in a dedicated file or via in-line configuration. For more information please read the User documentation"

		config := abapEnvironmentCheckoutBranchOptions{}
		err := checkCheckoutBranchRepositoryConfiguration(config)
		assert.Equal(t, expectedErrorMessage, err.Error(), "Different error message expected")
	})
}
