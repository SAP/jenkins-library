//go:build unit
// +build unit

package versioning

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoModGetCoordinates(t *testing.T) {
	t.Run("with full module name", func(t *testing.T) {
		// prepare
		gomod := createTestGoModFile(t, "module github.com/path-to/moduleName\n\ngo 1.24.0")
		// test
		coordinates, err := gomod.GetCoordinates()
		// assert
		assert.NoError(t, err)
		assert.Equal(t, "moduleName", coordinates.ArtifactID)
		assert.Equal(t, "github.com/path-to", coordinates.GroupID)
	})
	t.Run("with module name without path", func(t *testing.T) {
		// prepare
		gomod := createTestGoModFile(t, "module github.com/moduleName\n\ngo 1.24.0")
		// test
		coordinates, err := gomod.GetCoordinates()
		// assert
		assert.NoError(t, err)
		assert.Equal(t, "moduleName", coordinates.ArtifactID)
		assert.Equal(t, "github.com", coordinates.GroupID)
	})
	t.Run("with invalid simple module name", func(t *testing.T) {
		// prepare
		gomod := createTestGoModFile(t, "module moduleName\n\ngo 1.24.0")
		// test
		coordinates, err := gomod.GetCoordinates()
		// assert
		assert.ErrorContains(t, err, "missing dot in first path element")
		assert.Empty(t, coordinates.ArtifactID)
		assert.Empty(t, coordinates.GroupID)
	})
	t.Run("with invalid full module name", func(t *testing.T) {
		// prepare
		gomod := createTestGoModFile(t, "module path/to/test\n\ngo 1.24.0")
		// test
		coordinates, err := gomod.GetCoordinates()
		// assert
		assert.ErrorContains(t, err, "missing dot in first path element")
		assert.Empty(t, coordinates.ArtifactID)
		assert.Empty(t, coordinates.GroupID)
	})
}

func createTestGoModFile(t *testing.T, content string) *GoMod {
	tmpFolder := t.TempDir()
	goModFilePath := filepath.Join(tmpFolder, "go.mod")
	os.WriteFile(goModFilePath, []byte(content), 0666)
	gomod := &GoMod{
		path: goModFilePath,
	}
	return gomod
}
