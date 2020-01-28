package cmd

import (
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
)

func cloudFoundryDeleteService(CloudFoundryDeleteServiceOptions cloudFoundryDeleteServiceOptions) error {

	if CloudFoundryDeleteServiceOptions.API == "" || CloudFoundryDeleteServiceOptions.Organisation == "" || CloudFoundryDeleteServiceOptions.Space == "" || CloudFoundryDeleteServiceOptions.Username == "" || CloudFoundryDeleteServiceOptions.Password == "" {
		return errors.New("Parameters missing. Please provide EITHER the Cloud Foundry ApiEndpoint, Organization, Space, Username or Password!")
	} else if CloudFoundryDeleteServiceOptions.ServiceName == "" {
		return errors.New("Parameter missing. Please provide the Name of the Service Instance you want to delete!")
	} else {
		c := command.Command{}

		// reroute command output to logging framework
		c.Stdout(log.Entry().Writer())
		c.Stderr(log.Entry().Writer())

		cfloginconfig := [...]string{CloudFoundryDeleteServiceOptions.API, CloudFoundryDeleteServiceOptions.Organisation, CloudFoundryDeleteServiceOptions.Space, CloudFoundryDeleteServiceOptions.Username, CloudFoundryDeleteServiceOptions.Password}

		cloudFoundryLogin(cfloginconfig, &c)
		cloudFoundryDeleteServiceFunction(CloudFoundryDeleteServiceOptions.ServiceName, &c)
		cloudFoundryLogout(&c)

		return nil
	}
}

func cloudFoundryLogin(loginconfig [5]string, c shellRunner) error {
	var cfLoginScript = "cf login -a " + loginconfig[0] + " -o " + loginconfig[1] + " -s " + loginconfig[2] + " -u " + loginconfig[3] + " -p " + loginconfig[4]

	log.Entry().WithField("cfAPI:", loginconfig[0]).WithField("cfOrg", loginconfig[1]).WithField("space", loginconfig[2]).Info("Logging into Cloud Foundry..")

	err := c.RunShell("/bin/bash", cfLoginScript)
	if err != nil {
		log.Entry().
			WithError(err).
			Fatal("Failed to login to Cloud Foundry")
	}
	log.Entry().Info("Logged in successfully to Cloud Foundry..")
	return err
}

func cloudFoundryDeleteServiceFunction(service string, c shellRunner) error {
	var cfdeleteServiceScript = "cf delete-service " + service + " -f"

	log.Entry().WithField("cfService", service).Info("Deleting the requested Service")

	err := c.RunShell("/bin/bash", cfdeleteServiceScript)
	if err != nil {
		log.Entry().
			WithError(err).
			Fatal("Failed to delete Service")
	}
	log.Entry().Info("Deletion of Service is finished or Service has never existed before, thus can't need to be deleted")
	return err
}

func cloudFoundryLogout(c shellRunner) error {
	var cfLogoutScript = "cf logout"

	log.Entry().Info("Logging out of Cloud Foundry")

	err := c.RunShell("/bin/bash", cfLogoutScript)
	if err != nil {
		log.Entry().
			WithError(err).
			Fatal("Failed to Logout of Cloud Foudnry")
	}
	log.Entry().Info("Logged out successfully")
	return err
}
