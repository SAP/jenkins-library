package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGctsDeploySuccess(t *testing.T) {

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
				assert.Equal(t, "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo/pullByCommit?sap-client=000", httpClient.URL)
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

func TestGctsDeployFailure(t *testing.T) {

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
