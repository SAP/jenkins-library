//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestGradleIntegration ./integration/...

package main

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/piperutils"
)

const DOCKER_IMAGE_GRADLE = "gradle:6-jdk11-alpine"

func TestGradleIntegrationExecuteBuildJavaProjectBOMCreationUsingWrapper(t *testing.T) {
	t.Parallel()
	assert := NewContainerAssert(t)

	container := StartPiperContainer(t, ContainerConfig{
		Image:    DOCKER_IMAGE_GRADLE,
		TestData: "TestGradleIntegration/java-project",
		WorkDir:  "/java-project",
	})

	output := RunPiper(t, container, "/java-project", "gradleExecuteBuild")

	assert.Contains(output, "info  gradleExecuteBuild - running command: ./gradlew tasks")
	assert.Contains(output, "info  gradleExecuteBuild - running command: ./gradlew cyclonedxBom --init-script initScript.gradle.tmp")
	assert.Contains(output, "info  gradleExecuteBuild - running command: ./gradlew build")
	assert.Contains(output, "info  gradleExecuteBuild - BUILD SUCCESSFUL")
	assert.Contains(output, "info  gradleExecuteBuild - SUCCESS")

	assert.FileExists(container, "/java-project/build/reports/bom-gradle.xml")
}

func TestGradleIntegrationExecuteBuildJavaProjectWithBomPlugin(t *testing.T) {
	t.Parallel()
	assert := NewContainerAssert(t)

	container := StartPiperContainer(t, ContainerConfig{
		Image:    DOCKER_IMAGE_GRADLE,
		TestData: "TestGradleIntegration/java-project-with-bom-plugin",
		WorkDir:  "/java-project-with-bom-plugin",
	})

	output := RunPiper(t, container, "/java-project-with-bom-plugin", "gradleExecuteBuild")

	assert.Contains(output, "info  gradleExecuteBuild - running command: gradle tasks")
	assert.Contains(output, "info  gradleExecuteBuild - running command: gradle cyclonedxBom")
	assert.Contains(output, "info  gradleExecuteBuild - running command: gradle build")
	assert.Contains(output, "info  gradleExecuteBuild - BUILD SUCCESSFUL")
	assert.Contains(output, "info  gradleExecuteBuild - SUCCESS")

	assert.FileExists(container, "/java-project-with-bom-plugin/build/reports/bom-gradle.xml")
}

func TestGradleIntegrationExecuteBuildWithBOMValidation(t *testing.T) {
	t.Parallel()
	assert := NewContainerAssert(t)

	container := StartPiperContainer(t, ContainerConfig{
		Image:    DOCKER_IMAGE_GRADLE,
		TestData: "TestGradleIntegration/java-project",
		WorkDir:  "/java-project",
	})

	output := RunPiper(t, container, "/java-project", "gradleExecuteBuild")
	assert.Contains(output, "info  gradleExecuteBuild - SUCCESS")

	// Read BOM content and validate
	bomContent := ReadFile(t, container, "/java-project/build/reports/bom-gradle.xml")
	err := piperutils.ValidateBOM(bomContent)
	assert.NoError(err, "BOM validation should pass for Gradle project")
}
