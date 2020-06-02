package cmd

import (
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func npmExecuteLint(config npmExecuteLintOptions, telemetryData *telemetry.CustomData) {
	c := command.Command{}
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())
	err := runNpmExecuteLint(&config, telemetryData, &c)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runNpmExecuteLint(config *npmExecuteLintOptions, telemetryData *telemetry.CustomData, command execRunner) error {
	log.Entry().WithField("LogField", "Log field content").Info("This is just a demo for a simple step.")
	// if ci-lint in any package.json; runCiLint(), move findPackageJsonFiles to pkg? what about execute scripts? could also be useful for other steps?
		// move npmExecuteScriptsUtilsBundle to pkg
		// installDependencies if needed
	// else runDefaultLint(), retrieve ESLint config from resources? how to run eslint (npm/npx) from go?

	return nil
}

func runCiLint() error {

	return nil
}

func runDefaultLint() error {
	return nil
}
