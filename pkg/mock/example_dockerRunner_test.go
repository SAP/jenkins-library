//go:build unit
// +build unit

package mock_test

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/mock"
)

type ExecRunner interface {
	SetDir(d string)
	SetEnv(e []string)
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	RunExecutable(e string, p ...string) error
}

func getMavenVersion(runner ExecRunner) (string, error) {
	output := bytes.Buffer{}
	runner.Stdout(&output)
	err := runner.RunExecutable("mvn", "--version")
	if err != nil {
		return "", fmt.Errorf("failed to run maven: %w", err)
	}
	logLines := strings.Split(output.String(), "\n")
	if len(logLines) < 1 {
		return "", fmt.Errorf("failed to obtain maven output")
	}
	return logLines[0], nil
}

func ExampleDockerExecRunner_RunExecutable() {
	// getMavenVersion(runner ExecRunner) executes the command "mvn --version"
	// and returns the command output as string
	runner := command.Command{}
	localMavenVersion, _ := getMavenVersion(&runner)

	dockerRunner := mock.DockerExecRunner{Runner: &runner}
	_ = dockerRunner.AddExecConfig("mvn", mock.DockerExecConfig{
		Image: "maven:3.6.1-jdk-8",
	})

	dockerMavenVersion, _ := getMavenVersion(&dockerRunner)

	fmt.Printf("Your local mvn version is %v, while the version in docker is %v", localMavenVersion, dockerMavenVersion)
}
