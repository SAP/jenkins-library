package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type packBuildUtils interface {
	command.ExecRunner
	piperutils.FileUtils
}

type packBuildUtilsBundle struct {
	*command.Command
	*piperutils.Files
}

func newPackBuildUtils() packBuildUtils {
	utils := packBuildUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func packBuild(config packBuildOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *packBuildCommonPipelineEnvironment) {
	err := runPackBuild(&config, telemetryData, newPackBuildUtils(), commonPipelineEnvironment)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runPackBuild(config *packBuildOptions, _ *telemetry.CustomData, utils packBuildUtils, _ *packBuildCommonPipelineEnvironment) error {
	args := []string{
		"build",
		"--no-color",
		"--builder", config.Builder,
	}

	if err := downloadBinary(config.PackDownloadURL); err != nil {
		log.Entry().Error("failed to download pack cli")
		return err
	}

	if config.DockerConfigJSON != "" {
		os.Setenv("DOCKER_CONFIG", filepath.Dir(config.DockerConfigJSON))
		utils.AppendEnv([]string{fmt.Sprintf("DOCKER_CONFIG=%s", filepath.Dir(config.DockerConfigJSON))})
		utils.FileRename(config.DockerConfigJSON, filepath.Join(filepath.Dir(config.DockerConfigJSON), "config.json"))
	}

	for _, buildpack := range config.PreBuildpacks {
		args = append(args, "--pre-buildpack", buildpack)
	}

	for _, buildpack := range config.Buildpacks {
		args = append(args, "--buildpack", buildpack)
	}

	for _, buildpack := range config.PostBuildpacks {
		args = append(args, "--post-buildpack", buildpack)
	}

	if config.RunImage != "" {
		args = append(args, "--run-image", config.RunImage)
	}

	if config.Path != "" {
		args = append(args, "--path", config.Path)
	}

	if config.DefaultProcess != "" {
		args = append(args, "--default-process", config.DefaultProcess)
	}

	for k, v := range config.BuildEnvVars {
		utils.AppendEnv([]string{fmt.Sprintf("%s=%v", k, v)})
		args = append(args, "--env", k)
	}

	args = append(args, fmt.Sprintf("%s:%s", config.ContainerImageName, config.ContainerImageTag))

	if err := utils.RunExecutable("./pack", args...); err != nil {
		log.Entry().Error("failed to run pack cli")
		return err
	}

	return nil
}

func downloadBinary(downloadURL string) error {
	client := &piperhttp.Client{}

	if err := client.DownloadFile(downloadURL, "pack.tgz", nil, nil); err != nil {
		return err
	}

	return piperutils.Untar("pack.tgz", "", 0)
}
