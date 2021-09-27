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

type githubSetCommitStatusOptions struct {
	APIURL      string `json:"apiUrl,omitempty"`
	CommitID    string `json:"commitId,omitempty"`
	Context     string `json:"context,omitempty"`
	Description string `json:"description,omitempty"`
	Owner       string `json:"owner,omitempty"`
	Repository  string `json:"repository,omitempty"`
	Status      string `json:"status,omitempty" validate:"oneof=failure pending success"`
	TargetURL   string `json:"targetUrl,omitempty"`
	Token       string `json:"token,omitempty"`
}

// GithubSetCommitStatusCommand Set a status of a certain commit.
func GithubSetCommitStatusCommand() *cobra.Command {
	const STEP_NAME = "githubSetCommitStatus"

	metadata := githubSetCommitStatusMetadata()
	var stepConfig githubSetCommitStatusOptions
	var startTime time.Time
	var logCollector *log.CollectorHook

	var createGithubSetCommitStatusCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Set a status of a certain commit.",
		Long: `This step allows you to set a status for a certain commit.
Details can be found here: https://developer.github.com/v3/repos/statuses/.

Typically, following information is set:

* state (pending, failure, success)
* context
* target URL (link to details)

It can for example be used to create additional check indicators for a pull request which can be evaluated and also be enforced by GitHub configuration.`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			log.SetVerbose(GeneralConfig.Verbose)

			GeneralConfig.GitHubAccessTokens = ResolveAccessTokens(GeneralConfig.GitHubTokens)

			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

			validation, err := validation.New()
			if err != nil {
				return err
			}
			if err := validation.ValidateStruct(stepConfig); err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}

			err = PrepareConfig(cmd, &metadata, STEP_NAME, &stepConfig, config.OpenPiperFile)
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}
			log.RegisterSecret(stepConfig.Token)

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
			githubSetCommitStatus(stepConfig, &telemetryData)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addGithubSetCommitStatusFlags(createGithubSetCommitStatusCmd, &stepConfig)
	return createGithubSetCommitStatusCmd
}

func addGithubSetCommitStatusFlags(cmd *cobra.Command, stepConfig *githubSetCommitStatusOptions) {
	cmd.Flags().StringVar(&stepConfig.APIURL, "apiUrl", `https://api.github.com`, "Set the GitHub API URL.")
	cmd.Flags().StringVar(&stepConfig.CommitID, "commitId", os.Getenv("PIPER_commitId"), "The commitId for which the status should be set.")
	cmd.Flags().StringVar(&stepConfig.Context, "context", os.Getenv("PIPER_context"), "Label for the status which will for example show up in a pull request.")
	cmd.Flags().StringVar(&stepConfig.Description, "description", os.Getenv("PIPER_description"), "Short description of the status.")
	cmd.Flags().StringVar(&stepConfig.Owner, "owner", os.Getenv("PIPER_owner"), "Name of the GitHub organization.")
	cmd.Flags().StringVar(&stepConfig.Repository, "repository", os.Getenv("PIPER_repository"), "Name of the GitHub repository.")
	cmd.Flags().StringVar(&stepConfig.Status, "status", os.Getenv("PIPER_status"), "Status which should be set on the commitId.")
	cmd.Flags().StringVar(&stepConfig.TargetURL, "targetUrl", os.Getenv("PIPER_targetUrl"), "Target URL to associate the status with.")
	cmd.Flags().StringVar(&stepConfig.Token, "token", os.Getenv("PIPER_token"), "GitHub personal access token as per https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line.")

	cmd.MarkFlagRequired("apiUrl")
	cmd.MarkFlagRequired("commitId")
	cmd.MarkFlagRequired("context")
	cmd.MarkFlagRequired("owner")
	cmd.MarkFlagRequired("repository")
	cmd.MarkFlagRequired("status")
	cmd.MarkFlagRequired("token")
}

// retrieve step metadata
func githubSetCommitStatusMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "githubSetCommitStatus",
			Aliases:     []config.Alias{},
			Description: "Set a status of a certain commit.",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "githubTokenCredentialsId", Description: "Jenkins 'Secret text' credentials ID containing token to authenticate to GitHub.", Type: "jenkins"},
				},
				Parameters: []config.StepParameters{
					{
						Name:        "apiUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "githubApiUrl"}},
						Default:     `https://api.github.com`,
					},
					{
						Name: "commitId",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "git/commitId",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_commitId"),
					},
					{
						Name:        "context",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_context"),
					},
					{
						Name:        "description",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_description"),
					},
					{
						Name: "owner",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "github/owner",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{{Name: "githubOrg"}},
						Default:   os.Getenv("PIPER_owner"),
					},
					{
						Name: "repository",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "github/repository",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{{Name: "githubRepo"}},
						Default:   os.Getenv("PIPER_repository"),
					},
					{
						Name:        "status",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_status"),
					},
					{
						Name:        "targetUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_targetUrl"),
					},
					{
						Name: "token",
						ResourceRef: []config.ResourceReference{
							{
								Name: "githubTokenCredentialsId",
								Type: "secret",
							},

							{
								Name:    "githubVaultSecretName",
								Type:    "vaultSecret",
								Default: "github",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{{Name: "githubToken"}, {Name: "access_token"}},
						Default:   os.Getenv("PIPER_token"),
					},
				},
			},
		},
	}
	return theMetaData
}
