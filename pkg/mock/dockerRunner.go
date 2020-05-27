// +build !release

package mock

import (
	"fmt"
	"io"
	"os"
)

type baseRunner interface {
	SetDir(d string)
	SetEnv(e []string)
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	RunExecutable(e string, p ...string) error
}

// DockerExecConfig is the configuration for an individual tool that shall be executed on docker.
type DockerExecConfig struct {
	// Image is the fully qualified docker image name that is passed to docker for this tool.
	Image string
	// Workspace is the (optional) directory to which the current working directory is mapped
	// within the docker container.
	Workspace string
}

// DockerExecRunner can be used "in place" of another ExecRunner in order to transparently
// execute commands within a docker container. One use-case is to test locally with tools
// that are not available on the current platform.
type DockerExecRunner struct {
	// Runner is the ExecRunner to which all executions are forwarded in the end.
	Runner            baseRunner
	executablesToWrap map[string]DockerExecConfig
}

func (d *DockerExecRunner) SetDir(dir string) {
	d.Runner.SetDir(dir)
}

func (d *DockerExecRunner) SetEnv(env []string) {
	d.Runner.SetEnv(env)
}

func (d *DockerExecRunner) Stdout(out io.Writer) {
	d.Runner.Stdout(out)
}

func (d *DockerExecRunner) Stderr(err io.Writer) {
	d.Runner.Stderr(err)
}

func (d *DockerExecRunner) AddExecConfig(executable string, config DockerExecConfig) error {
	if executable == "" {
		return fmt.Errorf("'executable' needs to be provided")
	}
	if config.Image == "" {
		return fmt.Errorf("the DockerExecConfig must specify a docker image")
	}
	if d.executablesToWrap == nil {
		d.executablesToWrap = map[string]DockerExecConfig{}
	}
	d.executablesToWrap[executable] = config
	return nil
}

func (d *DockerExecRunner) RunExecutable(executable string, parameters ...string) error {
	if config, ok := d.executablesToWrap[executable]; ok {
		wrappedParameters := []string{"run", "--entrypoint=" + executable}
		if config.Workspace != "" {
			currentDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory for mounting in docker: %w", err)
			}
			wrappedParameters = append(wrappedParameters, "-v", currentDir+":"+config.Workspace)
		}
		wrappedParameters = append(wrappedParameters, config.Image)
		wrappedParameters = append(wrappedParameters, parameters...)
		executable = "docker"
		parameters = wrappedParameters
	}
	return d.Runner.RunExecutable(executable, parameters...)
}
