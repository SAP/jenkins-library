package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
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

func runAbapEnvironmentReadAddonDescriptor(config *abapEnvironmentReadAddonDescriptorOptions, telemetryData *telemetry.CustomData, command command.ExecRunner, cpe *abapEnvironmentReadAddonDescriptorCommonPipelineEnvironment) error {

	var addonYAMLFile []byte
	filelocation, err := filepath.Glob(config.FileName)

	if err != nil || len(filelocation) != 1 {
		return errors.New(fmt.Sprintf("Could not find %v.", config.FileName))
	}
	filename, err := filepath.Abs(filelocation[0])
	if err != nil {
		return errors.New(fmt.Sprintf("Could not get path of %v.", config.FileName))
	}
	addonYAMLFile, err = ioutil.ReadFile(filename)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not read %v.", config.FileName))
	}

	var jsonBytes []byte
	jsonBytes, err = yaml.YAMLToJSON(addonYAMLFile)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not parse %v.", config.FileName))
	}

	var addonDescriptor addonDescriptor
	err = json.Unmarshal(jsonBytes, &addonDescriptor)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not unmarshal %v.", config.FileName))
	}

	var repositoryNames []string
	for _, v := range addonDescriptor.Repositories {
		repositoryNames = append(repositoryNames, v.Name)
	}

	repositoryNamesJSONString, _ := json.Marshal(repositoryNames)
	repositories, _ := json.Marshal(addonDescriptor.Repositories)

	cpe.abap.repositoryNames = string(repositoryNamesJSONString)
	cpe.abap.addonProduct = addonDescriptor.AddonProduct
	cpe.abap.addonUniqueID = addonDescriptor.AddonUniqueID
	cpe.abap.addonVersion = addonDescriptor.AddonVersion
	cpe.abap.customerID = fmt.Sprintf("%v", addonDescriptor.CustomerID)
	cpe.abap.repositories = string(repositories)

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
	Name     string `json:"name"`
	Tag      string `json:"tag"`
	Version  string `json:"version"`
	Spslevel string `json:"SpsLevel"`
}
