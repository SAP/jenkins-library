package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
)

func TestTriggerPull(t *testing.T) {

	t.Run("Test trigger pull: success case", func(t *testing.T) {

		receivedURI := "example.com/Entity"
		uriExpected := receivedURI + "?$expand=to_Execution_log,to_Transport_log"
		tokenExpected := "myToken"

		client := &abaputils.ClientMock{
			Body:       `{"d" : { "__metadata" : { "uri" : "` + receivedURI + `" } } }`,
			Token:      tokenExpected,
			StatusCode: 200,
		}
		config := abapEnvironmentPullGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			RepositoryNames:   []string{"testRepo1", "testRepo2"},
		}

		con := abaputils.ConnectionDetailsHTTP{
			User:     "MY_USER",
			Password: "MY_PW",
			URL:      "https://api.endpoint.com/Entity/",
		}
		entityConnection, err := triggerPull(config.RepositoryNames[0], con, client)
		assert.Nil(t, err)
		assert.Equal(t, uriExpected, entityConnection.URL)
		assert.Equal(t, tokenExpected, entityConnection.XCsrfToken)
	})

	t.Run("Test trigger pull: ABAP Error", func(t *testing.T) {

		errorMessage := "ABAP Error Message"
		errorCode := "ERROR/001"
		HTTPErrorMessage := "HTTP Error Message"
		combinedErrorMessage := "HTTP Error Message: ERROR/001 - ABAP Error Message"

		client := &abaputils.ClientMock{
			Body:       `{"error" : { "code" : "` + errorCode + `", "message" : { "lang" : "en", "value" : "` + errorMessage + `" } } }`,
			Token:      "myToken",
			StatusCode: 400,
			Error:      errors.New(HTTPErrorMessage),
		}
		config := abapEnvironmentPullGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			RepositoryNames:   []string{"testRepo1", "testRepo2"},
		}

		con := abaputils.ConnectionDetailsHTTP{
			User:     "MY_USER",
			Password: "MY_PW",
			URL:      "https://api.endpoint.com/Entity/",
		}
		_, err := triggerPull(config.RepositoryNames[0], con, client)
		assert.Equal(t, combinedErrorMessage, err.Error(), "Different error message expected")
	})

}

/* func TestGetAbapCommunicationArrangementInfo(t *testing.T) {

	t.Run("Test cf cli command: success case", func(t *testing.T) {

		config := abaputils.AbapEnvironmentOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
		}

		options := abaputils.AbapEnvironmentPullGitRepoOptions{
			AbapEnvOptions: config,
		}

		execRunner := &mock.ExecMockRunner{}

		abaputils.GetAbapCommunicationArrangementInfo(options.AbapEnvOptions, execRunner, "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull")
		assert.Equal(t, "cf", execRunner.Calls[0].Exec, "Wrong command")
		assert.Equal(t, []string{"login", "-a", "https://api.endpoint.com", "-o", "testOrg", "-s", "testSpace", "-u", "testUser", "-p", "testPassword"}, execRunner.Calls[0].Params, "Wrong parameters")
		//assert.Equal(t, []string{"api", "https://api.endpoint.com"}, execRunner.Calls[0].Params, "Wrong parameters")

	})

	t.Run("Test cf cli command: params missing", func(t *testing.T) {

		config := abaputils.AbapEnvironmentOptions{
			//CfServiceKeyName:  "testServiceKey", this parameter will be missing
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			Username:          "testUser",
			Password:          "testPassword",
		}

		options := abaputils.AbapEnvironmentPullGitRepoOptions{
			AbapEnvOptions: config,
		}

		execRunner := &mock.ExecMockRunner{}

		var _, err = abaputils.GetAbapCommunicationArrangementInfo(options.AbapEnvOptions, execRunner, "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull")
		assert.Equal(t, "Parameters missing. Please provide EITHER the Host of the ABAP server OR the Cloud Foundry ApiEndpoint, Organization, Space, Service Instance and a corresponding Service Key for the Communication Scenario SAP_COM_0510", err.Error(), "Different error message expected")
	})

	t.Run("Test cf cli command: params missing", func(t *testing.T) {

		config := abaputils.AbapEnvironmentOptions{
			Username: "testUser",
			Password: "testPassword",
		}

		options := abaputils.AbapEnvironmentPullGitRepoOptions{
			AbapEnvOptions: config,
		}

		execRunner := &mock.ExecMockRunner{}

		var _, err = abaputils.GetAbapCommunicationArrangementInfo(options.AbapEnvOptions, execRunner, "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull")
		assert.Equal(t, "Parameters missing. Please provide EITHER the Host of the ABAP server OR the Cloud Foundry ApiEndpoint, Organization, Space, Service Instance and a corresponding Service Key for the Communication Scenario SAP_COM_0510", err.Error(), "Different error message expected")
	})

} */
