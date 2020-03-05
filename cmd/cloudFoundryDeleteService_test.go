package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCloudFoundryDeleteService(t *testing.T) {
	execRunner := mock.ExecMockRunner{}

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
			assert.Equal(t, "cf", execRunner.Calls[0].Exec)
			assert.Equal(t, "login", execRunner.Calls[0].Params[0])
			assert.Equal(t, "-a", execRunner.Calls[0].Params[1])
			assert.Equal(t, "https://api.endpoint.com", execRunner.Calls[0].Params[2])
			assert.Equal(t, "-o", execRunner.Calls[0].Params[3])
			assert.Equal(t, "testOrg", execRunner.Calls[0].Params[4])
			assert.Equal(t, "-s", execRunner.Calls[0].Params[5])
			assert.Equal(t, "testSpace", execRunner.Calls[0].Params[6])
			assert.Equal(t, "-u", execRunner.Calls[0].Params[7])
			assert.Equal(t, "testUser", execRunner.Calls[0].Params[8])
			assert.Equal(t, "-p", execRunner.Calls[0].Params[9])
			assert.Equal(t, "testPassword", execRunner.Calls[0].Params[10])
		}
	})
	t.Run("CF Delete Service: Success case", func(t *testing.T) {
		ServiceName := "testInstance"
		error := cloudFoundryDeleteServiceFunction(ServiceName, &execRunner)
		if error == nil {
			assert.Equal(t, "cf", execRunner.Calls[1].Exec)
			assert.Equal(t, "delete-service", execRunner.Calls[1].Params[0])
			assert.Equal(t, "testInstance", execRunner.Calls[1].Params[1])
			assert.Equal(t, "-f", execRunner.Calls[1].Params[2])
		}
	})
	t.Run("CF Logout: Success case", func(t *testing.T) {
		error := cloudFoundryLogout(&execRunner)
		if error == nil {
			assert.Equal(t, "cf", execRunner.Calls[2].Exec)
			assert.Equal(t, "logout", execRunner.Calls[2].Params[0])
		}
	})
	t.Run("CF Delete Service Keys: success case", func(t *testing.T) {
		config := cloudFoundryDeleteServiceOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			Username:          "testUser",
			Password:          "testPassword",
			CfServiceInstance: "testInstance",
		}
		error := cloudFoundryDeleteServiceKeys(config, &execRunner)
		if error == nil {
			assert.Equal(t, "cf", execRunner.Calls[3].Exec)
			assert.Equal(t, []string{"service-keys", "testInstance"}, execRunner.Calls[3].Params)
		}
	})
}
