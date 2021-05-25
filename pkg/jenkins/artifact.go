package jenkins

import (
	"context"

	"github.com/bndr/gojenkins"
)

// Artifact is an interface to abstract gojenkins.Artifact.
// mock generated with: mockery --name Artifact --dir pkg/jenkins --output pkg/jenkins/mocks
type Artifact interface {
	Save(path string) (bool, error)
	SaveToDir(dir string) (bool, error)
	GetData() ([]byte, error)
	FileName() string
}

// ArtifactImpl is a wrapper struct for gojenkins.Artifact that respects the Artifact interface.
type ArtifactImpl struct {
	artifact gojenkins.Artifact
}

// Save refers to the gojenkins.Artifact.Save function.
func (a *ArtifactImpl) Save(path string) (bool, error) {
	return a.artifact.Save(context.Background(), path)
}

// SaveToDir refers to the gojenkins.Artifact.SaveToDir function.
func (a *ArtifactImpl) SaveToDir(dir string) (bool, error) {
	return a.artifact.SaveToDir(context.Background(), dir)
}

// GetData refers to the gojenkins.Artifact.GetData function.
func (a *ArtifactImpl) GetData() ([]byte, error) {
	return a.artifact.GetData(context.Background())
}

// FileName refers to the gojenkins.Artifact.FileName field.
func (a *ArtifactImpl) FileName() string {
	return a.artifact.FileName
}
