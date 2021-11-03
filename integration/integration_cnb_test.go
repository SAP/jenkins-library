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

	container.whenRunningPiperCommand("cnbBuild", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", "test", "--buildEnvVars", "BP_NODE_VERSION=15.14.0")

	container.assertHasOutput(t, "running command: /cnb/lifecycle/detector")
	container.assertHasOutput(t, "Selected Node Engine version (using BP_NODE_VERSION): 15.14.0")
	container.assertHasOutput(t, "Paketo NPM Start Buildpack")
	container.assertHasOutput(t, "Saving test/not-found:0.0.1")
	container.assertHasOutput(t, "failed to write image to the following tags: [test/not-found:0.0.1")
}

func TestProjectDescriptor(t *testing.T) {
	t.Parallel()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "paketobuildpacks/builder:full",
		User:    "cnb",
		TestDir: []string{"testdata", "TestCnbIntegration", "project"},
	})

	container.whenRunningPiperCommand("cnbBuild", "-v", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", "test")

	container.assertHasOutput(t, "running command: /cnb/lifecycle/detector")
	container.assertHasOutput(t, "/project/Dockerfile doesn't match include pattern, ignoring")
	container.assertHasOutput(t, "/project/srv/hello.js matches include pattern")
	container.assertHasOutput(t, "/project/srv/hello.js matches include pattern")
	container.assertHasOutput(t, "Downloading buildpack")
	container.assertHasOutput(t, "Setting custom environment variables: '[BP_NODE_VERSION=15.14.0]'")
	container.assertHasOutput(t, "Selected Node Engine version (using BP_NODE_VERSION): 15.14.0")
	container.assertHasOutput(t, "Paketo NPM Start Buildpack")
	container.assertHasOutput(t, "Saving test/not-found:0.0.1")
	container.assertHasOutput(t, "failed to write image to the following tags: [test/not-found:0.0.1")
}

func TestZipPath(t *testing.T) {
	t.Parallel()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "paketobuildpacks/builder:full",
		User:    "cnb",
		TestDir: []string{"testdata", "TestCnbIntegration", "zip"},
	})

	container.whenRunningPiperCommand("cnbBuild", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", "test", "--path", "go.zip")

	container.assertHasOutput(t, "running command: /cnb/lifecycle/detector")
	container.assertHasOutput(t, "Installing Go")
	container.assertHasOutput(t, "Paketo Go Build Buildpack")
	container.assertHasOutput(t, "Saving test/not-found:0.0.1")
}

func TestNonZipPath(t *testing.T) {
	t.Parallel()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "paketobuildpacks/builder:full",
		User:    "cnb",
		TestDir: []string{"testdata", "TestMtaIntegration", "npm"},
	})

	container.whenRunningPiperCommand("cnbBuild", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", "test", "--path", "mta.yaml")

	container.assertHasOutput(t, "Copying  'mta.yaml' into '/workspace' failed: application path must be a directory or zip")
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

func TestBindings(t *testing.T) {
	t.Parallel()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "paketobuildpacks/builder:full",
		User:    "cnb",
		TestDir: []string{"testdata"},
	})

	container.whenRunningPiperCommand("cnbBuild", "--customConfig", "TestCnbIntegration/config.yml", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", "test", "--path", "TestMtaIntegration/maven")

	container.assertHasOutput(t, "bindings/maven-settings/settings.xml: only whitespace content allowed before start tag")
	container.assertHasFile(t, "/tmp/platform/bindings/dummy-binding/type")
	container.assertHasFile(t, "/tmp/platform/bindings/dummy-binding/dummy.yml")
}
