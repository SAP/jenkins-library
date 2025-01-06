package cmd

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/config"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/spf13/cobra"
)

// WritePipelineEnv Serializes the commonPipelineEnvironment JSON to disk
func WritePipelineEnv() *cobra.Command {
	var stepConfig artifactPrepareVersionOptions
	var encryptedCPE bool
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
			err := runWritePipelineEnv(stepConfig.Password, encryptedCPE)
			if err != nil {
				log.Entry().Fatalf("error when writing common Pipeline environment: %v", err)
			}
		},
	}

	writePipelineEnv.Flags().BoolVar(&encryptedCPE, "encryptedCPE", false, "Bool to use encryption in CPE")
	return writePipelineEnv
}

func runWritePipelineEnv(stepConfigPassword string, encryptedCPE bool) error {
	var err error
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
	if encryptedCPE {
		log.Entry().Debug("trying to decrypt CPE")
		if stepConfigPassword == "" {
			return fmt.Errorf("empty stepConfigPassword")
		}

		inBytes, err = decrypt([]byte(stepConfigPassword), inBytes)
		if err != nil {
			log.Entry().Fatal(err)
		}
	}

	commonPipelineEnv := piperenv.CPEMap{}
	decoder := json.NewDecoder(bytes.NewReader(inBytes))
	decoder.UseNumber()
	err = decoder.Decode(&commonPipelineEnv)
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
