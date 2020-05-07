package cmd

import (
	"fmt"

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

	config := cloudFoundryDeleteServiceOptions{
		CfAPIEndpoint: options.CfAPIEndpoint,
		CfOrg:         options.CfOrg,
		CfSpace:       options.CfSpace,
		Username:      options.Username,
		Password:      options.Password,
	}

	var err error

	err = cloudFoundryLogin(config, &c)

	if err == nil {
		err = runCloudFoundryCreateServiceKey(&options, telemetryData, &c)
	}

	var logoutErr error

	if err == nil {
		logoutErr = cloudFoundryLogout(&c)
		if logoutErr != nil {
			log.Entry().
				WithError(logoutErr).
				Fatal("Error while logging out occured.")
		}
	} else if err != nil {
		logoutErr = cloudFoundryLogout(&c)
		if logoutErr != nil {
			log.Entry().
				WithError(logoutErr).
				Fatal("Error while logging out occured.")
		}
		log.Entry().
			WithError(err).
			Fatal("Error occured during step.")
	}
}

func runCloudFoundryCreateServiceKey(config *cloudFoundryCreateServiceKeyOptions, telemetryData *telemetry.CustomData, c execRunner) error {

	log.Entry().Info("Creating Service Key")

	var cfCreateServiceKeyScript []string

	if config.CfServiceKeyConfig == "" {
		cfCreateServiceKeyScript = []string{"create-service-key", config.CfServiceInstance, config.CfServiceKeyName}
	} else {
		cfCreateServiceKeyScript = []string{"create-service-key", config.CfServiceInstance, config.CfServiceKeyName, "-c", config.CfServiceKeyConfig}
	}
	err := c.RunExecutable("cf", cfCreateServiceKeyScript...)
	if err != nil {
		return fmt.Errorf("Failed to Create Service Key: %w", err)
	}
	return nil
}
