package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

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
	"golang.org/x/mod/module"
)

const (
	coverageFile                = "cover.out"
	golangUnitTestOutput        = "TEST-go.xml"
	golangIntegrationTestOutput = "TEST-integration.xml"
	golangCoberturaPackage      = "github.com/boumenot/gocover-cobertura@latest"
	golangTestsumPackage        = "gotest.tools/gotestsum@latest"
	golangCycloneDXPackage      = "github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest"
	sbomFilename                = "bom.xml"
	golangciLintCurlUrl         = "https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh"
	golangciLintVersion         = "v1.46.2"
)

type golangBuildUtils interface {
	command.ExecRunner
	goget.Client

	piperutils.FileUtils
	piperhttp.Uploader

	DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error
	getDockerImageValue(stepName string) (string, error)

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The golangBuildUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type golangBuildUtilsBundle struct {
	*command.Command
	*piperutils.Files
	piperhttp.Uploader

	goget.Client

	// Embed more structs as necessary to implement methods or interfaces you add to golangBuildUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// golangBuildUtilsBundle and forward to the implementation of the dependency.
}

func (g *golangBuildUtilsBundle) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	return fmt.Errorf("not implemented")
}

func (g *golangBuildUtilsBundle) getDockerImageValue(stepName string) (string, error) {
	return GetDockerImageValue(stepName)
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
		Command:  &command.Command{},
		Files:    &piperutils.Files{},
		Uploader: &httpClient,
		Client: &goget.ClientImpl{
			HTTPClient: &httpClient,
		},
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

	if config.RunLinter {
		goPath := os.Getenv("GOPATH")
		golangciLintDir := filepath.Join(goPath, "bin")

		if err := retrieveGolangciLint(golangciLintDir); err != nil {
			return err
		}

		if err := runGolangciLint(golangciLintDir); err != nil {
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
		var err error
		ldflags, err = prepareLdflags(config, utils, GeneralConfig.EnvRootPath)
		if err != nil {
			return err
		}
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
		for _, binary := range binaries {

			targetPath := fmt.Sprintf("go/%s/%s/%s", goModFile.Module.Mod.Path, config.ArtifactVersion, binary)

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
		}
		commonPipelineEnvironment.custom.artifacts = binaryArtifacts

	}

	return nil
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

	repoURLs, err := lookupGolangPrivateModulesRepositories(goModFile, config.PrivateModules, utils)

	if err != nil {
		return err
	}

	// configure credentials git shall use for pulling repos
	for _, repoURL := range repoURLs {
		if match, _ := regexp.MatchString("(?i)^https?://", repoURL); !match {
			continue
		}

		authenticatedRepoURL := strings.Replace(repoURL, "://", fmt.Sprintf("://%s@", config.PrivateModulesGitToken), 1)

		err = utils.RunExecutable("git", "config", "--global", fmt.Sprintf("url.%s.insteadOf", authenticatedRepoURL), fmt.Sprintf("%s", repoURL))
		if err != nil {
			return err
		}
	}

	return nil
}

func runGolangTests(config *golangBuildOptions, utils golangBuildUtils) (bool, error) {
	// execute gotestsum in order to have more output options
	if err := utils.RunExecutable("gotestsum", "--junitfile", golangUnitTestOutput, "--", fmt.Sprintf("-coverprofile=%v", coverageFile), "./..."); err != nil {
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
	if err := utils.RunExecutable("gotestsum", "--junitfile", golangIntegrationTestOutput, "--", "-tags=integration", "./..."); err != nil {
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

		err = utils.FileWrite("cobertura-coverage.xml", coverageOutput.Bytes(), 0666)
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

func retrieveGolangciLint(golangciLintDir string) error {
	// from installation instructions: https://golangci-lint.run/usage/install/#linux-and-windows
	installationScript, err := exec.Command("curl", "-sSfL", golangciLintCurlUrl).Output()
	if err != nil {
		return fmt.Errorf("failed to install golangci-lint: curl command failed: %w", err)
	}

	cmd := exec.Command("sh", "-s", "--", "-b", golangciLintDir, golangciLintVersion)
	installationScriptBuffer := bytes.Buffer{}
	installationScriptBuffer.Write(installationScript)
	cmd.Stdin = &installationScriptBuffer
	cmdOutput, err := cmd.CombinedOutput()

	log.Entry().Infof(string(cmdOutput))
	if err != nil {
		return fmt.Errorf("failed to install golangci-lint: %w", err)
	}

	return nil
}

func runGolangciLint(golangciLintDir string) error {
	binaryPath := filepath.Join(golangciLintDir, "golangci-lint")
	reportOutputPath := "golangci-lint-report.xml"
	reportStyle := "checkstyle" // readable by Sonar

	lintRunCommand := fmt.Sprintf("%s run --out-format %s > %s", binaryPath, reportStyle, reportOutputPath)
	log.Entry().Infof("running command: %s", lintRunCommand)
	lintRunOutput, err := exec.Command("bash", "-c", lintRunCommand).CombinedOutput()
	log.Entry().Infof(string(lintRunOutput))

	// exit status 1 is returned when linter found issues, but ran fine
	if err != nil && err.Error() != "exit status 1" {
		return fmt.Errorf("running golangci-lint failed: %w", err)
	}

	lintReport, err := ioutil.ReadFile(reportOutputPath)
	if err != nil {
		return fmt.Errorf("running golangci-lint failed: couldn't read lint report: %w", err)
	}
	log.Entry().Infof("lint report: \n" + string(lintReport))

	return nil
}

func prepareLdflags(config *golangBuildOptions, utils golangBuildUtils, envRootPath string) (string, error) {
	cpe := piperenv.CPEMap{}
	err := cpe.LoadFromDisk(path.Join(envRootPath, "commonPipelineEnvironment"))
	if err != nil {
		log.Entry().Warning("failed to load values from commonPipelineEnvironment")
	}

	log.Entry().Debugf("ldflagsTemplate in use: %v", config.LdflagsTemplate)
	tmpl, err := template.New("ldflags").Parse(config.LdflagsTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse ldflagsTemplate '%v': %w", config.LdflagsTemplate, err)
	}

	ldflagsParams := struct {
		CPE map[string]interface{}
	}{
		CPE: map[string]interface{}(cpe),
	}
	var generatedLdflags bytes.Buffer
	err = tmpl.Execute(&generatedLdflags, ldflagsParams)
	if err != nil {
		return "", fmt.Errorf("failed to execute ldflagsTemplate '%v': %w", config.LdflagsTemplate, err)
	}

	return generatedLdflags.String(), nil
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

// lookupPrivateModulesRepositories returns a slice of all modules that match the given glob pattern
func lookupGolangPrivateModulesRepositories(goModFile *modfile.File, globPattern string, utils golangBuildUtils) ([]string, error) {
	if globPattern == "" {
		return []string{}, nil
	}

	if goModFile == nil {
		return nil, fmt.Errorf("couldn't find go.mod file")
	} else if goModFile.Require == nil {
		return []string{}, nil // no modules referenced, nothing to do
	}

	privateModules := []string{}

	for _, goModule := range goModFile.Require {
		if !module.MatchPrefixPatterns(globPattern, goModule.Mod.Path) {
			continue
		}

		repo, err := utils.GetRepositoryURL(goModule.Mod.Path)

		if err != nil {
			return nil, err
		}

		privateModules = append(privateModules, repo)
	}
	return privateModules, nil
}

func runBOMCreation(utils golangBuildUtils, outputFilename string) error {
	if err := utils.RunExecutable("cyclonedx-gomod", "mod", "-licenses", "-test", "-output", outputFilename); err != nil {
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
		return false, err
	}

	if outBuffer.String() != "main" {
		return false, nil
	}

	return true, nil
}
