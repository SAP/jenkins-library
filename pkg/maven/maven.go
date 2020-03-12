package maven

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

// ExecuteOptions are used by Execute() to construct the Maven command line.
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

// Execute constructs a mvn command line from the given options, and uses the provided
// mavenExecRunner to execute it.
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

// Evaluate constructs ExecuteOptions for using the maven-help-plugin's 'evaluate' goal to
// evaluate a given expression from a pom file. This allows to retrieve the value of - for
// example - 'project.version' from a pom file exactly as Maven itself evaluates it.
func Evaluate(pomFile, expression string, command mavenExecRunner) (string, error) {
	expressionDefine := "-Dexpression=" + expression
	options := ExecuteOptions{
		PomPath:      pomFile,
		Goals:        []string{"org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate"},
		Defines:      []string{expressionDefine, "-DforceStdout", "-q"},
		ReturnStdout: true,
	}
	value, err := Execute(&options, command)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(value, "null object or invalid expression") {
		return "", fmt.Errorf("expression '%s' in file '%s' could not be resolved", expression, pomFile)
	}
	return value, nil
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

func GetTestModulesExcludes() []string {
	var excludes []string
	exists, _ := piperutils.FileExists("unit-tests/pom.xml")
	if exists {
		excludes = append(excludes, "-pl", "!unit-tests")
	}
	exists, _ = piperutils.FileExists("integration-tests/pom.xml")
	if exists {
		excludes = append(excludes, "-pl", "!integration-tests")
	}
	return excludes
}
