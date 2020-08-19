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

type npmExecuteLintOptions struct {
	FailOnError        bool   `json:"failOnError,omitempty"`
	DefaultNpmRegistry string `json:"defaultNpmRegistry,omitempty"`
}

// NpmExecuteLintCommand Execute ci-lint script on all npm packages in a project or execute default linting
func NpmExecuteLintCommand() *cobra.Command {
	const STEP_NAME = "npmExecuteLint"

	metadata := npmExecuteLintMetadata()
	var stepConfig npmExecuteLintOptions
	var startTime time.Time

	var createNpmExecuteLintCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Execute ci-lint script on all npm packages in a project or execute default linting",
		Long: `Execute ci-lint script for all package json files, if they implement the script. If no ci-lint script is defined,
either use ESLint configurations present in the project or use the provided general purpose configuration to run ESLint.`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			log.SetVerbose(GeneralConfig.Verbose)

			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

			err := PrepareConfig(cmd, &metadata, STEP_NAME, &stepConfig, config.OpenPiperFile)
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}

			if len(GeneralConfig.HookConfig.SentryConfig.Dsn) > 0 {
				sentryHook := log.NewSentryHook(GeneralConfig.HookConfig.SentryConfig.Dsn, GeneralConfig.CorrelationID)
				log.RegisterHook(&sentryHook)
			}

			return nil
		},
		Run: func(_ *cobra.Command, _ []string) {
			telemetryData := telemetry.CustomData{}
			telemetryData.ErrorCode = "1"
			handler := func() {
				telemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				telemetry.Send(&telemetryData)
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetry.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			npmExecuteLint(stepConfig, &telemetryData)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addNpmExecuteLintFlags(createNpmExecuteLintCmd, &stepConfig)
	return createNpmExecuteLintCmd
}

func addNpmExecuteLintFlags(cmd *cobra.Command, stepConfig *npmExecuteLintOptions) {
	cmd.Flags().BoolVar(&stepConfig.FailOnError, "failOnError", false, "Defines the behavior in case linting errors are found.")
	cmd.Flags().StringVar(&stepConfig.DefaultNpmRegistry, "defaultNpmRegistry", os.Getenv("PIPER_defaultNpmRegistry"), "URL of the npm registry to use. Defaults to https://registry.npmjs.org/")

}

// retrieve step metadata
func npmExecuteLintMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:    "npmExecuteLint",
			Aliases: []config.Alias{{Name: "executeNpm", Deprecated: false}},
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "failOnError",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "defaultNpmRegistry",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "GENERAL", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "npm/defaultNpmRegistry"}},
					},
				},
			},
		},
	}
	return theMetaData
}
