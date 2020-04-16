package cmd

import (
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	FileUtils "github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func nodeJsBuild(config nodeJsBuildOptions, telemetryData *telemetry.CustomData) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())

	// for http calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runNodeJsBuild(&config, telemetryData, &c)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runNodeJsBuild(config *nodeJsBuildOptions, telemetryData *telemetry.CustomData, command execRunner) error {
	log.Entry().WithField("LogField", "Log field content").Info("This is just a demo for a simple step.")

	packageLockExists, err := FileUtils.FileExists("package-lock.json")
	if err != nil {
		return err
	}
	if config.Install {
		log.Entry().Info("run install")
		if packageLockExists {
			err = command.RunExecutable("npm", "ci")
			if err != nil {
				return err
			}
		} else {
			err = command.RunExecutable("npm", "install")
			if err != nil {
				return err
			}
		}
	}

	for _, v := range config.RunScripts {
		log.Entry().Info("run-script " + v)
		err = command.RunExecutable("npm", "run-script", v)
		if err != nil {
			return err
		}
	}
	return err
}
