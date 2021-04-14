package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGctsPullByCommitSuccess(t *testing.T) {

	config := gctsDeployOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Username:   "testUser",
		Password:   "testPassword",
	}

	t.Run("deploy latest commit", func(t *testing.T) {

		httpClient := httpMockGcts{StatusCode: 200, ResponseBody: `{
			"trkorr": "SIDK1234567",
			"fromCommit": "f1cdb6a032c1d8187c0990b51e94e8d8bb9898b2",
			"toCommit": "f1cdb6a032c1d8187c0990b51e94e8d8bb9898b2",
			"log": [
				{
					"time": 20180606130524,
					"user": "JENKINS",
					"section": "REPOSITORY_FACTORY",
					"action": "CREATE_REPOSITORY",
					"severity": "INFO",
					"message": "Start action CREATE_REPOSITORY review",
					"code": "GCTS.API.410"
				}
			]
		}`}

		err := pullByCommit(&config, nil, nil, &httpClient)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo/pullByCommit?sap-client=000&request=", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})

			t.Run("check user", func(t *testing.T) {
				assert.Equal(t, "testUser", httpClient.Options.Username)
			})

			t.Run("check password", func(t *testing.T) {
				assert.Equal(t, "testPassword", httpClient.Options.Password)
			})

		}

	})
}

func TestGctsPullByCommitFailure(t *testing.T) {

	config := gctsDeployOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Username:   "testUser",
		Password:   "testPassword",
	}

	t.Run("http error occurred", func(t *testing.T) {

		httpClient := httpMockGcts{StatusCode: 500, ResponseBody: `{
			"log": [
				{
					"time": 20180606130524,
					"user": "JENKINS",
					"section": "REPOSITORY_FACTORY",
					"action": "CREATE_REPOSITORY",
					"severity": "INFO",
					"message": "Start action CREATE_REPOSITORY review",
					"code": "GCTS.API.410"
				}
			],
			"errorLog": [
				{
					"time": 20180606130524,
					"user": "JENKINS",
					"section": "REPOSITORY_FACTORY",
					"action": "CREATE_REPOSITORY",
					"severity": "INFO",
					"message": "Start action CREATE_REPOSITORY review",
					"code": "GCTS.API.410"
				}
			],
			"exception": {
				"message": "repository_not_found",
				"description": "Repository not found",
				"code": 404
			}
		}`}

		err := pullByCommit(&config, nil, nil, &httpClient)

		assert.EqualError(t, err, "a http error occurred")

	})

}

func TestGctsGetRepositorySuccess(t *testing.T) {
	config := gctsDeployOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Username:   "testUser",
		Password:   "testPassword",
	}
	t.Run("Get Repository Success Test", func(t *testing.T) {
		var httpClient httpMockGcts
		if config.Repository == "testRepo" {
			httpClient = httpMockGcts{StatusCode: 200, ResponseBody: `{
				"result": {
				    "rid": "testrepo",
				    "name": "testRepo",
				    "role": "SOURCE",
				    "type": "GIT",
				    "vsid": "GIT",
				    "status": "READY",
				    "branch": "dummy_branch",
				    "url": "https://example.git.com/testRepo",
				    "createdBy": "testUser",
				    "createdDate": "dummy_date",
				    "config": [
					{
					    "key": "CURRENT_COMMIT",
					    "value": "dummy_commit_number",
					    "category": "GENERAL",
					    "scope": "local"
					}
				    ],
				    "objects": 1,
				    "currentCommit": "dummy_commit_number",
				    "connection": "ssl"
				}
			    }`}
		}

		repository, err := getRepository(&config, &httpClient)

		if assert.NoError(t, err) {
			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "https://example.git.com/testRepo", repository.Result.Url)
			})
			t.Run("check rid", func(t *testing.T) {
				assert.Equal(t, "testrepo", repository.Result.Rid)
			})
			t.Run("check commit id", func(t *testing.T) {
				assert.Equal(t, "dummy_commit_number", repository.Result.CurrentCommit)
			})
		}
	})
}

func TestGctsGetRepositoryFailure(t *testing.T) {
	config := gctsDeployOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepoNotExists",
		Username:   "testUser",
		Password:   "testPassword",
	}
	t.Run("Get Repository Success Test", func(t *testing.T) {
		var httpClient httpMockGcts
		if config.Repository == "testRepoNotExists" {
			httpClient = httpMockGcts{StatusCode: 500, ResponseBody: `{
				"exception": "No relation between system and repository"
			    }`}
		}

		_, err := getRepository(&config, &httpClient)

		assert.EqualError(t, err, "a http error occurred")
	})

}

func TestGctsSwitchBranchSuccess(t *testing.T) {
	config := gctsDeployOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Branch:     "dummyBranch",
		Username:   "testUser",
		Password:   "testPassword",
	}

	t.Run("Switch Branch success", func(t *testing.T) {
		var httpClient httpMockGcts
		if config.Branch == "dummyBranch" {
			httpClient = httpMockGcts{StatusCode: 200, ResponseBody: `{
				"result": {
				    "rid": "testrepo",
				    "checkoutTime": 20210413082242,
				    "fromCommit": "from_dummy_commit",
				    "toCommit": "to_dummy_commit",
				    "caller": "testUser",
				    "request": "GITKUKDUMMY",
				    "type": "BRANCH_SW",
				    "state": "DONE",
				    "rc": "0000"
				}
			    }`}
		}

		responseBody, err := switchBranch(&config, &httpClient, "dummyCurrentBranch", "dummyTargetBranch")

		if assert.NoError(t, err) {
			t.Run("check from commit", func(t *testing.T) {
				assert.Equal(t, "from_dummy_commit", responseBody.Result.FromCommit)
			})
			t.Run("check to commit", func(t *testing.T) {
				assert.Equal(t, "to_dummy_commit", responseBody.Result.ToCommit)
			})
		}
	})
}

func TestGctsSwitchBranchFailure(t *testing.T) {
	config := gctsDeployOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Branch:     "dummyBranchNotExists",
		Username:   "testUser",
		Password:   "testPassword",
	}
	t.Run("Switch Branch failure Test", func(t *testing.T) {
		var httpClient httpMockGcts
		if config.Branch == "dummyBranchNotExists" {
			httpClient = httpMockGcts{StatusCode: 500, ResponseBody: `{
				"errorLog": [
				    {
					"severity": "ERROR",
					"message": "The branch to switch to - 'feature1' - does not exist",
					"code": "GCTS.CLIENT.1320"
				    }
				],
				"log": [
				    {
					"time": 20210414102742,
					"user": "testUser",
					"section": "REPOSITORY",
					"action": "SWITCH_BRANCH",
					"severity": "ERROR",
					"message": "20210414102742: Error action SWITCH_BRANCH 20210414_102740_B4EC329722B5C611B35B345F3B5F8FAA"
				    },
				    {
					"time": 20210414102742,
					"user": "testUser",
					"section": "REPOSITORY",
					"action": "SWITCH_BRANCH",
					"severity": "ERROR",
					"message": "20210414102742: Error action SWITCH_BRANCH Client error"
				    }
				],
				"exception": "Cannot switch branch of local repository to selected branch."
			    }`}
		}

		_, err := getRepository(&config, &httpClient)

		assert.EqualError(t, err, "a http error occurred")
	})

}

func TestCreateRepositorySuccess(t *testing.T) {
	config := gctsCreateRepositoryOptions{
		Host:                "http://testHost.com:50000",
		Client:              "000",
		Repository:          "testRepo",
		Username:            "testUser",
		Password:            "testPassword",
		RemoteRepositoryURL: "http://testRepoUrl.com",
		Role:                "dummyRole",
		VSID:                "dummyVsid",
		Type:                "dummyType",
	}
	t.Run("Create Repository Success", func(t *testing.T) {
		var httpClient httpMockGcts
		if config.Repository == "testRepo" {
			httpClient = httpMockGcts{StatusCode: 200, ResponseBody: `{
				"repository": {
				    "rid": "testrepo",
				    "name": "testRepo",
				    "role": "dummyRole",
				    "type": "dummyType",
				    "vsid": "dummyVsid",
				    "status": "CREATED",
				    "branch": "dummyBranch",
				    "url": "http://testRepoUrl.com",
				    "createdBy": "testUser",
				    "createdDate": "2021-04-14",
				    "config": [
					{
					    "key": "CLIENT_VCS_CONNTYPE",
					    "value": "ssl",
					    "category": "CONNECTION",
					    "scope": "local"
					},
					{
					    "key": "CLIENT_VCS_URI",
					    "value": "http://testRepoUrl.com",
					    "category": "CONNECTION",
					    "scope": "local"
					}
				    ],
				    "connection": "ssl"
				}
			    }`}
		}

		err := createRepositoryForDeploy(&config, nil, nil, &httpClient, nil)
		assert.NoError(t, err)
	})
}

func TestCreateRepositoryFailure(t *testing.T) {
	config := gctsCreateRepositoryOptions{
		Host:                "http://testHost.com:50000",
		Client:              "000",
		Repository:          "testRepoExists",
		Username:            "testUser",
		Password:            "testPassword",
		RemoteRepositoryURL: "http://testRepoUrlFail.com",
		Role:                "dummyRole",
		VSID:                "dummyVsid",
		Type:                "dummyType",
	}
	t.Run("Create Repository Failure", func(t *testing.T) {
		var httpClient httpMockGcts
		if config.Repository == "testRepoExists" {
			httpClient = httpMockGcts{StatusCode: 500, ResponseBody: `{
				"errorLog": [
				    {
					"time": 20210414105442,
					"user": "testUser",
					"section": "SYSTEM",
					"action": "CREATE_REPOSITORY",
					"severity": "ERROR",
					"message": "20210414105442: Error action CREATE_REPOSITORY Repository already exists"
				    },
				    {
					"severity": "ERROR",
					"message": "Could not create repository 'testRepoExists'.",
					"code": "GCTS.API.703"
				    }
				],
				"log": [
				    {
					"time": 20210414105442,
					"user": "testUser",
					"section": "SYSTEM",
					"action": "CREATE_REPOSITORY",
					"severity": "INFO",
					"message": "20210414105442: Start action CREATE_REPOSITORY"
				    },
				    {
					"time": 20210414105442,
					"user": "testUser",
					"section": "REPOSITORY_FACTORY",
					"action": "CREATE_REPOSITORY",
					"severity": "INFO",
					"message": "20210414105442: Start action CREATE_REPOSITORY testrepo"
				    },
				    {
					"time": 20210414105442,
					"user": "testUser",
					"section": "SYSTEM",
					"action": "CREATE_REPOSITORY",
					"severity": "ERROR",
					"message": "20210414105442: Error action CREATE_REPOSITORY Repository already exists"
				    }
				],
				"exception": "Repository already exists"
			    }`}
		}

		err := createRepositoryForDeploy(&config, nil, nil, &httpClient, nil)
		assert.EqualError(t, err, "a http error occurred")
	})
}

func TestGctsSetConfigByKeySuccess(t *testing.T) {
	config := gctsDeployOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Branch:     "dummyBranch",
		Username:   "testUser",
		Password:   "testPassword",
	}
	configKey := setConfigKeyBody{
		Key:   "dummy_key",
		Value: "dummy_value",
	}
	t.Run("Set Config By key Success", func(t *testing.T) {
		var httpClient httpMockGcts
		if config.Repository == "testRepo" {
			httpClient = httpMockGcts{StatusCode: 200, ResponseBody: `{
				"result": {
				    "key": "dummy_key",
				    "value": "dummy_value"
				}
			    }`}
		}

		err := setConfigKey(&config, &httpClient, &configKey)

		assert.NoError(t, err)
	})

}

func TestGctsSetConfigByKeyFailure(t *testing.T) {
	config := gctsDeployOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepoNotExists",
		Branch:     "dummyBranchNotExists",
		Username:   "testUser",
		Password:   "testPassword",
	}
	configKey := setConfigKeyBody{
		Key:   "dummy_key",
		Value: "dummy_value",
	}
	t.Run("Set Config By key Success", func(t *testing.T) {
		var httpClient httpMockGcts
		if config.Repository == "testRepoNotExists" {
			httpClient = httpMockGcts{StatusCode: 500, ResponseBody: `{
				"exception": "No relation between system and repository"
			    }`}
		}

		err := setConfigKey(&config, &httpClient, &configKey)

		assert.EqualError(t, err, "a http error occurred")
	})

}
