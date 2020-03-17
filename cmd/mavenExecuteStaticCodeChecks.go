package cmd

import (
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func mavenExecuteStaticCodeChecks(config mavenExecuteStaticCodeChecksOptions, telemetryData *telemetry.CustomData) {
	c := command.Command{}
	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())
	err := runMavenStaticCodeChecks(&config, telemetryData, &c)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runMavenStaticCodeChecks(config *mavenExecuteStaticCodeChecksOptions, telemetryData *telemetry.CustomData, command execRunner) error {
	var defines []string
	var goals []string

	if !config.SpotBugs && !config.Pmd {
		log.Entry().Fatal("Neither SpotBugs nor Pmd are configured. At least one of those tools have to be enabled")
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
		Goals:   goals,
		Defines: defines,
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

	mavenOptions := maven.ExecuteOptions{
		Goals:   []string{"com.github.spotbugs:spotbugs-maven-plugin:3.1.12:spotbugs"},
		Defines: defines,
	}

	return &mavenOptions
}

func getPmdMavenParameters(config *mavenExecuteStaticCodeChecksOptions) *maven.ExecuteOptions {
	var defines []string
	if config.PmdExcludes != nil {
		defines = append(defines, "-Dpmd.excludes="+strings.Join(config.PmdExcludes, ","))
	}
	if config.PmdRuleSets != nil {
		defines = append(defines, "-Dpmd.rulesets="+strings.Join(config.PmdRuleSets, ","))
	}

	mavenOptions := maven.ExecuteOptions{
		Goals:   []string{"org.apache.maven.plugins:maven-pmd-plugin:3.13.0:pmd"},
		Defines: defines,
	}

	return &mavenOptions
}
