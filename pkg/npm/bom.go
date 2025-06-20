package npm

import (
	"fmt"
	"path/filepath"
	"runtime"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
)

const (
	// Output file names
	npmBomFilename  = "bom-npm.xml"
	tempBomFilename = "bom-npm.json"

	// Package versions
	cycloneDxNpmPackageVersion = "@cyclonedx/cyclonedx-npm@1.11.0"
	cycloneDxBomPackageVersion = "@cyclonedx/bom@^3.10.6"
	cdxgenPackageVersion       = "@cyclonedx/cdxgen@^11.3.2"
	cycloneDxCliVersion        = "v0.27.2"

	// Configuration
	cycloneDxSchemaVersion = "1.4"
)

var cycloneDxCliUrl = map[struct{ os, arch string }]string{
	{"windows", "amd64"}: "https://github.com/CycloneDX/cyclonedx-cli/releases/download/%s/cyclonedx-win-x64.exe",
	{"windows", "arm64"}: "https://github.com/CycloneDX/cyclonedx-cli/releases/download/%s/cyclonedx-win-arm64.exe",
	{"darwin", "amd64"}:  "https://github.com/CycloneDX/cyclonedx-cli/releases/download/%s/cyclonedx-osx-x64",
	{"darwin", "arm64"}:  "https://github.com/CycloneDX/cyclonedx-cli/releases/download/%s/cyclonedx-osx-arm64",
	{"linux", "amd64"}:   "https://github.com/CycloneDX/cyclonedx-cli/releases/download/%s/cyclonedx-linux-x64",
	{"linux", "arm64"}:   "https://github.com/CycloneDX/cyclonedx-cli/releases/download/%s/cyclonedx-linux-arm64",
	{"linux", "arm"}:     "https://github.com/CycloneDX/cyclonedx-cli/releases/download/%s/cyclonedx-linux-arm",
}

// CreateBOM generates a CycloneDX Bill of Materials (BOM) file for the given package.json files.
// It supports both pnpm and other package managers (npm/yarn) with different BOM generation strategies.
func (exec *Execute) CreateBOM(packageJSONFiles []string) error {
	log.Entry().Debug("Detecting package manager...")
	pm, err := exec.detectPackageManager("")
	if err != nil {
		return fmt.Errorf("failed to detect package manager (looking for package.json, package-lock.json, pnpm-lock.yaml): %w", err)
	}
	if pm != nil {
		log.Entry().Debugf("Detected package manager: %s", pm.Name)
	}

	if pm != nil && pm.Name == "pnpm" {
		return exec.createPnpmBOM(packageJSONFiles)
	}

	return exec.createNpmBOM(packageJSONFiles)
}

// createPnpmBOM generates a BOM for pnpm projects using cdxgen and cyclonedx-cli
func (exec *Execute) createPnpmBOM(packageJSONFiles []string) error {
	log.Entry().Info("Starting BOM generation for pnpm project...")
	log.Entry().Debug("Downloading CycloneDX CLI tool...")

	cliPath, err := exec.downloadCycloneDxCli()
	if err != nil {
		return fmt.Errorf("failed to setup CycloneDX CLI tool: %w", err)
	}
	log.Entry().Debugf("CycloneDX CLI downloaded successfully to: %s", cliPath)

	execRunner := exec.Utils.GetExecRunner()
	log.Entry().Debug("Installing cdxgen tool...")
	if err := exec.installCdxgen(execRunner); err != nil {
		return fmt.Errorf("cdxgen installation failed: %w", err)
	}
	log.Entry().Debug("cdxgen tool installed successfully")

	return exec.generatePnpmBOMFiles(packageJSONFiles, cliPath, execRunner)
}

// downloadCycloneDxCli downloads the appropriate cyclonedx-cli binary for the current OS/arch
func (exec *Execute) downloadCycloneDxCli() (string, error) {
	osArch := struct{ os, arch string }{runtime.GOOS, runtime.GOARCH}
	urlTemplate, ok := cycloneDxCliUrl[osArch]
	if !ok {
		return "", fmt.Errorf("cyclonedx-cli not available for OS/architecture combination: %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	url := fmt.Sprintf(urlTemplate, cycloneDxCliVersion)
	cliPath, err := piperhttp.DownloadExecutable("", exec.Utils.GetFileUtils(), exec.Utils.GetDownloadUtils(), url)
	if err != nil {
		return "", fmt.Errorf("failed to download cyclonedx-cli: %w", err)
	}

	return cliPath, nil
}

// installCdxgen installs the cdxgen package for pnpm projects
func (exec *Execute) installCdxgen(execRunner ExecRunner) error {
	log.Entry().Debugf("Installing cdxgen version %s to %s", cdxgenPackageVersion, tmpInstallFolder)
	err := execRunner.RunExecutable("npm", "install", cdxgenPackageVersion, "--prefix", tmpInstallFolder)
	if err != nil {
		return fmt.Errorf("failed to install cdxgen (path: %s): %w", tmpInstallFolder, err)
	}
	return nil
}

// generatePnpmBOMFiles generates BOM files for each package.json file using cdxgen and cyclonedx-cli
func (exec *Execute) generatePnpmBOMFiles(packageJSONFiles []string, cliPath string, execRunner ExecRunner) error {
	for _, packageJSONFile := range packageJSONFiles {
		path := filepath.Dir(packageJSONFile)
		jsonBomPath := filepath.Join(path, tempBomFilename)
		xmlBomPath := filepath.Join(path, npmBomFilename)

		cdxgenExecutable := tmpInstallFolder + "/node_modules/.bin/cdxgen"
		params := []string{
			"-r",
			"-o", jsonBomPath,
			"--spec-version", cycloneDxSchemaVersion,
		}

		log.Entry().Debugf("Executing cdxgen with params: %v", params)
		log.Entry().Debugf("cdxgen executable path: %s", cdxgenExecutable)

		if err := execRunner.RunExecutable(cdxgenExecutable, params...); err != nil {
			return fmt.Errorf("failed to generate CycloneDX BOM with cdxgen for package: %s. Error: %w", packageJSONFile, err)
		}

		log.Entry().Infof("Generated CycloneDX BOM in JSON format at %s", jsonBomPath)

		// Convert JSON to XML using cyclonedx-cli
		log.Entry().Debugf("Converting BOM from JSON to XML using cyclonedx-cli at: %s", cliPath)
		log.Entry().Debugf("Input file: %s, Output file: %s", jsonBomPath, xmlBomPath)

		if err := execRunner.RunExecutable(cliPath, "convert", "--input-file", jsonBomPath, "--output-format", "xml", "--output-file", xmlBomPath); err != nil {
			return fmt.Errorf("failed to convert BOM to XML format for package: %s. Input: %s, Output: %s, Error: %w",
				packageJSONFile, jsonBomPath, xmlBomPath, err)
		}

		log.Entry().Infof("Converted CycloneDX BOM to XML format at %s", xmlBomPath)
	}
	return nil
}

// createNpmBOM generates a BOM for npm/yarn projects using cyclonedx-npm or cyclonedx/bom as fallback
func (exec *Execute) createNpmBOM(packageJSONFiles []string) error {
	// Primary attempt with cyclonedx-npm
	cycloneDxNpmInstallParams := []string{"install", "--no-save", cycloneDxNpmPackageVersion, "--prefix", tmpInstallFolder}
	cycloneDxNpmRunParams := []string{"--output-format", "XML", "--spec-version", cycloneDxSchemaVersion, "--omit", "dev", "--output-file"}

	err := exec.createBOMWithParams(cycloneDxNpmInstallParams, cycloneDxNpmRunParams, packageJSONFiles, false)
	if err == nil {
		return nil
	}

	log.Entry().Infof("Failed to generate BOM with @cyclonedx/cyclonedx-npm@%s, falling back to @cyclonedx/bom: %v",
		cycloneDxNpmPackageVersion, err)
	log.Entry().Debug("Note: This fallback is normal if using an older Node.js version or if cyclonedx-npm installation fails")

	// Fallback attempt with cyclonedx/bom
	cycloneDxBomInstallParams := []string{"install", cycloneDxBomPackageVersion, "--no-save"}
	cycloneDxBomRunParams := []string{"cyclonedx-bom", "--output"}

	err = exec.createBOMWithParams(cycloneDxBomInstallParams, cycloneDxBomRunParams, packageJSONFiles, true)
	if err != nil {
		return fmt.Errorf("failed to generate BOM with fallback package cyclonedx/bom: %w", err)
	}

	return nil
}

// createBOMWithParams facilitates BOM generation with different CycloneDX packages
func (exec *Execute) createBOMWithParams(packageInstallParams []string, packageRunParams []string, packageJSONFiles []string, fallback bool) error {
	execRunner := exec.Utils.GetExecRunner()

	log.Entry().Debugf("Installing CycloneDX package with params: %v", packageInstallParams)
	if err := execRunner.RunExecutable("npm", packageInstallParams...); err != nil {
		return fmt.Errorf("failed to install CycloneDX BOM package (params: %v): %w", packageInstallParams, err)
	}

	for _, packageJSONFile := range packageJSONFiles {
		path := filepath.Dir(packageJSONFile)
		params := append(packageRunParams, filepath.Join(path, npmBomFilename))
		executable := "npx"

		if !fallback {
			params = append(params, packageJSONFile)
			executable = tmpInstallFolder + "/node_modules/.bin/cyclonedx-npm"
		} else {
			params = append(params, path)
		}

		log.Entry().Debugf("Generating BOM for package: %s", packageJSONFile)
		log.Entry().Debugf("Using executable: %s with params: %v", executable, params)

		if err := execRunner.RunExecutable(executable, params...); err != nil {
			return fmt.Errorf("failed to generate CycloneDX BOM for package %s using %s: %w",
				packageJSONFile, executable, err)
		}
	}

	return nil
}
