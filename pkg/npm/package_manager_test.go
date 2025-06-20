//go:build unit
// +build unit

package npm

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPackageManager(t *testing.T) {

	t.Run("Test detect package manager", func(t *testing.T) {
		tests := []struct {
			name          string
			existingFiles map[string]bool
			expectedPM    string
		}{
			{
				name: "npm with package-lock.json",
				existingFiles: map[string]bool{
					"package-lock.json": true,
					"yarn.lock":         false,
					"pnpm-lock.yaml":    false,
				},
				expectedPM: "npm",
			},
			{
				name: "yarn with yarn.lock",
				existingFiles: map[string]bool{
					"package-lock.json": false,
					"yarn.lock":         true,
					"pnpm-lock.yaml":    false,
				},
				expectedPM: "yarn",
			},
			{
				name: "pnpm with pnpm-lock.yaml",
				existingFiles: map[string]bool{
					"package-lock.json": false,
					"yarn.lock":         false,
					"pnpm-lock.yaml":    true,
				},
				expectedPM: "pnpm",
			},
			{
				name: "no lock file defaults to npm",
				existingFiles: map[string]bool{
					"package-lock.json": false,
					"yarn.lock":         false,
					"pnpm-lock.yaml":    false,
				},
				expectedPM: "npm",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				utils := newNpmMockUtilsBundle()
				for file, exists := range tt.existingFiles {
					if exists {
						utils.AddFile(file, []byte("{}"))
					}
				}

				exec := &Execute{
					Utils:   &utils,
					Options: ExecutorOptions{},
				}

				pm, err := exec.detectPackageManager("")

				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPM, pm.Name)
			})
		}
	})

	t.Run("Test package manager errors", func(t *testing.T) {
		tests := []struct {
			name          string
			mockSetup     func(utils *npmMockUtilsBundle)
			expectedError string
		}{
			{
				name: "file check error",
				mockSetup: func(utils *npmMockUtilsBundle) {
					utils.FileExistsErrors = map[string]error{
						"package-lock.json": fmt.Errorf("permission denied"),
					}
				},
				expectedError: "failed to check for package-lock.json: permission denied",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				utils := newNpmMockUtilsBundle()
				tt.mockSetup(&utils)

				exec := &Execute{
					Utils:   &utils,
					Options: ExecutorOptions{},
				}

				_, err := exec.detectPackageManager("")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			})
		}
	})

	t.Run("Test install errors", func(t *testing.T) {
		tests := []struct {
			name          string
			lockFile      string
			execError     error
			expectedError string
		}{
			{
				name:          "npm install command fails",
				lockFile:      "package-lock.json",
				execError:     fmt.Errorf("npm install failed"),
				expectedError: "npm install failed",
			},
			{
				name:          "yarn install command fails",
				lockFile:      "yarn.lock",
				execError:     fmt.Errorf("yarn install failed"),
				expectedError: "yarn install failed",
			},
			{
				name:          "pnpm install command fails",
				lockFile:      "pnpm-lock.yaml",
				execError:     fmt.Errorf("pnpm install failed"),
				expectedError: "pnpm install failed",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				utils := newNpmMockUtilsBundle()
				if tt.lockFile != "" {
					utils.AddFile(tt.lockFile, []byte("{}"))
				}
				if tt.execError != nil {
					utils.execRunner.ShouldFailOnCommand = map[string]error{
						"npm ci":                                tt.execError,
						"yarn install --frozen-lockfile":        tt.execError,
						pnpmPath + " install --frozen-lockfile": tt.execError,
					}
				}

				exec := &Execute{
					Utils:   &utils,
					Options: ExecutorOptions{},
				}

				err := exec.install("package.json", "")

				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			})
		}
	})

}
