//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestCloudFoundryCreateServiceKey(t *testing.T) {
	var telemetryData telemetry.CustomData
	t.Run("CF Create Service Key: Success case", func(t *testing.T) {
		config := cloudFoundryCreateServiceKeyOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			Username:          "testUser",
			Password:          "testPassword",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testKey",
			CfAsync:           true,
		}
		execRunner := mock.ExecMockRunner{}
		cfUtilsMock := cloudfoundry.CfUtilsMock{}
		defer cfUtilsMock.Cleanup()

		error := runCloudFoundryCreateServiceKey(&config, &telemetryData, &execRunner, &cfUtilsMock)
		if error == nil {
			assert.Equal(t, "cf", execRunner.Calls[0].Exec)
			assert.Equal(t, []string{"create-service-key", "testInstance", "testKey", "--wait"}, execRunner.Calls[0].Params)
		}
	})
	t.Run("CF Create Service Key asynchronous with service Key config: Success case", func(t *testing.T) {
		config := cloudFoundryCreateServiceKeyOptions{
			CfAPIEndpoint:      "https://api.endpoint.com",
			CfOrg:              "testOrg",
			CfSpace:            "testSpace",
			Username:           "testUser",
			Password:           "testPassword",
			CfServiceInstance:  "testInstance",
			CfServiceKeyName:   "testKey",
			CfServiceKeyConfig: "testconfig.yml",
			CfAsync:            true,
		}
		execRunner := mock.ExecMockRunner{}
		cfUtilsMock := cloudfoundry.CfUtilsMock{}
		defer cfUtilsMock.Cleanup()

		error := runCloudFoundryCreateServiceKey(&config, &telemetryData, &execRunner, &cfUtilsMock)
		if error == nil {
			assert.Equal(t, "cf", execRunner.Calls[0].Exec)
			assert.Equal(t, []string{"create-service-key", "testInstance", "testKey", "-c", "testconfig.yml", "--wait"}, execRunner.Calls[0].Params)
		}
	})
	t.Run("CF Create Service Key synchronous with service Key config: Success case", func(t *testing.T) {
		config := cloudFoundryCreateServiceKeyOptions{
			CfAPIEndpoint:      "https://api.endpoint.com",
			CfOrg:              "testOrg",
			CfSpace:            "testSpace",
			Username:           "testUser",
			Password:           "testPassword",
			CfServiceInstance:  "testInstance",
			CfServiceKeyName:   "testKey",
			CfServiceKeyConfig: "{\"scenario_id\":\"SAP_COM_0510\",\"type\":\"basic\"}",
			CfAsync:            false,
		}
		execRunner := mock.ExecMockRunner{}
		cfUtilsMock := cloudfoundry.CfUtilsMock{}
		defer cfUtilsMock.Cleanup()

		error := runCloudFoundryCreateServiceKey(&config, &telemetryData, &execRunner, &cfUtilsMock)
		if error == nil {
			assert.Equal(t, "cf", execRunner.Calls[0].Exec)
			assert.Equal(t, []string{"create-service-key", "testInstance", "testKey", "-c", "{\"scenario_id\":\"SAP_COM_0510\",\"type\":\"basic\"}", cfCliSynchronousRequestFlag}, execRunner.Calls[0].Params)
		}
	})
}

func TestCloudFoundryCreateServiceKeyErrorMessages(t *testing.T) {
	errorMessage := "errorMessage"
	var telemetryData telemetry.CustomData
	t.Run("CF Login Error", func(t *testing.T) {
		config := cloudFoundryCreateServiceKeyOptions{
			CfAPIEndpoint:      "https://api.endpoint.com",
			CfOrg:              "testOrg",
			CfSpace:            "testSpace",
			Username:           "testUser",
			Password:           "testPassword",
			CfServiceInstance:  "testInstance",
			CfServiceKeyName:   "testKey",
			CfServiceKeyConfig: "{\"scenario_id\":\"SAP_COM_0510\",\"type\":\"basic\"}",
			CfAsync:            true,
		}
		execRunner := mock.ExecMockRunner{}
		cfUtilsMock := cloudfoundry.CfUtilsMock{
			LoginError: errors.New(errorMessage),
		}
		defer cfUtilsMock.Cleanup()

		error := runCloudFoundryCreateServiceKey(&config, &telemetryData, &execRunner, &cfUtilsMock)
		assert.Equal(t, error.Error(), "Error while logging in occurred: "+errorMessage, "Wrong error message")
	})

	t.Run("CF Logout Error", func(t *testing.T) {
		config := cloudFoundryCreateServiceKeyOptions{
			CfAPIEndpoint:      "https://api.endpoint.com",
			CfOrg:              "testOrg",
			CfSpace:            "testSpace",
			Username:           "testUser",
			Password:           "testPassword",
			CfServiceInstance:  "testInstance",
			CfServiceKeyName:   "testKey",
			CfServiceKeyConfig: "{\"scenario_id\":\"SAP_COM_0510\",\"type\":\"basic\"}",
			CfAsync:            true,
		}
		execRunner := mock.ExecMockRunner{}
		cfUtilsMock := cloudfoundry.CfUtilsMock{
			LogoutError: errors.New(errorMessage),
		}
		defer cfUtilsMock.Cleanup()

		err := runCloudFoundryCreateServiceKey(&config, &telemetryData, &execRunner, &cfUtilsMock)
		assert.Equal(t, err.Error(), "Error while logging out occurred: "+errorMessage, "Wrong error message")
	})

	t.Run("CF Create Service Key Error", func(t *testing.T) {
		errorMessage := "errorMessage"
		config := cloudFoundryCreateServiceKeyOptions{
			CfAPIEndpoint:      "https://api.endpoint.com",
			CfOrg:              "testOrg",
			CfSpace:            "testSpace",
			Username:           "testUser",
			Password:           "testPassword",
			CfServiceInstance:  "testInstance",
			CfServiceKeyName:   "testKey",
			CfServiceKeyConfig: "{\"scenario_id\":\"SAP_COM_0510\",\"type\":\"basic\"}",
			CfAsync:            false,
		}
		m := make(map[string]error)
		m["cf create-service-key testInstance testKey -c {\"scenario_id\":\"SAP_COM_0510\",\"type\":\"basic\"} --wait"] = errors.New(errorMessage)
		execRunner := mock.ExecMockRunner{
			ShouldFailOnCommand: m,
		}
		cfUtilsMock := cloudfoundry.CfUtilsMock{
			LogoutError: errors.New(errorMessage),
		}
		defer cfUtilsMock.Cleanup()

		error := runCloudFoundryCreateServiceKey(&config, &telemetryData, &execRunner, &cfUtilsMock)
		assert.Equal(t, error.Error(), "Failed to Create Service Key: "+errorMessage, "Wrong error message")
	})
}
