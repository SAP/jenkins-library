package cmd

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func cloudFoundryCreateSpace(config cloudFoundryCreateSpaceOptions, telemetryData *telemetry.CustomData) {

	c := command.Command{}
	cf := cloudfoundry.CFUtils{Exec: &command.Command{}}

	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	err := runCloudFoundryCreateSpace(&config, telemetryData, cf, &c)

	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}

}

func runCloudFoundryCreateSpace(config *cloudFoundryCreateSpaceOptions, telemetryData *telemetry.CustomData, cf cloudfoundry.CFUtils, s command.ShellRunner) (err error) {

	var c = cf.Exec

	cfLoginError := s.RunShell("/bin/sh", fmt.Sprintf("yes '' | cf login -a %s -u %s -p %s", config.CfAPIEndpoint, config.Username, config.Password))

	if cfLoginError != nil {
		return fmt.Errorf("Error while logging in occured: %w", cfLoginError)
	}
	log.Entry().Info("Successfully logged into cloud foundry.")

	defer func() {
		logoutErr := cf.Logout()
		if logoutErr != nil {
			err = fmt.Errorf("Error while logging out occured: %w", logoutErr)
		}
	}()

	log.Entry().Infof("Creating Cloud Foundry Space: '%s'", config.CfSpace)

	cfCreateSpaceScript := []string{"create-space", config.CfSpace, "-o", config.CfOrg}

	err = c.RunExecutable("cf", cfCreateSpaceScript...)

	if err != nil {
		return fmt.Errorf("Creating a cf space has failed: %w", err)
	}

	log.Entry().Info("Cloud foundry space has been created successfully")

	return err
}
