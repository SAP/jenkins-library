package cmd

import (
	"os"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/gradle"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type gradleExecuteBuildUtils interface {
	command.ExecRunner
	FileExists(filename string) (bool, error)
	FileWrite(path string, content []byte, perm os.FileMode) error
	FileRemove(path string) error
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
		log.Entry().WithError(err).Fatal("step execution failed: %w", err)
	}
}

func runGradleExecuteBuild(config *gradleExecuteBuildOptions, telemetryData *telemetry.CustomData, utils gradleExecuteBuildUtils) error {
	opt := &gradle.ExecuteOptions{
		BuildGradlePath: config.Path,
		Task:            config.Task,
		CreateBOM:       config.CreateBOM,
	}

	if err := gradle.Execute(opt, utils); err != nil {
		log.Entry().WithError(err).Errorln("build.gradle execution was failed: %w", err)
		return err
	}

	return nil
}
