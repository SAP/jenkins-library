//go:build integration
// +build integration

// can be execute with go test -tags=integration ./integration/...

package main

import (
	"testing"
)

func TestNpmProject(t *testing.T) {
	t.Parallel()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "paketobuildpacks/builder:full",
		User:    "cnb",
		TestDir: []string{"testdata", "TestMtaIntegration", "npm"},
	})

	container.whenRunningPiperCommand("cnbBuild", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", "test")

	container.assertHasOutput(t, "running command: /cnb/lifecycle/detector")
	container.assertHasOutput(t, "Paketo NPM Start Buildpack")
	container.assertHasOutput(t, "Saving test/not-found:0.0.1")
	container.assertHasOutput(t, "failed to write image to the following tags: [test/not-found:0.0.1")
}

func TestNonZipPath(t *testing.T) {
	t.Parallel()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "paketobuildpacks/builder:full",
		User:    "cnb",
		TestDir: []string{"testdata", "TestMtaIntegration", "npm"},
	})

	container.runScriptInsideContainer("touch not_a_zip")
	container.whenRunningPiperCommand("cnbBuild", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", "test", "--path", "not_a_zip")

	container.assertHasOutput(t, "step execution failed - Copying  'not_a_zip' into '/workspace' failed: application path must be a directory or zip")
}

func TestNpmCustomBuildpacksFullProject(t *testing.T) {
	t.Parallel()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "paketobuildpacks/builder:full",
		User:    "cnb",
		TestDir: []string{"testdata", "TestMtaIntegration", "npm"},
	})

	container.whenRunningPiperCommand("cnbBuild", "--buildpacks", "gcr.io/paketo-buildpacks/nodejs", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", "test")

	container.assertHasOutput(t, "Setting custom buildpacks: '[gcr.io/paketo-buildpacks/nodejs]'")
	container.assertHasOutput(t, "Downloading buildpack 'gcr.io/paketo-buildpacks/nodejs' to /tmp/nodejs")
	container.assertHasOutput(t, "running command: /cnb/lifecycle/detector")
	container.assertHasOutput(t, "Paketo NPM Start Buildpack")
	container.assertHasOutput(t, "Saving test/not-found:0.0.1")
	container.assertHasOutput(t, "failed to write image to the following tags: [test/not-found:0.0.1")
}

func TestNpmCustomBuildpacksBuildpacklessProject(t *testing.T) {
	t.Parallel()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "paketobuildpacks/builder:buildpackless-full",
		User:    "cnb",
		TestDir: []string{"testdata", "TestMtaIntegration", "npm"},
	})

	container.whenRunningPiperCommand("cnbBuild", "--buildpacks", "gcr.io/paketo-buildpacks/nodejs", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", "test")

	container.assertHasOutput(t, "Setting custom buildpacks: '[gcr.io/paketo-buildpacks/nodejs]'")
	container.assertHasOutput(t, "Downloading buildpack 'gcr.io/paketo-buildpacks/nodejs' to /tmp/nodejs")
	container.assertHasOutput(t, "running command: /cnb/lifecycle/detector")
	container.assertHasOutput(t, "Paketo NPM Start Buildpack")
	container.assertHasOutput(t, "Saving test/not-found:0.0.1")
	container.assertHasOutput(t, "failed to write image to the following tags: [test/not-found:0.0.1")
}

func TestWrongBuilderProject(t *testing.T) {
	t.Parallel()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "nginx:latest",
		TestDir: []string{"testdata", "TestMtaIntegration", "npm"},
	})

	container.whenRunningPiperCommand("cnbBuild", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", "test")

	container.assertHasOutput(t, "the provided dockerImage is not a valid builder")
}
