package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/npm"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type lintUtils interface {
	Glob(pattern string) (matches []string, err error)

	getExecRunner() command.ExecRunner
	getGeneralPurposeConfig(configURL string)
}

type lintUtilsBundle struct {
	*piperutils.Files
	execRunner *command.Command
	client     *piperhttp.Client
}

func newLintUtilsBundle() *lintUtilsBundle {
	return &lintUtilsBundle{
		Files:  &piperutils.Files{},
		client: &piperhttp.Client{},
	}
}

func (u *lintUtilsBundle) getExecRunner() command.ExecRunner {
	if u.execRunner == nil {
		u.execRunner = &command.Command{}
		u.execRunner.Stdout(log.Writer())
		u.execRunner.Stderr(log.Writer())
	}
	return u.execRunner
}

func (u *lintUtilsBundle) getGeneralPurposeConfig(configURL string) {
	response, err := u.client.SendRequest(http.MethodGet, configURL, nil, nil, nil)
	if err != nil {
		log.Entry().Warnf("failed to download general purpose configuration: %v", err)
		return
	}

	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Entry().Warnf("error while reading the general purpose configuration: %v", err)
		return
	}

	err = u.FileWrite(filepath.Join(".pipeline", ".eslintrc.json"), content, os.ModePerm)
	if err != nil {
		log.Entry().Warnf("failed to write .eslintrc.json file to .pipeline/: %v", err)
	}
}

func npmExecuteLint(config npmExecuteLintOptions, telemetryData *telemetry.CustomData) {
	utils := newLintUtilsBundle()
	npmExecutorOptions := npm.ExecutorOptions{DefaultNpmRegistry: config.DefaultNpmRegistry, ExecRunner: utils.getExecRunner()}
	npmExecutor := npm.NewExecutor(npmExecutorOptions)

	err := runNpmExecuteLint(npmExecutor, utils, &config)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runNpmExecuteLint(npmExecutor npm.Executor, utils lintUtils, config *npmExecuteLintOptions) error {
	packageJSONFiles := npmExecutor.FindPackageJSONFiles()
	packagesWithLintScript, _ := npmExecutor.FindPackageJSONFilesWithScript(packageJSONFiles, config.RunScript)

	if len(packagesWithLintScript) > 0 {
		if config.Install {
			err := npmExecutor.InstallAllDependencies(packagesWithLintScript)
			if err != nil {
				return err
			}
		}

		err := runLintScript(npmExecutor, config.RunScript, config.FailOnError)
		if err != nil {
			return err
		}
	} else {
		if config.Install {
			err := npmExecutor.InstallAllDependencies(packageJSONFiles)
			if err != nil {
				return err
			}
		}

		err := runDefaultLint(npmExecutor, utils, config.FailOnError)
		if err != nil {
			return err
		}
	}
	return nil
}

func runLintScript(npmExecutor npm.Executor, runScript string, failOnError bool) error {
	runScripts := []string{runScript}
	runOptions := []string{"--silent"}

	err := npmExecutor.RunScriptsInAllPackages(runScripts, runOptions, nil, false, nil, nil)
	if err != nil {
		if failOnError {
			return fmt.Errorf(runScript+" script execution failed with error: %w. This might be the result of severe linting findings, or some other issue while executing the script. Please examine the linting results in the UI, the cilint.xml file, if available, or the log above. ", err)
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
			dir := filepath.Dir(config)
			if dir == "." {
				err = execRunner.RunExecutable("npx", "eslint", ".", "-f", "checkstyle", "-o", "./"+strconv.Itoa(i)+"_defaultlint.xml", "--ignore-pattern", "node_modules/", "--ignore-pattern", ".eslintrc.js")
			} else {
				lintPattern := dir + "/**/*.js"
				err = execRunner.RunExecutable("npx", "eslint", lintPattern, "-f", "checkstyle", "-o", "./"+strconv.Itoa(i)+"_defaultlint.xml", "--ignore-pattern", "node_modules/", "--ignore-pattern", ".eslintrc.js")
			}
			if err != nil {
				if failOnError {
					return fmt.Errorf("Lint execution failed. This might be the result of severe linting findings, problems with the provided ESLint configuration (%s), or another issue. Please examine the linting results in the UI or in %s, if available, or the log above. ", config, strconv.Itoa(i)+"_defaultlint.xml")
				}
			}
		}
	} else {
		// install dependencies manually, since npx cannot resolve the dependencies required for general purpose
		// ESLint config, e.g., TypeScript ESLint plugin
		log.Entry().Info("Run ESLint with general purpose config")
		utils.getGeneralPurposeConfig("https://raw.githubusercontent.com/SAP/jenkins-library/master/resources/.eslintrc.json")

		// Ignore possible errors when invoking ESLint to not fail the pipeline based on linting results
		_ = execRunner.RunExecutable("npm", "install", "eslint@^7.0.0", "typescript@^3.7.4", "@typescript-eslint/parser@^3.0.0", "@typescript-eslint/eslint-plugin@^3.0.0")
		_ = execRunner.RunExecutable("npx", "--no-install", "eslint", ".", "--ext", ".js,.jsx,.ts,.tsx", "-c", ".pipeline/.eslintrc.json", "-f", "checkstyle", "-o", "./defaultlint.xml", "--ignore-pattern", ".eslintrc.js")
	}
	return nil
}

func findEslintConfigs(utils lintUtils) []string {
	unfilteredListOfEslintConfigs, _ := utils.Glob("**/.eslintrc.*")

	var eslintConfigs []string

	for _, config := range unfilteredListOfEslintConfigs {
		if strings.Contains(config, "node_modules") {
			continue
		}

		if strings.HasPrefix(config, ".pipeline"+string(os.PathSeparator)) {
			continue
		}

		eslintConfigs = append(eslintConfigs, config)
		log.Entry().Info("Discovered ESLint config " + config)
	}
	return eslintConfigs
}
