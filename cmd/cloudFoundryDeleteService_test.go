//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestCloudFoundryDeleteService(t *testing.T) {
	newOptions := func(deleteKeys, async bool) cloudFoundryDeleteServiceOptions {
		return cloudFoundryDeleteServiceOptions{
			CfAPIEndpoint: "https://api.endpoint.com", CfOrg: "testOrg", CfSpace: "testSpace",
			Username: "testUser", Password: "testPassword", CfServiceInstance: "testInstance",
			CfDeleteServiceKeys: deleteKeys, CfAsync: async,
		}
	}
	serviceKeyOutput := map[string]string{
		"cf service testInstance --guid": "instance-guid",
		"cf curl /v3/service_credential_bindings?service_instance_guids=instance-guid": `{"resources":[
			{"name":"ExampleServiceKey1","type":"key"},
			{"name":"ExampleServiceKey2","type":"application"},
			{"name":"ExampleServiceKey3","type":"key"}
		]}`,
	}

	t.Run("deletes a service synchronously without service keys", func(t *testing.T) {
		runner := &mock.ExecMockRunner{}

		err := runCloudFoundryDeleteService(ptr(newOptions(false, false)), runner)

		if assert.NoError(t, err) {
			assert.Equal(t, []string{"delete-service", "testInstance", "-f", "--wait"}, runner.Calls[1].Params)
			assert.Equal(t, []string{"logout"}, runner.Calls[2].Params)
		}
	})

	t.Run("deletes a service asynchronously without service keys", func(t *testing.T) {
		runner := &mock.ExecMockRunner{}

		err := runCloudFoundryDeleteService(ptr(newOptions(false, true)), runner)

		if assert.NoError(t, err) {
			assert.Equal(t, []string{"delete-service", "testInstance", "-f"}, runner.Calls[1].Params)
		}
	})

	for _, test := range []struct {
		name   string
		async  bool
		key    []string
		delete []string
	}{
		{"deletes service keys synchronously", false, []string{"delete-service-key", "testInstance", "ExampleServiceKey1", "-f", "--wait"}, []string{"delete-service", "testInstance", "-f", "--wait"}},
		{"deletes service keys asynchronously", true, []string{"delete-service-key", "testInstance", "ExampleServiceKey1", "-f"}, []string{"delete-service", "testInstance", "-f"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			runner := &mock.ExecMockRunner{StdoutReturn: serviceKeyOutput}

			err := runCloudFoundryDeleteService(ptr(newOptions(true, test.async)), runner)

			if assert.NoError(t, err) {
				assert.Equal(t, []string{"login", "-a", "https://api.endpoint.com", "-o", "testOrg", "-s", "testSpace", "-u", "testUser", "-p", "testPassword"}, runner.Calls[0].Params)
				assert.Equal(t, []string{"service", "testInstance", "--guid"}, runner.Calls[1].Params)
				assert.Equal(t, []string{"curl", "/v3/service_credential_bindings?service_instance_guids=instance-guid"}, runner.Calls[2].Params)
				assert.Equal(t, test.key, runner.Calls[3].Params)
				secondKey := append([]string(nil), test.key...)
				secondKey[2] = "ExampleServiceKey3"
				assert.Equal(t, secondKey, runner.Calls[4].Params)
				assert.Equal(t, test.delete, runner.Calls[5].Params)
				assert.Equal(t, []string{"logout"}, runner.Calls[6].Params)
			}
		})
	}
}

func ptr[T any](value T) *T { return &value }
