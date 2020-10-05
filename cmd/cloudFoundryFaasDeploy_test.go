package cmd

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/npm"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestCloudFoundryFaasDeploy(t *testing.T) {
	var telemetryData telemetry.CustomData

	config := cloudFoundryFaasDeployOptions{
		CfAPIEndpoint:        "https://api.endpoint.com",
		CfOrg:                "testOrg",
		CfSpace:              "testSpace",
		Username:             "testUser",
		Password:             "testPassword",
		XfsrtServiceInstance: "testInstance",
		XfsrtServiceKeyName:  "testKey",
	}
	execRunner := mock.ExecMockRunner{}
	cfUtilsMock := cloudfoundry.CfUtilsMock{}
	npmUtilsMock := npmMockUtilsBundle{FilesMock: &mock.FilesMock{}, execRunner: &execRunner}

	t.Run("CF Deploy Faas: Success case", func(t *testing.T) {
		defer cfUtilsMock.Cleanup()

		gotError := runCloudFoundryFaasDeploy(&config, &telemetryData, &execRunner, &cfUtilsMock, &npm.Execute{Utils: &npmUtilsMock})
		if assert.NoError(t, gotError) {
			assert.Equal(t, "xfsrt-cli", execRunner.Calls[0].Exec)
			assert.Equal(t, []string{"login", "-s", "testInstance", "-b", "testKey", "--silent"}, execRunner.Calls[0].Params)
			assert.Equal(t, "xfsrt-cli", execRunner.Calls[1].Exec)
			assert.Equal(t, []string{"faas", "project", "deploy", "-y", "./deploy/values.yaml"}, execRunner.Calls[1].Params)
		}
	})

	t.Run("CF Login Error", func(t *testing.T) {
		defer cfUtilsMock.Cleanup()
		errorMessage := "cf login error"

		cfUtilsMock.LoginError = errors.New(errorMessage)

		gotError := runCloudFoundryFaasDeploy(&config, &telemetryData, &execRunner, &cfUtilsMock, &npm.Execute{Utils: &npmUtilsMock})
		assert.EqualError(t, gotError, "Error while logging in occured: "+errorMessage, "Wrong error message")
	})

	t.Run("xfsrt Login Error", func(t *testing.T) {
		defer cfUtilsMock.Cleanup()
		errorMessage := "xfsrt login error"

		execRunner.ShouldFailOnCommand = map[string]error{"xfsrt-cli login -s testInstance -b testKey --silent": fmt.Errorf(errorMessage)}

		gotError := runCloudFoundryFaasDeploy(&config, &telemetryData, &execRunner, &cfUtilsMock, &npm.Execute{Utils: &npmUtilsMock})
		assert.EqualError(t, gotError, "Failed to log in to xfsrt service instance 'testInstance' with service key 'testKey': "+errorMessage, "Wrong error message")
	})

	t.Run("xfsrt Deployment Failure", func(t *testing.T) {
		defer cfUtilsMock.Cleanup()
		errorMessage := "xfsrt deployment failure"

		execRunner.ShouldFailOnCommand = map[string]error{"xfsrt-cli faas project deploy -y ./deploy/values.yaml": fmt.Errorf(errorMessage)}

		gotError := runCloudFoundryFaasDeploy(&config, &telemetryData, &execRunner, &cfUtilsMock, &npm.Execute{Utils: &npmUtilsMock})
		assert.EqualError(t, gotError, "Failed to deploy faas project: "+errorMessage, "Wrong error message")
	})

}
