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

type mavenExecuteIntegrationOptions struct {
	Retry                       int    `json:"retry,omitempty"`
	ForkCount                   string `json:"forkCount,omitempty"`
	Goal                        string `json:"goal,omitempty"`
	InstallArtifacts            bool   `json:"installArtifacts,omitempty"`
	ProjectSettingsFile         string `json:"projectSettingsFile,omitempty"`
	GlobalSettingsFile          string `json:"globalSettingsFile,omitempty"`
	M2Path                      string `json:"m2Path,omitempty"`
	LogSuccessfulMavenTransfers bool   `json:"logSuccessfulMavenTransfers,omitempty"`
}

// MavenExecuteIntegrationCommand This step will execute backend integration tests via the Jacoco Maven-plugin.
func MavenExecuteIntegrationCommand() *cobra.Command {
	const STEP_NAME = "mavenExecuteIntegration"

	metadata := mavenExecuteIntegrationMetadata()
	var stepConfig mavenExecuteIntegrationOptions
	var startTime time.Time
	var logCollector *log.CollectorHook

	var createMavenExecuteIntegrationCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "This step will execute backend integration tests via the Jacoco Maven-plugin.",
		Long: `If the project contains a Maven module named "integration-tests", this step will execute
the integration tests via the Jacoco Maven-plugin.`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			log.SetVerbose(GeneralConfig.Verbose)

			GeneralConfig.GitHubAccessTokens = ResolveAccessTokens(GeneralConfig.GitHubTokens)

			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

			validation, err := validation.New()
			if err != nil {
				return err
			}
			if err := validation.ValidateStruct(stepConfig); err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}

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
			mavenExecuteIntegration(stepConfig, &telemetryData)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addMavenExecuteIntegrationFlags(createMavenExecuteIntegrationCmd, &stepConfig)
	return createMavenExecuteIntegrationCmd
}

func addMavenExecuteIntegrationFlags(cmd *cobra.Command, stepConfig *mavenExecuteIntegrationOptions) {
	cmd.Flags().IntVar(&stepConfig.Retry, "retry", 1, "The number of times that integration tests will be retried before failing the step. Note: This will consume more time for the step execution.")
	cmd.Flags().StringVar(&stepConfig.ForkCount, "forkCount", `1C`, "The number of JVM processes that are spawned to run the tests in parallel in case of using a maven based project structure. For more details visit the Surefire documentation at https://maven.apache.org/surefire/maven-surefire-plugin/test-mojo.html#forkCount.")
	cmd.Flags().StringVar(&stepConfig.Goal, "goal", `test`, "The name of the Maven goal to execute.")
	cmd.Flags().BoolVar(&stepConfig.InstallArtifacts, "installArtifacts", true, "If enabled, it will install all artifacts to the local maven repository to make them available before running the tests. This is required if the integration test module has dependencies to other modules in the repository and they were not installed before.")
	cmd.Flags().StringVar(&stepConfig.ProjectSettingsFile, "projectSettingsFile", os.Getenv("PIPER_projectSettingsFile"), "Path to the mvn settings file that should be used as project settings file.")
	cmd.Flags().StringVar(&stepConfig.GlobalSettingsFile, "globalSettingsFile", os.Getenv("PIPER_globalSettingsFile"), "Path to the mvn settings file that should be used as global settings file.")
	cmd.Flags().StringVar(&stepConfig.M2Path, "m2Path", os.Getenv("PIPER_m2Path"), "Path to the location of the local repository that should be used.")
	cmd.Flags().BoolVar(&stepConfig.LogSuccessfulMavenTransfers, "logSuccessfulMavenTransfers", false, "Configures maven to log successful downloads. This is set to `false` by default to reduce the noise in build logs.")

}

// retrieve step metadata
func mavenExecuteIntegrationMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "mavenExecuteIntegration",
			Aliases:     []config.Alias{{Name: "mavenExecute", Deprecated: false}},
			Description: "This step will execute backend integration tests via the Jacoco Maven-plugin.",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "retry",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STEPS", "STAGES"},
						Type:        "int",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     1,
					},
					{
						Name:        "forkCount",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STEPS", "STAGES"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `1C`,
					},
					{
						Name:        "goal",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `test`,
					},
					{
						Name:        "installArtifacts",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     true,
					},
					{
						Name:        "projectSettingsFile",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "maven/projectSettingsFile"}},
						Default:     os.Getenv("PIPER_projectSettingsFile"),
					},
					{
						Name:        "globalSettingsFile",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "maven/globalSettingsFile"}},
						Default:     os.Getenv("PIPER_globalSettingsFile"),
					},
					{
						Name:        "m2Path",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "maven/m2Path"}},
						Default:     os.Getenv("PIPER_m2Path"),
					},
					{
						Name:        "logSuccessfulMavenTransfers",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "maven/logSuccessfulMavenTransfers"}},
						Default:     false,
					},
				},
			},
			Containers: []config.Container{
				{Name: "mvn", Image: "maven:3.6-jdk-8"},
			},
			Sidecars: []config.Container{
				{},
			},
		},
	}
	return theMetaData
}
