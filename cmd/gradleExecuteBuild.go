package cmd

import (
	"fmt"
	"net/http"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/gradle"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type gradleExecuteBuildUtils interface {
	command.ExecRunner
	piperutils.FileUtils
	DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error
}

type gradleExecuteBuildUtilsBundle struct {
	*command.Command
	*piperutils.Files
}

func (g *gradleExecuteBuildUtilsBundle) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	return fmt.Errorf("not implemented")
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
		log.Entry().WithError(err).Fatalf("step execution failed: %v", err)
	}
}

func runGradleExecuteBuild(config *gradleExecuteBuildOptions, telemetryData *telemetry.CustomData, utils gradleExecuteBuildUtils) error {
	opt := &gradle.ExecuteOptions{
		BuildGradlePath:    config.Path,
		Task:               config.Task,
		CreateBOM:          config.CreateBOM,
		Publish:            config.Publish,
		RepositoryURL:      config.RepositoryURL,
		RepositoryPassword: config.RepositoryPassword,
		RepositoryUsername: config.RepositoryUsername,
		ArtifactVersion:    config.ArtifactVersion,
		ArtifactGroupID:    config.ArtifactGroupID,
		ArtifactID:         config.ArtifactID,
	}

	if err := gradle.Execute(opt, utils); err != nil {
		log.Entry().WithError(err).Errorf("gradle build execution was failed: %v", err)
		return err
	}

	return nil
}
