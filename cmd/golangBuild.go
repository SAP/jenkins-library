package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

const coverageFile = "cover.out"

type golangBuildUtils interface {
	command.ExecRunner

	FileExists(filename string) (bool, error)
	FileRead(path string) ([]byte, error)
	FileWrite(path string, content []byte, perm os.FileMode) error

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The golangBuildUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type golangBuildUtilsBundle struct {
	*command.Command
	*piperutils.Files

	// Embed more structs as necessary to implement methods or interfaces you add to golangBuildUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// golangBuildUtilsBundle and forward to the implementation of the dependency.
}

func newGolangBuildUtils() golangBuildUtils {
	utils := golangBuildUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func golangBuild(config golangBuildOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newGolangBuildUtils()

	// Error situations will be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runGolangBuild(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("execution of golang build failed")
	}
}

func runGolangBuild(config *golangBuildOptions, telemetryData *telemetry.CustomData, utils golangBuildUtils) error {

	// install test pre-requisites only in case testing should be performed
	if config.RunTests || config.RunIntegrationTests {
		if err := utils.RunExecutable("go", "install", "gotest.tools/gotestsum"); err != nil {
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
		failedTests = !success
	}

	if failedTests {
		log.SetErrorCategory(log.ErrorTest)
		return fmt.Errorf("some tests failed")
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

	for _, architecture := range config.TargetArchitectures {
		err := runGolangBuildPerArchitecture(config, utils, ldflags, architecture)
		if err != nil {
			return err
		}
	}

	return nil
}

func runGolangTests(config *golangBuildOptions, utils golangBuildUtils) (bool, error) {
	// execute gotestsum in order to have more output options
	if err := utils.RunExecutable("gotestsum", "--junitfile", "TEST-go.xml", "--", fmt.Sprintf("-coverprofile=%v", coverageFile), "./..."); err != nil {
		exists, fileErr := utils.FileExists("TEST-go.xml")
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
	if err := utils.RunExecutable("gotestsum", "--junitfile", "TEST-integration.xml", "--", "-tags=integration", "./..."); err != nil {
		exists, fileErr := utils.FileExists("TEST-integration.xml")
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
		if err := utils.RunExecutable("go", "install", "github.com/boumenot/gocover-cobertura"); err != nil {
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

func runGolangBuildPerArchitecture(config *golangBuildOptions, utils golangBuildUtils, ldflags, architecture string) error {
	envVars := os.Environ()
	goos, goarch := splitTargetArchitecture(architecture)
	envVars = append(envVars, fmt.Sprintf("GOOS=%v", goos), fmt.Sprintf("GOARCH=%v", goarch))

	if !config.CgoEnabled {
		envVars = append(envVars, "CGO_ENABLED=0")
	}
	utils.SetEnv(envVars)

	buildOptions := []string{"build"}
	if len(config.Output) > 0 {
		fileExtension := ""
		if goos == "windows" {
			fileExtension = ".exe"
		}
		buildOptions = append(buildOptions, "-o", fmt.Sprintf("%v-%v.%v%v", config.Output, goos, goarch, fileExtension))
	}
	buildOptions = append(buildOptions, config.BuildFlags...)
	buildOptions = append(buildOptions, config.Packages...)
	if len(ldflags) > 0 {
		buildOptions = append(buildOptions, "-ldflags", ldflags)
	}

	if err := utils.RunExecutable("go", buildOptions...); err != nil {
		log.Entry().Debugf("buildOptions: %v", buildOptions)
		log.SetErrorCategory(log.ErrorBuild)
		return fmt.Errorf("failed to run build for %v.%v: %w", goos, goarch, err)
	}
	return nil
}

func splitTargetArchitecture(architecture string) (string, string) {
	// architecture expected to be in format os,arch due to possibleValues check of step

	architectureParts := strings.Split(architecture, ",")
	return architectureParts[0], architectureParts[1]
}
