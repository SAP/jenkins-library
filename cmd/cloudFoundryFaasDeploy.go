package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/npm"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func cloudFoundryFaasDeploy(config cloudFoundryFaasDeployOptions, telemetryData *telemetry.CustomData) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	cfUtils := cloudfoundry.CFUtils{
		Exec: &c,
	}

	npmExecutorOptions := npm.ExecutorOptions{DefaultNpmRegistry: config.DefaultNpmRegistry, ExecRunner: &c}
	npmExecutor := npm.NewExecutor(npmExecutorOptions)

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runCloudFoundryFaasDeploy(&config, telemetryData, &c, &cfUtils, npmExecutor)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runCloudFoundryFaasDeploy(options *cloudFoundryFaasDeployOptions,
	telemetryData *telemetry.CustomData,
	c command.ExecRunner,
	cfUtils cloudfoundry.AuthenticationUtils,
	npmExecutor npm.Executor) (returnedError error) {
	// Login via cf cli
	config := cloudfoundry.LoginOptions{
		CfAPIEndpoint: options.CfAPIEndpoint,
		CfOrg:         options.CfOrg,
		CfSpace:       options.CfSpace,
		Username:      options.Username,
		Password:      options.Password,
	}
	loginErr := cfUtils.Login(config)
	if loginErr != nil {
		return fmt.Errorf("Error while logging in occured: %w", loginErr)
	}
	defer func() {
		logoutErr := cfUtils.Logout()
		if logoutErr != nil && returnedError == nil {
			returnedError = fmt.Errorf("Error while logging out occured: %w", logoutErr)
		}
	}()

	serviceInstance := options.XfsRuntimeServiceInstance
	serviceKey := options.XfsRuntimeServiceKeyName
	log.Entry().Infof("Logging into Extension Factory Serverless Runtime service instance '%s' with service key '%s'", serviceInstance, serviceKey)
	xfsrtLoginScript := []string{"login", "-s", serviceInstance, "-b", serviceKey, "--silent"}
	if err := c.RunExecutable("xfsrt-cli", xfsrtLoginScript...); err != nil {
		return fmt.Errorf("Failed to log in to xfsrt service instance '%s' with service key '%s': %w", serviceInstance, serviceKey, err)
	}
	log.Entry().Info("Logged in successfully to Extension Factory Serverless Runtime service.")

	if err := npmExecutor.InstallAllDependencies(npmExecutor.FindPackageJSONFiles()); err != nil {
		message := "Failed to install npm dependencies"
		registry := options.DefaultNpmRegistry
		if registry != "" {
			message += fmt.Sprintf(" with default registry set to '%s'", registry)
		}
		return fmt.Errorf("%s: %w", message, err)
	}

	log.Entry().Info("Deploying faas project to Extension Factory Serverless Runtime")
	xfsrtDeployScript := []string{"faas", "project", "deploy", "-y", "./deploy/values.yaml"}
	if err := c.RunExecutable("xfsrt-cli", xfsrtDeployScript...); err != nil {
		return fmt.Errorf("Failed to deploy faas project: %w", err)
	} else {
		log.Entry().Info("Deployment successful.")
	}

	return returnedError
}
