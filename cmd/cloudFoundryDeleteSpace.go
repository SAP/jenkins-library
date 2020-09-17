package cmd

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func cloudFoundryDeleteSpace(config cloudFoundryDeleteSpaceOptions, telemetryData *telemetry.CustomData) {
	cf := cloudfoundry.CFUtils{Exec: &command.Command{}}

	err := runCloudFoundryDeleteSpace(&config, telemetryData, cf)

	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runCloudFoundryDeleteSpace(config *cloudFoundryDeleteSpaceOptions, telemetryData *telemetry.CustomData, cf cloudfoundry.CFUtils) (err error) {
	var c = cf.Exec

	cfLogin := []string{"login", "-a", config.CfAPIEndpoint, "-u", config.Username, "-p", config.Password}

	//TODO use pipe mechanism to skip the user interference

	err = c.RunExecutable("cf", cfLogin...)

	if err != nil {
		return fmt.Errorf("Error while logging in occured: %w", err)
	}

	defer func() {
		logoutErr := cf.Logout()
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
