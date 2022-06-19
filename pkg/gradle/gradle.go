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

type (
	cancelFunc func()
	Utils      interface {
		Stdout(out io.Writer)
		Stderr(err io.Writer)
		RunExecutable(e string, p ...string) error

		FileExists(filename string) (bool, error)
		FileWrite(path string, content []byte, perm os.FileMode) error
		FileRemove(path string) error
	}
	// ExecuteOptions are used by Execute() to construct the Gradle command line.
	ExecuteOptions struct {
		BuildGradlePath   string            `json:"path,omitempty"`
		Tasks             []string          `json:"tasks,omitempty"`
		InitScriptTasks   []string          `json:"initScriptTasks,omitempty"`
		SkipTasks         []string          `json:"skipTasks,omitempty"`
		InitScriptContent string            `json:"initScriptContent,omitempty"`
		UseWrapper        bool              `json:"useWrapper,omitempty"`
		ProjectProperties map[string]string `json:"projectProperties,omitempty"`
		setInitScript     bool
	}
)

var cancelNothing = func() {}

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

	err, cancel := handleInitTasks(exec, options, utils, stdOutBuf)
	if err != nil {
		return "", err
	}
	defer cancel()
	parameters := getParametersFromOptions(options)
	err = utils.RunExecutable(exec, parameters...)
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		commandLine := append([]string{exec}, parameters...)
		return "", fmt.Errorf("failed to run executable, command: '%s', error: %v", commandLine, err)
	}
	return string(stdOutBuf.Bytes()), nil
}

func handleInitTasks(exec string, options *ExecuteOptions, utils Utils, stdOutBuf *bytes.Buffer) (error, cancelFunc) {
	cancel := cancelNothing
	if options.InitScriptContent != "" {
		hasTasks, err := hasInitTasks(exec, options, utils, stdOutBuf)
		if err != nil {
			return fmt.Errorf("failed list gradle tasks: %v", err), nil
		}
		if !hasTasks {
			err := utils.FileWrite(initScriptName, []byte(options.InitScriptContent), 0644)
			if err != nil {
				return fmt.Errorf("failed create init script: %v", err), nil
			}
			cancel = func() {
				utils.FileRemove(initScriptName)
			}
			options.setInitScript = true
			hasTasks, err := hasInitTasks(exec, options, utils, stdOutBuf)
			if err != nil {
				return fmt.Errorf("failed list gradle tasks with init script: %v", err), nil
			}
			if !hasTasks {
				options.InitScriptTasks = nil
			}
		}
	} else {
		options.InitScriptTasks = nil
	}
	return nil, cancel
}

func hasInitTasks(exec string, options *ExecuteOptions, utils Utils, stdOutBuf *bytes.Buffer) (bool, error) {
	parameters := []string{"tasks"}
	if options.BuildGradlePath != "" {
		parameters = append(parameters, "-p", options.BuildGradlePath)
	}
	if options.setInitScript {
		parameters = append(parameters, "--init-script", initScriptName)
	}
	if err := utils.RunExecutable(exec, parameters...); err != nil {
		return false, err
	}
	tasksOut := stdOutBuf.String()
	for _, task := range options.InitScriptTasks {
		if !strings.Contains(tasksOut, task) {
			return false, nil
		}
	}
	return true, nil
}

func getParametersFromOptions(options *ExecuteOptions) []string {
	var parameters []string

	// default value for task is 'build', so no necessary to checking for empty parameter
	parameters = append(parameters, append(options.Tasks, options.InitScriptTasks...)...)
	log.Entry().Infof("the folowing tasks will be called: %v", parameters)
	for _, task := range options.SkipTasks {
		parameters = append(parameters, "-x", task)
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
