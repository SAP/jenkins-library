package jenkins

import (
	"github.com/bndr/gojenkins"
)

// Artifact is an interface to abstract gojenkins.Artifact.
type Artifact interface {
	SaveToDir(dir string) (bool, error)
	GetData() ([]byte, error)
	FileName() string
}

// ArtifactImpl is a wrapper struct for gojenkins.Artifact that respects the Artifact interface.
type ArtifactImpl struct {
	artifact gojenkins.Artifact
}

// SaveToDir refers to the gojenkins.Artifact.SaveToDir function.
func (a *ArtifactImpl) SaveToDir(dir string) (bool, error) {
	return a.artifact.SaveToDir(dir)
}

// GetData refers to the gojenkins.Artifact.GetData function.
func (a *ArtifactImpl) GetData() ([]byte, error) {
	return a.artifact.GetData()
}

// FileName refers to the gojenkins.Artifact.FileName field.
func (a *ArtifactImpl) FileName() string {
	return a.artifact.FileName
}
