package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

type defaultsCommandOptions struct {
	output        string //output format of default configs, currently only YAML
	outputFile    string //if set: path to file where the output should be written to
	defaultsFiles []string
	useV1         bool
	openFile      func(s string, t map[string]string) (io.ReadCloser, error)
}

var defaultsOptions defaultsCommandOptions

type getDefaultsUtils interface {
	FileExists(filename string) (bool, error)
	DirExists(path string) (bool, error)
	FileWrite(path string, content []byte, perm os.FileMode) error
}

type getDefaultsUtilsBundle struct {
	*piperutils.Files
}

func newGetDefaultsUtilsUtils() getDefaultsUtils {
	utils := getDefaultsUtilsBundle{
		Files: &piperutils.Files{},
	}
	return &utils
}

// DefaultsCommand is the entry command for loading the configuration of a pipeline step
func DefaultsCommand() *cobra.Command {

	defaultsOptions.openFile = config.OpenPiperFile
	var createDefaultsCmd = &cobra.Command{
		Use:   "getDefaults",
		Short: "Retrieves multiple default configurations and outputs them embedded into a JSON object.",
		PreRun: func(cmd *cobra.Command, args []string) {
			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)
			initStageName(false)
			GeneralConfig.GitHubAccessTokens = ResolveAccessTokens(GeneralConfig.GitHubTokens)
		},
		Run: func(cmd *cobra.Command, _ []string) {
			utils := newGetDefaultsUtilsUtils()
			_, err := generateDefaults(utils)
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				log.Entry().WithError(err).Fatal("failed to retrieve default configurations")
			}
		},
	}

	addDefaultsFlags(createDefaultsCmd)
	return createDefaultsCmd
}

func getDefaults() ([]map[string]string, error) {

	var yamlDefaults []map[string]string

	for _, f := range defaultsOptions.defaultsFiles {
		fc, err := defaultsOptions.openFile(f, GeneralConfig.GitHubAccessTokens)
		if err != nil {
			return yamlDefaults, fmt.Errorf("defaults: retrieving defaults file failed: '%v': %w", f, err)
		}
		if err == nil {
			var yamlContent string

			if !defaultsOptions.useV1 {
				var c config.Config
				c.ReadConfig(fc)

				yamlContent, err = config.GetYAML(c)
				if err != nil {
					return yamlDefaults, fmt.Errorf("defaults: could not marshal YAML default file: '%v: %w", f, err)
				}
			} else {
				var rc config.RunConfigV1
				rc.StageConfigFile = fc
				rc.LoadConditionsV1()

				yamlContent, err = config.GetYAML(rc.PipelineConfig)
				if err != nil {
					return yamlDefaults, fmt.Errorf("defaults: could not marshal YAML default file: '%v: %w", f, err)
				}
			}

			yamlDefaults = append(yamlDefaults, map[string]string{"content": yamlContent, "filepath": f})
		}
	}

	return yamlDefaults, nil
}

func generateDefaults(utils getDefaultsUtils) ([]byte, error) {

	var jsonOutput []byte

	yamlDefaults, err := getDefaults()
	if err != nil {
		return jsonOutput, err
	}

	if len(yamlDefaults) > 1 {
		jsonOutput, err = json.Marshal(yamlDefaults)
	} else {
		jsonOutput, err = json.Marshal(yamlDefaults[0])
	}

	if err != nil {
		return jsonOutput, fmt.Errorf("defaults: could not embed YAML defaults into JSON: %w", err)
	}

	if len(defaultsOptions.outputFile) > 0 {
		err := utils.FileWrite(defaultsOptions.outputFile, []byte(jsonOutput), 0666)
		if err != nil {
			return jsonOutput, fmt.Errorf("failed to write output file %v: %w", defaultsOptions.outputFile, err)
		}
		return jsonOutput, nil
	}
	fmt.Println(string(jsonOutput))

	return jsonOutput, nil
}

func addDefaultsFlags(cmd *cobra.Command) {

	cmd.Flags().StringVar(&defaultsOptions.output, "output", "yaml", "Defines the format of the configs embedded into a JSON object")
	cmd.Flags().StringVar(&defaultsOptions.outputFile, "outputFile", "", "Defines the output filename")
	cmd.Flags().StringArrayVar(&defaultsOptions.defaultsFiles, "defaultsFile", []string{}, "Defines the input defaults file(s)")
	cmd.Flags().BoolVar(&defaultsOptions.useV1, "useV1", false, "Input files are CRD-style stage configuration")
	cmd.MarkFlagRequired("defaultsFile")
}
