// +build integration
// can be execute with go test -tags=integration ./integration/...

package main

import (
	"testing"
)

func TestMavenBuildCloudSdkSpringProject(t *testing.T) {
	t.Parallel()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "maven:3-openjdk-8-slim",
		User:    "root",
		TestDir: []string{"testdata", "TestMavenIntegration", "cloud-sdk-spring-archetype"},
		Mounts:  map[string]string{},
		Setup:   []string{},
	})

	container.whenRunningPiperCommand(t, "mavenBuild", "")

	container.assertHasOutput(t, "BUILD SUCCESS")
	container.assertHasFile(t, "/project/application/target/cloud-sdk-spring-archetype-application.jar")
	container.assertHasFile(t, "/tmp/.m2/repository")
}
