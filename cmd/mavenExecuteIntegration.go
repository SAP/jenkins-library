package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

type mavenExecuteIntegrationUtils interface {
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	RunExecutable(e string, p ...string) error

	FileExists(filename string) (bool, error)
}

type mavenExecuteIntegrationUtilsBundle struct {
	*command.Command
	*piperutils.Files
}

func newMavenExecuteIntegrationUtils() mavenExecuteIntegrationUtils {
	utils := mavenExecuteIntegrationUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func mavenExecuteIntegration(config mavenExecuteIntegrationOptions, _ *telemetry.CustomData) {
	utils := newMavenExecuteIntegrationUtils()
	err := runMavenExecuteIntegration(&config, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runMavenExecuteIntegration(config *mavenExecuteIntegrationOptions, utils mavenExecuteIntegrationUtils) error {
	pomPath := filepath.Join("integration-tests", "pom.xml")
	hasIntegrationTestsModule, _ := utils.FileExists(pomPath)
	if !hasIntegrationTestsModule {
		return fmt.Errorf("maven module 'integration-tests' does not exist in project structure")
	}

	if config.InstallArtifacts {
		err := maven.InstallMavenArtifacts(utils, maven.EvaluateOptions{
			M2Path:              config.M2Path,
			ProjectSettingsFile: config.ProjectSettingsFile,
			GlobalSettingsFile:  config.GlobalSettingsFile,
		})
		if err != nil {
			return err
		}
	}

	if err := validateForkCount(config.ForkCount); err != nil {
		return err
	}

	retryDefine := fmt.Sprintf("-Dsurefire.rerunFailingTestsCount=%v", config.Retry)
	forkCountDefine := fmt.Sprintf("-Dsurefire.forkCount=%v", config.ForkCount)

	mavenOptions := maven.ExecuteOptions{
		PomPath:             pomPath,
		M2Path:              config.M2Path,
		ProjectSettingsFile: config.ProjectSettingsFile,
		GlobalSettingsFile:  config.GlobalSettingsFile,
		Goals:               []string{"org.jacoco:jacoco-maven-plugin:prepare-agent", "test"},
		Defines:             []string{retryDefine, forkCountDefine},
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
