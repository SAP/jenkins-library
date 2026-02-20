package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/btp"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/stretchr/testify/assert"
)

func TestRunBtpDeleteServiceBinding(t *testing.T) {
	BindingName := "testServiceBinding"
	BindingId := "xxxx"
	InstanceName := "test_instance"
	m := &btp.BtpExecutorMock{}
	m.Stdout(new(bytes.Buffer))

	var telemetryData telemetry.CustomData

	t.Run("Delete service binding: no identity provider", func(t *testing.T) {
		defer btpMockCleanup(m)

		utils := btp.NewBTPUtils(m)
		m.StdoutReturn = map[string]string{
			"btp .* login .+": "Authentication successful",
			"btp .* get services/instance .+": fmt.Sprintf(`
				{
				"id": "xxxx",
				"name": "%s",
				"ready": true
				}`, InstanceName),
			"btp .* list services/binding .+": fmt.Sprintf(`
				[{
				"id": "%s",
				"name": "%s",
				"ready": true
				}]`, BindingId, BindingName),
		}
		m.ShouldFailOnCommand = map[string]error{
			"btp .* get services/binding": fmt.Errorf(`
				{
				"error": "BadRequest",
				"description": "Could not find such binding"
				}`),
		}

		// init
		config := btpDeleteServiceBindingOptions{
			Url:                 "https://api.endpoint.com",
			Subdomain:           "testSubdomain",
			Subaccount:          "testSubaccount",
			ServiceBindingName:  BindingName,
			ServiceInstanceName: InstanceName,
			Timeout:             60,
			PollInterval:        5,
			User:                "testUser",
			Password:            "testPassword",
		}

		// test
		err := runBtpDeleteServiceBinding(&config, &telemetryData, *utils)

		// assert
		if assert.NoError(t, err) {
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"--format", "json", "login", "--url", config.Url, "--subdomain", config.Subdomain, "--user", config.User, "--password", config.Password}},
				m.Calls[0])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"--format", "json", "delete", "services/binding", "--subaccount", config.Subaccount, "--id", BindingId, "--confirm"}},
				m.Calls[3])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"--format", "json", "logout"}},
				m.Calls[len(m.Calls)-1])
		}
	})

	t.Run("Delete service binding: full parameters", func(t *testing.T) {
		defer btpMockCleanup(m)

		utils := btp.NewBTPUtils(m)
		m.StdoutReturn = map[string]string{
			"btp .* login .+": "Authentication successful",
			"btp .* get services/instance .+": fmt.Sprintf(`
				{
				"id": "xxxx",
				"name": "%s",
				"ready": true
				}`, InstanceName),
			"btp .* list services/binding .+": fmt.Sprintf(`
				[{
				"id": "%s",
				"name": "%s",
				"ready": true
				}]`, BindingId, BindingName),
		}
		m.ShouldFailOnCommand = map[string]error{
			"btp .* get services/binding": fmt.Errorf(`
				{
				"error": "BadRequest",
				"description": "Could not find such binding"
				}`),
		}

		// init
		config := btpDeleteServiceBindingOptions{
			Url:                 "https://api.endpoint.com",
			Subdomain:           "testSubdomain",
			Idp:                 "testIdentityProvider",
			Subaccount:          "testSubaccount",
			ServiceBindingName:  BindingName,
			ServiceInstanceName: InstanceName,
			Timeout:             60,
			PollInterval:        5,
			User:                "testUser",
			Password:            "testPassword",
		}

		// test
		err := runBtpDeleteServiceBinding(&config, &telemetryData, *utils)

		// assert
		if assert.NoError(t, err) {
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"--format", "json", "login", "--url", config.Url, "--subdomain", config.Subdomain, "--user", config.User, "--password", config.Password, "--idp", config.Idp}},
				m.Calls[0])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"--format", "json", "delete", "services/binding", "--subaccount", config.Subaccount, "--id", BindingId, "--confirm"}},
				m.Calls[3])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"--format", "json", "logout"}},
				m.Calls[len(m.Calls)-1])
		}
	})
}
