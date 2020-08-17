package cmd

import (
	"encoding/json"

	"github.com/SAP/jenkins-library/pkg/abaputils"
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
	err := runAbapEnvironmentASimulate(&config, telemetryData, commonPipelineEnvironment)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentASimulate(config *abapEnvironmentASimulateOptions, telemetryData *telemetry.CustomData, cpe *abapEnvironmentASimulateCommonPipelineEnvironment) error {
	log.Entry().WithField("LogField", "Log field content").Info("This is just a demo for a simple step.")

	var reposBackToCPE []abaputils.Repository
	var repo abaputils.Repository
	repo.Name = config.SWC
	repo.VersionOtherFormat = config.SWCRelease
	repo.PackageName = config.PackageName
	repo.PackageType = "AOI"
	repo.SpsLevel = config.SpsLevel
	repo.Namespace = config.Namespace

	reposBackToCPE = append(reposBackToCPE, repo)

	backToCPE, _ := json.Marshal(reposBackToCPE)
	cpe.abap.repositories = string(backToCPE)
	return nil
}
