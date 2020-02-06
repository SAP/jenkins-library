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
	BuildTarget            string `json:"buildTarget,omitempty"`
	MtaBuildTool           string `json:"mtaBuildTool,omitempty"`
	Extensions             string `json:"extensions,omitempty"`
	Platform               string `json:"platform,omitempty"`
	ApplicationName        string `json:"applicationName,omitempty"`
	DefaultNpmRegistry     string `json:"defaultNpmRegistry,omitempty"`
	ProjectSettingsFileSrc string `json:"projectSettingsFileSrc,omitempty"`
	GlobalSettingsFileSrc  string `json:"globalSettingsFileSrc,omitempty"`
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
		Long:  `Performs an mta build`,
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
	cmd.Flags().StringVar(&stepConfig.BuildTarget, "buildTarget", os.Getenv("PIPER_buildTarget"), "For mtaBuildTool classic only. Valid values: CF, NEO, XSA")
	cmd.Flags().StringVar(&stepConfig.MtaBuildTool, "mtaBuildTool", "cloudMbt", "Valid values: 'classic', 'cloudMbt' (default)")
	cmd.Flags().StringVar(&stepConfig.Extensions, "extensions", os.Getenv("PIPER_extensions"), "Lorem ipsum")
	cmd.Flags().StringVar(&stepConfig.Platform, "platform", os.Getenv("PIPER_platform"), "Lorem ipsum")
	cmd.Flags().StringVar(&stepConfig.ApplicationName, "applicationName", os.Getenv("PIPER_applicationName"), "Lorem ipsum")
	cmd.Flags().StringVar(&stepConfig.DefaultNpmRegistry, "defaultNpmRegistry", os.Getenv("PIPER_defaultNpmRegistry"), "Lorem ipsum")
	cmd.Flags().StringVar(&stepConfig.ProjectSettingsFileSrc, "projectSettingsFileSrc", os.Getenv("PIPER_projectSettingsFileSrc"), "Lorem ipsum")
	cmd.Flags().StringVar(&stepConfig.GlobalSettingsFileSrc, "globalSettingsFileSrc", os.Getenv("PIPER_globalSettingsFileSrc"), "Lorem ipsum")

	cmd.MarkFlagRequired("buildTarget")
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
						Mandatory:   true,
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
						Name:        "projectSettingsFileSrc",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "globalSettingsFileSrc",
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
