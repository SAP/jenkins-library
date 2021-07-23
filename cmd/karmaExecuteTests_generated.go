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
	"github.com/spf13/cobra"
)

type karmaExecuteTestsOptions struct {
	InstallCommand string   `json:"installCommand,omitempty"`
	Modules        []string `json:"modules,omitempty"`
	RunCommand     string   `json:"runCommand,omitempty"`
}

// KarmaExecuteTestsCommand Executes the Karma test runner
func KarmaExecuteTestsCommand() *cobra.Command {
	const STEP_NAME = "karmaExecuteTests"

	metadata := karmaExecuteTestsMetadata()
	var stepConfig karmaExecuteTestsOptions
	var startTime time.Time
	var logCollector *log.CollectorHook

	var createKarmaExecuteTestsCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Executes the Karma test runner",
		Long: `In this step the ([Karma test runner](http://karma-runner.github.io)) is executed.

The step is using the ` + "`" + `seleniumExecuteTest` + "`" + ` step to spin up two containers in a Docker network:

* a Selenium/Chrome container (` + "`" + `selenium/standalone-chrome` + "`" + `)
* a NodeJS container (` + "`" + `node:lts-stretch` + "`" + `)

In the Docker network, the containers can be referenced by the values provided in ` + "`" + `dockerName` + "`" + ` and ` + "`" + `sidecarName` + "`" + `, the default values are ` + "`" + `karma` + "`" + ` and ` + "`" + `selenium` + "`" + `. These values must be used in the ` + "`" + `hostname` + "`" + ` properties of the test configuration ([Karma](https://karma-runner.github.io/1.0/config/configuration-file.html) and [WebDriver](https://github.com/karma-runner/karma-webdriver-launcher#usage)).

!!! note
    In a Kubernetes environment, the containers both need to be referenced with ` + "`" + `localhost` + "`" + `.`,
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
			karmaExecuteTests(stepConfig, &telemetryData)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addKarmaExecuteTestsFlags(createKarmaExecuteTestsCmd, &stepConfig)
	return createKarmaExecuteTestsCmd
}

func addKarmaExecuteTestsFlags(cmd *cobra.Command, stepConfig *karmaExecuteTestsOptions) {
	cmd.Flags().StringVar(&stepConfig.InstallCommand, "installCommand", `npm install --quiet`, "The command that is executed to install the test tool.")
	cmd.Flags().StringSliceVar(&stepConfig.Modules, "modules", []string{`.`}, "Define the paths of the modules to execute tests on.")
	cmd.Flags().StringVar(&stepConfig.RunCommand, "runCommand", `npm run karma`, "The command that is executed to start the tests.")

	cmd.MarkFlagRequired("installCommand")
	cmd.MarkFlagRequired("modules")
	cmd.MarkFlagRequired("runCommand")
}

// retrieve step metadata
func karmaExecuteTestsMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "karmaExecuteTests",
			Aliases:     []config.Alias{},
			Description: "Executes the Karma test runner",
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
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     `npm install --quiet`,
					},
					{
						Name:        "modules",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     []string{`.`},
					},
					{
						Name:        "runCommand",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     `npm run karma`,
					},
				},
			},
			Containers: []config.Container{
				{Name: "karma", Image: "node:lts-stretch", EnvVars: []config.EnvVar{{Name: "no_proxy", Value: "localhost,selenium,$no_proxy"}, {Name: "NO_PROXY", Value: "localhost,selenium,$NO_PROXY"}, {Name: "PIPER_SELENIUM_HOSTNAME", Value: "karma"}, {Name: "PIPER_SELENIUM_WEBDRIVER_HOSTNAME", Value: "selenium"}, {Name: "PIPER_SELENIUM_WEBDRIVER_PORT", Value: "4444"}}, WorkingDir: "/home/node"},
			},
			Sidecars: []config.Container{
				{Name: "selenium", Image: "selenium/standalone-chrome", EnvVars: []config.EnvVar{{Name: "NO_PROXY", Value: "localhost,karma,$NO_PROXY"}, {Name: "no_proxy", Value: "localhost,selenium,$no_proxy"}}},
			},
		},
	}
	return theMetaData
}
