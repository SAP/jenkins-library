package cmd

import (
	"os"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"

	"github.com/spf13/cobra"
)

type cloudFoundryDeleteServiceOptions struct {
	API          string `json:"API,omitempty"`
	Username     string `json:"Username,omitempty"`
	Password     string `json:"Password,omitempty"`
	Organisation string `json:"Organisation,omitempty"`
	Space        string `json:"Space,omitempty"`
	ServiceName  string `json:"ServiceName,omitempty"`
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
	cmd.Flags().StringVar(&myCloudFoundryDeleteServiceOptions.API, "API", os.Getenv("PIPER_API"), "Login API")
	cmd.Flags().StringVar(&myCloudFoundryDeleteServiceOptions.Username, "Username", os.Getenv("PIPER_Username"), "User E-Mail")
	cmd.Flags().StringVar(&myCloudFoundryDeleteServiceOptions.Password, "Password", os.Getenv("PIPER_Password"), "User Password")
	cmd.Flags().StringVar(&myCloudFoundryDeleteServiceOptions.Organisation, "Organisation", os.Getenv("PIPER_Organisation"), "CF org")
	cmd.Flags().StringVar(&myCloudFoundryDeleteServiceOptions.Space, "Space", os.Getenv("PIPER_Space"), "CF Space")
	cmd.Flags().StringVar(&myCloudFoundryDeleteServiceOptions.ServiceName, "ServiceName", os.Getenv("PIPER_ServiceName"), "Parameter to delete CloudFoundry Service")

	cmd.MarkFlagRequired("API")
	cmd.MarkFlagRequired("Username")
	cmd.MarkFlagRequired("Password")
	cmd.MarkFlagRequired("Organisation")
	cmd.MarkFlagRequired("Space")
	cmd.MarkFlagRequired("ServiceName")
}

// retrieve step metadata
func cloudFoundryDeleteServiceMetadata() config.StepData {
	var theMetaData = config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "API",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "cloudFoundry/apiEndpoint"}},
					},
					{
						Name:        "Username",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "cloudFoundry/Username"}},
					},
					{
						Name:        "Password",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "cloudFoundry/Password"}},
					},
					{
						Name:        "Organisation",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "cloudFoundry/Organisation"}},
					},
					{
						Name:        "Space",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "cloudFoundry/Space"}},
					},
					{
						Name:        "ServiceName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "cloudFoundry/ServiceName"}},
					},
				},
			},
		},
	}
	return theMetaData
}
