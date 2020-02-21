package cmd

import (
	"encoding/hex"
	"encoding/json"
	"hash"
	"io"
	"net/http"
	"os"
	"strings"

	"crypto/md5"
	"crypto/sha1"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type artifactDescription struct {
	ID         string `json:"artifactId"`
	Classifier string `json:"classifier"`
	Type       string `json:"type"`
	File       string `json:"file"`
}

func nexusUpload(config nexusUploadOptions, telemetryData *telemetry.CustomData) {
	artifacts := getArtifacts(config)
	baseUrl := getBaseUrl(config)
	client := createHttpClient(config)

	for _, artifact := range artifacts {
		url := getArtifactUrl(baseUrl, config.Version, artifact)

		uploadHash(client, artifact.File, url+".md5", md5.New(), 16)
		uploadHash(client, artifact.File, url+".sha1", sha1.New(), 20)
		uploadFile(client, artifact.File, url)
	}
}

func getArtifacts(config nexusUploadOptions) []artifactDescription {
	var artifacts []artifactDescription
	err := json.Unmarshal([]byte(config.Artifacts), &artifacts)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to convert artifact JSON '", config.Artifacts, "'")
	}
	return artifacts
}

func createHttpClient(config nexusUploadOptions) *piperHttp.Client {
	client := piperHttp.Client{}
	clientOptions := piperHttp.ClientOptions{Username: config.User, Password: config.Password, Logger: log.Entry().WithField("package", "github.com/SAP/jenkins-library/pkg/http")}
	client.SetOptions(clientOptions)
	return &client
}

func getBaseUrl(config nexusUploadOptions) string {
	baseUrl := config.Url
	switch config.NexusVersion {
	case "nexus2":
		baseUrl += "/content/repositories/"
	case "nexus3":
		baseUrl += "/repository/"
	default:
		log.Entry().Fatal("Unsupported Nexus version '", config.NexusVersion, "'")
	}
	groupPath := strings.ReplaceAll(config.GroupID, ".", "/")
	baseUrl += config.Repository + "/" + groupPath + "/"
	return baseUrl
}

func getArtifactUrl(baseUrl, version string, artifact artifactDescription) string {
	url := baseUrl

	// Generate artifacte name including optional classifier
	artifactName := artifact.ID + "-" + version
	if len(artifact.Classifier) > 0 {
		artifactName += "-" + artifact.Classifier
	}
	artifactName += "." + artifact.Type

	url += artifact.ID + "/" + version + "/" + artifactName

	// Remove any double slashes, as Nexus does not like them, and prepend protocol
	url = "http://" + strings.ReplaceAll(url, "//", "/")

	return url
}

func uploadFile(client *piperHttp.Client, filePath, url string) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to open artifact file ", filePath)
	}

	defer file.Close()

	err = uploadToNexus(client, file, url)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to upload artifact ", filePath)
	}
}

func uploadHash(client *piperHttp.Client, filePath, url string, hash hash.Hash, length int) {
	hashReader, err := generateHashReader(filePath, hash, length)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to generate hash")
	}
	err = uploadToNexus(client, hashReader, url)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to upload hash")
	}
}

func uploadToNexus(client *piperHttp.Client, stream io.Reader, url string) error {
	response, err := client.SendRequest(http.MethodPut, url, stream, nil, nil)
	if err == nil {
		log.Entry().Info("Uploaded '"+url+"', response: ", response.StatusCode)
	}
	return err
}

func generateHashReader(filePath string, hash hash.Hash, length int) (io.Reader, error) {
	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	// Read file and feed the hash
	if _, err := io.Copy(hash, file); err != nil {
		return nil, err
	}

	// Get the requested number of bytes from the hash
	hashInBytes := hash.Sum(nil)[:length]

	// Convert the bytes to a string
	hexString := hex.EncodeToString(hashInBytes)

	// Finally create an io.Reader wrapping the string
	return strings.NewReader(hexString), nil
}
