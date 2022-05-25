package gradle

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
)

const (
	gradleExecutable  = "gradle"
	gradlewExecutable = "./gradlew"

	groovyBuildScriptName = "build.gradle"
	kotlinBuildScriptName = "build.gradle.kts"
	initScriptName        = "initScript.gradle.tmp"
)

type Utils interface {
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	RunExecutable(e string, p ...string) error

	DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error
	Glob(pattern string) (matches []string, err error)
	FileExists(filename string) (bool, error)
	Copy(src, dest string) (int64, error)
	MkdirAll(path string, perm os.FileMode) error
	FileWrite(path string, content []byte, perm os.FileMode) error
	FileRead(path string) ([]byte, error)
	FileRemove(path string) error
}

// ExecuteOptions are used by Execute() to construct the Gradle command line.
type ExecuteOptions struct {
	BuildGradlePath   string `json:"path,omitempty"`
	Task              string `json:"task,omitempty"`
	InitScriptContent string `json:"initScriptContent,omitempty"`
	UseWrapper        bool   `json:"useWrapper,omitempty"`
	ReturnStdout      bool   `json:"returnStdout,omitempty"`
}

func Execute(options *ExecuteOptions, utils Utils) error {
	stdOutBuf, stdOut := evaluateStdOut(options)
	utils.Stdout(stdOut)
	utils.Stderr(log.Writer())

	_, err := searchBuildScript([]string{groovyBuildScriptName, kotlinBuildScriptName}, utils.FileExists)
	if err != nil {
		return err
	}

	exec := gradleExecutable
	if options.UseWrapper {
		wrapperExists, err := utils.FileExists(gradlewExecutable)
		if err != nil {
			return err
		}
		if !wrapperExists {
			return errors.New("gradle wrapper not found")
		}
		exec = gradlewExecutable
	}
	log.Entry().Infof("All commands will be executed with the '%s' tool", exec)

	if options.InitScriptContent != "" {
		if err := utils.RunExecutable(exec, "tasks"); err != nil {
			return fmt.Errorf("failed list gradle tasks: %v", err)
		}
		if !strings.Contains(stdOutBuf.String(), options.Task) {
			err := utils.FileWrite(initScriptName, []byte(options.InitScriptContent), 0644)
			if err != nil {
				return fmt.Errorf("failed create init script: %v", err)
			}
			defer utils.FileRemove(initScriptName)
		}
	}

	parameters := getParametersFromOptions(options)

	err = utils.RunExecutable(exec, parameters...)
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		commandLine := append([]string{exec}, parameters...)
		return fmt.Errorf("failed to run executable, command: '%s', error: %v", commandLine, err)
	}
	return nil
}

func getParametersFromOptions(options *ExecuteOptions) []string {
	var parameters []string

	// default value for task is 'build', so no necessary to checking for empty parameter
	parameters = append(parameters, options.Task)

	// resolve path for build.gradle execution
	if options.BuildGradlePath != "" {
		parameters = append(parameters, "-p", options.BuildGradlePath)
	}

	if options.InitScriptContent != "" {
		parameters = append(parameters, "--init-script", initScriptName)
	}

	return parameters
}

func searchBuildScript(supported []string, existsFunc func(string) (bool, error)) (string, error) {
	var descriptor string
	for _, f := range supported {
		exists, err := existsFunc(f)
		if err != nil {
			return "", err
		}
		if exists {
			descriptor = f
			break
		}
	}
	if len(descriptor) == 0 {
		return "", fmt.Errorf("no build script available, supported: %v", supported)
	}
	return descriptor, nil
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
