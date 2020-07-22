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
