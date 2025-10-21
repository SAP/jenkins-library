//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestMavenIntegration ./integration/...

package main

import (
	"testing"
)

func TestMavenIntegrationBuildCloudSdkSpringProject(t *testing.T) {
	t.Parallel()

	// Create a shared Maven cache directory to avoid re-downloading dependencies
	mavenCache, err := createTmpDir(t)
	if err != nil {
		t.Fatalf("Failed to create Maven cache directory: %s", err)
	}

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "maven:3-openjdk-8-slim",
		User:    "1000",
		TestDir: []string{"testdata", "TestMavenIntegration", "cloud-sdk-spring-archetype"},
		Mounts:  map[string]string{mavenCache: "/tmp/.m2"},
		Setup:   []string{"chown -R 1000:1000 /tmp/.m2"},
	})
	defer container.terminate(t)

	err = container.whenRunningPiperCommand("mavenBuild", "")
	if err != nil {
		t.Fatalf("Calling piper command failed %s", err)
	}

	container.assertHasOutput(t, "BUILD SUCCESS")
	container.assertHasFiles(t,
		"/project/application/target/cloud-sdk-spring-archetype-application.jar",
		"/tmp/.m2/repository",
	)

	err = container.whenRunningPiperCommand("mavenExecuteIntegration", "")
	if err != nil {
		t.Fatalf("Calling piper command failed %s", err)
	}

	container.assertHasOutput(t,
		"INFO mydemo.HelloWorldControllerTest - Starting HelloWorldControllerTest",
		"Tests run: 1, Failures: 0, Errors: 0, Skipped: 0",
	)

	container.assertHasFiles(t, "/project/integration-tests/target/coverage-reports/jacoco.exec")
}

func TestMavenIntegrationBuildCloudSdkTomeeProject(t *testing.T) {
	t.Parallel()

	// Create a shared Maven cache directory to avoid re-downloading dependencies
	mavenCache, err := createTmpDir(t)
	if err != nil {
		t.Fatalf("Failed to create Maven cache directory: %s", err)
	}

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "maven:3-openjdk-8-slim",
		User:    "1000",
		TestDir: []string{"testdata", "TestMavenIntegration", "cloud-sdk-tomee-archetype"},
		Mounts:  map[string]string{mavenCache: "/tmp/.m2"},
		Setup:   []string{"chown -R 1000:1000 /tmp/.m2"},
	})
	defer container.terminate(t)

	err = container.whenRunningPiperCommand("mavenBuild", "")
	if err != nil {
		t.Fatalf("Calling piper command failed %s", err)
	}

	container.assertHasOutput(t, "BUILD SUCCESS")
	container.assertHasFiles(t,
		"/project/application/target/cloud-sdk-tomee-archetype-application-classes.jar",
		"/project/application/target/cloud-sdk-tomee-archetype-application.war",
		"/tmp/.m2/repository",
	)

	err = container.whenRunningPiperCommand("mavenExecuteIntegration", "")
	if err != nil {
		t.Fatalf("Calling piper command failed %s", err)
	}

	container.assertHasOutput(t,
		"(prepare-agent) @ cloud-sdk-tomee-archetype-integration-tests",
		"Tests run: 1, Failures: 0, Errors: 0, Skipped: 0",
	)

	container.assertHasFiles(t, "/project/integration-tests/target/coverage-reports/jacoco.exec")
}
