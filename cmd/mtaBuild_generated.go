package cmd

import (
	"os"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/spf13/cobra"
)

type mtaBuildOptions struct {
	Extension           string `json:"extension,omitempty"`
	MtaBuildTool        string `json:"mtaBuildTool,omitempty"`
	MtaJarLocation      string `json:"mtaJarLocation,omitempty"`
	MtarName            string `json:"mtarName,omitempty"`
	DefaultNpmRegistry  string `json:"defaultNpmRegistry,omitempty"`
	GlobalSettingsFile  string `json:"globalSettingsFile,omitempty"`
	Platform            string `json:"platform,omitempty"`
	ProjectSettingsFile string `json:"projectSettingsFile,omitempty"`
	ApplicationName     string `json:"applicationName,omitempty"`
	BuildTarget         string `json:"buildTarget,omitempty"`
}

var myMtaBuildOptions mtaBuildOptions
var mtaBuildStepConfigJSON string

// MtaBuildCommand Executes the SAP Multitarget Application Archive Builder to create an mtar archive of the application.
func MtaBuildCommand() *cobra.Command {
	metadata := mtaBuildMetadata()
	var createMtaBuildCmd = &cobra.Command{
		Use:   "mtaBuild",
		Short: "Executes the SAP Multitarget Application Archive Builder to create an mtar archive of the application.",
		Long:  `Executes the SAP Multitarget Application Archive Builder to create an mtar archive of the application.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			log.SetStepName("mtaBuild")
			log.SetVerbose(GeneralConfig.Verbose)
			return PrepareConfig(cmd, &metadata, "mtaBuild", &myMtaBuildOptions, config.OpenPiperFile)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return mtaBuild(myMtaBuildOptions)
		},
	}

	addMtaBuildFlags(createMtaBuildCmd)
	return createMtaBuildCmd
}

func addMtaBuildFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&myMtaBuildOptions.Extension, "extension", os.Getenv("PIPER_extension"), "The path to the extension descriptor file.")
	cmd.Flags().StringVar(&myMtaBuildOptions.MtaBuildTool, "mtaBuildTool", "classic", "Tool to use when building the MTA")
	cmd.Flags().StringVar(&myMtaBuildOptions.MtaJarLocation, "mtaJarLocation", "/opt/sap/mta/lib/mta.jar", "The location of the SAP Multitarget Application Archive Builder jar file, including file name and extension. If it is not provided, the SAP Multitarget Application Archive Builder is expected on PATH.")
	cmd.Flags().StringVar(&myMtaBuildOptions.MtarName, "mtarName", os.Getenv("PIPER_mtarName"), "The name of the generated mtar file including its extension.")
	cmd.Flags().StringVar(&myMtaBuildOptions.DefaultNpmRegistry, "defaultNpmRegistry", os.Getenv("PIPER_defaultNpmRegistry"), "Url to the npm registry that should be used for installing npm dependencies.")
	cmd.Flags().StringVar(&myMtaBuildOptions.GlobalSettingsFile, "globalSettingsFile", os.Getenv("PIPER_globalSettingsFile"), "Path or url to the mvn settings file that should be used as global settings file.")
	cmd.Flags().StringVar(&myMtaBuildOptions.Platform, "platform", "cf", "mtaBuildTool cloudMbt only: The target platform to which the mtar can be deployed.")
	cmd.Flags().StringVar(&myMtaBuildOptions.ProjectSettingsFile, "projectSettingsFile", os.Getenv("PIPER_projectSettingsFile"), "Path or url to the mvn settings file that should be used as project settings file.")
	cmd.Flags().StringVar(&myMtaBuildOptions.ApplicationName, "applicationName", os.Getenv("PIPER_applicationName"), "The name of the application which is being built. If the parameter has been provided and no `mta.yaml` exists, the `mta.yaml` will be automatically generated using this parameter and the information (`name` and `version`) from `package.json` before the actual build starts.")
	cmd.Flags().StringVar(&myMtaBuildOptions.BuildTarget, "buildTarget", "NEO", "The target platform to which the mtar can be deployed.")

}

// retrieve step metadata
func mtaBuildMetadata() config.StepData {
	var theMetaData = config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:      "extension",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "mtaBuildTool",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "mtaJarLocation",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "mtarName",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "defaultNpmRegistry",
						Scope:     []string{"STAGES"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "globalSettingsFile",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "platform",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "projectSettingsFile",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "applicationName",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "buildTarget",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
				},
			},
		},
	}
	return theMetaData
}
