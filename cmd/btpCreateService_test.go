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

	t.Run("Create service: no tenant", func(t *testing.T) {
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
		config := btpCreateServiceOptions{
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
		err := runBtpCreateService(&config, &telemetryData, *utils)

		// assert
		if assert.NoError(t, err) {
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"login", "--url", "https://api.endpoint.com", "--subdomain", "testSubdomain", "--user", "testUser", "--password", "testPassword"}},
				m.Calls[1])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"create", "services/instance", "--name", "testServiceInstance", "--subaccount", "testSubaccount", "--parameters", "testCreateServiceConfig", "--plan-name", "testPlan", "--offering-name", "testOffering"}},
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
		config := btpCreateServiceOptions{
			Url:                 "https://api.endpoint.com",
			Subdomain:           "testSubdomain",
			Tenant:              "testTenant",
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
		err := runBtpCreateService(&config, &telemetryData, *utils)

		// assert
		if assert.NoError(t, err) {
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"login", "--url", "https://api.endpoint.com", "--subdomain", "testSubdomain", "--user", "testUser", "--password", "testPassword", "--idp", "testTenant"}},
				m.Calls[1])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"create", "services/instance", "--name", "testServiceInstance", "--subaccount", "testSubaccount", "--parameters", "testCreateServiceConfig", "--plan-name", "testPlan", "--offering-name", "testOffering"}},
				m.Calls[2])
			assert.Equal(t,
				btp.BtpExecCall{Exec: "btp", Params: []string{"logout"}},
				m.Calls[len(m.Calls)-1])
		}
	})
}
