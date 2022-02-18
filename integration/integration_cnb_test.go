//go:build integration
// +build integration

// can be executed with go test -tags=integration ./integration/...

package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
)

var registryURL = "localhost:5000"

func setupDockerRegistry(t *testing.T, ctx context.Context) testcontainers.Container {
	reqRegistry := testcontainers.ContainerRequest{
		Image:      "registry:2",
		SkipReaper: true,
	}

	regContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: reqRegistry,
		Started:          true,
	})
	assert.NoError(t, err)

	return regContainer
}

func TestNpmProject(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	registryContainer := setupDockerRegistry(t, ctx)
	defer registryContainer.Terminate(ctx)

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "paketobuildpacks/builder:full",
		User:    "cnb",
		TestDir: []string{"testdata"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
	})

	container.whenRunningPiperCommand("cnbBuild", "--customConfig", "TestCnbIntegration/config_env.yml", "--path", "TestMtaIntegration/npm", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", registryURL)

	container.assertHasOutput(t, "running command: /cnb/lifecycle/creator")
	container.assertHasOutput(t, "Selected Node Engine version (using BP_NODE_VERSION): 16")
	container.assertHasOutput(t, "Paketo NPM Start Buildpack")
	container.assertHasOutput(t, fmt.Sprintf("Saving %s/not-found:0.0.1", registryURL))
	container.assertHasOutput(t, "*** Images (sha256:")
	container.assertHasOutput(t, "SUCCESS")
	container.terminate(t)
}

func TestProjectDescriptor(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	registryContainer := setupDockerRegistry(t, ctx)
	defer registryContainer.Terminate(ctx)

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "paketobuildpacks/builder:full",
		User:    "cnb",
		TestDir: []string{"testdata", "TestCnbIntegration", "project"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
	})

	container.whenRunningPiperCommand("cnbBuild", "-v", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", registryURL)

	container.assertHasOutput(t, "running command: /cnb/lifecycle/creator")
	container.assertHasOutput(t, "Dockerfile doesn't match include pattern, ignoring")
	container.assertHasOutput(t, "srv/hello.js matches include pattern")
	container.assertHasOutput(t, "srv/hello.js matches include pattern")
	container.assertHasOutput(t, "package.json matches include pattern")
	container.assertHasOutput(t, "Downloading buildpack")
	container.assertHasOutput(t, "Setting custom environment variables: 'map[BP_NODE_VERSION:16]'")
	container.assertHasOutput(t, "Selected Node Engine version (using BP_NODE_VERSION): 16")
	container.assertHasOutput(t, "Paketo NPM Start Buildpack")
	container.assertHasOutput(t, fmt.Sprintf("Saving %s/not-found:0.0.1", registryURL))
	container.assertHasOutput(t, "*** Images (sha256:")
	container.assertHasOutput(t, "SUCCESS")
	container.terminate(t)
}

func TestZipPath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	registryContainer := setupDockerRegistry(t, ctx)
	defer registryContainer.Terminate(ctx)

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "paketobuildpacks/builder:full",
		User:    "cnb",
		TestDir: []string{"testdata", "TestCnbIntegration", "zip"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
	})

	container.whenRunningPiperCommand("cnbBuild", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", registryURL, "--path", "go.zip")

	container.assertHasOutput(t, "running command: /cnb/lifecycle/creator")
	container.assertHasOutput(t, "Installing Go")
	container.assertHasOutput(t, "Paketo Go Build Buildpack")
	container.assertHasOutput(t, fmt.Sprintf("Saving %s/not-found:0.0.1", registryURL))
	container.assertHasOutput(t, "*** Images (sha256:")
	container.assertHasOutput(t, "SUCCESS")
	container.terminate(t)
}

func TestNonZipPath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	registryContainer := setupDockerRegistry(t, ctx)
	defer registryContainer.Terminate(ctx)

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "paketobuildpacks/builder:full",
		User:    "cnb",
		TestDir: []string{"testdata", "TestMtaIntegration", "npm"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
	})

	container.whenRunningPiperCommand("cnbBuild", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", registryURL, "--path", "mta.yaml")

	container.assertHasOutput(t, "Copying  '/project/mta.yaml' into '/workspace' failed: application path must be a directory or zip")
	container.terminate(t)
}

func TestNpmCustomBuildpacksFullProject(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	registryContainer := setupDockerRegistry(t, ctx)
	defer registryContainer.Terminate(ctx)

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "paketobuildpacks/builder:full",
		User:    "cnb",
		TestDir: []string{"testdata", "TestMtaIntegration", "npm"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
	})

	container.whenRunningPiperCommand("cnbBuild", "--buildpacks", "gcr.io/paketo-buildpacks/nodejs", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", registryURL)

	container.assertHasOutput(t, "Setting custom buildpacks: '[gcr.io/paketo-buildpacks/nodejs]'")
	container.assertHasOutput(t, "Downloading buildpack 'gcr.io/paketo-buildpacks/nodejs' to /tmp/nodejs")
	container.assertHasOutput(t, "running command: /cnb/lifecycle/creator")
	container.assertHasOutput(t, "Paketo NPM Start Buildpack")
	container.assertHasOutput(t, fmt.Sprintf("Saving %s/not-found:0.0.1", registryURL))
	container.assertHasOutput(t, "*** Images (sha256:")
	container.assertHasOutput(t, "SUCCESS")
	container.terminate(t)
}

func TestNpmCustomBuildpacksBuildpacklessProject(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	registryContainer := setupDockerRegistry(t, ctx)
	defer registryContainer.Terminate(ctx)

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "paketobuildpacks/builder:buildpackless-full",
		User:    "cnb",
		TestDir: []string{"testdata", "TestMtaIntegration", "npm"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
	})

	container.whenRunningPiperCommand("cnbBuild", "--buildpacks", "gcr.io/paketo-buildpacks/nodejs", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", registryURL)

	container.assertHasOutput(t, "Setting custom buildpacks: '[gcr.io/paketo-buildpacks/nodejs]'")
	container.assertHasOutput(t, "Downloading buildpack 'gcr.io/paketo-buildpacks/nodejs' to /tmp/nodejs")
	container.assertHasOutput(t, "running command: /cnb/lifecycle/creator")
	container.assertHasOutput(t, "Paketo NPM Start Buildpack")
	container.assertHasOutput(t, fmt.Sprintf("Saving %s/not-found:0.0.1", registryURL))
	container.assertHasOutput(t, "*** Images (sha256:")
	container.assertHasOutput(t, "SUCCESS")
	container.terminate(t)
}

func TestWrongBuilderProject(t *testing.T) {
	t.Parallel()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "nginx:latest",
		TestDir: []string{"testdata", "TestMtaIntegration", "npm"},
	})

	container.whenRunningPiperCommand("cnbBuild", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", "test")

	container.assertHasOutput(t, "the provided dockerImage is not a valid builder")
	container.terminate(t)
}

func TestBindings(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	registryContainer := setupDockerRegistry(t, ctx)
	defer registryContainer.Terminate(ctx)

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "paketobuildpacks/builder:full",
		User:    "cnb",
		TestDir: []string{"testdata"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
	})

	container.whenRunningPiperCommand("cnbBuild", "--customConfig", "TestCnbIntegration/config.yml", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", registryURL, "--path", "TestMtaIntegration/maven")

	container.assertHasOutput(t, "bindings/maven-settings/settings.xml: only whitespace content allowed before start tag")
	container.assertHasFile(t, "/tmp/platform/bindings/dummy-binding/type")
	container.assertHasFile(t, "/tmp/platform/bindings/dummy-binding/dummy.yml")
	container.terminate(t)
}

func TestMultiImage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	registryContainer := setupDockerRegistry(t, ctx)
	defer registryContainer.Terminate(ctx)

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "paketobuildpacks/builder:full",
		User:    "cnb",
		TestDir: []string{"testdata", "TestCnbIntegration"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
	})

	container.whenRunningPiperCommand("cnbBuild", "--customConfig", "config_multi_image.yml")

	container.assertHasOutput(t, "Previous image with name \"localhost:5000/io-buildpacks-my-app:latest\" not found")
	container.assertHasOutput(t, "Saving localhost:5000/io-buildpacks-my-app:latest...")
	container.assertHasOutput(t, "Previous image with name \"localhost:5000/go-app:v1.0.0\" not found")
	container.assertHasOutput(t, "Saving localhost:5000/go-app:v1.0.0...")
	container.terminate(t)
}

func TestPreserveFiles(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	registryContainer := setupDockerRegistry(t, ctx)
	defer registryContainer.Terminate(ctx)

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "paketobuildpacks/builder:full",
		User:    "cnb",
		TestDir: []string{"testdata", "TestCnbIntegration"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
	})

	container.whenRunningPiperCommand("cnbBuild", "--customConfig", "config_preserve_files.yml")
	container.assertHasFile(t, "/project/project/node_modules/base/README.md")
	container.assertHasFile(t, "/project/project/package-lock.json")
	container.terminate(t)
}

func TestPreserveFilesIgnored(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	registryContainer := setupDockerRegistry(t, ctx)
	defer registryContainer.Terminate(ctx)

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "paketobuildpacks/builder:full",
		User:    "cnb",
		TestDir: []string{"testdata", "TestCnbIntegration"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
	})

	container.whenRunningPiperCommand("cnbBuild", "--customConfig", "config_preserve_files.yml", "--path", "zip/go.zip", "--containerImageName", "go-zip")
	container.assertHasOutput(t, "skipping preserving files because the source")
	container.terminate(t)
}
