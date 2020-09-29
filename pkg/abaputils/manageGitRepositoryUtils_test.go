package abaputils

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPollEntity(t *testing.T) {

	t.Run("Test poll entity - success case", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{"d" : { "status" : "S" } }`,
				`{"d" : { "status" : "R" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		options := AbapEnvironmentOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
		}

		config := AbapEnvironmentCheckoutBranchOptions{
			AbapEnvOptions: options,
			RepositoryName: "testRepo1",
		}

		con := ConnectionDetailsHTTP{
			User:       "MY_USER",
			Password:   "MY_PW",
			URL:        "https://api.endpoint.com/Entity/",
			XCsrfToken: "MY_TOKEN",
		}
		status, _ := PollEntity(config.RepositoryName, con, client, 0)
		assert.Equal(t, "S", status)
	})

	t.Run("Test poll entity - error case", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{"d" : { "status" : "E" } }`,
				`{"d" : { "status" : "R" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		options := AbapEnvironmentOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
		}

		config := AbapEnvironmentCheckoutBranchOptions{
			AbapEnvOptions: options,
			RepositoryName: "testRepo1",
		}

		con := ConnectionDetailsHTTP{
			User:       "MY_USER",
			Password:   "MY_PW",
			URL:        "https://api.endpoint.com/Entity/",
			XCsrfToken: "MY_TOKEN",
		}
		status, _ := PollEntity(config.RepositoryName, con, client, 0)
		assert.Equal(t, "E", status)
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

		dir, err := ioutil.TempDir("", "test abap utils")
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

		config := RepositoriesConfig{
			BranchName:      "testBranch",
			RepositoryName:  "testRepository",
			RepositoryNames: []string{"testRepository"},
			Repositories:    "repositoriesTest.yml",
		}

		repositories, err := GetRepositories(&config)

		assert.Equal(t, expectedRepositoryList, repositories)
		assert.NoError(t, err)
	})
	t.Run("Get Repositories from file config - failure case", func(t *testing.T) {
		expectedRepositoryList := []Repository([]Repository{})
		expectedErrorMessage := "Could not parse config file repositoriesTest.yml, AddonDescriptor doesn't contain any repositories"

		dir, err := ioutil.TempDir("", "test abap utils")
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
- repo: 'testRepo'
- repo: 'testRepo2'`

		err = ioutil.WriteFile("repositoriesTest.yml", []byte(manifestFileString), 0644)

		config := RepositoriesConfig{
			Repositories: "repositoriesTest.yml",
		}

		repositories, err := GetRepositories(&config)

		assert.Equal(t, expectedRepositoryList, repositories)
		assert.EqualError(t, err, expectedErrorMessage)
	})
	t.Run("Get Repositories from config - failure case", func(t *testing.T) {
		expectedRepositoryList := []Repository([]Repository{})
		expectedErrorMessage := "Could not parse config file repositoriesTest.yml, AddonDescriptor doesn't contain any repositories"

		dir, err := ioutil.TempDir("", "test  abap utils")
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
- repo: 'testRepo'
- repo: 'testRepo2'`

		err = ioutil.WriteFile("repositoriesTest.yml", []byte(manifestFileString), 0644)

		config := RepositoriesConfig{
			Repositories: "repositoriesTest.yml",
		}

		repositories, err := GetRepositories(&config)

		assert.Equal(t, expectedRepositoryList, repositories)
		assert.EqualError(t, err, expectedErrorMessage)
	})
	t.Run("Get Repositories from empty config - failure case", func(t *testing.T) {
		expectedRepositoryList := []Repository([]Repository{})
		expectedErrorMessage := "Failed to read repository configuration: You have not specified any repository configuration to be pulled into the ABAP Environment System. Please make sure that you specified the repositories with their branches that should be pulled either in a dedicated file or via in-line configuration. For more information please read the User documentation"

		config := RepositoriesConfig{}

		repositories, err := GetRepositories(&config)

		assert.Equal(t, expectedRepositoryList, repositories)
		assert.EqualError(t, err, expectedErrorMessage)
	})
}
