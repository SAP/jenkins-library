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
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/spf13/cobra"
)

type uiVeri5ExecuteTestsOptions struct {
	InstallCommand string   `json:"installCommand,omitempty" validate:""`
	RunCommand     string   `json:"runCommand,omitempty" validate:""`
	RunOptions     []string `json:"runOptions,omitempty" validate:""`
	TestOptions    string   `json:"testOptions,omitempty" validate:""`
	TestServerURL  string   `json:"testServerUrl,omitempty" validate:""`
}

// UiVeri5ExecuteTestsCommand Executes UI5 e2e tests using uiVeri5
func UiVeri5ExecuteTestsCommand() *cobra.Command {
	const STEP_NAME = "uiVeri5ExecuteTests"

	metadata := uiVeri5ExecuteTestsMetadata()
	var stepConfig uiVeri5ExecuteTestsOptions
	var startTime time.Time
	var logCollector *log.CollectorHook

	var createUiVeri5ExecuteTestsCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Executes UI5 e2e tests using uiVeri5",
		Long:  `In this step the ([UIVeri5 tests](https://github.com/SAP/ui5-uiveri5)) are executed.`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			log.SetVerbose(GeneralConfig.Verbose)

			validation, err := validation.New()
			if err != nil {
				return err
			}
			if err := validation.ValidateStruct(stepConfig); err != nil {
				return err
			}

			GeneralConfig.GitHubAccessTokens = ResolveAccessTokens(GeneralConfig.GitHubTokens)

			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

			err = PrepareConfig(cmd, &metadata, STEP_NAME, &stepConfig, config.OpenPiperFile)
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}

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
			uiVeri5ExecuteTests(stepConfig, &telemetryData)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addUiVeri5ExecuteTestsFlags(createUiVeri5ExecuteTestsCmd, &stepConfig)
	return createUiVeri5ExecuteTestsCmd
}

func addUiVeri5ExecuteTestsFlags(cmd *cobra.Command, stepConfig *uiVeri5ExecuteTestsOptions) {
	cmd.Flags().StringVar(&stepConfig.InstallCommand, "installCommand", `npm install @ui5/uiveri5 --global --quiet`, "The command that is executed to install the uiveri5 test tool.")
	cmd.Flags().StringVar(&stepConfig.RunCommand, "runCommand", `/home/node/.npm-global/bin/uiveri5`, "The command that is executed to start the tests.")
	cmd.Flags().StringSliceVar(&stepConfig.RunOptions, "runOptions", []string{`--seleniumAddress=http://localhost:4444/wd/hub`}, "Options to append to the runCommand, last parameter has to be path to conf.js (default if missing: ./conf.js).")
	cmd.Flags().StringVar(&stepConfig.TestOptions, "testOptions", os.Getenv("PIPER_testOptions"), "Deprecated and will result in an error if set. Please use runOptions instead. Split the testOptions string at the whitespaces when migrating it into a list of runOptions.")
	cmd.Flags().StringVar(&stepConfig.TestServerURL, "testServerUrl", os.Getenv("PIPER_testServerUrl"), "URL pointing to the deployment.")

	cmd.MarkFlagRequired("installCommand")
	cmd.MarkFlagRequired("runCommand")
	cmd.MarkFlagRequired("runOptions")
}

// retrieve step metadata
func uiVeri5ExecuteTestsMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "uiVeri5ExecuteTests",
			Aliases:     []config.Alias{},
			Description: "Executes UI5 e2e tests using uiVeri5",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Resources: []config.StepResources{
					{Name: "buildDescriptor", Type: "stash"},
					{Name: "tests", Type: "stash"},
				},
				Parameters: []config.StepParameters{
					{
						Name:        "installCommand",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     `npm install @ui5/uiveri5 --global --quiet`,
					},
					{
						Name:        "runCommand",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     `/home/node/.npm-global/bin/uiveri5`,
					},
					{
						Name:        "runOptions",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     []string{`--seleniumAddress=http://localhost:4444/wd/hub`},
					},
					{
						Name:        "testOptions",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_testOptions"),
					},
					{
						Name:        "testServerUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_testServerUrl"),
					},
				},
			},
			Containers: []config.Container{
				{Name: "uiVeri5", Image: "node:lts-stretch", EnvVars: []config.EnvVar{{Name: "no_proxy", Value: "localhost,selenium,$no_proxy"}, {Name: "NO_PROXY", Value: "localhost,selenium,$NO_PROXY"}}, WorkingDir: "/home/node"},
			},
			Sidecars: []config.Container{
				{Name: "selenium", Image: "selenium/standalone-chrome", EnvVars: []config.EnvVar{{Name: "NO_PROXY", Value: "localhost,selenium,$NO_PROXY"}, {Name: "no_proxy", Value: "localhost,selenium,$no_proxy"}}},
			},
		},
	}
	return theMetaData
}
