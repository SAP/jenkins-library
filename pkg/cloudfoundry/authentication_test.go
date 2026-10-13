//go:build unit
// +build unit

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

func TestCloudFoundryLogin(t *testing.T) {

	m := &mock.ExecMockRunner{}

	t.Run("CF Login: missing parameter", func(t *testing.T) {

		defer loginMockCleanup(m)

		cfconfig := LoginOptions{}
		err := Login(m, cfconfig)
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

		err := Login(m, cfconfig)
		if assert.EqualError(t, err, "Failed to login to Cloud Foundry: wrong password or account does not exist") {
			assert.Equal(t, []mock.ExecCall{
				{Exec: "cf", Params: []string{
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
		err := Login(m, cfconfig)
		if assert.NoError(t, err) {
			assert.Equal(t, []mock.ExecCall{
				{Exec: "cf", Params: []string{
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
		}
		err := Login(m, cfconfig)
		if assert.NoError(t, err) {
			assert.Equal(t, []mock.ExecCall{
				{Exec: "cf", Params: []string{
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
	runner := &mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{"cf logout": fmt.Errorf("logout rejected")}}
	assert.EqualError(t, Logout(runner), "Failed to Logout of Cloud Foundry: logout rejected")
}

func TestCloudFoundryAssertionLogin(t *testing.T) {
	api := mock.ExecCall{Exec: "cf", Params: []string{"api", "https://cf.example"}}
	auth := mock.ExecCall{Exec: "cf", Params: []string{"auth", "--assertion", "synthetic-assertion", "--origin", "ias-platform"}}
	target := mock.ExecCall{Exec: "cf", Params: []string{"target", "-o", "org", "-s", "space"}}
	logout := mock.ExecCall{Exec: "cf", Params: []string{"logout"}}
	for _, tc := range []struct {
		name      string
		configure func(*LoginOptions)
		failures  map[string]error
		wantCalls []mock.ExecCall
		wantError string
	}{
		{name: "token only", wantCalls: []mock.ExecCall{api, auth, target}},
		{name: "assertion wins over credentials", configure: func(o *LoginOptions) { o.Username, o.Password = "user", "password" }, wantCalls: []mock.ExecCall{api, auth, target}},
		{name: "missing target", configure: func(o *LoginOptions) { o.CfSpace = "" }, wantError: "Parameters missing. Please provide the Cloud Foundry Endpoint, Org and Space"},
		{name: "missing origin", configure: func(o *LoginOptions) { o.Origin = "" }, wantError: "Parameter missing. Please provide the Cloud Foundry identity provider origin for assertion authentication"},
		{name: "unsupported login parameters", configure: func(o *LoginOptions) { o.CfLoginOpts = []string{"--origin", "other"} }, wantError: "loginParameters cannot be used with assertion authentication"},
		{name: "api failure", failures: map[string]error{"cf api .*": fmt.Errorf("api rejected")}, wantCalls: []mock.ExecCall{api}, wantError: "api rejected"},
		{name: "auth rejection never downgrades", configure: func(o *LoginOptions) { o.Username, o.Password = "user", "password" }, failures: map[string]error{"cf auth .*": fmt.Errorf("assertion rejected")}, wantCalls: []mock.ExecCall{api, auth}, wantError: "assertion rejected"},
		{name: "target failure logs out", failures: map[string]error{"cf target .*": fmt.Errorf("target rejected")}, wantCalls: []mock.ExecCall{api, auth, target, logout}, wantError: "target rejected"},
		{name: "cleanup failure preserves target error", failures: map[string]error{"cf target .*": fmt.Errorf("target rejected"), "cf logout": fmt.Errorf("logout rejected")}, wantCalls: []mock.ExecCall{api, auth, target, logout}, wantError: "target rejected"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			options := LoginOptions{CfAPIEndpoint: "https://cf.example", CfOrg: "org", CfSpace: "space", Assertion: "synthetic-assertion", Origin: "ias-platform"}
			if tc.configure != nil {
				tc.configure(&options)
			}
			runner := &mock.ExecMockRunner{ShouldFailOnCommand: tc.failures}
			err := Login(runner, options)
			if tc.wantError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, "Failed to login to Cloud Foundry: "+tc.wantError)
			}
			assert.Equal(t, tc.wantCalls, runner.Calls)
		})
	}
}
