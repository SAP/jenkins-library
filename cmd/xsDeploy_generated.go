package cmd

import (
	"os"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/spf13/cobra"
)

type xsDeployOptions struct {
	DeployOpts    string `json:"deployOpts,omitempty"`
	MtaPath       string `json:"mtaPath,omitempty"`
	Action        string `json:"action,omitempty"`
	Mode          string `json:"mode,omitempty"`
	DeploymentID  string `json:"deploymentID,omitempty"`
	APIURL        string `json:"apiUrl,omitempty"`
	User          string `json:"user,omitempty"`
	Password      string `json:"password,omitempty"`
	Org           string `json:"org,omitempty"`
	Space         string `json:"space,omitempty"`
	LoginOpts     string `json:"loginOpts,omitempty"`
	XsSessionFile string `json:"xsSessionFile,omitempty"`
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
			return PrepareConfig(cmd, &metadata, "xsDeploy", &myXsDeployOptions, config.OpenPiperFile)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return xsDeploy(myXsDeployOptions)
		},
	}

	addXsDeployFlags(createXsDeployCmd)
	return createXsDeployCmd
}

func addXsDeployFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&myXsDeployOptions.DeployOpts, "deployOpts", os.Getenv("PIPER_deployOpts"), "Additional deploy options")
	cmd.Flags().StringVar(&myXsDeployOptions.MtaPath, "mtaPath", os.Getenv("PIPER_mtaPath"), "Path to deployable")
	cmd.Flags().StringVar(&myXsDeployOptions.Action, "action", "None", "The action")
	cmd.Flags().StringVar(&myXsDeployOptions.Mode, "mode", "Deploy", "The mode")
	cmd.Flags().StringVar(&myXsDeployOptions.DeploymentID, "deploymentID", os.Getenv("PIPER_deploymentID"), "The deployment ID")
	cmd.Flags().StringVar(&myXsDeployOptions.APIURL, "apiUrl", os.Getenv("PIPER_apiUrl"), "The api url (e.g. https://example.org:12345")
	cmd.Flags().StringVar(&myXsDeployOptions.User, "user", os.Getenv("PIPER_user"), "User")
	cmd.Flags().StringVar(&myXsDeployOptions.Password, "password", os.Getenv("PIPER_password"), "Password")
	cmd.Flags().StringVar(&myXsDeployOptions.Org, "org", os.Getenv("PIPER_org"), "The org")
	cmd.Flags().StringVar(&myXsDeployOptions.Space, "space", os.Getenv("PIPER_space"), "The space")
	cmd.Flags().StringVar(&myXsDeployOptions.LoginOpts, "loginOpts", os.Getenv("PIPER_loginOpts"), "Additional options for performing xs login.")
	cmd.Flags().StringVar(&myXsDeployOptions.XsSessionFile, "xsSessionFile", os.Getenv("PIPER_xsSessionFile"), "The file keeping the xs session.")

	cmd.MarkFlagRequired("mtaPath")
	cmd.MarkFlagRequired("mode")
	cmd.MarkFlagRequired("apiUrl")
	cmd.MarkFlagRequired("user")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("org")
	cmd.MarkFlagRequired("space")
	cmd.MarkFlagRequired("loginOpts")
}

// retrieve step metadata
func xsDeployMetadata() config.StepData {
	var theMetaData = config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:      "deployOpts",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "mtaPath",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "action",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "mode",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "deploymentID",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "apiUrl",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "user",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "password",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "org",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "space",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "loginOpts",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "xsSessionFile",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
				},
			},
		},
	}
	return theMetaData
}
