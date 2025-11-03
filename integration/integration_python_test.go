//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestPythonIntegration ./integration/...

package main

import (
	"testing"

	"github.com/SAP/jenkins-library/integration/testhelper"
	"github.com/stretchr/testify/assert"
)

func TestPythonIntegrationBuildProject(t *testing.T) {
	// t.Parallel()

	container := testhelper.StartPiperContainer(t, testhelper.ContainerConfig{
		Image:    "python:3.10",
		TestData: "TestPythonIntegration/python-project",
		WorkDir:  "/python-project",
	})

	output := testhelper.RunPiper(t, container, "/python-project", "pythonBuild")

	assert.Contains(t, output, "info  pythonBuild - running command: piperBuild-env/bin/python setup.py sdist bdist_wheel")
	assert.Contains(t, output, "info  pythonBuild - running command: piperBuild-env/bin/pip install --upgrade --root-user-action=ignore cyclonedx-bom==")
	assert.Contains(t, output, "info  pythonBuild - running command: piperBuild-env/bin/cyclonedx-py env --output-file bom-pip.xml --output-format XML --spec-version 1.4")
	assert.Contains(t, output, "info  pythonBuild - SUCCESS")

	lsOutput := testhelper.ExecCommand(t, container, "/python-project", []string{"ls", "-l", ".", "dist", "build"})
	assert.Contains(t, lsOutput, "example_pkg-0.0.1.tar.gz")
	assert.Contains(t, lsOutput, "example_pkg-0.0.1-py3-none-any.whl")
}

func TestPythonIntegrationBuildWithBOMValidation(t *testing.T) {
	// t.Parallel()

	container := testhelper.StartPiperContainer(t, testhelper.ContainerConfig{
		Image:    "python:3.10",
		TestData: "TestPythonIntegration/python-project",
		WorkDir:  "/python-project",
	})

	// First, run pythonBuild to generate the BOM
	output := testhelper.RunPiper(t, container, "/python-project", "pythonBuild")
	assert.Contains(t, output, "info  pythonBuild - running command: piperBuild-env/bin/cyclonedx-py env --output-file bom-pip.xml --output-format XML --spec-version 1.4")
	assert.Contains(t, output, "info  pythonBuild - SUCCESS")

	t.Log(output)

	// Now run validateBOM on the generated BOM
	output = testhelper.RunPiper(t, container, "/python-project", "validateBOM")

	t.Log(output)
	assert.Contains(t, output, "info  validateBOM - Found 1 BOM file(s) to validate")
	assert.Contains(t, output, "info  validateBOM - Validating BOM file:")
	assert.Contains(t, output, "bom-pip.xml")
	assert.Contains(t, output, "warn  validateBOM - BOM validation failed for:") // cyclonedx-py currently generates incomplete BOMs
	assert.Contains(t, output, "metadata.component.name is required but missing")
	assert.Contains(t, output, "info  validateBOM - BOM validation complete:")
}
