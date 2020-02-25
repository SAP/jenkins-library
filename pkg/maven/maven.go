package maven

import (
	"bytes"

	"github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"io"
	"strings"
)

type ExecuteOptions struct {
	PomPath                     string   `json:"pomPath,omitempty"`
	ProjectSettingsFile         string   `json:"projectSettingsFile,omitempty"`
	GlobalSettingsFile          string   `json:"globalSettingsFile,omitempty"`
	M2Path                      string   `json:"m2Path,omitempty"`
	Goals                       []string `json:"goals,omitempty"`
	Defines                     []string `json:"defines,omitempty"`
	Flags                       []string `json:"flags,omitempty"`
	LogSuccessfulMavenTransfers bool     `json:"logSuccessfulMavenTransfers,omitempty"`
	ReturnStdout                bool     `json:"returnStdout,omitempty"`
}

type mavenExecRunner interface {
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	RunExecutable(e string, p ...string) error
}

const mavenExecutable = "mvn"

func Execute(config *ExecuteOptions, command mavenExecRunner) (string, error) {
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

func evaluateStdOut(config *ExecuteOptions) (*bytes.Buffer, io.Writer) {
	var stdOutBuf *bytes.Buffer
	var stdOut io.Writer

	stdOut = log.Entry().Writer()
	if config.ReturnStdout {
		stdOutBuf = new(bytes.Buffer)
		stdOut = io.MultiWriter(stdOut, stdOutBuf)
	}
	return stdOutBuf, stdOut
}

func getParametersFromConfig(config *ExecuteOptions, client http.Downloader) []string {
	var parameters []string

	if config.GlobalSettingsFile != "" {
		globalSettingsFileParameter := "--global-settings " + config.GlobalSettingsFile
		if strings.HasPrefix(config.GlobalSettingsFile, "http:") || strings.HasPrefix(config.GlobalSettingsFile, "https:") {
			downloadSettingsFromURL(config.ProjectSettingsFile, "globalSettings.xml", client)
			globalSettingsFileParameter = "--global-settings " + "globalSettings.xml"
		}
		parameters = append(parameters, globalSettingsFileParameter)
	}

	if config.ProjectSettingsFile != "" {
		projectSettingsFileParameter := "--settings " + config.ProjectSettingsFile
		if strings.HasPrefix(config.ProjectSettingsFile, "http:") || strings.HasPrefix(config.ProjectSettingsFile, "https:") {
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
