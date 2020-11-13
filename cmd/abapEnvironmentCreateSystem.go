package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/ghodss/yaml"
	"github.com/google/uuid"
)

func abapEnvironmentCreateSystem(config abapEnvironmentCreateSystemOptions, telemetryData *telemetry.CustomData) {

	cf := cloudfoundry.CFUtils{Exec: &command.Command{}}
	u := &googleUUID{}

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runAbapEnvironmentCreateSystem(&config, telemetryData, cf, u)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentCreateSystem(config *abapEnvironmentCreateSystemOptions, telemetryData *telemetry.CustomData, cf cloudfoundry.CFUtils, u uuidGenerator) error {

	if config.ServiceManifest != "" {
		// if the manifest file is provided, it is directly passed through to cloudFoundryCreateService
		createServiceConfig := cloudFoundryCreateServiceOptions{
			CfAPIEndpoint:   config.CfAPIEndpoint,
			CfOrg:           config.CfOrg,
			CfSpace:         config.CfSpace,
			Username:        config.Username,
			Password:        config.Password,
			ServiceManifest: config.ServiceManifest,
		}
		runCloudFoundryCreateService(&createServiceConfig, telemetryData, cf)
	} else {
		// if no manifest file is provided, it is created with the provided config values
		manifestYAML, err := generateManifestYAML(config)

		// writing the yaml into a temporary file
		path, _ := os.Getwd()
		path = path + "/generated_service_manifest-" + u.getUUID() + ".yml"
		log.Entry().Debugf("Path: %s", path)
		err = ioutil.WriteFile(path, manifestYAML, 0644)
		if err != nil {
			return fmt.Errorf("%s: %w", "Could not generate manifest file for the cloud foundry cli", err)
		}

		defer os.Remove(path)

		// Calling cloudFoundryCreateService with the respective parameters
		createServiceConfig := cloudFoundryCreateServiceOptions{
			CfAPIEndpoint:   config.CfAPIEndpoint,
			CfOrg:           config.CfOrg,
			CfSpace:         config.CfSpace,
			Username:        config.Username,
			Password:        config.Password,
			ServiceManifest: path,
		}
		runCloudFoundryCreateService(&createServiceConfig, telemetryData, cf)
	}

	return nil
}

func generateManifestYAML(config *abapEnvironmentCreateSystemOptions) ([]byte, error) {
	addonProduct := ""
	addonVersion := ""
	if config.AddonDescriptorFileName != "" && config.IncludeAddon {
		descriptor, err := abaputils.ReadAddonDescriptor(config.AddonDescriptorFileName)
		if err != nil {
			return nil, fmt.Errorf("Cloud not read addonProduct and addonVersion from %s: %w", config.AddonDescriptorFileName, err)
		}
		addonProduct = descriptor.AddonProduct
		addonVersion = descriptor.AddonVersionYAML
	}
	params := abapSystemParameters{
		AdminEmail:           config.AbapSystemAdminEmail,
		Description:          config.AbapSystemDescription,
		IsDevelopmentAllowed: config.AbapSystemIsDevelopmentAllowed,
		SapSystemName:        config.AbapSystemID,
		SizeOfPersistence:    config.AbapSystemSizeOfPersistence,
		SizeOfRuntime:        config.AbapSystemSizeOfRuntime,
		AddonProductName:     addonProduct,
		AddonProductVersion:  addonVersion,
	}

	serviceParameters, err := json.Marshal(params)
	serviceParametersString := string(serviceParameters)
	log.Entry().Debugf("Service Parameters: %s", serviceParametersString)
	if err != nil {
		return nil, fmt.Errorf("Could not generate parameter string for the cloud foundry cli: %w", err)
	}

	/*
		Generating the temporary manifest yaml file
	*/
	service := Service{
		Name:       config.CfServiceInstance,
		Broker:     config.CfService,
		Plan:       config.CfServicePlan,
		Parameters: serviceParametersString,
	}

	serviceManifest := serviceManifest{CreateServices: []Service{service}}
	errorMessage := "Could not generate manifest for the cloud foundry cli"

	// converting the golang structure to json
	manifestJSON, err := json.Marshal(serviceManifest)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errorMessage, err)
	}

	// converting the json to yaml
	manifestYAML, err := yaml.JSONToYAML(manifestJSON)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errorMessage, err)
	}

	log.Entry().Debug(string(manifestYAML))

	return manifestYAML, nil
}

type abapSystemParameters struct {
	AdminEmail           string `json:"admin_email,omitempty"`
	Description          string `json:"description,omitempty"`
	IsDevelopmentAllowed bool   `json:"is_development_allowed,omitempty"`
	SapSystemName        string `json:"sapsystemname,omitempty"`
	SizeOfPersistence    int    `json:"size_of_persistence,omitempty"`
	SizeOfRuntime        int    `json:"size_of_runtime,omitempty"`
	AddonProductName     string `json:"addon_product_name,omitempty"`
	AddonProductVersion  string `json:"addon_product_version,omitempty"`
}

type serviceManifest struct {
	CreateServices []Service `json:"create-services"`
}

// Service struct for creating a cloud foundry service
type Service struct {
	Name       string `json:"name"`
	Broker     string `json:"broker"`
	Plan       string `json:"plan"`
	Parameters string `json:"parameters,omitempty"`
}

type uuidGenerator interface {
	getUUID() string
}

type googleUUID struct {
}

func (u *googleUUID) getUUID() string {
	return uuid.New().String()
}
