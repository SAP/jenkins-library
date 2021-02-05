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
	t.Run("Run Step Successful - repositoryName and branchName config", func(t *testing.T) {

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
	t.Run("Run Step Failure - empty config", func(t *testing.T) {
		expectedErrorMessage := "Something failed during the checkout: Checking configuration failed: You have not specified any repository or branch configuration to be checked out in the ABAP Environment System. Please make sure that you specified the repositories with their branches that should be checked out either in a dedicated file or via the parameters 'repositoryName' and 'branchName'. For more information please read the User documentation"

		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		config := abapEnvironmentCheckoutBranchOptions{}

		client := &abaputils.ClientMock{
			BodyList: []string{
				`{"d" : { "status" : "E" } }`,
				`{"d" : { "status" : "E" } }`,
				`{"d" : { "status" : "E" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		err := runAbapEnvironmentCheckoutBranch(&config, nil, &autils, client)
		assert.EqualError(t, err, expectedErrorMessage)
	})
	t.Run("Run Step Failure - wrong status", func(t *testing.T) {
		expectedErrorMessage := "Something failed during the checkout: Checkout failed: Checkout of branch testBranch failed on the ABAP System"

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
				`{"d" : { "status" : "E" } }`,
				`{"d" : { "status" : "E" } }`,
				`{"d" : { "status" : "E" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		err := runAbapEnvironmentCheckoutBranch(&config, nil, &autils, client)
		assert.EqualError(t, err, expectedErrorMessage)
	})
	t.Run("Success case: checkout Branches from file config", func(t *testing.T) {
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
repositories:
- name: 'testRepo'
  branch: 'testBranch'
- name: 'testRepo2'
  branch: 'testBranch2'
- name: 'testRepo3'
  branch: 'testBranch3'`

		err = ioutil.WriteFile("repositoriesTest.yml", []byte(manifestFileString), 0644)

		config := abapEnvironmentCheckoutBranchOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			Repositories:      "repositoriesTest.yml",
		}
		err = runAbapEnvironmentCheckoutBranch(&config, nil, &autils, client)
		assert.NoError(t, err)
	})
	t.Run("Failure case: checkout Branches from empty file config", func(t *testing.T) {
		expectedErrorMessage := "Something failed during the checkout: Error in config file repositoriesTest.yml, AddonDescriptor doesn't contain any repositories"

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
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			Repositories:      "repositoriesTest.yml",
		}
		err = runAbapEnvironmentCheckoutBranch(&config, nil, &autils, client)
		assert.EqualError(t, err, expectedErrorMessage)
	})
	t.Run("Failure case: checkout Branches from wrong file config", func(t *testing.T) {
		expectedErrorMessage := "Something failed during the checkout: Could not unmarshal repositoriesTest.yml"

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
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			Repositories:      "repositoriesTest.yml",
		}
		err = runAbapEnvironmentCheckoutBranch(&config, nil, &autils, client)
		assert.EqualError(t, err, expectedErrorMessage)
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
			Repositories: "test.file",
			BranchName:   "feature-unit-test",
		}
		err := checkCheckoutBranchRepositoryConfiguration(config)
		assert.NoError(t, err)
	})
	t.Run("Failure case: check empty config", func(t *testing.T) {
		expectedErrorMessage := "Checking configuration failed: You have not specified any repository or branch configuration to be checked out in the ABAP Environment System. Please make sure that you specified the repositories with their branches that should be checked out either in a dedicated file or via the parameters 'repositoryName' and 'branchName'. For more information please read the User documentation"

		config := abapEnvironmentCheckoutBranchOptions{}
		err := checkCheckoutBranchRepositoryConfiguration(config)
		assert.Equal(t, expectedErrorMessage, err.Error(), "Different error message expected")
	})
}
