package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/btp"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/stretchr/testify/assert"
)

func TestRunBtpDeleteService(t *testing.T) {
	InstanceName := "testServiceInstance"
	m := &btp.BtpExecutorMock{}
	m.Stdout(new(bytes.Buffer))

	var telemetryData telemetry.CustomData

	t.Run("Delete service: no tenant", func(t *testing.T) {
		defer btpMockCleanup(m)

		utils := btp.NewBTPUtils(m)
		m.StdoutReturn = map[string]string{
			"btp login .*": "Authentication successful",
		}
		m.ShouldFailOnCommand = map[string]error{
			"btp get services/instance": fmt.Errorf(`
				{
				"error": "BadRequest",
				"description": "Could not find such instance"
				}`),
		}

		// init
		config := btpDeleteServiceOptions{
			Url:                 "https://api.endpoint.com",
			Subdomain:           "testSubdomain",
			Subaccount:          "testSubaccount",
			ServiceInstanceName: InstanceName,
			Timeout:             60,
			PollInterval:        5,
			User:                "testUser",
			Password:            "testPassword",
		}

		// test
		err := runBtpDeleteService(&config, &telemetryData, *utils)

		// assert
		if assert.NoError(t, err) {
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"login", "--url", config.Url, "--subdomain", config.Subdomain, "--user", config.User, "--password", config.Password}},
				m.Calls[1])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"delete", "services/instance", "--name", config.ServiceInstanceName, "--subaccount", config.Subaccount, "--confirm"}},
				m.Calls[2])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"logout"}},
				m.Calls[len(m.Calls)-1])
		}
	})

	t.Run("Delete service: full parameters", func(t *testing.T) {
		defer btpMockCleanup(m)

		utils := btp.NewBTPUtils(m)
		m.StdoutReturn = map[string]string{
			"btp login .*": "Authentication successful",
		}
		m.ShouldFailOnCommand = map[string]error{
			"btp get services/instance": fmt.Errorf(`
				{
				"error": "BadRequest",
				"description": "Could not find such instance"
				}`),
		}

		// init
		config := btpDeleteServiceOptions{
			Url:                 "https://api.endpoint.com",
			Subdomain:           "testSubdomain",
			Tenant:              "testTenant",
			Subaccount:          "testSubaccount",
			ServiceInstanceName: InstanceName,
			Timeout:             60,
			PollInterval:        5,
			User:                "testUser",
			Password:            "testPassword",
		}

		// test
		err := runBtpDeleteService(&config, &telemetryData, *utils)

		// assert
		if assert.NoError(t, err) {
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"login", "--url", config.Url, "--subdomain", config.Subdomain, "--user", config.User, "--password", config.Password, "--idp", config.Tenant}},
				m.Calls[1])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"delete", "services/instance", "--name", config.ServiceInstanceName, "--subaccount", config.Subaccount, "--confirm"}},
				m.Calls[2])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"logout"}},
				m.Calls[len(m.Calls)-1])
		}
	})
}
