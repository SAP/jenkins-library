// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/spf13/cobra"
)

type terraformExecuteOptions struct {
	Command          string   `json:"command,omitempty"`
	TerraformSecrets string   `json:"terraformSecrets,omitempty"`
	AdditionalArgs   []string `json:"additionalArgs,omitempty"`
}

// TerraformExecuteCommand Executes Terraform
func TerraformExecuteCommand() *cobra.Command {
	const STEP_NAME = "terraformExecute"

	metadata := terraformExecuteMetadata()
	var stepConfig terraformExecuteOptions
	var startTime time.Time
	var logCollector *log.CollectorHook

	var createTerraformExecuteCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Executes Terraform",
		Long:  `This step executes the terraform binary with the given command, and is able to fetch additional variables from vault.`,
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

			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				logCollector = &log.CollectorHook{CorrelationID: GeneralConfig.CorrelationID}
				log.RegisterHook(logCollector)
			}

			return nil
		},
		Run: func(_ *cobra.Command, _ []string) {
			telemetryData := telemetry.CustomData{}
			telemetryData.ErrorCode = "1"
			handler := func() {
				config.RemoveVaultSecretFiles()
				telemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				telemetryData.ErrorCategory = log.GetErrorCategory().String()
				telemetry.Send(&telemetryData)
				if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
					splunk.Send(&telemetryData, logCollector)
				}
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetry.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				splunk.Initialize(GeneralConfig.CorrelationID,
					GeneralConfig.HookConfig.SplunkConfig.Dsn,
					GeneralConfig.HookConfig.SplunkConfig.Token,
					GeneralConfig.HookConfig.SplunkConfig.Index,
					GeneralConfig.HookConfig.SplunkConfig.SendLogs)
			}
			terraformExecute(stepConfig, &telemetryData)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addTerraformExecuteFlags(createTerraformExecuteCmd, &stepConfig)
	return createTerraformExecuteCmd
}

func addTerraformExecuteFlags(cmd *cobra.Command, stepConfig *terraformExecuteOptions) {
	cmd.Flags().StringVar(&stepConfig.Command, "command", `plan`, "")
	cmd.Flags().StringVar(&stepConfig.TerraformSecrets, "terraformSecrets", os.Getenv("PIPER_terraformSecrets"), "")
	cmd.Flags().StringSliceVar(&stepConfig.AdditionalArgs, "additionalArgs", []string{}, "")

}

// retrieve step metadata
func terraformExecuteMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "terraformExecute",
			Aliases:     []config.Alias{},
			Description: "Executes Terraform",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "command",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name: "terraformSecrets",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "",
								Paths: []string{"$(vaultPath)/terraformExecute", "$(vaultBasePath)/$(vaultPipelineName)/terraformExecute", "$(vaultBasePath)/GROUP-SECRETS/terraformExecute"},
								Type:  "vaultSecretFile",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:        "additionalArgs",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
				},
			},
			Containers: []config.Container{
				{Name: "terraform", Image: "hashicorp/terraform:0.14.7"},
			},
		},
	}
	return theMetaData
}
