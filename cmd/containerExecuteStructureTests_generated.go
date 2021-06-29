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

type containerExecuteStructureTestsOptions struct {
	PullImage          bool   `json:"pullImage,omitempty"`
	TestConfiguration  string `json:"testConfiguration,omitempty"`
	TestDriver         string `json:"testDriver,omitempty"`
	TestImage          string `json:"testImage,omitempty"`
	TestReportFilePath string `json:"testReportFilePath,omitempty"`
}

// ContainerExecuteStructureTestsCommand In this step [Container Structure Tests](https://github.com/GoogleContainerTools/container-structure-test) are executed.
func ContainerExecuteStructureTestsCommand() *cobra.Command {
	const STEP_NAME = "containerExecuteStructureTests"

	metadata := containerExecuteStructureTestsMetadata()
	var stepConfig containerExecuteStructureTestsOptions
	var startTime time.Time
	var logCollector *log.CollectorHook

	var createContainerExecuteStructureTestsCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "In this step [Container Structure Tests](https://github.com/GoogleContainerTools/container-structure-test) are executed.",
		Long: `This testing framework allows you to execute different test types against a Docker container, for example:
- Command tests (only if a Docker Deamon is available)
- File existence tests
- File content tests
- Metadata test`,
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
			containerExecuteStructureTests(stepConfig, &telemetryData)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addContainerExecuteStructureTestsFlags(createContainerExecuteStructureTestsCmd, &stepConfig)
	return createContainerExecuteStructureTestsCmd
}

func addContainerExecuteStructureTestsFlags(cmd *cobra.Command, stepConfig *containerExecuteStructureTestsOptions) {
	cmd.Flags().BoolVar(&stepConfig.PullImage, "pullImage", false, "Force a pull of the tested image before running tests. Only relevant for testDriver 'docker'.")
	cmd.Flags().StringVar(&stepConfig.TestConfiguration, "testConfiguration", os.Getenv("PIPER_testConfiguration"), "Container structure test configuration in yml or json format. You can pass a pattern in order to execute multiple tests.")
	cmd.Flags().StringVar(&stepConfig.TestDriver, "testDriver", os.Getenv("PIPER_testDriver"), "Container structure test driver to be used for testing, please see https://github.com/GoogleContainerTools/container-structure-test for details.")
	cmd.Flags().StringVar(&stepConfig.TestImage, "testImage", os.Getenv("PIPER_testImage"), "Image to be tested")
	cmd.Flags().StringVar(&stepConfig.TestReportFilePath, "testReportFilePath", `cst-report.json`, "Path and name of the test report which will be generated")

	cmd.MarkFlagRequired("testConfiguration")
	cmd.MarkFlagRequired("testImage")
}

// retrieve step metadata
func containerExecuteStructureTestsMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "containerExecuteStructureTests",
			Aliases:     []config.Alias{},
			Description: "In this step [Container Structure Tests](https://github.com/GoogleContainerTools/container-structure-test) are executed.",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "pullImage",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "testConfiguration",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_testConfiguration"),
					},
					{
						Name:        "testDriver",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_testDriver"),
					},
					{
						Name:        "testImage",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_testImage"),
					},
					{
						Name:        "testReportFilePath",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `cst-report.json`,
					},
				},
			},
			Containers: []config.Container{
				{Image: "ppiper/container-structure-test", Options: []config.Option{{Name: "-u", Value: "0"}, {Name: "--entrypoint", Value: "''"}}},
			},
		},
	}
	return theMetaData
}
