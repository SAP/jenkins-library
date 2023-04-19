package gradle

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

	FileExists(filename string) (bool, error)
	FileWrite(path string, content []byte, perm os.FileMode) error
	FileRemove(path string) error
}

// ExecuteOptions are used by Execute() to construct the Gradle command line.
type ExecuteOptions struct {
	BuildGradlePath   string            `json:"path,omitempty"`
	Task              string            `json:"task,omitempty"`
	BuildFlags        []string          `json:"buildFlags,omitempty"`
	InitScriptContent string            `json:"initScriptContent,omitempty"`
	UseWrapper        bool              `json:"useWrapper,omitempty"`
	ProjectProperties map[string]string `json:"projectProperties,omitempty"`
	setInitScript     bool
}

func Execute(options *ExecuteOptions, utils Utils) (string, error) {
	stdOutBuf := new(bytes.Buffer)
	utils.Stdout(io.MultiWriter(log.Writer(), stdOutBuf))
	utils.Stderr(log.Writer())

	_, err := searchBuildScript([]string{
		filepath.Join(options.BuildGradlePath, groovyBuildScriptName),
		filepath.Join(options.BuildGradlePath, kotlinBuildScriptName),
	}, utils.FileExists)
	if err != nil {
		return "", fmt.Errorf("the specified gradle build script could not be found: %v", err)
	}

	exec := gradleExecutable
	if options.UseWrapper {
		wrapperExists, err := utils.FileExists("gradlew")
		if err != nil {
			return "", err
		}
		if !wrapperExists {
			return "", errors.New("gradle wrapper not found")
		}
		exec = gradlewExecutable
	}
	log.Entry().Infof("All commands will be executed with the '%s' tool", exec)

	if options.InitScriptContent != "" {
		parameters := []string{"tasks"}
		if options.BuildGradlePath != "" {
			parameters = append(parameters, "-p", options.BuildGradlePath)
		}
		if err := utils.RunExecutable(exec, parameters...); err != nil {
			return "", fmt.Errorf("failed list gradle tasks: %v", err)
		}
		if !strings.Contains(stdOutBuf.String(), options.Task) {
			err := utils.FileWrite(initScriptName, []byte(options.InitScriptContent), 0644)
			if err != nil {
				return "", fmt.Errorf("failed create init script: %v", err)
			}
			defer utils.FileRemove(initScriptName)
			options.setInitScript = true
		}
	}

	parameters := getParametersFromOptions(options)

	err = utils.RunExecutable(exec, parameters...)
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		commandLine := append([]string{exec}, parameters...)
		return "", fmt.Errorf("failed to run executable, command: '%s', error: %v", commandLine, err)
	}

	return string(stdOutBuf.Bytes()), nil
}

func getParametersFromOptions(options *ExecuteOptions) []string {
	var parameters []string

	if len(options.BuildFlags) > 0 {
		// respect the list of tasks/flags user wants to execute
		parameters = append(parameters, options.BuildFlags...)
	} else {
		// default value for task is 'build', so no necessary to checking for empty parameter
		parameters = append(parameters, options.Task)
	}

	// resolve path for build.gradle execution
	if options.BuildGradlePath != "" {
		parameters = append(parameters, "-p", options.BuildGradlePath)
	}

	for k, v := range options.ProjectProperties {
		parameters = append(parameters, fmt.Sprintf("-P%s=%s", k, v))
	}

	if options.setInitScript {
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
