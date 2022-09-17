//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestCNBIntegration ./integration/...

package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
)

const (
	registryURL = "localhost:5000"
	baseBuilder = "paketobuildpacks/builder:0.3.26-base"
)

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

func TestCNBIntegrationNPMProject(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	registryContainer := setupDockerRegistry(t, ctx)
	defer registryContainer.Terminate(ctx)

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   baseBuilder,
		User:    "cnb",
		TestDir: []string{"testdata"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
	})
	defer container.terminate(t)

	container2 := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   baseBuilder,
		User:    "cnb",
		TestDir: []string{"testdata"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
	})
	defer container2.terminate(t)

	err := container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--path", "TestCnbIntegration/project", "--customConfig", "TestCnbIntegration/config.yml", "--containerImageName", "node", "--containerImageTag", "0.0.1", "--containerRegistryUrl", registryURL)
	assert.NoError(t, err)
	container.assertHasOutput(t, "running command: /cnb/lifecycle/creator")
	container.assertHasOutput(t, "Selected Node Engine version (using BP_NODE_VERSION): 16")
	container.assertHasOutput(t, "Paketo NPM Start Buildpack")
	container.assertHasOutput(t, fmt.Sprintf("Saving %s/node:0.0.1", registryURL))
	container.assertHasOutput(t, "*** Images (sha256:")
	container.assertHasOutput(t, "SUCCESS")
	container.terminate(t)

	err = container2.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--path", "TestCnbIntegration/project", "--customConfig", "TestCnbIntegration/config.yml", "--containerImageName", "node", "--containerImageTag", "0.0.1", "--containerRegistryUrl", registryURL, "--projectDescriptor", "project-with-id.toml")
	assert.NoError(t, err)
	container2.assertHasOutput(t, "running command: /cnb/lifecycle/creator")
	container2.assertHasOutput(t, "Selected Node Engine version (using BP_NODE_VERSION): 16")
	container2.assertHasOutput(t, "Paketo NPM Start Buildpack")
	container2.assertHasOutput(t, fmt.Sprintf("Saving %s/node:0.0.1", registryURL))
	container2.assertHasOutput(t, "*** Images (sha256:")
	container2.assertHasOutput(t, "SUCCESS")
	container2.terminate(t)
}

func TestCNBIntegrationProjectDescriptor(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	registryContainer := setupDockerRegistry(t, ctx)
	defer registryContainer.Terminate(ctx)

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   baseBuilder,
		User:    "cnb",
		TestDir: []string{"testdata", "TestCnbIntegration", "project"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
	})
	defer container.terminate(t)

	container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", registryURL)

	container.assertHasOutput(t, "running command: /cnb/lifecycle/creator",
		"Dockerfile doesn't match include pattern, ignoring",
		"srv/hello.js matches include pattern",
		"package.json matches include pattern",
		"Downloading buildpack",
		"Setting custom environment variables: 'map[BP_NODE_VERSION:16 TMPDIR:/tmp/cnbBuild-",
		"Selected Node Engine version (using BP_NODE_VERSION): 16",
		"Paketo NPM Start Buildpack",
		fmt.Sprintf("Saving %s/not-found:0.0.1", registryURL),
		"*** Images (sha256:",
		"SUCCESS",
	)
}

func TestCNBIntegrationZipPath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	registryContainer := setupDockerRegistry(t, ctx)
	defer registryContainer.Terminate(ctx)

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   baseBuilder,
		User:    "cnb",
		TestDir: []string{"testdata", "TestCnbIntegration", "zip"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
	})
	defer container.terminate(t)

	container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", registryURL, "--path", "go.zip")

	container.assertHasOutput(t,
		"running command: /cnb/lifecycle/creator",
		"Installing Go",
		"Paketo Go Build Buildpack",
		fmt.Sprintf("Saving %s/not-found:0.0.1", registryURL),
		"*** Images (sha256:",
		"SUCCESS",
	)
}

func TestCNBIntegrationNonZipPath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	registryContainer := setupDockerRegistry(t, ctx)
	defer registryContainer.Terminate(ctx)

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   baseBuilder,
		User:    "cnb",
		TestDir: []string{"testdata", "TestMtaIntegration", "npm"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
	})
	defer container.terminate(t)

	container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", registryURL, "--path", "mta.yaml")

	container.assertHasOutput(t, "Copying  '/project/mta.yaml' into '/workspace' failed: application path must be a directory or zip")
}

func TestCNBIntegrationNPMCustomBuildpacksFullProject(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	registryContainer := setupDockerRegistry(t, ctx)
	defer registryContainer.Terminate(ctx)

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   baseBuilder,
		User:    "cnb",
		TestDir: []string{"testdata", "TestMtaIntegration", "npm"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
	})
	defer container.terminate(t)

	container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--buildpacks", "gcr.io/paketo-buildpacks/nodejs:0.19.0", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", registryURL)

	container.assertHasOutput(t,
		"Setting custom buildpacks: '[gcr.io/paketo-buildpacks/nodejs:0.19.0]'",
		"Downloading buildpack 'gcr.io/paketo-buildpacks/nodejs:0.19.0' to /tmp/buildpacks_cache/sha256:",
		"running command: /cnb/lifecycle/creator",
		"Paketo NPM Start Buildpack",
		fmt.Sprintf("Saving %s/not-found:0.0.1", registryURL),
		"*** Images (sha256:",
		"SUCCESS",
	)
}

func TestCNBIntegrationNPMCustomBuildpacksBuildpacklessProject(t *testing.T) {
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
	defer container.terminate(t)

	container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--buildpacks", "gcr.io/paketo-buildpacks/nodejs:0.19.0", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", registryURL)

	container.assertHasOutput(t, "Setting custom buildpacks: '[gcr.io/paketo-buildpacks/nodejs:0.19.0]'",
		"Downloading buildpack 'gcr.io/paketo-buildpacks/nodejs:0.19.0' to /tmp/buildpacks_cache/sha256:",
		"running command: /cnb/lifecycle/creator",
		"Paketo NPM Start Buildpack",
		fmt.Sprintf("Saving %s/not-found:0.0.1", registryURL),
		"*** Images (sha256:",
		"SUCCESS",
	)
}

func TestCNBIntegrationWrongBuilderProject(t *testing.T) {
	t.Parallel()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "nginx:latest",
		TestDir: []string{"testdata", "TestMtaIntegration", "npm"},
	})
	defer container.terminate(t)

	container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", "test")

	container.assertHasOutput(t, "the provided dockerImage is not a valid builder")
}

func TestCNBIntegrationBindings(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	registryContainer := setupDockerRegistry(t, ctx)
	defer registryContainer.Terminate(ctx)

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   baseBuilder,
		User:    "cnb",
		TestDir: []string{"testdata"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
	})
	defer container.terminate(t)

	container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--customConfig", "TestCnbIntegration/config.yml", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", registryURL, "--path", "TestMtaIntegration/maven")

	container.assertHasOutput(t, "bindings/maven-settings/settings.xml: only whitespace content allowed before start tag")
	container.assertHasFiles(t,
		"/tmp/platform/bindings/dummy-binding/type",
		"/tmp/platform/bindings/dummy-binding/dummy.yml",
	)
}

func TestCNBIntegrationMultiImage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	registryContainer := setupDockerRegistry(t, ctx)
	defer registryContainer.Terminate(ctx)

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   baseBuilder,
		User:    "cnb",
		TestDir: []string{"testdata", "TestCnbIntegration"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
	})
	defer container.terminate(t)

	container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--customConfig", "config_multi_image.yml")

	container.assertHasOutput(t,
		"Previous image with name \"localhost:5000/io-buildpacks-my-app:latest\" not found",
		"Saving localhost:5000/io-buildpacks-my-app:latest...",
		"Previous image with name \"localhost:5000/go-app:v1.0.0\" not found",
		"Saving localhost:5000/go-app:v1.0.0...",
		"Using cached buildpack",
		"Saving localhost:5000/my-app2:latest...",
	)
}

func TestCNBIntegrationPreserveFiles(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	registryContainer := setupDockerRegistry(t, ctx)
	defer registryContainer.Terminate(ctx)

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   baseBuilder,
		User:    "cnb",
		TestDir: []string{"testdata", "TestCnbIntegration"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
	})
	defer container.terminate(t)

	container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--customConfig", "config_preserve_files.yml")
	container.assertHasFiles(t, "/project/project/node_modules/base/README.md", "/project/project/package-lock.json")
}

func TestCNBIntegrationPreserveFilesIgnored(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	registryContainer := setupDockerRegistry(t, ctx)
	defer registryContainer.Terminate(ctx)

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   baseBuilder,
		User:    "cnb",
		TestDir: []string{"testdata", "TestCnbIntegration"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
	})
	defer container.terminate(t)

	container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--customConfig", "config_preserve_files.yml", "--path", "zip/go.zip", "--containerImageName", "go-zip")
	container.assertHasOutput(t, "skipping preserving files because the source")
}
