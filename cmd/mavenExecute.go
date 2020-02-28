package cmd

import (
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/maven"

	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func mavenExecute(config mavenExecuteOptions, telemetryData *telemetry.CustomData) string {
	c := command.Command{}

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

	output, err := maven.Execute(&options, &c)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}

	return output
}
