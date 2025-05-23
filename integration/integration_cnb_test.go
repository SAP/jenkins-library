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
	baseBuilder = "paketobuildpacks/builder-jammy-base:0.4.252"
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
	// t.Parallel()
	ctx := context.Background()
	registryContainer := setupDockerRegistry(t, ctx)
	defer registryContainer.Terminate(ctx)

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   baseBuilder,
		User:    "0",
		TestDir: []string{"testdata", "TestCnbIntegration"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
		Environment: map[string]string{
			"PIPER_VAULTCREDENTIAL_DYNATRACE_API_KEY": "api-key-content",
		},
	})

	container2 := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   baseBuilder,
		User:    "0",
		TestDir: []string{"testdata", "TestCnbIntegration"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
		Environment: map[string]string{
			"PIPER_VAULTCREDENTIAL_DYNATRACE_API_KEY": "api-key-content",
		},
	})

	err := container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--path", "project", "--customConfig", "config.yml", "--containerImageName", "node", "--containerImageTag", "0.0.1", "--dockerConfigJSON", "config.json", "--containerRegistryUrl", fmt.Sprintf("http://%s", registryURL), "--defaultProcess", "greeter")
	assert.NoError(t, err)
	container.assertHasOutput(t, "running command: /cnb/lifecycle/creator")
	container.assertHasOutput(t, "Selected Node Engine version (using BP_NODE_VERSION): 16")
	container.assertHasOutput(t, "Paketo Buildpack for NPM Start")
	container.assertHasOutput(t, fmt.Sprintf("Saving %s/node:0.0.1", registryURL))
	container.assertHasOutput(t, "Setting default process type 'greeter'")
	container.assertHasOutput(t, "*** Images (sha256:")
	container.assertHasOutput(t, "SUCCESS")
	container.assertFileContentEquals(t, "/tmp/config.json", "{\n\t\"auths\": {\n\t\t\"test.registry.io\": {},\n\t\t\"test2.registry.io\": {}\n\t}\n}")
	container.terminate(t)

	err = container2.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--path", "project", "--customConfig", "config.yml", "--containerImageName", "node", "--containerImageTag", "0.0.1", "--containerRegistryUrl", registryURL, "--projectDescriptor", "project-with-id.toml")
	assert.NoError(t, err)
	container2.assertHasOutput(t, "running command: /cnb/lifecycle/creator")
	container2.assertHasOutput(t, "Selected Node Engine version (using BP_NODE_VERSION): 16")
	container2.assertHasOutput(t, "Paketo Buildpack for NPM Start")
	container2.assertHasOutput(t, fmt.Sprintf("Saving %s/node:0.0.1", registryURL))
	container2.assertHasOutput(t, "*** Images (sha256:")
	container2.assertHasOutput(t, "SUCCESS")
	container2.assertFileContentEquals(t, "/tmp/config.json", "{\n\t\"auths\": {\n\t\t\"test2.registry.io\": {}\n\t}\n}")
	container2.terminate(t)
}

func TestCNBIntegrationProjectDescriptor(t *testing.T) {
	// t.Parallel()
	ctx := context.Background()
	registryContainer := setupDockerRegistry(t, ctx)
	defer registryContainer.Terminate(ctx)

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   baseBuilder,
		User:    "0",
		TestDir: []string{"testdata", "TestCnbIntegration", "project"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
	})

	err := container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", registryURL)
	assert.NoError(t, err)

	container.assertHasOutput(t, "running command: /cnb/lifecycle/creator",
		"Dockerfile doesn't match include pattern, ignoring",
		"srv/hello.js matches include pattern",
		"package.json matches include pattern",
		"Downloading buildpack",
		"Setting custom environment variables: 'map[BP_NODE_VERSION:16 TMPDIR:/tmp/cnbBuild-",
		"Selected Node Engine version (using BP_NODE_VERSION): 16",
		"Paketo Buildpack for NPM Start",
		fmt.Sprintf("Saving %s/not-found:0.0.1", registryURL),
		"*** Images (sha256:",
		"SUCCESS",
	)
	container.terminate(t)
}
func TestCNBIntegrationBuildSummary(t *testing.T) {
	// t.Parallel()
	ctx := context.Background()
	registryContainer := setupDockerRegistry(t, ctx)
	defer registryContainer.Terminate(ctx)

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   baseBuilder,
		User:    "0",
		TestDir: []string{"testdata", "TestCnbIntegration", "project"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
	})

	err := container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", registryURL)
	assert.NoError(t, err)

	container.assertHasOutput(t, "*** Build Summary ***",
		"  Builder: \"paketobuildpacks/builder:base\"",
		"  Lifecycle: \"0.16.4+683e1b46\"",
		"  Image: \"localhost:5000/not-found@sha256:",
		"    Project descriptor: \"/project/project.toml\"",
		"    Env: \"TMPDIR, BP_NODE_VERSION\"")
	container.terminate(t)
}

func TestCNBIntegrationZipPath(t *testing.T) {
	// t.Parallel()
	ctx := context.Background()
	registryContainer := setupDockerRegistry(t, ctx)
	defer registryContainer.Terminate(ctx)

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   baseBuilder,
		User:    "0",
		TestDir: []string{"testdata", "TestCnbIntegration", "zip"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
	})

	err := container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", registryURL, "--path", "go.zip", "--createBOM")
	assert.NoError(t, err)

	container.assertHasOutput(t,
		"running command: /cnb/lifecycle/creator",
		"Installing Go",
		"Paketo Buildpack for Go Build",
		fmt.Sprintf("Saving %s/not-found:0.0.1", registryURL),
		"*** Images (sha256:",
		"SUCCESS",
		"syft scan registry:localhost:5000/not-found:0.0.1 -o cyclonedx-xml@1.4=bom-docker-0.xml -q",
	)
	container.assertHasFiles(t, "/project/bom-docker-0.xml")
	container.terminate(t)
}

func TestCNBIntegrationNonZipPath(t *testing.T) {
	// t.Parallel()
	ctx := context.Background()
	registryContainer := setupDockerRegistry(t, ctx)
	defer registryContainer.Terminate(ctx)

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   baseBuilder,
		User:    "0",
		TestDir: []string{"testdata", "TestMtaIntegration", "npm"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
	})

	err := container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", registryURL, "--path", "mta.yaml")
	assert.Error(t, err)

	container.assertHasOutput(t, "Copying  '/project/mta.yaml' into '/workspace' failed: application path must be a directory or zip")
	container.terminate(t)
}

func TestCNBIntegrationNPMCustomBuildpacksFullProject(t *testing.T) {
	// t.Parallel()
	ctx := context.Background()
	registryContainer := setupDockerRegistry(t, ctx)
	defer registryContainer.Terminate(ctx)

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   baseBuilder,
		User:    "0",
		TestDir: []string{"testdata", "TestMtaIntegration", "npm"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
	})

	err := container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--buildpacks", "docker.io/paketobuildpacks/nodejs:2.0.0", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", registryURL)
	assert.NoError(t, err)

	container.assertHasOutput(t,
		"Setting custom buildpacks: '[docker.io/paketobuildpacks/nodejs:2.0.0]'",
		"Downloading buildpack 'docker.io/paketobuildpacks/nodejs:2.0.0' to /tmp/buildpacks_cache/sha256:",
		"running command: /cnb/lifecycle/creator",
		"Paketo Buildpack for NPM Start",
		fmt.Sprintf("Saving %s/not-found:0.0.1", registryURL),
		"*** Images (sha256:",
		"SUCCESS",
	)
	container.terminate(t)
}

func TestCNBIntegrationNPMCustomBuildpacksBuildpacklessProject(t *testing.T) {
	// t.Parallel()
	ctx := context.Background()
	registryContainer := setupDockerRegistry(t, ctx)
	defer registryContainer.Terminate(ctx)

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "paketobuildpacks/builder-jammy-buildpackless-full",
		User:    "0",
		TestDir: []string{"testdata", "TestMtaIntegration", "npm"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
	})

	err := container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--buildpacks", "docker.io/paketobuildpacks/nodejs:2.0.0", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", registryURL)
	assert.NoError(t, err)

	container.assertHasOutput(t, "Setting custom buildpacks: '[docker.io/paketobuildpacks/nodejs:2.0.0]'",
		"Downloading buildpack 'docker.io/paketobuildpacks/nodejs:2.0.0' to /tmp/buildpacks_cache/sha256:",
		"running command: /cnb/lifecycle/creator",
		"Paketo Buildpack for NPM Start",
		fmt.Sprintf("Saving %s/not-found:0.0.1", registryURL),
		"*** Images (sha256:",
		"SUCCESS",
	)
	container.terminate(t)
}

func TestCNBIntegrationWrongBuilderProject(t *testing.T) {
	// t.Parallel()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "nginx:latest",
		TestDir: []string{"testdata", "TestMtaIntegration", "npm"},
	})

	err := container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", "test")
	assert.Error(t, err)

	container.assertHasOutput(t, "the provided dockerImage is not a valid builder")
	container.terminate(t)
}

func TestCNBIntegrationBindings(t *testing.T) {
	// t.Parallel()
	ctx := context.Background()
	registryContainer := setupDockerRegistry(t, ctx)
	defer registryContainer.Terminate(ctx)

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   baseBuilder,
		User:    "0",
		TestDir: []string{"testdata", "TestCnbIntegration"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
		Environment: map[string]string{
			"PIPER_VAULTCREDENTIAL_DYNATRACE_API_KEY": "api-key-content",
		},
	})

	err := container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--customConfig", "config.yml", "--containerImageName", "not-found", "--containerImageTag", "0.0.1", "--containerRegistryUrl", registryURL, "--path", "project")
	assert.NoError(t, err)

	container.assertHasFiles(t,
		"/tmp/platform/bindings/dummy-binding/type",
		"/tmp/platform/bindings/dummy-binding/dummy.yml",
	)
	container.assertFileContentEquals(t, "/tmp/platform/bindings/maven-settings/settings.xml", "invalid xml")
	container.assertFileContentEquals(t, "/tmp/platform/bindings/dynatrace/api-key", "api-key-content")
	container.terminate(t)
}

func TestCNBIntegrationMultiImage(t *testing.T) {
	// t.Parallel()
	ctx := context.Background()
	registryContainer := setupDockerRegistry(t, ctx)
	defer registryContainer.Terminate(ctx)

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   baseBuilder,
		User:    "0",
		TestDir: []string{"testdata", "TestCnbIntegration"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
	})

	err := container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--customConfig", "config_multi_image.yml", "--createBOM")
	assert.NoError(t, err)

	container.assertHasOutput(t,
		"Image with name \"localhost:5000/io-buildpacks-my-app:latest\" not found",
		"Saving localhost:5000/io-buildpacks-my-app:latest...",
		"Image with name \"localhost:5000/go-app:v1.0.0\" not found",
		"Saving localhost:5000/go-app:v1.0.0...",
		"Using cached buildpack",
		"Saving localhost:5000/my-app2:latest...",
		"syft scan registry:localhost:5000/io-buildpacks-my-app:latest -o cyclonedx-xml@1.4=bom-docker-0.xml -q",
		"syft scan registry:localhost:5000/go-app:v1.0.0 -o cyclonedx-xml@1.4=bom-docker-1.xml -q",
		"syft scan registry:localhost:5000/my-app2:latest -o cyclonedx-xml@1.4=bom-docker-2.xml -q",
	)

	container.assertHasFiles(t, "/project/bom-docker-0.xml")
	container.assertHasFiles(t, "/project/bom-docker-1.xml")
	container.assertHasFiles(t, "/project/bom-docker-2.xml")
	container.terminate(t)
}

func TestCNBIntegrationPreserveFiles(t *testing.T) {
	// t.Parallel()
	ctx := context.Background()
	registryContainer := setupDockerRegistry(t, ctx)
	defer registryContainer.Terminate(ctx)

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   baseBuilder,
		User:    "0",
		TestDir: []string{"testdata", "TestCnbIntegration"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
	})

	err := container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--customConfig", "config_preserve_files.yml")
	assert.NoError(t, err)

	container.assertHasFiles(t, "/project/project/node_modules/base/README.md", "/project/project/package-lock.json")
	container.terminate(t)
}

func TestCNBIntegrationPreserveFilesIgnored(t *testing.T) {
	// t.Parallel()
	ctx := context.Background()
	registryContainer := setupDockerRegistry(t, ctx)
	defer registryContainer.Terminate(ctx)

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   baseBuilder,
		User:    "0",
		TestDir: []string{"testdata", "TestCnbIntegration"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
	})

	err := container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--customConfig", "config_preserve_files.yml", "--path", "zip/go.zip", "--containerImageName", "go-zip")
	assert.NoError(t, err)
	container.assertHasOutput(t, "skipping preserving files because the source")
	container.terminate(t)
}

func TestCNBIntegrationPrePostBuildpacks(t *testing.T) {
	// t.Parallel()
	ctx := context.Background()
	registryContainer := setupDockerRegistry(t, ctx)
	defer registryContainer.Terminate(ctx)

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   baseBuilder,
		User:    "0",
		TestDir: []string{"testdata", "TestCnbIntegration"},
		Network: fmt.Sprintf("container:%s", registryContainer.GetContainerID()),
		Environment: map[string]string{
			"PIPER_VAULTCREDENTIAL_DYNATRACE_API_KEY": "api-key-content",
		},
	})

	err := container.whenRunningPiperCommand("cnbBuild", "--noTelemetry", "--verbose", "--projectDescriptor", "", "--path", "project", "--customConfig", "config.yml", "--containerImageTag", "0.0.1", "--containerImageName", "not-found", "--containerRegistryUrl", registryURL, "--postBuildpacks", "paketobuildpacks/datadog")
	assert.NoError(t, err)
	container.assertHasOutput(t, "Setting custom buildpacks: '[]'")
	container.assertHasOutput(t, "Pre-buildpacks: '[]'")
	container.assertHasOutput(t, "Post-buildpacks: '[paketobuildpacks/datadog]'")
	container.terminate(t)
}
