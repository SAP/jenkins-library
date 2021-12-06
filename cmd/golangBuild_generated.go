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

type golangBuildOptions struct {
	BuildFlags                   []string `json:"buildFlags,omitempty"`
	CgoEnabled                   bool     `json:"cgoEnabled,omitempty"`
	CoverageFormat               string   `json:"coverageFormat,omitempty" validate:"possible-values=cobertura html"`
	CreateBOM                    bool     `json:"createBOM,omitempty"`
	CustomTLSCertificateLinks    []string `json:"customTlsCertificateLinks,omitempty"`
	ExcludeGeneratedFromCoverage bool     `json:"excludeGeneratedFromCoverage,omitempty"`
	LdflagsTemplate              string   `json:"ldflagsTemplate,omitempty"`
	Output                       string   `json:"output,omitempty"`
	Packages                     []string `json:"packages,omitempty"`
	Publish                      bool     `json:"publish,omitempty"`
	ReportCoverage               bool     `json:"reportCoverage,omitempty"`
	RunTests                     bool     `json:"runTests,omitempty"`
	RunIntegrationTests          bool     `json:"runIntegrationTests,omitempty"`
	TargetArchitectures          []string `json:"targetArchitectures,omitempty"`
	TestOptions                  []string `json:"testOptions,omitempty"`
	TestResultFormat             string   `json:"testResultFormat,omitempty" validate:"possible-values=junit standard"`
}

// GolangBuildCommand This step will execute a golang build.
func GolangBuildCommand() *cobra.Command {
	const STEP_NAME = "golangBuild"

	metadata := golangBuildMetadata()
	var stepConfig golangBuildOptions
	var startTime time.Time
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createGolangBuildCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "This step will execute a golang build.",
		Long: `This step will build a golang project.
It will also execute golang-based tests using [gotestsum](https://github.com/gotestyourself/gotestsum) and with that allows for reporting test results and test coverage.

Besides execution of the default tests the step allows for running an additional integration test run using ` + "`" + `-tags=integration` + "`" + ` using pattern ` + "`" + `./...` + "`" + `

If the build is successful the resulting artifact can be uploaded to e.g. a binary repository automatically.`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			log.SetVerbose(GeneralConfig.Verbose)

			GeneralConfig.GitHubAccessTokens = ResolveAccessTokens(GeneralConfig.GitHubTokens)

			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

			err := PrepareConfig(cmd, &metadata, STEP_NAME, &stepConfig, config.OpenPiperFile)
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}

			if len(GeneralConfig.HookConfig.SentryConfig.Dsn) > 0 {
				sentryHook := log.NewSentryHook(GeneralConfig.HookConfig.SentryConfig.Dsn, GeneralConfig.CorrelationID)
				log.RegisterHook(&sentryHook)
			}

			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				splunkClient = &splunk.Splunk{}
				logCollector = &log.CollectorHook{CorrelationID: GeneralConfig.CorrelationID}
				log.RegisterHook(logCollector)
			}

			validation, err := validation.New(validation.WithJSONNamesForStructFields(), validation.WithPredefinedErrorMessages())
			if err != nil {
				return err
			}
			if err = validation.ValidateStruct(stepConfig); err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}

			return nil
		},
		Run: func(_ *cobra.Command, _ []string) {
			stepTelemetryData := telemetry.CustomData{}
			stepTelemetryData.ErrorCode = "1"
			handler := func() {
				config.RemoveVaultSecretFiles()
				stepTelemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				stepTelemetryData.ErrorCategory = log.GetErrorCategory().String()
				stepTelemetryData.PiperCommitHash = GitCommit
				telemetryClient.SetData(&stepTelemetryData)
				telemetryClient.Send()
				if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
					splunkClient.Send(telemetryClient.GetData(), logCollector)
				}
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetryClient.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				splunkClient.Initialize(GeneralConfig.CorrelationID,
					GeneralConfig.HookConfig.SplunkConfig.Dsn,
					GeneralConfig.HookConfig.SplunkConfig.Token,
					GeneralConfig.HookConfig.SplunkConfig.Index,
					GeneralConfig.HookConfig.SplunkConfig.SendLogs)
			}
			golangBuild(stepConfig, &stepTelemetryData)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addGolangBuildFlags(createGolangBuildCmd, &stepConfig)
	return createGolangBuildCmd
}

func addGolangBuildFlags(cmd *cobra.Command, stepConfig *golangBuildOptions) {
	cmd.Flags().StringSliceVar(&stepConfig.BuildFlags, "buildFlags", []string{}, "Defines list of build flags to be used.")
	cmd.Flags().BoolVar(&stepConfig.CgoEnabled, "cgoEnabled", false, "If active: enables the creation of Go packages that call C code.")
	cmd.Flags().StringVar(&stepConfig.CoverageFormat, "coverageFormat", `html`, "Defines the format of the coverage repository.")
	cmd.Flags().BoolVar(&stepConfig.CreateBOM, "createBOM", false, "Creates the bill of materials (BOM) using CycloneDX plugin.")
	cmd.Flags().StringSliceVar(&stepConfig.CustomTLSCertificateLinks, "customTlsCertificateLinks", []string{}, "List of download links to custom TLS certificates. This is required to ensure trusted connections to instances with repositories (like nexus) when publish flag is set to true.")
	cmd.Flags().BoolVar(&stepConfig.ExcludeGeneratedFromCoverage, "excludeGeneratedFromCoverage", true, "Defines if generated files should be excluded, according to [https://golang.org/s/generatedcode](https://golang.org/s/generatedcode).")
	cmd.Flags().StringVar(&stepConfig.LdflagsTemplate, "ldflagsTemplate", os.Getenv("PIPER_ldflagsTemplate"), "Defines the content of -ldflags option in a golang template format.")
	cmd.Flags().StringVar(&stepConfig.Output, "output", os.Getenv("PIPER_output"), "Defines the build result or output directory as per `go build` documentation.")
	cmd.Flags().StringSliceVar(&stepConfig.Packages, "packages", []string{}, "List of packages to be build as per `go build` documentation.")
	cmd.Flags().BoolVar(&stepConfig.Publish, "publish", false, "Configures the build to publish artifacts to a repository.")
	cmd.Flags().BoolVar(&stepConfig.ReportCoverage, "reportCoverage", true, "Defines if a coverage report should be created.")
	cmd.Flags().BoolVar(&stepConfig.RunTests, "runTests", true, "Activates execution of tests using [gotestsum](https://github.com/gotestyourself/gotestsum).")
	cmd.Flags().BoolVar(&stepConfig.RunIntegrationTests, "runIntegrationTests", false, "Activates execution of a second test run using tag `integration`.")
	cmd.Flags().StringSliceVar(&stepConfig.TargetArchitectures, "targetArchitectures", []string{`linux,amd64`}, "Defines the target architectures for which the build should run using OS and architecture separated by a comma.")
	cmd.Flags().StringSliceVar(&stepConfig.TestOptions, "testOptions", []string{}, "Options to pass to test as per `go test` documentation (comprises e.g. flags, packages).")
	cmd.Flags().StringVar(&stepConfig.TestResultFormat, "testResultFormat", `junit`, "Defines the output format of the test results.")

	cmd.MarkFlagRequired("targetArchitectures")
}

// retrieve step metadata
func golangBuildMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "golangBuild",
			Aliases:     []config.Alias{},
			Description: "This step will execute a golang build.",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "buildFlags",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name:        "cgoEnabled",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "coverageFormat",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `html`,
					},
					{
						Name:        "createBOM",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "customTlsCertificateLinks",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name:        "excludeGeneratedFromCoverage",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     true,
					},
					{
						Name:        "ldflagsTemplate",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_ldflagsTemplate"),
					},
					{
						Name:        "output",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_output"),
					},
					{
						Name:        "packages",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name:        "publish",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "reportCoverage",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     true,
					},
					{
						Name:        "runTests",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     true,
					},
					{
						Name:        "runIntegrationTests",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "targetArchitectures",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "[]string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     []string{`linux,amd64`},
					},
					{
						Name:        "testOptions",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name:        "testResultFormat",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `junit`,
					},
				},
			},
			Containers: []config.Container{
				{Name: "golang", Image: "golang:1", Options: []config.Option{{Name: "-u", Value: "0"}}},
			},
		},
	}
	return theMetaData
}
