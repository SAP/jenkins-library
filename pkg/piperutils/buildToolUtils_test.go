package piperutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProjectWithOnlyMtaFile(t *testing.T) {
	b := buildTools{directory: "testdata/mta"}
	resultMta := b.UsesMta()
	assert.True(t, resultMta)
	resultPom := b.UsesMaven()
	assert.False(t, resultPom)
	resultNpm := b.UsesNpm()
	assert.False(t, resultNpm)
}

func TestProjectWithOnlyPomFile(t *testing.T) {
	b := buildTools{directory: "testdata/maven"}
	resultMta := b.UsesMta()
	assert.False(t, resultMta)
	resultPom := b.UsesMaven()
	assert.True(t, resultPom)
	resultNpm := b.UsesNpm()
	assert.False(t, resultNpm)
}

func TestProjectWithOnlyNpmFile(t *testing.T) {
	b := buildTools{directory: "testdata/npm"}
	resultMta := b.UsesMta()
	assert.False(t, resultMta)
	resultPom := b.UsesMaven()
	assert.False(t, resultPom)
	resultNpm := b.UsesNpm()
	assert.True(t, resultNpm)
}

func TestDirectryParameterIsEmptyAndNoProjectFilesAreInIt(t *testing.T) {
	b := buildTools{}
	resultMta := b.UsesMta()
	assert.False(t, resultMta)
	resultPom := b.UsesMaven()
	assert.False(t, resultPom)
	resultNpm := b.UsesNpm()
	assert.False(t, resultNpm)
}
