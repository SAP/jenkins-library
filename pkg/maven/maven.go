package maven

import (
	"bytes"

	"io"
	"strings"

	"github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
)

type ExecuteOptions struct {
	PomPath                     string
	ProjectSettingsFile         string
	GlobalSettingsFile          string
	M2Path                      string
	Goals                       []string
	Defines                     []string
	Flags                       []string
	LogSuccessfulMavenTransfers bool
	ReturnStdout                bool
}

type mavenExecRunner interface {
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	RunExecutable(e string, p ...string) error
}

const mavenExecutable = "mvn"

func Execute(options *ExecuteOptions, command mavenExecRunner) (string, error) {
	stdOutBuf := new(bytes.Buffer)
	stdOut := io.MultiWriter(log.Entry().Writer(), stdOutBuf)

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

	return string(stdOutBuf.Bytes()), err
}

func getParametersFromOptions(options *ExecuteOptions, client http.Downloader) []string {
	var parameters []string

	if options.GlobalSettingsFile != "" {
		globalSettingsFileName := options.GlobalSettingsFile
		if strings.HasPrefix(options.GlobalSettingsFile, "http:") || strings.HasPrefix(options.GlobalSettingsFile, "https:") {
			downloadSettingsFromURL(options.ProjectSettingsFile, "globalSettings.xml", client)
			globalSettingsFileName = "globalSettings.xml"
		}
		parameters = append(parameters, "--global-settings", globalSettingsFileName)
	}

	if options.ProjectSettingsFile != "" {
		projectSettingsFileName := options.ProjectSettingsFile
		if strings.HasPrefix(options.ProjectSettingsFile, "http:") || strings.HasPrefix(options.ProjectSettingsFile, "https:") {
			downloadSettingsFromURL(options.ProjectSettingsFile, "projectSettings.xml", client)
			projectSettingsFileName = "projectSettings.xml"
		}
		parameters = append(parameters, "--settings", projectSettingsFileName)
	}

	if options.M2Path != "" {
		parameters = append(parameters, "-Dmaven.repo.local="+options.M2Path)
	}

	if options.PomPath != "" {
		parameters = append(parameters, "--file", options.PomPath)
	}

	if options.Flags != nil {
		parameters = append(parameters, options.Flags...)
	}

	if options.Defines != nil {
		parameters = append(parameters, options.Defines...)
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
