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

type npmExecuteScriptsOptions struct {
	Install            bool     `json:"install,omitempty"`
	RunScripts         []string `json:"runScripts,omitempty"`
	DefaultNpmRegistry string   `json:"defaultNpmRegistry,omitempty"`
	SapNpmRegistry     string   `json:"sapNpmRegistry,omitempty"`
}

// NpmExecuteScriptsCommand Execute npm run scripts with optional install before
func NpmExecuteScriptsCommand() *cobra.Command {
	metadata := npmExecuteScriptsMetadata()
	var stepConfig npmExecuteScriptsOptions
	var startTime time.Time

	var createNpmExecuteScriptsCmd = &cobra.Command{
		Use:   "npmExecuteScripts",
		Short: "Execute npm run scripts with optional install before",
		Long:  `Execute npm run scripts in all package json files, if they implement the scripts`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			startTime = time.Now()
			log.SetStepName("npmExecuteScripts")
			log.SetVerbose(GeneralConfig.Verbose)
			err := PrepareConfig(cmd, &metadata, "npmExecuteScripts", &stepConfig, config.OpenPiperFile)
			if err != nil {
				return err
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
			telemetry.Initialize(GeneralConfig.NoTelemetry, "npmExecuteScripts")
			npmExecuteScripts(stepConfig, &telemetryData)
			telemetryData.ErrorCode = "0"
		},
	}

	addNpmExecuteScriptsFlags(createNpmExecuteScriptsCmd, &stepConfig)
	return createNpmExecuteScriptsCmd
}

func addNpmExecuteScriptsFlags(cmd *cobra.Command, stepConfig *npmExecuteScriptsOptions) {
	cmd.Flags().BoolVar(&stepConfig.Install, "install", false, "Run npm install or similar commands depending on the project structure.")
	cmd.Flags().StringSliceVar(&stepConfig.RunScripts, "runScripts", []string{}, "List of additinal run scripts to execute from package.json.")
	cmd.Flags().StringVar(&stepConfig.DefaultNpmRegistry, "defaultNpmRegistry", os.Getenv("PIPER_defaultNpmRegistry"), "URL of the npm registry to use. Defaults to https://registry.npmjs.org/")
	cmd.Flags().StringVar(&stepConfig.SapNpmRegistry, "sapNpmRegistry", "https://npm.sap.com", "The default npm registry url to be used as the remote mirror for the SAP npm packages.")

}

// retrieve step metadata
func npmExecuteScriptsMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:    "npmExecuteScripts",
			Aliases: []config.Alias{},
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "install",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "runScripts",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
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
						Name:        "sapNpmRegistry",
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
