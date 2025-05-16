//go:build unit

package piperutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProjectWithOnlyMtaFile(t *testing.T) {
	projectStructure := ProjectStructure{directory: "testdata/mta"}
	resultMta := projectStructure.UsesMta()
	assert.True(t, resultMta)
	resultPom := projectStructure.UsesMaven()
	assert.False(t, resultPom)
	resultNpm := projectStructure.UsesNpm()
	assert.False(t, resultNpm)
}

func TestProjectWithOnlyPomFile(t *testing.T) {
	projectStructure := ProjectStructure{directory: "testdata/maven"}
	resultMta := projectStructure.UsesMta()
	assert.False(t, resultMta)
	resultPom := projectStructure.UsesMaven()
	assert.True(t, resultPom)
	resultNpm := projectStructure.UsesNpm()
	assert.False(t, resultNpm)
}

func TestProjectWithOnlyNpmFile(t *testing.T) {
	projectStructure := ProjectStructure{directory: "testdata/npm"}
	resultMta := projectStructure.UsesMta()
	assert.False(t, resultMta)
	resultPom := projectStructure.UsesMaven()
	assert.False(t, resultPom)
	resultNpm := projectStructure.UsesNpm()
	assert.True(t, resultNpm)
}

func TestDirectryParameterIsEmptyAndNoProjectFilesAreInIt(t *testing.T) {
	projectStructure := ProjectStructure{}
	resultMta := projectStructure.UsesMta()
	assert.False(t, resultMta)
	resultPom := projectStructure.UsesMaven()
	assert.False(t, resultPom)
	resultNpm := projectStructure.UsesNpm()
	assert.False(t, resultNpm)
}
