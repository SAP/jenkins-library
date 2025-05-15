package jenkins

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bndr/gojenkins"
)

// Build is an interface to abstract gojenkins.Build.
type Build interface {
	GetArtifacts() []gojenkins.Artifact
	IsRunning(ctx context.Context) bool
	Poll(ctx context.Context, options ...interface{}) (int, error)
}

// WaitForBuildToFinish waits till a build is finished.
func WaitForBuildToFinish(ctx context.Context, build Build, pollInterval time.Duration) error {
	//TODO: handle timeout?
	maxRetries := 4

	for build.IsRunning(ctx) {
		time.Sleep(pollInterval)
		//TODO: add 404/503 response code handling
		_, err := build.Poll(ctx)

		if err == nil {
			continue
		}

		fmt.Printf("Error occurred while waiting for build to finish: %v. Retrying...\n", err)

		for i := 0; i < maxRetries; i++ {
			time.Sleep(pollInterval)
			_, err = build.Poll(ctx)

			if err == nil {
				break
			}

			fmt.Printf("Error occurred while waiting for build to finish: %v. Retrying...\n", err)
		}

		if err != nil {
			return fmt.Errorf("Max retries (%v) exceeded while waiting for build to finish. Last error: %w", maxRetries, err)
		}
	}

	return nil
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
