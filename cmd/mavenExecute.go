package cmd

import (
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"os"

	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type mavenExecuteUtils interface {
	maven.Utils
	FileWrite(path string, content []byte, perm os.FileMode) error

}

type mavenExecuteUtilsBundle struct {
	*command.Command
	*piperutils.Files
	*piperhttp.Client
}

func newMavenExecuteUtilsBundle() mavenExecuteUtils {
	utils := mavenExecuteUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
		Client: &piperhttp.Client{},
	}
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func mavenExecute(config mavenExecuteOptions, _ *telemetry.CustomData) {
	err := runMavenExecute(config, newMavenExecuteUtilsBundle())
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runMavenExecute(config mavenExecuteOptions, utils mavenExecuteUtils) error {
	options := maven.ExecuteOptions{
		PomPath:                     config.PomPath,
		ProjectSettingsFile:         config.ProjectSettingsFile,
		GlobalSettingsFile:          config.GlobalSettingsFile,
		M2Path:                      config.M2Path,
		Goals:                       config.Goals,
		Defines:                     config.Defines,
		Flags:                       config.Flags,
		LogSuccessfulMavenTransfers: config.LogSuccessfulMavenTransfers,
		ReturnStdout:                config.ReturnStdout,
	}

	output, err := maven.Execute(&options, utils)
	if err == nil && config.ReturnStdout {
		err = utils.FileWrite(".pipeline/maven_output.txt", []byte(output), 0644)
	}
	return err
}
