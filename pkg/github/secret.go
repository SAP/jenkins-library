package github

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/google/go-github/v68/github"
	"golang.org/x/crypto/nacl/box"

	"github.com/SAP/jenkins-library/pkg/log"
)

// CreateEncryptedSecret creates an encrypted secret using a public key from a GitHub repository, which can be sent through the GitHub API
// https://github.com/google/go-github/blob/master/example/newreposecretwithxcrypto/main.go
func CreateEncryptedSecret(secretName, secretValue string, publicKey *github.PublicKey) (*github.EncryptedSecret, error) {
	decodedPublicKey, err := base64.StdEncoding.DecodeString(publicKey.GetKey())
	if err != nil {
		log.Entry().Warn("Could not decode public key from base64")
		return nil, err
	}

	var boxKey [32]byte
	copy(boxKey[:], decodedPublicKey)
	secretBytes := []byte(secretValue)
	encryptedSecretBytes, err := box.SealAnonymous([]byte{}, secretBytes, &boxKey, rand.Reader)
	if err != nil {
		log.Entry().Warn("Could not encrypt secret using public key")
		return nil, err
	}

	encryptedSecretString := base64.StdEncoding.EncodeToString(encryptedSecretBytes)

	githubSecret := &github.EncryptedSecret{
		Name:           secretName,
		KeyID:          publicKey.GetKeyID(),
		EncryptedValue: encryptedSecretString,
	}
	return githubSecret, nil
}
