package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func hadolintExecute(config hadolintExecuteOptions, telemetryData *telemetry.CustomData) {
	var outputBuffer bytes.Buffer
	var errorBuffer bytes.Buffer
	c := command.Command{}
	options := []string{
		"--format checkstyle",
	}

	c.Stdout(&outputBuffer)
	c.Stderr(&errorBuffer)
	//
	if len(config.ConfigurationURL) > 0 {
		loadConfigurationFile(config.ConfigurationURL, config.ConfigurationFile)
	}

	exists, err := piperutils.FileExists(config.ConfigurationFile)
	if err != nil {
		log.Entry().WithError(err).Error()
	}
	if exists {
		options = append(options, fmt.Sprintf("--config %s", config.ConfigurationFile))
	}

	runCommand := fmt.Sprintf("hadolint %s %s", config.DockerFile, strings.Join(options, " "))
	runCommandTokens := tokenize(runCommand)
	//command.Dir(config.ModulePath)
	err = c.RunExecutable(runCommandTokens[0], runCommandTokens[1:]...)

	//TODO: mind https://github.com/hadolint/hadolint/pull/392
	output := outputBuffer.String()
	if len(output) > 0 {
		log.Entry().WithField("report", output).Debug("Report created")
		ioutil.WriteFile(config.ReportFile, []byte(output), 0755)
	} else if err != nil {
		log.Entry().
			WithError(err).
			WithField("command", runCommand).
			Fatal(errorBuffer.String())
	}
	// persist report information
	piperutils.PersistReportsAndLinks(
		"hadolintExecute", "./",
		[]piperutils.Path{piperutils.Path{Target: config.ReportFile}},
		[]piperutils.Path{})
}

func loadConfigurationFile(url, file string) {
	exists, err := piperutils.FileExists(file)
	if err != nil {
		log.Entry().WithError(err).Error()
	}
	if !exists {
		log.Entry().WithField("file", url).Debug("Loading configuration from URL")
		if _, err := piperutils.Download(url, file); err != nil {
			log.Entry().
				WithError(err).
				WithField("file", url).
				Error("Failed to download configuration file from URL.")
		}
	}
}
