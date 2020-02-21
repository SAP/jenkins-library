package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/spf13/cobra"
)

type mtaBuildOptions struct {
	BuildTarget         string `json:"buildTarget,omitempty"`
	MtaBuildTool        string `json:"mtaBuildTool,omitempty"`
	MtarName            string `json:"mtarName,omitempty"`
	MtaJarLocation      string `json:"mtaJarLocation,omitempty"`
	Extensions          string `json:"extensions,omitempty"`
	Platform            string `json:"platform,omitempty"`
	ApplicationName     string `json:"applicationName,omitempty"`
	DefaultNpmRegistry  string `json:"defaultNpmRegistry,omitempty"`
	ProjectSettingsFile string `json:"projectSettingsFile,omitempty"`
	GlobalSettingsFile  string `json:"globalSettingsFile,omitempty"`
}

type mtaBuildCommonPipelineEnvironment struct {
	mtarFilePath string
}

func (p *mtaBuildCommonPipelineEnvironment) persist(path, resourceName string) {
	content := []struct {
		category string
		name     string
		value    string
	}{
		{category: "", name: "mtarFilePath", value: p.mtarFilePath},
	}

	errCount := 0
	for _, param := range content {
		err := piperenv.SetResourceParameter(path, resourceName, filepath.Join(param.category, param.name), param.value)
		if err != nil {
			log.Entry().WithError(err).Error("Error persisting piper environment.")
			errCount++
		}
	}
	if errCount > 0 {
		os.Exit(1)
	}
}

// MtaBuildCommand Performs an mta build
func MtaBuildCommand() *cobra.Command {
	metadata := mtaBuildMetadata()
	var stepConfig mtaBuildOptions
	var startTime time.Time
	var commonPipelineEnvironment mtaBuildCommonPipelineEnvironment

	var createMtaBuildCmd = &cobra.Command{
		Use:   "mtaBuild",
		Short: "Performs an mta build",
		Long:  `Executes the SAP Multitarget Application Archive Builder to create an mtar archive of the application.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			startTime = time.Now()
			log.SetStepName("mtaBuild")
			log.SetVerbose(GeneralConfig.Verbose)
			return PrepareConfig(cmd, &metadata, "mtaBuild", &stepConfig, config.OpenPiperFile)
		},
		Run: func(cmd *cobra.Command, args []string) {
			telemetryData := telemetry.CustomData{}
			telemetryData.ErrorCode = "1"
			handler := func() {
				commonPipelineEnvironment.persist(GeneralConfig.EnvRootPath, "commonPipelineEnvironment")
				telemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				telemetry.Send(&telemetryData)
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetry.Initialize(GeneralConfig.NoTelemetry, "mtaBuild")
			mtaBuild(stepConfig, &telemetryData, &commonPipelineEnvironment)
			telemetryData.ErrorCode = "0"
		},
	}

	addMtaBuildFlags(createMtaBuildCmd, &stepConfig)
	return createMtaBuildCmd
}

func addMtaBuildFlags(cmd *cobra.Command, stepConfig *mtaBuildOptions) {
	cmd.Flags().StringVar(&stepConfig.BuildTarget, "buildTarget", os.Getenv("PIPER_buildTarget"), "mtaBuildTool 'classic' only: The target platform to which the mtar can be deployed. Valid values: 'CF', 'NEO', 'XSA'.")
	cmd.Flags().StringVar(&stepConfig.MtaBuildTool, "mtaBuildTool", "cloudMbt", "Tool to use when building the MTA. Valid values: 'classic', 'cloudMbt'.")
	cmd.Flags().StringVar(&stepConfig.MtarName, "mtarName", os.Getenv("PIPER_mtarName"), "The name of the generated mtar file including its extension.")
	cmd.Flags().StringVar(&stepConfig.MtaJarLocation, "mtaJarLocation", os.Getenv("PIPER_mtaJarLocation"), "mtaBuildTool 'classic' only: The location of the SAP Multitarget Application Archive Builder jar file, including file name and extension. If you run on Docker, this must match the location of the jar file in the container as well.")
	cmd.Flags().StringVar(&stepConfig.Extensions, "extensions", os.Getenv("PIPER_extensions"), "The path to the extension descriptor file.")
	cmd.Flags().StringVar(&stepConfig.Platform, "platform", os.Getenv("PIPER_platform"), "mtaBuildTool 'cloudMbt' only: The target platform to which the mtar can be deployed.")
	cmd.Flags().StringVar(&stepConfig.ApplicationName, "applicationName", os.Getenv("PIPER_applicationName"), "The name of the application which is being built. If the parameter has been provided and no `mta.yaml` exists, the `mta.yaml` will be automatically generated using this parameter and the information (`name` and `version`) from 'package.json` before the actual build starts.")
	cmd.Flags().StringVar(&stepConfig.DefaultNpmRegistry, "defaultNpmRegistry", os.Getenv("PIPER_defaultNpmRegistry"), "Url to the npm registry that should be used for installing npm dependencies.")
	cmd.Flags().StringVar(&stepConfig.ProjectSettingsFile, "projectSettingsFile", os.Getenv("PIPER_projectSettingsFile"), "Path or url to the mvn settings file that should be used as project settings file.")
	cmd.Flags().StringVar(&stepConfig.GlobalSettingsFile, "globalSettingsFile", os.Getenv("PIPER_globalSettingsFile"), "Path or url to the mvn settings file that should be used as global settings file")

}

// retrieve step metadata
func mtaBuildMetadata() config.StepData {
	var theMetaData = config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "buildTarget",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "mtaBuildTool",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "mtarName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "mtaJarLocation",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "extensions",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "platform",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "applicationName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "defaultNpmRegistry",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "projectSettingsFile",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "globalSettingsFile",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
				},
			},
		},
	}
	return theMetaData
}
