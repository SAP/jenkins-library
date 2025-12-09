package cmd

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func mavenExecuteIntegration(config mavenExecuteIntegrationOptions, _ *telemetry.CustomData) {
	err := runMavenExecuteIntegration(&config, maven.NewUtilsBundle())
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runMavenExecuteIntegration(config *mavenExecuteIntegrationOptions, utils maven.Utils) error {
	integrationTestsPomPath := filepath.Join("integration-tests", "pom.xml")
	hasIntegrationTestsModule, _ := utils.FileExists(integrationTestsPomPath)
	if !hasIntegrationTestsModule {
		return fmt.Errorf("maven module 'integration-tests' does not exist in project structure")
	}

	if err := validateForkCount(config.ForkCount); err != nil {
		return err
	}

	retryDefine := fmt.Sprintf("-Dsurefire.rerunFailingTestsCount=%v", config.Retry)
	forkCountDefine := fmt.Sprintf("-Dsurefire.forkCount=%v", config.ForkCount)

	if config.InstallArtifacts {
		if config.UseReactorForMultiModuleBuild {
			err := maven.InstallModuleWithReactor("integration-tests", &maven.EvaluateOptions{
				M2Path:              config.M2Path,
				ProjectSettingsFile: config.ProjectSettingsFile,
				GlobalSettingsFile:  config.GlobalSettingsFile,
			}, utils)
			if err != nil {
				return err
			}
		} else {
			err := maven.InstallMavenArtifacts(&maven.EvaluateOptions{
				M2Path:              config.M2Path,
				ProjectSettingsFile: config.ProjectSettingsFile,
				GlobalSettingsFile:  config.GlobalSettingsFile,
			}, utils)
			if err != nil {
				return err
			}
		}
	}

	var targetPomPath string
	var flags []string

	if config.UseReactorForMultiModuleBuild {
		targetPomPath = "pom.xml"
		flags = []string{"-pl", "integration-tests"}
	} else {
		targetPomPath = integrationTestsPomPath
		flags = []string{}
	}

	mavenOptions := maven.ExecuteOptions{
		PomPath:             targetPomPath,
		M2Path:              config.M2Path,
		ProjectSettingsFile: config.ProjectSettingsFile,
		GlobalSettingsFile:  config.GlobalSettingsFile,
		Goals:               []string{"org.jacoco:jacoco-maven-plugin:prepare-agent", config.Goal},
		Defines:             []string{retryDefine, forkCountDefine},
		Flags:               flags,
	}

	_, err := maven.Execute(&mavenOptions, utils)
	return err
}

func validateForkCount(value string) error {
	var err error

	if strings.HasSuffix(value, "C") {
		value := strings.TrimSuffix(value, "C")
		for _, c := range value {
			if !unicode.IsDigit(c) && c != '.' {
				err = fmt.Errorf("only integers or floats allowed with 'C' suffix")
				break
			}
		}
		if err == nil {
			_, err = strconv.ParseFloat(value, 64)
		}
	} else {
		for _, c := range value {
			if !unicode.IsDigit(c) {
				err = fmt.Errorf("only integers allowed without 'C' suffix")
				break
			}
		}
		if err == nil {
			_, err = strconv.ParseInt(value, 10, 64)
		}
	}

	if err != nil {
		return fmt.Errorf("invalid forkCount parameter '%v': %w, please see https://maven.apache.org/surefire/maven-surefire-plugin/test-mojo.html#forkCount for details", value, err)
	}
	return nil
}
