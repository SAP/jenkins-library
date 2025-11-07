//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestPythonIntegration ./integration/...

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const DOCKER_IMAGE_PYTHON = "python:3.11"

func TestPythonIntegrationBuildProject(t *testing.T) {
	t.Parallel()

	container := StartPiperContainer(t, ContainerConfig{
		Image:    DOCKER_IMAGE_PYTHON,
		TestData: "TestPythonIntegration/python-project",
		WorkDir:  "/python-project",
	})

	output := RunPiper(t, container, "/python-project", "pythonBuild")

	assert.Contains(t, output, "info  pythonBuild - running command: piperBuild-env/bin/python -m build --no-isolation")
	assert.Contains(t, output, "info  pythonBuild - running command: piperBuild-env/bin/pip install --upgrade --root-user-action=ignore cyclonedx-bom==")
	assert.Contains(t, output, "info  pythonBuild - running command: piperBuild-env/bin/cyclonedx-py env --output-file bom-pip.xml --output-format XML --spec-version 1.4 --pyproject pyproject.toml")
	assert.Contains(t, output, "info  pythonBuild - SUCCESS")

	lsOutput := ExecCommand(t, container, "/python-project", []string{"ls", "-l", ".", "dist", "build"})
	assert.Contains(t, lsOutput, "example_pkg-0.0.1.tar.gz")
	assert.Contains(t, lsOutput, "example_pkg-0.0.1-py3-none-any.whl")
}

func TestPythonIntegrationBuildWithBOMValidation(t *testing.T) {
	t.Parallel()

	container := StartPiperContainer(t, ContainerConfig{
		Image:    DOCKER_IMAGE_PYTHON,
		TestData: "TestPythonIntegration/python-project",
		WorkDir:  "/python-project",
	})

	output := RunPiper(t, container, "/python-project", "pythonBuild")
	assert.Contains(t, output, "info  pythonBuild - running command: piperBuild-env/bin/cyclonedx-py env --output-file bom-pip.xml --output-format XML --spec-version 1.4 --pyproject pyproject.toml")
	assert.Contains(t, output, "info  pythonBuild - SUCCESS")

	// Verify BOM was generated and contains PURL
	bomOutput := ExecCommand(t, container, "/python-project", []string{"grep", "-o", "pkg:pypi/example[-_]pkg", "bom-pip.xml"})
	assert.Contains(t, bomOutput, "pkg:pypi/example", "BOM should contain PURL for the Python package")
}

func TestPythonIntegrationBuildLegacy(t *testing.T) {
	t.Parallel()

	container := StartPiperContainer(t, ContainerConfig{
		Image:    DOCKER_IMAGE_PYTHON,
		TestData: "TestPythonIntegration/python-project-legacy",
		WorkDir:  "/python-project-legacy",
	})

	output := RunPiper(t, container, "/python-project-legacy", "pythonBuild")

	// Should build using setup.py
	assert.Contains(t, output, "info  pythonBuild - running command: piperBuild-env/bin/python setup.py sdist bdist_wheel")
	assert.Contains(t, output, "info  pythonBuild - SUCCESS")

	// Verify BOM was generated but doesn't have proper metadata (no PURL expected for legacy)
	// Just verify the BOM file exists
	lsOutput := ExecCommand(t, container, "/python-project-legacy", []string{"ls", "-la", "bom-pip.xml"})
	assert.Contains(t, lsOutput, "bom-pip.xml", "BOM file should be generated")
}

func TestPythonIntegrationBuildMinimal(t *testing.T) {
	t.Parallel()

	container := StartPiperContainer(t, ContainerConfig{
		Image:    DOCKER_IMAGE_PYTHON,
		TestData: "TestPythonIntegration/python-project-minimal",
		WorkDir:  "/python-project-minimal",
	})

	output := RunPiper(t, container, "/python-project-minimal", "pythonBuild")

	// Should build using python -m build (modern approach)
	assert.Contains(t, output, "info  pythonBuild - running command: piperBuild-env/bin/python -m build --no-isolation")
	// Should generate BOM without --pyproject flag (no [project] metadata)
	assert.Contains(t, output, "info  pythonBuild - running command: piperBuild-env/bin/cyclonedx-py env --output-file bom-pip.xml --output-format XML --spec-version 1.4")
	assert.NotContains(t, output, "--pyproject")
	assert.Contains(t, output, "info  pythonBuild - SUCCESS")

	// Verify BOM was generated (even without project metadata)
	lsOutput := ExecCommand(t, container, "/python-project-minimal", []string{"ls", "-la", "bom-pip.xml"})
	assert.Contains(t, lsOutput, "bom-pip.xml", "BOM file should be generated")
}
