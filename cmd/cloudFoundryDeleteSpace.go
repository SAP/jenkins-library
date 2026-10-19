package cmd

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func cloudFoundryDeleteSpace(config cloudFoundryDeleteSpaceOptions, telemetryData *telemetry.CustomData) {

	c := command.Command{}

	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	err := runCloudFoundryDeleteSpace(&config, telemetryData, &c, &c)

	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runCloudFoundryDeleteSpace(config *cloudFoundryDeleteSpaceOptions, telemetryData *telemetry.CustomData, c command.ExecRunner, s command.ShellRunner) (err error) {
	cfLoginError := s.RunShell("/bin/sh", fmt.Sprintf("yes '' | cf login -a %s -u %s -p %s", config.CfAPIEndpoint, config.Username, config.Password))

	if cfLoginError != nil {
		return fmt.Errorf("Error while logging in occured: %w", cfLoginError)
	}

	defer func() {
		logoutErr := cloudfoundry.Logout(c)
		if logoutErr != nil {
			err = fmt.Errorf("Error while logging out occured: %w", logoutErr)
		}
	}()

	log.Entry().Infof("Deleting Cloud Foundry Space: '%s'", config.CfSpace)

	cfDeleteSpaceScript := []string{"delete-space", config.CfSpace, "-o", config.CfOrg, "-f"}

	err = c.RunExecutable("cf", cfDeleteSpaceScript...)

	if err != nil {
		return fmt.Errorf("Deletion of cf space has failed: %w", err)
	}

	log.Entry().Info("Cloud foundry space has been deleted successfully")

	return err
}
