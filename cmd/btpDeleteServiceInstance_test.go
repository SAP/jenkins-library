package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/btp"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/stretchr/testify/assert"
)

func TestRunBtpDeleteServiceInstance(t *testing.T) {
	InstanceName := "testServiceInstance"
	InstanceId := "xxxx"
	m := &btp.BtpExecutorMock{}
	m.Stdout(new(bytes.Buffer))

	var telemetryData telemetry.CustomData

	t.Run("Delete service: no identity provider", func(t *testing.T) {
		defer btpMockCleanup(m)

		utils := btp.NewBTPUtils(m)
		m.StdoutReturn = map[string]string{
			"btp .* login .+": "Authentication successful",
			"btp .* get services/instance (.*)--name": fmt.Sprintf(`
				{
				"id": "%s",
				"name": "%s",
				"ready": true
				}`, InstanceId, InstanceName),
		}
		m.ShouldFailOnCommand = map[string]error{
			"btp .* get services/instance (.*)--id": fmt.Errorf(`
				{
				"error": "BadRequest",
				"description": "Could not find such instance"
				}`),
		}

		// init
		config := btpDeleteServiceInstanceOptions{
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
		err := runBtpDeleteServiceInstance(&config, &telemetryData, *utils)

		// assert
		if assert.NoError(t, err) {
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"--format", "json", "login", "--url", config.Url, "--subdomain", config.Subdomain, "--user", config.User, "--password", config.Password}},
				m.Calls[0])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"--format", "json", "delete", "services/instance", "--id", InstanceId, "--subaccount", config.Subaccount, "--confirm"}},
				m.Calls[2])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"--format", "json", "logout"}},
				m.Calls[len(m.Calls)-1])
		}
	})

	t.Run("Delete service: full parameters", func(t *testing.T) {
		defer btpMockCleanup(m)

		utils := btp.NewBTPUtils(m)
		m.StdoutReturn = map[string]string{
			"btp .* login .+": "Authentication successful",
			"btp .* get services/instance (.*)--name": fmt.Sprintf(`
				{
				"id": "%s",
				"name": "%s",
				"ready": true
				}`, InstanceId, InstanceName),
		}
		m.ShouldFailOnCommand = map[string]error{
			"btp .* get services/instance (.*)--id": fmt.Errorf(`
				{
				"error": "BadRequest",
				"description": "Could not find such instance"
				}`),
		}

		// init
		config := btpDeleteServiceInstanceOptions{
			Url:                 "https://api.endpoint.com",
			Subdomain:           "testSubdomain",
			Idp:                 "testIdentityProvider",
			Subaccount:          "testSubaccount",
			ServiceInstanceName: InstanceName,
			Timeout:             60,
			PollInterval:        5,
			User:                "testUser",
			Password:            "testPassword",
		}

		// test
		err := runBtpDeleteServiceInstance(&config, &telemetryData, *utils)

		// assert
		if assert.NoError(t, err) {
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"--format", "json", "login", "--url", config.Url, "--subdomain", config.Subdomain, "--user", config.User, "--password", config.Password, "--idp", config.Idp}},
				m.Calls[0])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"--format", "json", "delete", "services/instance", "--id", InstanceId, "--subaccount", config.Subaccount, "--confirm"}},
				m.Calls[2])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"--format", "json", "logout"}},
				m.Calls[len(m.Calls)-1])
		}
	})
}
