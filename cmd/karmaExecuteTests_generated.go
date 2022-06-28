// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/gcs"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/bmatcuk/doublestar"
	"github.com/spf13/cobra"
)

type karmaExecuteTestsOptions struct {
	InstallCommand string   `json:"installCommand,omitempty"`
	Modules        []string `json:"modules,omitempty"`
	RunCommand     string   `json:"runCommand,omitempty"`
}

type karmaExecuteTestsReports struct {
}

func (p *karmaExecuteTestsReports) persist(stepConfig karmaExecuteTestsOptions, gcpJsonKeyFilePath string, gcsBucketId string, gcsFolderPath string, gcsSubFolder string) {
	if gcsBucketId == "" {
		log.Entry().Info("persisting reports to GCS is disabled, because gcsBucketId is empty")
		return
	}
	log.Entry().Info("Uploading reports to Google Cloud Storage...")
	content := []gcs.ReportOutputParam{
		{FilePattern: "**/TEST-*.xml", ParamRef: "", StepResultType: "karma"},
		{FilePattern: "**/cobertura-coverage.xml", ParamRef: "", StepResultType: "karma"},
		{FilePattern: "**/TEST-*.xml", ParamRef: "", StepResultType: "junit"},
		{FilePattern: "**/jacoco.xml", ParamRef: "", StepResultType: "jacoco-coverage"},
		{FilePattern: "**/cobertura-coverage.xml", ParamRef: "", StepResultType: "cobertura-coverage"},
		{FilePattern: "**/xmake_stage.json", ParamRef: "", StepResultType: "xmake"},
		{FilePattern: "**/requirement.mapping", ParamRef: "", StepResultType: "requirement-mapping"},
	}
	envVars := []gcs.EnvVar{
		{Name: "GOOGLE_APPLICATION_CREDENTIALS", Value: gcpJsonKeyFilePath, Modified: false},
	}
	gcsClient, err := gcs.NewClient(gcs.WithEnvVars(envVars))
	if err != nil {
		log.Entry().Errorf("creation of GCS client failed: %v", err)
		return
	}
	defer gcsClient.Close()
	structVal := reflect.ValueOf(&stepConfig).Elem()
	inputParameters := map[string]string{}
	for i := 0; i < structVal.NumField(); i++ {
		field := structVal.Type().Field(i)
		if field.Type.String() == "string" {
			paramName := strings.Split(field.Tag.Get("json"), ",")
			paramValue, _ := structVal.Field(i).Interface().(string)
			inputParameters[paramName[0]] = paramValue
		}
	}
	if err := gcs.PersistReportsToGCS(gcsClient, content, inputParameters, gcsFolderPath, gcsBucketId, gcsSubFolder, doublestar.Glob, os.Stat); err != nil {
		log.Entry().Errorf("failed to persist reports: %v", err)
	}
}

// KarmaExecuteTestsCommand Executes the Karma test runner
func KarmaExecuteTestsCommand() *cobra.Command {
	const STEP_NAME = "karmaExecuteTests"

	metadata := karmaExecuteTestsMetadata()
	var stepConfig karmaExecuteTestsOptions
	var startTime time.Time
	var reports karmaExecuteTestsReports
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

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
				splunkClient = &splunk.Splunk{}
				logCollector = &log.CollectorHook{CorrelationID: GeneralConfig.CorrelationID}
				log.RegisterHook(logCollector)
			}

			if err = log.RegisterANSHookIfConfigured(GeneralConfig.CorrelationID); err != nil {
				log.Entry().WithError(err).Warn("failed to set up SAP Alert Notification Service log hook")
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
				reports.persist(stepConfig, GeneralConfig.GCPJsonKeyFilePath, GeneralConfig.GCSBucketId, GeneralConfig.GCSFolderPath, GeneralConfig.GCSSubFolder)
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
			karmaExecuteTests(stepConfig, &stepTelemetryData)
			stepTelemetryData.ErrorCode = "0"
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
			Outputs: config.StepOutputs{
				Resources: []config.StepResources{
					{
						Name: "reports",
						Type: "reports",
						Parameters: []map[string]interface{}{
							{"filePattern": "**/TEST-*.xml", "type": "karma"},
							{"filePattern": "**/cobertura-coverage.xml", "type": "karma"},
							{"filePattern": "**/TEST-*.xml", "type": "junit"},
							{"filePattern": "**/jacoco.xml", "type": "jacoco-coverage"},
							{"filePattern": "**/cobertura-coverage.xml", "type": "cobertura-coverage"},
							{"filePattern": "**/xmake_stage.json", "type": "xmake"},
							{"filePattern": "**/requirement.mapping", "type": "requirement-mapping"},
						},
					},
				},
			},
		},
	}
	return theMetaData
}
