package helper

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetFileContent(t *testing.T) {
	workingDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal("Failed to create temporary directory")
	}
	defer os.RemoveAll(workingDir)

	testFileName := "myTestFile"
	testFileContent := "this is my test file"
	testFilePath := filepath.Join(workingDir, testFileName)
	ioutil.WriteFile(testFilePath, []byte(testFileContent), 0644)

	tests := []struct {
		name       string
		file       string
		want       string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:       "empty file string",
			file:       "",
			want:       "",
			wantErr:    true,
			wantErrMsg: "failed to open file",
		},
		{
			name:    "test file",
			file:    testFilePath,
			want:    "this is my test file",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetFileContent(tt.file, nil)
			if err != nil {
				if tt.wantErr == false {
					t.Errorf("GetFileContent() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("GetFileContent() error = '%v' does not contain wantErr '%v'", err, tt.wantErrMsg)
					return
				}
			}
			if got != tt.want {
				t.Errorf("GetFileContent() = %v, want %v", got, tt.want)
			}
		})
	}
}
