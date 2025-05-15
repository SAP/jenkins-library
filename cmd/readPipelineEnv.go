package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/encryption"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/spf13/cobra"
)

// ReadPipelineEnv reads the commonPipelineEnvironment from disk and outputs it as JSON
func ReadPipelineEnv() *cobra.Command {
	var stepConfig artifactPrepareVersionOptions
	var encryptedCPE bool
	metadata := artifactPrepareVersionMetadata()

	readPipelineEnvCmd := &cobra.Command{
		Use:   "readPipelineEnv",
		Short: "Reads the commonPipelineEnvironment from disk and outputs it as JSON",
		PreRun: func(cmd *cobra.Command, args []string) {
			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

			err := PrepareConfig(cmd, &metadata, "", &stepConfig, config.OpenPiperFile)
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return
			}
			log.RegisterSecret(stepConfig.Password)
			log.RegisterSecret(stepConfig.Username)
		},

		Run: func(cmd *cobra.Command, args []string) {
			err := runReadPipelineEnv(stepConfig.Password, encryptedCPE)
			if err != nil {
				log.Entry().Fatalf("error when writing reading Pipeline environment: %v", err)
			}
		},
	}

	readPipelineEnvCmd.Flags().BoolVar(&encryptedCPE, "encryptedCPE", false, "Bool to use encryption in CPE")
	return readPipelineEnvCmd
}

func runReadPipelineEnv(stepConfigPassword string, encryptedCPE bool) error {
	cpe := piperenv.CPEMap{}

	err := cpe.LoadFromDisk(path.Join(GeneralConfig.EnvRootPath, "commonPipelineEnvironment"))
	if err != nil {
		return err
	}

	// try to encrypt
	if encryptedCPE {
		log.Entry().Debug("trying to encrypt CPE")
		if stepConfigPassword == "" {
			return fmt.Errorf("empty stepConfigPassword")
		}

		cpeJsonBytes, _ := json.Marshal(cpe)
		encryptedCPEBytes, err := encryption.Encrypt([]byte(stepConfigPassword), cpeJsonBytes)
		if err != nil {
			log.Entry().Fatal(err)
		}

		os.Stdout.Write(encryptedCPEBytes)
		return nil
	}

	// fallback
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "\t")
	if err := encoder.Encode(cpe); err != nil {
		return err
	}

	return nil
}
