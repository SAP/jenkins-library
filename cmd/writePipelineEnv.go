package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/encryption"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/spf13/cobra"
)

// WritePipelineEnv Serializes the commonPipelineEnvironment JSON to disk
// Can be used in two modes:
// 1. JSON serialization: processes JSON input from stdin or PIPER_pipelineEnv environment variable
// 2. Direct value: writes a single key-value pair using the --value flag (format: key=value)
func WritePipelineEnv() *cobra.Command {
	var stepConfig artifactPrepareVersionOptions
	var encryptedCPE bool
	var directValue string
	metadata := artifactPrepareVersionMetadata()

	writePipelineEnv := &cobra.Command{
		Use:   "writePipelineEnv",
		Short: "Serializes the commonPipelineEnvironment JSON to disk",
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
			if directValue != "" {
				err := writeDirectValue(directValue)
				if err != nil {
					log.Entry().Fatalf("error when writing direct value: %v", err)
				}
				return
			}
			err := runWritePipelineEnv(stepConfig.Password, encryptedCPE)
			if err != nil {
				log.Entry().Fatalf("error when writing common Pipeline environment: %v", err)
			}
		},
	}

	writePipelineEnv.Flags().BoolVar(&encryptedCPE, "encryptedCPE", false, "Bool to use encryption in CPE")
	writePipelineEnv.Flags().StringVar(&directValue, "value", "", "Key-value pair to write directly (format: key=value)")
	return writePipelineEnv
}

func runWritePipelineEnv(stepConfigPassword string, encryptedCPE bool) error {
	inBytes, err := readInput()
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	if len(inBytes) == 0 {
		return nil
	}

	if encryptedCPE {
		if inBytes, err = handleEncryption(stepConfigPassword, inBytes); err != nil {
			return err
		}
	}

	commonPipelineEnv, err := parseInput(inBytes)
	if err != nil {
		return fmt.Errorf("failed to parse input: %w", err)
	}

	if _, err := writeOutput(commonPipelineEnv); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

func readInput() ([]byte, error) {
	if pipelineEnv, ok := os.LookupEnv("PIPER_pipelineEnv"); ok {
		return []byte(pipelineEnv), nil
	}
	return io.ReadAll(os.Stdin)
}

func handleEncryption(password string, data []byte) ([]byte, error) {
	if password == "" {
		return nil, fmt.Errorf("encryption enabled but password is empty")
	}
	log.Entry().Debug("decrypting CPE data")
	return encryption.Decrypt([]byte(password), data)
}

func parseInput(data []byte) (piperenv.CPEMap, error) {
	commonPipelineEnv := piperenv.CPEMap{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&commonPipelineEnv); err != nil {
		return nil, err
	}
	return commonPipelineEnv, nil
}

func writeOutput(commonPipelineEnv piperenv.CPEMap) (int, error) {
	rootPath := filepath.Join(GeneralConfig.EnvRootPath, "commonPipelineEnvironment")
	if err := commonPipelineEnv.WriteToDisk(rootPath); err != nil {
		return 0, err
	}

	writtenBytes, err := json.MarshalIndent(commonPipelineEnv, "", "\t")
	if err != nil {
		return 0, err
	}
	return os.Stdout.Write(writtenBytes)
}

// writeDirectValue writes a single value to a file in the commonPipelineEnvironment directory
// The key-value pair should be in the format "key=value"
// The key will be used as the file name and the value as its content
func writeDirectValue(keyValue string) error {
	parts := strings.SplitN(keyValue, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid key-value format. Expected 'key=value', got '%s'", keyValue)
	}

	key := parts[0]
	value := parts[1]

	rootPath := filepath.Join(GeneralConfig.EnvRootPath, "commonPipelineEnvironment")
	filePath := filepath.Join(rootPath, key)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	return os.WriteFile(filePath, []byte(value), 0644)
}
