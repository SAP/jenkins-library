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
	cycloneDxNpmInstallationFolder = "./tmp" // This folder is also added to npmignore in publish.go
	cycloneDxSchemaVersion         = "1.4"
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
	pm, err := exec.detectPackageManager()
	if err != nil {
		return fmt.Errorf("failed to detect package manager: %w", err)
	}

	if pm != nil && pm.Name == "pnpm" {
		return exec.createPnpmBOM(packageJSONFiles)
	}

	return exec.createNpmBOM(packageJSONFiles)
}

// createPnpmBOM generates a BOM for pnpm projects using cdxgen and cyclonedx-cli
func (exec *Execute) createPnpmBOM(packageJSONFiles []string) error {
	cliPath, err := exec.downloadCycloneDxCli()
	if err != nil {
		return err
	}

	execRunner := exec.Utils.GetExecRunner()
	if err := exec.installCdxgen(execRunner); err != nil {
		return err
	}

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
	err := execRunner.RunExecutable("npm", "install", cdxgenPackageVersion, "--prefix", cycloneDxNpmInstallationFolder)
	if err != nil {
		return fmt.Errorf("failed to install cdxgen: %w", err)
	}
	return nil
}

// generatePnpmBOMFiles generates BOM files for each package.json file using cdxgen and cyclonedx-cli
func (exec *Execute) generatePnpmBOMFiles(packageJSONFiles []string, cliPath string, execRunner ExecRunner) error {
	for _, packageJSONFile := range packageJSONFiles {
		path := filepath.Dir(packageJSONFile)
		jsonBomPath := filepath.Join(path, tempBomFilename)
		xmlBomPath := filepath.Join(path, npmBomFilename)

		cdxgenExecutable := cycloneDxNpmInstallationFolder + "/node_modules/.bin/cdxgen"
		params := []string{
			"-r",
			"-o", jsonBomPath,
			"--spec-version", cycloneDxSchemaVersion,
		}

		if err := execRunner.RunExecutable(cdxgenExecutable, params...); err != nil {
			return fmt.Errorf("failed to generate CycloneDX BOM with cdxgen: %w", err)
		}

		log.Entry().Infof("Generated CycloneDX BOM in JSON format at %s", jsonBomPath)

		// Convert JSON to XML using cyclonedx-cli
		if err := execRunner.RunExecutable(cliPath, "convert", "--input-file", jsonBomPath, "--output-format", "xml", "--output-file", xmlBomPath); err != nil {
			return fmt.Errorf("failed to convert BOM to XML format: %w", err)
		}

		log.Entry().Infof("Converted CycloneDX BOM to XML format at %s", xmlBomPath)
	}
	return nil
}

// createNpmBOM generates a BOM for npm/yarn projects using cyclonedx-npm or cyclonedx/bom as fallback
func (exec *Execute) createNpmBOM(packageJSONFiles []string) error {
	// Primary attempt with cyclonedx-npm
	cycloneDxNpmInstallParams := []string{"install", "--no-save", cycloneDxNpmPackageVersion, "--prefix", cycloneDxNpmInstallationFolder}
	cycloneDxNpmRunParams := []string{"--output-format", "XML", "--spec-version", cycloneDxSchemaVersion, "--omit", "dev", "--output-file"}

	err := exec.createBOMWithParams(cycloneDxNpmInstallParams, cycloneDxNpmRunParams, packageJSONFiles, false)
	if err == nil {
		return nil
	}

	log.Entry().Infof("Failed to generate BOM with cyclonedx-npm, falling back to cyclonedx/bom: %v", err)

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

	if err := execRunner.RunExecutable("npm", packageInstallParams...); err != nil {
		return fmt.Errorf("failed to install CycloneDX BOM package: %w", err)
	}

	for _, packageJSONFile := range packageJSONFiles {
		path := filepath.Dir(packageJSONFile)
		params := append(packageRunParams, filepath.Join(path, npmBomFilename))
		executable := "npx"

		if !fallback {
			params = append(params, packageJSONFile)
			executable = cycloneDxNpmInstallationFolder + "/node_modules/.bin/cyclonedx-npm"
		} else {
			params = append(params, path)
		}

		if err := execRunner.RunExecutable(executable, params...); err != nil {
			return fmt.Errorf("failed to generate CycloneDX BOM: %w", err)
		}
	}

	return nil
}
