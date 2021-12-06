package gradle

import (
	"bytes"
	"fmt"
	"io"

	"github.com/SAP/jenkins-library/pkg/log"
)

const exec = "gradle"

type Utils interface {
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	RunExecutable(e string, p ...string) error
}

// ExecuteOptions are used by Execute() to construct the Gradle command line.
type ExecuteOptions struct {
	BuildGradlePath string `json:"path,omitempty"`
	Task            string `json:"task,omitempty"`
	ReturnStdout    bool   `json:"returnStdout,omitempty"`
}

func Execute(options *ExecuteOptions, utils Utils) (string, error) {

	stdOutBuf, stdOut := evaluateStdOut(options)
	utils.Stdout(stdOut)
	utils.Stderr(log.Writer())

	parameters, err := getParametersFromOptions(options, utils)
	if err != nil {
		return "", fmt.Errorf("failed to construct parameters from options: %w", err)
	}

	err = utils.RunExecutable(exec, parameters...)
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		commandLine := append([]string{exec}, parameters...)
		return "", fmt.Errorf("failed to run executable, command: '%s', error: %w", commandLine, err)
	}

	if stdOutBuf == nil {
		return "", nil
	}
	return string(stdOutBuf.Bytes()), nil
}

func evaluateStdOut(options *ExecuteOptions) (*bytes.Buffer, io.Writer) {
	var stdOutBuf *bytes.Buffer
	stdOut := log.Writer()
	if options.ReturnStdout {
		stdOutBuf = new(bytes.Buffer)
		stdOut = io.MultiWriter(stdOut, stdOutBuf)
	}
	return stdOutBuf, stdOut
}

func getParametersFromOptions(options *ExecuteOptions, utils Utils) ([]string, error) {
	var parameters []string

	// default value for task is 'build', so no necessary to checking for empty parameter
	parameters = append(parameters, options.Task)

	// resolve path for build.gradle execution
	if options.BuildGradlePath != "" {
		parameters = append(parameters, "-p")
		parameters = append(parameters, options.BuildGradlePath)
	}

	return parameters, nil
}
