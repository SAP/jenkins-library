// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/spf13/cobra"
)

type containerSaveImageOptions struct {
	ContainerRegistryURL string `json:"containerRegistryUrl,omitempty"`
	ContainerImage       string `json:"containerImage,omitempty"`
	FilePath             string `json:"filePath,omitempty"`
	IncludeLayers        bool   `json:"includeLayers,omitempty"`
}

// ContainerSaveImageCommand Saves a container image as a tar file
func ContainerSaveImageCommand() *cobra.Command {
	const STEP_NAME = "containerSaveImage"

	metadata := containerSaveImageMetadata()
	var stepConfig containerSaveImageOptions
	var startTime time.Time

	var createContainerSaveImageCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Saves a container image as a tar file",
		Long: `This step allows you to save a container image, for example a Docker image into a tar file.

It can be used no matter if a Docker daemon is available or not. It will also work inside a Kubernetes cluster without access to a daemon.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			log.SetVerbose(GeneralConfig.Verbose)

			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

			err := PrepareConfig(cmd, &metadata, STEP_NAME, &stepConfig, config.OpenPiperFile)
			if err != nil {
				return err
			}

			if len(GeneralConfig.HookConfig.SentryConfig.Dsn) > 0 {
				sentryHook := log.NewSentryHook(GeneralConfig.HookConfig.SentryConfig.Dsn, GeneralConfig.CorrelationID)
				log.RegisterHook(&sentryHook)
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			telemetryData := telemetry.CustomData{}
			telemetryData.ErrorCode = "1"
			handler := func() {
				telemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				telemetry.Send(&telemetryData)
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetry.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			containerSaveImage(stepConfig, &telemetryData)
			telemetryData.ErrorCode = "0"
		},
	}

	addContainerSaveImageFlags(createContainerSaveImageCmd, &stepConfig)
	return createContainerSaveImageCmd
}

func addContainerSaveImageFlags(cmd *cobra.Command, stepConfig *containerSaveImageOptions) {
	cmd.Flags().StringVar(&stepConfig.ContainerRegistryURL, "containerRegistryUrl", os.Getenv("PIPER_containerRegistryUrl"), "The reference to the container registry where the image is located.")
	cmd.Flags().StringVar(&stepConfig.ContainerImage, "containerImage", os.Getenv("PIPER_containerImage"), "Container image to be saved.")
	cmd.Flags().StringVar(&stepConfig.FilePath, "filePath", os.Getenv("PIPER_filePath"), "The path to the file to which the image should be saved. Defaults to `imageName.tar`")
	cmd.Flags().BoolVar(&stepConfig.IncludeLayers, "includeLayers", false, "Flag if the docker layers should be included")

	cmd.MarkFlagRequired("containerRegistryUrl")
	cmd.MarkFlagRequired("containerImage")
}

// retrieve step metadata
func containerSaveImageMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:    "containerSaveImage",
			Aliases: []config.Alias{},
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "containerRegistryUrl",
						ResourceRef: []config.ResourceReference{{Name: "commonPipelineEnvironment", Param: "container/registryUrl"}},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "dockerRegistryUrl"}},
					},
					{
						Name:        "containerImage",
						ResourceRef: []config.ResourceReference{{Name: "commonPipelineEnvironment", Param: "container/imageNameTag"}},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "dockerImage"}, {Name: "scanImage"}},
					},
					{
						Name:        "filePath",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "includeLayers",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
				},
			},
		},
	}
	return theMetaData
}
