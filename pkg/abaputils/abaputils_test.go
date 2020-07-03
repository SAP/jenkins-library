package abaputils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadServiceKeyAbapEnvironment(t *testing.T) {
	t.Run("CF ReadServiceKeyAbapEnvironment", func(t *testing.T) {

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
		abapKey, _ = ReadServiceKeyAbapEnvironment(options, true)

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
		//assert.Error(t, err)
	})
}

func TestGetAbapCommunicationInfo(t *testing.T) {
	t.Run("GetAbapCommunicationArrangementInfo", func(t *testing.T) {

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
		connectionDetails, _ = GetAbapCommunicationArrangementInfo(options, "", false)

		//then
		assert.Equal(t, "", connectionDetails.URL)
		assert.Equal(t, "", connectionDetails.User)
		assert.Equal(t, "", connectionDetails.Password)
		assert.Equal(t, "", connectionDetails.XCsrfToken)
		//assert.Error(t, err)
	})
}
