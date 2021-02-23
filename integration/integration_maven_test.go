// +build integration
// can be execute with go test -tags=integration ./integration/...

package main

import (
	"testing"
)

func TestMavenBuildCloudSdkSpringProject(t *testing.T) {
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "maven:3-openjdk-8-slim",
		User:    "1000",
		TestDir: []string{"testdata", "TestMavenIntegration", "cloud-sdk-spring-archetype"},
		Mounts:  map[string]string{},
		Setup:   []string{},
	})

	err := container.whenRunningPiperCommand("mavenBuild", "")
	if err != nil {
		t.Fatalf("Calling piper command failed %s", err)
	}

	container.assertHasOutput(t, "BUILD SUCCESS")
	container.assertHasFile(t, "/project/application/target/cloud-sdk-spring-archetype-application.jar")
	container.assertHasFile(t, "/tmp/.m2/repository")

	err = container.whenRunningPiperCommand("mavenExecuteIntegration", "")
	if err != nil {
		t.Fatalf("Calling piper command failed %s", err)
	}

	container.assertHasOutput(t, "INFO mydemo.HelloWorldControllerTest - Starting HelloWorldControllerTest")
	container.assertHasOutput(t, "Tests run: 1, Failures: 0, Errors: 0, Skipped: 0")

	container.assertHasFile(t, "/project/integration-tests/target/coverage-reports/jacoco.exec")
}

func TestMavenBuildCloudSdkTomeeProject(t *testing.T) {
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "maven:3-openjdk-8-slim",
		User:    "1000",
		TestDir: []string{"testdata", "TestMavenIntegration", "cloud-sdk-tomee-archetype"},
		Mounts:  map[string]string{},
		Setup:   []string{},
	})

	err := container.whenRunningPiperCommand("mavenBuild", "")
	if err != nil {
		t.Fatalf("Calling piper command failed %s", err)
	}

	container.assertHasOutput(t, "BUILD SUCCESS")
	container.assertHasFile(t, "/project/application/target/cloud-sdk-tomee-archetype-application-classes.jar")
	container.assertHasFile(t, "/project/application/target/cloud-sdk-tomee-archetype-application.war")
	container.assertHasFile(t, "/tmp/.m2/repository")

	err = container.whenRunningPiperCommand("mavenExecuteIntegration", "")
	if err != nil {
		t.Fatalf("Calling piper command failed %s", err)
	}

	container.assertHasOutput(t, "(prepare-agent) @ cloud-sdk-tomee-archetype-integration-tests")
	container.assertHasOutput(t, "Tests run: 1, Failures: 0, Errors: 0, Skipped: 0")

	container.assertHasFile(t, "/project/integration-tests/target/coverage-reports/jacoco.exec")
}
