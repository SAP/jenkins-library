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
	cpe        map[string]string
	execRunner mock.ExecMockRunner
}

func newMockUtilsBundle(usesMta, usesMaven bool) mockUtilsBundle {
	utils := mockUtilsBundle{mta: usesMta, maven: usesMaven}
	utils.files = map[string][]byte{}
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

func (m *mockUtilsBundle) getExecRunner() execRunner {
	return &m.execRunner
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

var testMtaYmlNoVersion = []byte(`
_schema-version: 2.1.0
ID: test

modules:
- name: java
  type: java
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
	t.Run("Uploading MTA project without groupId parameter fails", func(t *testing.T) {
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
	t.Run("Uploading MTA project without artifactId parameter fails", func(t *testing.T) {
		utils := newMockUtilsBundle(true, false)
		utils.files["mta.yaml"] = testMtaYml
		utils.cpe[".pipeline/commonPipelineEnvironment/mtarFilePath"] = "test.mtar"
		uploader := mockUploader{}
		options := createOptions()
		options.ArtifactID = ""

		err := runNexusUpload(&utils, &uploader, &options)
		assert.EqualError(t, err, "the 'artifactId' parameter was not provided and could not be retrieved from the Common Pipeline Environment")
		assert.Equal(t, 0, len(uploader.GetArtifacts()))
	})
	t.Run("Uploading MTA project fails due to missing yaml file", func(t *testing.T) {
		utils := newMockUtilsBundle(true, false)
		utils.cpe[".pipeline/commonPipelineEnvironment/mtarFilePath"] = "test.mtar"
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(&utils, &uploader, &options)
		assert.EqualError(t, err, "could not read from required project descriptor file 'mta.yml'")
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
			"failed to parse contents of the project descriptor file 'mta.yaml'")
		assert.Equal(t, 0, len(uploader.GetArtifacts()))
	})
	t.Run("Uploading MTA project fails due invalid version in YAML content", func(t *testing.T) {
		utils := newMockUtilsBundle(true, false)
		utils.files["mta.yaml"] = []byte(testMtaYmlNoVersion)
		utils.cpe[".pipeline/commonPipelineEnvironment/mtarFilePath"] = "test.mtar"
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(&utils, &uploader, &options)
		assert.EqualError(t, err,
			"the project descriptor file 'mta.yaml' has an invalid version: version must not be empty")
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

func TestUploadArtifacts(t *testing.T) {
	t.Run("Uploading MTA project fails without any artifacts", func(t *testing.T) {
		utils := newMockUtilsBundle(false, true)
		uploader := mockUploader{}
		options := createOptions()

		err := uploadArtifacts(&utils, &uploader, &options)
		assert.EqualError(t, err, "no artifacts to upload")
	})
	t.Run("Uploading MTA project fails for unknown reasons", func(t *testing.T) {
		utils := newMockUtilsBundle(false, true)

		// Configure mocked execRunner to fail
		utils.execRunner.ShouldFailOnCommand = map[string]error{}
		utils.execRunner.ShouldFailOnCommand["mvn"] = fmt.Errorf("failed")

		uploader := mockUploader{}
		_ = uploader.AddArtifact(nexus.ArtifactDescription{
			File: "mta.yaml",
			Type: "yaml",
			ID:   "my.artifact",
		})
		_ = uploader.AddArtifact(nexus.ArtifactDescription{
			File: "artifact.mtar",
			Type: "yaml",
			ID:   "my.artifact",
		})

		options := createOptions()

		err := uploadArtifacts(&utils, &uploader, &options)
		assert.EqualError(t, err, "uploading artifacts failed: failed to run executable, command: '[mvn -Durl=http:// -DgroupId=my.group.id -DartifactId=my.artifact -Dversion= -Dfile=mta.yaml -DgeneratePom=false -Dfiles=artifact.mtar -Dclassifiers= -Dtypes=yaml --batch-mode deploy:deploy-file]', error: failed")
	})
	t.Run("Uploading MTA project fails with different artifact IDs", func(t *testing.T) {
		utils := newMockUtilsBundle(false, true)
		uploader := mockUploader{}
		options := createOptions()

		_ = uploader.AddArtifact(nexus.ArtifactDescription{
			File: "mta.yaml",
			Type: "yaml",
			ID:   "my.artifact",
		})
		_ = uploader.AddArtifact(nexus.ArtifactDescription{
			File: "pom.yml",
			Type: "pom",
			ID:   "my.artifact",
		})
		_ = uploader.AddArtifact(nexus.ArtifactDescription{
			File: "artifact.mtar",
			Type: "yaml",
			ID:   "artifact",
		})

		err := uploadArtifacts(&utils, &uploader, &options)
		assert.EqualError(t, err, "cannot deploy artifacts with different IDs in one run (my.artifact vs. artifact)")
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
	t.Run("Test uploading Maven project produces correct mvn command line", func(t *testing.T) {
		utils := newMockUtilsBundle(false, true)
		utils.files["pom.xml"] = testPomXml
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(&utils, &uploader, &options)
		if assert.NoError(t, err, "expected Maven upload to work") {
			expectedParameters := []string{"-Dmaven.test.skip", "-DaltDeploymentRepository=maven-releases::default::http://localhost:8081/repository/maven-releases/", "--batch-mode", "deploy"}
			assert.Equal(t, len(utils.execRunner.Calls[0].Params), len(expectedParameters))
			assert.Equal(t, utils.execRunner.Calls[0], mock.ExecCall{Exec: "mvn", Params: expectedParameters})
		}
	})
	t.Run("Test uploading Maven project with m2 path produces correct mvn command line", func(t *testing.T) {
		utils := newMockUtilsBundle(false, true)
		utils.files["pom.xml"] = testPomXml
		uploader := mockUploader{}
		options := createOptions()
		options.M2Path = ".pipeline/m2"

		err := runNexusUpload(&utils, &uploader, &options)
		if assert.NoError(t, err, "expected Maven upload to work") {
			expectedParameters := []string{"-Dmaven.repo.local=.pipeline/m2", "-Dmaven.test.skip", "-DaltDeploymentRepository=maven-releases::default::http://localhost:8081/repository/maven-releases/", "--batch-mode", "deploy"}
			assert.Equal(t, len(utils.execRunner.Calls[0].Params), len(expectedParameters))
			assert.Equal(t, utils.execRunner.Calls[0], mock.ExecCall{Exec: "mvn", Params: expectedParameters})
		}
	})
}

func TestUploadUnknownProjectFails(t *testing.T) {
	utils := newMockUtilsBundle(false, false)
	uploader := mockUploader{}
	options := createOptions()

	err := runNexusUpload(&utils, &uploader, &options)
	assert.EqualError(t, err, "unsupported project structure")
}
