package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/SAP/jenkins-library/pkg/build"
	"github.com/SAP/jenkins-library/pkg/buildsettings"
	"github.com/SAP/jenkins-library/pkg/certutils"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/versioning"
)

const (
	rustUnitTestOutput        = "TEST-rust.xml"
	rustIntegrationTestOutput = "TEST-rust-integration.xml"
)

type rustBuildUtils interface {
	command.ExecRunner
	piperutils.FileUtils
	piperhttp.Uploader

	getDockerImageValue(stepName string) (string, error)
	DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error
}

type rustBuildUtilsBundle struct {
	*command.Command
	*piperutils.Files
	piperhttp.Uploader
	httpClient *piperhttp.Client
}

func (r *rustBuildUtilsBundle) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	return r.httpClient.DownloadFile(url, filename, header, cookies)
}

func (r *rustBuildUtilsBundle) getDockerImageValue(stepName string) (string, error) {
	return GetDockerImageValue(stepName)
}

func newRustBuildUtils(config rustBuildOptions) rustBuildUtils {
	httpClientOptions := piperhttp.ClientOptions{}
	if len(config.CustomTLSCertificateLinks) > 0 {
		httpClientOptions.TransportSkipVerification = false
		httpClientOptions.TrustedCerts = config.CustomTLSCertificateLinks
	}

	httpClient := piperhttp.Client{}
	httpClient.SetOptions(httpClientOptions)

	utils := rustBuildUtilsBundle{
		Command: &command.Command{
			StepName: "rustBuild",
		},
		Files:      &piperutils.Files{},
		Uploader:   &httpClient,
		httpClient: &httpClient,
	}
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func rustBuild(config rustBuildOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *rustBuildCommonPipelineEnvironment) {
	utils := newRustBuildUtils(config)

	err := runRustBuild(&config, telemetryData, utils, commonPipelineEnvironment)
	if err != nil {
		log.Entry().WithError(err).Fatal("execution of rust build failed")
	}
}

func readCargoCoordinates(utils rustBuildUtils) (versioning.Coordinates, error) {
	type cargoManifest struct {
		Package struct {
			Name    string `toml:"name"`
			Version string `toml:"version"`
		} `toml:"package"`
	}
	content, err := utils.FileRead(versioning.CargoBuildDescriptor)
	if err != nil {
		return versioning.Coordinates{}, fmt.Errorf("failed to read Cargo.toml: %w", err)
	}
	var m cargoManifest
	if _, err := toml.Decode(string(content), &m); err != nil {
		return versioning.Coordinates{}, fmt.Errorf("failed to parse Cargo.toml: %w", err)
	}
	return versioning.Coordinates{ArtifactID: m.Package.Name, Version: m.Package.Version}, nil
}

func prepareRustEnvironment(config *rustBuildOptions, utils rustBuildUtils) error {
	if err := certutils.CertificateUpdate(config.CustomTLSCertificateLinks, utils, utils, "/etc/ssl/certs/ca-certificates.crt"); err != nil {
		return fmt.Errorf("failed to update certificates: %w", err)
	}
	if config.CargoRegistryToken != "" {
		utils.SetEnv(append(os.Environ(), fmt.Sprintf("CARGO_REGISTRY_TOKEN=%s", config.CargoRegistryToken)))
	}
	return nil
}

func runRustBuild(config *rustBuildOptions, _ *telemetry.CustomData, utils rustBuildUtils, commonPipelineEnvironment *rustBuildCommonPipelineEnvironment) error {
	coords, err := readCargoCoordinates(utils)
	if err != nil {
		return err
	}
	packageName := coords.ArtifactID

	if err := prepareRustEnvironment(config, utils); err != nil {
		return err
	}

	failedTests := false

	if config.RunTests {
		success, err := runRustTests(config, utils)
		if err != nil {
			return err
		}
		failedTests = !success
	}

	if config.RunTests && config.ReportCoverage {
		if err := reportRustTestCoverage(config, utils); err != nil {
			return err
		}
	} else if config.ReportCoverage && !config.RunTests {
		log.Entry().Warn("reportCoverage is enabled but runTests is false — skipping coverage report")
	}

	if config.RunIntegrationTests {
		success, err := runRustIntegrationTests(config, utils)
		if err != nil {
			return err
		}
		failedTests = failedTests || !success
	}

	if failedTests {
		log.SetErrorCategory(log.ErrorTest)
		return fmt.Errorf("some tests failed")
	}

	if config.RunLint {
		if err := runRustLint(config, utils); err != nil {
			return err
		}
	}

	if config.CreateBOM {
		if err := runRustBOMCreation(utils); err != nil {
			return err
		}
	}

	var binaries []string
	for _, target := range config.TargetArchitectures {
		target := target
		binary, err := runRustBuildPerArchitecture(config, packageName, utils, target)
		if err != nil {
			return err
		}
		if binary != "" {
			binaries = append(binaries, binary)
		}
	}

	log.Entry().Debugf("creating build settings information...")
	stepName := "rustBuild"
	dockerImage, err := utils.getDockerImageValue(stepName)
	if err != nil {
		return err
	}

	buildConfig := buildsettings.BuildOptions{
		CreateBOM:         config.CreateBOM,
		Publish:           config.Publish,
		BuildSettingsInfo: config.BuildSettingsInfo,
		DockerImage:       dockerImage,
	}
	buildSettingsInfo, err := buildsettings.CreateBuildSettingsInfo(&buildConfig, stepName)
	if err != nil {
		log.Entry().Warnf("failed to create build settings info: %v", err)
	}
	commonPipelineEnvironment.custom.buildSettingsInfo = buildSettingsInfo

	if config.Publish {
		if len(config.TargetRepositoryURL) == 0 {
			return fmt.Errorf("there's no target repository for binary publishing configured")
		}

		artifactVersion := config.ArtifactVersion
		if artifactVersion == "" {
			if coords.Version == "" {
				return fmt.Errorf("no version found in Cargo.toml and no artifactVersion configured")
			}
			artifactVersion = coords.Version
		}

		repoClientOptions := piperhttp.ClientOptions{
			Username:     config.TargetRepositoryUser,
			Password:     config.TargetRepositoryPassword,
			TrustedCerts: config.CustomTLSCertificateLinks,
		}
		utils.SetOptions(repoClientOptions)

		var binaryArtifacts piperenv.Artifacts
		var buildCoordinates []versioning.Coordinates

		for _, binary := range binaries {
			targetPath := fmt.Sprintf("rust/%s/%s/%s", packageName, artifactVersion, binary)

			separator := "/"
			if strings.HasSuffix(config.TargetRepositoryURL, "/") {
				separator = ""
			}
			targetURL := fmt.Sprintf("%s%s%s", config.TargetRepositoryURL, separator, targetPath)
			log.Entry().Infof("publishing artifact: %s", targetURL)

			response, err := utils.UploadRequest(http.MethodPut, targetURL, binary, "", nil, nil, "binary")
			if err != nil {
				return fmt.Errorf("couldn't upload artifact: %w", err)
			}
			if response.StatusCode != 200 && response.StatusCode != 201 {
				return fmt.Errorf("couldn't upload artifact, received status code %d", response.StatusCode)
			}

			binaryArtifacts = append(binaryArtifacts, piperenv.Artifact{Name: binary})

			if config.CreateBuildArtifactsMetadata {
				buildCoordinates = append(buildCoordinates, versioning.Coordinates{
					ArtifactID: binary,
					GroupID:    packageName,
					Version:    artifactVersion,
					URL:        config.TargetRepositoryURL,
					BuildPath:  filepath.Dir(binary),
				})
			}
		}
		commonPipelineEnvironment.custom.artifacts = binaryArtifacts

		if len(buildCoordinates) > 0 {
			var buildArtifacts build.BuildArtifacts
			buildArtifacts.Coordinates = buildCoordinates
			jsonResult, err := json.Marshal(buildArtifacts)
			if err != nil {
				log.Entry().Warnf("failed to marshal build artifacts: %v", err)
			} else {
				commonPipelineEnvironment.custom.rustBuildArtifacts = string(jsonResult)
			}
		}
	}

	return nil
}

func runRustTests(config *rustBuildOptions, utils rustBuildUtils) (bool, error) {
	testArgs := []string{"test", "--no-fail-fast"}
	testArgs = append(testArgs, config.TestOptions...)
	if err := utils.RunExecutable("cargo", testArgs...); err != nil {
		exists, fileErr := utils.FileExists(rustUnitTestOutput)
		if exists && fileErr == nil {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func runRustIntegrationTests(config *rustBuildOptions, utils rustBuildUtils) (bool, error) {
	testArgs := []string{"test", "--tests", "--no-fail-fast"}
	testArgs = append(testArgs, config.TestOptions...)
	if err := utils.RunExecutable("cargo", testArgs...); err != nil {
		exists, fileErr := utils.FileExists(rustIntegrationTestOutput)
		if exists && fileErr == nil {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func reportRustTestCoverage(config *rustBuildOptions, utils rustBuildUtils) error {
	switch config.CoverageTool {
	case "llvm-cov":
		return reportRustCoverageLlvmCov(config, utils)
	default:
		return reportRustCoverageTarpaulin(config, utils)
	}
}

func reportRustCoverageTarpaulin(config *rustBuildOptions, utils rustBuildUtils) error {
	if err := utils.RunExecutable("cargo", "install", "cargo-tarpaulin"); err != nil {
		return fmt.Errorf("failed to install cargo-tarpaulin: %w", err)
	}
	args := []string{"tarpaulin"}
	switch config.CoverageFormat {
	case "cobertura":
		args = append(args, "--out", "Xml", "--output-dir", ".")
	case "lcov":
		args = append(args, "--out", "Lcov", "--output-dir", ".")
	default:
		args = append(args, "--out", "Html", "--output-dir", ".")
	}
	if err := utils.RunExecutable("cargo", args...); err != nil {
		return fmt.Errorf("failed to run cargo-tarpaulin: %w", err)
	}
	return nil
}

func reportRustCoverageLlvmCov(config *rustBuildOptions, utils rustBuildUtils) error {
	if err := utils.RunExecutable("cargo", "install", "cargo-llvm-cov"); err != nil {
		return fmt.Errorf("failed to install cargo-llvm-cov: %w", err)
	}
	args := []string{"llvm-cov"}
	switch config.CoverageFormat {
	case "cobertura":
		args = append(args, "--cobertura", "--output-path", "cobertura-coverage.xml")
	case "lcov":
		args = append(args, "--lcov", "--output-path", "lcov.info")
	default:
		args = append(args, "--html")
	}
	if err := utils.RunExecutable("cargo", args...); err != nil {
		return fmt.Errorf("failed to run cargo-llvm-cov: %w", err)
	}
	return nil
}

func runRustLint(config *rustBuildOptions, utils rustBuildUtils) error {
	lintArgs := []string{"clippy", "--all-targets", "--", "-D", "warnings"}
	lintArgs = append(lintArgs, config.ClippyArgs...)
	if err := utils.RunExecutable("cargo", lintArgs...); err != nil {
		if config.FailOnLintingError {
			return fmt.Errorf("cargo clippy reported linting errors: %w", err)
		}
		log.Entry().Warnf("cargo clippy reported linting errors (ignored because failOnLintingError=false)")
	}
	return nil
}

func runRustBOMCreation(utils rustBuildUtils) error {
	if err := utils.RunExecutable("cargo", "install", "cargo-cyclonedx"); err != nil {
		return fmt.Errorf("failed to install cargo-cyclonedx: %w", err)
	}
	if err := utils.RunExecutable("cargo", "cyclonedx", "--format", "xml"); err != nil {
		return fmt.Errorf("BOM creation failed: %w", err)
	}
	return nil
}

func runRustBuildPerArchitecture(config *rustBuildOptions, packageName string, utils rustBuildUtils, target string) (string, error) {
	if err := utils.RunExecutable("rustup", "target", "add", target); err != nil {
		return "", fmt.Errorf("failed to add rust target %s: %w", target, err)
	}

	buildArgs := []string{"build", "--profile", config.CargoProfile, "--target", target}
	if len(config.CargoFeatures) > 0 {
		buildArgs = append(buildArgs, "--features", strings.Join(config.CargoFeatures, ","))
	}
	buildArgs = append(buildArgs, config.BuildFlags...)

	if err := utils.RunExecutable("cargo", buildArgs...); err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		return "", fmt.Errorf("failed to run cargo build for target %s: %w", target, err)
	}

	// locate binary: target/<triple>/<profile>/<name>
	binaryPath := filepath.Join("target", target, config.CargoProfile, packageName)

	if config.Output != "" {
		renamedPath := fmt.Sprintf("%s-%s", config.Output, target)
		if err := utils.FileRename(binaryPath, renamedPath); err != nil {
			return "", fmt.Errorf("failed to rename binary: %w", err)
		}
		return renamedPath, nil
	}

	return binaryPath, nil
}
