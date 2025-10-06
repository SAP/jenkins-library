package cmd

import (
	"os"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
)

var exportStepConfigOptions struct {
	stepName     string
	stepMetadata string
	outputFile   string
}

// ExportStepConfigCommand is the entry command for exporting the step configuration
func ExportStepConfigCommand() *cobra.Command {
	var exportStepConfigCmd = &cobra.Command{
		Use:   "exportStepConfig",
		Short: "For internal use by Piper team only. Do NOT use this command in production pipelines.",
		PreRun: func(cmd *cobra.Command, _ []string) {
			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)
			initStageName(false)
			log.SetVerbose(GeneralConfig.Verbose)
			GeneralConfig.GitHubAccessTokens = ResolveAccessTokens(GeneralConfig.GitHubTokens)
		},
		Run: func(cmd *cobra.Command, _ []string) {
			metadata, err := config.ResolveMetadata(GeneralConfig.GitHubAccessTokens, GeneralConfig.MetaDataResolver, exportStepConfigOptions.stepMetadata, "")
			if err != nil {
				log.Entry().WithError(err).Fatal("Failed to resolve metadata")
				return
			}

			var result map[string]interface{}
			err = PrepareConfig(cmd, &metadata, exportStepConfigOptions.stepName, nil, config.OpenPiperFile, &result)
			if err != nil {
				log.Entry().WithError(err).Fatal("Failed to prepare configuration")
				return
			}

			yamlData, err := yaml.Marshal(result)
			if err != nil {
				log.Entry().WithError(err).Fatal("Failed to marshal result to YAML")
				return
			}
			err = os.WriteFile(exportStepConfigOptions.outputFile, yamlData, 0644)
			if err != nil {
				log.Entry().WithError(err).Fatal("Failed to write YAML to file")
				return
			}
			log.Entry().Infof("Configuration exported to %s\n", exportStepConfigOptions.outputFile)
		},
	}
	addExportStepConfigFlags(exportStepConfigCmd)
	return exportStepConfigCmd
}

func addExportStepConfigFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&exportStepConfigOptions.stepName, "stepName", "", "Name of the step being checked")
	cmd.Flags().StringVar(&exportStepConfigOptions.stepMetadata, "metadataFile", "", "Step metadata, passed as path to yaml")
	cmd.Flags().StringVar(&exportStepConfigOptions.outputFile, "outputFilePath", "", "Defines a file path. If set, the output will be written to the defined file")
	_ = cmd.MarkFlagRequired("stepName")
	_ = cmd.MarkFlagRequired("metadataFile")
	_ = cmd.MarkFlagRequired("outputFilePath")
}
