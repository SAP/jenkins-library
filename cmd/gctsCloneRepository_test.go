//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGctsCloneRepositorySuccess(t *testing.T) {

	config := gctsCloneRepositoryOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Username:   "testUser",
		Password:   "testPassword",
	}

	t.Run("cloning successful", func(t *testing.T) {

		httpClient := httpMockGcts{StatusCode: 200, ResponseBody: `{
			"result": {
				"rid": "my-repository",
				"checkoutTime": 20180606130524,
				"fromCommit": "f1cdb6a032c1d8187c0990b51e94e8d8bb9898b2",
				"toCommit": "f1cdb6a032c1d8187c0990b51e94e8d8bb9898b2",
				"caller": "JOHNDOE",
				"request": "SIDK1234567",
				"type": "PULL"
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

		err := cloneRepository(&config, nil, &httpClient)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo/clone?sap-client=000", httpClient.URL)
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

	t.Run("repository has already been cloned", func(t *testing.T) {

		httpClient := httpMockGcts{StatusCode: 500, ResponseBody: `{
			"errorLog": [
				{
					"time": 20200414112900,
					"user": "USER",
					"section": "CLIENT_TP",
					"action": "GET_SOURCE_FROM_REMOTE",
					"severity": "ERROR",
					"message": "20200414112900: Error action GET_SOURCE_FROM_REMOTE",
					"protocol": [
						{
							"type": "Paramters",
							"protocol": [
								{
									"repouri": "https://github.com/test-repo",
									"logfile": "/usr/sap/SID/D00/gcts/repos/gcts/repo-name/log/20200414_112900_AD0F43952A5A3F47133637329BA760D8.log",
									"repodir": "/usr/sap/SID/D00/gcts/repos/gcts/repo-name/data/",
									"loglevel": "warning",
									"command": "clone"
								}
							]
						},
						{
							"type": "Client Log",
							"protocol": [
								"protocol logs"
							]
						},
						{
							"type": "Client Stack Log",
							"protocol": [
								{
									"client stack log key": "client stack log value",
									"stack": [
										"java",
										"stack"
									]
								}
							]
						}
					]
				},
				{
					"severity": "ERROR",
					"message": "Cloning a repository into a new working directory failed: Destination path 'data' already exists and is not an empty directory",
					"code": "GCTS.CLIENT.1420"
				},
				{
					"time": 20200414112900,
					"user": "USER",
					"section": "REPOSITORY",
					"action": "CLONE",
					"severity": "ERROR",
					"message": "20200414112900: Error action CLONE 20200414_112900_AD0F43952A5A3F47133637329BA760D8"
				}
			]
		}`}

		err := cloneRepository(&config, nil, &httpClient)
		assert.NoError(t, err)

	})

}

func TestGctsCloneRepositoryFailure(t *testing.T) {

	config := gctsCloneRepositoryOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Username:   "testUser",
		Password:   "testPassword",
	}

	t.Run("a http error occurred", func(t *testing.T) {
		httpClient := httpMockGcts{StatusCode: 500, ResponseBody: `{
			"errorLog": [
				{
					"time": 20200414112900,
					"user": "USER",
					"section": "CLIENT_TP",
					"action": "GET_SOURCE_FROM_REMOTE",
					"severity": "ERROR",
					"message": "20200414112900: Error action GET_SOURCE_FROM_REMOTE",
					"protocol": [
						{
							"type": "Paramters",
							"protocol": [
								{
									"repouri": "https://github.com/test-repo",
									"logfile": "/usr/sap/SID/D00/gcts/repos/gcts/repo-name/log/20200414_112900_AD0F43952A5A3F47133637329BA760D8.log",
									"repodir": "/usr/sap/SID/D00/gcts/repos/gcts/repo-name/data/",
									"loglevel": "warning",
									"command": "clone"
								}
							]
						},
						{
							"type": "Client Log",
							"protocol": [
								"protocol logs"
							]
						},
						{
							"type": "Client Stack Log",
							"protocol": [
								{
									"client stack log key": "client stack log value",
									"stack": [
										"java",
										"stack"
									]
								}
							]
						}
					]
				},
				{
					"severity": "ERROR",
					"message": "Cloning a repository into a new working directory failed: Destination path 'data' already exists and is not an empty directory",
					"code": "GCTS.CLIENT.9999"
				},
				{
					"time": 20200414112900,
					"user": "USER",
					"section": "REPOSITORY",
					"action": "CLONE",
					"severity": "ERROR",
					"message": "20200414112900: Error action CLONE 20200414_112900_AD0F43952A5A3F47133637329BA760D8"
				}
			]
		}`}

		err := cloneRepository(&config, nil, &httpClient)
		assert.EqualError(t, err, "cloning the repository failed: a http error occurred")

	})
}
