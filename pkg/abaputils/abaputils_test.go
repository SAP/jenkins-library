package abaputils

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/stretchr/testify/assert"
)

func TestCloudFoundryReadServiceKeyAbapEnvironment(t *testing.T) {
	t.Run("CF ReadServiceKeyAbapEnvironment", func(t *testing.T) {

		//given
		cfconfig := AbapEnvironmentOptions{
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
		abapKey, _ = ReadServiceKeyAbapEnvironment(cfconfig, true)

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
		assert.Equal(t, "", abapKey.Systemid)
		assert.Equal(t, "", abapKey.URL)
		//assert.Error(t, err)
	})
}

func TestGetAbapCommunicationInfo(t *testing.T) {
	t.Run("GetAbapCommunicationArrangementInfo", func(t *testing.T) {

		//given
		cfconfig := AbapEnvironmentOptions{
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
		var c = command.Command{}
		connectionDetails, _ = GetAbapCommunicationArrangementInfo(cfconfig, c)

		//then
		assert.Equal(t, "", connectionDetails.URL)
		assert.Equal(t, "", connectionDetails.User)
		assert.Equal(t, "", connectionDetails.Password)
		assert.Equal(t, "", connectionDetails.XCsrfToken)
		//assert.Error(t, err)
	})
}
