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

type abapEnvironmentPullGitRepoOptions struct {
	Username          string   `json:"username,omitempty"`
	Password          string   `json:"password,omitempty"`
	RepositoryNames   []string `json:"repositoryNames,omitempty"`
	Repositories      string   `json:"repositories,omitempty"`
	Host              string   `json:"host,omitempty"`
	CfAPIEndpoint     string   `json:"cfApiEndpoint,omitempty"`
	CfOrg             string   `json:"cfOrg,omitempty"`
	CfSpace           string   `json:"cfSpace,omitempty"`
	CfServiceInstance string   `json:"cfServiceInstance,omitempty"`
	CfServiceKeyName  string   `json:"cfServiceKeyName,omitempty"`
	IgnoreCommit      bool     `json:"ignoreCommit,omitempty"`
}

// AbapEnvironmentPullGitRepoCommand Pulls a git repository to a SAP Cloud Platform ABAP Environment system
func AbapEnvironmentPullGitRepoCommand() *cobra.Command {
	const STEP_NAME = "abapEnvironmentPullGitRepo"

	metadata := abapEnvironmentPullGitRepoMetadata()
	var stepConfig abapEnvironmentPullGitRepoOptions
	var startTime time.Time

	var createAbapEnvironmentPullGitRepoCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Pulls a git repository to a SAP Cloud Platform ABAP Environment system",
		Long: `Pulls a git repository (Software Component) to a SAP Cloud Platform ABAP Environment system.
Please provide either of the following options:

* The host and credentials the Cloud Platform ABAP Environment system itself. The credentials must be configured for the Communication Scenario [SAP_COM_0510](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/b04a9ae412894725a2fc539bfb1ca055.html).
* The Cloud Foundry parameters (API endpoint, organization, space), credentials, the service instance for the ABAP service and the service key for the Communication Scenario SAP_COM_0510.
* Only provide one of those options with the respective credentials. If all values are provided, the direct communication (via host) has priority.`,
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
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetry.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			abapEnvironmentPullGitRepo(stepConfig, &telemetryData)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addAbapEnvironmentPullGitRepoFlags(createAbapEnvironmentPullGitRepoCmd, &stepConfig)
	return createAbapEnvironmentPullGitRepoCmd
}

func addAbapEnvironmentPullGitRepoFlags(cmd *cobra.Command, stepConfig *abapEnvironmentPullGitRepoOptions) {
	cmd.Flags().StringVar(&stepConfig.Username, "username", os.Getenv("PIPER_username"), "User for either the Cloud Foundry API or the Communication Arrangement for SAP_COM_0510")
	cmd.Flags().StringVar(&stepConfig.Password, "password", os.Getenv("PIPER_password"), "Password for either the Cloud Foundry API or the Communication Arrangement for SAP_COM_0510")
	cmd.Flags().StringSliceVar(&stepConfig.RepositoryNames, "repositoryNames", []string{}, "Specifies a list of Repositories (Software Components) on the SAP Cloud Platform ABAP Environment system")
	cmd.Flags().StringVar(&stepConfig.Repositories, "repositories", os.Getenv("PIPER_repositories"), "Specifies a YAML file containing the repositories configuration")
	cmd.Flags().StringVar(&stepConfig.Host, "host", os.Getenv("PIPER_host"), "Specifies the host address of the SAP Cloud Platform ABAP Environment system")
	cmd.Flags().StringVar(&stepConfig.CfAPIEndpoint, "cfApiEndpoint", os.Getenv("PIPER_cfApiEndpoint"), "Cloud Foundry API Enpoint")
	cmd.Flags().StringVar(&stepConfig.CfOrg, "cfOrg", os.Getenv("PIPER_cfOrg"), "Cloud Foundry target organization")
	cmd.Flags().StringVar(&stepConfig.CfSpace, "cfSpace", os.Getenv("PIPER_cfSpace"), "Cloud Foundry target space")
	cmd.Flags().StringVar(&stepConfig.CfServiceInstance, "cfServiceInstance", os.Getenv("PIPER_cfServiceInstance"), "Cloud Foundry Service Instance")
	cmd.Flags().StringVar(&stepConfig.CfServiceKeyName, "cfServiceKeyName", os.Getenv("PIPER_cfServiceKeyName"), "Cloud Foundry Service Key")
	cmd.Flags().BoolVar(&stepConfig.IgnoreCommit, "ignoreCommit", false, "ingores a commit provided via the repositories file")

	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")
}

// retrieve step metadata
func abapEnvironmentPullGitRepoMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "abapEnvironmentPullGitRepo",
			Aliases:     []config.Alias{},
			Description: "Pulls a git repository to a SAP Cloud Platform ABAP Environment system",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "abapCredentialsId", Description: "Jenkins credentials ID containing user and password to authenticate to the Cloud Platform ABAP Environment system or the Cloud Foundry API", Type: "jenkins", Aliases: []config.Alias{{Name: "cfCredentialsId", Deprecated: false}, {Name: "credentialsId", Deprecated: false}}},
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
						Name:        "repositoryNames",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Default:     []string{},
						Aliases:     []config.Alias{},
					},
					{
						Name:        "repositories",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Default:     os.Getenv("PIPER_repositories"),
						Aliases:     []config.Alias{},
					},
					{
						Name:        "host",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Default:     os.Getenv("PIPER_host"),
						Aliases:     []config.Alias{},
					},
					{
						Name:        "cfApiEndpoint",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Default:     os.Getenv("PIPER_cfApiEndpoint"),
						Aliases:     []config.Alias{{Name: "cloudFoundry/apiEndpoint"}},
					},
					{
						Name:        "cfOrg",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Default:     os.Getenv("PIPER_cfOrg"),
						Aliases:     []config.Alias{{Name: "cloudFoundry/org"}},
					},
					{
						Name:        "cfSpace",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Default:     os.Getenv("PIPER_cfSpace"),
						Aliases:     []config.Alias{{Name: "cloudFoundry/space"}},
					},
					{
						Name:        "cfServiceInstance",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Default:     os.Getenv("PIPER_cfServiceInstance"),
						Aliases:     []config.Alias{{Name: "cloudFoundry/serviceInstance"}},
					},
					{
						Name:        "cfServiceKeyName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Default:     os.Getenv("PIPER_cfServiceKeyName"),
						Aliases:     []config.Alias{{Name: "cloudFoundry/serviceKey"}, {Name: "cloudFoundry/serviceKeyName"}, {Name: "cfServiceKey"}},
					},
					{
						Name:        "ignoreCommit",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Default:     false,
						Aliases:     []config.Alias{},
					},
				},
			},
			Containers: []config.Container{
				{Name: "cf", Image: "ppiper/cf-cli"},
			},
		},
	}
	return theMetaData
}
