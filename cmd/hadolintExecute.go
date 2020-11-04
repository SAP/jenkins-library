package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func hadolintExecute(config hadolintExecuteOptions, _ *telemetry.CustomData) {
	runner := command.Command{
		ErrorCategoryMapping: map[string][]string{},
	}
	// reroute command output to logging framework
	// runner.Stdout(log.Writer())
	// runner.Stderr(log.Writer())

	client := piperhttp.Client{}
	clientOptions := piperhttp.ClientOptions{TransportTimeout: 20 * time.Second}
	if len(config.ConfigurationUsername) > 0 {
		clientOptions.Username = config.ConfigurationUsername
		clientOptions.Password = config.ConfigurationPassword
	}
	client.SetOptions(clientOptions)

	if err := runHadolint(config, &client, &runner); err != nil {
		log.Entry().WithError(err).Fatal("Execution failed")
	}
}

func runHadolint(config hadolintExecuteOptions, client piperhttp.Downloader, runner command.ExecRunner) error {
	var outputBuffer bytes.Buffer
	var errorBuffer bytes.Buffer
	runner.Stdout(&outputBuffer)
	runner.Stderr(&errorBuffer)

	options := []string{
		"--format checkstyle",
	}
	// load config file from URL
	if !hasConfigurationFile(config.ConfigurationFile) && len(config.ConfigurationURL) > 0 {
		if err := loadConfigurationFile(config.ConfigurationURL, config.ConfigurationFile, client); err != nil {
			return errors.Wrap(err, "failed to load configuration file from URL")
		}
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
	err := runner.RunExecutable(runCommandTokens[0], runCommandTokens[1:]...)

	//TODO: related to https://github.com/hadolint/hadolint/issues/391
	// hadolint exists with 1 if there are processing issues but also if there are findings
	// thus check stdout first if a report was created
	if output := outputBuffer.String(); len(output) > 0 {
		log.Entry().WithField("report", output).Debug("Report created")
		ioutil.WriteFile(config.ReportFile, []byte(output), 0755)
	} else if err != nil {
		// if stdout is empty a processing issue occured
		return errors.Wrap(err, errorBuffer.String())
	}
	// persist report information
	piperutils.PersistReportsAndLinks("hadolintExecute", "./", []piperutils.Path{piperutils.Path{Target: config.ReportFile}}, []piperutils.Path{})
	return nil
}

// loadConfigurationFile loads a file from the provided url
func loadConfigurationFile(url, file string, client piperhttp.Downloader) error {
	log.Entry().WithField("url", url).Debug("Loading configuration file from URL")
	return client.DownloadFile(url, file, nil, nil)
}

// hasConfigurationFile checks if the given file exists
func hasConfigurationFile(file string) bool {
	exists, err := piperutils.FileExists(file)
	if err != nil {
		log.Entry().WithError(err).Error()
	}
	return exists
}
