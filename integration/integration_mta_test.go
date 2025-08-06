//go:build integration

// can be executed with
// go test -v -tags integration -run TestMTAIntegration ./integration/...

package main

import (
	"testing"
)

func TestMTAIntegrationMavenProject(t *testing.T) {
	// t.Parallel()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "devxci/mbtci-java11-node14",
		User:    "root",
		TestDir: []string{"testdata", "TestMtaIntegration", "maven"},
	})
	defer container.terminate(t)

	err := container.whenRunningPiperCommand("mtaBuild", "--installArtifacts", "--m2Path=mym2")
	if err != nil {
		t.Fatalf("Piper command failed %s", err)
	}

	container.assertHasOutput(t,
		"Installing /project/.flattened-pom.xml to /project/mym2/mygroup/mymvn/1.0-SNAPSHOT/mymvn-1.0-SNAPSHOT.pom",
		"Installing /project/app/target/mymvn-app-1.0-SNAPSHOT.war to /project/mym2/mygroup/mymvn-app/1.0-SNAPSHOT/mymvn-app-1.0-SNAPSHOT.war",
		"Installing /project/app/target/mymvn-app-1.0-SNAPSHOT-classes.jar to /project/mym2/mygroup/mymvn-app/1.0-SNAPSHOT/mymvn-app-1.0-SNAPSHOT-classes.jar",
		"added 2 packages from 3 contributors and audited 2 packages in",
	)
}

func TestMTAIntegrationMavenSpringProject(t *testing.T) {
	// t.Parallel()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "devxci/mbtci-java11-node14",
		User:    "root",
		TestDir: []string{"testdata", "TestMtaIntegration", "maven-spring"},
	})
	defer container.terminate(t)

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

func TestMTAIntegrationNPMProject(t *testing.T) {
	// t.Parallel()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "devxci/mbtci-java11-node14",
		User:    "root",
		TestDir: []string{"testdata", "TestMtaIntegration", "npm"},
	})
	defer container.terminate(t)

	err := container.whenRunningPiperCommand("mtaBuild", "")
	if err != nil {
		t.Fatalf("Piper command failed %s", err)
	}

	container.assertHasOutput(t, "INFO the MTA archive generated at: /project/test-mta-js.mtar")
}

func TestMTAIntegrationNPMProjectInstallsDevDependencies(t *testing.T) {
	// t.Parallel()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "devxci/mbtci-java11-node14",
		User:    "root",
		TestDir: []string{"testdata", "TestMtaIntegration", "npm-install-dev-dependencies"},
	})
	defer container.terminate(t)

	err := container.whenRunningPiperCommand("mtaBuild", "--installArtifacts")
	if err != nil {
		t.Fatalf("Piper command failed %s", err)
	}

	container.assertHasOutput(t, "added 2 packages in")
}
