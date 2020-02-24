package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/nexus"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/ghodss/yaml"
	"io/ioutil"
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

type MtaYaml struct {
	ID      string `json:"ID"`
	Version string `json:"version"`
}

func runNexusUpload(config *nexusUploadOptions, telemetryData *telemetry.CustomData, command execRunner) error {

	projectStructure := piperutils.ProjectStructure{}

	nexusClient := nexus.NexusUpload{Username: config.User, Password: config.Password}
	groupID := config.GroupID // TODO... Only expected to be provided for MTA projects, can be empty, though
	nexusClient.SetBaseURL(config.Url, config.Version, config.Repository, groupID)

	if projectStructure.UsesMta() {
		var mtaYaml MtaYaml
		mtaYamContent, _ := ioutil.ReadFile("mta.yaml")
		err := yaml.Unmarshal(mtaYamContent, &mtaYaml)
		if err != nil {
			fmt.Println(err)
		}
		nexusClient.Version = mtaYaml.Version
		_ = nexusClient.AddArtifact(nexus.ArtifactDescription{File: "mta.yaml", Type: "yaml", Classifier: "", ID: config.ArtifactID})
		_ = nexusClient.AddArtifact(nexus.ArtifactDescription{File: mtaYaml.ID + ".mtar", Type: "mtar", Classifier: "", ID: config.ArtifactID})
	}

	if projectStructure.UsesMaven() {
		//read pom
	}


	nexusClient.UploadArtifacts()

	//log.Entry().WithField("LogField", "Log field content").Info("This is just a demo for a simple step.")
	return nil
}
