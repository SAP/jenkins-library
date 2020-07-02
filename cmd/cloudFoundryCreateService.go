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

	var c command.ExecRunner = &command.Command{}

	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())

	var err, logouterr error

	loginOptions := cloudfoundry.LoginOptions{
		CfAPIEndpoint: config.CfAPIEndpoint,
		CfOrg:         config.CfOrg,
		CfSpace:       config.CfSpace,
		Username:      config.Username,
		Password:      config.Password,
	}

	err = cloudfoundry.Login(loginOptions)
	if err == nil {
		err = runCloudFoundryCreateService(&config, telemetryData, c)
	}
	if err != nil {
		logouterr = cloudfoundry.Logout()
		if logouterr != nil {
			log.Entry().WithError(logouterr).Fatal("step execution failed")
		}
		log.Entry().WithError(err).Fatal("step execution failed")
	}
	logouterr = cloudfoundry.Logout()
	if logouterr != nil {
		log.Entry().WithError(logouterr).Fatal("step execution failed")
	}
	log.Entry().Info("Service creation completed successfully")
}

func runCloudFoundryCreateService(config *cloudFoundryCreateServiceOptions, telemetryData *telemetry.CustomData, c command.ExecRunner) error {
	var err error
	log.Entry().Info("Creating Cloud Foundry Service")

	var cfCreateServiceScript []string
	cfCreateServiceScript = []string{"create-service", config.CfService, config.CfServicePlan, config.CfServiceInstanceName}

	if config.CfServiceBroker != "" {
		cfCreateServiceScript = append(cfCreateServiceScript, "-b", config.CfServiceBroker)
	}
	if config.CfCreateServiceConfig != "" {
		cfCreateServiceScript = append(cfCreateServiceScript, "-c", config.CfCreateServiceConfig)
	}
	if config.CfServiceTags != "" {
		cfCreateServiceScript = append(cfCreateServiceScript, "-t", config.CfServiceTags)
	}
	if config.ServiceManifest != "" {
		var varPart []string
		cfCreateServiceScript = []string{"create-service-push", "--no-push", "--service-manifest", config.ServiceManifest}

		if len(config.ManifestVariablesFiles) >= 0 {
			for _, v := range config.ManifestVariablesFiles {
				if fileExists(v) {
					cfCreateServiceScript = append(cfCreateServiceScript, "--vars-file", v)
				} else {
					return fmt.Errorf("Failed to append Manifest Variables File: %w", errors.New(v+" is not a file"))
				}
			}
		}
		if len(config.ManifestVariables) >= 0 {
			varPart, err = varOptions(config.ManifestVariables)
		}
		for _, s := range varPart {
			cfCreateServiceScript = append(cfCreateServiceScript, s)
		}
	}
	fmt.Print(cfCreateServiceScript)
	err = c.RunExecutable("cf", cfCreateServiceScript...)

	if err != nil {
		return fmt.Errorf("Failed to Create Service: %w", err)
	}
	return nil
}

func varOptions(options []string) ([]string, error) {
	var varOptionsString []string
	for _, s := range options {
		varOptionsString = append(varOptionsString, "--var", s)
	}
	return varOptionsString, nil
}
