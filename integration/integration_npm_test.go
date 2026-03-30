//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestNPMIntegration ./integration/...

package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const cyclonedxSchemaVersion = "1.4"

func TestNPMIntegrationRunScriptsWithOptions(t *testing.T) {
	t.Parallel()

	container := StartPiperContainer(t, ContainerConfig{
		Image:    "node:24-bookworm",
		TestData: "TestNpmIntegration/runScriptsWithOptions",
		WorkDir:  "/runScriptsWithOptions",
	})

	output := RunPiper(t, container, "/runScriptsWithOptions",
		"npmExecuteScripts",
		"--runScripts=start",
		"--scriptOptions=--tag,tag1")

	assert.Contains(t, output, "info  npmExecuteScripts - running command: npm run start -- --tag tag1")
	assert.Contains(t, output, "[ '--tag', 'tag1' ]")
}

func TestNPMIntegrationRegistrySetInFlags(t *testing.T) {
	t.Parallel()

	container := StartPiperContainer(t, ContainerConfig{
		Image:    "node:24-bookworm",
		TestData: "TestNpmIntegration/registrySetInFlags",
		WorkDir:  "/registrySetInFlags",
	})

	output := RunPiper(t, container, "/registrySetInFlags",
		"npmExecuteScripts",
		"--install",
		"--runScripts=ci-build",
		"--defaultNpmRegistry=https://foo.bar")

	assert.Contains(t, output, "https://foo.bar")
}

func TestNPMIntegrationRegistrySetInNpmrc(t *testing.T) {
	t.Parallel()

	container := StartPiperContainer(t, ContainerConfig{
		Image:    "node:24-bookworm",
		TestData: "TestNpmIntegration/registrySetInNpmrc",
		WorkDir:  "/registrySetInNpmrc",
	})

	output := RunPiper(t, container, "/registrySetInNpmrc",
		"npmExecuteScripts",
		"--install",
		"--runScripts=ci-build")

	assert.Contains(t, output, "https://example.com")
}

func TestNPMIntegrationRegistryWithTwoModules(t *testing.T) {
	t.Parallel()

	container := StartPiperContainer(t, ContainerConfig{
		Image:    "node:24-bookworm",
		TestData: "TestNpmIntegration/registryWithTwoModules",
		WorkDir:  "/registryWithTwoModules",
	})

	output := RunPiper(t, container, "/registryWithTwoModules",
		"npmExecuteScripts",
		"--install",
		"--runScripts=ci-build",
		"--defaultNpmRegistry=https://foo.bar")

	assert.Contains(t, output, "https://example.com")
	assert.Contains(t, output, "https://foo.bar")
}

// TestNPMIntegrationPublishPrerelease verifies that not passing publishTag flag
// runs npm publish with 'prerelease' tag for version with prerelease part (required by npm 11+ for prerelease versions)
func TestNPMIntegrationPublishPrerelease(t *testing.T) {
	t.Parallel()

	container := StartPiperContainer(t, ContainerConfig{
		Image:    "node:24-bookworm",
		TestData: "TestNpmIntegration/publishPrerelease",
		WorkDir:  "/publishPrerelease",
	})

	// We expect this to fail because we're using a fake registry,
	// but we want to verify that the --tag prerelease flag is added
	exitCode, output := RunPiperExpectFailure(t, container, "/publishPrerelease",
		"npmExecuteScripts",
		"--publish",
		"--repositoryUrl=https://fake-registry.example.com",
		"--repositoryUsername=test-user",
		"--repositoryPassword=test-pass")

	// Verify the command detected the prerelease version
	assert.Contains(t, output, "No publish tag provided, using 'prerelease' based on version 0.0.1-20251112123456")
	assert.Contains(t, output, "--tag prerelease")

	// Verify it attempted to publish (will fail due to fake registry, but that's expected)
	assert.Contains(t, output, "triggering publish for package.json")

	// Command should fail because the registry doesn't exist
	assert.NotEqual(t, 0, exitCode, "Expected command to fail with fake registry")
}

// TestNPMIntegrationPublishStable verifies that publishing a stable version
// does NOT add the --tag flag (default npm behavior for stable versions)
func TestNPMIntegrationPublishStable(t *testing.T) {
	t.Parallel()

	container := StartPiperContainer(t, ContainerConfig{
		Image:    "node:24-bookworm",
		TestData: "TestNpmIntegration/publishStable",
		WorkDir:  "/publishStable",
	})

	// We expect this to fail because we're using a fake registry,
	// but we want to verify that the --tag flag is NOT added for stable versions
	exitCode, output := RunPiperExpectFailure(t, container, "/publishStable",
		"npmExecuteScripts",
		"--publish",
		"--repositoryUrl=https://fake-registry.example.com",
		"--repositoryUsername=test-user",
		"--repositoryPassword=test-pass")

	// Verify it attempted to publish
	assert.Contains(t, output, "triggering publish for package.json")

	// Verify it did NOT detect a prerelease version or add --tag flag
	assert.NotContains(t, output, "Detected prerelease version")
	assert.NotContains(t, output, "--tag prerelease")

	// For stable versions, there should be no mention of --tag in the output
	// (we're checking the logs don't show the prerelease-specific logic)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Detected prerelease") {
			t.Errorf("Should not detect prerelease for stable version 1.0.0, but found: %s", line)
		}
	}

	// Command should fail because the registry doesn't exist
	assert.NotEqual(t, 0, exitCode, "Expected command to fail with fake registry")
}

// TestNPMIntegrationCreateBOMNpm verifies that running npmExecuteScripts with --createBOM
// on an npm project calls createNpmBOM via cyclonedx-npm and produces a bom-npm.xml file
// using CycloneDX schema version 1.4.
func TestNPMIntegrationCreateBOMNpm(t *testing.T) {
	t.Parallel()

	container := StartPiperContainer(t, ContainerConfig{
		Image:    "node:24-bookworm",
		TestData: "TestNpmIntegration/createBOM",
		WorkDir:  "/createBOM",
	})

	output := RunPiper(t, container, "/createBOM",
		"npmExecuteScripts",
		"--createBOM",
		"--install",
		"--runScripts=ci-build")

	// Verify cyclonedx-npm is called with expected arguments for npm
	assert.Contains(t, output,
		"running command: ./tmp/node_modules/.bin/cyclonedx-npm --output-format XML --spec-version "+cyclonedxSchemaVersion+" --omit dev --output-file bom-npm.xml package.json",
		"BOM generation should invoke cyclonedx-npm with the expected arguments for npm")

	// Verify the generated BOM file exists and references the correct CycloneDX schema version
	bomContent := ReadFile(t, container, "/createBOM/bom-npm.xml")

	assert.NotEmpty(t, bomContent, "bom-npm.xml should not be empty")
	assert.Contains(t, string(bomContent), "http://cyclonedx.org/schema/bom/"+cyclonedxSchemaVersion,
		"bom-npm.xml should reference CycloneDX schema version "+cyclonedxSchemaVersion)
}

// TestNPMIntegrationCreateBOMYarn verifies that running npmExecuteScripts with --createBOM
// on a yarn project (identified by yarn.lock) calls createNpmBOM via cyclonedx-npm and
// produces a bom-npm.xml file using CycloneDX schema version 1.4.
func TestNPMIntegrationCreateBOMYarn(t *testing.T) {
	t.Parallel()

	container := StartPiperContainer(t, ContainerConfig{
		Image:    "node:24-bookworm",
		TestData: "TestNpmIntegration/createBOMYarn",
		WorkDir:  "/createBOMYarn",
	})

	output := RunPiper(t, container, "/createBOMYarn",
		"npmExecuteScripts",
		"--createBOM",
		"--install",
		"--runScripts=ci-build")

	// Verify cyclonedx-npm is called with expected arguments for yarn
	assert.Contains(t, output,
		"running command: ./tmp/node_modules/.bin/cyclonedx-npm --output-format XML --spec-version "+cyclonedxSchemaVersion+" --omit dev --output-file bom-npm.xml package.json",
		"BOM generation should invoke cyclonedx-npm with the expected arguments for yarn")

	// Verify the generated BOM file exists and references the correct CycloneDX schema version
	bomContent := ReadFile(t, container, "/createBOMYarn/bom-npm.xml")

	assert.NotEmpty(t, bomContent, "bom-npm.xml should not be empty")
	assert.Contains(t, string(bomContent), "http://cyclonedx.org/schema/bom/"+cyclonedxSchemaVersion,
		"bom-npm.xml should reference CycloneDX schema version "+cyclonedxSchemaVersion)
}

// TestNPMIntegrationCreateBOMPnpm verifies that running npmExecuteScripts with --createBOM
// on a pnpm project (identified by pnpm-lock.yaml) calls createPnpmBOM via cdxgen and
// cyclonedx-cli, and produces a bom-npm.xml file using CycloneDX schema version 1.4.
func TestNPMIntegrationCreateBOMPnpm(t *testing.T) {
	t.Parallel()

	container := StartPiperContainer(t, ContainerConfig{
		Image:    "node:24-bookworm",
		TestData: "TestNpmIntegration/createBOMPnpm",
		WorkDir:  "/createBOMPnpm",
	})

	output := RunPiper(t, container, "/createBOMPnpm",
		"npmExecuteScripts",
		"--createBOM",
		"--install",
		"--runScripts=ci-build")

	// Verify cdxgen is called with expected arguments for pnpm
	assert.Contains(t, output,
		"running command: ./tmp/node_modules/.bin/cdxgen -r -o bom-npm.json --spec-version "+cyclonedxSchemaVersion,
		"BOM generation should invoke cdxgen with the expected arguments for pnpm")

	// Verify cyclonedx-cli is called to convert JSON to XML with expected arguments for pnpm
	// The output version for cyclonedx-cli is expected to be in the format "vX_Y_Z". Ex: 1.4 => v1_4
	outputVersion := fmt.Sprintf("v%s", strings.ReplaceAll(cyclonedxSchemaVersion, ".", "_"))
	assert.Contains(t, output,
		"running command: .pipeline/cyclonedx-linux-x64 convert --input-file bom-npm.json --output-format xml --output-file bom-npm.xml --output-version "+outputVersion,
		"BOM generation should invoke cyclonedx-cli to convert JSON to XML for pnpm")

	bomContent := ReadFile(t, container, "/createBOMPnpm/bom-npm.xml")

	assert.NotEmpty(t, bomContent, "bom-npm.xml should not be empty")
	assert.Contains(t, string(bomContent), "http://cyclonedx.org/schema/bom/"+cyclonedxSchemaVersion,
		"bom-npm.xml should reference CycloneDX schema version "+cyclonedxSchemaVersion)
}
