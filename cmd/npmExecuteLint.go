package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/npm"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"path"
	"strconv"
	"strings"
)

func npmExecuteLint(config npmExecuteLintOptions, telemetryData *telemetry.CustomData) {
	utils := npm.UtilsBundle{}

	err := runNpmExecuteLint(&utils, &config)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runNpmExecuteLint(utils npm.Utils, config *npmExecuteLintOptions) error {
	options := npm.ExecuteOptions{
		RunScripts:         []string{"ci-lint"},
		RunOptions:         []string{"--silent"},
		DefaultNpmRegistry: config.DefaultNpmRegistry,
		SapNpmRegistry:     config.SapNpmRegistry,
	}

	packageJSONFiles := npm.FindPackageJSONFiles(utils)

	packagesWithCiLint, _ := npm.FindPackageJSONFilesWithScript(utils, packageJSONFiles, options.RunScripts[0])

	if len(packagesWithCiLint) > 0 {
		err := runCiLint(utils, options, config.FailOnError)
		if err != nil {
			return err
		}
	} else {
		err := runDefaultLint(utils, options, config.FailOnError)
		if err != nil {
			return err
		}
	}
	return nil
}

func runCiLint(utils npm.Utils, options npm.ExecuteOptions, failOnError bool) error {
	err := npm.ExecuteAllScripts(utils, options)
	if err != nil {
		if failOnError {
			return err
		}
	}
	return nil
}

func runDefaultLint(utils npm.Utils, options npm.ExecuteOptions, failOnError bool) error {
	execRunner := utils.GetExecRunner()
	eslintConfigs := findEslintConfigs(utils)

	err := npm.SetNpmRegistries(execRunner, &options)
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
		// Install dependencies manually, since npx cannot resolve the dependencies required for general purpose
		// ESLint config, e.g., TypeScript ESLint plugin
		log.Entry().Info("Run ESLint with general purpose config")
		// Ignore possible errors when invoking ci-lint script to not fail the pipeline based on linting results
		_ = execRunner.RunExecutable("npm", "install", "eslint@^7.0.0", "typescript@^3.7.4", "@typescript-eslint/parser@^3.0.0", "@typescript-eslint/eslint-plugin@^3.0.0")
		_ = execRunner.RunExecutable("npx", "--no-install", "eslint", ".", "--ext", ".js,.jsx,.ts,.tsx", "-c", ".pipeline/.eslintrc.json", "-f", "checkstyle", "-o", "./defaultlint.xml", "--ignore-pattern", ".eslintrc.js")
	}
	return nil
}

func findEslintConfigs(utils npm.Utils) []string {
	unfilteredListOfEslintConfigs, _ := utils.Glob("**/.eslintrc.*")

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
