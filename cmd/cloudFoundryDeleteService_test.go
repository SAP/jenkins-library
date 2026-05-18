//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestCloudFoundryDeleteService(t *testing.T) {

	t.Run("CF Delete Service without service keys: success case", func(t *testing.T) {
		config := cloudFoundryDeleteServiceOptions{
			CfAPIEndpoint:       "https://api.endpoint.com",
			CfOrg:               "testOrg",
			CfSpace:             "testSpace",
			Username:            "testUser",
			Password:            "testPassword",
			CfServiceInstance:   "testInstance",
			CfDeleteServiceKeys: false,
			CfAsync:             false,
		}
		m := make(map[string]string)
		execRunner := mock.ExecMockRunner{
			StdoutReturn: m,
		}
		cfUtils := cloudfoundry.CfUtilsMock{}

		err := runCloudFoundryDeleteService(&config, &execRunner, &cfUtils)
		if assert.NoError(t, err) {
			assert.Equal(t, "cf", execRunner.Calls[0].Exec)
			assert.Equal(t, []string{"delete-service", "testInstance", "-f", "--wait"}, execRunner.Calls[0].Params)

		}
	})

	t.Run("CF Delete Service without service keys async: success case", func(t *testing.T) {
		config := cloudFoundryDeleteServiceOptions{
			CfAPIEndpoint:       "https://api.endpoint.com",
			CfOrg:               "testOrg",
			CfSpace:             "testSpace",
			Username:            "testUser",
			Password:            "testPassword",
			CfServiceInstance:   "testInstance",
			CfDeleteServiceKeys: false,
			CfAsync:             true,
		}
		m := make(map[string]string)
		execRunner := mock.ExecMockRunner{
			StdoutReturn: m,
		}
		cfUtils := cloudfoundry.CfUtilsMock{}

		err := runCloudFoundryDeleteService(&config, &execRunner, &cfUtils)
		if assert.NoError(t, err) {
			assert.Equal(t, "cf", execRunner.Calls[0].Exec)
			assert.Equal(t, []string{"delete-service", "testInstance", "-f"}, execRunner.Calls[0].Params)

		}
	})

	t.Run("CF Delete Service : success case", func(t *testing.T) {
		config := cloudFoundryDeleteServiceOptions{
			CfAPIEndpoint:       "https://api.endpoint.com",
			CfOrg:               "testOrg",
			CfSpace:             "testSpace",
			Username:            "testUser",
			Password:            "testPassword",
			CfServiceInstance:   "testInstance",
			CfDeleteServiceKeys: true,
			CfAsync:             false,
		}
		m := make(map[string]string)
		m["cf service testInstance --guid"] = `instance-guid`
		m["cf curl /v3/service_credential_bindings?service_instance_guids=instance-guid"] = `{
"resources": [
	{ "name": "ExampleServiceKey1", "type": "key"  },
    { "name": "ExampleServiceKey2", "type": "application"  },
    { "name": "ExampleServiceKey3", "type": "key"  }
]
}`

		execRunner := mock.ExecMockRunner{
			StdoutReturn: m,
		}
		cfUtils := cloudfoundry.CfUtilsMock{}

		err := runCloudFoundryDeleteService(&config, &execRunner, &cfUtils)
		if assert.NoError(t, err) {
			assert.Equal(t, "cf", execRunner.Calls[0].Exec)
			assert.Equal(t, "cf", execRunner.Calls[1].Exec)
			assert.Equal(t, "cf", execRunner.Calls[2].Exec)
			assert.Equal(t, "cf", execRunner.Calls[3].Exec)
			assert.Equal(t, []string{"service", "testInstance", "--guid"}, execRunner.Calls[0].Params)
			assert.Equal(t, []string{"curl", "/v3/service_credential_bindings?service_instance_guids=instance-guid"}, execRunner.Calls[1].Params)
			assert.Equal(t, []string{"delete-service-key", "testInstance", "ExampleServiceKey1", "-f", "--wait"}, execRunner.Calls[2].Params)
			assert.Equal(t, []string{"delete-service-key", "testInstance", "ExampleServiceKey3", "-f", "--wait"}, execRunner.Calls[3].Params)
			assert.Equal(t, []string{"delete-service", "testInstance", "-f", "--wait"}, execRunner.Calls[4].Params)
		}
	})
	t.Run("CF Delete Service async : success case", func(t *testing.T) {
		config := cloudFoundryDeleteServiceOptions{
			CfAPIEndpoint:       "https://api.endpoint.com",
			CfOrg:               "testOrg",
			CfSpace:             "testSpace",
			Username:            "testUser",
			Password:            "testPassword",
			CfServiceInstance:   "testInstance",
			CfDeleteServiceKeys: true,
			CfAsync:             true,
		}
		m := make(map[string]string)
		m["cf service testInstance --guid"] = `instance-guid`
		m["cf curl /v3/service_credential_bindings?service_instance_guids=instance-guid"] = `{
"resources": [
	{ "name": "ExampleServiceKey1", "type": "key" },
    { "name": "ExampleServiceKey2", "type": "application" },
    { "name": "ExampleServiceKey3", "type": "key" }
]
}`

		execRunner := mock.ExecMockRunner{
			StdoutReturn: m,
		}
		cfUtils := cloudfoundry.CfUtilsMock{}

		err := runCloudFoundryDeleteService(&config, &execRunner, &cfUtils)
		if assert.NoError(t, err) {
			assert.Equal(t, "cf", execRunner.Calls[0].Exec)
			assert.Equal(t, "cf", execRunner.Calls[1].Exec)
			assert.Equal(t, "cf", execRunner.Calls[2].Exec)
			assert.Equal(t, "cf", execRunner.Calls[3].Exec)
			assert.Equal(t, []string{"service", "testInstance", "--guid"}, execRunner.Calls[0].Params)
			assert.Equal(t, []string{"curl", "/v3/service_credential_bindings?service_instance_guids=instance-guid"}, execRunner.Calls[1].Params)
			assert.Equal(t, []string{"delete-service-key", "testInstance", "ExampleServiceKey1", "-f"}, execRunner.Calls[2].Params)
			assert.Equal(t, []string{"delete-service-key", "testInstance", "ExampleServiceKey3", "-f"}, execRunner.Calls[3].Params)
			assert.Equal(t, []string{"delete-service", "testInstance", "-f"}, execRunner.Calls[4].Params)
		}
	})
}
