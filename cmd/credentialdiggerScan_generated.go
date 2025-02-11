// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/gcp"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/spf13/cobra"
)

type credentialdiggerScanOptions struct {
	Repository       string   `json:"repository,omitempty"`
	Snapshot         string   `json:"snapshot,omitempty"`
	PrNumber         int      `json:"prNumber,omitempty"`
	ExportAll        bool     `json:"exportAll,omitempty"`
	APIURL           string   `json:"apiUrl,omitempty"`
	Debug            bool     `json:"debug,omitempty"`
	RulesDownloadURL string   `json:"rulesDownloadUrl,omitempty"`
	Models           []string `json:"models,omitempty"`
	Token            string   `json:"token,omitempty"`
	RulesFile        string   `json:"rulesFile,omitempty"`
}

// CredentialdiggerScanCommand Scan a repository on GitHub with Credential Digger
func CredentialdiggerScanCommand() *cobra.Command {
	const STEP_NAME = "credentialdiggerScan"

	metadata := credentialdiggerScanMetadata()
	var stepConfig credentialdiggerScanOptions
	var startTime time.Time
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createCredentialdiggerScanCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Scan a repository on GitHub with Credential Digger",
		Long: `This step allows you to scan a repository on Github using Credential Digger.

It can for example be used for DevSecOps scenarios to verify the source code does not contain hard-coded credentials before being merged or released for production.
It supports several scan flavors, i.e., full scans of a repo, scan of a snapshot, or scan of a pull request.`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			log.SetVerbose(GeneralConfig.Verbose)

			GeneralConfig.GitHubAccessTokens = ResolveAccessTokens(GeneralConfig.GitHubTokens)

			path, err := os.Getwd()
			if err != nil {
				return err
			}
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

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

			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 || len(GeneralConfig.HookConfig.SplunkConfig.ProdCriblEndpoint) > 0 {
				splunkClient = &splunk.Splunk{}
				logCollector = &log.CollectorHook{CorrelationID: GeneralConfig.CorrelationID}
				log.RegisterHook(logCollector)
			}

			if err = log.RegisterANSHookIfConfigured(GeneralConfig.CorrelationID); err != nil {
				log.Entry().WithError(err).Warn("failed to set up SAP Alert Notification Service log hook")
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
			vaultClient := config.GlobalVaultClient()
			if vaultClient != nil {
				defer vaultClient.MustRevokeToken()
			}

			stepTelemetryData := telemetry.CustomData{}
			stepTelemetryData.ErrorCode = "1"
			handler := func() {
				config.RemoveVaultSecretFiles()
				stepTelemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				stepTelemetryData.ErrorCategory = log.GetErrorCategory().String()
				stepTelemetryData.PiperCommitHash = GitCommit
				telemetryClient.SetData(&stepTelemetryData)
				telemetryClient.LogStepTelemetryData()
				if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
					splunkClient.Initialize(GeneralConfig.CorrelationID,
						GeneralConfig.HookConfig.SplunkConfig.Dsn,
						GeneralConfig.HookConfig.SplunkConfig.Token,
						GeneralConfig.HookConfig.SplunkConfig.Index,
						GeneralConfig.HookConfig.SplunkConfig.SendLogs)
					splunkClient.Send(telemetryClient.GetData(), logCollector)
				}
				if len(GeneralConfig.HookConfig.SplunkConfig.ProdCriblEndpoint) > 0 {
					splunkClient.Initialize(GeneralConfig.CorrelationID,
						GeneralConfig.HookConfig.SplunkConfig.ProdCriblEndpoint,
						GeneralConfig.HookConfig.SplunkConfig.ProdCriblToken,
						GeneralConfig.HookConfig.SplunkConfig.ProdCriblIndex,
						GeneralConfig.HookConfig.SplunkConfig.SendLogs)
					splunkClient.Send(telemetryClient.GetData(), logCollector)
				}
				if GeneralConfig.HookConfig.GCPPubSubConfig.Enabled {
					err := gcp.NewGcpPubsubClient(
						vaultClient,
						GeneralConfig.HookConfig.GCPPubSubConfig.ProjectNumber,
						GeneralConfig.HookConfig.GCPPubSubConfig.IdentityPool,
						GeneralConfig.HookConfig.GCPPubSubConfig.IdentityProvider,
						GeneralConfig.CorrelationID,
						GeneralConfig.HookConfig.OIDCConfig.RoleID,
					).Publish(GeneralConfig.HookConfig.GCPPubSubConfig.Topic, telemetryClient.GetDataBytes())
					if err != nil {
						log.Entry().WithError(err).Warn("event publish failed")
					}
				}
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetryClient.Initialize(STEP_NAME)
			credentialdiggerScan(stepConfig, &stepTelemetryData)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addCredentialdiggerScanFlags(createCredentialdiggerScanCmd, &stepConfig)
	return createCredentialdiggerScanCmd
}

func addCredentialdiggerScanFlags(cmd *cobra.Command, stepConfig *credentialdiggerScanOptions) {
	cmd.Flags().StringVar(&stepConfig.Repository, "repository", os.Getenv("PIPER_repository"), "URL of the GitHub repository (was name, but we need the url). In case it's missing, use the URL of the current repository.")
	cmd.Flags().StringVar(&stepConfig.Snapshot, "snapshot", os.Getenv("PIPER_snapshot"), "If set, scan the snapshot of the repository at this commit_id/branch.")
	cmd.Flags().IntVar(&stepConfig.PrNumber, "prNumber", 0, "If set, scan the pull request open with this number.")
	cmd.Flags().BoolVar(&stepConfig.ExportAll, "exportAll", false, "Export all the findings, i.e., including non-leaks.")
	cmd.Flags().StringVar(&stepConfig.APIURL, "apiUrl", `https://api.github.com`, "Set the GitHub API url. Needed for scanning a pull request.")
	cmd.Flags().BoolVar(&stepConfig.Debug, "debug", false, "Execute the scans in debug mode (i.e., print logs).")
	cmd.Flags().StringVar(&stepConfig.RulesDownloadURL, "rulesDownloadUrl", os.Getenv("PIPER_rulesDownloadUrl"), "URL where to download custom rules. The file published at this URL must be formatted as the default ruleset https://raw.githubusercontent.com/SAP/credential-digger/main/ui/backend/rules.yml")
	cmd.Flags().StringSliceVar(&stepConfig.Models, "models", []string{}, "Machine learning models to automatically verify the findings.")
	cmd.Flags().StringVar(&stepConfig.Token, "token", os.Getenv("PIPER_token"), "GitHub personal access token as per https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line")
	cmd.Flags().StringVar(&stepConfig.RulesFile, "rulesFile", `inputs/rules.yml`, "Name of the rules file used locally within the step. If a remote files for rules is declared as `rulesDownloadUrl`, the stashed file is ignored. If you change the file's name make sure your stashing configuration also reflects this.")

	cmd.MarkFlagRequired("apiUrl")
	cmd.MarkFlagRequired("token")
}

// retrieve step metadata
func credentialdiggerScanMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "credentialdiggerScan",
			Aliases:     []config.Alias{},
			Description: "Scan a repository on GitHub with Credential Digger",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "githubTokenCredentialsId", Description: "Jenkins 'Secret text' credentials ID containing token to authenticate to GitHub.", Type: "jenkins"},
				},
				Parameters: []config.StepParameters{
					{
						Name:        "repository",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "githubRepo"}},
						Default:     os.Getenv("PIPER_repository"),
					},
					{
						Name:        "snapshot",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_snapshot"),
					},
					{
						Name:        "prNumber",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "int",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     0,
					},
					{
						Name:        "exportAll",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
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
						Name:        "debug",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "verbose"}},
						Default:     false,
					},
					{
						Name:        "rulesDownloadUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_rulesDownloadUrl"),
					},
					{
						Name:        "models",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
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
					{
						Name:        "rulesFile",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `inputs/rules.yml`,
					},
				},
			},
			Containers: []config.Container{
				{Image: "saposs/credentialdigger:4.14.0"},
			},
			Outputs: config.StepOutputs{
				Resources: []config.StepResources{
					{
						Name: "report",
						Type: "report",
						Parameters: []map[string]interface{}{
							{"filePattern": "**/report*.csv", "type": "credentialdigger-report"},
						},
					},
				},
			},
		},
	}
	return theMetaData
}
