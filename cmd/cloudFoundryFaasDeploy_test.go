package cmd

import (
	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/npm"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCloudFoundryFaasDeploy(t *testing.T) {
	var telemetryData telemetry.CustomData

	existingFiles := make(map[string]string)
	existingFiles["package.json"] = "{\"name\": \"myName\", \"version\": \"1.2.3\"}"
	//fileUtils := mock.FilesMock{files: existingFiles}

	t.Run("CF Deploy Faas: Success case", func(t *testing.T) {
		config := cloudFoundryFaasDeployOptions{
			CfAPIEndpoint:             "https://api.endpoint.com",
			CfOrg:                     "testOrg",
			CfSpace:                   "testSpace",
			Username:                  "testUser",
			Password:                  "testPassword",
			XfsRuntimeServiceInstance: "testInstance",
			XfsRuntimeServiceKeyName:  "testKey",
		}
		execRunner := mock.ExecMockRunner{}
		cfUtilsMock := cloudfoundry.CfUtilsMock{}
		defer cfUtilsMock.Cleanup()
		npmUtilsMock := npmMockUtilsBundle{FilesMock: &mock.FilesMock{}, execRunner: &execRunner}

		error := runCloudFoundryFaasDeploy(&config, &telemetryData, &execRunner, &cfUtilsMock, &npm.Execute{Utils: &npmUtilsMock})
		if error == nil {
			assert.Equal(t, "xfsrt-cli", execRunner.Calls[0].Exec)
			assert.Equal(t, []string{"login", "-s", "testInstance", "-b", "testKey", "--silent"}, execRunner.Calls[0].Params)
			assert.Equal(t, "xfsrt-cli", execRunner.Calls[1].Exec)
			assert.Equal(t, []string{"faas", "project", "deploy", "-y", "./deploy/values.yaml"}, execRunner.Calls[1].Params)
		}
	})
}

type FaasTestFileUtilsMock struct {
	existingFiles map[string]string
	writtenFiles  map[string]string
	copiedFiles   map[string]string
}
