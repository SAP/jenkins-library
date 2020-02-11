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

type hadolintExecuteOptions struct {
	ConfigurationURL  string `json:"configurationUrl,omitempty"`
	DockerFile        string `json:"dockerFile,omitempty"`
	ConfigurationFile string `json:"configurationFile,omitempty"`
	QualityGates      string `json:"qualityGates,omitempty"`
	ReportFile        string `json:"reportFile,omitempty"`
}

// HadolintExecuteCommand Executes the Haskell Dockerfile Linter which is a smarter Dockerfile linter that helps you build [best practice](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/) Docker images.
func HadolintExecuteCommand() *cobra.Command {
	metadata := hadolintExecuteMetadata()
	var stepConfig hadolintExecuteOptions
	var startTime time.Time

	var createHadolintExecuteCmd = &cobra.Command{
		Use:   "hadolintExecute",
		Short: "Executes the Haskell Dockerfile Linter which is a smarter Dockerfile linter that helps you build [best practice](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/) Docker images.",
		Long: `Executes the Haskell Dockerfile Linter which is a smarter Dockerfile linter that helps you build [best practice](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/) Docker images.
The linter is parsing the Dockerfile into an abstract syntax tree (AST) and performs rules on top of the AST.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			startTime = time.Now()
			log.SetStepName("hadolintExecute")
			log.SetVerbose(GeneralConfig.Verbose)
			return PrepareConfig(cmd, &metadata, "hadolintExecute", &stepConfig, config.OpenPiperFile)
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
			telemetry.Initialize(GeneralConfig.NoTelemetry, "hadolintExecute")
			hadolintExecute(stepConfig, &telemetryData)
			telemetryData.ErrorCode = "0"
		},
	}

	addHadolintExecuteFlags(createHadolintExecuteCmd, &stepConfig)
	return createHadolintExecuteCmd
}

func addHadolintExecuteFlags(cmd *cobra.Command, stepConfig *hadolintExecuteOptions) {
	cmd.Flags().StringVar(&stepConfig.ConfigurationURL, "configurationUrl", os.Getenv("PIPER_configurationUrl"), "URL pointing to the .hadolint.yaml exclude configuration to be used for linting. Also have a look at `configurationFile` which could avoid central configuration download in case the file is part of your repository.")
	cmd.Flags().StringVar(&stepConfig.DockerFile, "dockerFile", "./Dockerfile", "Dockerfile to be used for the assessment.")
	cmd.Flags().StringVar(&stepConfig.ConfigurationFile, "configurationFile", ".hadolint.yaml", "Name of the configuration file used locally within the step. If a file with this name is detected as part of your repo downloading the central configuration via `configurationUrl` will be skipped. If you change the file's name make sure your stashing configuration also reflects this.")
	cmd.Flags().StringVar(&stepConfig.QualityGates, "qualityGates", os.Getenv("PIPER_qualityGates"), "Quality Gates to fail the build, see [warnings-ng plugin documentation](https://github.com/jenkinsci/warnings-plugin/blob/master/doc/Documentation.md#quality-gate-configuration).")
	cmd.Flags().StringVar(&stepConfig.ReportFile, "reportFile", "hadolint.xml", "Name of the result file used locally within the step.")

}

// retrieve step metadata
func hadolintExecuteMetadata() config.StepData {
	var theMetaData = config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "configurationUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "dockerFile",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "configurationFile",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "qualityGates",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "reportFile",
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
