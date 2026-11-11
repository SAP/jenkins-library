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
		assert.EqualError(t, err, "Failed to login to Cloud Foundry: Parameters missing. Please provide the Cloud Foundry Endpoint, Org and Space")
	})
	t.Run("CF Login: token authentication", func(t *testing.T) {

		defer loginMockCleanup(m)

		cfconfig := LoginOptions{
			CfAPIEndpoint: "https://api.endpoint.com",
			CfSpace:       "testSpace",
			CfOrg:         "testOrg",
			Token:         "testToken",
			TokenOrigin:   "testOrigin",
		}
		err := Login(m, cfconfig)
		if assert.NoError(t, err) {
			assert.Equal(t, []mock.ExecCall{
				{Exec: "cf", Params: []string{"api", "https://api.endpoint.com"}},
				{Exec: "cf", Params: []string{"auth", "--assertion", "testToken", "--origin", "testOrigin"}},
				{Exec: "cf", Params: []string{"target", "-o", "testOrg", "-s", "testSpace"}},
			}, m.Calls)
		}
	})

	t.Run("CF Login: token authentication without origin", func(t *testing.T) {

		defer loginMockCleanup(m)

		cfconfig := LoginOptions{
			CfAPIEndpoint: "https://api.endpoint.com",
			CfSpace:       "testSpace",
			CfOrg:         "testOrg",
			Token:         "testToken",
		}
		err := Login(m, cfconfig)
		if assert.NoError(t, err) {
			assert.Equal(t, []mock.ExecCall{
				{Exec: "cf", Params: []string{"api", "https://api.endpoint.com"}},
				{Exec: "cf", Params: []string{"auth", "--assertion", "testToken"}},
				{Exec: "cf", Params: []string{"target", "-o", "testOrg", "-s", "testSpace"}},
			}, m.Calls)
		}
	})

	t.Run("CF Login: token failure does not fall back to user credentials", func(t *testing.T) {

		defer loginMockCleanup(m)

		m.ShouldFailOnCommand = map[string]error{"cf auth .*": fmt.Errorf("invalid assertion")}
		cfconfig := LoginOptions{
			CfAPIEndpoint: "https://api.endpoint.com",
			CfSpace:       "testSpace",
			CfOrg:         "testOrg",
			Username:      "testUser",
			Password:      "testPassword",
			Token:         "testToken",
		}
		err := Login(m, cfconfig)
		if assert.EqualError(t, err, "Failed to login to Cloud Foundry: invalid assertion") {
			assert.Equal(t, []mock.ExecCall{
				{Exec: "cf", Params: []string{"api", "https://api.endpoint.com"}},
				{Exec: "cf", Params: []string{"auth", "--assertion", "testToken"}},
			}, m.Calls)
		}
	})

	t.Run("CF Login: token takes precedence over user credentials", func(t *testing.T) {

		defer loginMockCleanup(m)

		cfconfig := LoginOptions{
			CfAPIEndpoint: "https://api.endpoint.com",
			CfSpace:       "testSpace",
			CfOrg:         "testOrg",
			Username:      "testUser",
			Password:      "testPassword",
			Token:         "testToken",
		}
		err := Login(m, cfconfig)
		if assert.NoError(t, err) {
			assert.Equal(t, []mock.ExecCall{
				{Exec: "cf", Params: []string{"api", "https://api.endpoint.com"}},
				{Exec: "cf", Params: []string{"auth", "--assertion", "testToken"}},
				{Exec: "cf", Params: []string{"target", "-o", "testOrg", "-s", "testSpace"}},
			}, m.Calls)
		}
	})

	t.Run("CF Login: missing authentication credentials", func(t *testing.T) {

		defer loginMockCleanup(m)

		cfconfig := LoginOptions{
			CfAPIEndpoint: "https://api.endpoint.com",
			CfSpace:       "testSpace",
			CfOrg:         "testOrg",
		}
		err := Login(m, cfconfig)
		assert.EqualError(t, err, "Failed to login to Cloud Foundry: Parameters missing. Please provide a token or Username and Password")
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
	t.Run("CF Logout", func(t *testing.T) {
		runner := &mock.ExecMockRunner{}
		err := Logout(runner)
		if assert.NoError(t, err) {
		}
	})
}
