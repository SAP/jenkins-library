package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/SAP/jenkins-library/pkg/build"
	"github.com/SAP/jenkins-library/pkg/buildsettings"
	"github.com/SAP/jenkins-library/pkg/certutils"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/goget"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"

	"github.com/SAP/jenkins-library/pkg/multiarch"
	"github.com/SAP/jenkins-library/pkg/versioning"

	"golang.org/x/mod/modfile"
)

const (
	coverageFile                = "cover.out"
	golangUnitTestOutput        = "TEST-go.xml"
	golangIntegrationTestOutput = "TEST-integration.xml"
	unitJsonReport              = "unit-report.out"
	integrationJsonReport       = "integration-report.out"
	golangCoberturaPackage      = "github.com/boumenot/gocover-cobertura@latest"
	golangTestsumPackage        = "gotest.tools/gotestsum@latest"
	golangCycloneDXPackage      = "github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@v1.9.0"
	sbomFilename                = "bom-golang.xml"
)

type golangBuildUtils interface {
	command.ExecRunner
	goget.Client

	piperutils.FileUtils
	piperhttp.Uploader

	getDockerImageValue(stepName string) (string, error)
	GetExitCode() int
	DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error
	Untar(src string, dest string, stripComponentLevel int) error

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The golangBuildUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type golangBuildUtilsBundle struct {
	*command.Command
	*piperutils.Files
	piperhttp.Uploader
	httpClient *piperhttp.Client

	goget.Client

	// Embed more structs as necessary to implement methods or interfaces you add to golangBuildUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// golangBuildUtilsBundle and forward to the implementation of the dependency.
}

func (g *golangBuildUtilsBundle) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	return g.httpClient.DownloadFile(url, filename, header, cookies)
}

func (g *golangBuildUtilsBundle) getDockerImageValue(stepName string) (string, error) {
	return GetDockerImageValue(stepName)
}

func (g *golangBuildUtilsBundle) Untar(src string, dest string, stripComponentLevel int) error {
	return piperutils.Untar(src, dest, stripComponentLevel)
}

func newGolangBuildUtils(config golangBuildOptions) golangBuildUtils {
	httpClientOptions := piperhttp.ClientOptions{}

	if len(config.CustomTLSCertificateLinks) > 0 {
		httpClientOptions.TransportSkipVerification = false
		httpClientOptions.TrustedCerts = config.CustomTLSCertificateLinks
	}

	httpClient := piperhttp.Client{}
	httpClient.SetOptions(httpClientOptions)

	utils := golangBuildUtilsBundle{
		Command: &command.Command{
			StepName: "golangBuild",
		},
		Files:    &piperutils.Files{},
		Uploader: &httpClient,
		Client: &goget.ClientImpl{
			HTTPClient: &httpClient,
		},
		httpClient: &httpClient,
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func golangBuild(config golangBuildOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *golangBuildCommonPipelineEnvironment) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newGolangBuildUtils(config)

	// Error situations will be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runGolangBuild(&config, telemetryData, utils, commonPipelineEnvironment)
	if err != nil {
		log.Entry().WithError(err).Fatal("execution of golang build failed")
	}
}

func runGolangBuild(config *golangBuildOptions, telemetryData *telemetry.CustomData, utils golangBuildUtils, commonPipelineEnvironment *golangBuildCommonPipelineEnvironment) error {
	goModFile, err := readGoModFile(utils) // returns nil if go.mod doesnt exist
	if err != nil {
		return err
	}

	if err = prepareGolangEnvironment(config, goModFile, utils); err != nil {
		return err
	}

	// install test pre-requisites only in case testing should be performed
	if config.RunTests || config.RunIntegrationTests {
		if err := utils.RunExecutable("go", "install", golangTestsumPackage); err != nil {
			return fmt.Errorf("failed to install pre-requisite: %w", err)
		}
	}

	if config.CreateBOM {
		if err := utils.RunExecutable("go", "install", golangCycloneDXPackage); err != nil {
			return fmt.Errorf("failed to install pre-requisite: %w", err)
		}
	}

	failedTests := false

	if config.RunTests {
		success, err := runGolangTests(config, utils)
		if err != nil {
			return err
		}
		failedTests = !success
	}

	if config.RunTests && config.ReportCoverage {
		if err := reportGolangTestCoverage(config, utils); err != nil {
			return err
		}
	}

	if config.RunIntegrationTests {
		success, err := runGolangIntegrationTests(config, utils)
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
		goPath := os.Getenv("GOPATH")
		golangciLintDir := filepath.Join(goPath, "bin")

		if err := retrieveGolangciLint(utils, golangciLintDir, config.GolangciLintURL); err != nil {
			return err
		}

		// hardcode those for now
		lintSettings := map[string]string{
			"reportStyle":      "checkstyle", // readable by Sonar
			"reportOutputPath": "golangci-lint-report.xml",
			"additionalParams": "",
		}

		if err := runGolangciLint(utils, golangciLintDir, config.FailOnLintingError, lintSettings); err != nil {
			return err
		}
	}

	if config.CreateBOM {
		if err := runBOMCreation(utils, sbomFilename); err != nil {
			return err
		}
	}

	ldflags := ""

	if len(config.LdflagsTemplate) > 0 {
		ldf, err := prepareLdflags(config, utils, GeneralConfig.EnvRootPath)
		if err != nil {
			return err
		}
		ldflags = (*ldf).String()
		log.Entry().Infof("ldflags from template: '%v'", ldflags)
	}

	var binaries []string
	platforms, err := multiarch.ParsePlatformStrings(config.TargetArchitectures)
	if err != nil {
		return err
	}

	for _, platform := range platforms {
		binaryNames, err := runGolangBuildPerArchitecture(config, goModFile, utils, ldflags, platform)
		if err != nil {
			return err
		}

		if len(binaryNames) > 0 {
			binaries = append(binaries, binaryNames...)
		}
	}

	log.Entry().Debugf("creating build settings information...")
	stepName := "golangBuild"
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

		if len(artifactVersion) == 0 {
			artifactOpts := versioning.Options{
				VersioningScheme: "library",
			}

			artifact, err := versioning.GetArtifact("golang", "", &artifactOpts, utils)
			if err != nil {
				return err
			}

			artifactVersion, err = artifact.GetVersion()
			if err != nil {
				return err
			}
		}

		if goModFile == nil {
			return fmt.Errorf("go.mod file not found")
		} else if goModFile.Module == nil {
			return fmt.Errorf("go.mod doesn't declare a module path")
		}

		repoClientOptions := piperhttp.ClientOptions{
			Username:     config.TargetRepositoryUser,
			Password:     config.TargetRepositoryPassword,
			TrustedCerts: config.CustomTLSCertificateLinks,
		}

		utils.SetOptions(repoClientOptions)

		var binaryArtifacts piperenv.Artifacts
		buildCoordinates := []versioning.Coordinates{}

		for _, binary := range binaries {

			targetPath := fmt.Sprintf("go/%s/%s/%s", goModFile.Module.Mod.Path, artifactVersion, binary)

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

			if !(response.StatusCode == 200 || response.StatusCode == 201) {
				return fmt.Errorf("couldn't upload artifact, received status code %d", response.StatusCode)
			}

			binaryArtifacts = append(binaryArtifacts, piperenv.Artifact{
				Name: binary,
			})

			if config.CreateBuildArtifactsMetadata {
				err, coordinate := createGoBuildArtifactsMetadata(binary, config.TargetRepositoryURL, artifactVersion, utils)
				if err != nil {
					log.Entry().Warnf("unable to create build artifact metadata : %v", err)
				}
				buildCoordinates = append(buildCoordinates, coordinate)
			}
		}
		commonPipelineEnvironment.custom.artifacts = binaryArtifacts

		if len(buildCoordinates) == 0 {
			log.Entry().Warnf("unable to identify artifact coordinates for the go binary(s) published")
			return nil
		}

		var buildArtifacts build.BuildArtifacts
		buildArtifacts.Coordinates = buildCoordinates
		jsonResult, _ := json.Marshal(buildArtifacts)
		commonPipelineEnvironment.custom.goBuildArtifacts = string(jsonResult)

	}

	return nil
}

func createGoBuildArtifactsMetadata(binary string, repositoryURL string, artifactVersion string, utils golangBuildUtils) (error, versioning.Coordinates) {
	options := versioning.Options{}
	builtArtifact, err := versioning.GetArtifact("golang", "", &options, utils)
	coordinate, err := builtArtifact.GetCoordinates()
	purl := piperutils.GetPurl(filepath.Join(filepath.Dir("go.mod"), sbomFilename))
	// golang purls contain the hex code for & with GOOS and GOARC and should be reomved from the PURL
	purl = strings.ReplaceAll(purl, "\\u0026", "&")
	if err != nil {
		return err, coordinate
	}
	coordinate.ArtifactID = binary
	coordinate.URL = repositoryURL
	coordinate.BuildPath = filepath.Dir(binary)
	coordinate.PURL = purl
	coordinate.Version = artifactVersion

	return nil, coordinate
}

func prepareGolangEnvironment(config *golangBuildOptions, goModFile *modfile.File, utils golangBuildUtils) error {
	// configure truststore
	err := certutils.CertificateUpdate(config.CustomTLSCertificateLinks, utils, utils, "/etc/ssl/certs/ca-certificates.crt") // TODO reimplement

	if config.PrivateModules == "" {
		return nil
	}

	if config.PrivateModulesGitToken == "" {
		return fmt.Errorf("please specify a token for fetching private git modules")
	}

	// pass private repos to go process
	os.Setenv("GOPRIVATE", config.PrivateModules)

	err = gitConfigurationForPrivateModules(config.PrivateModules, config.PrivateModulesGitToken, utils)
	if err != nil {
		return err
	}

	return nil
}

func runGolangTests(config *golangBuildOptions, utils golangBuildUtils) (bool, error) {
	// execute gotestsum in order to have more output options
	testOptions := []string{"--junitfile", golangUnitTestOutput, "--jsonfile", unitJsonReport, "--", fmt.Sprintf("-coverprofile=%v", coverageFile), "-tags=unit", "./..."}
	testOptions = append(testOptions, config.TestOptions...)
	if err := utils.RunExecutable("gotestsum", testOptions...); err != nil {
		exists, fileErr := utils.FileExists(golangUnitTestOutput)
		if !exists || fileErr != nil {
			log.SetErrorCategory(log.ErrorBuild)
			return false, fmt.Errorf("running tests failed - junit result missing: %w", err)
		}
		exists, fileErr = utils.FileExists(coverageFile)
		if !exists || fileErr != nil {
			log.SetErrorCategory(log.ErrorBuild)
			return false, fmt.Errorf("running tests failed - coverage output missing: %w", err)
		}
		return false, nil
	}
	return true, nil
}

func runGolangIntegrationTests(config *golangBuildOptions, utils golangBuildUtils) (bool, error) {
	// execute gotestsum in order to have more output options
	// for integration tests coverage data is not meaningful and thus not being created
	if err := utils.RunExecutable("gotestsum", "--junitfile", golangIntegrationTestOutput, "--jsonfile", integrationJsonReport, "--", "-tags=integration", "./..."); err != nil {
		exists, fileErr := utils.FileExists(golangIntegrationTestOutput)
		if !exists || fileErr != nil {
			log.SetErrorCategory(log.ErrorBuild)
			return false, fmt.Errorf("running tests failed: %w", err)
		}
		return false, nil
	}
	return true, nil
}

func reportGolangTestCoverage(config *golangBuildOptions, utils golangBuildUtils) error {
	if config.CoverageFormat == "cobertura" {
		// execute gocover-cobertura in order to create cobertura report
		// install pre-requisites
		if err := utils.RunExecutable("go", "install", golangCoberturaPackage); err != nil {
			return fmt.Errorf("failed to install pre-requisite: %w", err)
		}

		coverageData, err := utils.FileRead(coverageFile)
		if err != nil {
			return fmt.Errorf("failed to read coverage file %v: %w", coverageFile, err)
		}
		utils.Stdin(bytes.NewBuffer(coverageData))

		coverageOutput := bytes.Buffer{}
		utils.Stdout(&coverageOutput)
		options := []string{}
		if config.ExcludeGeneratedFromCoverage {
			options = append(options, "-ignore-gen-files")
		}
		if err := utils.RunExecutable("gocover-cobertura", options...); err != nil {
			log.SetErrorCategory(log.ErrorTest)
			return fmt.Errorf("failed to convert coverage data to cobertura format: %w", err)
		}
		utils.Stdout(log.Writer())

		err = utils.FileWrite("cobertura-coverage.xml", coverageOutput.Bytes(), 0o666)
		if err != nil {
			return fmt.Errorf("failed to create cobertura coverage file: %w", err)
		}
		log.Entry().Info("created file cobertura-coverage.xml")
	} else {
		// currently only cobertura and html format supported, thus using html as fallback
		if err := utils.RunExecutable("go", "tool", "cover", "-html", coverageFile, "-o", "coverage.html"); err != nil {
			return fmt.Errorf("failed to create html coverage file: %w", err)
		}
	}
	return nil
}

func retrieveGolangciLint(utils golangBuildUtils, golangciLintDir, golangciLintURL string) error {
	archiveName := "golangci-lint.tar.gz"
	err := utils.DownloadFile(golangciLintURL, archiveName, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to download golangci-lint: %w", err)
	}

	err = utils.Untar(archiveName, golangciLintDir, 1)
	if err != nil {
		return fmt.Errorf("failed to install golangci-lint: %w", err)
	}

	return nil
}

func runGolangciLint(utils golangBuildUtils, golangciLintDir string, failOnError bool, lintSettings map[string]string) error {
	binaryPath := filepath.Join(golangciLintDir, "golangci-lint")

	var outputBuffer bytes.Buffer
	utils.Stdout(&outputBuffer)
	err := utils.RunExecutable(binaryPath, "run", "--out-format", lintSettings["reportStyle"])
	if err != nil && utils.GetExitCode() != 1 {
		return fmt.Errorf("running golangci-lint failed: %w", err)
	}

	log.Entry().Infof("lint report: \n%s", outputBuffer.String())
	log.Entry().Infof("writing lint report to %s", lintSettings["reportOutputPath"])
	err = utils.FileWrite(lintSettings["reportOutputPath"], outputBuffer.Bytes(), 0o644)
	if err != nil {
		return fmt.Errorf("writing golangci-lint report failed: %w", err)
	}

	if utils.GetExitCode() == 1 && failOnError {
		return fmt.Errorf("golangci-lint found issues, see report above")
	}

	return nil
}

func prepareLdflags(config *golangBuildOptions, utils golangBuildUtils, envRootPath string) (*bytes.Buffer, error) {
	cpe := piperenv.CPEMap{}
	err := cpe.LoadFromDisk(path.Join(envRootPath, "commonPipelineEnvironment"))
	if err != nil {
		log.Entry().Warning("failed to load values from commonPipelineEnvironment")
	}

	log.Entry().Debugf("ldflagsTemplate in use: %v", config.LdflagsTemplate)
	return cpe.ParseTemplate(config.LdflagsTemplate)
}

func runGolangBuildPerArchitecture(config *golangBuildOptions, goModFile *modfile.File, utils golangBuildUtils, ldflags string, architecture multiarch.Platform) ([]string, error) {
	var binaryNames []string

	envVars := os.Environ()
	envVars = append(envVars, fmt.Sprintf("GOOS=%v", architecture.OS), fmt.Sprintf("GOARCH=%v", architecture.Arch))

	if !config.CgoEnabled {
		envVars = append(envVars, "CGO_ENABLED=0")
	}
	utils.SetEnv(envVars)

	buildOptions := []string{"build", "-trimpath"}

	if len(config.Output) > 0 {
		if len(config.Packages) > 1 {
			binaries, outputDir, err := getOutputBinaries(config.Output, config.Packages, utils, architecture)
			if err != nil {
				log.SetErrorCategory(log.ErrorBuild)
				return nil, fmt.Errorf("failed to calculate output binaries or directory, error: %s", err.Error())
			}
			buildOptions = append(buildOptions, "-o", outputDir)
			binaryNames = append(binaryNames, binaries...)
		} else {
			fileExtension := ""
			if architecture.OS == "windows" {
				fileExtension = ".exe"
			}
			binaryName := fmt.Sprintf("%s-%s.%s%s", strings.TrimRight(config.Output, string(os.PathSeparator)), architecture.OS, architecture.Arch, fileExtension)
			buildOptions = append(buildOptions, "-o", binaryName)
			binaryNames = append(binaryNames, binaryName)
		}
	} else {
		// use default name in case no name is defined via Output
		binaryName := path.Base(goModFile.Module.Mod.Path)
		binaryNames = append(binaryNames, binaryName)
	}
	buildOptions = append(buildOptions, config.BuildFlags...)
	if len(ldflags) > 0 {
		buildOptions = append(buildOptions, "-ldflags", ldflags)
	}
	buildOptions = append(buildOptions, config.Packages...)

	if err := utils.RunExecutable("go", buildOptions...); err != nil {
		log.Entry().Debugf("buildOptions: %v", buildOptions)
		log.SetErrorCategory(log.ErrorBuild)
		return nil, fmt.Errorf("failed to run build for %v.%v: %w", architecture.OS, architecture.Arch, err)
	}

	return binaryNames, nil
}

func runBOMCreation(utils golangBuildUtils, outputFilename string) error {
	if err := utils.RunExecutable("cyclonedx-gomod", "mod", "-licenses", fmt.Sprintf("-verbose=%t", GeneralConfig.Verbose), "-test", "-output", outputFilename, "-output-version", "1.4"); err != nil {
		return fmt.Errorf("BOM creation failed: %w", err)
	}
	return nil
}

func readGoModFile(utils golangBuildUtils) (*modfile.File, error) {
	modFilePath := "go.mod"

	if modFileExists, err := utils.FileExists(modFilePath); err != nil {
		return nil, err
	} else if !modFileExists {
		return nil, nil
	}

	modFileContent, err := utils.FileRead(modFilePath)
	if err != nil {
		return nil, err
	}

	return modfile.Parse(modFilePath, modFileContent, nil)
}

func getOutputBinaries(out string, packages []string, utils golangBuildUtils, architecture multiarch.Platform) ([]string, string, error) {
	var binaries []string
	outDir := fmt.Sprintf("%s-%s-%s%c", strings.TrimRight(out, string(os.PathSeparator)), architecture.OS, architecture.Arch, os.PathSeparator)

	for _, pkg := range packages {
		ok, err := isMainPackage(utils, pkg)
		if err != nil {
			return nil, "", err
		}

		if ok {
			fileExt := ""
			if architecture.OS == "windows" {
				fileExt = ".exe"
			}
			binaries = append(binaries, filepath.Join(outDir, filepath.Base(pkg)+fileExt))
		}
	}

	return binaries, outDir, nil
}

func isMainPackage(utils golangBuildUtils, pkg string) (bool, error) {
	outBuffer := bytes.NewBufferString("")
	utils.Stdout(outBuffer)
	utils.Stderr(outBuffer)
	err := utils.RunExecutable("go", "list", "-f", "{{ .Name }}", pkg)
	if err != nil {
		return false, fmt.Errorf("%w: %s", err, outBuffer.String())
	}

	if outBuffer.String() != "main" {
		return false, nil
	}

	return true, nil
}

func gitConfigurationForPrivateModules(privateMod string, token string, utils golangBuildUtils) error {
	privateMod = strings.ReplaceAll(privateMod, "/*", "")
	privateMod = strings.ReplaceAll(privateMod, "*.", "")
	modules := strings.Split(privateMod, ",")
	for _, v := range modules {
		authenticatedRepoURL := fmt.Sprintf("https://%s@%s", token, v)
		repoBaseURL := fmt.Sprintf("https://%s", v)
		err := utils.RunExecutable("git", "config", "--global", fmt.Sprintf("url.%s.insteadOf", authenticatedRepoURL), repoBaseURL)
		if err != nil {
			return err
		}

	}

	return nil
}
