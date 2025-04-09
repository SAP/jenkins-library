//go:build unit

package maven

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

const aggregatorPomXML = `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <parent>
        <artifactId>artifact</artifactId>
        <groupId>group</groupId>
        <version>1.0-SNAPSHOT</version>
    </parent>
    <modelVersion>4.0.0</modelVersion>

    <artifactId>project-aggregator</artifactId>
    <packaging>pom</packaging>

    <modules>
        <module>sub1</module>
        <module>sub2</module>
    </modules>

</project>
`

func TestParsePOM(t *testing.T) {
	t.Parallel()

	t.Run("no XML data provided gives error", func(t *testing.T) {
		project, err := ParsePOM(nil)
		assert.EqualError(t, err, "failed to parse POM data: EOF")
		assert.Nil(t, project)
	})

	t.Run("modules evaluated", func(t *testing.T) {
		project, err := ParsePOM([]byte(aggregatorPomXML))
		require.NoError(t, err)
		require.NotNil(t, project)
		require.Len(t, project.Modules, 2)
		assert.Equal(t, project.Modules[0], "sub1")
		assert.Equal(t, project.Modules[1], "sub2")
	})

	t.Run("artifact coordinates", func(t *testing.T) {
		project, err := ParsePOM([]byte(aggregatorPomXML))
		require.NoError(t, err)
		require.NotNil(t, project)
		assert.Equal(t, project.ArtifactID, "project-aggregator")
		assert.Equal(t, project.Packaging, "pom")
	})

	t.Run("parent coordinates", func(t *testing.T) {
		project, err := ParsePOM([]byte(aggregatorPomXML))
		require.NoError(t, err)
		require.NotNil(t, project)
		assert.Equal(t, project.Parent.Version, "1.0-SNAPSHOT")
		assert.Equal(t, project.Parent.GroupID, "group")
		assert.Equal(t, project.Parent.ArtifactID, "artifact")
	})
}
