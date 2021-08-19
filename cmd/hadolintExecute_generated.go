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

type hadolintExecuteOptions struct {
	ConfigurationURL          string   `json:"configurationUrl,omitempty"`
	ConfigurationUsername     string   `json:"configurationUsername,omitempty"`
	ConfigurationPassword     string   `json:"configurationPassword,omitempty"`
	DockerFile                string   `json:"dockerFile,omitempty"`
	ConfigurationFile         string   `json:"configurationFile,omitempty"`
	ReportFile                string   `json:"reportFile,omitempty"`
	CustomTLSCertificateLinks []string `json:"customTlsCertificateLinks,omitempty"`
	UploadReportsToGCS        bool     `json:"uploadReportsToGCS,omitempty"`
	GcpJSONKeyFilePath        string   `json:"gcpJsonKeyFilePath,omitempty"`
	GcsTargetFolder           string   `json:"gcsTargetFolder,omitempty"`
	GcsBucketID               string   `json:"gcsBucketId,omitempty"`
}

// HadolintExecuteCommand Executes the Haskell Dockerfile Linter which is a smarter Dockerfile linter that helps you build [best practice](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/) Docker images.
func HadolintExecuteCommand() *cobra.Command {
	const STEP_NAME = "hadolintExecute"

	metadata := hadolintExecuteMetadata()
	var stepConfig hadolintExecuteOptions
	var startTime time.Time
	var logCollector *log.CollectorHook

	var createHadolintExecuteCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Executes the Haskell Dockerfile Linter which is a smarter Dockerfile linter that helps you build [best practice](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/) Docker images.",
		Long: `Executes the Haskell Dockerfile Linter which is a smarter Dockerfile linter that helps you build [best practice](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/) Docker images.
The linter is parsing the Dockerfile into an abstract syntax tree (AST) and performs rules on top of the AST.`,
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
			log.RegisterSecret(stepConfig.ConfigurationUsername)
			log.RegisterSecret(stepConfig.ConfigurationPassword)
			log.RegisterSecret(stepConfig.GcpJSONKeyFilePath)

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
			hadolintExecute(stepConfig, &telemetryData)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addHadolintExecuteFlags(createHadolintExecuteCmd, &stepConfig)
	return createHadolintExecuteCmd
}

func addHadolintExecuteFlags(cmd *cobra.Command, stepConfig *hadolintExecuteOptions) {
	cmd.Flags().StringVar(&stepConfig.ConfigurationURL, "configurationUrl", os.Getenv("PIPER_configurationUrl"), "URL pointing to the .hadolint.yaml exclude configuration to be used for linting. Also have a look at `configurationFile` which could avoid central configuration download in case the file is part of your repository.")
	cmd.Flags().StringVar(&stepConfig.ConfigurationUsername, "configurationUsername", os.Getenv("PIPER_configurationUsername"), "The username to authenticate")
	cmd.Flags().StringVar(&stepConfig.ConfigurationPassword, "configurationPassword", os.Getenv("PIPER_configurationPassword"), "The password to authenticate")
	cmd.Flags().StringVar(&stepConfig.DockerFile, "dockerFile", `./Dockerfile`, "Dockerfile to be used for the assessment.")
	cmd.Flags().StringVar(&stepConfig.ConfigurationFile, "configurationFile", `.hadolint.yaml`, "Name of the configuration file used locally within the step. If a file with this name is detected as part of your repo downloading the central configuration via `configurationUrl` will be skipped. If you change the file's name make sure your stashing configuration also reflects this.")
	cmd.Flags().StringVar(&stepConfig.ReportFile, "reportFile", `hadolint.xml`, "Name of the result file used locally within the step.")
	cmd.Flags().StringSliceVar(&stepConfig.CustomTLSCertificateLinks, "customTlsCertificateLinks", []string{}, "List of download links to custom TLS certificates. This is required to ensure trusted connections between Piper and the system where the configuration file is to be downloaded from.")
	cmd.Flags().BoolVar(&stepConfig.UploadReportsToGCS, "uploadReportsToGCS", false, "Enables uploading reports to Google Cloud Storage bucket.")
	cmd.Flags().StringVar(&stepConfig.GcpJSONKeyFilePath, "gcpJsonKeyFilePath", os.Getenv("PIPER_gcpJsonKeyFilePath"), "File path to Google Cloud Platform JSON key file.")
	cmd.Flags().StringVar(&stepConfig.GcsTargetFolder, "gcsTargetFolder", os.Getenv("PIPER_gcsTargetFolder"), "Target folder in the GCS. If the value is empty, the source path is used.")
	cmd.Flags().StringVar(&stepConfig.GcsBucketID, "gcsBucketId", os.Getenv("PIPER_gcsBucketId"), "Bucket name for Google Cloud Storage.")

	cmd.MarkFlagRequired("gcsBucketId")
}

// retrieve step metadata
func hadolintExecuteMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "hadolintExecute",
			Aliases:     []config.Alias{},
			Description: "Executes the Haskell Dockerfile Linter which is a smarter Dockerfile linter that helps you build [best practice](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/) Docker images.",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "configurationCredentialsId", Description: "Jenkins 'Username with password' credentials ID containing username/password for access to your remote configuration file.", Type: "jenkins"},
					{Name: "gcpFileCredentialsId", Description: "Jenkins 'File' credentials ID containing the key file to authenticate to the Google Cloud Platform.", Type: "jenkins"},
				},
				Parameters: []config.StepParameters{
					{
						Name:        "configurationUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_configurationUrl"),
					},
					{
						Name: "configurationUsername",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "configurationCredentialsId",
								Param: "username",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_configurationUsername"),
					},
					{
						Name: "configurationPassword",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "configurationCredentialsId",
								Param: "password",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_configurationPassword"),
					},
					{
						Name:        "dockerFile",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "dockerfile"}},
						Default:     `./Dockerfile`,
					},
					{
						Name:        "configurationFile",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `.hadolint.yaml`,
					},
					{
						Name:        "reportFile",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `hadolint.xml`,
					},
					{
						Name:        "customTlsCertificateLinks",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name:        "uploadReportsToGCS",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPSs"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name: "gcpJsonKeyFilePath",
						ResourceRef: []config.ResourceReference{
							{
								Name: "gcpFileCredentialsId",
								Type: "secret",
							},

							{
								Name:  "",
								Paths: []string{"$(vaultPath)/gcp", "$(vaultBasePath)/$(vaultPipelineName)/gcp", "$(vaultBasePath)/GROUP-SECRETS/gcp"},
								Type:  "vaultSecretFile",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_gcpJsonKeyFilePath"),
					},
					{
						Name:        "gcsTargetFolder",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_gcsTargetFolder"),
					},
					{
						Name: "gcsBucketId",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "",
								Paths: []string{"$(vaultPath)/gcp", "$(vaultBasePath)/$(vaultPipelineName)/gcp", "$(vaultBasePath)/GROUP-SECRETS/gcp"},
								Type:  "vaultSecret",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_gcsBucketId"),
					},
				},
			},
			Containers: []config.Container{
				{Name: "hadolint", Image: "hadolint/hadolint:latest-debian"},
			},
		},
	}
	return theMetaData
}
