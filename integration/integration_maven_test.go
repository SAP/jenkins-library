//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestMavenIntegration ./integration/...

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/exec"
)

const DOCKER_IMAGE_MAVEN = "maven:3-openjdk-8-slim"

func TestMavenIntegrationBuildCloudSdkSpringProject(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	reqNode := testcontainers.ContainerRequest{
		Image: DOCKER_IMAGE_MAVEN,
		Cmd:   []string{"tail", "-f"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      filepath.Join(pwd, "integration", "testdata", "TestMavenIntegration", "cloud-sdk-spring-archetype"),
				ContainerFilePath: "/",
				FileMode:          0755,
			},
		},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.Binds = []string{
				fmt.Sprintf("%s:/piperbin", pwd),
			}
		},
	}

	mavenContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: reqNode,
		Started:          true,
	})
	require.NoError(t, err)

	code, reader, err := mavenContainer.Exec(ctx, []string{"/piperbin/piper", "mavenBuild"}, exec.WithWorkingDir("/cloud-sdk-spring-archetype"))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	outputBytes, err := io.ReadAll(reader)
	assert.NoError(t, err)
	output := string(outputBytes)
	assert.Contains(t, output, "BUILD SUCCESS")

	code, _, err = mavenContainer.Exec(ctx, []string{"stat",
		"/cloud-sdk-spring-archetype/application/target/cloud-sdk-spring-archetype-application.jar",
		"/tmp/.m2/repository",
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	code, reader, err = mavenContainer.Exec(ctx, []string{"/piperbin/piper", "mavenExecuteIntegration"}, exec.WithWorkingDir("/cloud-sdk-spring-archetype"))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	outputBytes, err = io.ReadAll(reader)
	assert.NoError(t, err)
	output = string(outputBytes)
	assert.Contains(t, output, "INFO mydemo.HelloWorldControllerTest - Starting HelloWorldControllerTest")
	assert.Contains(t, output, "Tests run: 1, Failures: 0, Errors: 0, Skipped: 0")

	code, _, err = mavenContainer.Exec(ctx, []string{"stat", "/cloud-sdk-spring-archetype/integration-tests/target/coverage-reports/jacoco.exec"})
	assert.NoError(t, err)
	assert.Equal(t, 0, code)
}

func TestMavenIntegrationBuildCloudSdkTomeeProject(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	reqNode := testcontainers.ContainerRequest{
		Image: DOCKER_IMAGE_MAVEN,
		Cmd:   []string{"tail", "-f"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      filepath.Join(pwd, "integration", "testdata", "TestMavenIntegration", "cloud-sdk-tomee-archetype"),
				ContainerFilePath: "/",
				FileMode:          0755,
			},
		},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.Binds = []string{
				fmt.Sprintf("%s:/piperbin", pwd),
			}
		},
	}

	mavenContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: reqNode,
		Started:          true,
	})
	require.NoError(t, err)

	code, reader, err := mavenContainer.Exec(ctx, []string{"/piperbin/piper", "mavenBuild"}, exec.WithWorkingDir("/cloud-sdk-tomee-archetype"))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	outputBytes, err := io.ReadAll(reader)
	assert.NoError(t, err)
	output := string(outputBytes)
	assert.Contains(t, output, "BUILD SUCCESS")

	code, _, err = mavenContainer.Exec(ctx, []string{"stat",
		"/cloud-sdk-tomee-archetype/application/target/cloud-sdk-tomee-archetype-application-classes.jar",
		"/cloud-sdk-tomee-archetype/application/target/cloud-sdk-tomee-archetype-application.war",
		"/tmp/.m2/repository",
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	code, reader, err = mavenContainer.Exec(ctx, []string{"/piperbin/piper", "mavenExecuteIntegration"}, exec.WithWorkingDir("/cloud-sdk-tomee-archetype"))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	outputBytes, err = io.ReadAll(reader)
	assert.NoError(t, err)
	output = string(outputBytes)
	assert.Contains(t, output, "(prepare-agent) @ cloud-sdk-tomee-archetype-integration-tests")
	assert.Contains(t, output, "Tests run: 1, Failures: 0, Errors: 0, Skipped: 0")

	code, _, err = mavenContainer.Exec(ctx, []string{"stat", "/cloud-sdk-tomee-archetype/integration-tests/target/coverage-reports/jacoco.exec"})
	assert.NoError(t, err)
	assert.Equal(t, 0, code)
}

func TestMavenIntegrationBuildWithBOMValidation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	reqNode := testcontainers.ContainerRequest{
		Image: DOCKER_IMAGE_MAVEN,
		Cmd:   []string{"tail", "-f"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      filepath.Join(pwd, "integration", "testdata", "TestMavenIntegration", "cloud-sdk-spring-archetype"),
				ContainerFilePath: "/",
				FileMode:          0755,
			},
		},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.Binds = []string{
				fmt.Sprintf("%s:/piperbin", pwd),
			}
		},
	}

	mavenContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: reqNode,
		Started:          true,
	})
	require.NoError(t, err)

	code, reader, err := mavenContainer.Exec(ctx, []string{"/piperbin/piper", "mavenBuild"}, exec.WithWorkingDir("/cloud-sdk-spring-archetype"))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	outputBytes, err := io.ReadAll(reader)
	assert.NoError(t, err)
	output := string(outputBytes)
	assert.Contains(t, output, "BUILD SUCCESS")

	code, _, err = mavenContainer.Exec(ctx, []string{"stat", "/cloud-sdk-spring-archetype/target/bom-maven.xml"})
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	code, reader, err = mavenContainer.Exec(ctx, []string{"/piperbin/piper", "validateBOM"}, exec.WithWorkingDir("/cloud-sdk-spring-archetype"))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	outputBytes, err = io.ReadAll(reader)
	assert.NoError(t, err)
	output = string(outputBytes)
	assert.Contains(t, output, "info  validateBOM - Found 1 BOM file(s) to validate")
	assert.Contains(t, output, "info  validateBOM - Validating BOM file:")
	assert.Contains(t, output, "bom-maven.xml")
	assert.Contains(t, output, "info  validateBOM - BOM validation passed:")
	assert.Contains(t, output, "info  validateBOM - BOM PURL:")
	assert.Contains(t, output, "info  validateBOM - BOM validation complete: 1/1 files validated successfully")
}
