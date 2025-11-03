//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestPythonIntegration ./integration/...

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

func TestPythonIntegrationBuildProject(t *testing.T) {
	// t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	reqNode := testcontainers.ContainerRequest{
		Image: "python:3.10",
		Cmd:   []string{"tail", "-f"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      filepath.Join(pwd, "integration", "testdata", "TestPythonIntegration", "python-project"),
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

	pythonContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: reqNode,
		Started:          true,
	})
	require.NoError(t, err)

	code, reader, err := pythonContainer.Exec(ctx, []string{"/piperbin/piper", "pythonBuild"}, exec.WithWorkingDir("/python-project"))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	outputBytes, err := io.ReadAll(reader)
	assert.NoError(t, err)
	output := string(outputBytes)
	assert.Contains(t, output, "info  pythonBuild - running command: piperBuild-env/bin/python setup.py sdist bdist_wheel")
	assert.Contains(t, output, "info  pythonBuild - running command: piperBuild-env/bin/pip install --upgrade --root-user-action=ignore cyclonedx-bom==")
	assert.Contains(t, output, "info  pythonBuild - running command: piperBuild-env/bin/cyclonedx-py env --output-file bom-pip.xml --output-format XML --spec-version 1.4")
	assert.Contains(t, output, "info  pythonBuild - SUCCESS")

	code, reader, err = pythonContainer.Exec(ctx, []string{"ls", "-l", ".", "dist", "build"}, exec.WithWorkingDir("/python-project"))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	lsOutputBytes, err := io.ReadAll(reader)
	assert.NoError(t, err)
	lsOutput := string(lsOutputBytes)
	assert.Contains(t, lsOutput, "example_pkg-0.0.1.tar.gz")
	assert.Contains(t, lsOutput, "example_pkg-0.0.1-py3-none-any.whl")
}

func TestPythonIntegrationBuildWithBOMValidation(t *testing.T) {
	// t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	reqNode := testcontainers.ContainerRequest{
		Image: "python:3.10",
		Cmd:   []string{"tail", "-f"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      filepath.Join(pwd, "integration", "testdata", "TestPythonIntegration", "python-project"),
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

	pythonContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: reqNode,
		Started:          true,
	})
	require.NoError(t, err)

	// First, run pythonBuild to generate the BOM
	code, reader, err := pythonContainer.Exec(ctx, []string{"/piperbin/piper", "pythonBuild"}, exec.WithWorkingDir("/python-project"))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	outputBytes, err := io.ReadAll(reader)
	assert.NoError(t, err)
	output := string(outputBytes)
	assert.Contains(t, output, "info  pythonBuild - running command: piperBuild-env/bin/cyclonedx-py env --output-file bom-pip.xml --output-format XML --spec-version 1.4")
	assert.Contains(t, output, "info  pythonBuild - SUCCESS")

	t.Log(output)

	// Now run validateBOM on the generated BOM
	code, reader, err = pythonContainer.Exec(ctx, []string{"/piperbin/piper", "validateBOM"}, exec.WithWorkingDir("/python-project"))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	outputBytes, err = io.ReadAll(reader)
	assert.NoError(t, err)
	output = string(outputBytes)

	t.Log(output)
	assert.Contains(t, output, "info  validateBOM - Found 1 BOM file(s) to validate")
	assert.Contains(t, output, "info  validateBOM - Validating BOM file:")
	assert.Contains(t, output, "bom-pip.xml")
	assert.Contains(t, output, "warn  validateBOM - BOM validation failed for:") // cyclonedx-py currently generates incomplete BOMs
	assert.Contains(t, output, "metadata.component.name is required but missing")
	assert.Contains(t, output, "info  validateBOM - BOM validation complete:")
}
