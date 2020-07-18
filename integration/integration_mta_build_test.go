// +build integration
// can be execute with go test -tags=integration ./integration/...

package main

import (
	"testing"
)

func TestMavenProject(t *testing.T) {
	t.Parallel()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "maven:3-openjdk-8-slim",
		User:    "root",
		TestDir: []string{"testdata", "TestMtaIntegration", "maven"},
		Mounts:  map[string]string{},
		Setup: []string{
			"apt-get -yqq update; apt-get -yqq install make",
			"curl -OL https://github.com/SAP/cloud-mta-build-tool/releases/download/v1.0.14/cloud-mta-build-tool_1.0.14_Linux_amd64.tar.gz",
			"tar xzf cloud-mta-build-tool_1.0.14_Linux_amd64.tar.gz",
			"mv mbt /usr/bin",
			"curl -sL https://deb.nodesource.com/setup_12.x | bash -",
			"apt-get install -yqq nodejs",
		},
	})

	err := container.whenRunningPiperCommand("mtaBuild", "--installArtifacts", "--m2Path=mym2")
	if err != nil {
		t.Fatalf("Piper command failed %s", err)
	}

	container.assertHasOutput(t, "Installing /project/.flattened-pom.xml to /project/mym2/mygroup/mymvn/1.0-SNAPSHOT/mymvn-1.0-SNAPSHOT.pom")
	container.assertHasOutput(t, "Installing /project/app/target/mymvn-app-1.0-SNAPSHOT.war to /project/mym2/mygroup/mymvn-app/1.0-SNAPSHOT/mymvn-app-1.0-SNAPSHOT.war")
	container.assertHasOutput(t, "Installing /project/app/target/mymvn-app-1.0-SNAPSHOT-classes.jar to /project/mym2/mygroup/mymvn-app/1.0-SNAPSHOT/mymvn-app-1.0-SNAPSHOT-classes.jar")
	container.assertHasOutput(t, "added 2 packages from 3 contributors and audited 2 packages in")
}

func TestNPMProject(t *testing.T) {
	t.Parallel()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "node:12",
		User:    "root",
		TestDir: []string{"testdata", "TestMtaIntegration", "npm"},
		Mounts:  map[string]string{},
		Setup: []string{
			"apt-get -yqq update; apt-get -yqq install make",
			"curl -OL https://github.com/SAP/cloud-mta-build-tool/releases/download/v1.0.14/cloud-mta-build-tool_1.0.14_Linux_amd64.tar.gz",
			"tar xzf cloud-mta-build-tool_1.0.14_Linux_amd64.tar.gz",
			"mv mbt /usr/bin",
		},
	})

	err := container.whenRunningPiperCommand("mtaBuild", "")
	if err != nil {
		t.Fatalf("Piper command failed %s", err)
	}

	container.assertHasOutput(t, "INFO the MTA archive generated at: test-mta-js.mtar")
}

func TestNPMProjectInstallsDevDependencies(t *testing.T) {
	t.Parallel()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "node:12",
		User:    "root",
		TestDir: []string{"testdata", "TestMtaIntegration", "npm-install-dev-dependencies"},
		Mounts:  map[string]string{},
		Setup: []string{
			"apt-get -yqq update; apt-get -yqq install make",
			"curl -OL https://github.com/SAP/cloud-mta-build-tool/releases/download/v1.0.14/cloud-mta-build-tool_1.0.14_Linux_amd64.tar.gz",
			"tar xzf cloud-mta-build-tool_1.0.14_Linux_amd64.tar.gz",
			"mv mbt /usr/bin",
		},
	})

	err := container.whenRunningPiperCommand("mtaBuild", "--installArtifacts")
	if err != nil {
		t.Fatalf("Piper command failed %s", err)
	}

	container.assertHasOutput(t, "added 2 packages in")
}
