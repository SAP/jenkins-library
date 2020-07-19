// +build integration

package main

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"math/rand"
	"os"
	"path"
	"strconv"
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

	//todo mounts
	//todo env (secrets)
	err := testRunner.Runner.RunExecutable("docker", "run", "-d", "-u="+testRunner.User,
		"-v", localPiper+":/piper", "-v", projectDir+":/project",
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
	err = testRunner.Runner.RunExecutable("docker", "exec", "-u=root", testRunner.ContainerName, "chmod", "+x", "/piper-wrapper")
	if err != nil {
		t.Fatalf("Making command wrapper in container execuable has failed %s", err)
	}
	return testRunner
}

func (d *IntegrationTestDockerExecRunner) whenRunningPiperCommand(t *testing.T, command string, parameters ...string) {
	args := []string{"exec", "--workdir", "/project", d.ContainerName, "/bin/bash", "/piper-wrapper", "/piper", command}
	args = append(args, parameters...)
	err := d.Runner.RunExecutable("docker", args...)
	if err != nil {
		println("dbg>>>>>")
		_ = d.Runner.RunExecutable("docker", "exec", d.ContainerName, "cat", "/tmp/test-log.txt")
		t.Fatalf("Running piper command failed, error: %s", err)
	}
}

func (d *IntegrationTestDockerExecRunner) assertHasOutput(t *testing.T, want string) {
	//todo depends on bash for now. I did not find a way to make it work with RunExecutable so far.
	err := d.Runner.RunShell("/bin/bash", fmt.Sprintf("docker exec %s grep --count '%s' /tmp/test-log.txt", d.ContainerName, want))
	if err != nil {
		_ = d.Runner.RunExecutable("docker", "exec", d.ContainerName, "cat", "/tmp/test-log.txt")
		t.Fatalf("Assertion has failed. Expected output %s in command output. %s", want, err)
	}
}

func (d *IntegrationTestDockerExecRunner) assertHasFile(t *testing.T, want string) {
	//todo depends on bash for now. I did not find a way to make it work with RunExecutable so far.
	err := d.Runner.RunShell("/bin/bash", fmt.Sprintf("docker exec %s stat '%s' /tmp/test-log.txt", d.ContainerName, want))
	if err != nil {
		t.Fatalf("Assertion has failed. Expected file to exist %s in container. %s", want, err)
	}
}
