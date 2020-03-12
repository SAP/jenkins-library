package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/nexus"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

/*
func TestDeployMTA(t *testing.T) {
	err := os.Chdir("../integration/testdata/TestNexusIntegration/mta")
	assert.NoError(t, err)
	options := nexusUploadOptions{
		Url:        "localhost:8081",
		GroupID:    "nexus.upload",
		Repository: "maven-releases",
		ArtifactID: "my.mta.project",
		Version:    "nexus3",
	}
	nexusUpload(options, nil)
}

func TestDeployGettingStartedBookshot(t *testing.T) {
	err := os.Chdir("../../GettingStartedBookshop")
	assert.NoError(t, err)
	options := nexusUploadOptions{
		Url:        "localhost:8081",
		GroupID:    "nexus.upload",
		Repository: "maven-releases",
		ArtifactID: "GettingStartedBookshop",
		Version:    "nexus3",
	}
	nexusUpload(options, nil)
}

func TestDeployMaven(t *testing.T) {
	err := os.Chdir("../integration/testdata/TestNexusIntegration/maven")
	assert.NoError(t, err)
	options := nexusUploadOptions{
		Url:        "localhost:8081",
		Repository: "maven-releases",
		Version:    "nexus3",
	}
	nexusUpload(options, nil)
}
*/

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
	if value == "<empty>" {
		return "", nil
	}
	if value == "" {
		return "", fmt.Errorf("property '%s' not found in '%s'", expression, pomFile)
	}
	return value, nil
}

func (m *mockUtilsBundle) getExecRunner() execRunner {
	mockExecRunner := mock.ExecMockRunner{}
	return &mockExecRunner
}

type mockUploader struct {
	nexus.Upload
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
	t.Run("Uploading MTA project without group id parameter fails", func(t *testing.T) {
		utils := newMockUtilsBundle(true, false)
		utils.files["mta.yaml"] = testMtaYml
		utils.cpe[".pipeline/commonPipelineEnvironment/mtarFilePath"] = "test.mtar"
		uploader := mockUploader{}
		options := createOptions()
		options.GroupID = ""

		err := runNexusUpload(&utils, &uploader, &options)
		assert.EqualError(t, err, "the 'groupId' parameter needs to be provided for MTA projects")
		assert.Equal(t, 0, len(uploader.GetArtifacts()))
	})
	t.Run("Uploading MTA project fails due to missing yaml file", func(t *testing.T) {
		utils := newMockUtilsBundle(true, false)
		utils.cpe[".pipeline/commonPipelineEnvironment/mtarFilePath"] = "test.mtar"
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(&utils, &uploader, &options)
		assert.EqualError(t, err, "could not read 'mta.yml'")
		assert.Equal(t, 0, len(uploader.GetArtifacts()))
	})
	t.Run("Uploading MTA project fails due to garbage YAML content", func(t *testing.T) {
		utils := newMockUtilsBundle(true, false)
		utils.files["mta.yaml"] = []byte("garbage")
		utils.cpe[".pipeline/commonPipelineEnvironment/mtarFilePath"] = "test.mtar"
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(&utils, &uploader, &options)
		assert.EqualError(t, err,
			"error unmarshaling JSON: json: cannot unmarshal string into Go value of type cmd.mtaYaml")
		assert.Equal(t, 0, len(uploader.GetArtifacts()))
	})
	t.Run("Test uploading mta.yaml project fails due to missing mtar file", func(t *testing.T) {
		utils := newMockUtilsBundle(true, false)
		utils.files["mta.yaml"] = testMtaYml
		utils.cpe[".pipeline/commonPipelineEnvironment/mtarFilePath"] = "test.mtar"
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(&utils, &uploader, &options)
		assert.EqualError(t, err, "artifact file not found 'test.mtar'")

		assert.Equal(t, "0.3.0", uploader.GetArtifactsVersion())

		artifacts := uploader.GetArtifacts()
		if assert.Equal(t, 1, len(artifacts)) {
			assert.Equal(t, "mta.yaml", artifacts[0].File)
			assert.Equal(t, "yaml", artifacts[0].Type)
			assert.Equal(t, "artifact.id", artifacts[0].ID)
		}
	})
	t.Run("Test uploading mta.yaml project works", func(t *testing.T) {
		utils := newMockUtilsBundle(true, false)
		utils.files["mta.yaml"] = testMtaYml
		utils.files["test.mtar"] = []byte("contentsOfMtar")
		utils.cpe[".pipeline/commonPipelineEnvironment/mtarFilePath"] = "test.mtar"
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(&utils, &uploader, &options)
		assert.NoError(t, err, "expected mta.yaml project upload to work")

		assert.Equal(t, "0.3.0", uploader.GetArtifactsVersion())

		artifacts := uploader.GetArtifacts()
		if assert.Equal(t, 2, len(artifacts)) {
			assert.Equal(t, "mta.yaml", artifacts[0].File)
			assert.Equal(t, "yaml", artifacts[0].Type)
			assert.Equal(t, "artifact.id", artifacts[0].ID)

			assert.Equal(t, "test.mtar", artifacts[1].File)
			assert.Equal(t, "mtar", artifacts[1].Type)
			assert.Equal(t, "artifact.id", artifacts[1].ID)
		}
	})
	t.Run("Test uploading mta.yml project works", func(t *testing.T) {
		utils := newMockUtilsBundle(true, false)
		utils.files["mta.yml"] = testMtaYml
		utils.files["test.mtar"] = []byte("contentsOfMtar")
		utils.cpe[".pipeline/commonPipelineEnvironment/mtarFilePath"] = "test.mtar"
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(&utils, &uploader, &options)
		assert.NoError(t, err, "expected mta.yml project upload to work")

		assert.Equal(t, "0.3.0", uploader.GetArtifactsVersion())

		artifacts := uploader.GetArtifacts()
		if assert.Equal(t, 2, len(artifacts)) {
			assert.Equal(t, "mta.yml", artifacts[0].File)
			assert.Equal(t, "yaml", artifacts[0].Type)
			assert.Equal(t, "artifact.id", artifacts[0].ID)

			assert.Equal(t, "test.mtar", artifacts[1].File)
			assert.Equal(t, "mtar", artifacts[1].Type)
			assert.Equal(t, "artifact.id", artifacts[1].ID)
		}
	})
	t.Run("Test uploading mta.yml project works with artifactID from CPE", func(t *testing.T) {
		utils := newMockUtilsBundle(true, false)
		utils.files["mta.yml"] = testMtaYml
		utils.files["test.mtar"] = []byte("contentsOfMtar")
		utils.cpe[".pipeline/commonPipelineEnvironment/mtarFilePath"] = "test.mtar"
		utils.cpe[".pipeline/commonPipelineEnvironment/configuration/artifactId"] = "my-artifact-id"
		uploader := mockUploader{}
		options := createOptions()
		// Clear artifact ID to trigger reading it from the CPE
		options.ArtifactID = ""

		err := runNexusUpload(&utils, &uploader, &options)
		assert.NoError(t, err, "expected mta.yml project upload to work")

		artifacts := uploader.GetArtifacts()
		if assert.Equal(t, 2, len(artifacts)) {
			assert.Equal(t, "mta.yml", artifacts[0].File)
			assert.Equal(t, "yaml", artifacts[0].Type)
			assert.Equal(t, "my-artifact-id", artifacts[0].ID)

			assert.Equal(t, "test.mtar", artifacts[1].File)
			assert.Equal(t, "mtar", artifacts[1].Type)
			assert.Equal(t, "my-artifact-id", artifacts[1].ID)
		}
	})
}

func TestUploadMavenProjects(t *testing.T) {
	t.Run("Uploading Maven project fails due to missing pom.xml", func(t *testing.T) {
		utils := newMockUtilsBundle(false, true)
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(&utils, &uploader, &options)
		assert.EqualError(t, err, "pom.xml not found")
		assert.Equal(t, 0, len(uploader.GetArtifacts()))
	})
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

		artifacts := uploader.GetArtifacts()
		if assert.Equal(t, 1, len(artifacts)) {
			assert.Equal(t, "pom.xml", artifacts[0].File)
			assert.Equal(t, "my-app", artifacts[0].ID)
			assert.Equal(t, "pom", artifacts[0].Type)
		}
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

		artifacts := uploader.GetArtifacts()
		if assert.Equal(t, 2, len(artifacts)) {
			assert.Equal(t, "pom.xml", artifacts[0].File)
			assert.Equal(t, "my-app", artifacts[0].ID)
			assert.Equal(t, "pom", artifacts[0].Type)

			assert.Equal(t, "target/my-app-1.0.jar", artifacts[1].File)
			assert.Equal(t, "my-app", artifacts[1].ID)
			assert.Equal(t, "jar", artifacts[1].Type)
		}
	})
	t.Run("Test uploading Maven project with fall-back to JAR packaging works", func(t *testing.T) {
		utils := newMockUtilsBundle(false, true)
		utils.properties["project.version"] = "1.0"
		utils.properties["project.groupId"] = "com.mycompany.app"
		utils.properties["project.artifactId"] = "my-app"
		utils.properties["project.packaging"] = "<empty>"
		utils.properties["project.build.finalName"] = "my-app-1.0"
		utils.files["pom.xml"] = testPomXml
		utils.files["target/my-app-1.0.jar"] = []byte("contentsOfJar")
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(&utils, &uploader, &options)
		assert.NoError(t, err, "expected Maven upload to work")

		artifacts := uploader.GetArtifacts()
		if assert.Equal(t, 2, len(artifacts)) {
			assert.Equal(t, "pom.xml", artifacts[0].File)
			assert.Equal(t, "my-app", artifacts[0].ID)
			assert.Equal(t, "pom", artifacts[0].Type)

			assert.Equal(t, "target/my-app-1.0.jar", artifacts[1].File)
			assert.Equal(t, "my-app", artifacts[1].ID)
			assert.Equal(t, "jar", artifacts[1].Type)
		}
	})
	t.Run("Test uploading Maven project with fall-back to group id from parameters works", func(t *testing.T) {
		utils := newMockUtilsBundle(false, true)
		utils.properties["project.version"] = "1.0"
		utils.properties["project.artifactId"] = "my-app"
		utils.properties["project.packaging"] = "pom"
		utils.properties["project.build.finalName"] = "my-app-1.0"
		utils.files["pom.xml"] = testPomXml
		uploader := mockUploader{}
		options := createOptions()
		options.GroupID = "awesome.group"

		err := runNexusUpload(&utils, &uploader, &options)
		assert.NoError(t, err, "expected Maven upload to work")

		assert.Equal(t, "http://localhost:8081/nexus/repositories/maven-releases/awesome/group",
			uploader.GetBaseURL())

		artifacts := uploader.GetArtifacts()
		if assert.Equal(t, 1, len(artifacts)) {
			assert.Equal(t, "pom.xml", artifacts[0].File)
			assert.Equal(t, "my-app", artifacts[0].ID)
			assert.Equal(t, "pom", artifacts[0].Type)
		}
	})
	t.Run("Test uploading Maven project with JAR packaging fails without finalName", func(t *testing.T) {
		utils := newMockUtilsBundle(false, true)
		utils.properties["project.version"] = "1.0"
		utils.properties["project.groupId"] = "com.mycompany.app"
		utils.properties["project.artifactId"] = "my-app"
		utils.properties["project.packaging"] = "jar"
		utils.files["pom.xml"] = testPomXml
		utils.files["target/my-app-1.0.jar"] = []byte("contentsOfJar")
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(&utils, &uploader, &options)
		assert.EqualError(t, err, "property 'project.build.finalName' not found in 'pom.xml'")

		artifacts := uploader.GetArtifacts()
		if assert.Equal(t, 1, len(artifacts)) {
			assert.Equal(t, "pom.xml", artifacts[0].File)
			assert.Equal(t, "my-app", artifacts[0].ID)
			assert.Equal(t, "pom", artifacts[0].Type)
		}
	})
	t.Run("Test uploading Maven project with application module works", func(t *testing.T) {
		utils := newMockUtilsBundle(false, true)
		utils.properties["project.version"] = "1.0"
		utils.properties["project.groupId"] = "com.mycompany.app"
		utils.properties["project.artifactId"] = "my-app"
		utils.properties["project.packaging"] = "jar"
		utils.properties["project.build.finalName"] = "my-app-1.0"
		utils.files["pom.xml"] = testPomXml
		utils.files["application/pom.xml"] = testPomXml
		utils.files["target/my-app-1.0.jar"] = []byte("contentsOfJar")
		utils.files["application/target/my-app-1.0.jar"] = []byte("contentsOfJar")
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(&utils, &uploader, &options)
		assert.NoError(t, err, "expected upload of maven project with application module to succeed")

		artifacts := uploader.GetArtifacts()
		if assert.Equal(t, 4, len(artifacts)) {
			assert.Equal(t, "pom.xml", artifacts[0].File)
			assert.Equal(t, "my-app", artifacts[0].ID)
			assert.Equal(t, "pom", artifacts[0].Type)

			assert.Equal(t, "target/my-app-1.0.jar", artifacts[1].File)
			assert.Equal(t, "my-app", artifacts[1].ID)
			assert.Equal(t, "jar", artifacts[1].Type)

			assert.Equal(t, "application/pom.xml", artifacts[2].File)
			assert.Equal(t, "my-app", artifacts[2].ID)
			assert.Equal(t, "pom", artifacts[2].Type)

			assert.Equal(t, "application/target/my-app-1.0.jar", artifacts[3].File)
			assert.Equal(t, "my-app", artifacts[3].ID)
			assert.Equal(t, "jar", artifacts[3].Type)
		}
	})
	t.Run("Test uploading Maven project fails without packaging", func(t *testing.T) {
		utils := newMockUtilsBundle(false, true)
		utils.properties["project.version"] = "1.0"
		utils.properties["project.groupId"] = "com.mycompany.app"
		utils.properties["project.artifactId"] = "my-app"
		utils.files["pom.xml"] = testPomXml
		utils.files["target/my-app-1.0.jar"] = []byte("contentsOfJar")
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(&utils, &uploader, &options)
		assert.EqualError(t, err, "property 'project.packaging' not found in 'pom.xml'")

		artifacts := uploader.GetArtifacts()
		if assert.Equal(t, 1, len(artifacts)) {
			assert.Equal(t, "pom.xml", artifacts[0].File)
			assert.Equal(t, "my-app", artifacts[0].ID)
			assert.Equal(t, "pom", artifacts[0].Type)
		}
	})
}

func TestUploadUnknownProjectFails(t *testing.T) {
	utils := newMockUtilsBundle(false, false)
	uploader := mockUploader{}
	options := createOptions()

	err := runNexusUpload(&utils, &uploader, &options)
	assert.Error(t, err, "expected upload of unknown project structure to fail")
}
