//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestGradleIntegration ./integration/...

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

const DOCKER_IMAGE_GRADLE = "gradle:6-jdk11-alpine"

func TestGradleIntegrationExecuteBuildJavaProjectBOMCreationUsingWrapper(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	reqNode := testcontainers.ContainerRequest{
		Image: DOCKER_IMAGE_GRADLE,
		Cmd:   []string{"tail", "-f"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      filepath.Join(pwd, "integration", "testdata", "TestGradleIntegration", "java-project"),
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

	nodeContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: reqNode,
		Started:          true,
	})
	require.NoError(t, err)

	code, reader, err := nodeContainer.Exec(ctx, []string{"/piperbin/piper", "gradleExecuteBuild"}, exec.WithWorkingDir("/java-project"))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	outputBytes, err := io.ReadAll(reader)
	assert.NoError(t, err)
	output := string(outputBytes)
	assert.Contains(t, output, "info  gradleExecuteBuild - running command: ./gradlew tasks")
	assert.Contains(t, output, "info  gradleExecuteBuild - running command: ./gradlew cyclonedxBom --init-script initScript.gradle.tmp")
	assert.Contains(t, output, "info  gradleExecuteBuild - running command: ./gradlew build")
	assert.Contains(t, output, "info  gradleExecuteBuild - BUILD SUCCESSFUL")
	assert.Contains(t, output, "info  gradleExecuteBuild - SUCCESS")
	assert.Contains(t, output, "Validating generated SBOM:")
	assert.Contains(t, output, "SBOM validation passed")
	assert.Contains(t, output, "SBOM PURL:")

	code, reader, err = nodeContainer.Exec(ctx, []string{"ls", "-l", "./build/reports/"}, exec.WithWorkingDir("/java-project"))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	lsOutputBytes, err := io.ReadAll(reader)
	assert.NoError(t, err)
	lsOutput := string(lsOutputBytes)
	assert.Contains(t, lsOutput, "bom-gradle.xml")
}

func TestGradleIntegrationExecuteBuildJavaProjectWithBomPlugin(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	reqNode := testcontainers.ContainerRequest{
		Image: DOCKER_IMAGE_GRADLE,
		Cmd:   []string{"tail", "-f"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      filepath.Join(pwd, "integration", "testdata", "TestGradleIntegration", "java-project-with-bom-plugin"),
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

	nodeContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: reqNode,
		Started:          true,
	})
	require.NoError(t, err)

	code, reader, err := nodeContainer.Exec(ctx, []string{"/piperbin/piper", "gradleExecuteBuild"}, exec.WithWorkingDir("/java-project-with-bom-plugin"))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	outputBytes, err := io.ReadAll(reader)
	assert.NoError(t, err)
	output := string(outputBytes)
	assert.Contains(t, output, "info  gradleExecuteBuild - running command: gradle tasks")
	assert.Contains(t, output, "info  gradleExecuteBuild - running command: gradle cyclonedxBom")
	assert.Contains(t, output, "info  gradleExecuteBuild - running command: gradle build")
	assert.Contains(t, output, "info  gradleExecuteBuild - BUILD SUCCESSFUL")
	assert.Contains(t, output, "info  gradleExecuteBuild - SUCCESS")
	assert.Contains(t, output, "Validating generated SBOM:")
	assert.Contains(t, output, "SBOM validation passed")
	assert.Contains(t, output, "SBOM PURL:")

	code, reader, err = nodeContainer.Exec(ctx, []string{"ls", "-l", "./build/reports/"}, exec.WithWorkingDir("/java-project-with-bom-plugin"))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	lsOutputBytes, err := io.ReadAll(reader)
	assert.NoError(t, err)
	lsOutput := string(lsOutputBytes)
	assert.Contains(t, lsOutput, "bom-gradle.xml")
}
