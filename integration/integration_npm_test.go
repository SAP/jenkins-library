//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestNPMIntegration ./integration/...

package main

import (
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

// TestNPMIntegrationPublishPrerelease verifies that passing publishTag flag
// runs npm publish with passed tag (required by npm 11+ for prerelease versions)
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
		"--publishTag=prerelease",
		"--repositoryUrl=https://fake-registry.example.com",
		"--repositoryUsername=test-user",
		"--repositoryPassword=test-pass")

	// Verify the command detected the prerelease version
	assert.Contains(t, output, "Detected prerelease version")
	assert.Contains(t, output, "0.0.1-20251112123456")
	assert.Contains(t, output, "--tag prerelease")

	// Verify it attempted to publish (will fail due to fake registry, but that's expected)
	assert.Contains(t, output, "triggering publish for package.json")

	// Command should fail because the registry doesn't exist
	assert.NotEqual(t, 0, exitCode, "Expected command to fail with fake registry")
}
