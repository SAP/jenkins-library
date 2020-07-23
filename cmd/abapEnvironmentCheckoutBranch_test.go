package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestTriggerCheckout(t *testing.T) {

	t.Run("Test trigger checkout: success case", func(t *testing.T) {

		// given
		receivedURI := "example.com/Branches"
		uriExpected := receivedURI + "?$expand=to_Execution_log,to_Transport_log"
		tokenExpected := "myToken"

		client := &clientMock{
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

		client := &clientMock{
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

func TestPollEntityCheckoutStep(t *testing.T) {

	t.Run("Test poll entity: success case", func(t *testing.T) {

		client := &clientMock{
			BodyList: []string{
				`{"d" : { "status" : "S" } }`,
				`{"d" : { "status" : "R" } }`,
			},
			Token:      "myToken",
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
		}

		con := abaputils.ConnectionDetailsHTTP{
			User:       "MY_USER",
			Password:   "MY_PW",
			URL:        "https://api.endpoint.com/Entity/",
			XCsrfToken: "MY_TOKEN",
		}
		status, _ := pollEntity(config.RepositoryName, con, client, 0)
		assert.Equal(t, "S", status)
	})

	t.Run("Test poll entity: error case", func(t *testing.T) {

		client := &clientMock{
			BodyList: []string{
				`{"d" : { "status" : "E" } }`,
				`{"d" : { "status" : "R" } }`,
			},
			Token:      "myToken",
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
		}

		con := abaputils.ConnectionDetailsHTTP{
			User:       "MY_USER",
			Password:   "MY_PW",
			URL:        "https://api.endpoint.com/Entity/",
			XCsrfToken: "MY_TOKEN",
		}
		status, _ := pollEntity(config.RepositoryName, con, client, 0)
		assert.Equal(t, "E", status)
	})

}
