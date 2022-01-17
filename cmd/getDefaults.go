package cmd

import (
	"io"
	"os"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/spf13/cobra"
)

type defaultsCommandOptions struct {
	output        string //output format of default configs, currently only YAML
	outputFile    string //if set: path to file where the output should be written to
	defaultsFiles []string
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
	log.Entry().Info(defaultsOptions)
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
			err := generateDefaults(utils)
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				log.Entry().WithError(err).Fatal("failed to retrieve configuration")
			}
		},
	}

	addDefaultsFlags(createDefaultsCmd)
	return createDefaultsCmd
}

func addDefaultsFlags(cmd *cobra.Command) {

	cmd.Flags().StringVar(&defaultsOptions.output, "output", "yaml", "Defines the format of the configs embedded into a JSON object")
	cmd.Flags().StringVar(&defaultsOptions.outputFile, "outputFile", "", "Defines the output filename")
	cmd.Flags().StringArrayVar(&defaultsOptions.defaultsFiles, "defaultsFiles", []string{}, "Defines the input defaults files")

}
