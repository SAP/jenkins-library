//go:build unit
// +build unit

package cloudfoundry

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVarsFiles(t *testing.T) {
	t.Chdir(t.TempDir())
	assert.NoError(t, os.WriteFile("varsA.yml", []byte("file content does not matter"), 0o644))
	assert.NoError(t, os.WriteFile("varsB.yml", []byte("file content does not matter"), 0o644))

	t.Run("All vars files found", func(t *testing.T) {
		opts, err := GetVarsFileOptions([]string{"varsA.yml", "varsB.yml"})
		if assert.NoError(t, err) {
			assert.Equal(t, []string{"--vars-file", "varsA.yml", "--vars-file", "varsB.yml"}, opts)
		}
	})

	t.Run("Some vars files missing", func(t *testing.T) {
		opts, err := GetVarsFileOptions([]string{"varsA.yml", "varsC.yml", "varsD.yml"})
		if assert.EqualError(t, err, "Some vars files could not be found: [varsC.yml varsD.yml]") {
			assert.IsType(t, &VarsFilesNotFoundError{}, err)
			assert.Equal(t, []string{"--vars-file", "varsA.yml"}, opts)
		}
	})
}

func TestVars(t *testing.T) {
	t.Run("Empty vars", func(t *testing.T) {
		opts, err := GetVarsOptions([]string{})
		if assert.NoError(t, err) {
			assert.Equal(t, []string{}, opts)
		}
	})

	t.Run("Some vars", func(t *testing.T) {
		opts, err := GetVarsOptions([]string{"a=b", "x=y"})
		if assert.NoError(t, err) {
			assert.Equal(t, []string{"--var", "a=b", "--var", "x=y"}, opts)
		}
	})
}
