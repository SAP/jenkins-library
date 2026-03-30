//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestDockerBuildIntegration ./integration/...

package main

import (
	"fmt"
	"strings"
	"testing"
)

// TestDockerBuildIntegrationSmoke validates the DinD test infrastructure works.
func TestDockerBuildIntegrationSmoke(t *testing.T) {
	dind := StartDindContainer(t, DindContainerConfig{
		TestData: "TestDockerBuildIntegration/single-image",
	})

	// Verify Docker works inside DinD
	ExecCommand(t, dind.DindContainer, "/", []string{"docker", "info"})

	// Verify piper binary is available
	ExecCommand(t, dind.DindContainer, "/", []string{"/piperbin/piper", "version"})

	// Verify test data was copied
	ExecCommand(t, dind.DindContainer, "/", []string{"stat", "/single-image/Dockerfile"})
}

// --- T2: Core test cases ---

func TestDockerBuildIntegrationSingleImageWithNameAndTag(t *testing.T) {
	assert := NewContainerAssert(t)

	dind := StartDindContainer(t, DindContainerConfig{
		TestData: "TestDockerBuildIntegration/single-image",
	})

	output := RunPiper(t, dind.DindContainer, "/single-image",
		"dockerBuild",
		"--containerImageName=test-image",
		"--containerImageTag=1.0.0",
		fmt.Sprintf("--containerRegistryUrl=http://%s", dind.RegistryHost),
		"--noTelemetry",
		"--verbose",
	)

	assert.Contains(output, "docker buildx build")
	assert.Contains(output, "--push")
	assert.Contains(output, "SUCCESS")

	// Verify image was pushed to registry
	ExecCommand(t, dind.DindContainer, "/", []string{
		"docker", "pull", fmt.Sprintf("%s/test-image:1.0.0", dind.RegistryHost),
	})

	// Verify CPE output
	registryURL := strings.TrimSpace(string(ReadFile(t, dind.DindContainer,
		"/single-image/.pipeline/commonPipelineEnvironment/container/registryUrl")))
	assert.Assertions.Contains(registryURL, dind.RegistryHost)

	imageNameTag := strings.TrimSpace(string(ReadFile(t, dind.DindContainer,
		"/single-image/.pipeline/commonPipelineEnvironment/container/imageNameTag")))
	assert.Assertions.Equal("test-image:1.0.0", imageNameTag)
}

func TestDockerBuildIntegrationSingleImageWithFullReference(t *testing.T) {
	assert := NewContainerAssert(t)

	dind := StartDindContainer(t, DindContainerConfig{
		TestData: "TestDockerBuildIntegration/single-image",
	})

	output := RunPiper(t, dind.DindContainer, "/single-image",
		"dockerBuild",
		fmt.Sprintf("--containerImage=%s/full-ref-image:v2", dind.RegistryHost),
		"--noTelemetry",
		"--verbose",
	)

	assert.Contains(output, "--push")
	assert.Contains(output, "SUCCESS")

	// Verify image was pushed
	ExecCommand(t, dind.DindContainer, "/", []string{
		"docker", "pull", fmt.Sprintf("%s/full-ref-image:v2", dind.RegistryHost),
	})

	// Verify CPE
	imageNameTag := strings.TrimSpace(string(ReadFile(t, dind.DindContainer,
		"/single-image/.pipeline/commonPipelineEnvironment/container/imageNameTag")))
	assert.Assertions.Equal("full-ref-image:v2", imageNameTag)
}

func TestDockerBuildIntegrationSingleImageWithBuildOptionTag(t *testing.T) {
	assert := NewContainerAssert(t)

	dind := StartDindContainer(t, DindContainerConfig{
		TestData: "TestDockerBuildIntegration/single-image",
	})

	output := RunPiper(t, dind.DindContainer, "/single-image",
		"dockerBuild",
		fmt.Sprintf("--buildOptions=-t,%s/opt-tag-image:3.0.0", dind.RegistryHost),
		"--noTelemetry",
		"--verbose",
	)

	assert.Contains(output, "SUCCESS")

	// Verify image was pushed
	ExecCommand(t, dind.DindContainer, "/", []string{
		"docker", "pull", fmt.Sprintf("%s/opt-tag-image:3.0.0", dind.RegistryHost),
	})
}

func TestDockerBuildIntegrationNoPush(t *testing.T) {
	assert := NewContainerAssert(t)

	dind := StartDindContainer(t, DindContainerConfig{
		TestData: "TestDockerBuildIntegration/single-image",
	})

	output := RunPiper(t, dind.DindContainer, "/single-image",
		"dockerBuild",
		"--noTelemetry",
		"--verbose",
	)

	assert.Contains(output, "docker buildx build")
	assert.NotContains(output, "--push")
	assert.Contains(output, "SUCCESS")
}

func TestDockerBuildIntegrationMultiImageBuild(t *testing.T) {
	assert := NewContainerAssert(t)

	dind := StartDindContainer(t, DindContainerConfig{
		TestData: "TestDockerBuildIntegration/multi-image",
	})

	output := RunPiper(t, dind.DindContainer, "/multi-image",
		"dockerBuild",
		"--containerMultiImageBuild=true",
		"--containerImageName=multi-test",
		"--containerImageTag=1.0.0",
		fmt.Sprintf("--containerRegistryUrl=http://%s", dind.RegistryHost),
		"--noTelemetry",
		"--verbose",
	)

	assert.Contains(output, "SUCCESS")

	// Verify all three images were pushed (root, sub1, sub2)
	ExecCommand(t, dind.DindContainer, "/", []string{
		"docker", "pull", fmt.Sprintf("%s/multi-test:1.0.0", dind.RegistryHost),
	})
	ExecCommand(t, dind.DindContainer, "/", []string{
		"docker", "pull", fmt.Sprintf("%s/multi-test-sub1:1.0.0", dind.RegistryHost),
	})
	ExecCommand(t, dind.DindContainer, "/", []string{
		"docker", "pull", fmt.Sprintf("%s/multi-test-sub2:1.0.0", dind.RegistryHost),
	})

	// Verify CPE imageNameTag (root image)
	imageNameTag := strings.TrimSpace(string(ReadFile(t, dind.DindContainer,
		"/multi-image/.pipeline/commonPipelineEnvironment/container/imageNameTag")))
	assert.Assertions.Equal("multi-test:1.0.0", imageNameTag)
}

func TestDockerBuildIntegrationMultipleImagesExplicit(t *testing.T) {
	// Skip: the multipleImages parameter is of type []map[string]interface{} which
	// cannot be passed via CLI flags and the piper config framework deserializes
	// YAML lists as []interface{} instead of the expected type. This feature is
	// fully covered by unit tests in cmd/dockerBuild_test.go.
	t.Skip("multipleImages parameter cannot be set via CLI or YAML config in integration tests")
}

func TestDockerBuildIntegrationReadImageDigest(t *testing.T) {
	assert := NewContainerAssert(t)

	dind := StartDindContainer(t, DindContainerConfig{
		TestData: "TestDockerBuildIntegration/single-image",
	})

	output := RunPiper(t, dind.DindContainer, "/single-image",
		"dockerBuild",
		"--containerImageName=digest-test",
		"--containerImageTag=1.0.0",
		fmt.Sprintf("--containerRegistryUrl=http://%s", dind.RegistryHost),
		"--readImageDigest=true",
		"--noTelemetry",
		"--verbose",
	)

	assert.Contains(output, "--metadata-file")
	assert.Contains(output, "SUCCESS")

	// Verify CPE imageDigest contains sha256
	imageDigest := strings.TrimSpace(string(ReadFile(t, dind.DindContainer,
		"/single-image/.pipeline/commonPipelineEnvironment/container/imageDigest")))
	assert.Assertions.True(strings.HasPrefix(imageDigest, "sha256:"),
		"Expected imageDigest to start with sha256:, got: %s", imageDigest)
}

func TestDockerBuildIntegrationBuildArgsPassing(t *testing.T) {
	assert := NewContainerAssert(t)

	dind := StartDindContainer(t, DindContainerConfig{
		TestData: "TestDockerBuildIntegration/build-args",
	})

	output := RunPiper(t, dind.DindContainer, "/build-args",
		"dockerBuild",
		fmt.Sprintf("--containerImage=%s/args-test:1.0.0", dind.RegistryHost),
		"--buildOptions=--build-arg=MY_ARG=customvalue",
		"--noTelemetry",
		"--verbose",
	)

	assert.Contains(output, "--build-arg=MY_ARG=customvalue")
	assert.Contains(output, "SUCCESS")

	// Verify the build arg was applied by running the built image
	argOutput := ExecCommand(t, dind.DindContainer, "/", []string{
		"docker", "run", "--rm", fmt.Sprintf("%s/args-test:1.0.0", dind.RegistryHost), "cat", "/arg.txt",
	})
	assert.Contains(argOutput, "customvalue")
}

func TestDockerBuildIntegrationVerboseMode(t *testing.T) {
	assert := NewContainerAssert(t)

	dind := StartDindContainer(t, DindContainerConfig{
		TestData: "TestDockerBuildIntegration/single-image",
	})

	output := RunPiper(t, dind.DindContainer, "/single-image",
		"dockerBuild",
		fmt.Sprintf("--containerImage=%s/verbose-test:1.0.0", dind.RegistryHost),
		"--noTelemetry",
		"--verbose",
	)

	assert.Contains(output, "--progress=plain")
	assert.Contains(output, "SUCCESS")
}

func TestDockerBuildIntegrationDeprecatedContainerBuildOptions(t *testing.T) {
	assert := NewContainerAssert(t)

	dind := StartDindContainer(t, DindContainerConfig{
		TestData: "TestDockerBuildIntegration/single-image",
	})

	output := RunPiper(t, dind.DindContainer, "/single-image",
		"dockerBuild",
		fmt.Sprintf("--containerImage=%s/compat-test:1.0.0", dind.RegistryHost),
		"--containerBuildOptions=--label test=true",
		"--noTelemetry",
		"--verbose",
	)

	assert.Contains(output, "containerBuildOptions")
	assert.Contains(output, "--label")
	assert.Contains(output, "SUCCESS")

	// Verify image was pushed
	ExecCommand(t, dind.DindContainer, "/", []string{
		"docker", "pull", fmt.Sprintf("%s/compat-test:1.0.0", dind.RegistryHost),
	})
}

func TestDockerBuildIntegrationCustomDockerfilePath(t *testing.T) {
	assert := NewContainerAssert(t)

	dind := StartDindContainer(t, DindContainerConfig{
		TestData: "TestDockerBuildIntegration/custom-dockerfile",
	})

	output := RunPiper(t, dind.DindContainer, "/custom-dockerfile",
		"dockerBuild",
		fmt.Sprintf("--containerImage=%s/custom-df:1.0.0", dind.RegistryHost),
		"--dockerfilePath=docker/MyDockerfile",
		"--noTelemetry",
		"--verbose",
	)

	assert.Contains(output, "--file")
	assert.Contains(output, "docker/MyDockerfile")
	assert.Contains(output, "SUCCESS")

	// Verify image was pushed
	ExecCommand(t, dind.DindContainer, "/", []string{
		"docker", "pull", fmt.Sprintf("%s/custom-df:1.0.0", dind.RegistryHost),
	})
}

func TestDockerBuildIntegrationPlusSignInTag(t *testing.T) {
	assert := NewContainerAssert(t)

	dind := StartDindContainer(t, DindContainerConfig{
		TestData: "TestDockerBuildIntegration/single-image",
	})

	output := RunPiper(t, dind.DindContainer, "/single-image",
		"dockerBuild",
		"--containerImageName=plustag-test",
		"--containerImageTag=1.0.0-rc+build123",
		fmt.Sprintf("--containerRegistryUrl=http://%s", dind.RegistryHost),
		"--noTelemetry",
		"--verbose",
	)

	assert.Contains(output, "SUCCESS")

	// Plus sign should be replaced with dash
	ExecCommand(t, dind.DindContainer, "/", []string{
		"docker", "pull", fmt.Sprintf("%s/plustag-test:1.0.0-rc-build123", dind.RegistryHost),
	})

	// Verify CPE has sanitized tag
	imageNameTag := strings.TrimSpace(string(ReadFile(t, dind.DindContainer,
		"/single-image/.pipeline/commonPipelineEnvironment/container/imageNameTag")))
	assert.Assertions.Equal("plustag-test:1.0.0-rc-build123", imageNameTag)
}

// --- T3: Edge cases and error scenarios ---

func TestDockerBuildIntegrationErrorNoDockerfile(t *testing.T) {
	assert := NewContainerAssert(t)

	dind := StartDindContainer(t, DindContainerConfig{
		TestData: "TestDockerBuildIntegration/no-dockerfile",
	})

	_, output := RunPiperExpectFailure(t, dind.DindContainer, "/no-dockerfile",
		"dockerBuild",
		fmt.Sprintf("--containerImage=%s/nodf:1.0.0", dind.RegistryHost),
		"--noTelemetry",
		"--verbose",
	)

	assert.Contains(output, "failed")
}

func TestDockerBuildIntegrationErrorMultiImageAllExcluded(t *testing.T) {
	assert := NewContainerAssert(t)

	dind := StartDindContainer(t, DindContainerConfig{
		TestData: "TestDockerBuildIntegration/multi-image",
	})

	_, output := RunPiperExpectFailure(t, dind.DindContainer, "/multi-image",
		"dockerBuild",
		"--containerMultiImageBuild=true",
		"--containerImageName=excl-test",
		"--containerImageTag=1.0.0",
		fmt.Sprintf("--containerRegistryUrl=http://%s", dind.RegistryHost),
		"--containerMultiImageBuildExcludes=Dockerfile,sub1/Dockerfile,sub2/Dockerfile",
		"--noTelemetry",
		"--verbose",
	)

	assert.Contains(output, "no docker files")
}

func TestDockerBuildIntegrationErrorPushToUnreachableRegistry(t *testing.T) {
	assert := NewContainerAssert(t)

	dind := StartDindContainer(t, DindContainerConfig{
		TestData: "TestDockerBuildIntegration/single-image",
	})

	_, output := RunPiperExpectFailure(t, dind.DindContainer, "/single-image",
		"dockerBuild",
		"--containerImage=unreachable.registry.local:9999/test:1.0.0",
		"--noTelemetry",
		"--verbose",
	)

	assert.Contains(output, "failed")
}

func TestDockerBuildIntegrationMultiImageWithExcludes(t *testing.T) {
	assert := NewContainerAssert(t)

	dind := StartDindContainer(t, DindContainerConfig{
		TestData: "TestDockerBuildIntegration/multi-image",
	})

	output := RunPiper(t, dind.DindContainer, "/multi-image",
		"dockerBuild",
		"--containerMultiImageBuild=true",
		"--containerImageName=partial-test",
		"--containerImageTag=1.0.0",
		fmt.Sprintf("--containerRegistryUrl=http://%s", dind.RegistryHost),
		"--containerMultiImageBuildExcludes=sub2/Dockerfile",
		"--noTelemetry",
		"--verbose",
	)

	assert.Contains(output, "SUCCESS")

	// Root and sub1 should be built
	ExecCommand(t, dind.DindContainer, "/", []string{
		"docker", "pull", fmt.Sprintf("%s/partial-test:1.0.0", dind.RegistryHost),
	})
	ExecCommand(t, dind.DindContainer, "/", []string{
		"docker", "pull", fmt.Sprintf("%s/partial-test-sub1:1.0.0", dind.RegistryHost),
	})

	// sub2 should NOT be built (excluded)
	code, _ := ExecCommandExpectNonZero(t, dind.DindContainer, "/", []string{
		"docker", "pull", fmt.Sprintf("%s/partial-test-sub2:1.0.0", dind.RegistryHost),
	})
	assert.Assertions.NotEqual(0, code, "Expected docker pull of excluded image to fail")
}
