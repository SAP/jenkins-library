package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/syft"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func buildkitExecute(config buildkitExecuteOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *buildkitExecuteCommonPipelineEnvironment) {
	// for command execution use Command
	c := command.Command{
		ErrorCategoryMapping: map[string][]string{
			log.ErrorBuild.String(): {
				"failed to execute buildctl",
			},
		},
		StepName: "buildkitExecute",
	}

	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	client := &piperhttp.Client{}
	fileUtils := &piperutils.Files{}

	err := runBuildkitExecute(&config, telemetryData, commonPipelineEnvironment, &c, client, fileUtils)
	if err != nil {
		log.Entry().WithError(err).Fatal("Buildkit execution failed")
	}
}

func runBuildkitExecute(config *buildkitExecuteOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *buildkitExecuteCommonPipelineEnvironment, execRunner command.ExecRunner, httpClient piperhttp.Sender, fileUtils piperutils.FileUtils) error {
	log.Entry().Info("Starting buildkit execution in improved rootless mode...")
	log.Entry().Infof("Using Dockerfile at: %s", config.DockerfilePath)

	// Set environment variables for rootless BuildKit
	os.Setenv("BUILDKIT_HOST", "unix:///run/buildkit/buildkitd.sock")
	os.Setenv("XDG_RUNTIME_DIR", "/run/buildkit")
	os.Setenv("HOME", "/home/user")
	os.Setenv("USER", "user")

	// Wait for buildkit daemon to be available
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		err := execRunner.RunExecutable("buildctl", "debug", "workers")
		if err == nil {
			break
		}
		if i == maxRetries-1 {
			return errors.Wrap(err, "Buildkit daemon not available after max retries")
		}
		log.Entry().Infof("Waiting for rootless buildkit daemon to be available (attempt %d/%d)...", i+1, maxRetries)
		time.Sleep(2 * time.Second)
	}

	// Check buildctl version
	err := execRunner.RunExecutable("buildctl", "--version")
	if err != nil {
		return errors.Wrap(err, "Failed to execute buildctl command")
	}

	// Handle Docker authentication
	dockerConfigDir := "/home/user/.docker"
	if len(config.DockerConfigJSON) > 0 {
		dockerConfigJSON, err := fileUtils.FileRead(config.DockerConfigJSON)
		if err != nil {
			return fmt.Errorf("failed to read Docker config: %w", err)
		}
		if err := fileUtils.MkdirAll(dockerConfigDir, 0755); err != nil {
			return fmt.Errorf("failed to create .docker directory: %w", err)
		}
		if err := fileUtils.FileWrite(fmt.Sprintf("%s/config.json", dockerConfigDir), dockerConfigJSON, 0644); err != nil {
			return fmt.Errorf("failed to write Docker config: %w", err)
		}
	}

	// Build with buildkit using direct buildctl
	buildOpts := []string{
		"build",
		"--frontend=dockerfile.v0",
		"--local", "context=.",
		"--local", fmt.Sprintf("dockerfile=%s", config.DockerfilePath),
		"--progress=plain",
		"--max-parallel-jobs=4",
		"--export-cache", "type=inline",
		"--import-cache", "type=local,src=/home/user/.local/share/buildkit/cache",
	}

	// Add build options from config
	buildOpts = append(buildOpts, config.BuildOptions...)

	imageTag := "latest"
	if config.ContainerImageTag != "" {
		imageTag = config.ContainerImageTag
	}
	if config.ContainerImageName != "" && config.ContainerRegistryURL != "" {
		destination := fmt.Sprintf("%s/%s:%s", config.ContainerRegistryURL, config.ContainerImageName, imageTag)
		buildOpts = append(buildOpts, "--output", fmt.Sprintf("type=image,name=%s", destination))

		commonPipelineEnvironment.container.registryURL = config.ContainerRegistryURL
		commonPipelineEnvironment.container.imageNameTag = fmt.Sprintf("%s:%s", config.ContainerImageName, imageTag)
		commonPipelineEnvironment.container.imageNameTags = append(commonPipelineEnvironment.container.imageNameTags, fmt.Sprintf("%s:%s", config.ContainerImageName, imageTag))
		commonPipelineEnvironment.container.imageNames = append(commonPipelineEnvironment.container.imageNames, config.ContainerImageName)
	} else {
		// Build without pushing if no registry/name provided
		buildOpts = append(buildOpts, "--output", "type=docker")
	}

	log.Entry().Info("Executing buildkit build...")
	err = execRunner.RunExecutable("buildctl", buildOpts...)
	if err != nil {
		return fmt.Errorf("buildkit build failed: %w", err)
	}

	log.Entry().Info("Buildkit execution completed")

	if config.CreateBOM {
		log.Entry().Info("Generating bill of materials using syft...")
		if err := syft.GenerateSBOM(config.SyftDownloadURL, dockerConfigDir, execRunner, fileUtils, httpClient, commonPipelineEnvironment.container.registryURL, commonPipelineEnvironment.container.imageNameTags); err != nil {
			return fmt.Errorf("failed to generate BOM: %w", err)
		}
		log.Entry().Info("BOM generation completed")
	}

	return nil
}
