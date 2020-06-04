package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/npm"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"path"
)

func npmExecuteLint(config npmExecuteLintOptions, telemetryData *telemetry.CustomData) {
	c := command.Command{}
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())
	utils := npm.NpmUtilsBundle{}

	err := runNpmExecuteLint(&config, telemetryData, &c, &utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runNpmExecuteLint(config *npmExecuteLintOptions, telemetryData *telemetry.CustomData, command execRunner, utils npm.NpmUtils) error {
	log.Entry().WithField("LogField", "Log field content").Info("This is just a demo for a simple step.")
	// if ci-lint in any package.json; runCiLint(), move findPackageJsonFiles to pkg? what about execute scripts? could also be useful for other steps?
	// move npmUtilsBundle to pkg
	// installDependencies if needed
	// else runDefaultLint(), retrieve ESLint config from resources? how to run eslint (npm/npx) from go?
	//packageJSONFiles, err := npm.FindPackageJSONFiles(&npm.NpmUtilsBundle{})
	packageJSONFiles, err := npm.FindPackageJSONFiles(utils)
	if err != nil {
		return err
	}
	packagesWithCilint, err := findPackagesWithCilint(packageJSONFiles, utils)
	if err != nil {
		return err
	}
	fmt.Println(packageJSONFiles)
	fmt.Println("these files have a ci-lint script")
	fmt.Println(packagesWithCilint)

	if len(packagesWithCilint) > 0 {
		err = runCiLint(packagesWithCilint, utils)
	} else {
		err = runDefaultLint(utils)
	}

	return nil
}

func findPackagesWithCilint(packageJSONFiles []string, utils npm.NpmUtils) ([]string, error) {
	var packageJSON map[string]interface{}
	var packagesWithCilint []string

	for _, file := range packageJSONFiles {
		fmt.Println(file)
		packageRaw, err := utils.FileRead(file)
		if err != nil {
			return nil, err
		}

		json.Unmarshal(packageRaw, &packageJSON)

		scripts, ok := packageJSON["scripts"].(map[string]interface{})
		if ok {
			ciLint, ok := scripts["ci-lint"].(string)
			if ok {
				fmt.Println(ciLint)
				packagesWithCilint = append(packagesWithCilint, file)

			} else {
				fmt.Println("No ci-lint script defined")
			}
		} else {
			fmt.Println("No ci-lint script defined")
		}
		//fmt.Println("Thats the package.json: ")
		//fmt.Println(packageJSON)
		// if len(packagesWithCilint) > 0 {runCiLint() } else{ runDefaultLint}
		packageJSON = nil
	}
	return packagesWithCilint, nil
}

func runCiLint(packagesWithCilint []string, utils npm.NpmUtils) error {
	execRunner := utils.GetExecRunner()
	fmt.Println("Here the packageJSONs that have cilint again: ")
	fmt.Println(packagesWithCilint)
	fmt.Println("thats the length of packageswithcilint: ----------------")
	fmt.Println(len(packagesWithCilint))
	oldWorkingDirectory, err := utils.Getwd()
	if err != nil {
		return err
	}

	for i, packageJSON := range packagesWithCilint {
		fmt.Println(i)
		fmt.Println("thats the package json for which we need to run ci-lint:")
		fmt.Println(packageJSON)
		dir := path.Dir(packageJSON)
		err := utils.Chdir(dir)
		fmt.Println("thats the directory of the package json:")
		fmt.Println(dir)
		if err != nil {
			return err
		}
		fmt.Println("Now run ci-lint: ")
		err = execRunner.RunExecutable("npm", "run", "--silent", "ci-lint")
		if err != nil {
			fmt.Println("err was not nil -------------------------")
			fmt.Println(err)
			//return err
		}
		fmt.Println("reached the end of the loop")
		err = utils.Chdir(oldWorkingDirectory)
		if err != nil {
			return err
		}
	}
	return nil
}

func runDefaultLint(utils npm.NpmUtils) error {
	// if ESLint config in project: run eslint with user config } else { use config from resources }
	return nil
}

func findEslintConfigs(utils npm.NpmUtils) error {

	return nil
}
