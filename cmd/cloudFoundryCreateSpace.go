package cmd

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func cloudFoundryCreateSpace(config cloudFoundryCreateSpaceOptions, telemetryData *telemetry.CustomData) {

	cf := cloudfoundry.CFUtils{Exec: &command.Command{}}
	c := command.Command{}

	err := runCloudFoundryCreateSpace(&config, telemetryData, cf, &c)

	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}

}

func runCloudFoundryCreateSpace(config *cloudFoundryCreateSpaceOptions, telemetryData *telemetry.CustomData, cf cloudfoundry.CFUtils, s command.ShellRunner) (err error) {

	var c = cf.Exec

	cfLogin := s.RunShell("/bin/bash", fmt.Sprintf("yes '' | cf login -a %s -u %s -p %s", config.CfAPIEndpoint, config.Username, config.Pasword))

	if cfLogin != nil {
		return fmt.Errorf("Error while logging in occured: %w", err)
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
