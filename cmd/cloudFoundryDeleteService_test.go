//go:build unit

package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestCloudFoundryDeleteService(t *testing.T) {

	t.Run("CF Delete Service : success case", func(t *testing.T) {
		config := cloudFoundryDeleteServiceOptions{
			CfAPIEndpoint:       "https://api.endpoint.com",
			CfOrg:               "testOrg",
			CfSpace:             "testSpace",
			Username:            "testUser",
			Password:            "testPassword",
			CfServiceInstance:   "testInstance",
			CfDeleteServiceKeys: true,
		}
		m := make(map[string]string)
		m["cf service-keys testInstance"] = `line1
line2
line3
myServiceKey1
myServiceKey2
`
		execRunner := mock.ExecMockRunner{
			StdoutReturn: m,
		}
		cfUtils := cloudfoundry.CfUtilsMock{}

		err := runCloudFoundryDeleteService(config, &execRunner, &cfUtils)
		if assert.NoError(t, err) {
			assert.Equal(t, "cf", execRunner.Calls[0].Exec)
			assert.Equal(t, "cf", execRunner.Calls[1].Exec)
			assert.Equal(t, "cf", execRunner.Calls[2].Exec)
			assert.Equal(t, "cf", execRunner.Calls[3].Exec)
			assert.Equal(t, []string{"service-keys", "testInstance"}, execRunner.Calls[0].Params)
			assert.Equal(t, []string{"delete-service-key", "testInstance", "myServiceKey1", "-f"}, execRunner.Calls[1].Params)
			assert.Equal(t, []string{"delete-service-key", "testInstance", "myServiceKey2", "-f"}, execRunner.Calls[2].Params)
			assert.Equal(t, []string{"delete-service", "testInstance", "-f"}, execRunner.Calls[3].Params)
		}
	})
}
