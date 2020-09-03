package cmd

import (
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/bmatcuk/doublestar"
)

func mavenBuild(config mavenBuildOptions, telemetryData *telemetry.CustomData) {
	c := command.Command{}

	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	utils := piperutils.Files{}

	err := runMavenBuild(&config, telemetryData, &c, &utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runMavenBuild(config *mavenBuildOptions, telemetryData *telemetry.CustomData, command command.ExecRunner, utils piperutils.FileUtils) error {
	var flags = []string{"-update-snapshots", "--batch-mode"}

	exists, _ := utils.FileExists("integration-tests/pom.xml")
	if exists {
		flags = append(flags, "-pl", "!integration-tests")
	}

	var defines []string
	var goals []string

	// Setup Jacoco coverage recording
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
	if err != nil {
		return err
	}

	execFiles, _ := doublestar.Glob("**/*.exec")
	log.Entry().Infof("found .exec files: %v", execFiles)

	// Generate a Jacoco coverage report in XML format, needed by SonarQube scan
	mavenOptions.Goals = []string{"org.jacoco:jacoco-maven-plugin:report"}
	mavenOptions.Defines = []string{}
	_, err = maven.Execute(&mavenOptions, command)
	if err != nil {
		log.Entry().Warnf("failed to generate Jacoco coverage report: %v", err)
	}

	return nil
}
