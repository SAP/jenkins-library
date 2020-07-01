package abaputils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCloudFoundryReadServiceKeyAbapEnvironment(t *testing.T) {
	t.Run("CF ReadServiceKeyAbapEnvironment", func(t *testing.T) {
		cfconfig := ServiceKeyOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfSpace:           "testSpace",
			CfOrg:             "testOrg",
			CfServiceInstance: "testInstance",
			CfServiceKey:      "testKey",
			Username:          "testUser",
			Password:          "testPassword",
		}
		var abapKey ServiceKey
		abapKey, _ = ReadServiceKeyAbapEnvironment(cfconfig, true)
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

	// t.Run("CF ReadServiceKeyAbapEnvironment Fail", func(t *testing.T) {
	// 	cfconfig := ServiceKeyOptions{
	// 		CfAPIEndpoint:     "https://api.endpoint.com",
	// 		CfSpace:           "testSpace",
	// 		CfOrg:             "testOrg",
	// 		CfServiceInstance: "testInstance",
	// 		CfServiceKey:      "testKey",
	// 		Username:          "testUser",
	// 		Password:          "testPassword",
	// 	}
	// 	var abapKey ServiceKey
	// 	abapKey, err := ReadServiceKeyAbapEnvironment(cfconfig, true)
	// 	assert.Equal(t, "", abapKey.Abap.Password)
	// 	assert.Equal(t, "", abapKey.Abap.Username)
	// 	assert.Equal(t, "", abapKey.Abap.CommunicationArrangementID)
	// 	assert.Equal(t, "", abapKey.Abap.CommunicationScenarioID)
	// 	assert.Equal(t, "", abapKey.Abap.CommunicationSystemID)
	// 	assert.Equal(t, "", abapKey.Binding.Env)
	// 	assert.Equal(t, "", abapKey.Binding.Type)
	// 	assert.Equal(t, "", abapKey.Binding.ID)
	// 	assert.Equal(t, "", abapKey.Binding.Version)
	// 	assert.Equal(t, "", abapKey.Systemid)
	// 	assert.Equal(t, "", abapKey.URL)
	// 	assert.Error(t, err)
	// })
}
