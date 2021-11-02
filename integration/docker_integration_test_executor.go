//go:build integration
// +build integration

package main

import (
	"bytes"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"
)

// The functions in this file provide a convenient way to integration test the piper binary in docker containers.
// It follows the "given, when, then" approach for structuring tests.
// The general concept is that per test one container is started, one piper command is run and outcomes are asserted.
// Please note that so far this was only tested with debian/ubuntu based containers.
//
// Non-exhaustive list of assumptions those functions make:
// - Bash is available in the container
// - If the option TestDir is not provided, the test project must be in the container image  in the directory /project

// IntegrationTestDockerExecRunnerBundle is used to construct an instance of IntegrationTestDockerExecRunner
// This is what a test uses to specify the container it requires
type IntegrationTestDockerExecRunnerBundle struct {
	Image       string
	User        string
	TestDir     []string
	Mounts      map[string]string
	Environment map[string]string
	Setup       []string
}

// IntegrationTestDockerExecRunner keeps the state of an instance of a docker runner
type IntegrationTestDockerExecRunner struct {
	// Runner is the ExecRunner to which all executions are forwarded in the end.
	Runner        command.Command
	Image         string
	User          string
	TestDir       []string
	Mounts        map[string]string
	Environment   map[string]string
	Setup         []string
	ContainerName string
}

func givenThisContainer(t *testing.T, bundle IntegrationTestDockerExecRunnerBundle) IntegrationTestDockerExecRunner {
	runner := command.Command{}
	containerName := generateContainerName()

	testRunner := IntegrationTestDockerExecRunner{
		Runner:        runner,
		Image:         bundle.Image,
		User:          bundle.User,
		Mounts:        bundle.Mounts,
		Environment:   bundle.Environment,
		Setup:         bundle.Setup,
		ContainerName: containerName,
	}

	wd, _ := os.Getwd()
	localPiper := path.Join(wd, "..", "piper")
	if localPiper == "" {
		t.Fatal("Could not locate piper binary to test")
	}

	params := []string{"run", "--detach", "-v", localPiper + ":/piper", "--name=" + testRunner.ContainerName}
	if testRunner.User != "" {
		params = append(params, fmt.Sprintf("--user=%s", testRunner.User))
	}
	if len(bundle.TestDir) > 0 {
		projectDir := path.Join(wd, path.Join(bundle.TestDir...))
		// 1. Copy test files to a temp dir in order to avoid non-repeatable test executions because of changed state
		// 2. Don't remove the temp dir to allow investigation of failed tests. Maybe add an option for cleaning it later?
		tempDir, err := ioutil.TempDir("", "piper-integration-test")
		if err != nil {
			t.Fatal(err)
		}

		err = copyDir(projectDir, tempDir)
		if err != nil {
			t.Fatalf("Failed to copy files from %s into %s", projectDir, tempDir)
		}
		params = append(params, "-v", fmt.Sprintf("%s:/project", tempDir))
	}
	if len(testRunner.Environment) > 0 {
		for envVarName, envVarValue := range testRunner.Environment {
			params = append(params, "--env", fmt.Sprintf("%s=%s", envVarName, envVarValue))
		}
	}
	params = append(params, testRunner.Image, "sleep", "2000")

	//todo mounts
	err := testRunner.Runner.RunExecutable("docker", params...)
	if err != nil {
		t.Fatalf("Starting test container has failed %s", err)
	}

	if len(bundle.TestDir) > 0 {
		err = testRunner.Runner.RunExecutable("docker", "exec", "-u=root", testRunner.ContainerName, "chown", "-R", testRunner.User, "/project")
		if err != nil {
			t.Fatalf("Chown /project has failed %s", err)
		}
	}

	for _, scriptLine := range testRunner.Setup {
		err := testRunner.Runner.RunExecutable("docker", "exec", testRunner.ContainerName, "/bin/bash", "-c", scriptLine)
		if err != nil {
			t.Fatalf("Running setup script in test container has failed %s", err)
		}
	}

	setupPiperBinary(t, testRunner, localPiper)

	return testRunner
}

// generateContainerName creates a name with a common prefix and a random number so we can start a new container for each test method
// We don't rely on docker's random name generator for two reasons
// First, it is easier to save the name here compared to getting it from stdout
// Second, the common prefix allows batch stopping/deleting of containers if so desired
// The test code will not automatically delete containers as they might be useful for debugging
func generateContainerName() string {
	var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	return "piper-integration-test-" + strconv.Itoa(seededRand.Int())
}

// setupPiperBinary copies a wrapper script for calling the piper binary into the container and verifies that the piper binary is executable inside the container
// The wrapper script (piper-command-wrapper.sh) only calls the piper binary and redirects its output into a file
// The purpose of this is to capture piper's stdout/stderr in order to assert on the output
// This is not possible via "docker logs", cf https://github.com/moby/moby/issues/8662
func setupPiperBinary(t *testing.T, testRunner IntegrationTestDockerExecRunner, localPiper string) {
	err := testRunner.Runner.RunExecutable("docker", "cp", "piper-command-wrapper.sh", testRunner.ContainerName+":/piper-wrapper")
	if err != nil {
		t.Fatalf("Copying command wrapper to container has failed %s", err)
	}
	err = testRunner.Runner.RunExecutable("docker", "exec", "-u=root", testRunner.ContainerName, "chmod", "+x", "/piper-wrapper")
	if err != nil {
		t.Fatalf("Making command wrapper in container executable has failed %s", err)
	}
	err = testRunner.Runner.RunExecutable("docker", "exec", testRunner.ContainerName, "/bin/bash", "/piper-wrapper", "/piper", "version")
	if err != nil {
		t.Fatalf("Running piper failed. "+
			"Please check that '%s' is the correct binary, and is compiled for this configuration: 'GOOS=linux GOARCH=amd64'. Error text: %s", localPiper, err)
	}
}

func (d *IntegrationTestDockerExecRunner) whenRunningPiperCommand(command string, parameters ...string) error {
	args := []string{"exec", "--workdir", "/project", d.ContainerName, "/bin/bash", "--login", "/piper-wrapper", "/piper", command}
	args = append(args, parameters...)
	return d.Runner.RunExecutable("docker", args...)
}

func (d *IntegrationTestDockerExecRunner) runScriptInsideContainer(script string) error {
	args := []string{"exec", "--workdir", "/project", d.ContainerName, "/bin/bash", "--login", "-c", script}
	return d.Runner.RunExecutable("docker", args...)
}

func (d *IntegrationTestDockerExecRunner) assertHasOutput(t *testing.T, want string) {
	buffer := new(bytes.Buffer)
	d.Runner.Stdout(buffer)
	err := d.Runner.RunExecutable("docker", "exec", d.ContainerName, "cat", "/tmp/test-log.txt")
	d.Runner.Stdout(log.Writer())
	if err != nil {
		t.Fatalf("Failed to get log output of container %s", d.ContainerName)
	}

	if !strings.Contains(buffer.String(), want) {
		t.Fatalf("Assertion has failed. Expected output %s in command output.\n%s", want, buffer.String())
	}
}

func (d *IntegrationTestDockerExecRunner) assertHasFile(t *testing.T, want string) {
	err := d.Runner.RunExecutable("docker", "exec", d.ContainerName, "stat", want)
	if err != nil {
		t.Fatalf("Assertion has failed. Expected file %s to exist in container. %s", want, err)
	}
}
