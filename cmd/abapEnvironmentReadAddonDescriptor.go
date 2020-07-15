package cmd

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/ghodss/yaml"
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
	err := runAbapEnvironmentReadAddonDescriptor(&config, telemetryData, &c, commonPipelineEnvironment)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentReadAddonDescriptor(config *abapEnvironmentReadAddonDescriptorOptions, telemetryData *telemetry.CustomData, command command.ExecRunner, commonPipelineEnvironment *abapEnvironmentReadAddonDescriptorCommonPipelineEnvironment) error {

	var addonYAMLFile []byte
	filelocation, err := filepath.Glob(config.FileName)

	//Parse YAML ATC run configuration as body for ATC run trigger
	if err != nil {
		return err
	}
	filename, _ := filepath.Abs(filelocation[0])
	addonYAMLFile, err = ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var jsonBytes []byte
	jsonBytes, err = yaml.YAMLToJSON(addonYAMLFile)
	if err != nil {
		return err
	}

	var addonDescriptor addonDescriptor
	json.Unmarshal(jsonBytes, &addonDescriptor)

	var repositoryNames []string
	for _, v := range addonDescriptor.Repositories {
		repositoryNames = append(repositoryNames, v.Name)
	}

	repositoryNamesJSONString, _ := json.Marshal(repositoryNames)

	commonPipelineEnvironment.abap.repositoryNames = string(repositoryNamesJSONString)

	return nil
}

type addonDescriptor struct {
	AddonProduct  string         `json:"addonProduct"`
	AddonVersion  string         `json:"addonVersion"`
	AddonUniqueID string         `json:"addonUniqueID"`
	CustomerID    interface{}    `json:"customerID"`
	Repositories  []repositories `json:"repositories"`
}

type repositories struct {
	Name    string `json:"name"`
	Tag     string `json:"tag"`
	Version string `json:"version"`
}
