package versioning

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetArtifact(t *testing.T) {
	t.Run("dub", func(t *testing.T) {
		dub, err := GetArtifact("dub", "my/dub.json", &Options{}, nil)
		assert.NoError(t, err)
		assert.Equal(t, "semver2", dub.VersioningScheme())
	})

	t.Run("golang", func(t *testing.T) {
		golang, err := GetArtifact("golang", "my/VERSION", &Options{}, nil)
		assert.NoError(t, err)
		assert.Equal(t, "semver2", golang.VersioningScheme())
	})

	t.Run("maven", func(t *testing.T) {
		maven, err := GetArtifact("maven", "my/pom.xml", &Options{}, nil)
		assert.NoError(t, err)
		assert.Equal(t, "maven", maven.VersioningScheme())
	})

	t.Run("mta", func(t *testing.T) {
		mta, err := GetArtifact("mta", "my/mta.yaml", &Options{}, nil)
		assert.NoError(t, err)
		assert.Equal(t, "semver2", mta.VersioningScheme())
	})

	t.Run("npm", func(t *testing.T) {
		npm, err := GetArtifact("npm", "my/package.json", &Options{}, nil)
		assert.NoError(t, err)
		assert.Equal(t, "semver2", npm.VersioningScheme())
	})

	t.Run("pip", func(t *testing.T) {
		pip, err := GetArtifact("pip", "my/version.txt", &Options{}, nil)
		assert.NoError(t, err)
		assert.Equal(t, "pep440", pip.VersioningScheme())
	})

	t.Run("sbt", func(t *testing.T) {
		sbt, err := GetArtifact("sbt", "my/sbtDescriptor.json", &Options{}, nil)
		assert.NoError(t, err)
		assert.Equal(t, "semver2", sbt.VersioningScheme())
	})

	t.Run("version", func(t *testing.T) {
		version, err := GetArtifact("version", "my/VERSION", &Options{}, nil)
		assert.NoError(t, err)
		assert.Equal(t, "semver2", version.VersioningScheme())
	})

	t.Run("not supported build tool", func(t *testing.T) {
		_, err := GetArtifact("nosupport", "whatever", &Options{}, nil)
		assert.EqualError(t, err, "build tool 'nosupport' not supported")
	})
}
