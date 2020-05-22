// +build !release

package mock

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"io"
	"os"
)

type DockerExecRunner struct {
	runner            command.Command
	ExecutablesToWrap []string
	DockerImage       string
	DockerWorkspace   string
}

func (d *DockerExecRunner) SetDir(dir string) {
	d.runner.SetDir(dir)
}

func (d *DockerExecRunner) SetEnv(env []string) {
	d.runner.SetEnv(env)
}

func (d *DockerExecRunner) Stdout(out io.Writer) {
	d.runner.Stdout(out)
}

func (d *DockerExecRunner) Stderr(err io.Writer) {
	d.runner.Stderr(err)
}

func (d *DockerExecRunner) RunExecutable(executable string, parameters ...string) error {
	if piperutils.ContainsString(d.ExecutablesToWrap, executable) {
		if d.DockerImage == "" {
			return fmt.Errorf("no docker image specified")
		}
		wrappedParameters := []string{"run", "--entrypoint=" + executable}
		if d.DockerWorkspace != "" {
			currentDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory for mounting in docker: %w", err)
			}
			wrappedParameters = append(wrappedParameters, "-v", currentDir+":"+d.DockerWorkspace)
		}
		wrappedParameters = append(wrappedParameters, d.DockerImage)
		wrappedParameters = append(wrappedParameters, parameters...)
		executable = "docker"
		parameters = wrappedParameters
	}
	return d.runner.RunExecutable(executable, parameters...)
}
