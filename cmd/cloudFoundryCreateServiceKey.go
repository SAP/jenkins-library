package cmd

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func cloudFoundryCreateServiceKey(options cloudFoundryCreateServiceKeyOptions, telemetryData *telemetry.CustomData) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	cfUtils := cloudfoundry.CFUtils{
		Exec: &c,
	}

	err := runCloudFoundryCreateServiceKey(&options, telemetryData, &c, &cfUtils)
	if err != nil {
		log.Entry().
			WithError(err).
			Fatal("Error occurred during step.")
	}
}

func runCloudFoundryCreateServiceKey(options *cloudFoundryCreateServiceKeyOptions, telemetryData *telemetry.CustomData, c command.ExecRunner, cfUtils cloudfoundry.AuthenticationUtils) (returnedError error) {

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
		return fmt.Errorf("Error while logging in occurred: %w", loginErr)
	}
	defer func() {
		logoutErr := cfUtils.Logout()
		if logoutErr != nil && returnedError == nil {
			returnedError = fmt.Errorf("Error while logging out occurred: %w", logoutErr)
		}
	}()
	log.Entry().Info("Creating Service Key")

	var cfCreateServiceKeyScript []string
	if options.CfServiceKeyConfig == "" {
		cfCreateServiceKeyScript = []string{"create-service-key", options.CfServiceInstance, options.CfServiceKeyName, "--wait"}
	} else {
		cfCreateServiceKeyScript = []string{"create-service-key", options.CfServiceInstance, options.CfServiceKeyName, "-c", options.CfServiceKeyConfig, "--wait"}
	}
	err := c.RunExecutable("cf", cfCreateServiceKeyScript...)
	if err != nil {
		return fmt.Errorf("Failed to Create Service Key: %w", err)
	}

	return returnedError
}
