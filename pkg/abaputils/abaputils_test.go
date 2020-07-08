package abaputils

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/stretchr/testify/assert"
)

func TestReadServiceKeyAbapEnvironment(t *testing.T) {
	t.Run("ReadServiceKeyAbapEnvironment - Failed to login to Cloud Foundry", func(t *testing.T) {

		//given
		options := AbapEnvironmentOptions{
			Username:          "testUser",
			Password:          "testPassword",
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfSpace:           "testSpace",
			CfOrg:             "testOrg",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testKey",
		}

		//when
		var abapKey AbapServiceKey
		var err error
		abapKey, err = ReadServiceKeyAbapEnvironment(options, &command.Command{}, true)

		//then
		assert.Equal(t, "", abapKey.Abap.Password)
		assert.Equal(t, "", abapKey.Abap.Username)
		assert.Equal(t, "", abapKey.Abap.CommunicationArrangementID)
		assert.Equal(t, "", abapKey.Abap.CommunicationScenarioID)
		assert.Equal(t, "", abapKey.Abap.CommunicationSystemID)

		assert.Equal(t, "", abapKey.Binding.Env)
		assert.Equal(t, "", abapKey.Binding.Type)
		assert.Equal(t, "", abapKey.Binding.ID)
		assert.Equal(t, "", abapKey.Binding.Version)
		assert.Equal(t, "", abapKey.SystemID)
		assert.Equal(t, "", abapKey.URL)

		assert.Error(t, err)
	})

	// t.Run("ReadServiceKeyAbapEnvironment - Success", func(t *testing.T) {

	// 	//given

	// 	// todo: mock cf cli

	// 	options := AbapEnvironmentOptions{
	// 		Username:          "testUser",
	// 		Password:          "testPassword",
	// 		CfAPIEndpoint:     "https://api.endpoint.com",
	// 		CfSpace:           "testSpace",
	// 		CfOrg:             "testOrg",
	// 		CfServiceInstance: "testInstance",
	// 		CfServiceKeyName:  "testKey",
	// 	}

	// 	//when
	// 	var abapKey AbapServiceKey
	// 	var err error
	// 	abapKey, err = ReadServiceKeyAbapEnvironment(options, &command.Command{}, true)

	// 	//then
	// 	assert.Equal(t, "", abapKey.Abap.Password)
	// 	assert.Equal(t, "", abapKey.Abap.Username)
	// 	assert.Equal(t, "", abapKey.Abap.CommunicationArrangementID)
	// 	assert.Equal(t, "", abapKey.Abap.CommunicationScenarioID)
	// 	assert.Equal(t, "", abapKey.Abap.CommunicationSystemID)

	// 	assert.Equal(t, "", abapKey.Binding.Env)
	// 	assert.Equal(t, "", abapKey.Binding.Type)
	// 	assert.Equal(t, "", abapKey.Binding.ID)
	// 	assert.Equal(t, "", abapKey.Binding.Version)
	// 	assert.Equal(t, "", abapKey.SystemID)
	// 	assert.Equal(t, "", abapKey.URL)

	// 	assert.NoError(t, err)
	// })
}

func TestGetAbapCommunicationInfo(t *testing.T) {
	t.Run("GetAbapCommunicationArrangementInfo - Error reading service Key", func(t *testing.T) {

		//given
		options := AbapEnvironmentOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfSpace:           "testSpace",
			CfOrg:             "testOrg",
			CfServiceInstance: "testInstance",
			Username:          "testUser",
			Password:          "testPassword",
			CfServiceKeyName:  "testServiceKeyName",
		}

		//when
		var connectionDetails ConnectionDetailsHTTP
		var err error
		connectionDetails, err = GetAbapCommunicationArrangementInfo(options, &command.Command{}, "", false)

		//then
		assert.Equal(t, "", connectionDetails.URL)
		assert.Equal(t, "", connectionDetails.User)
		assert.Equal(t, "", connectionDetails.Password)
		assert.Equal(t, "", connectionDetails.XCsrfToken)

		assert.Error(t, err)
	})

	// t.Run("GetAbapCommunicationArrangementInfo - Success", func(t *testing.T) {

	// 	//given

	// 	// todo: mock cf cli

	// 	options := AbapEnvironmentOptions{
	// 		CfAPIEndpoint:     "https://api.endpoint.com",
	// 		CfSpace:           "testSpace",
	// 		CfOrg:             "testOrg",
	// 		CfServiceInstance: "testInstance",
	// 		Username:          "testUser",
	// 		Password:          "testPassword",
	// 		CfServiceKeyName:  "testServiceKeyName",
	// 	}

	// 	//when
	// 	var connectionDetails ConnectionDetailsHTTP
	// 	var err error
	// 	connectionDetails, err = GetAbapCommunicationArrangementInfo(options, &command.Command{}, "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull", false)

	// 	//then
	// 	assert.Equal(t, "", connectionDetails.URL)
	// 	assert.Equal(t, "", connectionDetails.User)
	// 	assert.Equal(t, "", connectionDetails.Password)
	// 	assert.Equal(t, "", connectionDetails.XCsrfToken)

	// 	assert.NoError(t, err)
	// })
}
