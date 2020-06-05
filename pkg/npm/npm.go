package npm

import (
	"bytes"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	FileUtils "github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/bmatcuk/doublestar"
	"io"
	"os"
	"strings"
)

// RegistryOptions holds the configured urls for npm registries
type RegistryOptions struct {
	DefaultNpmRegistry string
	SapNpmRegistry     string
}

type ExecRunner interface {
	Stdout(out io.Writer)
	RunExecutable(executable string, params ...string) error
}

type NpmUtils interface {
	FileExists(path string) (bool, error)
	FileRead(path string) ([]byte , error)
	Glob(pattern string) (matches []string, err error)
	Getwd() (dir string, err error)
	Chdir(dir string) error
	GetExecRunner() ExecRunner
}

type NpmUtilsBundle struct {
	projectStructure FileUtils.ProjectStructure
	fileUtils        FileUtils.Files
	execRunner       *command.Command
}

func (u *NpmUtilsBundle) FileExists(path string) (bool, error) {
	return u.fileUtils.FileExists(path)
}

func (u *NpmUtilsBundle) FileRead(path string) ([]byte, error) {
	return u.fileUtils.FileRead(path)
}

func (u *NpmUtilsBundle) Glob(pattern string) (matches []string, err error) {
	return doublestar.Glob(pattern)
}

func (u *NpmUtilsBundle) Getwd() (dir string, err error) {
	return os.Getwd()
}

func (u *NpmUtilsBundle) Chdir(dir string) error {
	return os.Chdir(dir)
}

func (u *NpmUtilsBundle) GetExecRunner() ExecRunner {
	if u.execRunner == nil {
		u.execRunner = &command.Command{}
		u.execRunner.Stdout(log.Writer())
		u.execRunner.Stderr(log.Writer())
	}
	return u.execRunner
}

// SetNpmRegistries configures the given npm registries.
// CAUTION: This will change the npm configuration in the user's home directory.
func SetNpmRegistries(options *RegistryOptions, execRunner ExecRunner) error {
	const sapRegistry = "@sap:registry"
	const npmRegistry = "registry"
	configurableRegistries := []string{npmRegistry, sapRegistry}
	for _, registry := range configurableRegistries {
		var buffer bytes.Buffer
		execRunner.Stdout(&buffer)
		err := execRunner.RunExecutable("npm", "config", "get", registry)
		execRunner.Stdout(log.Writer())
		if err != nil {
			return err
		}
		preConfiguredRegistry := buffer.String()

		if registryIsNonEmpty(preConfiguredRegistry) {
			log.Entry().Info("Discovered pre-configured npm registry " + registry + " with value " + preConfiguredRegistry)
		}

		if registry == npmRegistry && options.DefaultNpmRegistry != "" && registryRequiresConfiguration(preConfiguredRegistry, "https://registry.npmjs.org") {
			log.Entry().Info("npm registry " + registry + " was not configured, setting it to " + options.DefaultNpmRegistry)
			err = execRunner.RunExecutable("npm", "config", "set", registry, options.DefaultNpmRegistry)
			if err != nil {
				return err
			}
		}

		if registry == sapRegistry && options.SapNpmRegistry != "" && registryRequiresConfiguration(preConfiguredRegistry, "https://npm.sap.com") {
			log.Entry().Info("npm registry " + registry + " was not configured, setting it to " + options.SapNpmRegistry)
			err = execRunner.RunExecutable("npm", "config", "set", registry, options.SapNpmRegistry)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func registryIsNonEmpty(preConfiguredRegistry string) bool {
	return !strings.HasPrefix(preConfiguredRegistry, "undefined") && len(preConfiguredRegistry) > 0
}

func registryRequiresConfiguration(preConfiguredRegistry, url string) bool {
	return strings.HasPrefix(preConfiguredRegistry, "undefined") || strings.HasPrefix(preConfiguredRegistry, url)
}

func CheckIfLockFilesExist(utils NpmUtils) (bool, bool, error) {
	packageLockExists, err := utils.FileExists("package-lock.json")

	if err != nil {
		return false, false, err
	}
	yarnLockExists, err := utils.FileExists("yarn.lock")
	if err != nil {
		return false, false, err
	}
	return packageLockExists, yarnLockExists, nil
}

func InstallDependencies(dir string, packageLockExists bool, yarnLockExists bool, execRunner ExecRunner) (err error) {
	log.Entry().WithField("WorkingDirectory", dir).Info("Running install")
	if packageLockExists {
		err = execRunner.RunExecutable("npm", "ci")
		if err != nil {
			return err
		}
	} else if yarnLockExists {
		err = execRunner.RunExecutable("yarn", "install", "--frozen-lockfile")
		if err != nil {
			return err
		}
	} else {
		log.Entry().Warn("No package lock file found. " +
			"It is recommended to create a `package-lock.json` file by running `npm install` locally." +
			" Add this file to your version control. " +
			"By doing so, the builds of your application become more reliable.")
		err = execRunner.RunExecutable("npm", "install")
		if err != nil {
			return err
		}
	}
	return nil
}

func FindPackageJSONFiles(utils NpmUtils) ([]string, error) {
	unfilteredListOfPackageJSONFiles, err := utils.Glob("**/package.json")
	if err != nil {
		return nil, err
	}

	var packageJSONFiles []string

	for _, file := range unfilteredListOfPackageJSONFiles {
		if strings.Contains(file, "node_modules") {
			continue
		}
		if strings.HasPrefix(file, "gen/") || strings.Contains(file, "/gen/") {
			continue
		}
		packageJSONFiles = append(packageJSONFiles, file)
		log.Entry().Info("Discovered package.json file " + file)
	}
	return packageJSONFiles, nil
}
