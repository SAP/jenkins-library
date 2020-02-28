package nexus

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddArtifactValid(t *testing.T) {
	nexusUpload := Upload{}

	err := nexusUpload.AddArtifact(ArtifactDescription{ID: "artifact.id", Classifier: "", Type: "pom", File: "pom.xml"})

	assert.NoError(t, err, "Expected to add valid artifact")
	assert.True(t, len(nexusUpload.artifacts) == 1)

	assert.True(t, nexusUpload.artifacts[0].ID == "artifact.id")
	assert.True(t, nexusUpload.artifacts[0].Classifier == "")
	assert.True(t, nexusUpload.artifacts[0].Type == "pom")
	assert.True(t, nexusUpload.artifacts[0].File == "pom.xml")
}

func TestAddArtifactMissingID(t *testing.T) {
	nexusUpload := Upload{}

	err := nexusUpload.AddArtifact(ArtifactDescription{ID: "", Classifier: "", Type: "pom", File: "pom.xml"})

	assert.Error(t, err, "Expected to fail adding invalid artifact")
	assert.True(t, len(nexusUpload.artifacts) == 0)
}

func TestAddDuplicateArtifact(t *testing.T) {
	nexusUpload := Upload{}

	err := nexusUpload.AddArtifact(ArtifactDescription{ID: "blob", Classifier: "", Type: "pom", File: "pom.xml"})
	err = nexusUpload.AddArtifact(ArtifactDescription{ID: "blob", Classifier: "", Type: "pom", File: "pom.xml"})
	assert.NoError(t, err, "Expected to succeed adding duplicate artifact")
	assert.True(t, len(nexusUpload.artifacts) == 1)
}

func TestArtifactsNotDirectlyAccessible(t *testing.T) {
	nexusUpload := Upload{}

	err := nexusUpload.AddArtifact(ArtifactDescription{ID: "artifact.id", Classifier: "", Type: "pom", File: "pom.xml"})
	assert.NoError(t, err, "Expected to succeed adding valid artifact")

	artifacts := nexusUpload.GetArtifacts()
	// Overwrite array entry in the returned array...
	artifacts[0] = ArtifactDescription{ID: "another.id", Classifier: "", Type: "pom", File: "pom.xml"}
	// ... but expect the entry in nexusUpload object to be unchanged
	assert.True(t, nexusUpload.artifacts[0].ID == "artifact.id")
}

func TestSensibleBaseURLNexus2(t *testing.T) {
	baseURL, err := getBaseURL("localhost:8081/nexus", "nexus2", "maven-releases", "some.group.id")
	assert.NoError(t, err, "Expected getBaseURL() to succeed")
	assert.Equal(t, "localhost:8081/nexus/content/repositories/maven-releases/some/group/id/", baseURL)
}

func TestSensibleBaseURLNexus3(t *testing.T) {
	baseURL, err := getBaseURL("localhost:8081", "nexus3", "maven-releases", "some.group.id")
	assert.NoError(t, err, "Expected getBaseURL() to succeed")
	assert.Equal(t, "localhost:8081/repository/maven-releases/some/group/id/", baseURL)
}

func TestSetBaseURLParamChecking(t *testing.T) {
	nexusUpload := Upload{}
	err := nexusUpload.SetBaseURL("", "nexus3", "maven-releases", "some.group.id")
	assert.Error(t, err, "Expected SetBaseURL() to fail (no host)")
	err = nexusUpload.SetBaseURL("localhost:8081", "3", "maven-releases", "some.group.id")
	assert.Error(t, err, "Expected SetBaseURL() to fail (invalid nexus version)")
	err = nexusUpload.SetBaseURL("localhost:8081", "nexus3", "", "some.group.id")
	assert.Error(t, err, "Expected SetBaseURL() to fail (no repository)")
	err = nexusUpload.SetBaseURL("localhost:8081", "nexus3", "maven-releases", "")
	assert.Error(t, err, "Expected SetBaseURL() to fail (no groupID)")
}
