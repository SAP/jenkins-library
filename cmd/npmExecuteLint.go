package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/npm"
	"github.com/SAP/jenkins-library/pkg/telemetry"
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
	var packageJSON map[string]interface{}
	var packageWithCilint []string
	for _, file := range packageJSONFiles {
		fmt.Println(file)
		packageRaw, err := utils.FileRead(file)
		if err != nil {
			return err
		}

		json.Unmarshal(packageRaw, &packageJSON)

		scripts, ok := packageJSON["scripts"].(map[string]interface{})
		if ok {
			ciLint, ok := scripts["ci-lint"].(string)
			if ok {
				fmt.Println(ciLint)
				packageWithCilint = append(packageWithCilint, file)

			} else {
				fmt.Println("No ci-lint script defined")
			}
		} else {
			fmt.Println("No ci-lint script defined")
		}
		fmt.Println("Thats the package.json: ")
		fmt.Println(packageJSON)
		// if len(packageWithCilint) > 0 {runCiLint() } else{ runDefaultLint}
		packageJSON = nil
	}
	fmt.Println(packageJSONFiles)
	fmt.Println("these files have a ci-lint script")
	fmt.Println(packageWithCilint)
	return nil
}

func runCiLint() error {

	return nil
}

func runDefaultLint() error {
	// if ESLint config in project: run eslint with user config } else { use config from resources }
	return nil
}
