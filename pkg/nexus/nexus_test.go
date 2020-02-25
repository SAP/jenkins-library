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

func TestArtifactsNotDirectlyAccessible(t *testing.T) {
	nexusUpload := Upload{}

	nexusUpload.AddArtifact(ArtifactDescription{ID: "artifact.id", Classifier: "", Type: "pom", File: "pom.xml"})

	artifacts := nexusUpload.GetArtifacts()
	// Overwrite array entry in the returned array...
	artifacts[0] = ArtifactDescription{ID: "another.id", Classifier: "", Type: "pom", File: "pom.xml"}
	// ... but expect the entry in nexusUpload object to be unchanged
	assert.True(t, nexusUpload.artifacts[0].ID == "artifact.id")
}
