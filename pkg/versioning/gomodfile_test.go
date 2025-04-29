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
	t.Run("simple module name", func(t *testing.T) {
		// prepare
		tmpFolder := t.TempDir()
		goModFilePath := filepath.Join(tmpFolder, "go.mod")
		os.WriteFile(goModFilePath, []byte("module test\n\ngo 1.24.0"), 0666)
		gomod := &GoMod{
			path: goModFilePath,
		}  
		// test
		coordinates, err := gomod.GetCoordinates()
		// assert
		assert.NoError(t, err)
		assert.Equal(t, "test", coordinates.ArtifactID)
	})
}
