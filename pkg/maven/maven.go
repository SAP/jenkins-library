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

func Execute(options *ExecuteOptions, command mavenExecRunner) (string, error) {
	stdOutBuf, stdOut := evaluateStdOut(options)
	command.Stdout(stdOut)
	command.Stderr(log.Entry().Writer())

	parameters := getParametersFromOptions(options, &http.Client{})

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

func getParametersFromOptions(options *ExecuteOptions, client http.Downloader) []string {
	var parameters []string

	if options.GlobalSettingsFile != "" {
		globalSettingsFileParameter := "--global-settings " + options.GlobalSettingsFile
		if strings.HasPrefix(options.GlobalSettingsFile, "http:") || strings.HasPrefix(options.GlobalSettingsFile, "https:") {
			downloadSettingsFromURL(options.ProjectSettingsFile, "globalSettings.xml", client)
			globalSettingsFileParameter = "--global-settings " + "globalSettings.xml"
		}
		parameters = append(parameters, globalSettingsFileParameter)
	}

	if options.ProjectSettingsFile != "" {
		projectSettingsFileParameter := "--settings " + options.ProjectSettingsFile
		if strings.HasPrefix(options.ProjectSettingsFile, "http:") || strings.HasPrefix(options.ProjectSettingsFile, "https:") {
			downloadSettingsFromURL(options.ProjectSettingsFile, "projectSettings.xml", client)
			projectSettingsFileParameter = "--settings " + "projectSettings.xml"
		}
		parameters = append(parameters, projectSettingsFileParameter)
	}

	if options.M2Path != "" {
		m2PathParameter := "-Dmaven.repo.local=" + options.M2Path
		parameters = append(parameters, m2PathParameter)
	}

	if options.PomPath != "" {
		pomPathParameter := "--file " + options.PomPath
		parameters = append(parameters, pomPathParameter)
	}

	if options.Flags != nil {
		parameters = append(parameters, options.Flags...)
	}

	parameters = append(parameters, "--batch-mode")

	if options.LogSuccessfulMavenTransfers {
		parameters = append(parameters, "-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn")
	}

	parameters = append(parameters, options.Goals...)
	return parameters
}

// ToDo replace with pkg/maven/settings GetSettingsFile
func downloadSettingsFromURL(url, filename string, client http.Downloader) {
	err := client.DownloadFile(url, filename, nil, nil)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to download maven settings from: " + url)
	}
}
