package cmd

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/config"
	"io"
	"os"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/spf13/cobra"
)

type writePipelineEnvOptions struct {
	Secret string `json:"secret,omitempty"`
}

// WritePipelineEnv Serializes the commonPipelineEnvironment JSON to disk
func WritePipelineEnv() *cobra.Command {
	const STEP_NAME = "writePipelineEnv"
	var stepConfig writePipelineEnvOptions
	metadata := writePipelineEnvMetadata()

	return &cobra.Command{
		Use:   "writePipelineEnv",
		Short: "Serializes the commonPipelineEnvironment JSON to disk",
		PreRun: func(cmd *cobra.Command, args []string) {
			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

			err := PrepareConfig(cmd, &metadata, STEP_NAME, &stepConfig, config.OpenPiperFile)
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return
			}
			log.RegisterSecret(stepConfig.Secret)
		},

		Run: func(cmd *cobra.Command, args []string) {
			err := runWritePipelineEnv(&stepConfig)
			if err != nil {
				log.Entry().Fatalf("error when writing common Pipeline environment: %v", err)
			}
		},
	}
}

func runWritePipelineEnv(config *writePipelineEnvOptions) error {
	pipelineEnv, ok := os.LookupEnv("PIPER_pipelineEnv")
	inBytes := []byte(pipelineEnv)
	if !ok {
		var err error
		inBytes, err = io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
	}
	if len(inBytes) == 0 {
		return nil
	}

	// try to decrypt
	if config.Secret != "" && orchestrator.DetectOrchestrator() != orchestrator.Jenkins {
		log.Entry().Debug("found PIPER_pipelineEnv_SECRET, trying to decrypt CPE")
		var err error
		inBytes, err = decrypt([]byte(config.Secret), inBytes)
		if err != nil {
			log.Entry().Fatal(err)
		}
	}

	commonPipelineEnv := piperenv.CPEMap{}
	decoder := json.NewDecoder(bytes.NewReader(inBytes))
	decoder.UseNumber()
	err := decoder.Decode(&commonPipelineEnv)
	if err != nil {
		return err
	}

	rootPath := filepath.Join(GeneralConfig.EnvRootPath, "commonPipelineEnvironment")
	err = commonPipelineEnv.WriteToDisk(rootPath)
	if err != nil {
		return err
	}

	writtenBytes, err := json.MarshalIndent(commonPipelineEnv, "", "\t")
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(writtenBytes)
	if err != nil {
		return err
	}
	return nil
}

func decrypt(secret, base64CipherText []byte) ([]byte, error) {
	// decode from base64
	cipherText, err := b64.StdEncoding.DecodeString(string(base64CipherText))
	if err != nil {
		return nil, fmt.Errorf("failed to decode from base64: %v", err)
	}

	// use SHA256 as key
	key := sha256.Sum256(secret)
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create new cipher: %v", err)
	}

	if len(cipherText) < aes.BlockSize {
		return nil, fmt.Errorf("invalid ciphertext block size")
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return cipherText, nil
}

// retrieve step metadata
func writePipelineEnvMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name: "writePipelineEnv",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name: "secret",
						ResourceRef: []config.ResourceReference{
							{
								Name: "cpeSecret",
								Type: "vaultSecret",
							},
						},
						Type:      "string",
						Mandatory: false,
						Default:   os.Getenv("PIPER_pipelineEnv_SECRET"),
					},
				},
			},
		},
	}
	return theMetaData
}
