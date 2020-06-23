package cmd

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func cloudFoundryDeleteService(options cloudFoundryDeleteServiceOptions, telemetryData *telemetry.CustomData) {

	c := command.Command{}

	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	var err error

	err = cloudFoundryLogin(options, &c)

	if err == nil && options.CfDeleteServiceKeys == true {
		err = cloudFoundryDeleteServiceKeys(options, &c)
	}

	if err == nil {
		err = cloudFoundryDeleteServiceFunction(options.CfServiceInstance, &c)
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

func cloudFoundryDeleteServiceKeys(options cloudFoundryDeleteServiceOptions, c execRunner) error {

	log.Entry().Info("Deleting inherent Service Keys")

	var cfFindServiceKeysScript = []string{"service-keys", options.CfServiceInstance}

	var serviceKeyBytes bytes.Buffer
	c.Stdout(&serviceKeyBytes)

	err := c.RunExecutable("cf", cfFindServiceKeysScript...)

	if err != nil {
		return fmt.Errorf("Failed to Delete Service Key, most likely your service doesn't exist: %w", err)
	}

	if len(serviceKeyBytes.String()) == 0 {
		log.Entry().Info("No service key could be retrieved for your requested Service")
		return nil
	}

	var lines []string = strings.Split(serviceKeyBytes.String(), "\n")
	if len(lines) <= 4 {
		log.Entry().Info("No Service Keys active to be deleted")
		return nil
	}
	var numberOfLines = len(lines)
	log.Entry().WithField("Number of service keys :", numberOfLines-4).Info("ServiceKey")
	//Deleting all matched Service Keys for Service
	for i := 3; i <= numberOfLines-2; i++ {
		log.Entry().WithField("Deleting Service Key", lines[i]).Info("ServiceKeyDeletion")
		var cfDeleteServiceKeyScript = []string{"delete-service-key", options.CfServiceInstance, lines[i], "-f"}
		err := c.RunExecutable("cf", cfDeleteServiceKeyScript...)
		if err != nil {
			return fmt.Errorf("Failed to Delete Service Key: %w", err)
		}
	}
	log.Entry().Info("ServiceKeys have been deleted!")
	return nil
}

func cloudFoundryLogin(options cloudFoundryDeleteServiceOptions, c execRunner) error {
	var cfLoginScript = []string{"login", "-a", options.CfAPIEndpoint, "-o", options.CfOrg, "-s", options.CfSpace, "-u", options.Username, "-p", options.Password}

	log.Entry().WithField("cfAPI:", options.CfAPIEndpoint).WithField("cfOrg", options.CfOrg).WithField("space", options.CfSpace).Info("Logging into Cloud Foundry..")

	err := c.RunExecutable("cf", cfLoginScript...)

	if err != nil {
		return fmt.Errorf("Failed to login to Cloud Foundry: %w", err)
	}
	log.Entry().Info("Logged in successfully to Cloud Foundry..")
	return nil
}

func cloudFoundryDeleteServiceFunction(service string, c execRunner) error {
	var cfdeleteServiceScript = []string{"delete-service", service, "-f"}

	log.Entry().WithField("cfService", service).Info("Deleting the requested Service")

	err := c.RunExecutable("cf", cfdeleteServiceScript...)

	if err != nil {
		return fmt.Errorf("Failed to delete Service: %w", err)
	}
	log.Entry().Info("Deletion of Service is finished or the Service has never existed")
	return nil
}

func cloudFoundryLogout(c execRunner) error {
	var cfLogoutScript = "logout"

	log.Entry().Info("Logging out of Cloud Foundry")

	err := c.RunExecutable("cf", cfLogoutScript)
	if err != nil {
		return fmt.Errorf("Failed to Logout of Cloud Foundry: %w", err)
	}
	log.Entry().Info("Logged out successfully")
	return nil
}
