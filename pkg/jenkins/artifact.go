package jenkins

import (
	"context"

	"github.com/bndr/gojenkins"
)

// Artifact is an interface to abstract gojenkins.Artifact.
type Artifact interface {
	Save(ctx context.Context, path string) (bool, error)
	SaveToDir(ctx context.Context, dir string) (bool, error)
	GetData(ctx context.Context) ([]byte, error)
	FileName() string
}

// ArtifactImpl is a wrapper struct for gojenkins.Artifact that respects the Artifact interface.
type ArtifactImpl struct {
	artifact gojenkins.Artifact
}

// Save refers to the gojenkins.Artifact.Save function.
func (a *ArtifactImpl) Save(ctx context.Context, path string) (bool, error) {
	return a.artifact.Save(ctx, path)
}

// SaveToDir refers to the gojenkins.Artifact.SaveToDir function.
func (a *ArtifactImpl) SaveToDir(ctx context.Context, dir string) (bool, error) {
	return a.artifact.SaveToDir(ctx, dir)
}

// GetData refers to the gojenkins.Artifact.GetData function.
func (a *ArtifactImpl) GetData(ctx context.Context) ([]byte, error) {
	return a.artifact.GetData(ctx)
}

// FileName refers to the gojenkins.Artifact.FileName field.
func (a *ArtifactImpl) FileName() string {
	return a.artifact.FileName
}
