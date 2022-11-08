package cmd

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func cloudFoundryDeleteService(options cloudFoundryDeleteServiceOptions, telemetryData *telemetry.CustomData) {

	c := command.Command{}

	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	cfUtils := cloudfoundry.CFUtils{
		Exec: &c,
	}

	err := runCloudFoundryDeleteService(options, &c, &cfUtils)
	if err != nil {
		log.Entry().
			WithError(err).
			Fatal("Error occurred during step.")
	}
}

func runCloudFoundryDeleteService(options cloudFoundryDeleteServiceOptions, c command.ExecRunner, cfUtils cloudfoundry.AuthenticationUtils) (returnedError error) {

	config := cloudfoundry.LoginOptions{
		CfAPIEndpoint: options.CfAPIEndpoint,
		CfOrg:         options.CfOrg,
		CfSpace:       options.CfSpace,
		Username:      options.Username,
		Password:      options.Password,
	}
	loginErr := cfUtils.Login(config)
	if loginErr != nil {
		return fmt.Errorf("Error while logging in occurred: %w", loginErr)
	}
	defer func() {
		logoutErr := cfUtils.Logout()
		if logoutErr != nil && returnedError == nil {
			returnedError = fmt.Errorf("Error while logging out occurred: %w", logoutErr)
		}
	}()

	if options.CfDeleteServiceKeys {
		err := cloudFoundryDeleteServiceKeys(options, c)
		if err != nil {
			return err
		}
	}

	err := cloudFoundryDeleteServiceFunction(options.CfServiceInstance, c)
	if err != nil {
		return err
	}

	return returnedError
}

func cloudFoundryDeleteServiceKeys(options cloudFoundryDeleteServiceOptions, c command.ExecRunner) error {

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

func cloudFoundryDeleteServiceFunction(service string, c command.ExecRunner) error {
	var cfdeleteServiceScript = []string{"delete-service", service, "-f"}

	log.Entry().WithField("cfService", service).Info("Deleting the requested Service")

	err := c.RunExecutable("cf", cfdeleteServiceScript...)

	if err != nil {
		return fmt.Errorf("Failed to delete Service: %w", err)
	}
	log.Entry().Info("Deletion of Service is finished or the Service has never existed")
	return nil
}
