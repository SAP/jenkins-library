package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
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
	client := piperhttp.Client{}

	c.Stdout(&outputBuffer)
	c.Stderr(&errorBuffer)
	// load config file from URL
	if !hasConfigurationFile(config.ConfigurationFile) && len(config.ConfigurationURL) > 0 {
		loadConfigurationFile(config.ConfigurationURL, config.ConfigurationFile, &client)
	}
	// use config
	if hasConfigurationFile(config.ConfigurationFile) {
		options = append(options, fmt.Sprintf("--config %s", config.ConfigurationFile))
		log.Entry().WithField("file", config.ConfigurationFile).Debug("Using configuration file")
	} else {
		log.Entry().Debug("No configuration file found.")
	}
	// execute scan command
	runCommand := fmt.Sprintf("hadolint %s %s", config.DockerFile, strings.Join(options, " "))
	runCommandTokens := tokenize(runCommand)
	err := c.RunExecutable(runCommandTokens[0], runCommandTokens[1:]...)
	//TODO: incorporate https://github.com/hadolint/hadolint/pull/392 if merged
	output := outputBuffer.String()
	// hadolint exists with 1 if there are processing issues but also if there are findings
	// thus check stdout first if a report was created
	if len(output) > 0 {
		log.Entry().WithField("report", output).Debug("Report created")
		ioutil.WriteFile(config.ReportFile, []byte(output), 0755)
	} else if err != nil {
		// if stdout is empty a processing issue occured
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

// loadConfigurationFile loads a file from the provided url
func loadConfigurationFile(url, file string, client piperhttp.Downloader) {
	log.Entry().WithField("url", url).Debug("Loading configuration file from URL")

	if err := client.DownloadFile(url, file, nil, nil); err != nil {
		log.Entry().
			WithError(err).
			WithField("file", url).
			Error("Failed to load configuration file from URL.")
	}
}

// hasConfigurationFile checks if the given file exists
func hasConfigurationFile(file string) bool {
	exists, err := piperutils.FileExists(file)
	if err != nil {
		log.Entry().WithError(err).Error()
	}
	return exists
}
