package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/nexus"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestMavenEvaluateGroupID(t *testing.T) {
	// This is a temporary test which should be moved into maven pkg
	// together with evaluate functionality (needs separate PR)
	evaluator := mavenExecutor{execRunner: command.Command{}}
	value, err := evaluator.evaluateProperty("../pom.xml", "project.groupId")

	assert.NoError(t, err, "expected evaluation to succeed")
	assert.Equal(t, "com.sap.cp.jenkins", value)
}

type mockProjectStructure struct {
	mta   bool
	maven bool
}

func (ps *mockProjectStructure) UsesMta() bool {
	return ps.mta
}

func (ps *mockProjectStructure) UsesMaven() bool {
	return ps.maven
}

type mockPropertyEvaluator struct {
	properties map[string]string
}

func newMockPropertyEvaluator() mockPropertyEvaluator {
	e := mockPropertyEvaluator{}
	e.properties = map[string]string{}
	return e
}

func (e *mockPropertyEvaluator) evaluateProperty(pomFile, expression string) (string, error) {
	value := e.properties[expression]
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

type mockFileUtils struct {
	files map[string][]byte
}

func newMockFileUtils() mockFileUtils {
	f := mockFileUtils{}
	f.files = map[string][]byte{}
	return f
}

func (f *mockFileUtils) FileExists(path string) (bool, error) {
	content := f.files[path]
	if content == nil {
		return false, fmt.Errorf("'%s': %w", path, os.ErrNotExist)
	}
	return true, nil
}

func (f *mockFileUtils) Copy(src, dest string) (int64, error) {
	// Not used, only needed to complete FileUtils interface
	return 42, nil
}

func (f *mockFileUtils) FileRead(path string) ([]byte, error) {
	content := f.files[path]
	if content == nil {
		return nil, fmt.Errorf("could not read '%s'", path)
	}
	return content, nil
}

func (f *mockFileUtils) FileWrite(path string, content []byte, perm os.FileMode) error {
	f.files[path] = content
	return nil
}

func (f *mockFileUtils) MkdirAll(path string, perm os.FileMode) error {
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
		projectStructure := mockProjectStructure{mta: true}
		evaluator := mockPropertyEvaluator{}
		uploader := mockUploader{}
		options := createOptions()
		fileUtils := newMockFileUtils()
		fileUtils.files["mta.yaml"] = testMtaYml
		err := runNexusUpload(&options, &uploader, &projectStructure, &fileUtils, &evaluator)
		assert.NoError(t, err, "expected mta.yaml project upload to work")
	})
	t.Run("Test uploading mta.yml project works", func(t *testing.T) {
		projectStructure := mockProjectStructure{mta: true}
		evaluator := mockPropertyEvaluator{}
		uploader := mockUploader{}
		options := createOptions()
		fileUtils := newMockFileUtils()
		fileUtils.files["mta.yml"] = testMtaYml
		err := runNexusUpload(&options, &uploader, &projectStructure, &fileUtils, &evaluator)
		assert.NoError(t, err, "expected mta.yml project upload to work")
	})
}

func TestUploadMavenProjects(t *testing.T) {
	t.Run("Test uploading Maven project with POM packaging works", func(t *testing.T) {
		projectStructure := mockProjectStructure{maven: true}
		evaluator := newMockPropertyEvaluator()
		evaluator.properties["project.version"] = "1.0"
		evaluator.properties["project.groupId"] = "com.mycompany.app"
		evaluator.properties["project.artifactId"] = "my-app"
		evaluator.properties["project.packaging"] = "pom"
		evaluator.properties["project.build.finalName"] = "my-app-1.0.jar"
		uploader := mockUploader{}
		options := createOptions()
		fileUtils := newMockFileUtils()
		fileUtils.files["pom.xml"] = testPomXml
		err := runNexusUpload(&options, &uploader, &projectStructure, &fileUtils, &evaluator)
		assert.NoError(t, err, "expected Maven upload to work")
	})
	t.Run("Test uploading Maven project with JAR packaging works", func(t *testing.T) {
		projectStructure := mockProjectStructure{maven: true}
		evaluator := newMockPropertyEvaluator()
		evaluator.properties["project.version"] = "1.0"
		evaluator.properties["project.groupId"] = "com.mycompany.app"
		evaluator.properties["project.artifactId"] = "my-app"
		evaluator.properties["project.packaging"] = "jar"
		evaluator.properties["project.build.finalName"] = "my-app-1.0.jar"
		uploader := mockUploader{}
		options := createOptions()
		fileUtils := newMockFileUtils()
		fileUtils.files["pom.xml"] = testPomXml
		fileUtils.files["target/my-app-1.0.jar"] = []byte("contentsOfJar")
		err := runNexusUpload(&options, &uploader, &projectStructure, &fileUtils, &evaluator)
		assert.NoError(t, err, "expected Maven upload to work")
	})
}

func TestUploadUnknownProjectFails(t *testing.T) {
	projectStructure := mockProjectStructure{}
	evaluator := newMockPropertyEvaluator()
	uploader := mockUploader{}
	options := createOptions()
	fileUtils := newMockFileUtils()
	err := runNexusUpload(&options, &uploader, &projectStructure, &fileUtils, &evaluator)
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
	return &client, addAdditionalClassifierArtifacts(additionalClassifiers, "some folder", "artifact-id", &client)
}
