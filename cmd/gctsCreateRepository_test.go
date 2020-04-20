package cmd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGctsCreateRepositorySuccess(t *testing.T) {

	config := gctsCreateRepositoryOptions{
		Host:           "testHost.wdf.sap.corp:50000",
		Client:         "000",
		RepositoryName: "testRepo",
		Username:       "testUser",
		Password:       "testPassword",
		GithubURL:      "https://github.com/org/testRepo",
		Role:           "SOURCE",
		VSID:           "TST",
	}

	t.Run("creating repository locally successfull", func(t *testing.T) {

		httpClient := httpMock{StatusCode: 200, ResponseBody: `{
			"repository": {
				"rid": "com.sap.cts.example",
				"name": "Example repository",
				"role": "SOURCE",
				"type": "GIT",
				"vsid": "GI7",
				"status": "READY",
				"branch": "master",
				"url": "https://github.com/git/git",
				"version": "1.0.1",
				"objects": 1337,
				"currentCommit": "f1cdb6a032c1d8187c0990b51e94e8d8bb9898b2",
				"connection": "ssl",
				"config": [
					{
						"key": "CLIENT_VCS_URI",
						"value": "git@github.wdf.sap.corp/example.git"
					}
				]
			},
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

		err := createRepository(&config, nil, nil, &httpClient)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "http://testHost.wdf.sap.corp:50000/sap/bc/cts_abapvcs/repository?sap-client=000", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "POST", httpClient.Method)
			})

			t.Run("check user", func(t *testing.T) {
				assert.Equal(t, "testUser", httpClient.Options.Username)
			})

			t.Run("check password", func(t *testing.T) {
				assert.Equal(t, "testPassword", httpClient.Options.Password)
			})

		}

	})

	t.Run("repository already exists locally", func(t *testing.T) {

		httpClient := httpMock{StatusCode: 500, ResponseBody: `{
			"exception": "Repository already exists"
		}`}

		err := createRepository(&config, nil, nil, &httpClient)

		assert.NoError(t, err)

	})

}
func TestGctsCreateRepositoryFailure(t *testing.T) {

	config := gctsCreateRepositoryOptions{
		Host:           "testHost.wdf.sap.corp:50000",
		Client:         "000",
		RepositoryName: "testRepo",
		Username:       "testUser",
		Password:       "testPassword",
		GithubURL:      "https://github.com/org/testRepo",
		Role:           "SOURCE",
		VSID:           "TST",
	}

	t.Run("creating repository locally failed", func(t *testing.T) {

		httpClient := httpMock{StatusCode: 500, ResponseBody: `{
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

		err := createRepository(&config, nil, nil, &httpClient)

		assert.EqualError(t, err, "creating the repository locally failed: a http error occurred")

	})

}
