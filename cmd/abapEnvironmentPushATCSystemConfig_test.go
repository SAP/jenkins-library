package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/stretchr/testify/assert"
)

func TestRunAbapEnvironmentPushATCSystemConfig(t *testing.T) {
	t.Parallel()

	t.Run("run Step Successful", func(t *testing.T) {
		t.Parallel()
		// init
		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		config := abapEnvironmentPushATCSystemConfigOptions{
			AtcSystemConfigFilePath:   "test.json",
			PatchExistingSystemConfig: true,
			CfAPIEndpoint:             "https://api.endpoint.com",
			CfOrg:                     "testOrg",
			CfSpace:                   "testSpace",
			CfServiceInstance:         "testInstance",
			CfServiceKeyName:          "testServiceKey",
			Username:                  "testUser",
			Password:                  "testPassword",
			Host:                      "testHost",
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
		// test
		err := runAbapEnvironmentPushATCSystemConfig(&config, nil, &autils, client)
		// assert
		assert.NoError(t, err)
	})

	t.Run("run Step Failure", func(t *testing.T) {
		t.Parallel()

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
