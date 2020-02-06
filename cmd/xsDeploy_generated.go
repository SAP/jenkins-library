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

type xsDeployOptions struct {
	DeployOpts            string `json:"deployOpts,omitempty"`
	OperationIDLogPattern string `json:"operationIdLogPattern,omitempty"`
	MtaPath               string `json:"mtaPath,omitempty"`
	Action                string `json:"action,omitempty"`
	Mode                  string `json:"mode,omitempty"`
	OperationID           string `json:"operationId,omitempty"`
	APIURL                string `json:"apiUrl,omitempty"`
	User                  string `json:"user,omitempty"`
	Password              string `json:"password,omitempty"`
	Org                   string `json:"org,omitempty"`
	Space                 string `json:"space,omitempty"`
	LoginOpts             string `json:"loginOpts,omitempty"`
	XsSessionFile         string `json:"xsSessionFile,omitempty"`
}

// XsDeployCommand Performs xs deployment
func XsDeployCommand() *cobra.Command {
	metadata := xsDeployMetadata()
	var stepConfig xsDeployOptions
	var startTime time.Time

	var createXsDeployCmd = &cobra.Command{
		Use:   "xsDeploy",
		Short: "Performs xs deployment",
		Long:  `Performs xs deployment`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			startTime = time.Now()
			log.SetStepName("xsDeploy")
			log.SetVerbose(GeneralConfig.Verbose)
			return PrepareConfig(cmd, &metadata, "xsDeploy", &stepConfig, config.OpenPiperFile)
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
			telemetry.Initialize(GeneralConfig.NoTelemetry, "xsDeploy")
			xsDeploy(stepConfig, &telemetryData)
			telemetryData.ErrorCode = "0"
		},
	}

	addXsDeployFlags(createXsDeployCmd, &stepConfig)
	return createXsDeployCmd
}

func addXsDeployFlags(cmd *cobra.Command, stepConfig *xsDeployOptions) {
	cmd.Flags().StringVar(&stepConfig.DeployOpts, "deployOpts", os.Getenv("PIPER_deployOpts"), "Additional options appended to the deploy command. Only needed for sophisticated cases. When provided it is the duty of the provider to ensure proper quoting / escaping.")
	cmd.Flags().StringVar(&stepConfig.OperationIDLogPattern, "operationIdLogPattern", "^.*xs bg-deploy -i (.*) -a.*$", "Regex pattern for retrieving the ID of the operation from the xs log.")
	cmd.Flags().StringVar(&stepConfig.MtaPath, "mtaPath", os.Getenv("PIPER_mtaPath"), "Path to deployable")
	cmd.Flags().StringVar(&stepConfig.Action, "action", "NONE", "Used for finalizing the blue-green deployment.")
	cmd.Flags().StringVar(&stepConfig.Mode, "mode", "DEPLOY", "Controls if there is a standard deployment or a blue green deployment. Values: 'DEPLOY', 'BG_DEPLOY'")
	cmd.Flags().StringVar(&stepConfig.OperationID, "operationId", os.Getenv("PIPER_operationId"), "The operation ID. Used in case of bg-deploy in order to resume or abort a previously started deployment.")
	cmd.Flags().StringVar(&stepConfig.APIURL, "apiUrl", os.Getenv("PIPER_apiUrl"), "The api url (e.g. https://example.org:12345")
	cmd.Flags().StringVar(&stepConfig.User, "user", os.Getenv("PIPER_user"), "User")
	cmd.Flags().StringVar(&stepConfig.Password, "password", os.Getenv("PIPER_password"), "Password")
	cmd.Flags().StringVar(&stepConfig.Org, "org", os.Getenv("PIPER_org"), "The org")
	cmd.Flags().StringVar(&stepConfig.Space, "space", os.Getenv("PIPER_space"), "The space")
	cmd.Flags().StringVar(&stepConfig.LoginOpts, "loginOpts", os.Getenv("PIPER_loginOpts"), "Additional options appended to the login command. Only needed for sophisticated cases. When provided it is the duty of the provider to ensure proper quoting / escaping.")
	cmd.Flags().StringVar(&stepConfig.XsSessionFile, "xsSessionFile", os.Getenv("PIPER_xsSessionFile"), "The file keeping the xs session.")

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
						Name:        "deployOpts",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "operationIdLogPattern",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "deployIdLogPattern"}},
					},
					{
						Name:        "mtaPath",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "action",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "mode",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "operationId",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "apiUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
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
						Name:        "org",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "space",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "loginOpts",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "xsSessionFile",
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
