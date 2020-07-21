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
		t.Fatalf("Calling piper command filed %s", err)
	}

	container.assertHasOutput(t, "BUILD SUCCESS")
	container.assertHasFile(t, "/project/application/target/cloud-sdk-spring-archetype-application.jar")
	container.assertHasFile(t, "/tmp/.m2/repository")
}
