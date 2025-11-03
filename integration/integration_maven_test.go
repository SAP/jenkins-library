//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestMavenIntegration ./integration/...

package main

import (
	"testing"

	"github.com/SAP/jenkins-library/integration/testhelper"
	"github.com/stretchr/testify/assert"
)

const DOCKER_IMAGE_MAVEN = "maven:3-openjdk-8-slim"

func TestMavenIntegrationBuildCloudSdkSpringProject(t *testing.T) {
	t.Parallel()

	container := testhelper.StartPiperContainer(t, testhelper.ContainerConfig{
		Image:    DOCKER_IMAGE_MAVEN,
		TestData: "TestMavenIntegration/cloud-sdk-spring-archetype",
		WorkDir:  "/cloud-sdk-spring-archetype",
	})

	output := testhelper.RunPiper(t, container, "/cloud-sdk-spring-archetype", "mavenBuild")
	assert.Contains(t, output, "BUILD SUCCESS")

	testhelper.AssertFileExists(t, container,
		"/cloud-sdk-spring-archetype/application/target/cloud-sdk-spring-archetype-application.jar",
		"/tmp/.m2/repository",
	)

	output = testhelper.RunPiper(t, container, "/cloud-sdk-spring-archetype", "mavenExecuteIntegration")
	assert.Contains(t, output, "INFO mydemo.HelloWorldControllerTest - Starting HelloWorldControllerTest")
	assert.Contains(t, output, "Tests run: 1, Failures: 0, Errors: 0, Skipped: 0")

	testhelper.AssertFileExists(t, container, "/cloud-sdk-spring-archetype/integration-tests/target/coverage-reports/jacoco.exec")
}

func TestMavenIntegrationBuildCloudSdkTomeeProject(t *testing.T) {
	t.Parallel()

	container := testhelper.StartPiperContainer(t, testhelper.ContainerConfig{
		Image:    DOCKER_IMAGE_MAVEN,
		TestData: "TestMavenIntegration/cloud-sdk-tomee-archetype",
		WorkDir:  "/cloud-sdk-tomee-archetype",
	})

	output := testhelper.RunPiper(t, container, "/cloud-sdk-tomee-archetype", "mavenBuild")
	assert.Contains(t, output, "BUILD SUCCESS")

	testhelper.AssertFileExists(t, container,
		"/cloud-sdk-tomee-archetype/application/target/cloud-sdk-tomee-archetype-application-classes.jar",
		"/cloud-sdk-tomee-archetype/application/target/cloud-sdk-tomee-archetype-application.war",
		"/tmp/.m2/repository",
	)

	output = testhelper.RunPiper(t, container, "/cloud-sdk-tomee-archetype", "mavenExecuteIntegration")
	assert.Contains(t, output, "(prepare-agent) @ cloud-sdk-tomee-archetype-integration-tests")
	assert.Contains(t, output, "Tests run: 1, Failures: 0, Errors: 0, Skipped: 0")

	testhelper.AssertFileExists(t, container, "/cloud-sdk-tomee-archetype/integration-tests/target/coverage-reports/jacoco.exec")
}

func TestMavenIntegrationBuildWithBOMValidation(t *testing.T) {
	t.Parallel()

	container := testhelper.StartPiperContainer(t, testhelper.ContainerConfig{
		Image:    DOCKER_IMAGE_MAVEN,
		TestData: "TestMavenIntegration/cloud-sdk-spring-archetype",
		WorkDir:  "/cloud-sdk-spring-archetype",
	})

	output := testhelper.RunPiper(t, container, "/cloud-sdk-spring-archetype", "mavenBuild")
	assert.Contains(t, output, "BUILD SUCCESS")

	testhelper.AssertFileExists(t, container, "/cloud-sdk-spring-archetype/target/bom-maven.xml")

	output = testhelper.RunPiper(t, container, "/cloud-sdk-spring-archetype", "validateBOM")
	assert.Contains(t, output, "info  validateBOM - Found 1 BOM file(s) to validate")
	assert.Contains(t, output, "info  validateBOM - Validating BOM file:")
	assert.Contains(t, output, "bom-maven.xml")
	assert.Contains(t, output, "info  validateBOM - BOM validation passed:")
	assert.Contains(t, output, "info  validateBOM - BOM PURL:")
	assert.Contains(t, output, "info  validateBOM - BOM validation complete: 1/1 files validated successfully")
}
