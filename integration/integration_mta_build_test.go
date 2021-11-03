// +build integration
// can be execute with go test -tags=integration ./integration/...

package main

import (
	"testing"
)

func TestMavenProject(t *testing.T) {
	t.Parallel()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "devxci/mbtci:1.1.1",
		User:    "root",
		TestDir: []string{"testdata", "TestMtaIntegration", "maven"},
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

func TestMavenSpringProject(t *testing.T) {
	t.Parallel()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "devxci/mbtci:1.1.1",
		User:    "root",
		TestDir: []string{"testdata", "TestMtaIntegration", "maven-spring"},
	})

	err := container.whenRunningPiperCommand("mtaBuild", "--installArtifacts", "--m2Path=mym2")
	if err != nil {
		t.Fatalf("Piper command failed %s", err)
	}
	err = container.whenRunningPiperCommand("mavenExecuteIntegration", "--m2Path=mym2")
	if err != nil {
		t.Fatalf("Piper command failed %s", err)
	}

	container.assertHasOutput(t, "Tests run: 1, Failures: 0, Errors: 0, Skipped: 0")
}

func TestNPMProject(t *testing.T) {
	t.Parallel()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "devxci/mbtci:1.1.1",
		User:    "root",
		TestDir: []string{"testdata", "TestMtaIntegration", "npm"},
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
		Image:   "devxci/mbtci:1.1.1",
		User:    "root",
		TestDir: []string{"testdata", "TestMtaIntegration", "npm-install-dev-dependencies"},
	})

	err := container.whenRunningPiperCommand("mtaBuild", "--installArtifacts")
	if err != nil {
		t.Fatalf("Piper command failed %s", err)
	}

	container.assertHasOutput(t, "added 2 packages in")
}
