package cmd

import (
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"io/ioutil"
	"strings"

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
		Goals:                       separateAndTrimParams(config.Goals),
		Defines:                     separateAndTrimParams(config.Defines),
		Flags:                       separateAndTrimParams(config.Flags),
		LogSuccessfulMavenTransfers: config.LogSuccessfulMavenTransfers,
		ReturnStdout:                config.ReturnStdout,
	}

	output, err := maven.Execute(&options, runner)
	if err == nil && output != "" {
		err = ioutil.WriteFile(".pipeline/maven_output.txt", []byte(output), 0644)
	}
	return err
}

func separateAndTrimParams(params []string) []string {
	if len(params) == 0 {
		return params
	}
	var cleanedParams []string
	for _, param := range params {
		splitParams := strings.Split(param, " ")
		for _, part := range splitParams {
			part = strings.TrimSpace(part)
			if part != "" && !piperutils.ContainsString(cleanedParams, part) {
				cleanedParams = append(cleanedParams, part)
			}
		}
	}
	return cleanedParams
}
