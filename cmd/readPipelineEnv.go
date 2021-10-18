package cmd

import (
	"encoding/json"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/spf13/cobra"
	"os"
	"path"
)

// ReadPipelineEnv reads the commonPipelineEnvironment from disk and outputs it as JSON
func ReadPipelineEnv() *cobra.Command {
	return &cobra.Command{
		Use:   "readPipelineEnv",
		Short: "Reads the commonPipelineEnvironment from disk and outputs it as JSON",
		PreRun: func(cmd *cobra.Command, args []string) {
			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)
		},

		Run: func(cmd *cobra.Command, args []string) {
			err := runReadPipelineEnv()
			if err != nil {
				log.Entry().Fatalf("error when writing reading Pipeline environment: %v", err)
			}
		},
	}
}

func runReadPipelineEnv() error {
	cpe := piperenv.CPEMap{}

	err := cpe.LoadFromDisk(path.Join(GeneralConfig.EnvRootPath, "commonPipelineEnvironment"))
	if err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(cpe, "", "\t")
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}
