package piperenv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Artifacts_FindByKind(t *testing.T) {
	artifacts := Artifacts([]Artifact{{
		LocalPath: "target/a.jar",
		Kind:      "jar",
	}, {
		LocalPath: "b",
		Kind:      "elf",
	}})

	assert.Len(t, artifacts.FindByKind("garbage"), 0)

	filtered := artifacts.FindByKind("jar")
	require.Len(t, filtered, 1)
	assert.Equal(t, "target/a.jar", filtered[0].LocalPath)
}

func Test_Artifacts_FindByName(t *testing.T) {
	artifacts := Artifacts([]Artifact{{
		LocalPath: "target/a.jar",
		Kind:      "jar",
		Name:      "a.jar",
	}})

	assert.Len(t, artifacts.FindByName("garbage"), 0)

	filtered := artifacts.FindByName("a.jar")
	require.Len(t, filtered, 1)
	assert.Equal(t, "a.jar", filtered[0].Name)
}
