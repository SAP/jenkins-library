//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestNPMIntegration ./integration/...

package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

// TestNPMIntegrationPublishPrerelease verifies that publishing a prerelease version
// automatically adds the --tag prerelease flag (required by npm 11+)
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
	assert.Contains(t, output, "Detected prerelease version")
	assert.Contains(t, output, "0.0.1-20251112123456")
	assert.Contains(t, output, "adding --tag prerelease")

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
	assert.NotContains(t, output, "adding --tag prerelease")

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
