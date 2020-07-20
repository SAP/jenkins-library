package cloudfoundry

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func loginMockCleanup(m *mock.ExecMockRunner) {
	m.ShouldFailOnCommand = map[string]error{}
	m.StdoutReturn = map[string]string{}
	m.Calls = []mock.ExecCall{}
}

func TestCloudFoundryLoginCheck(t *testing.T) {

	m := &mock.ExecMockRunner{}

	t.Run("CF Login check: logged in", func(t *testing.T) {

		defer loginMockCleanup(m)

		cfconfig := LoginOptions{
			CfAPIEndpoint: "https://api.endpoint.com",
		}
		cf := CFUtils{Exec: m, loggedIn: true}
		loggedIn, err := cf.LoginCheck(cfconfig)
		if assert.NoError(t, err) {
			assert.True(t, loggedIn)
		}
	})

	t.Run("CF Login check: not logged in", func(t *testing.T) {

		defer loginMockCleanup(m)

		cfconfig := LoginOptions{
			CfAPIEndpoint: "https://api.endpoint.com",
		}
		cf := CFUtils{Exec: m, loggedIn: false}
		loggedIn, err := cf.LoginCheck(cfconfig)
		if assert.NoError(t, err) {
			assert.False(t, loggedIn)
		}
	})
}

func TestCloudFoundryLogin(t *testing.T) {

	m := &mock.ExecMockRunner{}

	t.Run("CF Login: missing parameter", func(t *testing.T) {

		defer loginMockCleanup(m)

		cfconfig := LoginOptions{}
		cf := CFUtils{Exec: m}
		err := cf.Login(cfconfig)
		assert.EqualError(t, err, "Failed to login to Cloud Foundry: Parameters missing. Please provide the Cloud Foundry Endpoint, Org, Space, Username and Password")
	})
	t.Run("CF Login: failure", func(t *testing.T) {

		defer loginMockCleanup(m)

		m.ShouldFailOnCommand = map[string]error{"cf login .*": fmt.Errorf("wrong password or account does not exist")}

		cfconfig := LoginOptions{
			CfAPIEndpoint: "https://api.endpoint.com",
			CfSpace:       "testSpace",
			CfOrg:         "testOrg",
			Username:      "testUser",
			Password:      "testPassword",
		}

		cf := CFUtils{Exec: m}
		err := cf.Login(cfconfig)
		if assert.EqualError(t, err, "Failed to login to Cloud Foundry: wrong password or account does not exist") {
			assert.False(t, cf.loggedIn)
			assert.Equal(t, []mock.ExecCall{
				mock.ExecCall{Exec: "cf", Params: []string{
					"login",
					"-a", "https://api.endpoint.com",
					"-o", "testOrg",
					"-s", "testSpace",
					"-u", "testUser",
					"-p", "testPassword",
				}},
			}, m.Calls)
		}
	})

	t.Run("CF Login: success", func(t *testing.T) {

		defer loginMockCleanup(m)

		m.StdoutReturn = map[string]string{"cf api:*": "Not logged in"}

		cfconfig := LoginOptions{
			CfAPIEndpoint: "https://api.endpoint.com",
			CfSpace:       "testSpace",
			CfOrg:         "testOrg",
			Username:      "testUser",
			Password:      "testPassword",
		}
		cf := CFUtils{Exec: m}
		err := cf.Login(cfconfig)
		if assert.NoError(t, err) {
			assert.True(t, cf.loggedIn)
			assert.Equal(t, []mock.ExecCall{
				mock.ExecCall{Exec: "cf", Params: []string{
					"login",
					"-a", "https://api.endpoint.com",
					"-o", "testOrg",
					"-s", "testSpace",
					"-u", "testUser",
					"-p", "testPassword",
				}},
			}, m.Calls)
		}
	})

	t.Run("CF Login: with additional login options", func(t *testing.T) {

		defer loginMockCleanup(m)

		cfconfig := LoginOptions{
			CfAPIEndpoint: "https://api.endpoint.com",
			CfSpace:       "testSpace",
			CfOrg:         "testOrg",
			Username:      "testUser",
			Password:      "testPassword",
			CfLoginOpts: []string{
				// should never used in productive environment, but it is useful for rapid prototyping/troubleshooting
				"--skip-ssl-validation",
				"--origin", "ldap",
			},
			CfAPIOpts: []string{
				"--skip-ssl-validation",
			},
		}
		cf := CFUtils{Exec: m}
		err := cf.Login(cfconfig)
		if assert.NoError(t, err) {
			assert.True(t, cf.loggedIn)
			assert.Equal(t, []mock.ExecCall{
				mock.ExecCall{Exec: "cf", Params: []string{
					"login",
					"-a", "https://api.endpoint.com",
					"-o", "testOrg",
					"-s", "testSpace",
					"-u", "testUser",
					"-p", "testPassword",
					"--skip-ssl-validation",
					"--origin", "ldap",
				}},
			}, m.Calls)
		}
	})
}

func TestCloudFoundryLogout(t *testing.T) {
	t.Run("CF Logout", func(t *testing.T) {
		cf := CFUtils{Exec: &mock.ExecMockRunner{}, loggedIn: true}
		err := cf.Logout()
		if assert.NoError(t, err) {
			assert.False(t, cf.loggedIn)
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
		cf := CFUtils{Exec: &command.Command{}}
		abapKey, err := cf.ReadServiceKeyAbapEnvironment(cfconfig, true)
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
