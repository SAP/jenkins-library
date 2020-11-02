// Code generated by piper's step-generator. DO NOT EDIT.

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

type mavenBuildOptions struct {
	PomPath                     string `json:"pomPath,omitempty"`
	Flatten                     bool   `json:"flatten,omitempty"`
	Verify                      bool   `json:"verify,omitempty"`
	ProjectSettingsFile         string `json:"projectSettingsFile,omitempty"`
	GlobalSettingsFile          string `json:"globalSettingsFile,omitempty"`
	M2Path                      string `json:"m2Path,omitempty"`
	LogSuccessfulMavenTransfers bool   `json:"logSuccessfulMavenTransfers,omitempty"`
}

// MavenBuildCommand This step will install the maven project into the local maven repository.
func MavenBuildCommand() *cobra.Command {
	const STEP_NAME = "mavenBuild"

	metadata := mavenBuildMetadata()
	var stepConfig mavenBuildOptions
	var startTime time.Time

	var createMavenBuildCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "This step will install the maven project into the local maven repository.",
		Long: `This step will install the maven project into the local maven repository.
It will also prepare jacoco to record the code coverage and
supports ci friendly versioning by flattening the pom before installing.`,
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
				telemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				telemetryData.ErrorCategory = log.GetErrorCategory().String()
				telemetry.Send(&telemetryData)
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetry.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			mavenBuild(stepConfig, &telemetryData)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addMavenBuildFlags(createMavenBuildCmd, &stepConfig)
	return createMavenBuildCmd
}

func addMavenBuildFlags(cmd *cobra.Command, stepConfig *mavenBuildOptions) {
	cmd.Flags().StringVar(&stepConfig.PomPath, "pomPath", `pom.xml`, "Path to the pom file which should be installed including all children.")
	cmd.Flags().BoolVar(&stepConfig.Flatten, "flatten", true, "Defines if the pom files should be flattened to support ci friendly maven versioning.")
	cmd.Flags().BoolVar(&stepConfig.Verify, "verify", false, "Instead of installing the artifact only the verify lifecycle phase is executed.")
	cmd.Flags().StringVar(&stepConfig.ProjectSettingsFile, "projectSettingsFile", os.Getenv("PIPER_projectSettingsFile"), "Path to the mvn settings file that should be used as project settings file.")
	cmd.Flags().StringVar(&stepConfig.GlobalSettingsFile, "globalSettingsFile", os.Getenv("PIPER_globalSettingsFile"), "Path to the mvn settings file that should be used as global settings file.")
	cmd.Flags().StringVar(&stepConfig.M2Path, "m2Path", os.Getenv("PIPER_m2Path"), "Path to the location of the local repository that should be used.")
	cmd.Flags().BoolVar(&stepConfig.LogSuccessfulMavenTransfers, "logSuccessfulMavenTransfers", false, "Configures maven to log successful downloads. This is set to `false` by default to reduce the noise in build logs.")

}

// retrieve step metadata
func mavenBuildMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:    "mavenBuild",
			Aliases: []config.Alias{{Name: "mavenExecute", Deprecated: false}},
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "pomPath",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "flatten",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "verify",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "projectSettingsFile",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "maven/projectSettingsFile"}},
					},
					{
						Name:        "globalSettingsFile",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "maven/globalSettingsFile"}},
					},
					{
						Name:        "m2Path",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "maven/m2Path"}},
					},
					{
						Name:        "logSuccessfulMavenTransfers",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "maven/logSuccessfulMavenTransfers"}},
					},
				},
			},
		},
	}
	return theMetaData
}
