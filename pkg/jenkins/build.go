package jenkins

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bndr/gojenkins"
)

// Build is an interface to abstract gojenkins.Build.
// mock generated with: mockery --name Build --dir pkg/jenkins --output pkg/jenkins/mocks
type Build interface {
	GetArtifacts() []gojenkins.Artifact
	IsRunning(ctx context.Context) bool
}

// WaitForBuildToFinish waits till a build is finished.
func WaitForBuildToFinish(ctx context.Context, build Build, pollInterval time.Duration) {
	//TODO: handle timeout?
	for build.IsRunning(ctx) {
		time.Sleep(pollInterval)
		//TODO: build.Poll() needed?
	}
}

// FetchBuildArtifact is fetching a build artifact from a finished build with a certain name.
// Fails if build is running or no artifact is with the given name is found.
func FetchBuildArtifact(ctx context.Context, build Build, fileName string) (Artifact, error) {
	if build.IsRunning(ctx) {
		return &ArtifactImpl{}, errors.New("failed to fetch artifact: Job is still running")
	}
	for _, artifact := range build.GetArtifacts() {
		if artifact.FileName == fileName {
			return &ArtifactImpl{artifact: artifact}, nil
		}
	}
	return &ArtifactImpl{}, fmt.Errorf("failed to fetch artifact: Artifact '%s' not found", fileName)
}
