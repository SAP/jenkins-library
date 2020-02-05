package cmd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCloudFoundryDeleteService(t *testing.T) {
	execRunner := execMockRunner{}

	t.Run("CF Login: success case", func(t *testing.T) {
		config := cloudFoundryDeleteServiceOptions{
			CfAPIEndpoint: "https://api.endpoint.com",
			CfOrg:         "testOrg",
			CfSpace:       "testSpace",
			Username:      "testUser",
			Password:      "testPassword",
		}
		error := cloudFoundryLogin(config, &execRunner)
		if error == nil {
			assert.Equal(t, "cf", execRunner.calls[0].exec)
			assert.Equal(t, "login", execRunner.calls[0].params[0])
			assert.Equal(t, "-a", execRunner.calls[0].params[1])
			assert.Equal(t, "https://api.endpoint.com", execRunner.calls[0].params[2])
			assert.Equal(t, "-o", execRunner.calls[0].params[3])
			assert.Equal(t, "testOrg", execRunner.calls[0].params[4])
			assert.Equal(t, "-s", execRunner.calls[0].params[5])
			assert.Equal(t, "testSpace", execRunner.calls[0].params[6])
			assert.Equal(t, "-u", execRunner.calls[0].params[7])
			assert.Equal(t, "testUser", execRunner.calls[0].params[8])
			assert.Equal(t, "-p", execRunner.calls[0].params[9])
			assert.Equal(t, "testPassword", execRunner.calls[0].params[10])
		}
	})
	t.Run("CF Delete Service: Success case", func(t *testing.T) {
		ServiceName := "testInstance"
		error := cloudFoundryDeleteServiceFunction(ServiceName, &execRunner)
		if error == nil {
			assert.Equal(t, "cf", execRunner.calls[1].exec)
			assert.Equal(t, "delete-service", execRunner.calls[1].params[0])
			assert.Equal(t, "testInstance", execRunner.calls[1].params[1])
			assert.Equal(t, "-f", execRunner.calls[1].params[2])
		}
	})
	t.Run("CF Logout: Success case", func(t *testing.T) {
		error := cloudFoundryLogout(&execRunner)
		if error == nil {
			assert.Equal(t, "cf", execRunner.calls[2].exec)
			assert.Equal(t, "logout", execRunner.calls[2].params[0])
		}
	})
	t.Run("CF Delete Service: Error case", func(t *testing.T) {
		ServiceName := "testInstance"
		error := cloudFoundryDeleteServiceFunction(ServiceName, &execRunner)
		if error == nil {
		}
	})
}
