//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestGolangIntegration ./integration/...

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// In this test the piper command golangBuild performs testing, BOM file creation and building a project with entry point in the cmd/server/server.go
// The configuration for golangBuild can be found in testdata/TestGolangIntegration/golang-project1/.pipeline/config.yml
func TestGolangIntegrationBuildProject1(t *testing.T) {
	// t.Parallel()

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:       "golang:1",
		TestDir:     []string{"testdata", "TestGolangIntegration", "golang-project1"},
		ExecNoLogin: true,
	})
	defer container.terminate(t)

	// Debug: Show initial directory structure
	t.Log("=== DEBUG: Initial project directory structure ===")
	err := container.runScriptInsideContainer("find /project -type f -name '*.go' -o -name '*.yml' -o -name '*.yaml' | head -20")
	if err != nil {
		t.Logf("Failed to list project files: %v", err)
	}

	err = container.whenRunningPiperCommand("golangBuild")
	assert.NoError(t, err)

	// Debug: Show complete piper output
	t.Log("=== DEBUG: Complete piper command output ===")
	buffer, outputErr := container.getPiperOutput()
	if outputErr != nil {
		t.Logf("Failed to get piper output: %v", outputErr)
	} else {
		t.Logf("Piper output:\n%s", buffer.String())
	}

	// Debug: Show final directory contents
	t.Log("=== DEBUG: Final project directory contents ===")
	err = container.runScriptInsideContainer("ls -la /project/")
	if err != nil {
		t.Logf("Failed to list final directory: %v", err)
	}

	// Debug: Show specific file contents
	t.Log("=== DEBUG: Generated file details ===")
	err = container.runScriptInsideContainer("if [ -f /project/TEST-go.xml ]; then echo 'TEST-go.xml size:'; wc -l /project/TEST-go.xml; echo 'First 10 lines:'; head -10 /project/TEST-go.xml; fi")
	if err != nil {
		t.Logf("Failed to show TEST-go.xml details: %v", err)
	}

	err = container.runScriptInsideContainer("if [ -f /project/bom-golang.xml ]; then echo 'bom-golang.xml size:'; wc -l /project/bom-golang.xml; echo 'First 10 lines:'; head -10 /project/bom-golang.xml; fi")
	if err != nil {
		t.Logf("Failed to show bom-golang.xml details: %v", err)
	}

	err = container.runScriptInsideContainer("if [ -f /project/golang-app-linux.amd64 ]; then echo 'Binary details:'; file /project/golang-app-linux.amd64; ls -lh /project/golang-app-linux.amd64; fi")
	if err != nil {
		t.Logf("Failed to show binary details: %v", err)
	}

	container.assertHasOutput(t,
		"info  golangBuild - running command: go install gotest.tools/gotestsum@latest",
		"info  golangBuild - running command: go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@v1.9.0",
		"info  golangBuild - running command: gotestsum --junitfile TEST-go.xml --jsonfile unit-report.out -- -coverprofile=cover.out -tags=unit ./...",
		"info  golangBuild - DONE 8 tests",
		"info  golangBuild - running command: go tool cover -html cover.out -o coverage.html",
		"info  golangBuild - running command: gotestsum --junitfile TEST-integration.xml --jsonfile integration-report.out -- -tags=integration ./...",
		"info  golangBuild - running command: cyclonedx-gomod mod -licenses -verbose=false -test -output bom-golang.xml",
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

// This test extends TestGolangIntegrationBuildProject1 with multi-package build
func TestGolangIntegrationBuildProject1MultiPackage(t *testing.T) {
	// t.Parallel()

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:       "golang:1",
		TestDir:     []string{"testdata", "TestGolangIntegration", "golang-project1"},
		ExecNoLogin: true,
	})
	defer container.terminate(t)

	err := container.whenRunningPiperCommand("golangBuild", "--packages", "github.com/example/golang-app/cmd/server,github.com/example/golang-app/cmd/helper")
	assert.NoError(t, err)

	container.assertHasOutput(t, "info  golangBuild - running command: go install gotest.tools/gotestsum@latest",
		"info  golangBuild - running command: go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@v1.9.0",
		"info  golangBuild - running command: gotestsum --junitfile TEST-go.xml --jsonfile unit-report.out -- -coverprofile=cover.out -tags=unit ./...",
		"info  golangBuild - DONE 8 tests",
		"info  golangBuild - running command: go tool cover -html cover.out -o coverage.html",
		"info  golangBuild - running command: gotestsum --junitfile TEST-integration.xml --jsonfile integration-report.out -- -tags=integration ./...",
		"info  golangBuild - running command: cyclonedx-gomod mod -licenses -verbose=false -test -output bom-golang.xml",
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
func TestGolangIntegrationBuildProject2(t *testing.T) {
	// t.Parallel()

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
		"info  golangBuild - running command: go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@v1.9.0",
		"info  golangBuild - running command: gotestsum --junitfile TEST-go.xml --jsonfile unit-report.out -- -coverprofile=cover.out -tags=unit ./...",
		"info  golangBuild - running command: go tool cover -html cover.out -o coverage.html",
		"info  golangBuild - running command: gotestsum --junitfile TEST-integration.xml --jsonfile integration-report.out -- -tags=integration ./...",
		"info  golangBuild - running command: cyclonedx-gomod mod -licenses -verbose=false -test -output bom-golang.xml",
	)

	container.assertHasOutput(t,
		"info  golangBuild - running command: go build -trimpath -o golang-app-linux.amd64",
		"info  golangBuild - SUCCESS",
	)
}
