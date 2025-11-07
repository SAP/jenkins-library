package python

import (
	"fmt"
	"os"
	"regexp"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

const (
	BOMFilename = "bom-pip.xml"
)

func CreateBOM(
	executeFn func(executable string, params ...string) error,
	existsFn func(path string) (bool, error),
	virtualEnv string,
	requirementsFile string,
	cycloneDxVersion string,
	cycloneDxSchemaVersion string,
) error {
	if exists, _ := existsFn(requirementsFile); exists {
		if err := InstallRequirements(executeFn, virtualEnv, requirementsFile); err != nil {
			return fmt.Errorf("failed to install requirements.txt: %w", err)
		}
	} else {
		log.Entry().Warnf("unable to find requirements.txt file at %s , continuing SBOM generation without requirements.txt", requirementsFile)
	}

	// Install current package to ensure it's in the venv for cyclonedx-py to detect
	// For pyproject.toml builds, this is already done in BuildWithPyProjectToml, but doing it again is safe
	// For setup.py builds, this ensures the package is installed
	log.Entry().Debug("installing current package for SBOM generation")
	if err := InstallProjectDependencies(executeFn, virtualEnv); err != nil {
		return fmt.Errorf("failed to install project dependencies: %w", err)
	}

	if err := InstallCycloneDX(executeFn, virtualEnv, cycloneDxVersion); err != nil {
		return fmt.Errorf("failed to install cyclonedx module: %w", err)
	}

	log.Entry().Debug("creating BOM")
	args := []string{"env"}

	args = append(args,
		"--output-file", BOMFilename,
		"--output-format", "XML",
		"--spec-version", cycloneDxSchemaVersion,
	)

	// Add pyproject.toml only if it exists AND contains [project] metadata
	// Without [project] metadata, cyclonedx-py will fail when using --pyproject flag
	if hasMetadata := pyprojectHasMetadata("pyproject.toml"); hasMetadata {
		args = append(args, "--pyproject", "pyproject.toml")
	}

	if err := executeFn(getBinary(virtualEnv, "cyclonedx-py"), args...); err != nil {
		return fmt.Errorf("failed to create BOM: %w", err)
	}

	// Post-process BOM to add purl to root component
	// cyclonedx-py with --pyproject doesn't generate purl for the root component
	// We use the official CycloneDX library via piperutils.UpdatePurl for robust handling
	if err := addPurlToRootComponent(BOMFilename); err != nil {
		log.Entry().Warnf("failed to add purl to root component: %v", err)
	}

	return nil
}

// addPurlToRootComponent adds a purl element to the root component in the BOM
// This is needed because cyclonedx-py doesn't generate purl when using --pyproject flag
// Uses piperutils.UpdatePurl which leverages the official CycloneDX Go library
func addPurlToRootComponent(bomFile string) error {
	component := piperutils.GetComponent(bomFile)

	if component.Name == "" || component.Version == "" {
		return fmt.Errorf("could not extract name and version from BOM metadata")
	}

	// Generate PURL for PyPI package
	purl := fmt.Sprintf("pkg:pypi/%s@%s", component.Name, component.Version)

	log.Entry().Debugf("Adding purl to root component: %s", purl)

	// Update PURL using official CycloneDX library
	if err := piperutils.UpdatePurl(bomFile, purl); err != nil {
		return fmt.Errorf("failed to update purl in BOM: %w", err)
	}

	return nil
}

// pyprojectHasMetadata checks if pyproject.toml exists and contains [project] metadata section
// Returns true only if the file exists and contains a [project] section
func pyprojectHasMetadata(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		// File doesn't exist or can't be read
		return false
	}

	content := string(data)

	// Check if the file contains a [project] section
	// We use a regex to match [project] at the start of a line (possibly with whitespace)
	projectSectionRegex := regexp.MustCompile(`(?m)^\s*\[project\]`)
	return projectSectionRegex.MatchString(content)
}
