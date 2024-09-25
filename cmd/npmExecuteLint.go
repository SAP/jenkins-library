package cmd

import (
	"fmt"
	"io"
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

	content, err := io.ReadAll(response.Body)
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
	if len(config.RunScript) == 0 {
		return fmt.Errorf("runScript is not allowed to be empty!")
	}

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

		err := runDefaultLint(npmExecutor, utils, config.FailOnError, config.OutputFormat, config.OutputFileName)

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
			return fmt.Errorf("%s script execution failed with error: %w. This might be the result of severe linting findings, or some other issue while executing the script. Please examine the linting results in the UI, the cilint.xml file, if available, or the log above. ", runScript, err)
		}
	}
	return nil
}

func runDefaultLint(npmExecutor npm.Executor, utils lintUtils, failOnError bool, outputFormat string, outputFileName string) error {
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
			lintPattern := "."
			dir := filepath.Dir(config)
			if dir != "." {
				lintPattern = dir + "/**/*.js"
			}

			args := prepareArgs([]string{
				"eslint",
				lintPattern,
				"-f", outputFormat,
				"--ignore-pattern", "node_modules/",
				"--ignore-pattern", ".eslintrc.js",
			}, fmt.Sprintf("./%s_%%s", strconv.Itoa(i)), outputFileName)

			err = execRunner.RunExecutable("npx", args...)
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
		generalPurposeLintConfigURI := "https://raw.githubusercontent.com/SAP/jenkins-library/master/resources/.eslintrc.json"
		utils.getGeneralPurposeConfig(generalPurposeLintConfigURI)

		err = execRunner.RunExecutable("npm", "install", "eslint@^7.0.0", "typescript@^3.7.4", "@typescript-eslint/parser@^3.0.0", "@typescript-eslint/eslint-plugin@^3.0.0")
		if err != nil {
			if failOnError {
				return fmt.Errorf("linter installation failed: %s", err)
			}
		}

		args := prepareArgs([]string{
			"--no-install",
			"eslint",
			".",
			"--ext", ".js,.jsx,.ts,.tsx",
			"-c", ".pipeline/.eslintrc.json",
			"-f", outputFormat,
			"--ignore-pattern", ".eslintrc.js",
		}, "./%s", outputFileName)

		err = execRunner.RunExecutable("npx", args...)
		if err != nil {
			if failOnError {
				return fmt.Errorf("lint execution failed. This might be the result of severe linting findings. The lint configuration used can be found here: %s", generalPurposeLintConfigURI)
			}
		}
	}
	return nil
}

func findEslintConfigs(utils lintUtils) []string {
	unfilteredListOfEslintConfigs, err := utils.Glob("**/.eslintrc*")
	if err != nil {
		log.Entry().Warnf("Error during resolving lint config files: %v", err)
	}
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

func prepareArgs(defaultArgs []string, outputFileNamePattern, outputFileName string) []string {
	if outputFileName != "" { // in this case we omit the -o flag and output will go to the log
		defaultArgs = append(defaultArgs, "-o", fmt.Sprintf(outputFileNamePattern, outputFileName))
	}
	return defaultArgs

}
