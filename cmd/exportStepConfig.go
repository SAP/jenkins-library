package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var exportStepConfigOptions struct {
	stepName         string
	stepMetadataFile string
	outputFile       string
}

var exportStepConfigFlags map[string]bool

// ExportStepConfigCommand is the entry command for exporting the step configuration
// It also handles target step specific flags the same way as if they are passed to the target step directly.
func ExportStepConfigCommand() *cobra.Command {
	var exportStepConfigCmd = &cobra.Command{
		DisableFlagParsing: true, // this step receives flags for other steps, so we need to disable flag parsing and do it during step execution manually
		Use:                "exportStepConfig",
		Short:              "For internal use by Piper team only. Do NOT use this command in production pipelines.",
		PreRun: func(cmd *cobra.Command, _ []string) {
			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)
			initStageName(false)
			log.SetVerbose(GeneralConfig.Verbose)
			GeneralConfig.GitHubAccessTokens = ResolveAccessTokens(GeneralConfig.GitHubTokens)
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Flags().Parse(args) // parse flags to populate exportStepConfigOptions
			metadata, err := config.ResolveMetadata(nil, nil, exportStepConfigOptions.stepMetadataFile, "")
			if err != nil {
				log.Entry().WithError(err).Fatal("Failed to resolve metadata")
				return
			}

			parsedFlags := parseFlagValues(cmd, args, metadata)

			result, err := prepareConfig(parsedFlags, &metadata, exportStepConfigOptions.stepName)
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

// parseFlagValues processes command-line arguments and maps them to their appropriate types based on step metadata.
// It filters out flags specific to the exportStepConfig step itself.
func parseFlagValues(cmd *cobra.Command, args []string, metadata config.StepData) map[string]any {
	parsedArgs := parseArgs(args) // by this point, parsedArgs contains all flags passed to exportStepConfig step execution

	result := make(map[string]any)
	passedFlags := config.AvailableFlagValues(cmd, &config.StepFilters{})
	for k, v := range passedFlags {
		delete(parsedArgs, k) // remove already processed flags
		if exportStepConfigFlags[k] {
			// drop exportStepConfig step specific flags
			continue
		}

		result[k] = v
	}

	// Process remaining (target step (e.g. sonarExecuteScan) specific) flags
	for paramName, value := range parsedArgs {
		paramMeta := getParamMetadataByName(metadata.Spec.Inputs.Parameters, paramName)
		if paramMeta == nil {
			log.Entry().Warningf("No metadata found for parameter '%s', skipping", paramName)
			continue
		}

		switch paramMeta.Type {
		case "string":
			result[paramName] = value
		case "[]string":
			result[paramName], _ = csv.NewReader(strings.NewReader(value)).Read()
		case "bool":
			if value == "true" {
				result[paramName] = true
			} else {
				result[paramName] = false
			}
		case "int":
			intVal, err := strconv.Atoi(value)
			if err != nil {
				log.Entry().Warningf("Failed to convert value '%s' to int for parameter '%s', skipping", value, paramName)
				continue
			}
			result[paramName] = intVal
		case "int64":
			intVal, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				log.Entry().Warningf("Failed to convert value '%s' to int64 for parameter '%s', skipping", value, paramName)
				continue
			}
			result[paramName] = intVal
		default:
			log.Entry().Warningf("Unknown parameter type '%s' for parameter '%s', skipping", paramMeta.Type, paramName)
			continue
		}
	}

	return result
}

// parseArgs returns args in form of a map[string]string
// with flag names as keys and their corresponding values.
func parseArgs(args []string) map[string]string {
	result := make(map[string]string)
	i := 0
	for i < len(args) {
		arg := args[i]
		if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-") {
			// Remove leading dashes
			key := strings.TrimLeft(arg, "-")

			// Handle --flag=value and -f=value
			if strings.Contains(key, "=") {
				parts := strings.SplitN(key, "=", 2)
				result[parts[0]] = parts[1]
			} else {
				// Check if next arg exists and is not a flag
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					result[key] = args[i+1]
					i++ // Skip value
				} else {
					// Boolean flag
					result[key] = "true"
				}
			}
		}
		i++
	}
	return result
}

func getParamMetadataByName(params []config.StepParameters, name string) *config.StepParameters {
	for _, param := range params {
		if param.Name == name {
			return &param
		}
	}
	return nil
}

// prepareConfig is minimized and adapted version of PrepareConfig() from cmd/piper.go
func prepareConfig(stepFlags map[string]any, metadata *config.StepData, stepName string) (map[string]any, error) {
	var customConfig io.ReadCloser
	{
		projectConfigFile := getProjectConfigFile(GeneralConfig.CustomConfig)
		if exists, err := piperutils.FileExists(projectConfigFile); exists {
			log.Entry().Debugf("Project config: '%s'", projectConfigFile)
			if customConfig, err = config.OpenPiperFile(projectConfigFile, GeneralConfig.GitHubAccessTokens); err != nil {
				return nil, fmt.Errorf("could not open project config file '%s': %w", projectConfigFile, err)
			}
		} else {
			log.Entry().Infof("Project config: NONE ('%s' does not exist)", projectConfigFile)
			customConfig = nil
		}
	}
	var defaultConfig []io.ReadCloser
	{
		if len(GeneralConfig.DefaultConfig) == 0 {
			log.Entry().Info("Project defaults: NONE")
		}
		for _, projectDefaultFile := range GeneralConfig.DefaultConfig {
			fc, err := config.OpenPiperFile(projectDefaultFile, GeneralConfig.GitHubAccessTokens)
			// only create error for non-default values
			if err != nil {
				if projectDefaultFile != ".pipeline/defaults.yaml" {
					log.Entry().Debugf("Project defaults: '%s'", projectDefaultFile)
					return nil, fmt.Errorf("%w: Cannot read '%s'", err, projectDefaultFile)
				}
			} else {
				log.Entry().Debugf("Project defaults: '%s'", projectDefaultFile)
				defaultConfig = append(defaultConfig, fc)
			}
		}
	}

	filters := metadata.GetParameterFilters()
	filters.All = append(filters.All, "collectTelemetryData")
	filters.General = append(filters.General, "collectTelemetryData")
	filters.Parameters = append(filters.Parameters, "collectTelemetryData")
	for param := range stepFlags {
		filters.Parameters = append(filters.Parameters, param)
	}

	envParams := metadata.GetResourceParameters(GeneralConfig.EnvRootPath, "commonPipelineEnvironment")
	reportingEnvParams := config.ReportingParameters.GetResourceParameters(GeneralConfig.EnvRootPath, "commonPipelineEnvironment")
	resourceParams := mergeResourceParameters(envParams, reportingEnvParams)

	var myConfig config.Config
	stepConfig, err := myConfig.GetStepConfig(stepFlags, "", customConfig, defaultConfig, GeneralConfig.IgnoreCustomDefaults, filters, *metadata, resourceParams, GeneralConfig.StageName, stepName)
	if err != nil {
		return nil, errors.Wrap(err, "retrieving step configuration failed")
	}

	return stepConfig.Config, nil
}

// flags for configuring the exportStepConfig step only.
// They are named with "export" prefix to avoid conflicts with flags of other steps
// which are passed along with exportStepConfig flags.
//
// For example:
//
//	exportStepConfig --exportStepName sonarExecuteScan --exportMetadataFile resources/metadata/sonarExecuteScan.yaml --exportOutputFilePath test.yml --projectKey my-project --verbose
//
// Here, --projectKey is not a flag for exportStepConfig but for sonarExecuteScan step.
// and --verbose is a general flag, that is applied to exportStepConfig as well as to sonarExecuteScan step.
func addExportStepConfigFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&exportStepConfigOptions.stepName, "exportStepName", "", "Name of the step being checked")
	cmd.Flags().StringVar(&exportStepConfigOptions.stepMetadataFile, "exportMetadataFile", "", "Step metadata, passed as path to yaml")
	cmd.Flags().StringVar(&exportStepConfigOptions.outputFile, "exportOutputFilePath", "", "Defines a file path. If set, the output will be written to the defined file")
	_ = cmd.MarkFlagRequired("exportStepName")
	_ = cmd.MarkFlagRequired("exportMetadataFile")
	_ = cmd.MarkFlagRequired("exportOutputFilePath")

	// This is used to filter out exportStepConfig specific flags from all other flags that should be
	// passed for building the step configuration.
	exportStepConfigFlags = map[string]bool{
		"exportStepName":       true,
		"exportMetadataFile":   true,
		"exportOutputFilePath": true,
	}
}
