//go:build integration
// +build integration

// can be executed with go test -tags=integration ./integration/...

package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
)

const (
	registryURL = "localhost:5000"
	baseBuilder = "paketobuildpacks/builder:0.3.26-base"
)

func setupDockerRegistry(t *testing.T, ctx context.Context) testcontainers.Container {
	defer testTimer("setupDockerRegistry", timeNow())

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
	defer testTimer("TestNpmProject", timeNow())

	t.Parallel()
	ctx := context.Background()
	registryContainer := setupDockerRegistry(t, ctx)
	defer registryContainer.Terminate(ctx)

	wg, _ := errgroup.WithContext(context.TODO())

	wg.Go(func() error {
		container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
			Image:   baseBuilder,
			User:    "cnb",
			TestDir: []string{"testdata"},
			Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
		})
		assert.NoError(t, container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--path", "TestCnbIntegration/project", "--customConfig", "TestCnbIntegration/config.yml", "--containerImageName", "node", "--containerImageTag", "0.0.1", "--containerRegistryUrl", registryURL))
		container.assertHasOutput(t, "running command: /cnb/lifecycle/creator")
		container.assertHasOutput(t, "Selected Node Engine version (using BP_NODE_VERSION): 16")
		container.assertHasOutput(t, "Paketo NPM Start Buildpack")
		container.assertHasOutput(t, fmt.Sprintf("Saving %s/node:0.0.1", registryURL))
		container.assertHasOutput(t, "*** Images (sha256:")
		container.assertHasOutput(t, "SUCCESS")
		container.terminate(t)
		return nil
	})

	wg.Go(func() error {
		container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
			Image:   baseBuilder,
			User:    "cnb",
			TestDir: []string{"testdata"},
			Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
		})
		assert.NoError(t, container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--path", "TestCnbIntegration/project", "--customConfig", "TestCnbIntegration/config.yml", "--containerImageName", "node", "--containerImageTag", "0.0.1", "--containerRegistryUrl", registryURL, "--projectDescriptor", "project-with-id.toml"))
		container.assertHasOutput(t, "running command: /cnb/lifecycle/creator")
		container.assertHasOutput(t, "Selected Node Engine version (using BP_NODE_VERSION): 16")
		container.assertHasOutput(t, "Paketo NPM Start Buildpack")
		container.assertHasOutput(t, fmt.Sprintf("Saving %s/node:0.0.1", registryURL))
		container.assertHasOutput(t, "*** Images (sha256:")
		container.assertHasOutput(t, "SUCCESS")
		container.terminate(t)
		return nil
	})

	wg.Wait()

}

func TestProjectDescriptor(t *testing.T) {
	defer testTimer("TestProjectDescriptor", timeNow())

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

	container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", registryURL)

	container.assertHasOutput(t, "running command: /cnb/lifecycle/creator")
	container.assertHasOutput(t, "Dockerfile doesn't match include pattern, ignoring")
	container.assertHasOutput(t, "srv/hello.js matches include pattern")
	container.assertHasOutput(t, "package.json matches include pattern")
	container.assertHasOutput(t, "Downloading buildpack")
	container.assertHasOutput(t, "Setting custom environment variables: 'map[BP_NODE_VERSION:16 TMPDIR:/tmp/cnbBuild-")
	container.assertHasOutput(t, "Selected Node Engine version (using BP_NODE_VERSION): 16")
	container.assertHasOutput(t, "Paketo NPM Start Buildpack")
	container.assertHasOutput(t, fmt.Sprintf("Saving %s/not-found:0.0.1", registryURL))
	container.assertHasOutput(t, "*** Images (sha256:")
	container.assertHasOutput(t, "SUCCESS")
	container.terminate(t)
}

func TestZipPath(t *testing.T) {
	defer testTimer("TestZipPath", timeNow())

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

	container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", registryURL, "--path", "go.zip")

	container.assertHasOutput(t, "running command: /cnb/lifecycle/creator")
	container.assertHasOutput(t, "Installing Go")
	container.assertHasOutput(t, "Paketo Go Build Buildpack")
	container.assertHasOutput(t, fmt.Sprintf("Saving %s/not-found:0.0.1", registryURL))
	container.assertHasOutput(t, "*** Images (sha256:")
	container.assertHasOutput(t, "SUCCESS")
	container.terminate(t)
}

func TestNonZipPath(t *testing.T) {
	defer testTimer("TestNonZipPath", timeNow())

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

	container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", registryURL, "--path", "mta.yaml")

	container.assertHasOutput(t, "Copying  '/project/mta.yaml' into '/workspace' failed: application path must be a directory or zip")
	container.terminate(t)
}

func TestNpmCustomBuildpacksFullProject(t *testing.T) {
	defer testTimer("TestNpmCustomBuildpacksFullProject", timeNow())

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

	container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--buildpacks", "gcr.io/paketo-buildpacks/nodejs:0.19.0", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", registryURL)

	container.assertHasOutput(t, "Setting custom buildpacks: '[gcr.io/paketo-buildpacks/nodejs:0.19.0]'")
	container.assertHasOutput(t, "Downloading buildpack 'gcr.io/paketo-buildpacks/nodejs:0.19.0' to /tmp/buildpacks_cache/sha256:")
	container.assertHasOutput(t, "running command: /cnb/lifecycle/creator")
	container.assertHasOutput(t, "Paketo NPM Start Buildpack")
	container.assertHasOutput(t, fmt.Sprintf("Saving %s/not-found:0.0.1", registryURL))
	container.assertHasOutput(t, "*** Images (sha256:")
	container.assertHasOutput(t, "SUCCESS")
	container.terminate(t)
}

func TestNpmCustomBuildpacksBuildpacklessProject(t *testing.T) {
	defer testTimer("TestNpmCustomBuildpacksBuildpacklessProject", timeNow())

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

	container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--buildpacks", "gcr.io/paketo-buildpacks/nodejs:0.19.0", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", registryURL)

	container.assertHasOutput(t, "Setting custom buildpacks: '[gcr.io/paketo-buildpacks/nodejs:0.19.0]'")
	container.assertHasOutput(t, "Downloading buildpack 'gcr.io/paketo-buildpacks/nodejs:0.19.0' to /tmp/buildpacks_cache/sha256:")
	container.assertHasOutput(t, "running command: /cnb/lifecycle/creator")
	container.assertHasOutput(t, "Paketo NPM Start Buildpack")
	container.assertHasOutput(t, fmt.Sprintf("Saving %s/not-found:0.0.1", registryURL))
	container.assertHasOutput(t, "*** Images (sha256:")
	container.assertHasOutput(t, "SUCCESS")
	container.terminate(t)
}

func TestWrongBuilderProject(t *testing.T) {
	defer testTimer("TestWrongBuilderProject", timeNow())

	t.Parallel()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "nginx:latest",
		TestDir: []string{"testdata", "TestMtaIntegration", "npm"},
	})

	container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", "test")

	container.assertHasOutput(t, "the provided dockerImage is not a valid builder")
	container.terminate(t)
}

func TestBindings(t *testing.T) {
	defer testTimer("TestBindings", timeNow())

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

	container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--customConfig", "TestCnbIntegration/config.yml", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", registryURL, "--path", "TestMtaIntegration/maven")

	container.assertHasOutput(t, "bindings/maven-settings/settings.xml: only whitespace content allowed before start tag")
	container.assertHasFile(t, "/tmp/platform/bindings/dummy-binding/type")
	container.assertHasFile(t, "/tmp/platform/bindings/dummy-binding/dummy.yml")
	container.terminate(t)
}

func TestMultiImage(t *testing.T) {
	defer testTimer("TestMultiImage", timeNow())

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

	container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--customConfig", "config_multi_image.yml")

	container.assertHasOutput(t, "Previous image with name \"localhost:5000/io-buildpacks-my-app:latest\" not found")
	container.assertHasOutput(t, "Saving localhost:5000/io-buildpacks-my-app:latest...")
	container.assertHasOutput(t, "Previous image with name \"localhost:5000/go-app:v1.0.0\" not found")
	container.assertHasOutput(t, "Saving localhost:5000/go-app:v1.0.0...")
	container.assertHasOutput(t, "Using cached buildpack")
	container.assertHasOutput(t, "Saving localhost:5000/my-app2:latest...")
	container.terminate(t)
}

func TestPreserveFiles(t *testing.T) {
	defer testTimer("TestPreserveFiles", timeNow())

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

	container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--customConfig", "config_preserve_files.yml")
	container.assertHasFile(t, "/project/project/node_modules/base/README.md")
	container.assertHasFile(t, "/project/project/package-lock.json")
	container.terminate(t)
}

func TestPreserveFilesIgnored(t *testing.T) {
	defer testTimer("TestPreserveFilesIgnored", timeNow())

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

	container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--customConfig", "config_preserve_files.yml", "--path", "zip/go.zip", "--containerImageName", "go-zip")
	container.assertHasOutput(t, "skipping preserving files because the source")
	container.terminate(t)
}
