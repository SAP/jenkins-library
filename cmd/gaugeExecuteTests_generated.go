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
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/spf13/cobra"
)

type gaugeExecuteTestsOptions struct {
	InstallCommand string `json:"installCommand,omitempty"`
	LanguageRunner string `json:"languageRunner,omitempty"`
	RunCommand     string `json:"runCommand,omitempty"`
	TestOptions    string `json:"testOptions,omitempty"`
}

type gaugeExecuteTestsInflux struct {
	step_data struct {
		fields struct {
			gauge bool
		}
		tags struct {
		}
	}
}

func (i *gaugeExecuteTestsInflux) persist(path, resourceName string) {
	measurementContent := []struct {
		measurement string
		valType     string
		name        string
		value       interface{}
	}{
		{valType: config.InfluxField, measurement: "step_data", name: "gauge", value: i.step_data.fields.gauge},
	}

	errCount := 0
	for _, metric := range measurementContent {
		err := piperenv.SetResourceParameter(path, resourceName, filepath.Join(metric.measurement, fmt.Sprintf("%vs", metric.valType), metric.name), metric.value)
		if err != nil {
			log.Entry().WithError(err).Error("Error persisting influx environment.")
			errCount++
		}
	}
	if errCount > 0 {
		log.Entry().Error("failed to persist Influx environment")
	}
}

// GaugeExecuteTestsCommand Installs gauge and executes specified gauge tests.
func GaugeExecuteTestsCommand() *cobra.Command {
	const STEP_NAME = "gaugeExecuteTests"

	metadata := gaugeExecuteTestsMetadata()
	var stepConfig gaugeExecuteTestsOptions
	var startTime time.Time
	var influx gaugeExecuteTestsInflux
	var logCollector *log.CollectorHook

	var createGaugeExecuteTestsCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Installs gauge and executes specified gauge tests.",
		Long: `In this step Gauge ([getgauge.io](https://getgauge.io)) acceptance tests are executed. Using Gauge it will be possible to have a three-tier test layout:

Acceptance Criteria
Test implemenation layer
Application driver layer

This layout is propagated by Jez Humble and Dave Farley in their book "Continuous Delivery" as a way to create maintainable acceptance test suites (see "Continuous Delivery", p. 190ff).

Using Gauge it is possible to write test specifications in [Markdown syntax](http://daringfireball.net/projects/markdown/syntax) and therefore allow e.g. product owners to write the relevant acceptance test specifications. At the same time it allows the developer to implement the steps described in the specification in her development environment.

You can use the [sample projects](https://github.com/getgauge/gauge-mvn-archetypes) of Gauge.`,
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
			telemetryData := telemetry.CustomData{}
			telemetryData.ErrorCode = "1"
			handler := func() {
				config.RemoveVaultSecretFiles()
				influx.persist(GeneralConfig.EnvRootPath, "influx")
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
			gaugeExecuteTests(stepConfig, &telemetryData, &influx)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addGaugeExecuteTestsFlags(createGaugeExecuteTestsCmd, &stepConfig)
	return createGaugeExecuteTestsCmd
}

func addGaugeExecuteTestsFlags(cmd *cobra.Command, stepConfig *gaugeExecuteTestsOptions) {
	cmd.Flags().StringVar(&stepConfig.InstallCommand, "installCommand", os.Getenv("PIPER_installCommand"), "Defines the command for installing Gauge. Gauge should be installed using npm. Example: npm install -g @getgauge/cli@1.2.1")
	cmd.Flags().StringVar(&stepConfig.LanguageRunner, "languageRunner", os.Getenv("PIPER_languageRunner"), "Defines the Gauge language runner to be used. Example: java")
	cmd.Flags().StringVar(&stepConfig.RunCommand, "runCommand", os.Getenv("PIPER_runCommand"), "Defines the command which is used for executing Gauge. Example: run -s -p specs/")
	cmd.Flags().StringVar(&stepConfig.TestOptions, "testOptions", os.Getenv("PIPER_testOptions"), "Allows to set specific options for the Gauge execution. Details can be found for example [in the Gauge Maven plugin documentation](https://github.com/getgauge-contrib/gauge-maven-plugin#executing-specs)")

	cmd.MarkFlagRequired("runCommand")
}

// retrieve step metadata
func gaugeExecuteTestsMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "gaugeExecuteTests",
			Aliases:     []config.Alias{},
			Description: "Installs gauge and executes specified gauge tests.",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "seleniumHubCredentialsId", Description: "Defines the id of the user/password credentials to be used to connect to a Selenium Hub. The credentials are provided in the environment variables `PIPER_SELENIUM_HUB_USER` and `PIPER_SELENIUM_HUB_PASSWORD`.", Type: "jenkins"},
				},
				Resources: []config.StepResources{
					{Name: "buildDescriptor", Type: "stash"},
					{Name: "tests", Type: "stash"},
				},
				Parameters: []config.StepParameters{
					{
						Name:        "installCommand",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_installCommand"),
					},
					{
						Name:        "languageRunner",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_languageRunner"),
					},
					{
						Name:        "runCommand",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_runCommand"),
					},
					{
						Name:        "testOptions",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_testOptions"),
					},
				},
			},
			Containers: []config.Container{
				{Name: "gauge", Image: "node:lts-stretch", EnvVars: []config.EnvVar{{Name: "no_proxy", Value: "localhost,selenium,$no_proxy"}, {Name: "NO_PROXY", Value: "localhost,selenium,$NO_PROXY"}}, WorkingDir: "/home/node"},
			},
			Sidecars: []config.Container{
				{Name: "selenium", Image: "selenium/standalone-chrome", EnvVars: []config.EnvVar{{Name: "NO_PROXY", Value: "localhost,selenium,$NO_PROXY"}, {Name: "no_proxy", Value: "localhost,selenium,$no_proxy"}}},
			},
			Outputs: config.StepOutputs{
				Resources: []config.StepResources{
					{
						Name: "influx",
						Type: "influx",
						Parameters: []map[string]interface{}{
							{"Name": "step_data"}, {"fields": []map[string]string{{"name": "gauge"}}},
						},
					},
				},
			},
		},
	}
	return theMetaData
}
