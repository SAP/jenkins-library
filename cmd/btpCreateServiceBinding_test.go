package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/btp"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/stretchr/testify/assert"
)

func TestRunBtpCreateServiceBinding(t *testing.T) {
	t.Parallel()
	InstanceName := "test_instance"
	BindingName := "testServiceBinding"
	m := &btp.BtpExecutorMock{}
	m.Stdout(new(bytes.Buffer))

	var telemetryData telemetry.CustomData

	t.Run("Create service binding: no identity provider", func(t *testing.T) {
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
			"btp .* list services/binding": fmt.Sprintf(`
				[{
				"id": "xxxx",
				"name": "%s",
				"ready": true
				}]`, BindingName),
		}

		// init
		config := btpCreateServiceBindingOptions{
			Url:                 "https://api.endpoint.com",
			Subdomain:           "testSubdomain",
			Subaccount:          "testSubaccount",
			ServiceInstanceName: InstanceName,
			ServiceBindingName:  BindingName,
			Parameters:          "testCreateServiceConfig.json",
			Timeout:             60,
			PollInterval:        5,
			User:                "testUser",
			Password:            "testPassword",
		}

		// test
		err := runBtpCreateServiceBinding(&config, &telemetryData, *utils)

		// assert
		if assert.NoError(t, err) {
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"--format", "json", "login", "--url", config.Url, "--subdomain", config.Subdomain, "--user", config.User, "--password", config.Password}},
				m.Calls[0])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"--format", "json", "create", "services/binding", "--name", config.ServiceBindingName, "--instance-name", config.ServiceInstanceName, "--subaccount", config.Subaccount, "--parameters", config.Parameters}},
				m.Calls[2])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"--format", "json", "logout"}},
				m.Calls[len(m.Calls)-1])
		}
	})

	t.Run("Create service binding: no parameters", func(t *testing.T) {
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
			"btp .* list services/binding": fmt.Sprintf(`
				[{
				"id": "xxxx",
				"name": "%s",
				"ready": true
				}]`, BindingName),
		}

		// init
		config := btpCreateServiceBindingOptions{
			Url:                 "https://api.endpoint.com",
			Subdomain:           "testSubdomain",
			Subaccount:          "testSubaccount",
			ServiceInstanceName: InstanceName,
			ServiceBindingName:  BindingName,
			Timeout:             60,
			PollInterval:        5,
			User:                "testUser",
			Password:            "testPassword",
		}

		// test
		err := runBtpCreateServiceBinding(&config, &telemetryData, *utils)

		// assert
		if assert.NoError(t, err) {
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"--format", "json", "login", "--url", config.Url, "--subdomain", config.Subdomain, "--user", config.User, "--password", config.Password}},
				m.Calls[0])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"--format", "json", "create", "services/binding", "--name", config.ServiceBindingName, "--instance-name", config.ServiceInstanceName, "--subaccount", config.Subaccount}},
				m.Calls[2])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"--format", "json", "logout"}},
				m.Calls[len(m.Calls)-1])
		}
	})

	t.Run("Create service binding: full parameters", func(t *testing.T) {
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
			"btp .* list services/binding": fmt.Sprintf(`
				[{
					"id": "xxx",
					"name": "%s",
					"ready": true
				}]`, BindingName),
		}

		// init
		config := btpCreateServiceBindingOptions{
			Url:                 "https://api.endpoint.com",
			Subdomain:           "testSubdomain",
			Idp:                 "testIdentityProvider",
			Subaccount:          "testSubaccount",
			ServiceInstanceName: InstanceName,
			ServiceBindingName:  BindingName,
			Parameters:          "testCreateServiceBindingConfig.json",
			Timeout:             60,
			PollInterval:        5,
			User:                "testUser",
			Password:            "testPassword",
		}

		// test
		err := runBtpCreateServiceBinding(&config, &telemetryData, *utils)

		// assert
		if assert.NoError(t, err) {
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"--format", "json", "login", "--url", config.Url, "--subdomain", config.Subdomain, "--user", config.User, "--password", config.Password, "--idp", config.Idp}},
				m.Calls[0])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"--format", "json", "create", "services/binding", "--name", config.ServiceBindingName, "--instance-name", config.ServiceInstanceName, "--subaccount", config.Subaccount, "--parameters", config.Parameters}},
				m.Calls[2])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"--format", "json", "logout"}},
				m.Calls[len(m.Calls)-1])
		}
	})
}
