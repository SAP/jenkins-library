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
	err := runGradleExecuteBuild(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runGradleExecuteBuild(config *gradleExecuteBuildOptions, telemetryData *telemetry.CustomData, utils gradleExecuteBuildUtils) error {
	opt := &gradle.ExecuteOptions{BuildGradlePath: config.Path, Task: config.Task}

	_, err := gradle.Execute(opt, utils)
	if err != nil {
		log.Entry().WithError(err).Errorln("build.gradle execution was failed")
	}

	return nil
}
