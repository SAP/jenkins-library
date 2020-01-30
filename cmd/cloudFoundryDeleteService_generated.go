package cmd

import (
	"os"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"

	"github.com/spf13/cobra"
)

type cloudFoundryDeleteServiceOptions struct {
	CfAPIEndpoint     string `json:"cfApiEndpoint,omitempty"`
	Username          string `json:"username,omitempty"`
	Password          string `json:"password,omitempty"`
	CfOrg             string `json:"cfOrg,omitempty"`
	CfSpace           string `json:"cfSpace,omitempty"`
	CfServiceInstance string `json:"cfServiceInstance,omitempty"`
}

var myCloudFoundryDeleteServiceOptions cloudFoundryDeleteServiceOptions

// CloudFoundryDeleteServiceCommand DeleteCloudFoundryService
func CloudFoundryDeleteServiceCommand() *cobra.Command {
	metadata := cloudFoundryDeleteServiceMetadata()

	var createCloudFoundryDeleteServiceCmd = &cobra.Command{
		Use:   "cloudFoundryDeleteService",
		Short: "DeleteCloudFoundryService",
		Long:  `Delete CloudFoundryService`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			log.SetStepName("cloudFoundryDeleteService")
			log.SetVerbose(GeneralConfig.Verbose)
			return PrepareConfig(cmd, &metadata, "cloudFoundryDeleteService", &myCloudFoundryDeleteServiceOptions, config.OpenPiperFile)
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			return cloudFoundryDeleteService(myCloudFoundryDeleteServiceOptions)
		},
	}

	addCloudFoundryDeleteServiceFlags(createCloudFoundryDeleteServiceCmd)
	return createCloudFoundryDeleteServiceCmd
}

func addCloudFoundryDeleteServiceFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&myCloudFoundryDeleteServiceOptions.CfAPIEndpoint, "cfApiEndpoint", os.Getenv("PIPER_cfApiEndpoint"), "Login API")
	cmd.Flags().StringVar(&myCloudFoundryDeleteServiceOptions.Username, "username", os.Getenv("PIPER_username"), "User E-Mail")
	cmd.Flags().StringVar(&myCloudFoundryDeleteServiceOptions.Password, "password", os.Getenv("PIPER_password"), "User Password")
	cmd.Flags().StringVar(&myCloudFoundryDeleteServiceOptions.CfOrg, "cfOrg", os.Getenv("PIPER_cfOrg"), "CF org")
	cmd.Flags().StringVar(&myCloudFoundryDeleteServiceOptions.CfSpace, "cfSpace", os.Getenv("PIPER_cfSpace"), "CF Space")
	cmd.Flags().StringVar(&myCloudFoundryDeleteServiceOptions.CfServiceInstance, "cfServiceInstance", os.Getenv("PIPER_cfServiceInstance"), "Parameter to delete CloudFoundry Service")

	cmd.MarkFlagRequired("cfApiEndpoint")
	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("cfOrg")
	cmd.MarkFlagRequired("cfSpace")
	cmd.MarkFlagRequired("cfServiceInstance")
}

// retrieve step metadata
func cloudFoundryDeleteServiceMetadata() config.StepData {
	var theMetaData = config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "cfApiEndpoint",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "cloudFoundry/apiEndpoint"}},
					},
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
						Name:        "cfOrg",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "cloudFoundry/org"}},
					},
					{
						Name:        "cfSpace",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "cloudFoundry/space"}},
					},
					{
						Name:        "cfServiceInstance",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "cloudFoundry/serviceInstance"}},
					},
				},
			},
		},
	}
	return theMetaData
}
