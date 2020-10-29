package cmd

import (
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

		serviceParameters := ""

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
			CfCreateServiceConfig: serviceParameters,
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
}
