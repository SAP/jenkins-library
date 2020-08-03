package abaputils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPollEntity(t *testing.T) {

	t.Run("Test poll entity multiple repos: success case", func(t *testing.T) {

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

		config := AbapEnvironmentPullGitRepoOptions{
			AbapEnvOptions:  options,
			RepositoryNames: []string{"testRepo1", "testRepo2"},
		}

		con := ConnectionDetailsHTTP{
			User:       "MY_USER",
			Password:   "MY_PW",
			URL:        "https://api.endpoint.com/Entity/",
			XCsrfToken: "MY_TOKEN",
		}
		status, _ := PollEntity(config.RepositoryNames[0], con, client, 0)
		assert.Equal(t, "S", status)
	})

	t.Run("Test poll entity multiple repos: error case", func(t *testing.T) {

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

		config := AbapEnvironmentPullGitRepoOptions{
			AbapEnvOptions:  options,
			RepositoryNames: []string{"testRepo1", "testRepo2"},
		}

		con := ConnectionDetailsHTTP{
			User:       "MY_USER",
			Password:   "MY_PW",
			URL:        "https://api.endpoint.com/Entity/",
			XCsrfToken: "MY_TOKEN",
		}
		status, _ := PollEntity(config.RepositoryNames[0], con, client, 0)
		assert.Equal(t, "E", status)
	})

	t.Run("Test poll entity single repo: success case", func(t *testing.T) {

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

	t.Run("Test poll entity single repo: error case", func(t *testing.T) {

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
