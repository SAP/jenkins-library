package cmd

import (
	"encoding/json"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/npm"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"path"
	"strconv"
	"strings"
)

func npmExecuteLint(config npmExecuteLintOptions, telemetryData *telemetry.CustomData) {
	utils := npm.NpmUtilsBundle{}

	err := runNpmExecuteLint(&utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runNpmExecuteLint(utils npm.NpmUtils) error {
	packageJSONFiles, err := npm.FindPackageJSONFiles(utils)
	if err != nil {
		return err
	}

	packagesWithCilint, err := findPackagesWithCilint(packageJSONFiles, utils)
	if err != nil {
		return err
	}

	if len(packagesWithCilint) > 0 {
		err = runCiLint(packagesWithCilint, utils)
		if err != nil {
			return err
		}
	} else {
		err = runDefaultLint(utils)
		if err != nil {
			return err
		}
	}

	return nil
}

func findPackagesWithCilint(packageJSONFiles []string, utils npm.NpmUtils) ([]string, error) {
	var packageJSON map[string]interface{}
	var packagesWithCilint []string

	for _, file := range packageJSONFiles {
		packageRaw, err := utils.FileRead(file)
		if err != nil {
			return nil, err
		}

		json.Unmarshal(packageRaw, &packageJSON)

		scripts, ok := packageJSON["scripts"].(map[string]interface{})
		if ok {
			_, ok := scripts["ci-lint"].(string)
			if ok {
				packagesWithCilint = append(packagesWithCilint, file)
				log.Entry().Info("Discovered ci-lint script in " + file)
			}
		}
		packageJSON = nil
	}
	return packagesWithCilint, nil
}

func runCiLint(packagesWithCilint []string, utils npm.NpmUtils) error {
	execRunner := utils.GetExecRunner()
	oldWorkingDirectory, err := utils.Getwd()
	if err != nil {
		return err
	}

	for _, packageJSON := range packagesWithCilint {
		dir := path.Dir(packageJSON)
		err := utils.Chdir(dir)
		if err != nil {
			return err
		}

		// Ignore possible errors when invoking ci-lint script to not fail the pipeline based on linting results
		err = execRunner.RunExecutable("npm", "run", "--silent", "ci-lint")

		err = utils.Chdir(oldWorkingDirectory)
		if err != nil {
			return err
		}
	}
	return nil
}

func runDefaultLint(utils npm.NpmUtils) error {
	execRunner := utils.GetExecRunner()
	eslintConfigs, err := findEslintConfigs(utils)
	if err != nil {
		return err
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
				err = execRunner.RunExecutable("npx", "eslint", lintPattern , "-f", "checkstyle", "-o", "./"+strconv.Itoa(i)+"_defaultlint.xml", "--ignore-pattern", "node_modules/", "--ignore-pattern", ".eslintrc.js")
			}
		}
	} else {
		// Install dependencies manually, since npx cannot resolve the dependencies required for general purpose
		// ESLint config, e.g., TypeScript ESLint plugin
		log.Entry().Info("Run ESLint with general purpose config")
		// Ignore possible errors when invoking ci-lint script to not fail the pipeline based on linting results
		err = execRunner.RunExecutable("npm", "install", "eslint@^7.0.0", "typescript@^3.7.4", "@typescript-eslint/parser@^3.0.0", "@typescript-eslint/eslint-plugin@^3.0.0" )
		err = execRunner.RunExecutable("npx", "--no-install", "eslint", ".", "--ext", ".js,.jsx,.ts,.tsx", "-c", ".pipeline/.eslintrc.json", "-f", "checkstyle", "-o", "./defaultlint.xml", "--ignore-pattern", ".eslintrc.js" )
	}
	return nil
}

func findEslintConfigs(utils npm.NpmUtils) ([]string, error) {
	unfilteredListOfEslintConfigs, err := utils.Glob("**/.eslintrc.*")
	if err != nil {
		return nil, err
	}

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
	return eslintConfigs, nil
}
