package nexus

import (
	"errors"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/log"
	"strings"
)

// ArtifactDescription describes a single artifact that can be uploaded to a Nexus repository manager.
// The File string must point to an existing file. The Classifier can be empty.
type ArtifactDescription struct {
	Classifier string `json:"classifier"`
	Type       string `json:"type"`
	File       string `json:"file"`
}

// Upload combines information about an artifact and its sub-artifacts which are supposed to be uploaded together.
// Call SetRepoURL(), SetArtifactsVersion(), SetArtifactID(), and add at least one artifact via AddArtifact().
type Upload struct {
	repoURL    string
	version    string
	artifactID string
	artifacts  []ArtifactDescription
}

// Uploader provides an interface for configuring the target Nexus Repository and adding artifacts.
type Uploader interface {
	SetRepoURL(nexusURL, nexusVersion, repository string) error
	GetRepoURL() string
	SetArtifactsID(version string) error
	GetArtifactsID() string
	SetArtifactsVersion(version string) error
	GetArtifactsVersion() string
	AddArtifact(artifact ArtifactDescription) error
	GetArtifacts() []ArtifactDescription
	Clear()
}

// SetRepoURL constructs the base URL to the Nexus repository. No parameter can be empty.
func (nexusUpload *Upload) SetRepoURL(nexusURL, nexusVersion, repository string) error {
	repoURL, err := getBaseURL(nexusURL, nexusVersion, repository)
	if err != nil {
		return err
	}
	nexusUpload.repoURL = repoURL
	return nil
}

// GetRepoURL returns the base URL for the nexus repository.
func (nexusUpload *Upload) GetRepoURL() string {
	return nexusUpload.repoURL
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

// GetArtifactsVersion returns the common version for all artifacts.
func (nexusUpload *Upload) GetArtifactsVersion() string {
	return nexusUpload.version
}

// SetArtifactsID sets the common ID for all uploaded artifacts. The ID is external to
// the artifact descriptions so that it is consistent for all of them.
func (nexusUpload *Upload) SetArtifactsID(id string) error {
	if id == "" {
		return errors.New("id must not be empty")
	}
	if strings.Contains(id, "/") {
		return fmt.Errorf("artifact ID may not include slashes")
	}
	nexusUpload.artifactID = id
	return nil
}

// GetArtifactsID returns the common version for all artifacts.
func (nexusUpload *Upload) GetArtifactsID() string {
	return nexusUpload.artifactID
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
	if artifact.File == "" || artifact.Type == "" {
		return fmt.Errorf("Artifact.File (%v) or Type (%v) is empty",
			artifact.File, artifact.Type)
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

// Clear removes any contained artifact descriptions.
func (nexusUpload *Upload) Clear() {
	nexusUpload.artifacts = []ArtifactDescription{}
}

func getBaseURL(nexusURL, nexusVersion, repository string) (string, error) {
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
	baseURL := nexusURL
	switch nexusVersion {
	case "nexus2":
		baseURL += "/content/repositories/"
	case "nexus3":
		baseURL += "/repository/"
	default:
		return "", fmt.Errorf("unsupported Nexus version '%s', must be 'nexus2' or 'nexus3'", nexusVersion)
	}
	baseURL += repository + "/"
	// Replace any double slashes, as nexus does not like them
	baseURL = strings.ReplaceAll(baseURL, "//", "/")
	return baseURL, nil
}
