package cmd

import (
	"github.com/SAP/jenkins-library/pkg/nexus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMavenEvaluateGroupID(t *testing.T) {
	// This is just a temporary test to facilitate debugging
	value, err := evaluateMavenProperty("../pom.xml", "project.groupId")

	assert.NoError(t, err, "expected evaluation to succeed")
	assert.Equal(t, "com.sap.cp.jenkins", value)
}

func TestAdditionalClassifierEmpty(t *testing.T) {
	client, err := testAdditionalClassifierArtifacts("")
	assert.NoError(t, err, "expected empty additional classifiers to succeed")
	assert.True(t, len(client.GetArtifacts()) == 0)
}

func TestAdditionalClassifierInvalidJSON(t *testing.T) {
	client, err := testAdditionalClassifierArtifacts("some random string")
	assert.Error(t, err, "expected invalid additional classifiers to fail")
	assert.True(t, len(client.GetArtifacts()) == 0)
}

func TestAdditionalClassifierValidButWrongJSON(t *testing.T) {
	client, err := testAdditionalClassifierArtifacts("[{\"classifier\":\"source\",\"type\":\"jar\"},{}]")
	assert.Error(t, err, "expected invalid additional classifiers to fail")
	assert.True(t, len(client.GetArtifacts()) == 1)
}

func TestAdditionalClassifierValidJSON(t *testing.T) {
	client, err := testAdditionalClassifierArtifacts("[{\"classifier\":\"source\",\"type\":\"jar\"},{\"classifier\":\"classes\",\"type\":\"jar\"}]")
	assert.NoError(t, err, "expected valid additional classifiers to succeed")
	assert.True(t, len(client.GetArtifacts()) == 2)
}

func testAdditionalClassifierArtifacts(additionalClassifiers string) (*nexus.Upload, error) {
	client := nexus.Upload{}
	return &client, addAdditionalClassifierArtifacts(additionalClassifiers, "some folder", "artifact-id", &client)
}
