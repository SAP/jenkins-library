// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
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
	Username              string `json:"username,omitempty"`
	Password              string `json:"password,omitempty"`
	Org                   string `json:"org,omitempty"`
	Space                 string `json:"space,omitempty"`
	LoginOpts             string `json:"loginOpts,omitempty"`
	XsSessionFile         string `json:"xsSessionFile,omitempty"`
}

type xsDeployCommonPipelineEnvironment struct {
	operationID string
}

func (p *xsDeployCommonPipelineEnvironment) persist(path, resourceName string) {
	content := []struct {
		category string
		name     string
		value    interface{}
	}{
		{category: "", name: "operationId", value: p.operationID},
	}

	errCount := 0
	for _, param := range content {
		err := piperenv.SetResourceParameter(path, resourceName, filepath.Join(param.category, param.name), param.value)
		if err != nil {
			log.Entry().WithError(err).Error("Error persisting piper environment.")
			errCount++
		}
	}
	if errCount > 0 {
		log.Entry().Fatal("failed to persist Piper environment")
	}
}

// XsDeployCommand Performs xs deployment
func XsDeployCommand() *cobra.Command {
	const STEP_NAME = "xsDeploy"

	metadata := xsDeployMetadata()
	var stepConfig xsDeployOptions
	var startTime time.Time
	var commonPipelineEnvironment xsDeployCommonPipelineEnvironment

	var createXsDeployCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Performs xs deployment",
		Long:  `Performs xs deployment`,
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
				commonPipelineEnvironment.persist(GeneralConfig.EnvRootPath, "commonPipelineEnvironment")
				telemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				telemetryData.ErrorCategory = log.GetErrorCategory().String()
				telemetry.Send(&telemetryData)
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetry.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			xsDeploy(stepConfig, &telemetryData, &commonPipelineEnvironment)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addXsDeployFlags(createXsDeployCmd, &stepConfig)
	return createXsDeployCmd
}

func addXsDeployFlags(cmd *cobra.Command, stepConfig *xsDeployOptions) {
	cmd.Flags().StringVar(&stepConfig.DeployOpts, "deployOpts", os.Getenv("PIPER_deployOpts"), "Additional options appended to the deploy command. Only needed for sophisticated cases. When provided it is the duty of the provider to ensure proper quoting / escaping.")
	cmd.Flags().StringVar(&stepConfig.OperationIDLogPattern, "operationIdLogPattern", `^.*xs bg-deploy -i (.*) -a.*$`, "Regex pattern for retrieving the ID of the operation from the xs log.")
	cmd.Flags().StringVar(&stepConfig.MtaPath, "mtaPath", os.Getenv("PIPER_mtaPath"), "Path to deployable")
	cmd.Flags().StringVar(&stepConfig.Action, "action", `NONE`, "Used for finalizing the blue-green deployment.")
	cmd.Flags().StringVar(&stepConfig.Mode, "mode", `DEPLOY`, "Controls if there is a standard deployment or a blue green deployment. Values: 'DEPLOY', 'BG_DEPLOY'")
	cmd.Flags().StringVar(&stepConfig.OperationID, "operationId", os.Getenv("PIPER_operationId"), "The operation ID. Used in case of bg-deploy in order to resume or abort a previously started deployment.")
	cmd.Flags().StringVar(&stepConfig.APIURL, "apiUrl", os.Getenv("PIPER_apiUrl"), "The api url (e.g. https://example.org:12345")
	cmd.Flags().StringVar(&stepConfig.Username, "username", os.Getenv("PIPER_username"), "Username")
	cmd.Flags().StringVar(&stepConfig.Password, "password", os.Getenv("PIPER_password"), "Password")
	cmd.Flags().StringVar(&stepConfig.Org, "org", os.Getenv("PIPER_org"), "The org")
	cmd.Flags().StringVar(&stepConfig.Space, "space", os.Getenv("PIPER_space"), "The space")
	cmd.Flags().StringVar(&stepConfig.LoginOpts, "loginOpts", os.Getenv("PIPER_loginOpts"), "Additional options appended to the login command. Only needed for sophisticated cases. When provided it is the duty of the provider to ensure proper quoting / escaping.")
	cmd.Flags().StringVar(&stepConfig.XsSessionFile, "xsSessionFile", os.Getenv("PIPER_xsSessionFile"), "The file keeping the xs session.")

	cmd.MarkFlagRequired("mtaPath")
	cmd.MarkFlagRequired("mode")
	cmd.MarkFlagRequired("apiUrl")
	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("org")
	cmd.MarkFlagRequired("space")
	cmd.MarkFlagRequired("loginOpts")
}

// retrieve step metadata
func xsDeployMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "xsDeploy",
			Aliases:     []config.Alias{},
			Description: "Performs xs deployment",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "credentialsId", Description: "Jenkins 'Username with password' credentials ID containing username/password for accessing xs endpoint.", Type: "jenkins"},
				},
				Parameters: []config.StepParameters{
					{
						Name:        "deployOpts",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Default:     os.Getenv("PIPER_deployOpts"),
						Aliases:     []config.Alias{},
					},
					{
						Name:        "operationIdLogPattern",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Default:     `^.*xs bg-deploy -i (.*) -a.*$`,
						Aliases:     []config.Alias{{Name: "deployIdLogPattern"}},
					},
					{
						Name: "mtaPath",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "mtaPath",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Default:   os.Getenv("PIPER_mtaPath"),
						Aliases:   []config.Alias{},
					},
					{
						Name:        "action",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Default:     `NONE`,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "mode",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Default:     `DEPLOY`,
						Aliases:     []config.Alias{},
					},
					{
						Name: "operationId",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "operationId",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Default:   os.Getenv("PIPER_operationId"),
						Aliases:   []config.Alias{},
					},
					{
						Name:        "apiUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Default:     os.Getenv("PIPER_apiUrl"),
						Aliases:     []config.Alias{},
					},
					{
						Name: "username",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "credentialsId",
								Param: "username",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Default:   os.Getenv("PIPER_username"),
						Aliases:   []config.Alias{{Name: "user"}},
					},
					{
						Name: "password",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "credentialsId",
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
						Name:        "org",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Default:     os.Getenv("PIPER_org"),
						Aliases:     []config.Alias{},
					},
					{
						Name:        "space",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Default:     os.Getenv("PIPER_space"),
						Aliases:     []config.Alias{},
					},
					{
						Name:        "loginOpts",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Default:     os.Getenv("PIPER_loginOpts"),
						Aliases:     []config.Alias{},
					},
					{
						Name:        "xsSessionFile",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Default:     os.Getenv("PIPER_xsSessionFile"),
						Aliases:     []config.Alias{},
					},
				},
			},
			Containers: []config.Container{
				{Name: "xs", Image: "ppiper/xs-cli"},
			},
			Outputs: config.StepOutputs{
				Resources: []config.StepResources{
					{
						Name: "commonPipelineEnvironment",
						Type: "piperEnvironment",
						Parameters: []map[string]interface{}{
							{"Name": "operationId"},
						},
					},
				},
			},
		},
	}
	return theMetaData
}
