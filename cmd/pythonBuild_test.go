//go:build unit
// +build unit

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/telemetry"

	"github.com/stretchr/testify/assert"
)

type pythonBuildMockUtils struct {
	config *pythonBuildOptions
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newPythonBuildTestsUtils() pythonBuildMockUtils {
	utils := pythonBuildMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func (f *pythonBuildMockUtils) GetConfig() *pythonBuildOptions {
	return f.config
}

func TestRunPythonBuild(t *testing.T) {
	cpe := pythonBuildCommonPipelineEnvironment{}
	t.Run("success - build", func(t *testing.T) {
		config := pythonBuildOptions{
			VirtualEnvironmentName: "dummy",
		}
		utils := newPythonBuildTestsUtils()
		telemetryData := telemetry.CustomData{}

		runPythonBuild(&config, &telemetryData, utils, &cpe)
		assert.Equal(t, "python3", utils.ExecMockRunner.Calls[0].Exec)
		assert.Equal(t, []string{"-m", "venv", "dummy"}, utils.ExecMockRunner.Calls[0].Params)
	})

	t.Run("failure - build failure", func(t *testing.T) {
		config := pythonBuildOptions{}
		utils := newPythonBuildTestsUtils()
		utils.ShouldFailOnCommand = map[string]error{"python setup.py sdist bdist_wheel": fmt.Errorf("build failure")}
		telemetryData := telemetry.CustomData{}

		err := runPythonBuild(&config, &telemetryData, utils, &cpe)
		assert.EqualError(t, err, "failed to build package: build failure")
	})

	t.Run("success - publishes binaries", func(t *testing.T) {
		config := pythonBuildOptions{
			Publish:                  true,
			TargetRepositoryURL:      "https://my.target.repository.local",
			TargetRepositoryUser:     "user",
			TargetRepositoryPassword: "password",
			VirtualEnvironmentName:   "dummy",
		}
		utils := newPythonBuildTestsUtils()
		telemetryData := telemetry.CustomData{}

		runPythonBuild(&config, &telemetryData, utils, &cpe)
		assert.Equal(t, "python3", utils.ExecMockRunner.Calls[0].Exec)
		assert.Equal(t, []string{"-m", "venv", config.VirtualEnvironmentName}, utils.ExecMockRunner.Calls[0].Params)
		assert.Equal(t, "bash", utils.ExecMockRunner.Calls[1].Exec)
		assert.Equal(t, []string{"-c", "source " + filepath.Join("dummy", "bin", "activate")}, utils.ExecMockRunner.Calls[1].Params)
		assert.Equal(t, "dummy/bin/pip", utils.ExecMockRunner.Calls[2].Exec)
		assert.Equal(t, []string{"install", "--upgrade", "--root-user-action=ignore", "wheel"}, utils.ExecMockRunner.Calls[2].Params)
		assert.Equal(t, "dummy/bin/python", utils.ExecMockRunner.Calls[3].Exec)
		assert.Equal(t, []string{"setup.py", "sdist", "bdist_wheel"}, utils.ExecMockRunner.Calls[3].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "pip"), utils.ExecMockRunner.Calls[4].Exec)
		assert.Equal(t, []string{"install", "--upgrade", "--root-user-action=ignore", "twine"}, utils.ExecMockRunner.Calls[4].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "twine"), utils.ExecMockRunner.Calls[5].Exec)
		assert.Equal(t, []string{"upload", "--username", config.TargetRepositoryUser,
			"--password", config.TargetRepositoryPassword, "--repository-url", config.TargetRepositoryURL,
			"--disable-progress-bar", "dist/*"}, utils.ExecMockRunner.Calls[5].Params)
	})

	t.Run("success - create BOM", func(t *testing.T) {
		config := pythonBuildOptions{
			CreateBOM:              true,
			Publish:                false,
			VirtualEnvironmentName: "dummy",
		}
		utils := newPythonBuildTestsUtils()
		telemetryData := telemetry.CustomData{}

		runPythonBuild(&config, &telemetryData, utils, &cpe)
		// assert.NoError(t, err)
		assert.Equal(t, "python3", utils.ExecMockRunner.Calls[0].Exec)
		assert.Equal(t, []string{"-m", "venv", config.VirtualEnvironmentName}, utils.ExecMockRunner.Calls[0].Params)
		assert.Equal(t, "bash", utils.ExecMockRunner.Calls[1].Exec)
		assert.Equal(t, []string{"-c", "source " + filepath.Join("dummy", "bin", "activate")}, utils.ExecMockRunner.Calls[1].Params)
		assert.Equal(t, "dummy/bin/pip", utils.ExecMockRunner.Calls[2].Exec)
		assert.Equal(t, []string{"install", "--upgrade", "--root-user-action=ignore", "wheel"}, utils.ExecMockRunner.Calls[2].Params)
		assert.Equal(t, "dummy/bin/python", utils.ExecMockRunner.Calls[3].Exec)
		assert.Equal(t, []string{"setup.py", "sdist", "bdist_wheel"}, utils.ExecMockRunner.Calls[3].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "pip"), utils.ExecMockRunner.Calls[4].Exec)
		assert.Equal(t, []string{"install", "--upgrade", "--root-user-action=ignore", "cyclonedx-bom==6.1.1"}, utils.ExecMockRunner.Calls[4].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "cyclonedx-py"), utils.ExecMockRunner.Calls[5].Exec)
		assert.Equal(t, []string{"env", "--output-file", "bom-pip.xml", "--output-format", "XML", "--spec-version", "1.4"}, utils.ExecMockRunner.Calls[5].Params)
	})
}

// func Test_renameArtifactsInDistM(t *testing.T) {
// 	// Define test cases
// 	tests := []struct {
// 		testName         string
// 		inputFilename    string
// 		expectedFilename string
// 		shouldRename     bool // Whether renaming is expected
// 	}{
// 		{
// 			testName:         "Rename file with underscore in the name",
// 			inputFilename:    "test_artifact_1.0.0.tar.gz",
// 			expectedFilename: "test_artifact-1.0.0.tar.gz",
// 			shouldRename:     true,
// 		},
// 		{
// 			testName:         "Rename file with multiple underscores",
// 			inputFilename:    "another_test_artifact_1_0_0.tar.gz",
// 			expectedFilename: "another_test_artifact-1-0-0.tar.gz",
// 			shouldRename:     true,
// 		},
// 		{
// 			testName:         "Do not rename file without underscore",
// 			inputFilename:    "test-artifact-1.0.0.tar.gz",
// 			expectedFilename: "test-artifact-1.0.0.tar.gz",
// 			shouldRename:     false,
// 		},
// 		{
// 			testName:         "Ignore non-Python artifact file (ZIP)",
// 			inputFilename:    "random_file_1.0.0.zip",
// 			expectedFilename: "random_file_1.0.0.zip",
// 			shouldRename:     false,
// 		},
// 	}

// 	// Create a temporary directory for testing
// 	distTempDir := t.TempDir()

// 	// Create test files in the temporary directory
// 	for _, tt := range tests {
// 		filePath := filepath.Join(distTempDir, tt.inputFilename)
// 		if err := os.WriteFile(filePath, []byte("sample"), 0644); err != nil {
// 			t.Fatalf("Failed to create test file %s: %v", tt.inputFilename, err)
// 		}
// 	}

// 	// Run the function under test
// 	renameArtifactsInDist(distTempDir)

// 	// Validate the results
// 	for _, tt := range tests {
// 		t.Run(tt.testName, func(t *testing.T) {
// 			oldPath := filepath.Join(distTempDir, tt.inputFilename)
// 			newPath := filepath.Join(distTempDir, tt.expectedFilename)

// 			if tt.inputFilename == tt.expectedFilename {
// 				// Special case: No renaming expected
// 				if _, err := os.Stat(oldPath); os.IsNotExist(err) {
// 					t.Errorf("Expected file %s to remain, but it does not exist", oldPath)
// 				}
// 				return
// 			}

// 			// General case: Renaming logic
// 			_, oldExistsErr := os.Stat(oldPath)
// 			_, newExistsErr := os.Stat(newPath)

// 			if tt.shouldRename {
// 				if oldExistsErr == nil {
// 					t.Errorf("Expected old file %s to be renamed, but it still exists", oldPath)
// 				}
// 				if os.IsNotExist(newExistsErr) {
// 					t.Errorf("Expected new file %s to exist, but it does not", newPath)
// 				}
// 			} else {
// 				if os.IsNotExist(oldExistsErr) {
// 					t.Errorf("Expected file %s to remain, but it does not exist", oldPath)
// 				}
// 				if newExistsErr == nil {
// 					t.Errorf("Expected file %s not to be renamed, but it was renamed to %s", tt.inputFilename, tt.expectedFilename)
// 				}
// 			}
// 		})
// 	}
// }
