package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/nexus"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestMavenEvaluateGroupID(t *testing.T) {
	// This is a temporary test which should be moved into maven pkg
	// together with evaluate functionality (needs separate PR)
	utils := newUtilsBundle()
	value, err := utils.evaluateProperty("../pom.xml", "project.groupId")

	assert.NoError(t, err, "expected evaluation to succeed")
	assert.Equal(t, "com.sap.cp.jenkins", value)
}

type mockUtilsBundle struct {
	mta        bool
	maven      bool
	files      map[string][]byte
	properties map[string]string
	cpe        map[string]string
}

func newMockUtilsBundle(usesMta, usesMaven bool) mockUtilsBundle {
	utils := mockUtilsBundle{mta: usesMta, maven: usesMaven}
	utils.files = map[string][]byte{}
	utils.properties = map[string]string{}
	utils.cpe = map[string]string{}
	return utils
}

func (m *mockUtilsBundle) usesMta() bool {
	return m.mta
}

func (m *mockUtilsBundle) usesMaven() bool {
	return m.maven
}

func (m *mockUtilsBundle) fileExists(path string) (bool, error) {
	content := m.files[path]
	if content == nil {
		return false, fmt.Errorf("'%s': %w", path, os.ErrNotExist)
	}
	return true, nil
}

func (m *mockUtilsBundle) fileRead(path string) ([]byte, error) {
	content := m.files[path]
	if content == nil {
		return nil, fmt.Errorf("could not read '%s'", path)
	}
	return content, nil
}

func (m *mockUtilsBundle) getEnvParameter(path, name string) string {
	path = path + "/" + name
	return m.cpe[path]
}

func (m *mockUtilsBundle) evaluateProperty(pomFile, expression string) (string, error) {
	value := m.properties[expression]
	if value == "" {
		return "", fmt.Errorf("property '%s' not found in '%s'", expression, pomFile)
	}
	return value, nil
}

type mockUploader struct {
	baseURL   string
	version   string
	artifacts []nexus.ArtifactDescription
}

func (u *mockUploader) SetBaseURL(nexusURL, nexusVersion, repository, groupID string) error {
	u.baseURL = "http://" + nexusURL + "/nexus/repositories/" + repository
	u.baseURL += "/" + strings.ReplaceAll(groupID, ".", "/")
	return nil
}

func (u *mockUploader) SetArtifactsVersion(version string) error {
	u.version = version
	return nil
}

func (u *mockUploader) AddArtifact(artifact nexus.ArtifactDescription) error {
	u.artifacts = append(u.artifacts, artifact)
	return nil
}

func (u *mockUploader) GetArtifacts() []nexus.ArtifactDescription {
	return u.artifacts
}

func (u *mockUploader) UploadArtifacts() error {
	return nil
}

func createOptions() nexusUploadOptions {
	return nexusUploadOptions{
		Repository: "maven-releases",
		GroupID:    "my.group.id",
		ArtifactID: "artifact.id",
		Version:    "nexus3",
		Url:        "localhost:8081",
	}
}

var testMtaYml = []byte(`
_schema-version: 2.1.0
ID: test
version: 0.3.0

modules:

- name: java
  type: java
  path: srv
`)

var testPomXml = []byte(`
<project>
  <modelVersion>4.0.0</modelVersion>
  <groupId>com.mycompany.app</groupId>
  <artifactId>my-app</artifactId>
  <version>1.0</version>
</project>
`)

func TestUploadMTAProjects(t *testing.T) {
	t.Run("Test uploading mta.yaml project works", func(t *testing.T) {
		utils := newMockUtilsBundle(true, false)
		utils.files["mta.yaml"] = testMtaYml
		utils.cpe[".pipeline/commonPipelineEnvironment/mtarFilePath"] = "test.mtar"
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(&utils, &uploader, &options)
		assert.NoError(t, err, "expected mta.yaml project upload to work")

		assert.Equal(t, 2, len(uploader.artifacts))

		assert.Equal(t, "mta.yaml", uploader.artifacts[0].File)
		assert.Equal(t, "yaml", uploader.artifacts[0].Type)
		assert.Equal(t, "artifact.id", uploader.artifacts[0].ID)

		assert.Equal(t, "test.mtar", uploader.artifacts[1].File)
		assert.Equal(t, "mtar", uploader.artifacts[1].Type)
		assert.Equal(t, "artifact.id", uploader.artifacts[1].ID)
	})
	t.Run("Test uploading mta.yml project works", func(t *testing.T) {
		utils := newMockUtilsBundle(true, false)
		utils.files["mta.yml"] = testMtaYml
		utils.cpe[".pipeline/commonPipelineEnvironment/mtarFilePath"] = "test.mtar"
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(&utils, &uploader, &options)
		assert.NoError(t, err, "expected mta.yml project upload to work")

		assert.Equal(t, 2, len(uploader.artifacts))

		assert.Equal(t, "mta.yml", uploader.artifacts[0].File)
		assert.Equal(t, "yaml", uploader.artifacts[0].Type)
		assert.Equal(t, "artifact.id", uploader.artifacts[0].ID)

		assert.Equal(t, "test.mtar", uploader.artifacts[1].File)
		assert.Equal(t, "mtar", uploader.artifacts[1].Type)
		assert.Equal(t, "artifact.id", uploader.artifacts[1].ID)
	})
	t.Run("Test uploading mta.yml project works with artifactID from CPE", func(t *testing.T) {
		utils := newMockUtilsBundle(true, false)
		utils.files["mta.yml"] = testMtaYml
		utils.cpe[".pipeline/commonPipelineEnvironment/mtarFilePath"] = "test.mtar"
		utils.cpe[".pipeline/commonPipelineEnvironment/configuration/artifactId"] = "my-artifact-id"
		uploader := mockUploader{}
		options := createOptions()
		// Clear artifact ID to trigger reading it from the CPE
		options.ArtifactID = ""

		err := runNexusUpload(&utils, &uploader, &options)
		assert.NoError(t, err, "expected mta.yml project upload to work")

		assert.Equal(t, 2, len(uploader.artifacts))

		assert.Equal(t, "mta.yml", uploader.artifacts[0].File)
		assert.Equal(t, "yaml", uploader.artifacts[0].Type)
		assert.Equal(t, "my-artifact-id", uploader.artifacts[0].ID)

		assert.Equal(t, "test.mtar", uploader.artifacts[1].File)
		assert.Equal(t, "mtar", uploader.artifacts[1].Type)
		assert.Equal(t, "my-artifact-id", uploader.artifacts[1].ID)
	})
}

func TestUploadMavenProjects(t *testing.T) {
	t.Run("Test uploading Maven project with POM packaging works", func(t *testing.T) {
		utils := newMockUtilsBundle(false, true)
		utils.properties["project.version"] = "1.0"
		utils.properties["project.groupId"] = "com.mycompany.app"
		utils.properties["project.artifactId"] = "my-app"
		utils.properties["project.packaging"] = "pom"
		utils.properties["project.build.finalName"] = "my-app-1.0"
		utils.files["pom.xml"] = testPomXml
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(&utils, &uploader, &options)
		assert.NoError(t, err, "expected Maven upload to work")

		assert.Equal(t, 1, len(uploader.artifacts))

		assert.Equal(t, "pom.xml", uploader.artifacts[0].File)
		assert.Equal(t, "my-app", uploader.artifacts[0].ID)
		assert.Equal(t, "pom", uploader.artifacts[0].Type)
	})
	t.Run("Test uploading Maven project with JAR packaging works", func(t *testing.T) {
		utils := newMockUtilsBundle(false, true)
		utils.properties["project.version"] = "1.0"
		utils.properties["project.groupId"] = "com.mycompany.app"
		utils.properties["project.artifactId"] = "my-app"
		utils.properties["project.packaging"] = "jar"
		utils.properties["project.build.finalName"] = "my-app-1.0"
		utils.files["pom.xml"] = testPomXml
		utils.files["target/my-app-1.0.jar"] = []byte("contentsOfJar")
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(&utils, &uploader, &options)
		assert.NoError(t, err, "expected Maven upload to work")

		assert.Equal(t, 2, len(uploader.artifacts))

		assert.Equal(t, "pom.xml", uploader.artifacts[0].File)
		assert.Equal(t, "my-app", uploader.artifacts[0].ID)
		assert.Equal(t, "pom", uploader.artifacts[0].Type)

		assert.Equal(t, "target/my-app-1.0.jar", uploader.artifacts[1].File)
		assert.Equal(t, "my-app", uploader.artifacts[1].ID)
		assert.Equal(t, "jar", uploader.artifacts[1].Type)
	})
}

func TestUploadUnknownProjectFails(t *testing.T) {
	utils := newMockUtilsBundle(false, false)
	uploader := mockUploader{}
	options := createOptions()

	err := runNexusUpload(&utils, &uploader, &options)
	assert.Error(t, err, "expected upload of unknown project structure to fail")
}

func TestAdditionalClassifierEmpty(t *testing.T) {
	t.Run("Empty additional classifiers", func(t *testing.T) {
		client, err := testAdditionalClassifierArtifacts("")
		assert.NoError(t, err, "expected empty additional classifiers to succeed")
		assert.True(t, len(client.GetArtifacts()) == 0)
	})
	t.Run("Additional classifiers is invalid JSON", func(t *testing.T) {
		client, err := testAdditionalClassifierArtifacts("some random string")
		assert.Error(t, err, "expected invalid additional classifiers to fail")
		assert.True(t, len(client.GetArtifacts()) == 0)
	})
	t.Run("Classifiers valid but wrong JSON", func(t *testing.T) {
		json := `
		[
			{
				"classifier" : "source",
				"type"       : "jar"
			},
			{}
		]
	`
		client, err := testAdditionalClassifierArtifacts(json)
		assert.Error(t, err, "expected invalid additional classifiers to fail")
		assert.True(t, len(client.GetArtifacts()) == 1)
	})
	t.Run("Additional classifiers is valid JSON", func(t *testing.T) {
		json := `
		[
			{
				"classifier" : "source",
				"type"       : "jar"
			},
			{
				"classifier" : "classes",
				"type"       : "jar"
			}
		]
	`
		client, err := testAdditionalClassifierArtifacts(json)
		assert.NoError(t, err, "expected valid additional classifiers to succeed")
		assert.True(t, len(client.GetArtifacts()) == 2)
	})
}

func testAdditionalClassifierArtifacts(additionalClassifiers string) (*nexus.Upload, error) {
	client := nexus.Upload{}
	return &client, addAdditionalClassifierArtifacts(&client, additionalClassifiers, "some folder", "artifact-id")
}
