//go:build unit
// +build unit

package cmd

import (
	"errors"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/stretchr/testify/assert"
)

func TestCloudFoundryCreateServiceKey(t *testing.T) {
	var telemetryData telemetry.CustomData

	t.Run("creates a key without configuration", func(t *testing.T) {
		config := cloudFoundryCreateServiceKeyOptions{
			CfAPIEndpoint: "https://api.endpoint.com", CfOrg: "testOrg", CfSpace: "testSpace",
			Username: "testUser", Password: "testPassword", CfServiceInstance: "testInstance", CfServiceKeyName: "testKey", CfAsync: true,
		}
		runner := &mock.ExecMockRunner{}

		err := runCloudFoundryCreateServiceKey(&config, &telemetryData, runner)

		if assert.NoError(t, err) {
			assert.Equal(t, []string{"create-service-key", "testInstance", "testKey"}, runner.Calls[1].Params)
		}
	})

	t.Run("creates an asynchronous configured key", func(t *testing.T) {
		config := cloudFoundryCreateServiceKeyOptions{
			CfAPIEndpoint: "https://api.endpoint.com", CfOrg: "testOrg", CfSpace: "testSpace",
			Username: "testUser", Password: "testPassword", CfServiceInstance: "testInstance", CfServiceKeyName: "testKey", CfServiceKeyConfig: "testconfig.yml", CfAsync: true,
		}
		runner := &mock.ExecMockRunner{}

		err := runCloudFoundryCreateServiceKey(&config, &telemetryData, runner)

		if assert.NoError(t, err) {
			assert.Equal(t, []string{"create-service-key", "testInstance", "testKey", "-c", "testconfig.yml"}, runner.Calls[1].Params)
		}
	})

	t.Run("creates a synchronous configured key", func(t *testing.T) {
		config := cloudFoundryCreateServiceKeyOptions{
			CfAPIEndpoint: "https://api.endpoint.com", CfOrg: "testOrg", CfSpace: "testSpace",
			Username: "testUser", Password: "testPassword", CfServiceInstance: "testInstance", CfServiceKeyName: "testKey", CfServiceKeyConfig: `{"scenario_id":"SAP_COM_0510","type":"basic"}`,
		}
		runner := &mock.ExecMockRunner{}

		err := runCloudFoundryCreateServiceKey(&config, &telemetryData, runner)

		if assert.NoError(t, err) {
			assert.Equal(t, []string{"create-service-key", "testInstance", "testKey", "-c", `{"scenario_id":"SAP_COM_0510","type":"basic"}`, cfCliSynchronousRequestFlag}, runner.Calls[1].Params)
		}
	})
}

func TestCloudFoundryCreateServiceKeyErrors(t *testing.T) {
	config := cloudFoundryCreateServiceKeyOptions{
		CfAPIEndpoint: "https://api.endpoint.com", CfOrg: "testOrg", CfSpace: "testSpace",
		Username: "testUser", Password: "testPassword", CfServiceInstance: "testInstance", CfServiceKeyName: "testKey", CfServiceKeyConfig: `{"scenario_id":"SAP_COM_0510","type":"basic"}`,
	}
	var telemetryData telemetry.CustomData

	t.Run("returns login errors", func(t *testing.T) {
		runner := &mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{"cf login .*": errors.New("login failed")}}

		err := runCloudFoundryCreateServiceKey(&config, &telemetryData, runner)

		assert.EqualError(t, err, "Error while logging in occurred: Failed to login to Cloud Foundry: login failed")
	})

	t.Run("returns logout errors after success", func(t *testing.T) {
		runner := &mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{"cf logout": errors.New("logout failed")}}

		err := runCloudFoundryCreateServiceKey(&config, &telemetryData, runner)

		assert.EqualError(t, err, "Error while logging out occurred: Failed to Logout of Cloud Foundry: logout failed")
	})

	t.Run("returns service key creation errors before logout errors", func(t *testing.T) {
		runner := &mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{
			`cf create-service-key testInstance testKey -c {"scenario_id":"SAP_COM_0510","type":"basic"} --wait`: errors.New("create failed"),
			"cf logout": errors.New("logout failed"),
		}}

		err := runCloudFoundryCreateServiceKey(&config, &telemetryData, runner)

		assert.EqualError(t, err, "Failed to Create Service Key: create failed")
	})
}
