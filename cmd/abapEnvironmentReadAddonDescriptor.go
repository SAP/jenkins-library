package cmd

import (
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func abapEnvironmentReadAddonDescriptor(config abapEnvironmentReadAddonDescriptorOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *abapEnvironmentReadAddonDescriptorCommonPipelineEnvironment) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	// for http calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runAbapEnvironmentReadAddonDescriptor(&config, telemetryData, &c, &commonPipelineEnvironment)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentReadAddonDescriptor(config *abapEnvironmentReadAddonDescriptorOptions, telemetryData *telemetry.CustomData, command execRunner, commonPipelineEnvironment *abapEnvironmentReadAddonDescriptorCommonPipelineEnvironment) error {
	log.Entry().WithField("LogField", "Log field content").Info("This is just a demo for a simple step.")
	return nil
}
