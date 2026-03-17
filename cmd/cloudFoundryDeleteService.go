package cmd

import (
	"bytes"
	"encoding/json"
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

	err := runCloudFoundryDeleteService(&options, &c, &cfUtils)
	if err != nil {
		log.Entry().
			WithError(err).
			Fatal("Error occurred during step.")
	}
}

func runCloudFoundryDeleteService(options *cloudFoundryDeleteServiceOptions, c command.ExecRunner, cfUtils cloudfoundry.AuthenticationUtils) (returnedError error) {

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

	err := cloudFoundryDeleteServiceFunction(options, c)
	if err != nil {
		return err
	}

	return returnedError
}

func cloudFoundryDeleteServiceKeys(options *cloudFoundryDeleteServiceOptions, c command.ExecRunner) error {

	log.Entry().Info("Deleting inherent Service Keys")

	ServiceGUID, err := cloudFoundryFindServiceGUID(options, c)
	if err != nil {

	}

	log.Entry().WithField("Service Instance GUID :", ServiceGUID).Info("ServiceGUID")

	ServiceKeyNames, err := cloudFoundryFindServiceKeyNames(ServiceGUID, c)
	if err != nil {
		return fmt.Errorf("Failed to determine Service Key names: %w", err)
	}
	if len(ServiceKeyNames) == 0 {
		log.Entry().Info("No service key names could be retrieved for your requested Service")
		return nil
	}

	log.Entry().WithField("Number of service keys :", len(ServiceKeyNames)).Info("ServiceKey")
	//Deleting all matched Service Keys for Service
	for _, serviceKey := range ServiceKeyNames {
		log.Entry().WithField("Service key :", serviceKey).Info("ServiceKey")
		log.Entry().WithField("Deleting Service Key", serviceKey).Info("ServiceKeyDeletion")

		var cfDeleteServiceKeyScript = []string{"delete-service-key", options.CfServiceInstance, serviceKey, "-f"}
		if !options.CfAsync {
			cfDeleteServiceKeyScript = append(cfDeleteServiceKeyScript, "--wait")
		}
		err := c.RunExecutable("cf", cfDeleteServiceKeyScript...)
		if err != nil {
			return fmt.Errorf("Failed to Delete Service Key: %w", err)
		}
	}

	log.Entry().Info("ServiceKeys have been deleted!")
	return nil
}

func cloudFoundryFindServiceKeyNames(ServiceGUID string, c command.ExecRunner) (ServiceKeyNames []string, err error) {

	type Resource struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	type ServiceKeysResponse struct {
		Resources []Resource `json:"resources"`
	}

	// Retrieve list of Service Keys bound to Service Instance
	var cfFindServiceKeysScript = []string{"curl", "/v3/service_credential_bindings?service_instance_guids=" + ServiceGUID}
	var serviceKeyBytes bytes.Buffer

	c.Stdout(&serviceKeyBytes)
	err = c.RunExecutable("cf", cfFindServiceKeysScript...)

	if err != nil {
		return ServiceKeyNames, fmt.Errorf("Failed to any related Service Keys for the Service Instance, most likely your service doesn't exist: %w", err)
	}

	if len(serviceKeyBytes.String()) == 0 {
		log.Entry().Info("No service key could be retrieved for your requested Service")
		return ServiceKeyNames, nil
	}
	var serviceKeys = serviceKeyBytes.String()
	var response ServiceKeysResponse

	err = json.Unmarshal([]byte(serviceKeys), &response)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}
	if len(response.Resources) == 0 {
		log.Entry().Info("No service key could be retrieved for your requested Service")
		return ServiceKeyNames, nil
	}
	for _, resource := range response.Resources {
		if resource.Type == "key" {
			ServiceKeyNames = append(ServiceKeyNames, resource.Name)
		}
	}
	return ServiceKeyNames, err
}

func cloudFoundryFindServiceGUID(options *cloudFoundryDeleteServiceOptions, c command.ExecRunner) (GUID string, err error) {

	// Read GUID of Cloud Foundry Instance
	var cfFindServiceGUIDScript = []string{"service", options.CfServiceInstance, "--guid"}
	var serviceGUIDBytes bytes.Buffer
	c.Stdout(&serviceGUIDBytes)
	err = c.RunExecutable("cf", cfFindServiceGUIDScript...)
	if err != nil {
		return GUID, fmt.Errorf("Failed to Find Service Instance GUID, most likely your service doesn't exist: %w", err)
	}
	if len(serviceGUIDBytes.String()) == 0 {
		log.Entry().Info("No Service Instance GUID could be retrieved for your requested Service")
		return GUID, nil
	}
	unformattedGUID := serviceGUIDBytes.String()
	GUID = strings.Replace(unformattedGUID, "\n", "", 1)
	return GUID, nil
}

func cloudFoundryDeleteServiceFunction(options *cloudFoundryDeleteServiceOptions, c command.ExecRunner) error {
	var cfdeleteServiceScript = []string{"delete-service", options.CfServiceInstance, "-f"}

	if !options.CfAsync {
		cfdeleteServiceScript = append(cfdeleteServiceScript, "--wait")
	}

	log.Entry().WithField("cfService", options.CfServiceInstance).Info("Deleting the requested Service")

	err := c.RunExecutable("cf", cfdeleteServiceScript...)

	if err != nil {
		return fmt.Errorf("Failed to delete Service: %w", err)
	}
	log.Entry().Info("Deletion of Service is finished or the Service has never existed")
	return nil
}
