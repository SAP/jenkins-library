package cmd

import (
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func checkSpotBugs(config checkSpotBugsOptions, telemetryData *telemetry.CustomData) {
	c := command.Command{}
	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())

	err := runCheckSpotBugs(&config, telemetryData, &c)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runCheckSpotBugs(config *checkSpotBugsOptions, telemetryData *telemetry.CustomData, command execRunner) error {
	var defines []string
	if config.IncludeFilterFile != "" {
		defines = append(defines, "-Dspotbugs.includeFilterFile="+config.IncludeFilterFile)
	}
	if config.ExcludeFilterFile != "" {
		defines = append(defines, "-Dspotbugs.excludeFilterFile="+config.ExcludeFilterFile)
	}
	mavenOptions := maven.ExecuteOptions{
		Goals:   []string{"com.github.spotbugs:spotbugs-maven-plugin:3.1.12:spotbugs"},
		Defines: defines,
	}

	_, err := maven.Execute(&mavenOptions, command)
	return err
}
