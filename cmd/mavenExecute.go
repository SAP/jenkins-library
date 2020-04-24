package cmd

import (
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/maven"
	sliceUtils "github.com/SAP/jenkins-library/pkg/piperutils"
	"io/ioutil"

	"github.com/SAP/jenkins-library/pkg/telemetry"
)

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
		Goals:                       splitTrimAndDeDupParams(config.Goals),
		Defines:                     splitTrimAndDeDupParams(config.Defines),
		Flags:                       splitTrimAndDeDupParams(config.Flags),
		LogSuccessfulMavenTransfers: config.LogSuccessfulMavenTransfers,
		ReturnStdout:                config.ReturnStdout,
	}

	output, err := maven.Execute(&options, runner)
	if err == nil && config.ReturnStdout {
		err = ioutil.WriteFile(".pipeline/maven_output.txt", []byte(output), 0644)
	}
	return err
}

func splitTrimAndDeDupParams(params []string) []string {
	return sliceUtils.SplitTrimAndDeDup(params, " ")
}
