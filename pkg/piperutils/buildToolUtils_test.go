package piperutils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProjectWithOnlyMtaFile(t *testing.T) {
	wd, _ := os.Getwd()
	os.Chdir("testdata/mta")
	defer os.Chdir(wd)
	resultMta := UsesMta()
	assert.True(t, resultMta)
	resultPom := UsesMaven()
	assert.False(t, resultPom)
	resultNpm := UsesNpm()
	assert.False(t, resultNpm)
}

func TestProjectWithOnlyPomFile(t *testing.T) {
	wd, _ := os.Getwd()
	os.Chdir("testdata/maven")
	defer os.Chdir(wd)
	resultMta := UsesMta()
	assert.False(t, resultMta)
	resultPom := UsesMaven()
	assert.True(t, resultPom)
	resultNpm := UsesNpm()
	assert.False(t, resultNpm)
}

func TestProjectWithOnlyNpmFile(t *testing.T) {
	wd, _ := os.Getwd()
	os.Chdir("testdata/npm")
	defer os.Chdir(wd)
	resultMta := UsesMta()
	assert.False(t, resultMta)
	resultPom := UsesMaven()
	assert.False(t, resultPom)
	resultNpm := UsesNpm()
	assert.True(t, resultNpm)
}
