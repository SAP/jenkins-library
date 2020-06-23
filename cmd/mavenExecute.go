package cmd

import (
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/maven"
	"io/ioutil"

	"github.com/SAP/jenkins-library/pkg/telemetry"
)

var writeFile = ioutil.WriteFile

func mavenExecute(config mavenExecuteOptions, _ *telemetry.CustomData) {
	runner := command.Command{}
	err := runMavenExecute(config, &runner)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runMavenExecute(config mavenExecuteOptions, runner execRunner) error {
	options := maven.ExecuteOptions{
		PomPath:                     config.PomPath,
		ProjectSettingsFile:         config.ProjectSettingsFile,
		GlobalSettingsFile:          config.GlobalSettingsFile,
		M2Path:                      config.M2Path,
		Goals:                       config.Goals,
		Defines:                     config.Defines,
		Flags:                       config.Flags,
		LogSuccessfulMavenTransfers: config.LogSuccessfulMavenTransfers,
		ReturnStdout:                config.ReturnStdout,
	}

	output, err := maven.Execute(&options, runner)
	if err == nil && config.ReturnStdout {
		err = writeFile(".pipeline/maven_output.txt", []byte(output), 0644)
	}
	return err
}
