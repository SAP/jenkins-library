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
	defer testTimer("TestGolangBuild_Project1", time.Now())

	t.Parallel()

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:       "golang:1",
		TestDir:     []string{"testdata", "TestGolangIntegration", "golang-project1"},
		ExecNoLogin: true,
	})
	err := container.whenRunningPiperCommand("golangBuild")
	assert.NoError(t, err)

	container.assertHasOutput(t, "info  golangBuild - running command: go install gotest.tools/gotestsum@latest")
	container.assertHasOutput(t, "info  golangBuild - running command: go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest")
	container.assertHasOutput(t, "info  golangBuild - running command: gotestsum --junitfile TEST-go.xml -- -coverprofile=cover.out ./...")
	container.assertHasOutput(t, "info  golangBuild - DONE 8 tests")
	container.assertHasOutput(t, "info  golangBuild - running command: go tool cover -html cover.out -o coverage.html")
	container.assertHasOutput(t, "info  golangBuild - running command: gotestsum --junitfile TEST-integration.xml -- -tags=integration ./...")
	container.assertHasOutput(t, "info  golangBuild - running command: cyclonedx-gomod mod -licenses -test -output bom-golang.xml")
	container.assertHasOutput(t, "info  golangBuild - running command: go build -trimpath -o golang-app-linux.amd64 cmd/server/server.go")
	container.assertHasOutput(t, "info  golangBuild - SUCCESS")

	container.assertHasFile(t, "/project/TEST-go.xml")
	container.assertHasFile(t, "/project/TEST-integration.xml")
	container.assertHasFile(t, "/project/bom-golang.xml")
	container.assertHasFile(t, "/project/cover.out")
	container.assertHasFile(t, "/project/coverage.html")
	container.assertHasFile(t, "/project/golang-app-linux.amd64")
}

// This test extends TestGolangBuild_Project1 with multi-package build
func TestGolangBuild_Project1_Multipackage(t *testing.T) {
	defer testTimer("TestGolangBuild_Project1_Multipackage", time.Now())

	t.Parallel()

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:       "golang:1",
		TestDir:     []string{"testdata", "TestGolangIntegration", "golang-project1"},
		ExecNoLogin: true,
	})
	err := container.whenRunningPiperCommand("golangBuild", "--packages", "github.com/example/golang-app/cmd/server,github.com/example/golang-app/cmd/helper")
	assert.NoError(t, err)

	container.assertHasOutput(t, "info  golangBuild - running command: go install gotest.tools/gotestsum@latest")
	container.assertHasOutput(t, "info  golangBuild - running command: go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest")
	container.assertHasOutput(t, "info  golangBuild - running command: gotestsum --junitfile TEST-go.xml -- -coverprofile=cover.out ./...")
	container.assertHasOutput(t, "info  golangBuild - DONE 8 tests")
	container.assertHasOutput(t, "info  golangBuild - running command: go tool cover -html cover.out -o coverage.html")
	container.assertHasOutput(t, "info  golangBuild - running command: gotestsum --junitfile TEST-integration.xml -- -tags=integration ./...")
	container.assertHasOutput(t, "info  golangBuild - running command: cyclonedx-gomod mod -licenses -test -output bom-golang.xml")
	container.assertHasOutput(t, "info  golangBuild - running command: go build -trimpath -o golang-app-linux-amd64/ github.com/example/golang-app/cmd/server github.com/example/golang-app/cmd/helper")
	container.assertHasOutput(t, "info  golangBuild - SUCCESS")

	container.assertHasFile(t, "/project/TEST-go.xml")
	container.assertHasFile(t, "/project/TEST-integration.xml")
	container.assertHasFile(t, "/project/bom-golang.xml")
	container.assertHasFile(t, "/project/cover.out")
	container.assertHasFile(t, "/project/coverage.html")
	container.assertHasFile(t, "/project/golang-app-linux-amd64/server")
	container.assertHasFile(t, "/project/golang-app-linux-amd64/helper")
}

// In this test, the piper golangBuild command only builds the project with the entry point at the project root.
// The configuration for golangBuild can be found in testdata/TestGolangIntegration/golang-project2/.pipeline/config.yml
func TestGolangBuild_Project2(t *testing.T) {
	defer testTimer("TestGolangBuild_Project2", time.Now())

	t.Parallel()

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:       "golang:1",
		TestDir:     []string{"testdata", "TestGolangIntegration", "golang-project2"},
		ExecNoLogin: true,
	})
	err := container.whenRunningPiperCommand("golangBuild")
	assert.NoError(t, err)

	container.assertHasNoOutput(t, "info  golangBuild - running command: go install gotest.tools/gotestsum@latest")
	container.assertHasNoOutput(t, "info  golangBuild - running command: go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest")
	container.assertHasNoOutput(t, "info  golangBuild - running command: gotestsum --junitfile TEST-go.xml -- -coverprofile=cover.out ./...")
	container.assertHasNoOutput(t, "info  golangBuild - running command: go tool cover -html cover.out -o coverage.html")
	container.assertHasNoOutput(t, "info  golangBuild - running command: gotestsum --junitfile TEST-integration.xml -- -tags=integration ./...")
	container.assertHasNoOutput(t, "info  golangBuild - running command: cyclonedx-gomod mod -licenses -test -output bom-golang.xml")
	container.assertHasOutput(t, "info  golangBuild - running command: go build -trimpath -o golang-app-linux.amd64")
	container.assertHasOutput(t, "info  golangBuild - SUCCESS")
}
