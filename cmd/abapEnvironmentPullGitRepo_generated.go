package cmd

import (
	"os"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"

	"github.com/spf13/cobra"
)

type abapEnvironmentPullGitRepoOptions struct {
	User              string `json:"user,omitempty"`
	Password          string `json:"password,omitempty"`
	RepositoryName    string `json:"repositoryName,omitempty"`
	Host              string `json:"host,omitempty"`
	CfAPIEndpoint     string `json:"cfApiEndpoint,omitempty"`
	CfOrg             string `json:"cfOrg,omitempty"`
	CfSpace           string `json:"cfSpace,omitempty"`
	CfServiceInstance string `json:"cfServiceInstance,omitempty"`
	CfServiceKey      string `json:"cfServiceKey,omitempty"`
}

var myAbapEnvironmentPullGitRepoOptions abapEnvironmentPullGitRepoOptions

// AbapEnvironmentPullGitRepoCommand Pulls a git repository to a SAP Cloud Platform ABAP Environment system
func AbapEnvironmentPullGitRepoCommand() *cobra.Command {
	metadata := abapEnvironmentPullGitRepoMetadata()

	var createAbapEnvironmentPullGitRepoCmd = &cobra.Command{
		Use:   "abapEnvironmentPullGitRepo",
		Short: "Pulls a git repository to a SAP Cloud Platform ABAP Environment system",
		Long:  `Pulls a git repository (Software Component) to a SAP Cloud Platform ABAP Environment system.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			log.SetStepName("abapEnvironmentPullGitRepo")
			log.SetVerbose(GeneralConfig.Verbose)
			return PrepareConfig(cmd, &metadata, "abapEnvironmentPullGitRepo", &myAbapEnvironmentPullGitRepoOptions, config.OpenPiperFile)
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			return abapEnvironmentPullGitRepo(myAbapEnvironmentPullGitRepoOptions)
		},
	}

	addAbapEnvironmentPullGitRepoFlags(createAbapEnvironmentPullGitRepoCmd)
	return createAbapEnvironmentPullGitRepoCmd
}

func addAbapEnvironmentPullGitRepoFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&myAbapEnvironmentPullGitRepoOptions.User, "user", os.Getenv("PIPER_user"), "User for either the Cloud Foundry API or the Communication Arrangement for SAP_COM_0510")
	cmd.Flags().StringVar(&myAbapEnvironmentPullGitRepoOptions.Password, "password", os.Getenv("PIPER_password"), "Password for either the Cloud Foundry API or the Communication Arrangement for SAP_COM_0510")
	cmd.Flags().StringVar(&myAbapEnvironmentPullGitRepoOptions.RepositoryName, "repositoryName", os.Getenv("PIPER_repositoryName"), "Specifies the name of the Repository (Software Component) on the SAP Cloud Platform ABAP Environment system")
	cmd.Flags().StringVar(&myAbapEnvironmentPullGitRepoOptions.Host, "host", os.Getenv("PIPER_host"), "Specifies the host address of the SAP Cloud Platform ABAP Environment system")
	cmd.Flags().StringVar(&myAbapEnvironmentPullGitRepoOptions.CfAPIEndpoint, "cfApiEndpoint", os.Getenv("PIPER_cfApiEndpoint"), "Cloud Foundry API Enpoint")
	cmd.Flags().StringVar(&myAbapEnvironmentPullGitRepoOptions.CfOrg, "cfOrg", os.Getenv("PIPER_cfOrg"), "Cloud Foundry target organization")
	cmd.Flags().StringVar(&myAbapEnvironmentPullGitRepoOptions.CfSpace, "cfSpace", os.Getenv("PIPER_cfSpace"), "Cloud Foundry target space")
	cmd.Flags().StringVar(&myAbapEnvironmentPullGitRepoOptions.CfServiceInstance, "cfServiceInstance", os.Getenv("PIPER_cfServiceInstance"), "Cloud Foundry Service Instance")
	cmd.Flags().StringVar(&myAbapEnvironmentPullGitRepoOptions.CfServiceKey, "cfServiceKey", os.Getenv("PIPER_cfServiceKey"), "Cloud Foundry Service Key")

	cmd.MarkFlagRequired("user")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("repositoryName")
}

// retrieve step metadata
func abapEnvironmentPullGitRepoMetadata() config.StepData {
	var theMetaData = config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "user",
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
						Name:        "repositoryName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
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
						Aliases:     []config.Alias{},
					},
					{
						Name:        "cfOrg",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "cfSpace",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "cfServiceInstance",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "cfServiceKey",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
				},
			},
		},
	}
	return theMetaData
}
