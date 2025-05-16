//go:build unit

package versioning

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mavenMock struct {
	groupID    string
	artifactID string
	version    string
	packaging  string
}

func (m *mavenMock) VersioningScheme() string {
	return "maven"
}
func (m *mavenMock) GetVersion() (string, error) {
	return m.version, nil
}
func (m *mavenMock) SetVersion(v string) error {
	m.version = v
	return nil
}
func (m *mavenMock) GetCoordinates() (Coordinates, error) {
	return Coordinates{GroupID: m.groupID, ArtifactID: m.artifactID, Version: m.version, Packaging: m.packaging}, nil
}

type pipMock struct {
	artifactID string
	version    string
}

func (p *pipMock) VersioningScheme() string {
	return "pep440"
}
func (p *pipMock) GetVersion() (string, error) {
	return p.version, nil
}
func (p *pipMock) SetVersion(v string) error {
	p.version = v
	return nil
}
func (p *pipMock) GetCoordinates() (Coordinates, error) {
	return Coordinates{ArtifactID: p.artifactID, Version: p.version}, nil
}

func TestDetermineProjectCoordinatesWithCustomVersion(t *testing.T) {
	nameTemplate := `{{list .GroupID .ArtifactID | join "-" | trimAll "-"}}`

	t.Run("default", func(t *testing.T) {
		gav, _ := (&mavenMock{groupID: "com.test.pkg", artifactID: "analyzer", version: "1.2.3"}).GetCoordinates()
		name, version := DetermineProjectCoordinatesWithCustomVersion(nameTemplate, "major", "", gav)
		assert.Equal(t, "com.test.pkg-analyzer", name, "Expected different project name")
		assert.Equal(t, "1", version, "Expected different project version")
	})

	t.Run("custom", func(t *testing.T) {
		gav, _ := (&mavenMock{groupID: "com.test.pkg", artifactID: "analyzer", version: "1.2.3"}).GetCoordinates()
		_, version := DetermineProjectCoordinatesWithCustomVersion(nameTemplate, "major", "customVersion", gav)
		assert.Equal(t, "customVersion", version, "Expected different project version")
	})
}

func TestDetermineProjectCoordinates(t *testing.T) {
	nameTemplate := `{{list .GroupID .ArtifactID | join "-" | trimAll "-"}}`

	t.Run("maven", func(t *testing.T) {
		gav, _ := (&mavenMock{groupID: "com.test.pkg", artifactID: "analyzer", version: "1.2.3"}).GetCoordinates()
		name, version := DetermineProjectCoordinates(nameTemplate, "major", gav)
		assert.Equal(t, "com.test.pkg-analyzer", name, "Expected different project name")
		assert.Equal(t, "1", version, "Expected different project version")
	})

	t.Run("maven major-minor", func(t *testing.T) {
		gav, _ := (&mavenMock{groupID: "com.test.pkg", artifactID: "analyzer", version: "1.2.3"}).GetCoordinates()
		name, version := DetermineProjectCoordinates(nameTemplate, "major-minor", gav)
		assert.Equal(t, "com.test.pkg-analyzer", name, "Expected different project name")
		assert.Equal(t, "1.2", version, "Expected different project version")
	})

	t.Run("maven full", func(t *testing.T) {
		gav, _ := (&mavenMock{groupID: "com.test.pkg", artifactID: "analyzer", version: "1.2.3-7864387648746"}).GetCoordinates()
		name, version := DetermineProjectCoordinates(nameTemplate, "full", gav)
		assert.Equal(t, "com.test.pkg-analyzer", name, "Expected different project name")
		assert.Equal(t, "1.2.3-7864387648746", version, "Expected different project version")
	})

	t.Run("maven semantic", func(t *testing.T) {
		gav, _ := (&mavenMock{groupID: "com.test.pkg", artifactID: "analyzer", version: "1.2.3-7864387648746"}).GetCoordinates()
		name, version := DetermineProjectCoordinates(nameTemplate, "semantic", gav)
		assert.Equal(t, "com.test.pkg-analyzer", name, "Expected different project name")
		assert.Equal(t, "1.2.3", version, "Expected different project version")
	})

	t.Run("maven empty", func(t *testing.T) {
		gav, _ := (&mavenMock{groupID: "com.test.pkg", artifactID: "analyzer", version: "0-SNAPSHOT"}).GetCoordinates()
		name, version := DetermineProjectCoordinates(nameTemplate, "snapshot", gav)
		assert.Equal(t, "com.test.pkg-analyzer", name, "Expected different project name")
		assert.Equal(t, "", version, "Expected different project version")
	})

	t.Run("python", func(t *testing.T) {
		gav, _ := (&pipMock{artifactID: "python-test", version: "2.2.3"}).GetCoordinates()
		name, version := DetermineProjectCoordinates(nameTemplate, "major", gav)
		assert.Equal(t, "python-test", name, "Expected different project name")
		assert.Equal(t, "2", version, "Expected different project version")
	})

	t.Run("python semantic", func(t *testing.T) {
		gav, _ := (&pipMock{artifactID: "python-test", version: "2.2.3.20200101"}).GetCoordinates()
		name, version := DetermineProjectCoordinates(nameTemplate, "semantic", gav)
		assert.Equal(t, "python-test", name, "Expected different project name")
		assert.Equal(t, "2.2.3", version, "Expected different project version")
	})
}
