package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func abapEnvironmentCreateSystem(config abapEnvironmentCreateSystemOptions, telemetryData *telemetry.CustomData) {

	cf := cloudfoundry.CFUtils{Exec: &command.Command{}}

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runAbapEnvironmentCreateSystem(&config, telemetryData, cf)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentCreateSystem(config *abapEnvironmentCreateSystemOptions, telemetryData *telemetry.CustomData, cf cloudfoundry.CFUtils) error {

	// this is used to ensure compatibility for old pipeline configurations (using the step cloudFoundryCreateService)
	if config.ServiceManifest != "" {
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

		addonProduct := ""
		addonVersion := ""
		if config.AddonDescriptor != "" {
			descriptor, err := abaputils.ReadAddonDescriptor(config.AddonDescriptor)
			if err != nil {
				return fmt.Errorf("Cloud not read addonProduct and addonVersion from %s: %w", config.AddonDescriptor, err)
			}
			addonProduct = descriptor.AddonProduct
			addonVersion = descriptor.AddonVersionYAML
		}
		params := abapSystemParameters{
			AdminEmail:           config.AdminEmail,
			Description:          config.Description,
			IsDevelopmentAllowed: config.IsDevelopmentAllowed,
			SapSystemName:        config.SapSystemName,
			SizeOfPersistence:    config.SizeOfPersistence,
			SizeOfRuntime:        config.SizeOfRuntime,
			AddonProductName:     addonProduct,
			AddonProductVersion:  addonVersion,
		}
		serviceParameters, err := json.Marshal(params)
		if err != nil {
			return fmt.Errorf("Could not generate parameter string for the cloud foundry cli: %w", err)
		}

		createServiceConfig := cloudFoundryCreateServiceOptions{
			CfAPIEndpoint:         config.CfAPIEndpoint,
			CfOrg:                 config.CfOrg,
			CfSpace:               config.CfSpace,
			Username:              config.Username,
			Password:              config.Password,
			CfServiceInstanceName: config.CfServiceInstanceName,
			CfServiceBroker:       config.CfServiceBroker,
			CfService:             config.CfService,
			CfServicePlan:         config.CfServicePlan,
			CfCreateServiceConfig: string(serviceParameters),
		}
		runCloudFoundryCreateService(&createServiceConfig, telemetryData, cf)
	}

	return nil
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
