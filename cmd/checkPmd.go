package cmd

import (
	"fmt"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func checkPmd(config checkPmdOptions, telemetryData *telemetry.CustomData) {
	c := command.Command{}
	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())

	err := runCheckPmd(&config, telemetryData, &c)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runCheckPmd(config *checkPmdOptions, telemetryData *telemetry.CustomData, command execRunner) error {
	var defines []string

	if config.Excludes != nil {
		defines = append(defines, "-Dpmd.excludes="+strings.Join(config.Excludes, ","))
	}
	if config.RuleSets != nil {
		defines = append(defines, "-Dpmd.rulesets="+strings.Join(config.RuleSets, ","))
	}
	if config.MavenModulesExcludes != nil {
		for t, module := range config.MavenModulesExcludes {
			fmt.Printf("my module: %v \n", t)
			defines = append(defines, strings.Join([]string{"-pl", module}, " "))
		}
	}

	// build in exclusion of integration-tests and unit-tests
	mavenOptions := maven.ExecuteOptions{
		Goals:   []string{"org.apache.maven.plugins:maven-pmd-plugin:3.13.0:pmd"},
		Defines: defines,
	}

	_, err := maven.Execute(&mavenOptions, command)
	return err
}
