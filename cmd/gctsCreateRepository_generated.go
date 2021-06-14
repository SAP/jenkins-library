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

type gctsCreateRepositoryOptions struct {
	Username            string `json:"username,omitempty"`
	Password            string `json:"password,omitempty"`
	Repository          string `json:"repository,omitempty"`
	Host                string `json:"host,omitempty"`
	Client              string `json:"client,omitempty"`
	RemoteRepositoryURL string `json:"remoteRepositoryURL,omitempty"`
	Role                string `json:"role,omitempty"`
	VSID                string `json:"vSID,omitempty"`
	Type                string `json:"type,omitempty"`
}

// GctsCreateRepositoryCommand Creates a Git repository on an ABAP system
func GctsCreateRepositoryCommand() *cobra.Command {
	const STEP_NAME = "gctsCreateRepository"

	metadata := gctsCreateRepositoryMetadata()
	var stepConfig gctsCreateRepositoryOptions
	var startTime time.Time
	var logCollector *log.CollectorHook

	var createGctsCreateRepositoryCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Creates a Git repository on an ABAP system",
		Long:  `Creates a local Git repository on an ABAP system if it does not already exist.`,
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
			log.RegisterSecret(stepConfig.Username)
			log.RegisterSecret(stepConfig.Password)

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
			gctsCreateRepository(stepConfig, &telemetryData)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addGctsCreateRepositoryFlags(createGctsCreateRepositoryCmd, &stepConfig)
	return createGctsCreateRepositoryCmd
}

func addGctsCreateRepositoryFlags(cmd *cobra.Command, stepConfig *gctsCreateRepositoryOptions) {
	cmd.Flags().StringVar(&stepConfig.Username, "username", os.Getenv("PIPER_username"), "Username to authenticate to the ABAP system")
	cmd.Flags().StringVar(&stepConfig.Password, "password", os.Getenv("PIPER_password"), "Password to authenticate to the ABAP system")
	cmd.Flags().StringVar(&stepConfig.Repository, "repository", os.Getenv("PIPER_repository"), "Specifies the name (ID) of the local repository on the ABAP system")
	cmd.Flags().StringVar(&stepConfig.Host, "host", os.Getenv("PIPER_host"), "Specifies the protocol and host address, including the port. Please provide in the format `<protocol>://<host>:<port>`. Supported protocols are `http` and `https`.")
	cmd.Flags().StringVar(&stepConfig.Client, "client", os.Getenv("PIPER_client"), "Specifies the client of the ABAP system to be addressed")
	cmd.Flags().StringVar(&stepConfig.RemoteRepositoryURL, "remoteRepositoryURL", os.Getenv("PIPER_remoteRepositoryURL"), "URL of the corresponding remote repository")
	cmd.Flags().StringVar(&stepConfig.Role, "role", os.Getenv("PIPER_role"), "Role of the local repository. Choose between 'TARGET' and 'SOURCE'. Local repositories with a TARGET role will NOT be able to be the source of code changes")
	cmd.Flags().StringVar(&stepConfig.VSID, "vSID", os.Getenv("PIPER_vSID"), "Virtual SID of the local repository. The vSID corresponds to the transport route that delivers content to the remote Git repository")
	cmd.Flags().StringVar(&stepConfig.Type, "type", `GIT`, "Type of the used source code management tool")

	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("repository")
	cmd.MarkFlagRequired("host")
	cmd.MarkFlagRequired("client")
}

// retrieve step metadata
func gctsCreateRepositoryMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "gctsCreateRepository",
			Aliases:     []config.Alias{},
			Description: "Creates a Git repository on an ABAP system",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "abapCredentialsId", Description: "Jenkins credentials ID containing username and password for authentication to the ABAP system on which you want to create the repository", Type: "jenkins"},
				},
				Parameters: []config.StepParameters{
					{
						Name: "username",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "abapCredentialsId",
								Param: "username",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Default:   os.Getenv("PIPER_username"),
						Aliases:   []config.Alias{},
					},
					{
						Name: "password",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "abapCredentialsId",
								Param: "password",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Default:   os.Getenv("PIPER_password"),
						Aliases:   []config.Alias{},
					},
					{
						Name:        "repository",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Default:     os.Getenv("PIPER_repository"),
						Aliases:     []config.Alias{},
					},
					{
						Name:        "host",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Default:     os.Getenv("PIPER_host"),
						Aliases:     []config.Alias{},
					},
					{
						Name:        "client",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Default:     os.Getenv("PIPER_client"),
						Aliases:     []config.Alias{},
					},
					{
						Name:        "remoteRepositoryURL",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Default:     os.Getenv("PIPER_remoteRepositoryURL"),
						Aliases:     []config.Alias{},
					},
					{
						Name:        "role",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Default:     os.Getenv("PIPER_role"),
						Aliases:     []config.Alias{},
					},
					{
						Name:        "vSID",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Default:     os.Getenv("PIPER_vSID"),
						Aliases:     []config.Alias{},
					},
					{
						Name:        "type",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Default:     `GIT`,
						Aliases:     []config.Alias{},
					},
				},
			},
		},
	}
	return theMetaData
}
