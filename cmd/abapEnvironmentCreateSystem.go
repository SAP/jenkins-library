package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
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
		return runCloudFoundryCreateService(&createServiceConfig, telemetryData, cf)
	}

	cfConfig, err := generateServiceParameterString(config)
	if err != nil {
		log.Entry().Fatalf("Could not generate parameter string")
	}
	// Calling cloudFoundryCreateService with the respective parameters
	createServiceConfig := cloudFoundryCreateServiceOptions{
		CfAPIEndpoint:         config.CfAPIEndpoint,
		CfOrg:                 config.CfOrg,
		CfSpace:               config.CfSpace,
		Username:              config.Username,
		Password:              config.Password,
		CfService:             config.CfService,
		CfServicePlan:         config.CfServicePlan,
		CfServiceInstanceName: config.CfServiceInstance,
		CfCreateServiceConfig: cfConfig,
		CfAsync:               false,
	}
	return runCloudFoundryCreateService(&createServiceConfig, telemetryData, cf)
}

func generateServiceParameterString(config *abapEnvironmentCreateSystemOptions) (string, error) {
	addonProduct := ""
	addonVersion := ""
	parentSaaSAppName := ""
	if config.AddonDescriptorFileName != "" && config.IncludeAddon {
		descriptor, err := abaputils.ReadAddonDescriptor(config.AddonDescriptorFileName)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return "", fmt.Errorf("Cloud not read addonProduct and addonVersion from %s: %w", config.AddonDescriptorFileName, err)
		}
		addonProduct = descriptor.AddonProduct
		addonVersion = descriptor.AddonVersionYAML
		parentSaaSAppName = "addon_test"

	}
	params := abapSystemParameters{
		AdminEmail:           config.AbapSystemAdminEmail,
		Description:          config.AbapSystemDescription,
		IsDevelopmentAllowed: &config.AbapSystemIsDevelopmentAllowed,
		SapSystemName:        config.AbapSystemID,
		SizeOfPersistence:    config.AbapSystemSizeOfPersistence,
		SizeOfRuntime:        config.AbapSystemSizeOfRuntime,
		AddonProductName:     addonProduct,
		AddonProductVersion:  addonVersion,
		ParentSaaSAppName:    parentSaaSAppName,
	}

	serviceParameters, err := json.Marshal(params)
	serviceParametersString := string(serviceParameters)
	log.Entry().Debugf("Service Parameters: %s", serviceParametersString)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return "", fmt.Errorf("Could not generate parameter string for the cloud foundry cli: %w", err)
	}

	return serviceParametersString, nil
}

type abapSystemParameters struct {
	AdminEmail           string `json:"admin_email,omitempty"`
	Description          string `json:"description,omitempty"`
	IsDevelopmentAllowed *bool  `json:"is_development_allowed,omitempty"`
	SapSystemName        string `json:"sapsystemname,omitempty"`
	SizeOfPersistence    int    `json:"size_of_persistence,omitempty"`
	SizeOfRuntime        int    `json:"size_of_runtime,omitempty"`
	AddonProductName     string `json:"addon_product_name,omitempty"`
	AddonProductVersion  string `json:"addon_product_version,omitempty"`
	ParentSaaSAppName    string `json:"parent_saas_appname,omitempty"`
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
