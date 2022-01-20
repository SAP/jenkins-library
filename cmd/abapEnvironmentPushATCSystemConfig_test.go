package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/stretchr/testify/assert"
)

func TestFetchXcsrfTokenFromHead(t *testing.T) {
	t.Parallel()
	t.Run("FetchXcsrfToken Test", func(t *testing.T) {
		tokenExpected := "myToken"

		client := &abaputils.ClientMock{
			Body:  `Xcsrf Token test`,
			Token: tokenExpected,
		}

		con := abaputils.ConnectionDetailsHTTP{
			User:     "Test",
			Password: "Test",
			URL:      "https://api.endpoint.com/Entity/",
		}
		token, error := fetchXcsrfTokenFromHead(con, client)
		if error == nil {
			assert.Equal(t, tokenExpected, token)
		}
	})
	t.Run("failure case: fetch token", func(t *testing.T) {
		tokenExpected := ""

		client := &abaputils.ClientMock{
			Body:  `Xcsrf Token test`,
			Token: "",
		}

		con := abaputils.ConnectionDetailsHTTP{
			User:     "Test",
			Password: "Test",
			URL:      "https://api.endpoint.com/Entity/",
		}
		token, error := fetchXcsrfTokenFromHead(con, client)
		if error == nil {
			assert.Equal(t, tokenExpected, token)
		}
	})
}

func TestRunAbapEnvironmentPushATCSystemConfig(t *testing.T) {
	t.Parallel()

	t.Run("run Step Failure - File not exist", func(t *testing.T) {
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

		config := abapEnvironmentPushATCSystemConfigOptions{AtcSystemConfigFilePath: "test.json"}

		expectedErrorMessage := "pushing ATC System Configuration failed. Reason: Configured File does not exist(File: " + config.AtcSystemConfigFilePath + ")"

		err := runAbapEnvironmentPushATCSystemConfig(&config, nil, &autils, client)
		assert.Equal(t, expectedErrorMessage, err.Error(), "Different error message expected")
	})
}
