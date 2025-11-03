//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestGolangIntegration ./integration/...

package main

import (
	"testing"

	"github.com/SAP/jenkins-library/integration/testhelper"
	"github.com/stretchr/testify/assert"
)

const DOCKER_IMAGE_GOLANG = "golang:1"

// In this test the piper command golangBuild performs testing, BOM file creation and building a project with entry point in the cmd/server/server.go
// The configuration for golangBuild can be found in testdata/TestGolangIntegration/golang-project1/.pipeline/config.yml
func TestGolangIntegrationBuildProject1(t *testing.T) {
	t.Parallel()

	container := testhelper.StartPiperContainer(t, testhelper.ContainerConfig{
		Image:    DOCKER_IMAGE_GOLANG,
		TestData: "TestGolangIntegration/golang-project1",
		WorkDir:  "/golang-project1",
	})

	output := testhelper.RunPiper(t, container, "/golang-project1", "golangBuild")

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
	testhelper.AssertFileExists(t, container,
		"/golang-project1/TEST-go.xml",
		"/golang-project1/TEST-integration.xml",
		"/golang-project1/bom-golang.xml",
		"/golang-project1/cover.out",
		"/golang-project1/coverage.html",
		"/golang-project1/golang-app-linux.amd64",
	)
}

// This test extends TestGolangIntegrationBuildProject1 with multi-package build
func TestGolangIntegrationBuildProject1MultiPackage(t *testing.T) {
	t.Parallel()

	container := testhelper.StartPiperContainer(t, testhelper.ContainerConfig{
		Image:    DOCKER_IMAGE_GOLANG,
		TestData: "TestGolangIntegration/golang-project1",
		WorkDir:  "/golang-project1",
	})

	output := testhelper.RunPiper(t, container, "/golang-project1", "golangBuild", "--packages", "github.com/example/golang-app/cmd/server,github.com/example/golang-app/cmd/helper")

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
	testhelper.AssertFileExists(t, container,
		"/golang-project1/TEST-go.xml",
		"/golang-project1/TEST-integration.xml",
		"/golang-project1/bom-golang.xml",
		"/golang-project1/cover.out",
		"/golang-project1/coverage.html",
		"/golang-project1/golang-app-linux-amd64/server",
		"/golang-project1/golang-app-linux-amd64/helper",
	)
}

// In this test, the piper golangBuild command only builds the project with the entry point at the project root.
// The configuration for golangBuild can be found in testdata/TestGolangIntegration/golang-project2/.pipeline/config.yml
func TestGolangIntegrationBuildProject2(t *testing.T) {
	t.Parallel()

	container := testhelper.StartPiperContainer(t, testhelper.ContainerConfig{
		Image:    DOCKER_IMAGE_GOLANG,
		TestData: "TestGolangIntegration/golang-project2",
		WorkDir:  "/golang-project2",
	})

	output := testhelper.RunPiper(t, container, "/golang-project2", "golangBuild")

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

	container := testhelper.StartPiperContainer(t, testhelper.ContainerConfig{
		Image:    DOCKER_IMAGE_GOLANG,
		TestData: "TestGolangIntegration/golang-project1",
		WorkDir:  "/golang-project1",
	})

	// First, run golangBuild to generate the BOM
	output := testhelper.RunPiper(t, container, "/golang-project1", "golangBuild")
	assert.Contains(t, output, "info  golangBuild - running command: cyclonedx-gomod mod -licenses -verbose=false -test -output bom-golang.xml")
	assert.Contains(t, output, "info  golangBuild - SUCCESS")

	// Verify BOM file was created
	testhelper.AssertFileExists(t, container, "/golang-project1/bom-golang.xml")

	// Now run validateBOM on the generated BOM
	output = testhelper.RunPiper(t, container, "/golang-project1", "validateBOM")
	assert.Contains(t, output, "info  validateBOM - Found 1 BOM file(s) to validate")
	assert.Contains(t, output, "info  validateBOM - Validating BOM file:")
	assert.Contains(t, output, "bom-golang.xml")
	assert.Contains(t, output, "info  validateBOM - BOM validation passed:")
	assert.Contains(t, output, "info  validateBOM - BOM validation complete: 1/1 files validated successfully")
}
