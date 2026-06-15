//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestPythonIntegration ./integration/...

package main

import (
	"testing"

	"github.com/SAP/jenkins-library/cmd"
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
	assert.Contains(output, "info  pythonBuild - running command: piperBuild-env/bin/cyclonedx-py env --output-file bom-pip.xml --output-format XML --spec-version "+cmd.CycloneDxSchemaVersion+" --pyproject pyproject.toml")
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
	assert.Contains(output, "info  pythonBuild - running command: piperBuild-env/bin/cyclonedx-py env --output-file bom-pip.xml --output-format XML --spec-version "+cmd.CycloneDxSchemaVersion+" --pyproject pyproject.toml")
	assert.Contains(output, "info  pythonBuild - SUCCESS")

	// Read BOM content and validate
	bomContent := ReadFile(t, container, "/python-project/bom-pip.xml")
	err := piperutils.ValidateBOM(bomContent)
	assert.NoError(err, "BOM validation should pass for Python project with valid metadata")

	// Verify BOM references correct CycloneDX schema version
	schemaVersion, err := piperutils.GetBomSchemaVersionFromContent(bomContent)
	assert.NoError(err, "bom-pip.xml should contain the CycloneDX schema version")
	assert.Equal(schemaVersion, cmd.CycloneDxSchemaVersion, "bom-pip.xml should reference CycloneDX schema version "+cmd.CycloneDxSchemaVersion)
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

	// Verify BOM references correct CycloneDX schema version
	schemaVersion, err := piperutils.GetBomSchemaVersionFromContent(bomContent)
	assert.NoError(err, "bom-pip.xml should contain the CycloneDX schema version")
	assert.Equal(schemaVersion, cmd.CycloneDxSchemaVersion, "bom-pip.xml should reference CycloneDX schema version "+cmd.CycloneDxSchemaVersion)
}

func TestPythonIntegrationRunTests(t *testing.T) {
	t.Parallel()
	assert := NewContainerAssert(t)

	container := StartPiperContainer(t, ContainerConfig{
		Image:    DOCKER_IMAGE_PYTHON,
		TestData: "TestPythonIntegration/python-project-with-tests",
		WorkDir:  "/python-project-with-tests",
	})

	output := RunPiper(t, container, "/python-project-with-tests", "pythonBuild")

	assert.Contains(output, "info  pythonBuild - SUCCESS")

	assert.FileExists(container,
		"/python-project-with-tests/TEST-python.xml",
		"/python-project-with-tests/cobertura-coverage.xml",
	)
}

func TestPythonIntegrationRunTestsFailure(t *testing.T) {
	t.Parallel()
	assert := NewContainerAssert(t)

	container := StartPiperContainer(t, ContainerConfig{
		Image:    DOCKER_IMAGE_PYTHON,
		TestData: "TestPythonIntegration/python-project-with-failing-tests",
		WorkDir:  "/python-project-with-failing-tests",
	})

	_, output := RunPiperExpectFailure(t, container, "/python-project-with-failing-tests", "pythonBuild")

	assert.Contains(output, "failed to run python tests")
}

func TestPythonIntegrationRunTestsNoTests(t *testing.T) {
	t.Parallel()
	assert := NewContainerAssert(t)

	container := StartPiperContainer(t, ContainerConfig{
		Image:    DOCKER_IMAGE_PYTHON,
		TestData: "TestPythonIntegration/python-project-no-tests",
		WorkDir:  "/python-project-no-tests",
	})

	_, output := RunPiperExpectFailure(t, container, "/python-project-no-tests", "pythonBuild")

	// "pytest collected no tests" is the prefix of the error returned by pkg/python/test.go RunTests — keep in sync.
	assert.Contains(output, "pytest collected no tests")
}

func TestPythonIntegrationRunTestsWithVerboseTestOption(t *testing.T) {
	t.Parallel()
	assert := NewContainerAssert(t)

	container := StartPiperContainer(t, ContainerConfig{
		Image:    DOCKER_IMAGE_PYTHON,
		TestData: "TestPythonIntegration/python-project-with-tests",
		WorkDir:  "/python-project-with-tests",
	})

	output := RunPiper(t, container, "/python-project-with-tests", "pythonBuild")

	assert.Contains(output, "info  pythonBuild - SUCCESS")

	// testOptions: ['-v'] is wired to pytest argv (pkg/python/test.go:44-49).
	// The injected report flags are prepended; user options are appended after them.
	assert.Contains(output, "pytest --junitxml=TEST-python.xml --cov --cov-report=xml:cobertura-coverage.xml -v")

	// With -v, pytest prints each test function name; without -v it prints dots.
	// These are the two tests in tests/test_example.py.
	assert.Contains(output, "test_add_one")
	assert.Contains(output, "test_add_one_negative")

	// The injected report paths must be unchanged even when testOptions are present.
	assert.FileExists(container,
		"/python-project-with-tests/TEST-python.xml",
		"/python-project-with-tests/cobertura-coverage.xml",
	)
}

func TestPythonIntegrationRunTestsRejectsJunitxmlOverride(t *testing.T) {
	t.Parallel()
	assert := NewContainerAssert(t)

	container := StartPiperContainer(t, ContainerConfig{
		Image:    DOCKER_IMAGE_PYTHON,
		TestData: "TestPythonIntegration/python-project-test-options-invalid",
		WorkDir:  "/python-project-test-options-invalid",
	})

	_, output := RunPiperExpectFailure(t, container, "/python-project-test-options-invalid", "pythonBuild")

	// Validation message from pkg/python/test.go:33 — keep in sync if reworded.
	assert.Contains(output, "testOptions must not override --junitxml/--junit-xml")
	// The offending option is echoed via %q in the error.
	assert.Contains(output, "--junitxml=hijack.xml")
	// Outer wrap from cmd/pythonBuild.go:96.
	assert.Contains(output, "failed to run python tests")

	// Validation fires before pytest is invoked; the command must never appear.
	assert.NotContains(output, "running command: piperBuild-env/bin/pytest")
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
	assert.Contains(output, "info  pythonBuild - running command: piperBuild-env/bin/cyclonedx-py env --output-file bom-pip.xml --output-format XML --spec-version "+cmd.CycloneDxSchemaVersion)
	assert.NotContains(output, "--pyproject")
	assert.Contains(output, "info  pythonBuild - SUCCESS")

	// Verify BOM was generated
	assert.FileExists(container, "/python-project-minimal/bom-pip.xml")

	// Read BOM content and validate - should fail (no [project] metadata in pyproject.toml)
	bomContent := ReadFile(t, container, "/python-project-minimal/bom-pip.xml")
	err := piperutils.ValidateBOM(bomContent)
	assert.Error(err, "BOM validation should fail for minimal Python project without metadata")
	assert.Regexp("metadata\\.component\\.(name|purl)", err.Error())

	// Verify BOM references correct CycloneDX schema version
	schemaVersion, err := piperutils.GetBomSchemaVersionFromContent(bomContent)
	assert.NoError(err, "bom-pip.xml should contain the CycloneDX schema version")
	assert.Equal(schemaVersion, cmd.CycloneDxSchemaVersion, "bom-pip.xml should reference CycloneDX schema version "+cmd.CycloneDxSchemaVersion)
}
