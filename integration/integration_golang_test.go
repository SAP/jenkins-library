//go:build integration
// +build integration

// can be execute with go test -tags=integration ./integration/...

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// In this test the piper command golangBuild performs testing, BOM file creation and building a project with entry point in the cmd/server/server.go
// The configuration for golangBuild can be found in testdata/TestGolangIntegration/golang-project1/.pipeline/config.yml
func TestGolangBuild_Project1(t *testing.T) {
	t.Parallel()

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:       "golang:1",
		TestDir:     []string{"testdata", "TestGolangIntegration", "golang-project1"},
		ExecNoLogin: true,
	})
	defer container.terminate(t)

	err := container.whenRunningPiperCommand("golangBuild")
	assert.NoError(t, err)

	container.assertHasOutput(t,
		"info  golangBuild - running command: go install gotest.tools/gotestsum@latest",
		"info  golangBuild - running command: go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest",
		"info  golangBuild - running command: gotestsum --junitfile TEST-go.xml -- -coverprofile=cover.out ./...",
		"info  golangBuild - DONE 8 tests",
		"info  golangBuild - running command: go tool cover -html cover.out -o coverage.html",
		"info  golangBuild - running command: gotestsum --junitfile TEST-integration.xml -- -tags=integration ./...",
		"info  golangBuild - running command: cyclonedx-gomod mod -licenses -test -output bom-golang.xml",
		"info  golangBuild - running command: go build -trimpath -o golang-app-linux.amd64 cmd/server/server.go",
		"info  golangBuild - SUCCESS",
	)

	container.assertHasFiles(t,
		"/project/TEST-go.xml",
		"/project/TEST-integration.xml",
		"/project/bom-golang.xml",
		"/project/cover.out",
		"/project/coverage.html",
		"/project/golang-app-linux.amd64",
	)
}

// This test extends TestGolangBuild_Project1 with multi-package build
func TestGolangBuild_Project1_Multipackage(t *testing.T) {
	t.Parallel()

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:       "golang:1",
		TestDir:     []string{"testdata", "TestGolangIntegration", "golang-project1"},
		ExecNoLogin: true,
	})
	defer container.terminate(t)

	err := container.whenRunningPiperCommand("golangBuild", "--packages", "github.com/example/golang-app/cmd/server,github.com/example/golang-app/cmd/helper")
	assert.NoError(t, err)

	container.assertHasOutput(t, "info  golangBuild - running command: go install gotest.tools/gotestsum@latest",
		"info  golangBuild - running command: go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest",
		"info  golangBuild - running command: gotestsum --junitfile TEST-go.xml -- -coverprofile=cover.out ./...",
		"info  golangBuild - DONE 8 tests",
		"info  golangBuild - running command: go tool cover -html cover.out -o coverage.html",
		"info  golangBuild - running command: gotestsum --junitfile TEST-integration.xml -- -tags=integration ./...",
		"info  golangBuild - running command: cyclonedx-gomod mod -licenses -test -output bom-golang.xml",
		"info  golangBuild - running command: go build -trimpath -o golang-app-linux-amd64/ github.com/example/golang-app/cmd/server github.com/example/golang-app/cmd/helper",
		"info  golangBuild - SUCCESS",
	)

	container.assertHasFiles(t,
		"/project/TEST-go.xml",
		"/project/TEST-integration.xml",
		"/project/bom-golang.xml",
		"/project/cover.out",
		"/project/coverage.html",
		"/project/golang-app-linux-amd64/server",
		"/project/golang-app-linux-amd64/helper",
	)
}

// In this test, the piper golangBuild command only builds the project with the entry point at the project root.
// The configuration for golangBuild can be found in testdata/TestGolangIntegration/golang-project2/.pipeline/config.yml
func TestGolangBuild_Project2(t *testing.T) {
	t.Parallel()

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:       "golang:1",
		TestDir:     []string{"testdata", "TestGolangIntegration", "golang-project2"},
		ExecNoLogin: true,
	})
	defer container.terminate(t)

	err := container.whenRunningPiperCommand("golangBuild")
	assert.NoError(t, err)

	container.assertHasNoOutput(t,
		"info  golangBuild - running command: go install gotest.tools/gotestsum@latest",
		"info  golangBuild - running command: go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest",
		"info  golangBuild - running command: gotestsum --junitfile TEST-go.xml -- -coverprofile=cover.out ./...",
		"info  golangBuild - running command: go tool cover -html cover.out -o coverage.html",
		"info  golangBuild - running command: gotestsum --junitfile TEST-integration.xml -- -tags=integration ./...",
		"info  golangBuild - running command: cyclonedx-gomod mod -licenses -test -output bom-golang.xml",
	)

	container.assertHasOutput(t,
		"info  golangBuild - running command: go build -trimpath -o golang-app-linux.amd64",
		"info  golangBuild - SUCCESS",
	)
}
