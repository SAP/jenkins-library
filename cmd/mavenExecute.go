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
		Goals:                       splitAndTrimParams(config.Goals),
		Defines:                     splitAndTrimParams(config.Defines),
		Flags:                       splitAndTrimParams(config.Flags),
		LogSuccessfulMavenTransfers: config.LogSuccessfulMavenTransfers,
		ReturnStdout:                config.ReturnStdout,
	}

	output, err := maven.Execute(&options, runner)
	if err == nil && config.ReturnStdout {
		err = ioutil.WriteFile(".pipeline/maven_output.txt", []byte(output), 0644)
	}
	return err
}

// splitAndTrimParams iterates over the strings in params and splits each string on spaces. Each resulting
// sub-string is then a separate entry in the returned array. Duplicate entries are eliminated.
func splitAndTrimParams(params []string) []string {
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
