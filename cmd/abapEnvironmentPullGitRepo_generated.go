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
	Host              string   `json:"host,omitempty"`
	CfAPIEndpoint     string   `json:"cfApiEndpoint,omitempty"`
	CfOrg             string   `json:"cfOrg,omitempty"`
	CfSpace           string   `json:"cfSpace,omitempty"`
	CfServiceInstance string   `json:"cfServiceInstance,omitempty"`
	CfServiceKey      string   `json:"cfServiceKey,omitempty"`
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

* The host and credentials the Cloud Platform ABAP Environment system itself. The credentials must be configured for the Communication Scenario SAP_COM_0510.
* The Cloud Foundry parameters (API endpoint, organization, space), credentials, the service instance for the ABAP service and the service key for the Communication Scenario SAP_COM_0510.
* Only provide one of those options with the respective credentials. If all values are provided, the direct communication (via host) has priority.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			log.SetVerbose(GeneralConfig.Verbose)

			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

			err := PrepareConfig(cmd, &metadata, STEP_NAME, &stepConfig, config.OpenPiperFile)
			if err != nil {
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
		Run: func(cmd *cobra.Command, args []string) {
			telemetryData := telemetry.CustomData{}
			telemetryData.ErrorCode = "1"
			handler := func() {
				telemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				telemetry.Send(&telemetryData)
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetry.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			abapEnvironmentPullGitRepo(stepConfig, &telemetryData)
			telemetryData.ErrorCode = "0"
		},
	}

	addAbapEnvironmentPullGitRepoFlags(createAbapEnvironmentPullGitRepoCmd, &stepConfig)
	return createAbapEnvironmentPullGitRepoCmd
}

func addAbapEnvironmentPullGitRepoFlags(cmd *cobra.Command, stepConfig *abapEnvironmentPullGitRepoOptions) {
	cmd.Flags().StringVar(&stepConfig.Username, "username", os.Getenv("PIPER_username"), "User for either the Cloud Foundry API or the Communication Arrangement for SAP_COM_0510")
	cmd.Flags().StringVar(&stepConfig.Password, "password", os.Getenv("PIPER_password"), "Password for either the Cloud Foundry API or the Communication Arrangement for SAP_COM_0510")
	cmd.Flags().StringSliceVar(&stepConfig.RepositoryNames, "repositoryNames", []string{}, "Specifies the name of the Repository (Software Component) on the SAP Cloud Platform ABAP Environment system")
	cmd.Flags().StringVar(&stepConfig.Host, "host", os.Getenv("PIPER_host"), "Specifies the host address of the SAP Cloud Platform ABAP Environment system")
	cmd.Flags().StringVar(&stepConfig.CfAPIEndpoint, "cfApiEndpoint", os.Getenv("PIPER_cfApiEndpoint"), "Cloud Foundry API Enpoint")
	cmd.Flags().StringVar(&stepConfig.CfOrg, "cfOrg", os.Getenv("PIPER_cfOrg"), "Cloud Foundry target organization")
	cmd.Flags().StringVar(&stepConfig.CfSpace, "cfSpace", os.Getenv("PIPER_cfSpace"), "Cloud Foundry target space")
	cmd.Flags().StringVar(&stepConfig.CfServiceInstance, "cfServiceInstance", os.Getenv("PIPER_cfServiceInstance"), "Cloud Foundry Service Instance")
	cmd.Flags().StringVar(&stepConfig.CfServiceKey, "cfServiceKey", os.Getenv("PIPER_cfServiceKey"), "Cloud Foundry Service Key")

	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("repositoryNames")
}

// retrieve step metadata
func abapEnvironmentPullGitRepoMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:    "abapEnvironmentPullGitRepo",
			Aliases: []config.Alias{},
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "username",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "password",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "repositoryNames",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "host",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "cfApiEndpoint",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/apiEndpoint"}},
					},
					{
						Name:        "cfOrg",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/org"}},
					},
					{
						Name:        "cfSpace",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/space"}},
					},
					{
						Name:        "cfServiceInstance",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/serviceInstance"}},
					},
					{
						Name:        "cfServiceKey",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/serviceKey"}},
					},
				},
			},
		},
	}
	return theMetaData
}
