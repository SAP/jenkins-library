package cmd

import (
	"fmt"
	"reflect"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func cloudFoundryCreateServiceKey(options cloudFoundryCreateServiceKeyOptions, telemetryData *telemetry.CustomData) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())

	config := cloudFoundryDeleteServiceOptions{
		CfAPIEndpoint:     options.CfAPIEndpoint,
		CfOrg:             options.CfOrg,
		CfSpace:           options.CfSpace,
		Username:          options.Username,
		Password:          options.Password,
		CfServiceInstance: options.CfServiceInstance,
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
			Fatal("Error occured.")
	}
}

func runCloudFoundryCreateServiceKey(config *cloudFoundryCreateServiceKeyOptions, telemetryData *telemetry.CustomData, c execRunner) error {

	log.Entry().Info("Creating Service Key")

	var cfFindServiceKeysScript []string

	if config.CfServiceKeyConfig == "" {
		cfFindServiceKeysScript = []string{"create-service-key", config.CfServiceInstance, config.CfServiceKeyName}
	} else {
		fmt.Println(reflect.TypeOf(config.CfServiceKeyConfig))
		fmt.Println(config.CfServiceKeyConfig)

		cfFindServiceKeysScript = []string{"create-service-key", config.CfServiceInstance, config.CfServiceKeyName, "[-c", config.CfServiceKeyConfig, "]"}
		fmt.Println(cfFindServiceKeysScript)
	}

	err := c.RunExecutable("cf", cfFindServiceKeysScript...)

	if err != nil {
		return fmt.Errorf("Failed to Create Service Key: %w", err)
	}
	return nil
}
