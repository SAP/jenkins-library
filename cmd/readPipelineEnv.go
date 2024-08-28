package cmd

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/SAP/jenkins-library/pkg/config"
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
		encryptedCPEBytes, err := encrypt([]byte(stepConfigPassword), cpeJsonBytes)
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

func encrypt(secret, inBytes []byte) ([]byte, error) {
	// use SHA256 as key
	key := sha256.Sum256(secret)
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create new cipher: %v", err)
	}

	// Make the cipher text a byte array of size BlockSize + the length of the message
	cipherText := make([]byte, aes.BlockSize+len(inBytes))

	// iv is the ciphertext up to the blocksize (16)
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("failed to init iv: %v", err)
	}

	// Encrypt the data:
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], inBytes)

	// Return string encoded in base64
	return []byte(base64.StdEncoding.EncodeToString(cipherText)), err
}
