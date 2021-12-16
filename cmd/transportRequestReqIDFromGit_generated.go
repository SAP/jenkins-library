// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/spf13/cobra"
)

type transportRequestReqIDFromGitOptions struct {
	GitFrom               string `json:"gitFrom,omitempty"`
	GitTo                 string `json:"gitTo,omitempty"`
	TransportRequestLabel string `json:"transportRequestLabel,omitempty"`
}

type transportRequestReqIDFromGitCommonPipelineEnvironment struct {
	custom struct {
		transportRequestID string
	}
}

func (p *transportRequestReqIDFromGitCommonPipelineEnvironment) persist(path, resourceName string) {
	content := []struct {
		category string
		name     string
		value    interface{}
	}{
		{category: "custom", name: "transportRequestId", value: p.custom.transportRequestID},
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
		log.Entry().Error("failed to persist Piper environment")
	}
}

// TransportRequestReqIDFromGitCommand Retrieves the transport request ID from Git repository
func TransportRequestReqIDFromGitCommand() *cobra.Command {
	const STEP_NAME = "transportRequestReqIDFromGit"

	metadata := transportRequestReqIDFromGitMetadata()
	var stepConfig transportRequestReqIDFromGitOptions
	var startTime time.Time
	var commonPipelineEnvironment transportRequestReqIDFromGitCommonPipelineEnvironment
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createTransportRequestReqIDFromGitCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Retrieves the transport request ID from Git repository",
		Long: `This step scans the commit messages of the Git repository for a pattern to retrieve the transport request ID.
It is primarily made for the transport request upload steps to provide the transport request ID by Git means.`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			log.SetVerbose(GeneralConfig.Verbose)

			GeneralConfig.GitHubAccessTokens = ResolveAccessTokens(GeneralConfig.GitHubTokens)

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
				splunkClient = &splunk.Splunk{}
				logCollector = &log.CollectorHook{CorrelationID: GeneralConfig.CorrelationID}
				log.RegisterHook(logCollector)
			}

			validation, err := validation.New(validation.WithJSONNamesForStructFields(), validation.WithPredefinedErrorMessages())
			if err != nil {
				return err
			}
			if err = validation.ValidateStruct(stepConfig); err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}

			return nil
		},
		Run: func(_ *cobra.Command, _ []string) {
			stepTelemetryData := telemetry.CustomData{}
			stepTelemetryData.ErrorCode = "1"
			handler := func() {
				commonPipelineEnvironment.persist(GeneralConfig.EnvRootPath, "commonPipelineEnvironment")
				config.RemoveVaultSecretFiles()
				stepTelemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				stepTelemetryData.ErrorCategory = log.GetErrorCategory().String()
				stepTelemetryData.PiperCommitHash = GitCommit
				telemetryClient.SetData(&stepTelemetryData)
				telemetryClient.Send()
				if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
					splunkClient.Send(telemetryClient.GetData(), logCollector)
				}
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetryClient.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				splunkClient.Initialize(GeneralConfig.CorrelationID,
					GeneralConfig.HookConfig.SplunkConfig.Dsn,
					GeneralConfig.HookConfig.SplunkConfig.Token,
					GeneralConfig.HookConfig.SplunkConfig.Index,
					GeneralConfig.HookConfig.SplunkConfig.SendLogs)
			}
			transportRequestReqIDFromGit(stepConfig, &stepTelemetryData, &commonPipelineEnvironment)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addTransportRequestReqIDFromGitFlags(createTransportRequestReqIDFromGitCmd, &stepConfig)
	return createTransportRequestReqIDFromGitCmd
}

func addTransportRequestReqIDFromGitFlags(cmd *cobra.Command, stepConfig *transportRequestReqIDFromGitOptions) {
	cmd.Flags().StringVar(&stepConfig.GitFrom, "gitFrom", `origin/master`, "GIT starting point for retrieving the transport request ID")
	cmd.Flags().StringVar(&stepConfig.GitTo, "gitTo", `HEAD`, "GIT ending point for retrieving the transport request ID")
	cmd.Flags().StringVar(&stepConfig.TransportRequestLabel, "transportRequestLabel", `TransportRequest`, "Pattern used for identifying lines holding the transport request ID. The GIT commit log messages are scanned for this label")

}

// retrieve step metadata
func transportRequestReqIDFromGitMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "transportRequestReqIDFromGit",
			Aliases:     []config.Alias{},
			Description: "Retrieves the transport request ID from Git repository",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "gitFrom",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "changeManagement/git/from"}},
						Default:     `origin/master`,
					},
					{
						Name:        "gitTo",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "changeManagement/git/to"}},
						Default:     `HEAD`,
					},
					{
						Name:        "transportRequestLabel",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "changeManagement/transportRequestLabel"}},
						Default:     `TransportRequest`,
					},
				},
			},
			Outputs: config.StepOutputs{
				Resources: []config.StepResources{
					{
						Name: "commonPipelineEnvironment",
						Type: "piperEnvironment",
						Parameters: []map[string]interface{}{
							{"name": "custom/transportRequestId"},
						},
					},
				},
			},
		},
	}
	return theMetaData
}
