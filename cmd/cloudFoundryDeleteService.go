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

		cloudFoundryLogin(CloudFoundryDeleteServiceOptions, &c)
		cloudFoundryDeleteServiceFunction(CloudFoundryDeleteServiceOptions.ServiceName, &c)
		cloudFoundryLogout(&c)

		return nil
	}
}

func cloudFoundryLogin(CloudFoundryDeleteServiceOptions cloudFoundryDeleteServiceOptions, c execRunner) error {
	var cfLoginScript = []string{"login", "-a", CloudFoundryDeleteServiceOptions.API, "-o", CloudFoundryDeleteServiceOptions.Organisation, "-s", CloudFoundryDeleteServiceOptions.Space, "-u", CloudFoundryDeleteServiceOptions.Username, "-p", CloudFoundryDeleteServiceOptions.Password}

	log.Entry().WithField("cfAPI:", CloudFoundryDeleteServiceOptions.API).WithField("cfOrg", CloudFoundryDeleteServiceOptions.Organisation).WithField("space", CloudFoundryDeleteServiceOptions.Space).Info("Logging into Cloud Foundry..")

	err := c.RunExecutable("cf", cfLoginScript...)
	if err != nil {
		log.Entry().
			WithError(err).
			Fatal("Failed to login to Cloud Foundry")
	}
	log.Entry().Info("Logged in successfully to Cloud Foundry..")
	return err
}

func cloudFoundryDeleteServiceFunction(service string, c execRunner) error {
	var cfdeleteServiceScript = []string{"delete-service", service, "-f"}

	log.Entry().WithField("cfService", service).Info("Deleting the requested Service")

	err := c.RunExecutable("cf", cfdeleteServiceScript...)
	if err != nil {
		log.Entry().
			WithError(err).
			Fatal("Failed to delete Service")
	}
	log.Entry().Info("Deletion of Service is finished or Service has never existed before, thus can't need to be deleted")
	return err
}

func cloudFoundryLogout(c execRunner) error {
	var cfLogoutScript = "logout"

	log.Entry().Info("Logging out of Cloud Foundry")

	err := c.RunExecutable("cf", cfLogoutScript)
	if err != nil {
		log.Entry().
			WithError(err).
			Fatal("Failed to Logout of Cloud Foudnry")
	}
	log.Entry().Info("Logged out successfully")
	return err
}
