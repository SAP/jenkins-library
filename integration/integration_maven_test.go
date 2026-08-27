//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestMavenIntegration ./integration/...

package main

import (
	"strings"
	"testing"

	"github.com/SAP/jenkins-library/pkg/piperutils"
)

const DOCKER_IMAGE_MAVEN = "maven:3-openjdk-8-slim"

func TestMavenIntegrationBuildCloudSdkSpringProject(t *testing.T) {
	t.Parallel()
	assert := NewContainerAssert(t)

	container := StartPiperContainer(t, ContainerConfig{
		Image:    DOCKER_IMAGE_MAVEN,
		TestData: "TestMavenIntegration/cloud-sdk-spring-archetype",
		WorkDir:  "/cloud-sdk-spring-archetype",
	})

	output := RunPiper(t, container, "/cloud-sdk-spring-archetype", "mavenBuild")
	assert.Contains(output, "BUILD SUCCESS")

	assert.FileExists(container,
		"/cloud-sdk-spring-archetype/application/target/cloud-sdk-spring-archetype-application.jar",
		"/tmp/.m2/repository",
	)

	output = RunPiper(t, container, "/cloud-sdk-spring-archetype", "mavenExecuteIntegration")
	assert.Contains(output, "INFO mydemo.HelloWorldControllerTest - Starting HelloWorldControllerTest")
	assert.Contains(output, "Tests run: 1, Failures: 0, Errors: 0, Skipped: 0")

	assert.FileExists(container, "/cloud-sdk-spring-archetype/integration-tests/target/coverage-reports/jacoco.exec")
}

func TestMavenIntegrationBuildCloudSdkTomeeProject(t *testing.T) {
	t.Parallel()
	assert := NewContainerAssert(t)

	container := StartPiperContainer(t, ContainerConfig{
		Image:    DOCKER_IMAGE_MAVEN,
		TestData: "TestMavenIntegration/cloud-sdk-tomee-archetype",
		WorkDir:  "/cloud-sdk-tomee-archetype",
	})

	output := RunPiper(t, container, "/cloud-sdk-tomee-archetype", "mavenBuild")
	assert.Contains(output, "BUILD SUCCESS")

	assert.FileExists(container,
		"/cloud-sdk-tomee-archetype/application/target/cloud-sdk-tomee-archetype-application-classes.jar",
		"/cloud-sdk-tomee-archetype/application/target/cloud-sdk-tomee-archetype-application.war",
		"/tmp/.m2/repository",
	)

	output = RunPiper(t, container, "/cloud-sdk-tomee-archetype", "mavenExecuteIntegration")
	assert.Contains(output, "(prepare-agent) @ cloud-sdk-tomee-archetype-integration-tests")
	assert.Contains(output, "Tests run: 1, Failures: 0, Errors: 0, Skipped: 0")

	assert.FileExists(container, "/cloud-sdk-tomee-archetype/integration-tests/target/coverage-reports/jacoco.exec")
}

func TestMavenIntegrationBuildWithBOMValidation(t *testing.T) {
	t.Parallel()
	assert := NewContainerAssert(t)

	container := StartPiperContainer(t, ContainerConfig{
		Image:    DOCKER_IMAGE_MAVEN,
		TestData: "TestMavenIntegration/cloud-sdk-spring-archetype",
		WorkDir:  "/cloud-sdk-spring-archetype",
	})

	output := RunPiper(t, container, "/cloud-sdk-spring-archetype", "mavenBuild")
	assert.Contains(output, "BUILD SUCCESS")

	assert.FileExists(container, "/cloud-sdk-spring-archetype/target/bom-maven.xml")

	// Read BOM content and validate
	bomContent := ReadFile(t, container, "/cloud-sdk-spring-archetype/target/bom-maven.xml")
	err := piperutils.ValidateBOM(bomContent)
	assert.NoError(err, "BOM validation should pass for Maven project")
}

// DOCKER_IMAGE_MAVEN_JDK8 is the image which piper uses by default for maven steps.
const DOCKER_IMAGE_MAVEN_JDK8 = "maven:3.8.6-jdk-8"

// TestMavenIntegrationArtifactPrepareVersionCiFriendly makes sure that artifactPrepareVersion
// reads and writes the version of a multi module project which uses ci-friendly versions
// (`${revision}`) with as few maven calls as possible. Every maven call starts a JVM and builds
// the project model, which dominates the runtime of the step on build agents with a cold maven
// repository.
func TestMavenIntegrationArtifactPrepareVersionCiFriendly(t *testing.T) {
	t.Parallel()
	assert := NewContainerAssert(t)

	const workDir = "/multi-module-ci-friendly"

	container := StartPiperContainer(t, ContainerConfig{
		Image:    DOCKER_IMAGE_MAVEN_JDK8,
		TestData: "TestMavenIntegration/multi-module-ci-friendly",
		WorkDir:  workDir,
	})

	// artifactPrepareVersion requires a git repository with at least one commit
	ExecCommand(t, container, workDir, []string{"sh", "-c",
		"git init -q . && git config user.email piper@example.com && git config user.name piper && git add -A && git commit -q -m 'initial commit'"})

	output := RunPiper(t, container, workDir, "artifactPrepareVersion",
		"--buildTool", "maven", "--versioningType", "cloud_noTag", "--fetchCoordinates=true")

	assert.Contains(output, "Version before automatic versioning: 1.0.0-SNAPSHOT")

	// the coordinates are published to the common pipeline environment, unchanged behavior
	assert.Equal("com.sap.piper.test", strings.TrimSpace(string(ReadFile(t, container, workDir+"/.pipeline/commonPipelineEnvironment/groupId"))))
	assert.Equal("multi-module-ci-friendly", strings.TrimSpace(string(ReadFile(t, container, workDir+"/.pipeline/commonPipelineEnvironment/artifactId"))))
	assert.Equal("pom", strings.TrimSpace(string(ReadFile(t, container, workDir+"/.pipeline/commonPipelineEnvironment/packaging"))))
	assert.Equal("1.0.0-SNAPSHOT", strings.TrimSpace(string(ReadFile(t, container, workDir+"/.pipeline/commonPipelineEnvironment/originalArtifactVersion"))))

	newVersion := strings.TrimSpace(string(ReadFile(t, container, workDir+"/.pipeline/commonPipelineEnvironment/artifactVersion")))
	assert.True(strings.HasPrefix(newVersion, "1.0.0-SNAPSHOT-"), "unexpected version '%s'", newVersion)

	// the new version has been written to all descriptors of the reactor
	for _, pomPath := range []string{workDir + "/pom.xml", workDir + "/application/pom.xml", workDir + "/library/pom.xml"} {
		assert.Contains(string(ReadFile(t, container, pomPath)), newVersion, "version not updated in %s", pomPath)
	}

	// regression guard for the runtime of the step: reading the version, reading the coordinates
	// and setting the version must not result in one maven call per property
	assert.Equal(2, strings.Count(output, "running command: mvn"),
		"expected one evaluation and one versions:set call, got:\n%s", output)
}

// TestMavenIntegrationUnknownContainerUser makes sure that maven uses a usable local repository if
// the container is started with a user which the image does not know - which is what the piper
// GitHub action does with `docker run --user 1000:1000`. In that case the JVM cannot resolve
// ${user.home} and maven would otherwise download the artifacts into a directory named '?' within
// the checked out repository.
func TestMavenIntegrationUnknownContainerUser(t *testing.T) {
	t.Parallel()
	assert := NewContainerAssert(t)

	const workDir = "/tmp/project"

	container := StartPiperContainer(t, ContainerConfig{
		Image:    DOCKER_IMAGE_MAVEN_JDK8,
		User:     "1000:1000",
		TestData: "TestMavenIntegration/multi-module-ci-friendly",
		WorkDir:  "/multi-module-ci-friendly",
	})

	// the test data belongs to root, therefore work on a copy which the container user can write
	ExecCommand(t, container, "/", []string{"sh", "-c",
		"cp -r /multi-module-ci-friendly " + workDir + " && cd " + workDir +
			" && git init -q . && git config user.email piper@example.com && git config user.name piper && git add -A && git commit -q -m 'initial commit'"})

	output := RunPiper(t, container, workDir, "artifactPrepareVersion", "--buildTool", "maven", "--versioningType", "library")
	assert.Contains(output, "Version before automatic versioning: 1.0.0-SNAPSHOT")

	// maven must not create a local repository within the workspace
	entries := ExecCommand(t, container, workDir, []string{"ls", "-a"})
	assert.NotContains(entries, "?", "maven created a local repository named '?' in the workspace")
}
