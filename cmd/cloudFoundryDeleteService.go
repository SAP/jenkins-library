package cmd

import (
	"bytes"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"strings"
)

func cloudFoundryDeleteService(options cloudFoundryDeleteServiceOptions, telemetryData *telemetry.CustomData) error {

	c := command.Command{}

	// reroute command output to logging framework
	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())

	cloudFoundryLogin(options, &c)

	if options.CfDeleteServiceKeys == true {
		log.Entry().Info("Deleting inherent Service Keys")
		cloudFoundryDeleteServiceKeys(options, &c)
	}

	cloudFoundryDeleteServiceFunction(options.CfServiceInstance, &c)

	cloudFoundryLogout(&c)

	return nil
}

func cloudFoundryDeleteServiceKeys(options cloudFoundryDeleteServiceOptions, c execRunner) error {

	var cfFindServiceKeysScript = []string{"service-keys", options.CfServiceInstance}

	var serviceKeyBytes bytes.Buffer
	c.Stdout(&serviceKeyBytes)

	err := c.RunExecutable("cf", cfFindServiceKeysScript...)
	if err != nil {
		cloudFoundryLogout(c)
		log.Entry().
			WithError(err).
			Fatal("Failed to Delete Service Key, most likely your service doesn't exist")
	}

	if len(serviceKeyBytes.String()) > 0 {
		var lines []string = strings.Split(serviceKeyBytes.String(), "\n")
		if len(lines) <= 4 {
			log.Entry().Info("No Service Keys active to be deleted")
			return err
		}
		var numberOfLines = len(lines)
		log.Entry().WithField("Number of service keys :", numberOfLines-4).Info("ServiceKey")

		//Deleting all matched Service Keys for Service
		for i := 3; i <= numberOfLines-2; i++ {
			log.Entry().WithField("Deleting Service Key", lines[i]).Info("ServiceKeyDeletion")

			var cfDeleteServiceKeyScript = []string{"delete-service-key", options.CfServiceInstance, lines[i], "-f"}

			err := c.RunExecutable("cf", cfDeleteServiceKeyScript...)
			if err != nil {
				cloudFoundryLogout(c)
				log.Entry().
					WithError(err).
					Fatal("Failed to Delete Service Key")
			}
			log.Entry().Info("ServiceKeys have been deleted!")
		}
	} else {
		log.Entry().Info("No service key could be retrieved for your requested Service")
		return err
	}
	return err
}

func cloudFoundryLogin(options cloudFoundryDeleteServiceOptions, c execRunner) error {
	var cfLoginScript = []string{"login", "-a", options.CfAPIEndpoint, "-o", options.CfOrg, "-s", options.CfSpace, "-u", options.Username, "-p", options.Password}

	log.Entry().WithField("cfAPI:", options.CfAPIEndpoint).WithField("cfOrg", options.CfOrg).WithField("space", options.CfSpace).Info("Logging into Cloud Foundry..")

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
		cloudFoundryLogout(c)
		log.Entry().
			WithError(err).
			Fatal("Failed to delete Service!")
	}
	log.Entry().Info("Deletion of Service is finished or the Service has never existed")
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
