package versioning

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetArtifact(t *testing.T) {
	t.Run("maven", func(t *testing.T) {
		maven, err := GetArtifact("maven", "my/pom.xml", &Options{}, nil)
		assert.NoError(t, err)
		assert.Equal(t, "maven", maven.VersioningScheme())
	})

	t.Run("npm", func(t *testing.T) {
		npm, err := GetArtifact("npm", "my/package.json", &Options{}, nil)
		assert.NoError(t, err)
		assert.Equal(t, "semver2", npm.VersioningScheme())
	})

	t.Run("not supported build tool", func(t *testing.T) {
		_, err := GetArtifact("nosupport", "whatever", &Options{}, nil)
		assert.EqualError(t, err, "build tool 'nosupport' not supported")
	})
}
