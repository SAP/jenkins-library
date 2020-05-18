package cloudfoundry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCloudFoundryLoginCheck(t *testing.T) {
	t.Run("CF Login check: missing parameter", func(t *testing.T) {
		cfconfig := LoginOptions{}
		loggedIn, err := LoginCheck(cfconfig)
		assert.Equal(t, false, loggedIn)
		assert.EqualError(t, err, "Cloud Foundry API endpoint parameter missing. Please provide the Cloud Foundry Endpoint")
	})

	t.Run("CF Login check: failure case", func(t *testing.T) {
		cfconfig := LoginOptions{
			CfAPIEndpoint: "https://api.endpoint.com",
		}
		loggedIn, err := LoginCheck(cfconfig)
		assert.Equal(t, false, loggedIn)
		assert.Error(t, err)
	})
}

func TestCloudFoundryLogin(t *testing.T) {
	t.Run("CF Login: missing parameter", func(t *testing.T) {
		cfconfig := LoginOptions{}
		err := Login(cfconfig)
		assert.EqualError(t, err, "Failed to login to Cloud Foundry: Parameters missing. Please provide the Cloud Foundry Endpoint, Org, Space, Username and Password")
	})
	t.Run("CF Login: failure", func(t *testing.T) {
		cfconfig := LoginOptions{
			CfAPIEndpoint: "https://api.endpoint.com",
			CfSpace:       "testSpace",
			CfOrg:         "testOrg",
			Username:      "testUser",
			Password:      "testPassword",
		}
		err := Login(cfconfig)
		assert.Error(t, err)
	})
}

func TestCloudFoundryLogout(t *testing.T) {
	t.Run("CF Logout", func(t *testing.T) {
		err := Logout()
		if err == nil {
			assert.Equal(t, nil, err)
		} else {
			assert.Error(t, err)
		}
	})
}

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
		abapKey, err := ReadServiceKeyAbapEnvironment(cfconfig, true)
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
		assert.Error(t, err)
	})
}
