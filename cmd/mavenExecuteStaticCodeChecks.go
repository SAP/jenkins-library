package cmd

import (
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"strconv"
)

func mavenExecuteStaticCodeChecks(config mavenExecuteStaticCodeChecksOptions, telemetryData *telemetry.CustomData) {
	c := command.Command{}
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())
	err := runMavenStaticCodeChecks(&config, telemetryData, &c)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runMavenStaticCodeChecks(config *mavenExecuteStaticCodeChecksOptions, telemetryData *telemetry.CustomData, command command.ExecRunner) error {
	var defines []string
	var goals []string

	if !config.SpotBugs && !config.Pmd {
		log.Entry().Warnf("Neither SpotBugs nor Pmd are configured. Skipping step execution")
		return nil
	}

	if testModulesExcludes := maven.GetTestModulesExcludes(); testModulesExcludes != nil {
		defines = append(defines, testModulesExcludes...)
	}
	if config.MavenModulesExcludes != nil {
		for _, module := range config.MavenModulesExcludes {
			defines = append(defines, "-pl")
			defines = append(defines, "!"+module)
		}
	}

	if config.SpotBugs {
		spotBugsMavenParameters := getSpotBugsMavenParameters(config)
		defines = append(defines, spotBugsMavenParameters.Defines...)
		goals = append(goals, spotBugsMavenParameters.Goals...)
	}
	if config.Pmd {
		pmdMavenParameters := getPmdMavenParameters(config)
		defines = append(defines, pmdMavenParameters.Defines...)
		goals = append(goals, pmdMavenParameters.Goals...)
	}
	finalMavenOptions := maven.ExecuteOptions{
		Goals:                       goals,
		Defines:                     defines,
		ProjectSettingsFile:         config.ProjectSettingsFile,
		GlobalSettingsFile:          config.GlobalSettingsFile,
		M2Path:                      config.M2Path,
		LogSuccessfulMavenTransfers: config.LogSuccessfulMavenTransfers,
	}
	_, err := maven.Execute(&finalMavenOptions, command)
	return err
}

func getSpotBugsMavenParameters(config *mavenExecuteStaticCodeChecksOptions) *maven.ExecuteOptions {
	var defines []string
	if config.SpotBugsIncludeFilterFile != "" {
		defines = append(defines, "-Dspotbugs.includeFilterFile="+config.SpotBugsIncludeFilterFile)
	}
	if config.SpotBugsExcludeFilterFile != "" {
		defines = append(defines, "-Dspotbugs.excludeFilterFile="+config.SpotBugsExcludeFilterFile)
	}
	if config.SpotBugsMaxAllowedViolations != 0 {
		defines = append(defines, "-Dspotbugs.maxAllowedViolations="+strconv.Itoa(config.SpotBugsMaxAllowedViolations))
	}

	mavenOptions := maven.ExecuteOptions{
		// check goal executes spotbugs goal first and fails the build if any bugs were found
		Goals:   []string{"com.github.spotbugs:spotbugs-maven-plugin:4.1.4:check"},
		Defines: defines,
	}

	return &mavenOptions
}

func getPmdMavenParameters(config *mavenExecuteStaticCodeChecksOptions) *maven.ExecuteOptions {
	var defines []string
	if config.PmdMaxAllowedViolations != 0 {
		defines = append(defines, "-Dpmd.maxAllowedViolations="+strconv.Itoa(config.PmdMaxAllowedViolations))
	}
	if config.PmdFailurePriority >= 1 && config.PmdFailurePriority <= 5 {
		defines = append(defines, "-Dpmd.failurePriority="+strconv.Itoa(config.PmdFailurePriority))
	} else if config.PmdFailurePriority != 0 {
		log.Entry().Warningf("Pmd failure priority must be a value between 1 and 5. %v was configured. Defaulting to 5.", config.PmdFailurePriority)
	}

	mavenOptions := maven.ExecuteOptions{
		// check goal executes pmd goal first and fails the build if any violations were found
		Goals:   []string{"org.apache.maven.plugins:maven-pmd-plugin:3.13.0:check"},
		Defines: defines,
	}

	return &mavenOptions
}
