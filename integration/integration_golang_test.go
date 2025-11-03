//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestGolangIntegration ./integration/...

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

const DOCKER_IMAGE_GOLANG = "golang:1"

// In this test the piper command golangBuild performs testing, BOM file creation and building a project with entry point in the cmd/server/server.go
// The configuration for golangBuild can be found in testdata/TestGolangIntegration/golang-project1/.pipeline/config.yml
func TestGolangIntegrationBuildProject1(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	reqNode := testcontainers.ContainerRequest{
		Image: DOCKER_IMAGE_GOLANG,
		Cmd:   []string{"tail", "-f"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      filepath.Join(pwd, "integration", "testdata", "TestGolangIntegration", "golang-project1"),
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

	golangContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: reqNode,
		Started:          true,
	})
	require.NoError(t, err)

	code, reader, err := golangContainer.Exec(ctx, []string{"/piperbin/piper", "golangBuild"}, exec.WithWorkingDir("/golang-project1"))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	outputBytes, err := io.ReadAll(reader)
	assert.NoError(t, err)
	output := string(outputBytes)
	assert.Contains(t, output, "info  golangBuild - running command: go install gotest.tools/gotestsum@latest")
	assert.Contains(t, output, "info  golangBuild - running command: go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@v1.9.0")
	assert.Contains(t, output, "info  golangBuild - running command: gotestsum --junitfile TEST-go.xml --jsonfile unit-report.out -- -coverprofile=cover.out -tags=unit ./...")
	assert.Contains(t, output, "info  golangBuild - DONE 8 tests")
	assert.Contains(t, output, "info  golangBuild - running command: go tool cover -html cover.out -o coverage.html")
	assert.Contains(t, output, "info  golangBuild - running command: gotestsum --junitfile TEST-integration.xml --jsonfile integration-report.out -- -tags=integration ./...")
	assert.Contains(t, output, "info  golangBuild - running command: cyclonedx-gomod mod -licenses -verbose=false -test -output bom-golang.xml")
	assert.Contains(t, output, "info  golangBuild - running command: go build -trimpath -o golang-app-linux.amd64 cmd/server/server.go")
	assert.Contains(t, output, "info  golangBuild - SUCCESS")

	// Verify files were created
	code, _, err = golangContainer.Exec(ctx, []string{"stat",
		"/golang-project1/TEST-go.xml",
		"/golang-project1/TEST-integration.xml",
		"/golang-project1/bom-golang.xml",
		"/golang-project1/cover.out",
		"/golang-project1/coverage.html",
		"/golang-project1/golang-app-linux.amd64",
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, code)
}

// This test extends TestGolangIntegrationBuildProject1 with multi-package build
func TestGolangIntegrationBuildProject1MultiPackage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	reqNode := testcontainers.ContainerRequest{
		Image: DOCKER_IMAGE_GOLANG,
		Cmd:   []string{"tail", "-f"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      filepath.Join(pwd, "integration", "testdata", "TestGolangIntegration", "golang-project1"),
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

	golangContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: reqNode,
		Started:          true,
	})
	require.NoError(t, err)

	code, reader, err := golangContainer.Exec(ctx, []string{"/piperbin/piper", "golangBuild", "--packages", "github.com/example/golang-app/cmd/server,github.com/example/golang-app/cmd/helper"}, exec.WithWorkingDir("/golang-project1"))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	outputBytes, err := io.ReadAll(reader)
	assert.NoError(t, err)
	output := string(outputBytes)
	assert.Contains(t, output, "info  golangBuild - running command: go install gotest.tools/gotestsum@latest")
	assert.Contains(t, output, "info  golangBuild - running command: go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@v1.9.0")
	assert.Contains(t, output, "info  golangBuild - running command: gotestsum --junitfile TEST-go.xml --jsonfile unit-report.out -- -coverprofile=cover.out -tags=unit ./...")
	assert.Contains(t, output, "info  golangBuild - DONE 8 tests")
	assert.Contains(t, output, "info  golangBuild - running command: go tool cover -html cover.out -o coverage.html")
	assert.Contains(t, output, "info  golangBuild - running command: gotestsum --junitfile TEST-integration.xml --jsonfile integration-report.out -- -tags=integration ./...")
	assert.Contains(t, output, "info  golangBuild - running command: cyclonedx-gomod mod -licenses -verbose=false -test -output bom-golang.xml")
	assert.Contains(t, output, "info  golangBuild - running command: go build -trimpath -o golang-app-linux-amd64/ github.com/example/golang-app/cmd/server github.com/example/golang-app/cmd/helper")
	assert.Contains(t, output, "info  golangBuild - SUCCESS")

	// Verify files were created
	code, _, err = golangContainer.Exec(ctx, []string{"stat",
		"/golang-project1/TEST-go.xml",
		"/golang-project1/TEST-integration.xml",
		"/golang-project1/bom-golang.xml",
		"/golang-project1/cover.out",
		"/golang-project1/coverage.html",
		"/golang-project1/golang-app-linux-amd64/server",
		"/golang-project1/golang-app-linux-amd64/helper",
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, code)
}

// In this test, the piper golangBuild command only builds the project with the entry point at the project root.
// The configuration for golangBuild can be found in testdata/TestGolangIntegration/golang-project2/.pipeline/config.yml
func TestGolangIntegrationBuildProject2(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	reqNode := testcontainers.ContainerRequest{
		Image: DOCKER_IMAGE_GOLANG,
		Cmd:   []string{"tail", "-f"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      filepath.Join(pwd, "integration", "testdata", "TestGolangIntegration", "golang-project2"),
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

	golangContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: reqNode,
		Started:          true,
	})
	require.NoError(t, err)

	code, reader, err := golangContainer.Exec(ctx, []string{"/piperbin/piper", "golangBuild"}, exec.WithWorkingDir("/golang-project2"))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	outputBytes, err := io.ReadAll(reader)
	assert.NoError(t, err)
	output := string(outputBytes)

	// Should NOT run testing or BOM generation
	assert.NotContains(t, output, "info  golangBuild - running command: go install gotest.tools/gotestsum@latest")
	assert.NotContains(t, output, "info  golangBuild - running command: go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@v1.9.0")
	assert.NotContains(t, output, "info  golangBuild - running command: gotestsum --junitfile TEST-go.xml --jsonfile unit-report.out -- -coverprofile=cover.out -tags=unit ./...")
	assert.NotContains(t, output, "info  golangBuild - running command: go tool cover -html cover.out -o coverage.html")
	assert.NotContains(t, output, "info  golangBuild - running command: gotestsum --junitfile TEST-integration.xml --jsonfile integration-report.out -- -tags=integration ./...")
	assert.NotContains(t, output, "info  golangBuild - running command: cyclonedx-gomod mod -licenses -verbose=false -test -output bom-golang.xml")

	// Should only run build
	assert.Contains(t, output, "info  golangBuild - running command: go build -trimpath -o golang-app-linux.amd64")
	assert.Contains(t, output, "info  golangBuild - SUCCESS")
}

// This test verifies that the validateBOM step can validate a BOM generated by golangBuild
func TestGolangIntegrationBuildWithBOMValidation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	reqNode := testcontainers.ContainerRequest{
		Image: DOCKER_IMAGE_GOLANG,
		Cmd:   []string{"tail", "-f"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      filepath.Join(pwd, "integration", "testdata", "TestGolangIntegration", "golang-project1"),
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

	golangContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: reqNode,
		Started:          true,
	})
	require.NoError(t, err)

	// First, run golangBuild to generate the BOM
	code, reader, err := golangContainer.Exec(ctx, []string{"/piperbin/piper", "golangBuild"}, exec.WithWorkingDir("/golang-project1"))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	outputBytes, err := io.ReadAll(reader)
	assert.NoError(t, err)
	output := string(outputBytes)
	assert.Contains(t, output, "info  golangBuild - running command: cyclonedx-gomod mod -licenses -verbose=false -test -output bom-golang.xml")
	assert.Contains(t, output, "info  golangBuild - SUCCESS")

	// Verify BOM file was created
	code, _, err = golangContainer.Exec(ctx, []string{"stat", "/golang-project1/bom-golang.xml"})
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	// Now run validateBOM on the generated BOM
	code, reader, err = golangContainer.Exec(ctx, []string{"/piperbin/piper", "validateBOM"}, exec.WithWorkingDir("/golang-project1"))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	outputBytes, err = io.ReadAll(reader)
	assert.NoError(t, err)
	output = string(outputBytes)
	assert.Contains(t, output, "info  validateBOM - Found 1 BOM file(s) to validate")
	assert.Contains(t, output, "info  validateBOM - Validating BOM file:")
	assert.Contains(t, output, "bom-golang.xml")
	assert.Contains(t, output, "info  validateBOM - BOM validation passed:")
	assert.Contains(t, output, "info  validateBOM - BOM validation complete: 1/1 files validated successfully")
}
