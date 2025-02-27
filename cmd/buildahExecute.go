package cmd

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/syft"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func buildahExecute(config buildahExecuteOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *buildahExecuteCommonPipelineEnvironment) {
	c := command.Command{
		ErrorCategoryMapping: map[string][]string{
			log.ErrorBuild.String(): {
				"failed to execute buildah",
			},
		},
		StepName: "buildahExecute",
	}

	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	client := &piperhttp.Client{}
	fileUtils := &piperutils.Files{}

	err := runBuildahExecute(&config, telemetryData, commonPipelineEnvironment, &c, client, fileUtils)
	if err != nil {
		log.Entry().WithError(err).Fatal("Buildah execution failed")
	}
}

func runBuildahExecute(config *buildahExecuteOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *buildahExecuteCommonPipelineEnvironment, execRunner command.ExecRunner, httpClient piperhttp.Sender, fileUtils piperutils.FileUtils) error {
	log.Entry().Info("Starting buildah execution...")
	log.Entry().Infof("Using Dockerfile at: %s", config.DockerfilePath)

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

	// Check buildah version
	err := execRunner.RunExecutable("buildah", "--version")
	if err != nil {
		return errors.Wrap(err, "Failed to execute buildah command")
	}

	// Build options for buildah
	buildOpts := []string{
		"build-using-dockerfile",
		"--format=docker", // Use Docker format for compatibility
		"--file", config.DockerfilePath,
	}

	// Add build options from config
	buildOpts = append(buildOpts, config.BuildOptions...)

	imageTag := "latest"
	if config.ContainerImageTag != "" {
		imageTag = config.ContainerImageTag
	}

	if config.ContainerImageName != "" && config.ContainerRegistryURL != "" {
		destination := fmt.Sprintf("%s/%s:%s", config.ContainerRegistryURL, config.ContainerImageName, imageTag)
		buildOpts = append(buildOpts, "--tag", destination)

		commonPipelineEnvironment.container.registryURL = config.ContainerRegistryURL
		commonPipelineEnvironment.container.imageNameTag = fmt.Sprintf("%s:%s", config.ContainerImageName, imageTag)
		commonPipelineEnvironment.container.imageNameTags = append(commonPipelineEnvironment.container.imageNameTags, fmt.Sprintf("%s:%s", config.ContainerImageName, imageTag))
		commonPipelineEnvironment.container.imageNames = append(commonPipelineEnvironment.container.imageNames, config.ContainerImageName)

		// Set auth file for registry authentication
		buildOpts = append(buildOpts, "--authfile", fmt.Sprintf("%s/config.json", dockerConfigDir))
	} else {
		// Build without registry
		buildOpts = append(buildOpts, "--tag", fmt.Sprintf("%s:%s", config.ContainerImageName, imageTag))
	}

	// Add context directory as the final argument
	buildOpts = append(buildOpts, ".")

	log.Entry().Info("Executing buildah build...")
	err = execRunner.RunExecutable("buildah", buildOpts...)
	if err != nil {
		return fmt.Errorf("buildah build failed: %w", err)
	}

	// If registry is configured, push the image
	if config.ContainerImageName != "" && config.ContainerRegistryURL != "" {
		log.Entry().Info("Pushing image to registry...")
		pushOpts := []string{
			"push",
			"--authfile", fmt.Sprintf("%s/config.json", dockerConfigDir),
			fmt.Sprintf("%s/%s:%s", config.ContainerRegistryURL, config.ContainerImageName, imageTag),
		}
		err = execRunner.RunExecutable("buildah", pushOpts...)
		if err != nil {
			return fmt.Errorf("failed to push image: %w", err)
		}
	}

	log.Entry().Info("Buildah execution completed")

	if config.CreateBOM {
		log.Entry().Info("Generating bill of materials using syft...")
		if err := syft.GenerateSBOM(config.SyftDownloadURL, dockerConfigDir, execRunner, fileUtils, httpClient, commonPipelineEnvironment.container.registryURL, commonPipelineEnvironment.container.imageNameTags); err != nil {
			return fmt.Errorf("failed to generate BOM: %w", err)
		}
		log.Entry().Info("BOM generation completed")
	}

	return nil
}
