package cmd

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func cloudFoundryCreateService(config cloudFoundryCreateServiceOptions, telemetryData *telemetry.CustomData) {

	cf := cloudfoundry.CFUtils{Exec: &command.Command{}}

	err := runCloudFoundryCreateService(&config, telemetryData, cf)

	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}

}

func runCloudFoundryCreateService(config *cloudFoundryCreateServiceOptions, telemetryData *telemetry.CustomData, cf cloudfoundry.CFUtils) (err error) {
	var c = cf.Exec

	loginOptions := cloudfoundry.LoginOptions{
		CfAPIEndpoint: config.CfAPIEndpoint,
		CfOrg:         config.CfOrg,
		CfSpace:       config.CfSpace,
		Username:      config.Username,
		Password:      config.Password,
	}

	err = cf.Login(loginOptions)

	if err != nil {
		return fmt.Errorf("Error while logging in: %w", err)
	}

	defer func() {
		logoutErr := cf.Logout()
		if logoutErr != nil {
			err = fmt.Errorf("Error while logging out occurred: %w", logoutErr)
		}
	}()

	err = cloudFoundryCreateServiceRequest(config, telemetryData, c)

	if err != nil {
		return fmt.Errorf("Service creation failed: %w", err)
	}

	log.Entry().Info("Service creation completed successfully")

	return err

}

func cloudFoundryCreateServiceRequest(config *cloudFoundryCreateServiceOptions, telemetryData *telemetry.CustomData, c command.ExecRunner) error {
	var err error
	log.Entry().Info("Creating Cloud Foundry Service")

	cfCreateServiceScript := []string{"create-service", config.CfService, config.CfServicePlan, config.CfServiceInstanceName}

	if config.CfServiceBroker != "" {
		cfCreateServiceScript = append(cfCreateServiceScript, "-b", config.CfServiceBroker)
	}
	if config.CfCreateServiceConfig != "" {
		cfCreateServiceScript = append(cfCreateServiceScript, "-c", config.CfCreateServiceConfig)
	}
	if config.CfServiceTags != "" {
		cfCreateServiceScript = append(cfCreateServiceScript, "-t", config.CfServiceTags)
	}
	if !config.CfAsync {
		cfCreateServiceScript = append(cfCreateServiceScript, "--wait")
	}
	if config.ServiceManifest != "" && fileExists(config.ServiceManifest) {

		cfCreateServiceScript = []string{"create-service-push", "--no-push", "--service-manifest", config.ServiceManifest}

		if len(config.ManifestVariablesFiles) >= 0 {
			varFileOpts, err := cloudfoundry.GetVarsFileOptions(config.ManifestVariablesFiles)
			if err != nil {
				return errors.Wrapf(err, "Cannot prepare var-file-options: '%v'", config.ManifestVariablesFiles)
			}
			cfCreateServiceScript = append(cfCreateServiceScript, varFileOpts...)
		}
		if len(config.ManifestVariables) >= 0 {
			varOptions, err := cloudfoundry.GetVarsOptions(config.ManifestVariables)
			if err != nil {
				return errors.Wrapf(err, "Cannot prepare var-options: '%v'", config.ManifestVariables)
			}
			cfCreateServiceScript = append(cfCreateServiceScript, varOptions...)
		}
	}
	err = c.RunExecutable("cf", cfCreateServiceScript...)

	if err != nil {
		return fmt.Errorf("Failed to Create Service: %w", err)
	}
	return nil
}
