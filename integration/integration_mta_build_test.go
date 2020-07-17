// +build integration
// can be execute with go test -tags=integration ./integration/...

package main

import (
	"context"
	"github.com/SAP/jenkins-library/pkg/command"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
)

func TestMavenProject(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getting current working directory failed: %v", err)
	}
	pwd = filepath.Dir(pwd)

	// using custom createTmpDir function to avoid issues with symlinks on Docker for Mac
	tempDir, err := createTmpDir("")
	defer os.RemoveAll(tempDir) // clean up

	if err != nil {
		t.Fatalf("Error when creating temp dir: %v", err)
	}

	err = copyDir(filepath.Join(pwd, "integration", "testdata", "TestMtaIntegration", "maven"), tempDir)
	if err != nil {
		t.Fatal("Failed to copy test project.")
	}

	//workaround to use test script util it is possible to set workdir for Exec call
	testScript := `#!/bin/sh
cd /test
apt-get -yqq update; apt-get -yqq install make
curl -OL https://github.com/SAP/cloud-mta-build-tool/releases/download/v1.0.14/cloud-mta-build-tool_1.0.14_Linux_amd64.tar.gz
tar xzf cloud-mta-build-tool_1.0.14_Linux_amd64.tar.gz
curl -sL https://deb.nodesource.com/setup_12.x | bash -
apt-get install -yqq nodejs
mv mbt /usr/bin
mkdir mym2
/piperbin/piper mtaBuild --installArtifacts --m2Path=mym2 >test-log.txt 2>&1
`
	ioutil.WriteFile(filepath.Join(tempDir, "runPiper.sh"), []byte(testScript), 0700)

	reqNode := testcontainers.ContainerRequest{
		Image: "maven:3-openjdk-8-slim",
		Cmd:   []string{"tail", "-f"},

		BindMounts: map[string]string{
			pwd:     "/piperbin",
			tempDir: "/test",
		},
	}

	mbtContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: reqNode,
		Started:          true,
	})

	code, err := mbtContainer.Exec(ctx, []string{"sh", "/test/runPiper.sh"})

	if err != nil {
		t.Fatalf("Script returened error: %v", err)
	}
	assert.Equal(t, 0, code)

	content, err := ioutil.ReadFile(filepath.Join(tempDir, "/test-log.txt"))
	if err != nil {
		t.Fatal("Could not read test-log.txt.", err)
	}
	output := string(content)
	assert.Contains(t, output, "Installing /test/.flattened-pom.xml to /test/mym2/mygroup/mymvn/1.0-SNAPSHOT/mymvn-1.0-SNAPSHOT.pom")
	assert.Contains(t, output, "Installing /test/app/targe/mymvn-app-1.0-SNAPSHOT.war to /test/mym2/mygroup/mymvn-app/1.0-SNAPSHOT/mymvn-app-1.0-SNAPSHOT.war")
	assert.Contains(t, output, "Installing /test/app/target/mymvn-app-1.0-SNAPSHOT-classes.jar to /test/mym2/mygroup/mymvn-app/1.0-SNAPSHOT/mymvn-app-1.0-SNAPSHOT-classes.jar")
	assert.Contains(t, output, "added 2 packages from 3 contributors and audited 2 packages in")

	if t.Failed() {
		t.Fatal(output)
	}
}

func TestNPMProject(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getting current working directory failed: %v", err)
	}
	pwd = filepath.Dir(pwd)

	// using custom createTmpDir function to avoid issues with symlinks on Docker for Mac
	tempDir, err := createTmpDir("")
	defer os.RemoveAll(tempDir) // clean up

	if err != nil {
		t.Fatalf("Error when creating temp dir: %v", err)
	}

	err = copyDir(filepath.Join(pwd, "integration", "testdata", "TestMtaIntegration", "npm"), tempDir)
	if err != nil {
		t.Fatal("Failed to copy test project.")
	}

	//workaround to use test script util it is possible to set workdir for Exec call
	testScript := `#!/bin/sh
cd /test
apt-get -yqq update; apt-get -yqq install make
curl -OL https://github.com/SAP/cloud-mta-build-tool/releases/download/v1.0.14/cloud-mta-build-tool_1.0.14_Linux_amd64.tar.gz
tar xzf cloud-mta-build-tool_1.0.14_Linux_amd64.tar.gz
mv mbt /usr/bin
/piperbin/piper mtaBuild >test-log.txt 2>&1
`
	ioutil.WriteFile(filepath.Join(tempDir, "runPiper.sh"), []byte(testScript), 0700)

	reqNode := testcontainers.ContainerRequest{
		Image: "node:12",
		Cmd:   []string{"tail", "-f"},

		BindMounts: map[string]string{
			pwd:     "/piperbin",
			tempDir: "/test",
		},
	}

	mbtContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: reqNode,
		Started:          true,
	})

	code, err := mbtContainer.Exec(ctx, []string{"sh", "/test/runPiper.sh"})

	if err != nil {
		t.Fatalf("Script returened error: %v", err)
	}
	assert.Equal(t, 0, code)

	content, err := ioutil.ReadFile(filepath.Join(tempDir, "/test-log.txt"))
	if err != nil {
		t.Fatal("Could not read test-log.txt.", err)
	}
	output := string(content)
	assert.Contains(t, output, "INFO the MTA archive gendfserated at: test-mta-js.mtar")
}

func TestNPMProjectInstallsDevDependencies(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getting current working directory failed: %v", err)
	}
	pwd = filepath.Dir(pwd)

	// using custom createTmpDir function to avoid issues with symlinks on Docker for Mac
	tempDir, err := createTmpDir("")
	defer os.RemoveAll(tempDir) // clean up

	if err != nil {
		t.Fatalf("Error when creating temp dir: %v", err)
	}

	err = copyDir(filepath.Join(pwd, "integration", "testdata", "TestMtaIntegration", "npm-install-dev-dependencies"), tempDir)
	if err != nil {
		t.Fatal("Failed to copy test project.")
	}

	//workaround to use test script util it is possible to set workdir for Exec call
	testScript := `#!/bin/sh
cd /test
apt-get -yqq update; apt-get -yqq install make
curl -OL https://github.com/SAP/cloud-mta-build-tool/releases/download/v1.0.14/cloud-mta-build-tool_1.0.14_Linux_amd64.tar.gz
tar xzf cloud-mta-build-tool_1.0.14_Linux_amd64.tar.gz
mv mbt /usr/bin
/piperbin/piper mtaBuild --installArtifacts >test-log.txt 2>&1
`
	ioutil.WriteFile(filepath.Join(tempDir, "runPiper.sh"), []byte(testScript), 0700)

	reqNode := testcontainers.ContainerRequest{
		Image: "node:12",
		Cmd:   []string{"tail", "-f"},

		BindMounts: map[string]string{
			pwd:     "/piperbin",
			tempDir: "/test",
		},
	}

	mbtContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: reqNode,
		Started:          true,
	})

	code, err := mbtContainer.Exec(ctx, []string{"sh", "/test/runPiper.sh"})

	if err != nil {
		t.Fatalf("Script returened error: %v", err)
	}
	assert.Equal(t, 0, code)

	content, err := ioutil.ReadFile(filepath.Join(tempDir, "/test-log.txt"))
	if err != nil {
		t.Fatal("Could not read test-log.txt.", err)
	}
	output := string(content)
	assert.Contains(t, output, "added 2 packages in")
}

func TestWithNativeDockerClient(t *testing.T) {
	//t.Parallel()
	testRunner := given(IntegrationTestDockerExecRunner{
		Image:  "node:12",
		User:   "root",
		Mounts: map[string]string{},
		Setup: []string{
			"apt-get -yqq update; apt-get -yqq install make",
			"curl -OL https://github.com/SAP/cloud-mta-build-tool/releases/download/v1.0.14/cloud-mta-build-tool_1.0.14_Linux_amd64.tar.gz",
			"tar xzf cloud-mta-build-tool_1.0.14_Linux_amd64.tar.gz",
			"mv mbt /usr/bin",
		},
	})

	testRunner.whenRunningPiperCommand("mtaBuild", "--installArtifacts")

	testRunner.assertHasOutput("")

}

func given(foo IntegrationTestDockerExecRunner) IntegrationTestDockerExecRunner {
	runner := command.Command{}

	testRunner := IntegrationTestDockerExecRunner{
		Runner: runner,
		Image:  foo.Image,
		User:   foo.User,
		Mounts: foo.Mounts,
		Setup:  foo.Setup,
	}

	//todo ensure it is a linux binary
	wd, _ := os.Getwd()
	localPiper := path.Join(wd, "..", "piper")

	projectDir := path.Join(wd, "testdata", "TestMtaIntegration", "npm-install-dev-dependencies")

	//todo mounts
	//todo random name gen
	err := testRunner.Runner.RunExecutable("docker", "run", "-d", "-u=" + testRunner.User, "-v", localPiper + ":/piper", "-v", projectDir + ":/project", "--name=foobar", testRunner.Image, "sleep", "2000")
	if err != nil {
		panic(err)
	}
	for _, scriptLine := range testRunner.Setup {
		err := testRunner.Runner.RunExecutable("docker", "exec", "foobar", "/bin/bash", "-c", scriptLine)
		if err != nil {
			panic(err)
		}
	}
	return testRunner
}


type IntegrationTestDockerExecRunner struct {
	// Runner is the ExecRunner to which all executions are forwarded in the end.
	Runner command.Command
	Image  string
	User   string
	Mounts map[string]string
	Setup  []string
}

func (d *IntegrationTestDockerExecRunner) whenRunningPiperCommand(command string, parameters ...string) error {
	args := []string{"exec", "--workdir", "/project", "foobar", "/piper", command}
	args = append(args, parameters...)
	return d.Runner.RunExecutable("docker", args...)
}

func (d *IntegrationTestDockerExecRunner) assertHasOutput(want string) {
	d.Runner.RunExecutable("docker", "logs", "foobar")
}
