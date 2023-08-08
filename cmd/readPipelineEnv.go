package cmd

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/spf13/cobra"
	"io"
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

	// try to encrypt
	if secret, ok := os.LookupEnv("PIPER_pipelineEnv_SECRET"); ok && secret != "" {
		log.Entry().Debug("found PIPER_pipelineEnv_SECRET, trying to encrypt CPE")
		jsonBytes, _ := json.Marshal(cpe)
		encrypted, err := encrypt([]byte(secret), jsonBytes)
		if err != nil {
			log.Entry().Fatal(err)
		}

		os.Stdout.Write(encrypted)
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
	return []byte(base64.RawStdEncoding.EncodeToString(cipherText)), err
}
