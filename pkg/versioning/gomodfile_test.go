package versioning

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoModGetVersion(t *testing.T) {
	testCases := []struct {
		name           string
		goModContent   string
		versionFile    string // "VERSION", "version.txt", or "" for none
		versionContent string
		fileExistsFunc func(string) (bool, error)
		expectedVer    string
		expectedErr    string
	}{
		{
			name:           "reads from VERSION file when present",
			goModContent:   "module github.com/test/module\n\ngo 1.21",
			versionFile:    "VERSION",
			versionContent: "1.2.3",
			expectedVer:    "1.2.3",
			expectedErr:    "",
		},
		{
			name:           "reads from version.txt file when present",
			goModContent:   "module github.com/test/module\n\ngo 1.21",
			versionFile:    "version.txt",
			versionContent: "2.0.0",
			expectedVer:    "2.0.0",
			expectedErr:    "",
		},
		{
			name:           "prefers version.txt over VERSION",
			goModContent:   "module github.com/test/module\n\ngo 1.21",
			versionFile:    "both", // special case: create both files
			versionContent: "from-version-txt",
			expectedVer:    "from-version-txt",
			expectedErr:    "",
		},
		{
			name:           "error when no version file and go.mod has no version",
			goModContent:   "module github.com/test/module\n\ngo 1.21",
			versionFile:    "",
			versionContent: "",
			expectedVer:    "",
			expectedErr:    "go.mod has no version",
		},
		{
			name:           "error includes version file search failure",
			goModContent:   "module github.com/test/module\n\ngo 1.21",
			versionFile:    "",
			versionContent: "",
			expectedVer:    "",
			expectedErr:    "no version file found",
		},
		{
			name:           "handles version with trailing newline",
			goModContent:   "module github.com/test/module\n\ngo 1.21",
			versionFile:    "VERSION",
			versionContent: "1.0.0\n",
			expectedVer:    "1.0.0",
			expectedErr:    "",
		},
		{
			name:         "fileExists returns error falls back to go.mod",
			goModContent: "module github.com/test/module\n\ngo 1.21",
			fileExistsFunc: func(f string) (bool, error) {
				return false, fmt.Errorf("permission denied")
			},
			expectedVer: "",
			expectedErr: "go.mod has no version",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// prepare temp directory and change to it
			tmpFolder := t.TempDir()
			originalWd, _ := os.Getwd()
			os.Chdir(tmpFolder)
			defer os.Chdir(originalWd)

			goModFilePath := filepath.Join(tmpFolder, "go.mod")
			os.WriteFile(goModFilePath, []byte(tc.goModContent), 0o666)

			// create version file(s) if specified
			fileExistsFunc := tc.fileExistsFunc
			if fileExistsFunc == nil {
				switch tc.versionFile {
				case "VERSION":
					os.WriteFile("VERSION", []byte(tc.versionContent), 0o666)
					fileExistsFunc = func(f string) (bool, error) {
						return f == "VERSION", nil
					}
				case "version.txt":
					os.WriteFile("version.txt", []byte(tc.versionContent), 0o666)
					fileExistsFunc = func(f string) (bool, error) {
						return f == "version.txt", nil
					}
				case "both":
					os.WriteFile("version.txt", []byte(tc.versionContent), 0o666)
					os.WriteFile("VERSION", []byte("from-VERSION"), 0o666)
					fileExistsFunc = func(f string) (bool, error) {
						return f == "version.txt" || f == "VERSION", nil
					}
				default:
					fileExistsFunc = func(f string) (bool, error) { return false, nil }
				}
			}

			gomod := &GoMod{
				path:       goModFilePath,
				fileExists: fileExistsFunc,
			}

			// test
			version, err := gomod.GetVersion()

			// assert
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectedVer, version)
		})
	}
}

func TestGoModGetVersionNonGoMod(t *testing.T) {
	t.Run("reads directly when path is not go.mod", func(t *testing.T) {
		tmpFolder := t.TempDir()
		versionPath := filepath.Join(tmpFolder, "VERSION")
		os.WriteFile(versionPath, []byte("5.0.0"), 0o666)

		gomod := &GoMod{
			path: versionPath,
		}

		version, err := gomod.GetVersion()

		assert.NoError(t, err)
		assert.Equal(t, "5.0.0", version)
	})
}

func TestGoModGetCoordinates(t *testing.T) {
	testCases := []struct {
		name       string
		moduleName string
		artifact   string
		group      string
		err        string
	}{
		{"with full module name", "github.com/path-to/moduleName", "moduleName", "github.com/path-to", ""},
		{"with module name without path", "github.com/moduleName", "moduleName", "github.com", ""},
		{"with invalid simple module name", "moduleName", "", "", "missing dot in first path element"},
		{"with invalid full module name", "path/to/module", "", "", "missing dot in first path element"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// prepare
			tmpFolder := t.TempDir()
			goModFilePath := filepath.Join(tmpFolder, "go.mod")
			os.WriteFile(goModFilePath, []byte(fmt.Sprintf("module %s\n\ngo 1.24.0", tc.moduleName)), 0o666)
			gomod := &GoMod{
				path: goModFilePath,
			}
			// test
			coordinates, err := gomod.GetCoordinates()
			// assert
			if tc.err != "" {
				assert.ErrorContains(t, err, tc.err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.artifact, coordinates.ArtifactID)
			assert.Equal(t, tc.group, coordinates.GroupID)
		})
	}
}
