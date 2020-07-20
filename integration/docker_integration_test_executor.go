// +build integration

package main

import (
	"bytes"
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

// IntegrationTestDockerExecRunnerBundle is used to construct an instance of IntegrationTestDockerExecRunner
type IntegrationTestDockerExecRunnerBundle struct {
	Image       string
	User        string
	TestDir     []string
	Mounts      map[string]string
	Environment map[string]string
	Setup       []string
}

func givenThisContainer(t *testing.T, bundle IntegrationTestDockerExecRunnerBundle) IntegrationTestDockerExecRunner {
	runner := command.Command{}

	// Generate a random container name so we can start a new one for each test method
	// We don't rely on docker's random name generator for two reasons
	// First, it is easier to save the name here compared to getting it from stdout
	// Second, the common prefix allows batch stopping/deleting of containers if so desired
	// The test code will not automatically delete containers as they might be useful for debugging
	var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	containerName := "piper-integration-test-" + strconv.Itoa(seededRand.Int())

	testRunner := IntegrationTestDockerExecRunner{
		Runner:        runner,
		Image:         bundle.Image,
		User:          bundle.User,
		Mounts:        bundle.Mounts,
		Setup:         bundle.Setup,
		ContainerName: containerName,
	}

	//todo ensure it is a linux binary
	wd, _ := os.Getwd()
	localPiper := path.Join(wd, "..", "piper")
	if localPiper == "" {
		t.Fatal("Could not locate piper binary to test")
	}

	projectDir := path.Join(wd, path.Join(bundle.TestDir...))

	// 1. Copy test files to a temp dir in order to avoid non-repeatable test executions because of changed state
	// 2. Don't remove the temp dir to allow investigation of failed tests. Maybe add an option for cleaning it later?
	tempDir, err := ioutil.TempDir("", "piper-integration-test")
	if err != nil {
		t.Fatal(err)
	}

	err = copyDir(projectDir, tempDir)
	if err != nil {
		t.Fatalf("")
	}

	//todo mounts
	//todo env (secrets)
	err = testRunner.Runner.RunExecutable("docker", "run", "-d", "-u="+testRunner.User,
		"-v", localPiper+":/piper", "-v", tempDir+":/project",
		"--name="+testRunner.ContainerName,
		testRunner.Image,
		"sleep", "2000")
	if err != nil {
		t.Fatalf("Starting test container has failed %s", err)
	}
	for _, scriptLine := range testRunner.Setup {
		err := testRunner.Runner.RunExecutable("docker", "exec", testRunner.ContainerName, "/bin/bash", "-c", scriptLine)
		if err != nil {
			t.Fatalf("Running setup script in test container has failed %s", err)
		}
	}

	err = testRunner.Runner.RunExecutable("docker", "cp", "piper-command-wrapper.sh", testRunner.ContainerName+":/piper-wrapper")
	if err != nil {
		t.Fatalf("Copying command wrapper to container has failed %s", err)
	}
	err = testRunner.Runner.RunExecutable("docker", "exec", testRunner.ContainerName, "chmod", "+x", "/piper-wrapper")
	if err != nil {
		t.Fatalf("Making command wrapper in container executable has failed %s", err)
	}
	return testRunner
}

func (d *IntegrationTestDockerExecRunner) whenRunningPiperCommand(command string, parameters ...string) error {
	args := []string{"exec", "--workdir", "/project", d.ContainerName, "/bin/bash", "/piper-wrapper", "/piper", command}
	args = append(args, parameters...)
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
