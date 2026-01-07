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
	m := &btp.BtpExecutorMock{}
	m.Stdout(new(bytes.Buffer))

	var telemetryData telemetry.CustomData

	t.Run("Delete service binding: no identity provider", func(t *testing.T) {
		defer btpMockCleanup(m)

		utils := btp.NewBTPUtils(m)
		m.StdoutReturn = map[string]string{
			"btp login .*": "Authentication successful",
		}
		m.ShouldFailOnCommand = map[string]error{
			"btp get services/binding": fmt.Errorf(`
				{
				"error": "BadRequest",
				"description": "Could not find such binding"
				}`),
		}

		// init
		config := btpDeleteServiceBindingOptions{
			Url:                "https://api.endpoint.com",
			Subdomain:          "testSubdomain",
			Subaccount:         "testSubaccount",
			ServiceBindingName: BindingName,
			Timeout:            60,
			PollInterval:       5,
			User:               "testUser",
			Password:           "testPassword",
		}

		// test
		err := runBtpDeleteServiceBinding(&config, &telemetryData, *utils)

		// assert
		if assert.NoError(t, err) {
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"login", "--url", config.Url, "--subdomain", config.Subdomain, "--user", config.User, "--password", config.Password}},
				m.Calls[1])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"delete", "services/binding", "--subaccount", config.Subaccount, "--name", config.ServiceBindingName, "--confirm"}},
				m.Calls[2])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"logout"}},
				m.Calls[len(m.Calls)-1])
		}
	})

	t.Run("Delete service binding: full parameters", func(t *testing.T) {
		defer btpMockCleanup(m)

		utils := btp.NewBTPUtils(m)
		m.StdoutReturn = map[string]string{
			"btp login .*": "Authentication successful",
		}
		m.ShouldFailOnCommand = map[string]error{
			"btp get services/binding": fmt.Errorf(`
				{
				"error": "BadRequest",
				"description": "Could not find such binding"
				}`),
		}

		// init
		config := btpDeleteServiceBindingOptions{
			Url:                "https://api.endpoint.com",
			Subdomain:          "testSubdomain",
			Idp:                "testIdentityProvider",
			Subaccount:         "testSubaccount",
			ServiceBindingName: BindingName,
			Timeout:            60,
			PollInterval:       5,
			User:               "testUser",
			Password:           "testPassword",
		}

		// test
		err := runBtpDeleteServiceBinding(&config, &telemetryData, *utils)

		// assert
		if assert.NoError(t, err) {
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"login", "--url", config.Url, "--subdomain", config.Subdomain, "--user", config.User, "--password", config.Password, "--idp", config.Idp}},
				m.Calls[1])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"delete", "services/binding", "--subaccount", config.Subaccount, "--name", config.ServiceBindingName, "--confirm"}},
				m.Calls[2])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"logout"}},
				m.Calls[len(m.Calls)-1])
		}
	})
}
