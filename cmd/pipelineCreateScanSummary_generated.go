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
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/spf13/cobra"
)

type pipelineCreateScanSummaryOptions struct {
	FailedOnly     bool   `json:"failedOnly,omitempty" validate:""`
	OutputFilePath string `json:"outputFilePath,omitempty" validate:""`
	PipelineLink   string `json:"pipelineLink,omitempty" validate:""`
}

// PipelineCreateScanSummaryCommand Collect scan result information anc create a summary report
func PipelineCreateScanSummaryCommand() *cobra.Command {
	const STEP_NAME = "pipelineCreateScanSummary"

	metadata := pipelineCreateScanSummaryMetadata()
	var stepConfig pipelineCreateScanSummaryOptions
	var startTime time.Time
	var logCollector *log.CollectorHook

	var createPipelineCreateScanSummaryCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Collect scan result information anc create a summary report",
		Long: `This step allows you to create a summary report of your scan results.

It is for example used to create a markdown file which can be used to create a GitHub issue.`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			log.SetVerbose(GeneralConfig.Verbose)

			validation, err := validation.New()
			if err != nil {
				return err
			}
			if err := validation.ValidateStruct(stepConfig); err != nil {
				return err
			}

			GeneralConfig.GitHubAccessTokens = ResolveAccessTokens(GeneralConfig.GitHubTokens)

			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

			err = PrepareConfig(cmd, &metadata, STEP_NAME, &stepConfig, config.OpenPiperFile)
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
			pipelineCreateScanSummary(stepConfig, &telemetryData)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addPipelineCreateScanSummaryFlags(createPipelineCreateScanSummaryCmd, &stepConfig)
	return createPipelineCreateScanSummaryCmd
}

func addPipelineCreateScanSummaryFlags(cmd *cobra.Command, stepConfig *pipelineCreateScanSummaryOptions) {
	cmd.Flags().BoolVar(&stepConfig.FailedOnly, "failedOnly", false, "Defines if only failed scans should be included into the summary.")
	cmd.Flags().StringVar(&stepConfig.OutputFilePath, "outputFilePath", `scanSummary.md`, "Defines the filepath to the target file which will be created by the step.")
	cmd.Flags().StringVar(&stepConfig.PipelineLink, "pipelineLink", os.Getenv("PIPER_pipelineLink"), "Link to the pipeline (e.g. Jenkins job url) for reference in the scan summary.")

}

// retrieve step metadata
func pipelineCreateScanSummaryMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "pipelineCreateScanSummary",
			Aliases:     []config.Alias{},
			Description: "Collect scan result information anc create a summary report",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "failedOnly",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "outputFilePath",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `scanSummary.md`,
					},
					{
						Name:        "pipelineLink",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_pipelineLink"),
					},
				},
			},
		},
	}
	return theMetaData
}
