package nexus

import (
	"errors"
	"fmt"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
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
	protocol     string
	mavenRepoURL string
	npmRepoURL   string
	groupID      string
	version      string
	artifactID   string
	artifacts    []ArtifactDescription
}

// Uploader provides an interface for configuring the target Nexus Repository and adding artifacts.
type Uploader interface {
	SetRepoURL(nexusURL, nexusVersion, mavenRepository, npmRepository string) error
	GetNexusURLProtocol() string
	GetMavenRepoURL() string
	GetNpmRepoURL() string
	SetInfo(groupID, artifactsID, version string) error
	GetGroupID() string
	GetArtifactsID() string
	GetArtifactsVersion() string
	AddArtifact(artifact ArtifactDescription) error
	GetArtifacts() []ArtifactDescription
	Clear()
}

// SetRepoURL constructs the base URL to the Nexus repository. mavenRepository or npmRepository may be empty.
func (nexusUpload *Upload) SetRepoURL(nexusURL, nexusVersion, mavenRepository, npmRepository string) error {
	protocol, err := _GetNexusURLProtocol(nexusURL)
	if err != nil {
		return err
	}
	nexusUpload.protocol = protocol
	mavenRepoURL, err := getBaseURL(nexusURL, nexusVersion, mavenRepository)
	if err != nil {
		return err
	}
	nexusUpload.mavenRepoURL = mavenRepoURL
	npmRepositoryURL, err := getBaseURL(nexusURL, nexusVersion, npmRepository)
	if err != nil {
		return err
	}
	nexusUpload.npmRepoURL = npmRepositoryURL
	return nil
}

func getBaseURL(nexusURL, nexusVersion, repository string) (string, error) {
	if nexusURL == "" {
		return "", errors.New("nexusURL must not be empty")
	}
	nexusURL = strings.ToLower(nexusURL)
	var protocols = []string{"http://", "https://"}
	for _, protocol := range protocols {
		if strings.HasPrefix(nexusURL, protocol) {
			nexusURL = strings.TrimPrefix(nexusURL, protocol)
			break
		}
	}

	baseURL := nexusURL

	if repository != "" {
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
	}

	return baseURL, nil
}

// _GetNexusURLProtocol returns the protocol specified in the nexusUrl which was set thru setNexusUrl (internal method)
func _GetNexusURLProtocol(nexusURL string) (string, error) {
	if nexusURL == "" {
		return "", errors.New("nexusURL must not be empty")
	}
	nexusURL = strings.ToLower(nexusURL)
	var protocols = []string{"http://", "https://"}
	for _, protocol := range protocols {
		if strings.HasPrefix(nexusURL, protocol) {
			return strings.ReplaceAll(protocol, "://", ""), nil
		}
	}
	return "http", nil
}

// GetMavenRepoURL returns the base URL for the nexus-maven repository.
func (nexusUpload *Upload) GetMavenRepoURL() string {
	return nexusUpload.mavenRepoURL
}

// GetNexusURLProtocol returns the protocol specified in the nexusUrl which was set thru setNexusUrl
func (nexusUpload *Upload) GetNexusURLProtocol() string {
	if nexusUpload.protocol == "" {
		return "http"
	}
	return nexusUpload.protocol
}

// GetNpmRepoURL returns the base URL for the nexus-npm repository.
func (nexusUpload *Upload) GetNpmRepoURL() string {
	return nexusUpload.npmRepoURL
}

// ErrEmptyGroupID is returned from SetInfo, if groupID is empty.
var ErrEmptyGroupID = errors.New("groupID must not be empty")

// ErrEmptyArtifactID is returned from SetInfo, if artifactID is empty.
var ErrEmptyArtifactID = errors.New("artifactID must not be empty")

// ErrInvalidArtifactID is returned from SetInfo, if artifactID contains slashes.
var ErrInvalidArtifactID = errors.New("artifactID may not include slashes")

// ErrEmptyVersion is returned from SetInfo, if version is empty.
var ErrEmptyVersion = errors.New("version must not be empty")

// SetInfo sets the common info for all uploaded artifacts. This info is external to
// the artifact descriptions so that it is consistent for all of them.
func (nexusUpload *Upload) SetInfo(groupID, artifactID, version string) error {
	if groupID == "" {
		return ErrEmptyGroupID
	}
	if artifactID == "" {
		return ErrEmptyArtifactID
	}
	if strings.Contains(artifactID, "/") {
		return ErrInvalidArtifactID
	}
	if version == "" {
		return ErrEmptyVersion
	}
	nexusUpload.groupID = groupID
	nexusUpload.artifactID = artifactID
	nexusUpload.version = version
	return nil
}

// GetArtifactsVersion returns the common version for all artifacts.
func (nexusUpload *Upload) GetArtifactsVersion() string {
	return nexusUpload.version
}

// GetGroupID returns the common groupId for all artifacts.
func (nexusUpload *Upload) GetGroupID() string {
	return nexusUpload.groupID
}

// GetArtifactsID returns the common artifactId for all artifacts.
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
