package maven

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/piperutils"

	"github.com/SAP/jenkins-library/pkg/log"
)

// ExecuteOptions are used by Execute() to construct the Maven command line.
type ExecuteOptions struct {
	PomPath                     string   `json:"pomPath,omitempty"`
	ProjectSettingsFile         string   `json:"projectSettingsFile,omitempty"`
	GlobalSettingsFile          string   `json:"globalSettingsFile,omitempty"`
	M2Path                      string   `json:"m2Path,omitempty"`
	Goals                       []string `json:"goals,omitempty"`
	Defines                     []string `json:"defines,omitempty"`
	Flags                       []string `json:"flags,omitempty"`
	LogSuccessfulMavenTransfers bool     `json:"logSuccessfulMavenTransfers,omitempty"`
	ReturnStdout                bool     `json:"returnStdout,omitempty"`
	SkipGlobalSettings          bool     `json:"skipGlobalSettings,omitempty"`
}

// EvaluateOptions are used by Evaluate() to construct the Maven command line.
// In contrast to ExecuteOptions, fewer settings are required for Evaluate and thus a separate type is needed.
type EvaluateOptions struct {
	PomPath             string   `json:"pomPath,omitempty"`
	ProjectSettingsFile string   `json:"projectSettingsFile,omitempty"`
	GlobalSettingsFile  string   `json:"globalSettingsFile,omitempty"`
	M2Path              string   `json:"m2Path,omitempty"`
	Defines             []string `json:"defines,omitempty"`
}

type Utils interface {
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	RunExecutable(e string, p ...string) error

	DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error
	Glob(pattern string) (matches []string, err error)
	FileExists(filename string) (bool, error)
	Copy(src, dest string) (int64, error)
	MkdirAll(path string, perm os.FileMode) error
	FileWrite(path string, content []byte, perm os.FileMode) error
	FileRead(path string) ([]byte, error)
}

type utilsBundle struct {
	*command.Command
	*piperutils.Files
	*piperhttp.Client
}

func NewUtilsBundle() Utils {
	utils := utilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
		Client:  &piperhttp.Client{},
	}
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

const mavenExecutable = "mvn"

// Execute constructs a mvn command line from the given options, and uses the provided
// mavenExecRunner to execute it.
func Execute(options *ExecuteOptions, utils Utils) (string, error) {
	stdOutBuf, stdOut := evaluateStdOut(options)
	utils.Stdout(stdOut)
	utils.Stderr(log.Writer())

	parameters, err := getParametersFromOptions(options, utils)
	if err != nil {
		return "", fmt.Errorf("failed to construct parameters from options: %w", err)
	}

	err = utils.RunExecutable(mavenExecutable, parameters...)
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		commandLine := append([]string{mavenExecutable}, parameters...)
		return "", fmt.Errorf("failed to run executable, command: '%s', error: %w", commandLine, err)
	}

	if stdOutBuf == nil {
		return "", nil
	}
	return string(stdOutBuf.Bytes()), nil
}

// Evaluate constructs ExecuteOptions for using the maven-help-plugin's 'evaluate' goal to
// evaluate a given expression from a pom file. This allows to retrieve the value of - for
// example - 'project.version' from a pom file exactly as Maven itself evaluates it.
func Evaluate(options *EvaluateOptions, expression string, utils Utils) (string, error) {
	defines := []string{"-Dexpression=" + expression, "-DforceStdout", "-q"}
	defines = append(defines, options.Defines...)
	executeOptions := ExecuteOptions{
		PomPath:             options.PomPath,
		M2Path:              options.M2Path,
		ProjectSettingsFile: options.ProjectSettingsFile,
		GlobalSettingsFile:  options.GlobalSettingsFile,
		Goals:               []string{"org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate"},
		Defines:             defines,
		ReturnStdout:        true,
	}
	value, err := Execute(&executeOptions, utils)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(value, "null object or invalid expression") {
		return "", fmt.Errorf("expression '%s' in file '%s' could not be resolved", expression, options.PomPath)
	}
	return value, nil
}

func InstallModuleWithReactor(moduleName string, options *EvaluateOptions, utils Utils) error {
	var defines = options.Defines

	if !slices.Contains(defines, "-DskipTests") {
		defines = append(defines, "-DskipTests")
	}

	mavenOptionsInstall := ExecuteOptions{
		Goals:               []string{"install"},
		Flags:               []string{"-pl", moduleName, "-am"},
		Defines:             defines,
		M2Path:              options.M2Path,
		ProjectSettingsFile: options.ProjectSettingsFile,
		GlobalSettingsFile:  options.GlobalSettingsFile,
	}

	_, err := Execute(&mavenOptionsInstall, utils)
	if err != nil {
		return fmt.Errorf("failed to install maven artifacts: %w", err)
	}

	return nil
}

// InstallFile installs a maven artifact and its pom into the local maven repository.
// If "file" is empty, only the pom is installed. "pomFile" must not be empty.
func InstallFile(file, pomFile string, options *EvaluateOptions, utils Utils) error {
	if len(pomFile) == 0 {
		return fmt.Errorf("pomFile can't be empty")
	}

	var defines []string
	if len(file) > 0 {
		defines = append(defines, "-Dfile="+file)
		if strings.Contains(file, ".jar") {
			defines = append(defines, "-Dpackaging=jar")
		}
		if strings.Contains(file, "-classes") {
			defines = append(defines, "-Dclassifier=classes")
		}

	} else {
		defines = append(defines, "-Dfile="+pomFile)
	}
	defines = append(defines, "-DpomFile="+pomFile)
	mavenOptionsInstall := ExecuteOptions{
		Goals:               []string{"install:install-file"},
		Defines:             defines,
		M2Path:              options.M2Path,
		ProjectSettingsFile: options.ProjectSettingsFile,
		GlobalSettingsFile:  options.GlobalSettingsFile,
	}
	_, err := Execute(&mavenOptionsInstall, utils)
	if err != nil {
		return fmt.Errorf("failed to install maven artifacts: %w", err)
	}
	return nil
}

// InstallMavenArtifacts finds maven modules (identified by pom.xml files) and installs the artifacts into the local maven repository.
func InstallMavenArtifacts(options *EvaluateOptions, utils Utils) error {
	return doInstallMavenArtifacts(options, utils)
}

func doInstallMavenArtifacts(options *EvaluateOptions, utils Utils) error {
	err := flattenPom(options, utils)
	if err != nil {
		return err
	}

	pomFiles, err := utils.Glob(filepath.Join("**", "pom.xml"))
	if err != nil {
		return err
	}

	// Ensure m2 path is an absolute path, even if it is given relative
	// This is important to avoid getting multiple m2 directories in a maven multimodule project
	if options.M2Path != "" {
		options.M2Path, err = filepath.Abs(options.M2Path)
		if err != nil {
			return err
		}
	}

	for _, pomFile := range pomFiles {
		log.Entry().Info("Installing maven artifacts from module: " + pomFile)

		// Set this module's pom file as the pom file for evaluating the packaging,
		// otherwise we would evaluate the root pom in all iterations.
		evaluateProjectPackagingOptions := *options
		evaluateProjectPackagingOptions.PomPath = pomFile
		packaging, err := Evaluate(&evaluateProjectPackagingOptions, "project.packaging", utils)
		if err != nil {
			return err
		}

		currentModuleDir := filepath.Dir(pomFile)

		// Use flat pom if available to avoid issues with unresolved variables.
		pathToPomFile := pomFile
		flattenedPomExists, _ := utils.FileExists(filepath.Join(currentModuleDir, ".flattened-pom.xml"))
		if flattenedPomExists {
			pathToPomFile = filepath.Join(currentModuleDir, ".flattened-pom.xml")
		}

		if packaging == "pom" {
			err = InstallFile("", pathToPomFile, options, utils)
			if err != nil {
				return err
			}
		} else {

			err = installJarWarArtifacts(pathToPomFile, currentModuleDir, options, utils)
			if err != nil {
				return err
			}
		}
	}
	return err
}

func installJarWarArtifacts(pomFile, dir string, options *EvaluateOptions, utils Utils) error {
	options.PomPath = filepath.Join(dir, "pom.xml")
	finalName, err := Evaluate(options, "project.build.finalName", utils)
	if err != nil {
		return err
	}
	if finalName == "" {
		log.Entry().Warn("project.build.finalName is empty, skipping install of artifact. Installing only the pom file.")
		err = InstallFile("", pomFile, options, utils)
		if err != nil {
			return err
		}
		return nil
	}

	jarExists, _ := utils.FileExists(jarFile(dir, finalName))
	warExists, _ := utils.FileExists(warFile(dir, finalName))
	classesJarExists, _ := utils.FileExists(classesJarFile(dir, finalName))
	originalJarExists, _ := utils.FileExists(originalJarFile(dir, finalName))

	log.Entry().Infof("JAR file with name %s does exist: %t", jarFile(dir, finalName), jarExists)
	log.Entry().Infof("Classes-JAR file with name %s does exist: %t", classesJarFile(dir, finalName), classesJarExists)
	log.Entry().Infof("Original-JAR file with name %s does exist: %t", originalJarFile(dir, finalName), originalJarExists)
	log.Entry().Infof("WAR file with name %s does exist: %t", warFile(dir, finalName), warExists)

	// Due to spring's jar repackaging we need to check for an "original" jar file because the repackaged one is no suitable source for dependent maven modules
	if originalJarExists {
		err = InstallFile(originalJarFile(dir, finalName), pomFile, options, utils)
		if err != nil {
			return err
		}
	} else if jarExists {
		err = InstallFile(jarFile(dir, finalName), pomFile, options, utils)
		if err != nil {
			return err
		}
	}

	if warExists {
		err = InstallFile(warFile(dir, finalName), pomFile, options, utils)
		if err != nil {
			return err
		}
	}

	if classesJarExists {
		err = InstallFile(classesJarFile(dir, finalName), pomFile, options, utils)
		if err != nil {
			return err
		}
	}
	return nil
}

func jarFile(dir, finalName string) string {
	return filepath.Join(dir, "target", finalName+".jar")
}

func classesJarFile(dir, finalName string) string {
	return filepath.Join(dir, "target", finalName+"-classes.jar")
}

func originalJarFile(dir, finalName string) string {
	return filepath.Join(dir, "target", finalName+".jar.original")
}

func warFile(dir, finalName string) string {
	return filepath.Join(dir, "target", finalName+".war")
}

func flattenPom(options *EvaluateOptions, utils Utils) error {
	mavenOptionsFlatten := ExecuteOptions{
		Goals:               []string{"flatten:flatten"},
		Defines:             []string{"-Dflatten.mode=resolveCiFriendliesOnly"},
		PomPath:             options.PomPath,
		M2Path:              options.M2Path,
		ProjectSettingsFile: options.ProjectSettingsFile,
		GlobalSettingsFile:  options.GlobalSettingsFile,
	}
	_, err := Execute(&mavenOptionsFlatten, utils)
	return err
}

func evaluateStdOut(options *ExecuteOptions) (*bytes.Buffer, io.Writer) {
	var stdOutBuf *bytes.Buffer
	stdOut := log.Writer()
	if options.ReturnStdout {
		stdOutBuf = new(bytes.Buffer)
		stdOut = io.MultiWriter(stdOut, stdOutBuf)
	}
	return stdOutBuf, stdOut
}

func getParametersFromOptions(options *ExecuteOptions, utils Utils) ([]string, error) {
	var parameters []string

	parameters, err := DownloadAndGetMavenParameters(options.GlobalSettingsFile, options.ProjectSettingsFile, utils)
	if err != nil {
		return nil, err
	}

	if options.M2Path != "" {
		parameters = append(parameters, "-Dmaven.repo.local="+options.M2Path)
	}

	if options.PomPath != "" {
		parameters = append(parameters, "--file", options.PomPath)
	}

	if options.Flags != nil {
		parameters = append(parameters, options.Flags...)
	}

	if options.Defines != nil {
		parameters = append(parameters, options.Defines...)
	}

	if !options.LogSuccessfulMavenTransfers {
		parameters = append(parameters, "-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn")
	}

	parameters = append(parameters, "--batch-mode")

	parameters = append(parameters, options.Goals...)

	return parameters, nil
}

// GetTestModulesExcludes return testing modules that you be excluded from reactor
func GetTestModulesExcludes(utils Utils) []string {
	var excludes []string
	exists, _ := utils.FileExists("unit-tests/pom.xml")
	if exists {
		excludes = append(excludes, "-pl", "!unit-tests")
	}
	exists, _ = utils.FileExists("integration-tests/pom.xml")
	if exists {
		excludes = append(excludes, "-pl", "!integration-tests")
	}
	return excludes
}
