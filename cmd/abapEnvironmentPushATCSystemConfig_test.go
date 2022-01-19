package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/stretchr/testify/assert"
)

func TestRunAbapEnvironmentPushATCSystemConfig(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
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

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := abapEnvironmentPushATCSystemConfigOptions{}

		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()

		// test
		err := runAbapEnvironmentPushATCSystemConfig(&config, nil, &autils, nil)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
