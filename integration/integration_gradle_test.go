//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestGradleIntegration ./integration/...

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const DOCKER_IMAGE_GRADLE = "gradle:6-jdk11-alpine"

func TestGradleIntegrationExecuteBuildJavaProjectBOMCreationUsingWrapper(t *testing.T) {
	t.Parallel()

	container := StartPiperContainer(t, ContainerConfig{
		Image:    DOCKER_IMAGE_GRADLE,
		TestData: "TestGradleIntegration/java-project",
		WorkDir:  "/java-project",
	})

	output := RunPiper(t, container, "/java-project", "gradleExecuteBuild")

	assert.Contains(t, output, "info  gradleExecuteBuild - running command: ./gradlew tasks")
	assert.Contains(t, output, "info  gradleExecuteBuild - running command: ./gradlew cyclonedxBom --init-script initScript.gradle.tmp")
	assert.Contains(t, output, "info  gradleExecuteBuild - running command: ./gradlew build")
	assert.Contains(t, output, "info  gradleExecuteBuild - BUILD SUCCESSFUL")
	assert.Contains(t, output, "info  gradleExecuteBuild - SUCCESS")

	lsOutput := ExecCommand(t, container, "/java-project", []string{"ls", "-l", "./build/reports/"})
	assert.Contains(t, lsOutput, "bom-gradle.xml")
}

func TestGradleIntegrationExecuteBuildJavaProjectWithBomPlugin(t *testing.T) {
	t.Parallel()

	container := StartPiperContainer(t, ContainerConfig{
		Image:    DOCKER_IMAGE_GRADLE,
		TestData: "TestGradleIntegration/java-project-with-bom-plugin",
		WorkDir:  "/java-project-with-bom-plugin",
	})

	output := RunPiper(t, container, "/java-project-with-bom-plugin", "gradleExecuteBuild")

	assert.Contains(t, output, "info  gradleExecuteBuild - running command: gradle tasks")
	assert.Contains(t, output, "info  gradleExecuteBuild - running command: gradle cyclonedxBom")
	assert.Contains(t, output, "info  gradleExecuteBuild - running command: gradle build")
	assert.Contains(t, output, "info  gradleExecuteBuild - BUILD SUCCESSFUL")
	assert.Contains(t, output, "info  gradleExecuteBuild - SUCCESS")

	lsOutput := ExecCommand(t, container, "/java-project-with-bom-plugin", []string{"ls", "-l", "./build/reports/"})
	assert.Contains(t, lsOutput, "bom-gradle.xml")
}

func TestGradleIntegrationExecuteBuildWithBOMValidation(t *testing.T) {
	t.Parallel()

	container := StartPiperContainer(t, ContainerConfig{
		Image:    DOCKER_IMAGE_GRADLE,
		TestData: "TestGradleIntegration/java-project",
		WorkDir:  "/java-project",
	})

	output := RunPiper(t, container, "/java-project", "gradleExecuteBuild")
	assert.Contains(t, output, "info  gradleExecuteBuild - SUCCESS")

	output = RunPiper(t, container, "/java-project", "validateBOM")
	assert.Contains(t, output, "info  validateBOM - Found 1 BOM file(s) to validate")
	assert.Contains(t, output, "info  validateBOM - Validating BOM file:")
	assert.Contains(t, output, "bom-gradle.xml")
	assert.Contains(t, output, "info  validateBOM - BOM validation passed:")
	assert.Contains(t, output, "info  validateBOM - BOM PURL:")
	assert.Contains(t, output, "info  validateBOM - BOM validation complete: 1/1 files validated successfully")
}
