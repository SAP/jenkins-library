package cmd

import (
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/nexus"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func nexusUpload(config nexusUploadOptions, telemetryData *telemetry.CustomData) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())

	// for http calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runNexusUpload(&config, telemetryData, &c)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runNexusUpload(config *nexusUploadOptions, telemetryData *telemetry.CustomData, command execRunner) error {

	projectStructure := piperutils.ProjectStructure{}

	nexusClient := nexus.NexusUpload{Username: config.User, Password: config.Password}
	groupID := "" // TODO... expected to be provided for MTA projects, can be empty, though
	nexusClient.SetBaseUrl(config.Url, config.Version, config.Repository, groupID)

	// TODO:
	artifact := nexus.ArtifactDescription{}
	// TODO: Artifact ID is also expected to be provided for MTA projects, for compatibility
	// it would also have to be read from the "commonPipelineEnvironment"
	nexusClient.AddArtifact(artifact)

	nexusClient.UploadArtifacts()

	//log.Entry().WithField("LogField", "Log field content").Info("This is just a demo for a simple step.")
	return nil
}
