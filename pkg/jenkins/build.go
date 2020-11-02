package jenkins

import (
	"errors"
	"fmt"
	"time"

	"github.com/bndr/gojenkins"
)

// Build is an interface to abstract gojenkins.Build.
type Build interface {
	GetArtifacts() []gojenkins.Artifact
	IsRunning() bool
}

// WaitForBuildToFinish waits till a build is finished.
func WaitForBuildToFinish(build Build, pollInterval time.Duration) {
	//TODO: handle timeout?
	for build.IsRunning() {
		time.Sleep(pollInterval)
		//TODO: build.Poll() needed?
	}
}

// FetchBuildArtifact is fetching a build artifact from a finished build with a certain name.
// Fails if build is running or no artifact is with the given name is found.
func FetchBuildArtifact(build Build, fileName string) (Artifact, error) {
	if build.IsRunning() {
		return &ArtifactImpl{}, errors.New("Failed to fetch artifact: Job is still running")
	}
	for _, artifact := range build.GetArtifacts() {
		if artifact.FileName == fileName {
			return &ArtifactImpl{artifact: artifact}, nil
		}
	}
	return &ArtifactImpl{}, fmt.Errorf("Failed to fetch artifact: Artifact '%s' not found", fileName)
}
