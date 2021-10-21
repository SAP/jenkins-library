package gcs

import "path"

type TargetFolderHandler interface {
	GetTargetFolder() (string, error)
}

type TargetFolderHandle struct {
	FolderPath     string
	StepResultType string
	SubFolder      string
}

func (t *TargetFolderHandle) GetTargetFolder() (string, error) {
	return path.Join(t.FolderPath, t.StepResultType, t.SubFolder), nil
}
