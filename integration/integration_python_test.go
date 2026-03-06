//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestPythonIntegration ./integration/...

package main

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/piperutils"
)

const DOCKER_IMAGE_PYTHON = "python:3.11"

func TestPythonIntegrationBuildProject(t *testing.T) {
	t.Parallel()
	assert := NewContainerAssert(t)

	container := StartPiperContainer(t, ContainerConfig{
		Image:    DOCKER_IMAGE_PYTHON,
		TestData: "TestPythonIntegration/python-project",
		WorkDir:  "/python-project",
	})

	output := RunPiper(t, container, "/python-project", "pythonBuild")

	assert.Contains(output, "info  pythonBuild - running command: piperBuild-env/bin/python -m build --no-isolation")
	assert.Contains(output, "info  pythonBuild - running command: piperBuild-env/bin/pip install --upgrade --root-user-action=ignore cyclonedx-bom==")
	assert.Contains(output, "info  pythonBuild - running command: piperBuild-env/bin/cyclonedx-py env --output-file bom-pip.xml --output-format XML --spec-version 1.4 --pyproject pyproject.toml")
	assert.Contains(output, "info  pythonBuild - SUCCESS")

	assert.FileExists(container,
		"/python-project/dist/example_pkg-0.0.1.tar.gz",
		"/python-project/dist/example_pkg-0.0.1-py3-none-any.whl",
	)
}

func TestPythonIntegrationBuildWithBOMValidation(t *testing.T) {
	t.Parallel()
	assert := NewContainerAssert(t)

	container := StartPiperContainer(t, ContainerConfig{
		Image:    DOCKER_IMAGE_PYTHON,
		TestData: "TestPythonIntegration/python-project",
		WorkDir:  "/python-project",
	})

	output := RunPiper(t, container, "/python-project", "pythonBuild")
	assert.Contains(output, "info  pythonBuild - running command: piperBuild-env/bin/cyclonedx-py env --output-file bom-pip.xml --output-format XML --spec-version 1.4 --pyproject pyproject.toml")
	assert.Contains(output, "info  pythonBuild - SUCCESS")

	// Read BOM content and validate
	bomContent := ReadFile(t, container, "/python-project/bom-pip.xml")
	err := piperutils.ValidateBOM(bomContent)
	assert.NoError(err, "BOM validation should pass for Python project with valid metadata")
}

func TestPythonIntegrationBuildLegacy(t *testing.T) {
	t.Parallel()
	assert := NewContainerAssert(t)

	container := StartPiperContainer(t, ContainerConfig{
		Image:    DOCKER_IMAGE_PYTHON,
		TestData: "TestPythonIntegration/python-project-legacy",
		WorkDir:  "/python-project-legacy",
	})

	output := RunPiper(t, container, "/python-project-legacy", "pythonBuild")

	// Should build using setup.py
	assert.Contains(output, "info  pythonBuild - running command: piperBuild-env/bin/python setup.py sdist bdist_wheel")
	assert.Contains(output, "info  pythonBuild - SUCCESS")

	// Read BOM content and validate - should fail due to missing metadata
	bomContent := ReadFile(t, container, "/python-project-legacy/bom-pip.xml")
	err := piperutils.ValidateBOM(bomContent)
	assert.Error(err, "BOM validation should fail for legacy Python project without metadata")
	assert.Regexp("metadata\\.component\\.(name|purl)", err.Error())
}

func TestPythonIntegrationBuildMinimal(t *testing.T) {
	t.Parallel()
	assert := NewContainerAssert(t)

	container := StartPiperContainer(t, ContainerConfig{
		Image:    DOCKER_IMAGE_PYTHON,
		TestData: "TestPythonIntegration/python-project-minimal",
		WorkDir:  "/python-project-minimal",
	})

	output := RunPiper(t, container, "/python-project-minimal", "pythonBuild")

	// Should build using python -m build (modern approach)
	assert.Contains(output, "info  pythonBuild - running command: piperBuild-env/bin/python -m build --no-isolation")
	// Should generate BOM without --pyproject flag (no [project] metadata)
	assert.Contains(output, "info  pythonBuild - running command: piperBuild-env/bin/cyclonedx-py env --output-file bom-pip.xml --output-format XML --spec-version 1.4")
	assert.NotContains(output, "--pyproject")
	assert.Contains(output, "info  pythonBuild - SUCCESS")

	// Verify BOM was generated
	assert.FileExists(container, "/python-project-minimal/bom-pip.xml")

	// Read BOM content and validate - should fail (no [project] metadata in pyproject.toml)
	bomContent := ReadFile(t, container, "/python-project-minimal/bom-pip.xml")
	err := piperutils.ValidateBOM(bomContent)
	assert.Error(err, "BOM validation should fail for minimal Python project without metadata")
	assert.Regexp("metadata\\.component\\.(name|purl)", err.Error())
}
