package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/npm"
	FileUtils "github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/bmatcuk/doublestar"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

type lintUtils interface {
	fileWrite(path string, content []byte, perm os.FileMode) error
	getExecRunner() execRunner
	getGeneralPurposeConfig(configURL string) error
	glob(pattern string) (matches []string, err error)
}

type lintUtilsBundle struct {
	fileUtils  FileUtils.Files
	execRunner *command.Command
	client     http.Client
}

func newLintUtilsBundle() *lintUtilsBundle {
	return &lintUtilsBundle{
		fileUtils: FileUtils.Files{},
		client:    http.Client{},
	}
}

func (u *lintUtilsBundle) fileWrite(path string, content []byte, perm os.FileMode) error {
	parent := filepath.Dir(path)
	if parent != "" {
		err := u.fileUtils.MkdirAll(parent, 0775)
		if err != nil {
			return err
		}
	}
	return u.fileUtils.FileWrite(path, content, perm)
}

func (u *lintUtilsBundle) getExecRunner() execRunner {
	if u.execRunner == nil {
		u.execRunner = &command.Command{}
		u.execRunner.Stdout(log.Writer())
		u.execRunner.Stderr(log.Writer())
	}
	return u.execRunner
}

func (u *lintUtilsBundle) getGeneralPurposeConfig(configURL string) error {
	response, err := u.client.SendRequest("GET", configURL, nil, nil, nil)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("error reading %v: %w", response.Body, err)
	}

	err = u.fileWrite(".pipeline/.eslintrc.json", content, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write .eslintrc.json file to .pipeline/: %w", err)
	}

	return nil
}

func (u *lintUtilsBundle) glob(pattern string) (matches []string, err error) {
	return doublestar.Glob(pattern)
}

func npmExecuteLint(config npmExecuteLintOptions, telemetryData *telemetry.CustomData) {
	utils := newLintUtilsBundle()
	npmExecutor, err := npm.NewExecutor(false, []string{"ci-lint"}, []string{"--silent"}, config.DefaultNpmRegistry, config.SapNpmRegistry)

	err = runNpmExecuteLint(npmExecutor, utils, &config)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runNpmExecuteLint(npmExecutor npm.Executor, utils lintUtils, config *npmExecuteLintOptions) error {
	packageJSONFiles := npmExecutor.FindPackageJSONFiles()
	packagesWithCiLint, _ := npmExecutor.FindPackageJSONFilesWithScript(packageJSONFiles, "ci-lint")

	if len(packagesWithCiLint) > 0 {
		err := runCiLint(npmExecutor, config.FailOnError)
		if err != nil {
			return err
		}
	} else {
		err := runDefaultLint(npmExecutor, utils, config.FailOnError)
		if err != nil {
			return err
		}
	}
	return nil
}

func runCiLint(npmExecutor npm.Executor, failOnError bool) error {
	err := npmExecutor.ExecuteAllScripts()
	if err != nil {
		if failOnError {
			return err
		}
	}
	return nil
}

func runDefaultLint(npmExecutor npm.Executor, utils lintUtils, failOnError bool) error {
	execRunner := utils.getExecRunner()
	eslintConfigs := findEslintConfigs(utils)

	err := npmExecutor.SetNpmRegistries()
	if err != nil {
		log.Entry().Warnf("failed to set npm registries before running default lint: %v", err)
	}

	// If the user has ESLint configs in the project we use them to lint existing JS files. In this case we do not lint other types of files,
	// i.e., .jsx, .ts, .tsx, since we can not be sure that the provided config enables parsing of these file types.
	if len(eslintConfigs) > 0 {
		for i, config := range eslintConfigs {
			dir := path.Dir(config)
			if dir == "." {
				// Ignore possible errors when invoking ci-lint script to not fail the pipeline based on linting results
				err = execRunner.RunExecutable("npx", "eslint", ".", "-f", "checkstyle", "-o", "./"+strconv.Itoa(i)+"_defaultlint.xml", "--ignore-pattern", "node_modules/", "--ignore-pattern", ".eslintrc.js")
			} else {
				lintPattern := dir + "/**/*.js"
				// Ignore possible errors when invoking ci-lint script to not fail the pipeline based on linting results
				err = execRunner.RunExecutable("npx", "eslint", lintPattern, "-f", "checkstyle", "-o", "./"+strconv.Itoa(i)+"_defaultlint.xml", "--ignore-pattern", "node_modules/", "--ignore-pattern", ".eslintrc.js")
			}
			if err != nil {
				if failOnError {
					return fmt.Errorf("failed to run ESLint with config %s: %w", config, err)
				}
			}
		}
	} else {
		// install dependencies manually, since npx cannot resolve the dependencies required for general purpose
		// ESLint config, e.g., TypeScript ESLint plugin
		log.Entry().Info("Run ESLint with general purpose config")
		err = utils.getGeneralPurposeConfig("https://raw.githubusercontent.com/SAP/jenkins-library/stepNpmLint/resources/.eslintrc.json")
		if err != nil {
			return err
		}
		// Ignore possible errors when invoking ci-lint script to not fail the pipeline based on linting results
		_ = execRunner.RunExecutable("npm", "install", "eslint@^7.0.0", "typescript@^3.7.4", "@typescript-eslint/parser@^3.0.0", "@typescript-eslint/eslint-plugin@^3.0.0")
		_ = execRunner.RunExecutable("npx", "--no-install", "eslint", ".", "--ext", ".js,.jsx,.ts,.tsx", "-c", ".pipeline/.eslintrc.json", "-f", "checkstyle", "-o", "./defaultlint.xml", "--ignore-pattern", ".eslintrc.js")
	}
	return nil
}

func findEslintConfigs(utils lintUtils) []string {
	unfilteredListOfEslintConfigs, _ := utils.glob("**/.eslintrc.*")

	var eslintConfigs []string

	for _, config := range unfilteredListOfEslintConfigs {
		if strings.Contains(config, "node_modules") {
			continue
		}

		if strings.HasPrefix(config, ".pipeline/") {
			continue
		}

		eslintConfigs = append(eslintConfigs, config)
		log.Entry().Info("Discovered ESLint config " + config)
	}
	return eslintConfigs
}
