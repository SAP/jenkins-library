package cmd

import (
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func mavenBuild(config mavenBuildOptions, telemetryData *telemetry.CustomData) {
	c := command.Command{}

	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())

	err := runMavenBuild(&config, telemetryData, &c)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runMavenBuild(config *mavenBuildOptions, telemetryData *telemetry.CustomData, command execRunner) error {
	var flags = []string{"-update-snapshots", "--batch-mode"}

	exists, _ := piperutils.FileExists("integration-tests/pom.xml")
	if exists {
		flags = append(flags, "-pl", "!integration-tests")
	}

	var defines []string
	var goals []string

	goals = append(goals, "org.jacoco:jacoco-maven-plugin:prepare-agent")

	if config.Flatten {
		goals = append(goals, "flatten:flatten")
		defines = append(defines, "-Dflatten.mode=resolveCiFriendliesOnly", "-DupdatePomFile=true")
	}

	if config.Verify {
		goals = append(goals, "verify")
	} else {
		goals = append(goals, "install")
	}

	mavenOptions := maven.ExecuteOptions{
		Flags:                       flags,
		Goals:                       goals,
		Defines:                     defines,
		PomPath:                     config.PomPath,
		ProjectSettingsFile:         config.ProjectSettingsFile,
		GlobalSettingsFile:          config.GlobalSettingsFile,
		M2Path:                      config.M2Path,
		LogSuccessfulMavenTransfers: config.LogSuccessfulMavenTransfers,
	}

	_, err := maven.Execute(&mavenOptions, command)
	return err
}
