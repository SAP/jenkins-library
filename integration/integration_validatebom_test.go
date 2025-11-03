//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestValidateBOMIntegration ./integration/...

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

const DOCKER_IMAGE_VALIDATEBOM = "alpine:latest"

// TestValidateBOMIntegrationInvalidBOM tests that validateBOM logs warnings for invalid BOMs
// but does not fail the build by default
func TestValidateBOMIntegrationInvalidBOM(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	reqNode := testcontainers.ContainerRequest{
		Image: DOCKER_IMAGE_VALIDATEBOM,
		Cmd:   []string{"tail", "-f"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      filepath.Join(pwd, "integration", "testdata", "TestValidateBOMIntegration", "invalid-bom"),
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

	validateBOMContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: reqNode,
		Started:          true,
	})
	require.NoError(t, err)

	code, reader, err := validateBOMContainer.Exec(ctx, []string{"/piperbin/piper", "validateBOM"}, exec.WithWorkingDir("/invalid-bom"))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	outputBytes, err := io.ReadAll(reader)
	assert.NoError(t, err)
	output := string(outputBytes)
	assert.Contains(t, output, "info  validateBOM - Found 1 BOM file(s) to validate")
	assert.Contains(t, output, "info  validateBOM - Validating BOM file:")
	assert.Contains(t, output, "bom-invalid.xml")
	assert.Contains(t, output, "warn  validateBOM - BOM validation failed for:")
	assert.Contains(t, output, "info  validateBOM - BOM validation complete: 0/1 files validated successfully")
}

// TestValidateBOMIntegrationFailOnError tests that validateBOM fails the build when
// failOnValidationError is true and a BOM is invalid
func TestValidateBOMIntegrationFailOnError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	reqNode := testcontainers.ContainerRequest{
		Image: DOCKER_IMAGE_VALIDATEBOM,
		Cmd:   []string{"tail", "-f"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      filepath.Join(pwd, "integration", "testdata", "TestValidateBOMIntegration", "invalid-bom"),
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

	validateBOMContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: reqNode,
		Started:          true,
	})
	require.NoError(t, err)

	code, reader, err := validateBOMContainer.Exec(ctx, []string{"/piperbin/piper", "validateBOM", "--failOnValidationError"}, exec.WithWorkingDir("/invalid-bom"))
	assert.NoError(t, err)
	assert.NotEqual(t, 0, code)

	outputBytes, err := io.ReadAll(reader)
	assert.NoError(t, err)
	output := string(outputBytes)
	assert.Contains(t, output, "info  validateBOM - Found 1 BOM file(s) to validate")
	assert.Contains(t, output, "info  validateBOM - Validating BOM file:")
	assert.Contains(t, output, "bom-invalid.xml")
	assert.Contains(t, output, "fatal validateBOM")
	assert.Contains(t, output, "BOM validation failed for")
}

// TestValidateBOMIntegrationSkip tests that validateBOM can be skipped entirely
func TestValidateBOMIntegrationSkip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	reqNode := testcontainers.ContainerRequest{
		Image: DOCKER_IMAGE_VALIDATEBOM,
		Cmd:   []string{"tail", "-f"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      filepath.Join(pwd, "integration", "testdata", "TestValidateBOMIntegration", "invalid-bom"),
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

	validateBOMContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: reqNode,
		Started:          true,
	})
	require.NoError(t, err)

	code, reader, err := validateBOMContainer.Exec(ctx, []string{"/piperbin/piper", "validateBOM", "--skip"}, exec.WithWorkingDir("/invalid-bom"))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	outputBytes, err := io.ReadAll(reader)
	assert.NoError(t, err)
	output := string(outputBytes)
	assert.Contains(t, output, "info  validateBOM - BOM validation skipped (skip: true)")

	assert.NotContains(t, output, "info  validateBOM - Found")
	assert.NotContains(t, output, "info  validateBOM - Validating BOM file:")
}

// TestValidateBOMIntegrationNoBOMs tests that validateBOM succeeds silently when no BOMs are found
func TestValidateBOMIntegrationNoBOMs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	reqNode := testcontainers.ContainerRequest{
		Image: DOCKER_IMAGE_VALIDATEBOM,
		Cmd:   []string{"tail", "-f"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      filepath.Join(pwd, "integration", "testdata", "TestValidateBOMIntegration", "no-boms"),
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

	validateBOMContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: reqNode,
		Started:          true,
	})
	require.NoError(t, err)

	code, reader, err := validateBOMContainer.Exec(ctx, []string{"/piperbin/piper", "validateBOM"}, exec.WithWorkingDir("/no-boms"))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	outputBytes, err := io.ReadAll(reader)
	assert.NoError(t, err)
	output := string(outputBytes)
	assert.Contains(t, output, "info  validateBOM - No BOM files found matching pattern")
	assert.Contains(t, output, "skipping validation")
}

// TestValidateBOMIntegrationMultipleBOMs tests that validateBOM can handle multiple BOMs
func TestValidateBOMIntegrationMultipleBOMs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	reqNode := testcontainers.ContainerRequest{
		Image: DOCKER_IMAGE_VALIDATEBOM,
		Cmd:   []string{"tail", "-f"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      filepath.Join(pwd, "integration", "testdata", "TestValidateBOMIntegration", "multiple-boms"),
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

	validateBOMContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: reqNode,
		Started:          true,
	})
	require.NoError(t, err)

	code, reader, err := validateBOMContainer.Exec(ctx, []string{"/piperbin/piper", "validateBOM"}, exec.WithWorkingDir("/multiple-boms"))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	outputBytes, err := io.ReadAll(reader)
	assert.NoError(t, err)
	output := string(outputBytes)
	assert.Contains(t, output, "info  validateBOM - Found 3 BOM file(s) to validate")
	assert.Contains(t, output, "info  validateBOM - BOM validation complete: 3/3 files validated successfully")
}
