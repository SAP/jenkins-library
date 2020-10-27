package cmd

import (
	"fmt"
	"strings"
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

	t.Run("CF Deploy Faas without deploy values: Success case", func(t *testing.T) {
		defer func() {
			cfUtilsMock.Cleanup()
			execRunner.Calls = nil
			execRunner.ShouldFailOnCommand = nil
		}()

		gotError := runCloudFoundryFaasDeploy(&config, &telemetryData, &execRunner, &cfUtilsMock, &npm.Execute{Utils: &npmUtilsMock})
		if assert.NoError(t, gotError) {
			assert.Equal(t, "xfsrt-cli", execRunner.Calls[0].Exec)
			assert.Equal(t, []string{"login", "-s", "testInstance", "-b", "testKey", "--silent"}, execRunner.Calls[0].Params)
			assert.Equal(t, "xfsrt-cli", execRunner.Calls[1].Exec)
			assert.Equal(t, []string{"faas", "project", "deploy"}, execRunner.Calls[1].Params)
		}
	})

	t.Run("CF Deploy Faas with deploy values: Success case", func(t *testing.T) {
		defer func() {
			cfUtilsMock.Cleanup()
			execRunner.Calls = nil
			execRunner.ShouldFailOnCommand = nil
			config.XfsrtValues = ""
		}()

		config.XfsrtValues = `
{
   "secret-values": {
      "credentials": {
	 "username": "xxx",
	 "password": "yyy"
      }
   }
}`
		deployValues := strings.ReplaceAll(config.XfsrtValues, "\n", " ")

		gotError := runCloudFoundryFaasDeploy(&config, &telemetryData, &execRunner, &cfUtilsMock, &npm.Execute{Utils: &npmUtilsMock})
		if assert.NoError(t, gotError) {
			assert.Equal(t, "xfsrt-cli", execRunner.Calls[0].Exec)
			assert.Equal(t, []string{"login", "-s", "testInstance", "-b", "testKey", "--silent"}, execRunner.Calls[0].Params)
			assert.Equal(t, "xfsrt-cli", execRunner.Calls[1].Exec)
			assert.Equal(t, []string{"faas", "project", "deploy", "-c", deployValues}, execRunner.Calls[1].Params)
		}
	})

	t.Run("CF Login Error", func(t *testing.T) {
		defer func() {
			cfUtilsMock.Cleanup()
			execRunner.Calls = nil
			execRunner.ShouldFailOnCommand = nil
		}()

		errorMessage := "cf login error"

		cfUtilsMock.LoginError = errors.New(errorMessage)

		gotError := runCloudFoundryFaasDeploy(&config, &telemetryData, &execRunner, &cfUtilsMock, &npm.Execute{Utils: &npmUtilsMock})
		assert.EqualError(t, gotError, "Error while logging in occured: "+errorMessage, "Wrong error message")
	})

	t.Run("xfsrt Login Error", func(t *testing.T) {
		defer func() {
			cfUtilsMock.Cleanup()
			execRunner.Calls = nil
			execRunner.ShouldFailOnCommand = nil
		}()

		errorMessage := "xfsrt login error"

		execRunner.ShouldFailOnCommand = map[string]error{"xfsrt-cli login -s testInstance -b testKey --silent": fmt.Errorf(errorMessage)}

		gotError := runCloudFoundryFaasDeploy(&config, &telemetryData, &execRunner, &cfUtilsMock, &npm.Execute{Utils: &npmUtilsMock})
		assert.EqualError(t, gotError, "Failed to log in to xfsrt service instance 'testInstance' with service key 'testKey': "+errorMessage, "Wrong error message")
	})

	t.Run("xfsrt Deployment Failure", func(t *testing.T) {
		defer func() {
			cfUtilsMock.Cleanup()
			execRunner.Calls = nil
			execRunner.ShouldFailOnCommand = nil
		}()

		errorMessage := "xfsrt deployment failure"

		execRunner.ShouldFailOnCommand = map[string]error{"xfsrt-cli faas project deploy": fmt.Errorf(errorMessage)}

		gotError := runCloudFoundryFaasDeploy(&config, &telemetryData, &execRunner, &cfUtilsMock, &npm.Execute{Utils: &npmUtilsMock})
		assert.EqualError(t, gotError, "Failed to deploy faas project: "+errorMessage, "Wrong error message")
	})

}
