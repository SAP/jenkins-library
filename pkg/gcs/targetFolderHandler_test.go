package gcs

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTargetFolder(t *testing.T) {
	tests := []struct {
		target   TargetFolderHandle
		expected string
	}{
		{target: TargetFolderHandle{FolderPath: "folder/path/", StepResultType: "general", SubFolder: "sub/folder"}, expected: "folder/path/general/sub/folder"},
		{target: TargetFolderHandle{FolderPath: "folder/path/", SubFolder: "sub/folder"}, expected: "folder/path/sub/folder"},
		{target: TargetFolderHandle{FolderPath: "folder/path/", StepResultType: "general"}, expected: "folder/path/general"},
		{target: TargetFolderHandle{FolderPath: "folder1", StepResultType: "general", SubFolder: "folder2/"}, expected: "folder1/general/folder2"}}

	for key, tt := range tests {
		t.Run(fmt.Sprintf("Row %v", key+1), func(t *testing.T) {
			actualTargetFolder, err := tt.target.GetTargetFolder()
			assert.NoError(t, err)
			if actualTargetFolder != tt.expected {
				t.Errorf("Expected '%v' was '%v'", tt.expected, actualTargetFolder)
			}
		})
	}
}
