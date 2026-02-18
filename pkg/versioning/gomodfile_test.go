package versioning

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
			os.WriteFile(goModFilePath, fmt.Appendf(nil, "module %s\n\ngo 1.24.0", tc.moduleName), 0666)
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
