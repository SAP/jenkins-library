package nexus

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
)

type ArtifactDescription struct {
	ID         string `json:"artifactId"`
	Classifier string `json:"classifier"`
	Type       string `json:"type"`
	File       string `json:"file"`
}

type NexusUpload struct {
	BaseUrl   string
	Version   string
	Username  string
	Password  string
	Artifacts []ArtifactDescription
}

func (nexusUpload *NexusUpload) UploadArtifacts() {
	client := nexusUpload.createHttpClient()

	for _, artifact := range nexusUpload.Artifacts {
		url := getArtifactUrl(nexusUpload.BaseUrl, nexusUpload.Version, artifact)

		uploadHash(client, artifact.File, url+".md5", md5.New(), 16)
		uploadHash(client, artifact.File, url+".sha1", sha1.New(), 20)
		uploadFile(client, artifact.File, url)
	}
}

func (nexusUpload *NexusUpload) SetArtifacts(json string) {
	nexusUpload.Artifacts = GetArtifacts(json)
}

func GetArtifacts(artifactsAsJSON string) []ArtifactDescription {
	var artifacts []ArtifactDescription
	err := json.Unmarshal([]byte(artifactsAsJSON), &artifacts)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to convert artifact JSON '", artifactsAsJSON, "'")
	}
	return artifacts
}

func (nexusUpload *NexusUpload) createHttpClient() *piperHttp.Client {
	client := piperHttp.Client{}
	clientOptions := piperHttp.ClientOptions{Username: nexusUpload.Username, Password: nexusUpload.Password, Logger: log.Entry().WithField("package", "github.com/SAP/jenkins-library/pkg/http")}
	client.SetOptions(clientOptions)
	return &client
}

func GetBaseUrl(nexusUrl, nexusVersion, repository, groupID string) string {
	baseUrl := nexusUrl
	switch nexusVersion {
	case "nexus2":
		baseUrl += "/content/repositories/"
	case "nexus3":
		baseUrl += "/repository/"
	default:
		log.Entry().Fatal("Unsupported Nexus version '", nexusVersion, "'")
	}
	groupPath := strings.ReplaceAll(groupID, ".", "/")
	baseUrl += repository + "/" + groupPath + "/"
	return baseUrl
}

func getArtifactUrl(baseUrl, version string, artifact ArtifactDescription) string {
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
