package nexus

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/sirupsen/logrus"
)

// ArtifactDescription describes a single artifact that can be uploaded to a Nexus repository manager.
// The File string must point to an existing file. The Classifier can be empty.
type ArtifactDescription struct {
	ID         string `json:"artifactId"`
	Classifier string `json:"classifier"`
	Type       string `json:"type"`
	File       string `json:"file"`
}

// Upload holds state for an upload session. Call SetBaseURL(), SetArtifactsVersion() and add at least
// one artifact via AddArtifact(). Then call UploadArtifacts().
type Upload struct {
	baseURL   string
	version   string
	Username  string
	Password  string
	artifacts []ArtifactDescription
	Logger    *logrus.Entry
	Timeout   time.Duration
}

// Uploader provides an interface to the nexus upload for configuring the target Nexus Repository and
// adding artifacts.
type Uploader interface {
	SetBaseURL(nexusURL, nexusVersion, repository, groupID string) error
	SetArtifactsVersion(version string) error
	AddArtifact(artifact ArtifactDescription) error
	GetArtifacts() []ArtifactDescription
	UploadArtifacts() error
}

func (nexusUpload *Upload) initLogger() {
	if nexusUpload.Logger == nil {
		nexusUpload.Logger = log.Entry().WithField("package", "SAP/jenkins-library/pkg/nexus")
	}
}

// SetBaseURL constructs the base URL for all uploaded artifacts. No parameter can be empty.
func (nexusUpload *Upload) SetBaseURL(nexusURL, nexusVersion, repository, groupID string) error {
	baseURL, err := getBaseURL(nexusURL, nexusVersion, repository, groupID)
	if err != nil {
		return err
	}
	nexusUpload.baseURL = baseURL
	return nil
}

// SetArtifactsVersion sets the common version for all uploaded artifacts. The version is external to
// the artifact descriptions so that it is consistent for all of them.
func (nexusUpload *Upload) SetArtifactsVersion(version string) error {
	if version == "" {
		return errors.New("version must not be empty")
	}
	nexusUpload.version = version
	return nil
}

// AddArtifact adds a single artifact to be uploaded later via UploadArtifacts(). If an identical artifact
// description is already contained in the Upload, the function does nothing and returns no error.
func (nexusUpload *Upload) AddArtifact(artifact ArtifactDescription) error {
	err := validateArtifact(artifact)
	if err != nil {
		return err
	}
	if nexusUpload.containsArtifact(artifact) {
		log.Entry().Infof("Nexus Upload already contains artifact %v\n", artifact)
		return nil
	}
	nexusUpload.artifacts = append(nexusUpload.artifacts, artifact)
	return nil
}

func validateArtifact(artifact ArtifactDescription) error {
	if artifact.File == "" || artifact.ID == "" || artifact.Type == "" {
		return fmt.Errorf("Artifact.File (%v), ID (%v) or Type (%v) is empty",
			artifact.File, artifact.ID, artifact.Type)
	}
	if strings.Contains(artifact.ID, "/") {
		return fmt.Errorf("Artifact.ID may not include slashes")
	}
	return nil
}

func (nexusUpload *Upload) containsArtifact(artifact ArtifactDescription) bool {
	for _, n := range nexusUpload.artifacts {
		if artifact == n {
			return true
		}
	}
	return false
}

// GetArtifacts returns a copy of the artifact descriptions array stored in the Upload.
func (nexusUpload *Upload) GetArtifacts() []ArtifactDescription {
	artifacts := make([]ArtifactDescription, len(nexusUpload.artifacts))
	copy(artifacts, nexusUpload.artifacts)
	return artifacts
}

// UploadArtifacts performs the actual upload of all added artifacts to the Nexus server.
func (nexusUpload *Upload) UploadArtifacts() error {
	client := nexusUpload.createHTTPClient()
	return nexusUpload.uploadArtifacts(client)
}

func (nexusUpload *Upload) uploadArtifacts(client piperHttp.Sender) error {
	if nexusUpload.baseURL == "" {
		return fmt.Errorf("the nexus.Upload needs to be configured by calling SetBaseURL() first")
	}
	if nexusUpload.version == "" {
		return fmt.Errorf("the nexus.Upload needs to be configured by calling SetArtifactsVersion() first")
	}
	if len(nexusUpload.artifacts) == 0 {
		return fmt.Errorf("no artifacts to upload, call AddArtifact() first")
	}

	nexusUpload.initLogger()

	for _, artifact := range nexusUpload.artifacts {
		url := getArtifactURL(nexusUpload.baseURL, nexusUpload.version, artifact)

		var err error
		err = uploadHash(client, artifact.File, url+".md5", md5.New(), 16)
		if err != nil {
			return err
		}
		err = uploadHash(client, artifact.File, url+".sha1", sha1.New(), 20)
		if err != nil {
			return err
		}
		err = uploadFile(client, artifact.File, url)
		if err != nil {
			return err
		}
	}

	// Reset all artifacts already uploaded, so the object could be re-used
	nexusUpload.artifacts = nil
	return nil
}

func (nexusUpload *Upload) createHTTPClient() *piperHttp.Client {
	client := piperHttp.Client{}
	clientOptions := piperHttp.ClientOptions{
		Username: nexusUpload.Username,
		Password: nexusUpload.Password,
		Logger:   nexusUpload.Logger,
		Timeout:  nexusUpload.Timeout,
	}
	client.SetOptions(clientOptions)
	return &client
}

func getBaseURL(nexusURL, nexusVersion, repository, groupID string) (string, error) {
	if nexusURL == "" {
		return "", errors.New("nexusURL must not be empty")
	}
	nexusURL = strings.ToLower(nexusURL)
	if strings.HasPrefix(nexusURL, "http://") || strings.HasPrefix(nexusURL, "https://") {
		return "", errors.New("nexusURL must not start with 'http://' or 'https://'")
	}
	if repository == "" {
		return "", errors.New("repository must not be empty")
	}
	if groupID == "" {
		return "", errors.New("groupID must not be empty")
	}
	baseURL := nexusURL
	switch nexusVersion {
	case "nexus2":
		baseURL += "/content/repositories/"
	case "nexus3":
		baseURL += "/repository/"
	default:
		return "", fmt.Errorf("unsupported Nexus version '%s', must be 'nexus2' or 'nexus3'", nexusVersion)
	}
	groupPath := strings.ReplaceAll(groupID, ".", "/")
	baseURL += repository + "/" + groupPath + "/"
	return baseURL, nil
}

func getArtifactURL(baseURL, version string, artifact ArtifactDescription) string {
	url := baseURL

	// Generate artifact name including optional classifier
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

func uploadFile(client piperHttp.Sender, filePath, url string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open artifact file %s: %w", filePath, err)
	}

	defer file.Close()

	err = uploadToNexus(client, file, url)
	if err != nil {
		return fmt.Errorf("failed to upload artifact file %s: %w", filePath, err)
	}
	return nil
}

func uploadHash(client piperHttp.Sender, filePath, url string, hash hash.Hash, length int) error {
	hashReader, err := generateHashReader(filePath, hash, length)
	if err != nil {
		return fmt.Errorf("failed to generate hash %w", err)
	}
	err = uploadToNexus(client, hashReader, url)
	if err != nil {
		return fmt.Errorf("failed to upload hash %w", err)
	}
	return nil
}

func uploadToNexus(client piperHttp.Sender, stream io.Reader, url string) error {
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
	_, err = io.Copy(hash, file)
	if err != nil {
		return nil, err
	}

	// Get the requested number of bytes from the hash
	hashInBytes := hash.Sum(nil)[:length]

	// Convert the bytes to a string
	hexString := hex.EncodeToString(hashInBytes)

	// Finally create an io.Reader wrapping the string
	return strings.NewReader(hexString), nil
}
