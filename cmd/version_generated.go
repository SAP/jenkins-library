package cmd

import (
	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"

	"github.com/spf13/cobra"
)

type versionOptions struct {
}

var myVersionOptions versionOptions

// VersionCommand Returns the version of the piper binary
func VersionCommand() *cobra.Command {
	metadata := versionMetadata()

	var createVersionCmd = &cobra.Command{
		Use:   "version",
		Short: "Returns the version of the piper binary",
		Long:  `Writes the commit hash and the tag (if any) to stdout and exits with 0.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			log.SetStepName("version")
			log.SetVerbose(GeneralConfig.Verbose)
			return PrepareConfig(cmd, &metadata, "version", &myVersionOptions, config.OpenPiperFile)
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			telemetry.Initialize(GeneralConfig.NoTelemetry, "version")
			telemetry.Send(&telemetry.CustomData{})
			return version(myVersionOptions)
		},
	}

	addVersionFlags(createVersionCmd)
	return createVersionCmd
}

func addVersionFlags(cmd *cobra.Command) {

}

// retrieve step metadata
func versionMetadata() config.StepData {
	var theMetaData = config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{},
			},
		},
	}
	return theMetaData
}
