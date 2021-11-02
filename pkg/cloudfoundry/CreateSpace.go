package cloudfoundry

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type CloudFoundryCreateSpaceOptions struct {
	CfAPIEndpoint string `json:"cfApiEndpoint,omitempty"`
	Username      string `json:"username,omitempty"`
	Password      string `json:"password,omitempty"`
	CfOrg         string `json:"cfOrg,omitempty"`
	CfSpace       string `json:"cfSpace,omitempty"`
}

func CreateSpace(config *CloudFoundryCreateSpaceOptions, telemetryData *telemetry.CustomData, cf CFUtils, s command.ShellRunner) (err error) {
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
