//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestValidateBOMIntegration ./integration/...

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const DOCKER_IMAGE_VALIDATEBOM = "alpine:latest"

// TestValidateBOMIntegrationInvalidBOM tests that validateBOM logs warnings for invalid BOMs
// but does not fail the build by default
func TestValidateBOMIntegrationInvalidBOM(t *testing.T) {
	t.Parallel()

	container := StartPiperContainer(t, ContainerConfig{
		Image:    DOCKER_IMAGE_VALIDATEBOM,
		TestData: "TestValidateBOMIntegration/invalid-bom",
		WorkDir:  "/invalid-bom",
	})

	output := RunPiper(t, container, "/invalid-bom", "validateBOM")

	assert.Contains(t, output, "info  validateBOM - Found 1 BOM file(s) to validate")
	assert.Contains(t, output, "info  validateBOM - Validating BOM file:")
	assert.Contains(t, output, "bom-invalid.xml")
	assert.Contains(t, output, "warn  validateBOM - BOM validation failed for:")
	assert.Contains(t, output, "info  validateBOM - BOM validation complete: 0/1 files validated successfully")
}

// TestValidateBOMIntegrationFailOnError tests that validateBOM fails the build when
// failOnValidationError is true and a BOM is invalid
func TestValidateBOMIntegrationFailOnError(t *testing.T) {
	t.Parallel()

	container := StartPiperContainer(t, ContainerConfig{
		Image:    DOCKER_IMAGE_VALIDATEBOM,
		TestData: "TestValidateBOMIntegration/invalid-bom",
		WorkDir:  "/invalid-bom",
	})

	exitCode, output := RunPiperExpectFailure(t, container, "/invalid-bom", "validateBOM", "--failOnValidationError")

	assert.NotEqual(t, 0, exitCode)
	assert.Contains(t, output, "info  validateBOM - Found 1 BOM file(s) to validate")
	assert.Contains(t, output, "info  validateBOM - Validating BOM file:")
	assert.Contains(t, output, "bom-invalid.xml")
	assert.Contains(t, output, "fatal validateBOM")
	assert.Contains(t, output, "BOM validation failed for")
}

// TestValidateBOMIntegrationSkip tests that validateBOM can be skipped entirely
func TestValidateBOMIntegrationSkip(t *testing.T) {
	t.Parallel()

	container := StartPiperContainer(t, ContainerConfig{
		Image:    DOCKER_IMAGE_VALIDATEBOM,
		TestData: "TestValidateBOMIntegration/invalid-bom",
		WorkDir:  "/invalid-bom",
	})

	output := RunPiper(t, container, "/invalid-bom", "validateBOM", "--skip")

	assert.Contains(t, output, "info  validateBOM - BOM validation skipped (skip: true)")
	assert.NotContains(t, output, "info  validateBOM - Found")
	assert.NotContains(t, output, "info  validateBOM - Validating BOM file:")
}

// TestValidateBOMIntegrationNoBOMs tests that validateBOM succeeds silently when no BOMs are found
func TestValidateBOMIntegrationNoBOMs(t *testing.T) {
	t.Parallel()

	container := StartPiperContainer(t, ContainerConfig{
		Image:    DOCKER_IMAGE_VALIDATEBOM,
		TestData: "TestValidateBOMIntegration/no-boms",
		WorkDir:  "/no-boms",
	})

	output := RunPiper(t, container, "/no-boms", "validateBOM")

	assert.Contains(t, output, "info  validateBOM - No BOM files found matching pattern")
	assert.Contains(t, output, "skipping validation")
}

// TestValidateBOMIntegrationMultipleBOMs tests that validateBOM can handle multiple BOMs
func TestValidateBOMIntegrationMultipleBOMs(t *testing.T) {
	t.Parallel()

	container := StartPiperContainer(t, ContainerConfig{
		Image:    DOCKER_IMAGE_VALIDATEBOM,
		TestData: "TestValidateBOMIntegration/multiple-boms",
		WorkDir:  "/multiple-boms",
	})

	output := RunPiper(t, container, "/multiple-boms", "validateBOM")

	assert.Contains(t, output, "info  validateBOM - Found 3 BOM file(s) to validate")
	assert.Contains(t, output, "info  validateBOM - BOM validation complete: 3/3 files validated successfully")
}
