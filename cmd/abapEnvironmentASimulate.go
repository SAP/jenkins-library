package cmd

import (
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func abapEnvironmentASimulate(config abapEnvironmentASimulateOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *abapEnvironmentASimulateCommonPipelineEnvironment) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	// for http calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runAbapEnvironmentASimulate(&config, telemetryData, &c, commonPipelineEnvironment)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentASimulate(config *abapEnvironmentASimulateOptions, telemetryData *telemetry.CustomData, command command.ExecRunner, commonPipelineEnvironment *abapEnvironmentASimulateCommonPipelineEnvironment) error {
	log.Entry().WithField("LogField", "Log field content").Info("This is just a demo for a simple step.")
	log.Entry().Infof("package Type %v", config.PackageType)
	log.Entry().Infof("packageName %v", config.PackageName)
	log.Entry().Infof("SWC %v", config.SWC)
	log.Entry().Infof("SWCRelease %v", config.SWCRelease)
	log.Entry().Infof("SpsLevel %v", config.SpsLevel)
	log.Entry().Infof("Namespace %v", config.Namespace)
	log.Entry().Infof("commit %v", config.PreviousDeliveryCommit)
	// ins environment schreibne
	commonPipelineEnvironment.PackageType = config.PackageType
	commonPipelineEnvironment.PackageName = config.PackageName
	commonPipelineEnvironment.SWC = config.SWC
	commonPipelineEnvironment.SWCRelease = config.SWCRelease
	commonPipelineEnvironment.SpsLevel = config.SpsLevel
	commonPipelineEnvironment.Namespace = config.Namespace
	commonPipelineEnvironment.PreviousDeliveryCommit = config.PreviousDeliveryCommit
	commonPipelineEnvironment.persist(".pipeline", "commonPipelineEnvironment")
	return nil
}
