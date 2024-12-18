//go:build unit

package cloudfoundry

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
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
				{Exec: "cf", Params: []string{
					"login",
					"-a", "https://api.endpoint.com",
					"-o", "testOrg",
					"-s", "testSpace",
					"-u", "'testUser'",
					"-p", "'testPassword'",
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
				{Exec: "cf", Params: []string{
					"login",
					"-a", "https://api.endpoint.com",
					"-o", "testOrg",
					"-s", "testSpace",
					"-u", "'testUser'",
					"-p", "'testPassword'",
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
		}
		cf := CFUtils{Exec: m}
		err := cf.Login(cfconfig)
		if assert.NoError(t, err) {
			assert.True(t, cf.loggedIn)
			assert.Equal(t, []mock.ExecCall{
				{Exec: "cf", Params: []string{
					"login",
					"-a", "https://api.endpoint.com",
					"-o", "testOrg",
					"-s", "testSpace",
					"-u", "'testUser'",
					"-p", "'testPassword'",
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

	t.Run("CF ReadServiceKey", func(t *testing.T) {

		//given
		m := &mock.ExecMockRunner{}
		defer loginMockCleanup(m)

		const testURL = "testurl.com"
		const oDataURL = "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull"
		const username = "test_user"
		const password = "test_password"
		const serviceKey = `
		cf comment test \n\n
		{"sap.cloud.service":"com.sap.cloud.abap","url": "` + testURL + `" ,"systemid":"H01","abap":{"username":"'` + username + `'","password":"'` + password + `'","communication_scenario_id": "SAP_COM_0510","communication_arrangement_id": "SK_I6CBIRFZPPJDKYNATQA32W","communication_system_id": "SK_I6CBIRFZPPJDKYNATQA32W","communication_inbound_user_id": "CC0000000001","communication_inbound_user_auth_mode": "2"},"binding":{"env": "cf","version": "0.0.1.1","type": "basic","id": "i6cBiRfZppJdKynaTqa32W"},"preserve_host_header": true}`

		m.StdoutReturn = map[string]string{"cf service-key testInstance testServiceKeyName": serviceKey}

		cfconfig := ServiceKeyOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfSpace:           "testSpace",
			CfOrg:             "testOrg",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKeyName",
			Username:          "testUser",
			Password:          "testPassword",
		}

		//when
		var err error
		var abapServiceKey string
		cf := CFUtils{Exec: m}

		abapServiceKey, err = cf.ReadServiceKey(cfconfig)

		//then
		if assert.NoError(t, err) {
			assert.Equal(t, []mock.ExecCall{
				{Exec: "cf", Params: []string{
					"login",
					"-a", "https://api.endpoint.com",
					"-o", "testOrg",
					"-s", "testSpace",
					"-u", "'testUser'",
					"-p", "'testPassword'",
				}},
				{Exec: "cf", Params: []string{"service-key", "testInstance", "testServiceKeyName"}},
				{Exec: "cf", Params: []string{"logout"}},
			}, m.Calls)
		}
		assert.Equal(t, `		{"sap.cloud.service":"com.sap.cloud.abap","url": "`+testURL+`" ,"systemid":"H01","abap":{"username":"'`+username+`'","password":"'`+password+`'","communication_scenario_id": "SAP_COM_0510","communication_arrangement_id": "SK_I6CBIRFZPPJDKYNATQA32W","communication_system_id": "SK_I6CBIRFZPPJDKYNATQA32W","communication_inbound_user_id": "CC0000000001","communication_inbound_user_auth_mode": "2"},"binding":{"env": "cf","version": "0.0.1.1","type": "basic","id": "i6cBiRfZppJdKynaTqa32W"},"preserve_host_header": true}`, abapServiceKey)
	})
}
