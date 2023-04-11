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
	Poll(ctx context.Context, options ...interface{}) (int, error)
}

// WaitForBuildToFinish waits till a build is finished.
func WaitForBuildToFinishWithRetry(ctx context.Context, build Build, pollInterval time.Duration) error {
	var err error
	var maxRetries int = 5
	var retryInterval time.Duration = 10 * time.Second

	for build.IsRunning(ctx) {
		time.Sleep(pollInterval)
	}

	for i := 0; i < maxRetries; i++ {
		time.Sleep(pollInterval)
		_, err = build.Poll(ctx)
		if err != nil {
			break
		}
		if err == nil {
			return nil
		}
		fmt.Printf("Error occurred while waiting for build to finish: %v. Retrying after %v\n", err, retryInterval)
		// Sleep for the retry interval before trying again.
		time.Sleep(retryInterval)
	}
	fmt.Printf("Max retries (%v) exceeded while waiting for build to finish. Last error: %v\n", maxRetries, err)
	return err
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
