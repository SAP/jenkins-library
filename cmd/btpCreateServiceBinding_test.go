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
			"btp login .*": "Authentication successful",
			"btp get services/binding": fmt.Sprintf(`
				{
				"id": "xxxx",
				"name": "%s",
				"ready": true
				}`, BindingName),
		}

		// init
		config := btpCreateServiceBindingOptions{
			Url:                        "https://api.endpoint.com",
			Subdomain:                  "testSubdomain",
			Subaccount:                 "testSubaccount",
			ServiceInstanceName:        InstanceName,
			ServiceBindingName:         BindingName,
			CreateServiceBindingConfig: "testCreateServiceConfig",
			Timeout:                    60,
			PollInterval:               5,
			User:                       "testUser",
			Password:                   "testPassword",
		}

		// test
		err := runBtpCreateServiceBinding(&config, &telemetryData, *utils)

		// assert
		if assert.NoError(t, err) {
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"login", "--url", config.Url, "--subdomain", config.Subdomain, "--user", config.User, "--password", config.Password}},
				m.Calls[1])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"create", "services/binding", "--name", config.ServiceBindingName, "--instance-name", config.ServiceInstanceName, "--subaccount", config.Subaccount, "--parameters", config.CreateServiceBindingConfig}},
				m.Calls[2])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"logout"}},
				m.Calls[len(m.Calls)-1])
		}
	})

	t.Run("Create service binding: full parameters", func(t *testing.T) {
		defer btpMockCleanup(m)

		utils := btp.NewBTPUtils(m)
		m.StdoutReturn = map[string]string{
			"btp login .*": "Authentication successful",
			"btp get services/binding": fmt.Sprintf(`
				{
					"id": "xxx",
					"name": "%s",
					"ready": true
				}`, InstanceName),
		}

		// init
		config := btpCreateServiceBindingOptions{
			Url:                        "https://api.endpoint.com",
			Subdomain:                  "testSubdomain",
			Idp:                        "testIdentityProvider",
			Subaccount:                 "testSubaccount",
			ServiceInstanceName:        InstanceName,
			ServiceBindingName:         BindingName,
			CreateServiceBindingConfig: "testCreateServiceBindingConfig",
			Timeout:                    60,
			PollInterval:               5,
			User:                       "testUser",
			Password:                   "testPassword",
		}

		// test
		err := runBtpCreateServiceBinding(&config, &telemetryData, *utils)

		// assert
		if assert.NoError(t, err) {
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"login", "--url", config.Url, "--subdomain", config.Subdomain, "--user", config.User, "--password", config.Password, "--idp", config.Idp}},
				m.Calls[1])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"create", "services/binding", "--name", config.ServiceBindingName, "--instance-name", config.ServiceInstanceName, "--subaccount", config.Subaccount, "--parameters", config.CreateServiceBindingConfig}},
				m.Calls[2])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"logout"}},
				m.Calls[len(m.Calls)-1])
		}
	})
}
