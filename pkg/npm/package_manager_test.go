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
					pnpmSetup: pnpmSetupState{
						rootDir: "/", // Mock root directory
					},
				}

				pm, err := exec.detectPackageManager()

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
					pnpmSetup: pnpmSetupState{
						rootDir: "/", // Mock root directory
					},
				}

				_, err := exec.detectPackageManager()
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
					// Mock expects absolute path for locally installed pnpm
					absolutePnpmPath := "/tmp/node_modules/.bin/pnpm"
					utils.execRunner.ShouldFailOnCommand = map[string]error{
						"npm ci":                                        tt.execError,
						"yarn install --frozen-lockfile":                tt.execError,
						"pnpm --version":                                fmt.Errorf("not found"),
						absolutePnpmPath + " --version":                 fmt.Errorf("not found"),
						absolutePnpmPath + " install --frozen-lockfile": tt.execError,
					}
				}

				exec := &Execute{
					Utils:   &utils,
					Options: ExecutorOptions{},
					pnpmSetup: pnpmSetupState{
						rootDir: "/", // Mock root directory
					},
				}

				err := exec.install("package.json")

				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			})
		}
	})

	t.Run("Test pnpm version configuration", func(t *testing.T) {
		tests := []struct {
			name        string
			pnpmVersion string
			expectedCmd string
			globalPnpm  bool
			localPnpm   bool
		}{
			{
				name:        "pnpm version empty",
				pnpmVersion: "",
				expectedCmd: "npm install pnpm --prefix ./tmp",
				globalPnpm:  false,
				localPnpm:   false,
			},
			{
				name:        "global pnpm available no version specified",
				pnpmVersion: "",
				expectedCmd: "",
				globalPnpm:  true,
				localPnpm:   false,
			},
			{
				name:        "local pnpm available no version specified",
				pnpmVersion: "",
				expectedCmd: "",
				globalPnpm:  false,
				localPnpm:   true,
			},
			{
				name:        "pnpm version latest",
				pnpmVersion: "latest",
				expectedCmd: "npm install pnpm@latest --prefix ./tmp",
				globalPnpm:  false,
				localPnpm:   false,
			},
			{
				name:        "pnpm version 8.15.0",
				pnpmVersion: "8.15.0",
				expectedCmd: "npm install pnpm@8.15.0 --prefix ./tmp",
				globalPnpm:  false,
				localPnpm:   false,
			},
			{
				name:        "global pnpm available with latest version specified",
				pnpmVersion: "latest",
				expectedCmd: "npm install pnpm@latest --prefix ./tmp",
				globalPnpm:  true,
				localPnpm:   false,
			},
			{
				name:        "local pnpm available with latest version specified",
				pnpmVersion: "latest",
				expectedCmd: "npm install pnpm@latest --prefix ./tmp",
				globalPnpm:  false,
				localPnpm:   true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				utils := newNpmMockUtilsBundle()
				utils.AddFile("pnpm-lock.yaml", []byte("{}"))

				// Mock pnpm version checks
				if tt.globalPnpm {
					utils.execRunner.ShouldFailOnCommand = map[string]error{
						"pnpm --version": nil,
					}
				} else if tt.localPnpm {
					absolutePnpmPath := "/tmp/node_modules/.bin/pnpm"
					utils.execRunner.ShouldFailOnCommand = map[string]error{
						"pnpm --version":                fmt.Errorf("command not found"),
						absolutePnpmPath + " --version": nil,
					}
				} else {
					absolutePnpmPath := "/tmp/node_modules/.bin/pnpm"
					utils.execRunner.ShouldFailOnCommand = map[string]error{
						"pnpm --version":                fmt.Errorf("command not found"),
						absolutePnpmPath + " --version": fmt.Errorf("command not found"),
					}
				}

				exec := &Execute{
					Utils:   &utils,
					Options: ExecutorOptions{PnpmVersion: tt.pnpmVersion},
					pnpmSetup: pnpmSetupState{
						rootDir: "/", // Mock root directory
					},
				}

				pm, err := exec.detectPackageManager()
				assert.NoError(t, err)
				assert.Equal(t, "pnpm", pm.Name)

				// Check the install command is set correctly
				if tt.pnpmVersion == "" && tt.globalPnpm {
					assert.Equal(t, "pnpm", pm.InstallCommand)
				} else {
					absolutePnpmPath := "/tmp/node_modules/.bin/pnpm"
					assert.Equal(t, absolutePnpmPath, pm.InstallCommand)
					// Verify the npm install command was called with correct version
					if tt.expectedCmd != "" {
						expectedParams := []string{"install", "pnpm", "--prefix", "/tmp"}
						if tt.pnpmVersion != "" {
							expectedParams[1] = fmt.Sprintf("pnpm@%s", tt.pnpmVersion)
						}
						found := false
						for _, call := range utils.execRunner.Calls {
							if call.Exec == "npm" && len(call.Params) == 4 &&
								call.Params[0] == "install" && call.Params[2] == "--prefix" && call.Params[3] == "/tmp" {
								if tt.pnpmVersion == "" && call.Params[1] == "pnpm" {
									found = true
									break
								} else if tt.pnpmVersion != "" && call.Params[1] == fmt.Sprintf("pnpm@%s", tt.pnpmVersion) {
									found = true
									break
								}
							}
						}
						assert.True(t, found, "Expected npm install command not found in exec calls")
					}
				}
			})
		}
	})
}
