//go:build integration
// +build integration

package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/magiconair/properties/assert"
	"github.com/pkg/errors"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
)

// The functions in this file provide a convenient way to integration test the piper binary in docker containers.
// It follows the "given, when, then" approach for structuring tests.
// The general concept is that per test one container is started, one piper command is run and outcomes are asserted.
// Please note that so far this was only tested with debian/ubuntu based containers.
//
// Non-exhaustive list of assumptions those functions make:
// - the following commands are available in the container: sh, chown, sleep
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
	Network     string
	ExecNoLogin bool
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
	Network       string
	ContainerName string
	ExecNoLogin   bool
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
		Network:       bundle.Network,
		ExecNoLogin:   bundle.ExecNoLogin,
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
		tempDir, err := os.MkdirTemp("", "piper-integration-test")
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

	if testRunner.Mounts != nil {
		wd, _ := os.Getwd()
		for src, dst := range testRunner.Mounts {
			localSrc := path.Join(wd, src)
			params = append(params, "-v", fmt.Sprintf("%s:%s", localSrc, dst))
		}
	}

	if testRunner.Network != "" {
		params = append(params, "--network", testRunner.Network)
	}
	params = append(params, testRunner.Image, "sleep", "2000")

	err := testRunner.Runner.RunExecutable("docker", params...)
	if err != nil {
		t.Fatalf("Starting test container has failed %s", err)
	}

	if len(bundle.TestDir) > 0 && testRunner.User != "" {
		err = testRunner.Runner.RunExecutable("docker", "exec", "-u=root", testRunner.ContainerName, "chown", "-R", testRunner.User, "/project")
		if err != nil {
			t.Fatalf("Chown /project has failed %s", err)
		}
	}

	if err = testRunner.Runner.RunExecutable(
		"docker", "exec", testRunner.ContainerName, "sh", "-c",
		strings.Join(testRunner.Setup, "\n"),
	); err != nil {
		t.Fatalf("Running setup script in test container has failed %s", err)
	}

	setupPiperBinary(t, testRunner, localPiper)

	return testRunner
}

// generateContainerName creates a name with a common prefix and a random number, so we can start a new container for each test method
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
	err = testRunner.Runner.RunExecutable("docker", "exec", testRunner.ContainerName, "/bin/sh", "/piper-wrapper", "/piper", "version")
	if err != nil {
		t.Fatalf("Running piper failed. "+
			"Please check that '%s' is the correct binary, and is compiled for this configuration: 'GOOS=linux GOARCH=amd64'. Error text: %s", localPiper, err)
	}
}

func (d *IntegrationTestDockerExecRunner) whenRunningPiperCommand(command string, parameters ...string) error {
	args := []string{"exec", "--workdir", "/project", d.ContainerName, "/bin/sh"}

	if !d.ExecNoLogin {
		args = append(args, "-l")
	}

	args = append(args, "/piper-wrapper", "/piper", command)
	err := d.Runner.RunExecutable("docker", append(args, parameters...)...)
	if err != nil {
		stdOut, outputErr := d.getPiperOutput()
		if outputErr != nil {
			return errors.Wrap(outputErr, "unable to get output after Piper command failure")
		}
		return errors.Wrapf(err, "piper output: \n%s", stdOut.String())
	}
	return err
}

func (d *IntegrationTestDockerExecRunner) runScriptInsideContainer(script string) error {
	args := []string{"exec", "--workdir", "/project", d.ContainerName, "/bin/sh"}

	if !d.ExecNoLogin {
		args = append(args, "-l")
	}

	args = append(args, "-c", script)
	return d.Runner.RunExecutable("docker", args...)
}

func (d *IntegrationTestDockerExecRunner) assertHasNoOutput(t *testing.T, inconsistencies ...string) {
	count := len(inconsistencies)
	buffer, err := d.getPiperOutput()
	if err != nil {
		t.Fatalf("Failed to get log output of container %s", d.ContainerName)
	}
	scanner := bufio.NewScanner(buffer)
	for scanner.Scan() && (len(inconsistencies) != 0) {
		for i, str := range inconsistencies {
			if strings.Contains(scanner.Text(), str) {
				inconsistencies = append(inconsistencies[:i], inconsistencies[i+1:]...)
				break
			}
		}
	}
	assert.Equal(t, len(inconsistencies), count, fmt.Sprintf(
		"[assertHasNoOutput] Unexpected command output:\n%s\n%s\n", buffer.String(), strings.Join(inconsistencies, "\n")),
	)
}

func (d *IntegrationTestDockerExecRunner) assertHasOutput(t *testing.T, consistencies ...string) {
	buffer, err := d.getPiperOutput()
	if err != nil {
		t.Fatalf("Failed to get log output of container %s", d.ContainerName)
	}
	scanner := bufio.NewScanner(buffer)
	for scanner.Scan() && (len(consistencies) != 0) {
		for i, str := range consistencies {
			if strings.Contains(scanner.Text(), str) {
				consistencies = append(consistencies[:i], consistencies[i+1:]...)
				break
			}
		}
	}
	assert.Equal(t, len(consistencies), 0, fmt.Sprintf(
		"[assertHasOutput] Unexpected command output:\n%s\n%s\n", buffer.String(), strings.Join(consistencies, "\n")),
	)
}

func (d *IntegrationTestDockerExecRunner) getPiperOutput() (*bytes.Buffer, error) {
	buffer := new(bytes.Buffer)
	d.Runner.Stdout(buffer)
	err := d.Runner.RunExecutable("docker", "exec", d.ContainerName, "cat", "/tmp/test-log.txt")
	d.Runner.Stdout(log.Writer())
	return buffer, err
}

func (d *IntegrationTestDockerExecRunner) assertHasFiles(t *testing.T, consistencies ...string) {
	buffer := new(bytes.Buffer)
	d.Runner.Stderr(buffer)
	if d.Runner.RunExecutable(
		"docker",
		append(append(make([]string, 0), "exec", d.ContainerName, "stat"), consistencies...)...,
	) != nil {
		t.Fatalf("[assertHasFiles] Assertion has failed: %v", errors.New(buffer.String()))
	}
}

func (d *IntegrationTestDockerExecRunner) assertFileContentEquals(t *testing.T, fileWant string, contentWant string) {
	d.assertHasFiles(t, fileWant)

	buffer := new(bytes.Buffer)
	d.Runner.Stdout(buffer)
	err := d.Runner.RunExecutable("docker", "cp", d.ContainerName+":/"+fileWant, "-")
	if err != nil {
		t.Fatalf("Copy file has failed. Expected file %s to exist in container. %s", fileWant, err)
	}

	tarReader := tar.NewReader(buffer)
	header, err := tarReader.Next()
	if err == io.EOF {
		t.Fatal("Empty tar received")
	}
	if err != nil {
		t.Fatalf("Cant read tar: %s", err)
	}
	if header.Typeflag != tar.TypeReg {
		t.Fatalf("Expected a file, but received %c", header.Typeflag)
	}
	str := new(bytes.Buffer)
	_, err = io.Copy(str, tarReader)
	if err != nil {
		t.Fatalf("unable to get tar file content: %s", err)
	}

	assert.Equal(t, str.String(), contentWant, fmt.Sprintf("Unexpected content of file '%s'", fileWant))
}

func (d *IntegrationTestDockerExecRunner) terminate(t *testing.T) {
	err := d.Runner.RunExecutable("docker", "rm", "-f", d.ContainerName)
	if err != nil {
		t.Fatalf("Failed to terminate container '%s'", d.ContainerName)
	}
}
