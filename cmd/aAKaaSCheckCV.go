package cmd

import (
	"encoding/json"

	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func aAKaaSCheckCV(config aAKaaSCheckCVOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *aAKaaSCheckCVCommonPipelineEnvironment) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	// for http calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runAAKaaSCheckCV(&config, telemetryData, &c, commonPipelineEnvironment)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAAKaaSCheckCV(config *aAKaaSCheckCVOptions, telemetryData *telemetry.CustomData, command command.ExecRunner, commonPipelineEnvironment *aAKaaSCheckCVCommonPipelineEnvironment) error {
	log.Entry().WithField("LogField", "Log field content").Info("This is just a demo for a simple step.")

	conn := new(connector)
	conn.initAAK(config.AAKaaSEndpoint, config.Username, config.Password, &piperhttp.Client{})

	c := cv{
		connector: *conn,
	}
	c.validate(*config)
	commonPipelineEnvironment.CVersion = c.Version
	commonPipelineEnvironment.CSpspLevel = c.SpsLevel
	commonPipelineEnvironment.CPatchLevel = c.PatchLevel
	return nil
}

type jsonCV struct {
	CV *cv `json:"d"`
}

type cv struct {
	connector
	Name       string `json:"Name"`
	Version    string `json:"Version"`
	SpsLevel   string `json:"SpsLevel"`
	PatchLevel string `json:"PatchLevel"`
}

func (c *cv) validate(options aAKaaSCheckCVOptions) error {
	appendum := "/ValidateComponentVersion?Name='" + options.AddonComponent + "'&Version='" + options.AddonComponentVersion + "'"
	body, err := c.connector.get(appendum)
	if err != nil {
		return err
	}
	var jCV jsonCV
	json.Unmarshal(body, &jCV)
	c.Name = jCV.CV.Name
	c.Version = jCV.CV.Version
	c.SpsLevel = jCV.CV.SpsLevel
	c.PatchLevel = jCV.CV.PatchLevel
	return nil
}
