package nexus

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddArtifact(t *testing.T) {
	t.Run("Test valid artifact", func(t *testing.T) {
		nexusUpload := Upload{}

		err := nexusUpload.AddArtifact(ArtifactDescription{
			Classifier: "",
			Type:       "pom",
			File:       "pom.xml",
		})

		assert.NoError(t, err, "Expected to add valid artifact")
		assert.True(t, len(nexusUpload.artifacts) == 1)

		assert.True(t, nexusUpload.artifacts[0].Classifier == "")
		assert.True(t, nexusUpload.artifacts[0].Type == "pom")
		assert.True(t, nexusUpload.artifacts[0].File == "pom.xml")
	})
	t.Run("Test missing type", func(t *testing.T) {
		nexusUpload := Upload{}

		err := nexusUpload.AddArtifact(ArtifactDescription{
			Classifier: "",
			Type:       "",
			File:       "pom.xml",
		})

		assert.Error(t, err, "Expected to fail adding invalid artifact")
		assert.True(t, len(nexusUpload.artifacts) == 0)
	})
	t.Run("Test missing file", func(t *testing.T) {
		nexusUpload := Upload{}

		err := nexusUpload.AddArtifact(ArtifactDescription{
			Classifier: "",
			Type:       "pom",
			File:       "",
		})

		assert.Error(t, err, "Expected to fail adding invalid artifact")
		assert.True(t, len(nexusUpload.artifacts) == 0)
	})
	t.Run("Test adding duplicate artifact is ignored", func(t *testing.T) {
		nexusUpload := Upload{}

		_ = nexusUpload.AddArtifact(ArtifactDescription{
			Classifier: "",
			Type:       "pom",
			File:       "pom.xml",
		})
		err := nexusUpload.AddArtifact(ArtifactDescription{
			Classifier: "",
			Type:       "pom",
			File:       "pom.xml",
		})
		assert.NoError(t, err, "Expected to succeed adding duplicate artifact")
		assert.True(t, len(nexusUpload.artifacts) == 1)
	})
}

func TestGetArtifacts(t *testing.T) {
	nexusUpload := Upload{}

	err := nexusUpload.AddArtifact(ArtifactDescription{
		Classifier: "",
		Type:       "pom",
		File:       "pom.xml",
	})
	assert.NoError(t, err, "Expected to succeed adding valid artifact")

	artifacts := nexusUpload.GetArtifacts()
	// Overwrite array entry in the returned array...
	artifacts[0] = ArtifactDescription{
		Classifier: "",
		Type:       "jar",
		File:       "app.jar",
	}
	// ... but expect the entry in nexusUpload object to be unchanged
	assert.Equal(t, "pom", nexusUpload.artifacts[0].Type)
	assert.Equal(t, "pom.xml", nexusUpload.artifacts[0].File)
}

func TestGetBaseURL(t *testing.T) {
	// Invalid parameters to getBaseURL() already tested via SetRepoURL() tests
	t.Run("Test base URL for nexus2 is sensible", func(t *testing.T) {
		baseURL, err := getBaseURL("localhost:8081/nexus", "nexus2", "maven-releases")
		assert.NoError(t, err, "Expected getBaseURL() to succeed")
		assert.Equal(t, "localhost:8081/nexus/content/repositories/maven-releases/", baseURL)
	})
	t.Run("Test base URL for nexus3 is sensible", func(t *testing.T) {
		baseURL, err := getBaseURL("localhost:8081", "nexus3", "maven-releases")
		assert.NoError(t, err, "Expected getBaseURL() to succeed")
		assert.Equal(t, "localhost:8081/repository/maven-releases/", baseURL)
	})
}

func TestSetBaseURL(t *testing.T) {
	t.Run("Test no host provided", func(t *testing.T) {
		nexusUpload := Upload{}
		err := nexusUpload.SetRepoURL("", "nexus3", "maven-releases", "npm-repo")
		assert.Error(t, err, "Expected SetRepoURL() to fail (no host)")
	})
	t.Run("Test host wrongly includes protocol http://", func(t *testing.T) {
		nexusUpload := Upload{}
		err := nexusUpload.SetRepoURL("htTp://localhost:8081", "nexus3", "maven-releases", "npm-repo")
		if assert.NoError(t, err, "Expected SetRepoURL() to work") {
			assert.Equal(t, "localhost:8081/repository/maven-releases/", nexusUpload.mavenRepoURL)
		}
	})
	t.Run("Test host wrongly includes protocol https://", func(t *testing.T) {
		nexusUpload := Upload{}
		err := nexusUpload.SetRepoURL("htTpS://localhost:8081", "nexus3", "maven-releases", "npm-repo")
		if assert.NoError(t, err, "Expected SetRepoURL() to work") {
			assert.Equal(t, "localhost:8081/repository/maven-releases/", nexusUpload.mavenRepoURL)
		}
	})
	t.Run("Test invalid version provided", func(t *testing.T) {
		nexusUpload := Upload{}
		err := nexusUpload.SetRepoURL("localhost:8081", "3", "maven-releases", "npm-repo")
		assert.Error(t, err, "Expected SetRepoURL() to fail (invalid nexus version)")
	})
	t.Run("Test no nexus version provided", func(t *testing.T) {
		nexusUpload := Upload{}
		err := nexusUpload.SetRepoURL("localhost:8081", "nexus1", "maven-releases", "npm-repo")
		assert.Error(t, err, "Expected SetRepoURL() to fail (unsupported nexus version)")
	})
	t.Run("Test unsupported nexus version provided", func(t *testing.T) {
		nexusUpload := Upload{}
		err := nexusUpload.SetRepoURL("localhost:8081", "nexus1", "maven-releases", "npm-repo")
		assert.Error(t, err, "Expected SetRepoURL() to fail (unsupported nexus version)")
	})
}

func TestSetInfo(t *testing.T) {
	t.Run("Test invalid artifact version", func(t *testing.T) {
		nexusUpload := Upload{}
		err := nexusUpload.SetInfo("my.group", "artifact.id", "")
		assert.Error(t, err, "Expected SetInfo() to fail (empty version)")
		assert.Equal(t, "", nexusUpload.groupID)
		assert.Equal(t, "", nexusUpload.artifactID)
		assert.Equal(t, "", nexusUpload.version)
	})
	t.Run("Test valid artifact version", func(t *testing.T) {
		nexusUpload := Upload{}
		err := nexusUpload.SetInfo("my.group", "artifact.id", "1.0.0-SNAPSHOT")
		assert.NoError(t, err, "Expected SetInfo() to succeed")
	})
	t.Run("Test empty artifactID", func(t *testing.T) {
		nexusUpload := Upload{}
		err := nexusUpload.SetInfo("my.group", "", "1.0")
		assert.Error(t, err, "Expected to fail setting empty artifactID")
		assert.Equal(t, "", nexusUpload.groupID)
		assert.Equal(t, "", nexusUpload.artifactID)
		assert.Equal(t, "", nexusUpload.version)
	})
	t.Run("Test empty groupID", func(t *testing.T) {
		nexusUpload := Upload{}
		err := nexusUpload.SetInfo("", "id", "1.0")
		assert.Error(t, err, "Expected to fail setting empty groupID")
		assert.Equal(t, "", nexusUpload.groupID)
		assert.Equal(t, "", nexusUpload.artifactID)
		assert.Equal(t, "", nexusUpload.version)
	})
	t.Run("Test invalid ID", func(t *testing.T) {
		nexusUpload := Upload{}
		err := nexusUpload.SetInfo("my.group", "artifact/id", "1.0.0-SNAPSHOT")
		assert.Error(t, err, "Expected to fail adding invalid artifact")
		assert.Equal(t, "", nexusUpload.groupID)
		assert.Equal(t, "", nexusUpload.artifactID)
		assert.Equal(t, "", nexusUpload.version)
	})
}

func TestClear(t *testing.T) {
	nexusUpload := Upload{}

	err := nexusUpload.AddArtifact(ArtifactDescription{
		Classifier: "",
		Type:       "pom",
		File:       "pom.xml",
	})
	assert.NoError(t, err, "Expected to succeed adding valid artifact")
	assert.Equal(t, 1, len(nexusUpload.GetArtifacts()))

	nexusUpload.Clear()

	assert.Equal(t, 0, len(nexusUpload.GetArtifacts()))
}
