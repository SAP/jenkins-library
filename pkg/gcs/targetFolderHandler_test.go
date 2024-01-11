//go:build unit
// +build unit

package gcs

import (
	"fmt"
	"testing"
)

func TestGetTargetFolder(t *testing.T) {
	tests := []struct {
		folderPath     string
		stepResultType string
		subFolder      string
		expected       string
	}{
		{folderPath: "folder/path/", stepResultType: "general", subFolder: "sub/folder", expected: "folder/path/general/sub/folder"},
		{folderPath: "folder/path/", subFolder: "sub/folder", expected: "folder/path/sub/folder"},
		{folderPath: "folder/path/", stepResultType: "general", expected: "folder/path/general"},
		{folderPath: "folder1", stepResultType: "general", subFolder: "folder2/", expected: "folder1/general/folder2"}}

	for key, tt := range tests {
		t.Run(fmt.Sprintf("Row %v", key+1), func(t *testing.T) {
			actualTargetFolder := GetTargetFolder(tt.folderPath, tt.stepResultType, tt.subFolder)
			if actualTargetFolder != tt.expected {
				t.Errorf("Expected '%v' was '%v'", tt.expected, actualTargetFolder)
			}
		})
	}
}
