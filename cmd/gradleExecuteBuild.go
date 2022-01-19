package cmd

import (
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/gradle"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type gradleExecuteBuildUtils interface {
	command.ExecRunner
	FileExists(filename string) (bool, error)
}

type gradleExecuteBuildUtilsBundle struct {
	*command.Command
	*piperutils.Files
}

func newGradleExecuteBuildUtils() gradleExecuteBuildUtils {
	utils := gradleExecuteBuildUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func gradleExecuteBuild(config gradleExecuteBuildOptions, telemetryData *telemetry.CustomData) {
	utils := newGradleExecuteBuildUtils()
	fileUtils := &piperutils.Files{}
	err := runGradleExecuteBuild(&config, telemetryData, utils, fileUtils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed: %w", err)
	}
}

func runGradleExecuteBuild(config *gradleExecuteBuildOptions, telemetryData *telemetry.CustomData, utils gradleExecuteBuildUtils, fileUtils piperutils.FileUtils) error {
	opt := &gradle.ExecuteOptions{BuildGradlePath: config.Path, Task: config.Task}

	_, err := gradle.Execute(opt, utils, fileUtils)
	if err != nil {
		log.Entry().WithError(err).Errorln("build.gradle execution was failed: %w", err)
		return err
	}

	return nil
}
