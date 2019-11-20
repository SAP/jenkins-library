package cmd

import (
	"os"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/spf13/cobra"
)

type xsDeployOptions struct {
	DeployOpts    string `json:"DeployOpts,omitempty"`
	MtaPath       string `json:"MtaPath,omitempty"`
	Action        string `json:"Action,omitempty"`
	Mode          string `json:"Mode,omitempty"`
	DeploymentID  string `json:"DeploymentID,omitempty"`
	APIURL        string `json:"ApiUrl,omitempty"`
	User          string `json:"User,omitempty"`
	Password      string `json:"Password,omitempty"`
	Org           string `json:"Org,omitempty"`
	Space         string `json:"Space,omitempty"`
	LoginOpts     string `json:"LoginOpts,omitempty"`
	XsSessionFile string `json:"XsSessionFile,omitempty"`
}

var myXsDeployOptions xsDeployOptions
var xsDeployStepConfigJSON string

// XsDeployCommand Performs xs deployment
func XsDeployCommand() *cobra.Command {
	metadata := xsDeployMetadata()
	var createXsDeployCmd = &cobra.Command{
		Use:   "xsDeploy",
		Short: "Performs xs deployment",
		Long:  `Performs xs deployment`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			log.SetStepName("xsDeploy")
			log.SetVerbose(GeneralConfig.Verbose)
			return PrepareConfig(cmd, &metadata, "xsDeploy", &myXsDeployOptions, OpenPiperFile)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return xsDeploy(myXsDeployOptions)
		},
	}

	addXsDeployFlags(createXsDeployCmd)
	return createXsDeployCmd
}

func addXsDeployFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&myXsDeployOptions.DeployOpts, "DeployOpts", os.Getenv("PIPER_DeployOpts"), "Additional deploy options")
	cmd.Flags().StringVar(&myXsDeployOptions.MtaPath, "MtaPath", os.Getenv("PIPER_MtaPath"), "Path to deployable")
	cmd.Flags().StringVar(&myXsDeployOptions.Action, "Action", "None", "The action")
	cmd.Flags().StringVar(&myXsDeployOptions.Mode, "Mode", "Deploy", "The mode")
	cmd.Flags().StringVar(&myXsDeployOptions.DeploymentID, "DeploymentID", os.Getenv("PIPER_DeploymentID"), "The deployment ID")
	cmd.Flags().StringVar(&myXsDeployOptions.APIURL, "ApiUrl", os.Getenv("PIPER_ApiUrl"), "The api url (e.g. https://example.org:12345")
	cmd.Flags().StringVar(&myXsDeployOptions.User, "User", os.Getenv("PIPER_User"), "User")
	cmd.Flags().StringVar(&myXsDeployOptions.Password, "Password", os.Getenv("PIPER_Password"), "Password")
	cmd.Flags().StringVar(&myXsDeployOptions.Org, "Org", os.Getenv("PIPER_Org"), "The org")
	cmd.Flags().StringVar(&myXsDeployOptions.Space, "Space", os.Getenv("PIPER_Space"), "The space")
	cmd.Flags().StringVar(&myXsDeployOptions.LoginOpts, "LoginOpts", os.Getenv("PIPER_LoginOpts"), "Additional options for performing xs login.")
	cmd.Flags().StringVar(&myXsDeployOptions.XsSessionFile, "XsSessionFile", os.Getenv("PIPER_XsSessionFile"), "The file keeping the xs session.")

	cmd.MarkFlagRequired("MtaPath")
	cmd.MarkFlagRequired("Mode")
	cmd.MarkFlagRequired("ApiUrl")
	cmd.MarkFlagRequired("User")
	cmd.MarkFlagRequired("Password")
	cmd.MarkFlagRequired("Org")
	cmd.MarkFlagRequired("Space")
	cmd.MarkFlagRequired("LoginOpts")
}

// retrieve step metadata
func xsDeployMetadata() config.StepData {
	var theMetaData = config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:      "DeployOpts",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
					},
					{
						Name:      "MtaPath",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
					},
					{
						Name:      "Action",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
					},
					{
						Name:      "Mode",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
					},
					{
						Name:      "DeploymentID",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
					},
					{
						Name:      "ApiUrl",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
					},
					{
						Name:      "User",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
					},
					{
						Name:      "Password",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
					},
					{
						Name:      "Org",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
					},
					{
						Name:      "Space",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
					},
					{
						Name:      "LoginOpts",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
					},
					{
						Name:      "XsSessionFile",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
					},
				},
			},
		},
	}
	return theMetaData
}
