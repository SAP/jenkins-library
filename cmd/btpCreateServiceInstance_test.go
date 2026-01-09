package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/btp"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/stretchr/testify/assert"
)

func btpMockCleanup(m *btp.BtpExecutorMock) {
	m.ShouldFailOnCommand = map[string]error{}
	m.StdoutReturn = map[string]string{}
	m.Calls = []btp.BtpExecCall{}
}

func TestRunBtpCreateService(t *testing.T) {
	InstanceName := "testServiceInstance"
	m := &btp.BtpExecutorMock{}
	m.Stdout(new(bytes.Buffer))

	var telemetryData telemetry.CustomData

	t.Run("Create service: no identity provider", func(t *testing.T) {
		defer btpMockCleanup(m)

		utils := btp.NewBTPUtils(m)
		m.StdoutReturn = map[string]string{
			"btp login .*": "Authentication successful",
			"btp get services/instance": fmt.Sprintf(`
				{
					"id": "xxx",
					"name": "%s",
					"ready": true
				}`, InstanceName),
		}

		// init
		config := btpCreateServiceInstanceOptions{
			Url:                 "https://api.endpoint.com",
			Subdomain:           "testSubdomain",
			Subaccount:          "testSubaccount",
			PlanName:            "testPlan",
			OfferingName:        "testOffering",
			ServiceInstanceName: InstanceName,
			CreateServiceConfig: "testCreateServiceConfig",
			Timeout:             60,
			PollInterval:        5,
			User:                "testUser",
			Password:            "testPassword",
		}

		// test
		err := runBtpCreateServiceInstance(&config, &telemetryData, *utils)

		// assert
		if assert.NoError(t, err) {
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"login", "--url", config.Url, "--subdomain", config.Subdomain, "--user", config.User, "--password", config.Password}},
				m.Calls[1])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"create", "services/instance", "--name", config.ServiceInstanceName, "--subaccount", config.Subaccount, "--parameters", config.CreateServiceConfig, "--plan-name", config.PlanName, "--offering-name", config.OfferingName}},
				m.Calls[2])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"logout"}},
				m.Calls[len(m.Calls)-1])
		}
	})

	t.Run("Create service: full parameters", func(t *testing.T) {
		defer btpMockCleanup(m)

		utils := btp.NewBTPUtils(m)
		m.StdoutReturn = map[string]string{
			"btp login .*": "Authentication successful",
			"btp get services/instance": fmt.Sprintf(`
				{
					"id": "xxx",
					"name": "%s",
					"ready": true
				}`, InstanceName),
		}

		// init
		config := btpCreateServiceInstanceOptions{
			Url:                 "https://api.endpoint.com",
			Subdomain:           "testSubdomain",
			Idp:                 "testIdentityProvider",
			Subaccount:          "testSubaccount",
			PlanName:            "testPlan",
			OfferingName:        "testOffering",
			ServiceInstanceName: InstanceName,
			CreateServiceConfig: "testCreateServiceConfig",
			Timeout:             60,
			PollInterval:        5,
			User:                "testUser",
			Password:            "testPassword",
		}

		// test
		err := runBtpCreateServiceInstance(&config, &telemetryData, *utils)

		// assert
		if assert.NoError(t, err) {
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"login", "--url", config.Url, "--subdomain", config.Subdomain, "--user", config.User, "--password", config.Password, "--idp", config.Idp}},
				m.Calls[1])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"create", "services/instance", "--name", config.ServiceInstanceName, "--subaccount", config.Subaccount, "--parameters", config.CreateServiceConfig, "--plan-name", config.PlanName, "--offering-name", config.OfferingName}},
				m.Calls[2])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"logout"}},
				m.Calls[len(m.Calls)-1])
		}
	})
}
