package cmd

import (
	"bytes"
	"github.com/SAP/jenkins-library/pkg/http"
	"io"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

const mavenExecutable = "mvn"

func mavenExecute(config mavenExecuteOptions, telemetryData *telemetry.CustomData) string {
	c := command.Command{}
	output, err := runMavenExecute(&config, &c)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}

	return output
}

func runMavenExecute(config *mavenExecuteOptions, command execRunner) (string, error) {
	stdOutBuf, stdOut := evaluateStdOut(config)
	command.Stdout(stdOut)
	command.Stderr(log.Entry().Writer())

	parameters := getParametersFromConfig(config, &http.Client{})

	err := command.RunExecutable(mavenExecutable, parameters...)
	if err != nil {
		log.Entry().
			WithError(err).
			WithField("command", append([]string{mavenExecutable}, parameters...)).
			Fatal("failed to execute run command")
	}

	if stdOutBuf == nil {
		return "", nil
	}
	return string(stdOutBuf.Bytes()), nil
}

func evaluateStdOut(config *mavenExecuteOptions) (*bytes.Buffer, io.Writer) {
	var stdOutBuf *bytes.Buffer
	var stdOut io.Writer

	stdOut = log.Entry().Writer()
	if config.ReturnStdout {
		stdOutBuf = new(bytes.Buffer)
		stdOut = io.MultiWriter(stdOut, stdOutBuf)
	}
	return stdOutBuf, stdOut
}

func getParametersFromConfig(config *mavenExecuteOptions, client http.Downloader) []string {
	var parameters []string

	if config.GlobalSettingsFile != "" {
		globalSettingsFileParameter := "--global-settings " + config.GlobalSettingsFile
		if strings.HasPrefix(config.GlobalSettingsFile, "http") {
			downloadSettingsFromURL(config.ProjectSettingsFile, "globalSettings.xml", client)
			globalSettingsFileParameter = "--global-settings " + "globalSettings.xml"
		}
		parameters = append(parameters, globalSettingsFileParameter)
	}
	// do we need some global state to store that a settings.xml was downloaded and should always be used?
	if config.ProjectSettingsFile != "" {
		projectSettingsFileParameter := "--settings " + config.ProjectSettingsFile
		if strings.HasPrefix(config.ProjectSettingsFile, "http") {
			downloadSettingsFromURL(config.ProjectSettingsFile, "projectSettings.xml", client)
			projectSettingsFileParameter = "--settings " + "projectSettings.xml"
		}
		parameters = append(parameters, projectSettingsFileParameter)
	}

	if config.M2Path != "" {
		m2PathParameter := "-Dmaven.repo.local=" + config.M2Path
		parameters = append(parameters, m2PathParameter)
	}

	if config.PomPath != "" {
		pomPathParameter := "--file " + config.PomPath
		parameters = append(parameters, pomPathParameter)
	}

	if config.Flags != nil {
		parameters = append(parameters, config.Flags...)
	}

	parameters = append(parameters, "--batch-mode")

	if config.LogSuccessfulMavenTransfers {
		parameters = append(parameters, "-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn")
	}

	parameters = append(parameters, config.Goals...)
	return parameters
}

// ToDo replace with pkg/maven/settings GetSettingsFile
func downloadSettingsFromURL(url, filename string, client http.Downloader) {
	err := client.DownloadFile(url, filename, nil, nil)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to download maven settings from: " + url)
	}
}
