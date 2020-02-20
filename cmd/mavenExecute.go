package cmd

import (
	"bytes"
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
	var stdOutBuf *bytes.Buffer
	var stdOut io.Writer

	stdOut = log.Entry().Writer()
	if config.ReturnStdout {
		stdOutBuf = new(bytes.Buffer)
		stdOut = io.MultiWriter(stdOut, stdOutBuf)
	}
	command.Stdout(stdOut)
	command.Stderr(log.Entry().Writer())

	parameters := []string{}

	if config.GlobalSettingsFile != "" {
		globalSettingsFileParameter := "--global-settings " + config.GlobalSettingsFile
		if strings.HasPrefix(config.GlobalSettingsFile, "http") {
			globalSettingsFileParameter = "--global-settings " + downloadSettingsFromUrl(config.GlobalSettingsFile)
		}
		parameters = append(parameters, globalSettingsFileParameter)
	}

	if config.ProjectSettingsFile != "" {
		projectSettingsFileParameter := "--settings " + config.ProjectSettingsFile
		if strings.HasPrefix(config.ProjectSettingsFile, "http") {
			projectSettingsFileParameter = "--settings " + downloadSettingsFromUrl(config.ProjectSettingsFile)
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

func downloadSettingsFromURL(settingsURL string) string {
	//client := &p

	return "fileName"
}
