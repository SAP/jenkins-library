//go:build unit
// +build unit

package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/nexus"
	"github.com/stretchr/testify/assert"
)

type mockUtilsBundle struct {
	*mock.FilesMock
	*mock.ExecMockRunner
	mta        bool
	maven      bool
	npm        bool
	properties map[string]map[string]string
	cpe        map[string]string
}

func (m *mockUtilsBundle) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	return errors.New("test should not download files")
}

func newMockUtilsBundle(usesMta, usesMaven, usesNpm bool) *mockUtilsBundle {
	utils := mockUtilsBundle{
		FilesMock:      &mock.FilesMock{},
		ExecMockRunner: &mock.ExecMockRunner{},
		mta:            usesMta,
		maven:          usesMaven,
		npm:            usesNpm,
	}
	utils.properties = map[string]map[string]string{}
	utils.cpe = map[string]string{}
	return &utils
}

func (m *mockUtilsBundle) UsesMta() bool {
	return m.mta
}

func (m *mockUtilsBundle) UsesMaven() bool {
	return m.maven
}

func (m *mockUtilsBundle) UsesNpm() bool {
	return m.npm
}

func (m *mockUtilsBundle) getEnvParameter(path, name string) string {
	path = path + "/" + name
	return m.cpe[path]
}

func (m *mockUtilsBundle) setProperty(pomFile, expression, value string) {
	pomFile = strings.ReplaceAll(pomFile, "/", string(os.PathSeparator))
	pomFile = strings.ReplaceAll(pomFile, "\\", string(os.PathSeparator))

	pom := m.properties[pomFile]
	if pom == nil {
		pom = map[string]string{}
		m.properties[pomFile] = pom
	}
	pom[expression] = value
}

func (m *mockUtilsBundle) evaluate(options *maven.EvaluateOptions, expression string) (string, error) {
	pom := m.properties[options.PomPath]
	if pom == nil {
		return "", fmt.Errorf("pom file '%s' not found", options.PomPath)
	}
	value := pom[expression]
	if value == "<empty>" {
		return "", nil
	}
	if value == "" {
		return "", fmt.Errorf("property '%s' not found in '%s'", expression, options.PomPath)
	}
	return value, nil
}

type mockUploader struct {
	nexus.Upload
	uploadedArtifacts []nexus.ArtifactDescription
}

func (m *mockUploader) Clear() {
	// Clear is called after a successful upload. Record the artifacts that are present before
	// they are cleared. This way we can later peek into the set of all artifacts that were
	// uploaded across multiple bundles.
	m.uploadedArtifacts = append(m.uploadedArtifacts, m.GetArtifacts()...)
	m.Upload.Clear()
}

func createOptions() nexusUploadOptions {
	return nexusUploadOptions{
		MavenRepository: "maven-releases",
		NpmRepository:   "npm-repo",
		GroupID:         "my.group.id",
		ArtifactID:      "artifact.id",
		Version:         "nexus3",
		Url:             "localhost:8081",
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

var testPackageJson = []byte(`{
  "name": "npm-nexus-upload-test",
  "version": "1.0.0"
}
`)

func TestUploadMTAProjects(t *testing.T) {
	t.Parallel()
	t.Run("Uploading MTA project without groupId parameter fails", func(t *testing.T) {
		t.Parallel()
		utils := newMockUtilsBundle(true, false, false)
		utils.AddFile("mta.yaml", testMtaYml)
		utils.cpe[".pipeline/commonPipelineEnvironment/mtarFilePath"] = "test.mtar"
		uploader := mockUploader{}
		options := createOptions()
		options.GroupID = ""

		err := runNexusUpload(utils, &uploader, &options)
		assert.EqualError(t, err, "the 'groupId' parameter needs to be provided for MTA projects")
		assert.Equal(t, 0, len(uploader.GetArtifacts()))
		assert.Equal(t, 0, len(uploader.uploadedArtifacts))
	})
	t.Run("Uploading MTA project without artifactId parameter works", func(t *testing.T) {
		t.Parallel()
		utils := newMockUtilsBundle(true, false, false)
		utils.AddFile("mta.yaml", testMtaYml)
		utils.AddFile("test.mtar", []byte("contentsOfMtar"))
		utils.cpe[".pipeline/commonPipelineEnvironment/mtarFilePath"] = "test.mtar"
		uploader := mockUploader{}
		options := createOptions()
		options.ArtifactID = ""

		err := runNexusUpload(utils, &uploader, &options)
		if assert.NoError(t, err) {
			assert.Equal(t, 2, len(uploader.uploadedArtifacts))
			assert.Equal(t, "test", uploader.GetArtifactsID())
		}
	})
	t.Run("Uploading MTA project fails due to missing yaml file", func(t *testing.T) {
		t.Parallel()
		utils := newMockUtilsBundle(true, false, false)
		utils.cpe[".pipeline/commonPipelineEnvironment/mtarFilePath"] = "test.mtar"
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(utils, &uploader, &options)
		assert.EqualError(t, err, "could not read from required project descriptor file 'mta.yml'")
		assert.Equal(t, 0, len(uploader.GetArtifacts()))
		assert.Equal(t, 0, len(uploader.uploadedArtifacts))
	})
	t.Run("Uploading MTA project fails due to garbage YAML content", func(t *testing.T) {
		t.Parallel()
		utils := newMockUtilsBundle(true, false, false)
		utils.AddFile("mta.yaml", []byte("garbage"))
		utils.cpe[".pipeline/commonPipelineEnvironment/mtarFilePath"] = "test.mtar"
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(utils, &uploader, &options)
		assert.EqualError(t, err,
			"failed to parse contents of the project descriptor file 'mta.yaml'")
		assert.Equal(t, 0, len(uploader.GetArtifacts()))
		assert.Equal(t, 0, len(uploader.uploadedArtifacts))
	})
	t.Run("Uploading MTA project fails due invalid version in YAML content", func(t *testing.T) {
		t.Parallel()
		utils := newMockUtilsBundle(true, false, false)
		utils.AddFile("mta.yaml", []byte(testMtaYmlNoVersion))
		utils.cpe[".pipeline/commonPipelineEnvironment/mtarFilePath"] = "test.mtar"
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(utils, &uploader, &options)
		assert.EqualError(t, err,
			"the project descriptor file 'mta.yaml' has an invalid version: version must not be empty")
		assert.Equal(t, 0, len(uploader.GetArtifacts()))
		assert.Equal(t, 0, len(uploader.uploadedArtifacts))
	})
	t.Run("Test uploading mta.yaml project fails due to missing mtar file", func(t *testing.T) {
		t.Parallel()
		utils := newMockUtilsBundle(true, false, false)
		utils.AddFile("mta.yaml", testMtaYml)
		utils.cpe[".pipeline/commonPipelineEnvironment/mtarFilePath"] = "test.mtar"
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(utils, &uploader, &options)
		assert.EqualError(t, err, "artifact file not found 'test.mtar'")

		assert.Equal(t, "0.3.0", uploader.GetArtifactsVersion())
		assert.Equal(t, "artifact.id", uploader.GetArtifactsID())

		// Check the artifacts that /would/ have been uploaded
		artifacts := uploader.GetArtifacts()
		if assert.Equal(t, 1, len(artifacts)) {
			assert.Equal(t, "mta.yaml", artifacts[0].File)
			assert.Equal(t, "yaml", artifacts[0].Type)
		}
		assert.Equal(t, 0, len(uploader.uploadedArtifacts))
	})
	t.Run("Test uploading mta.yaml project works", func(t *testing.T) {
		t.Parallel()
		utils := newMockUtilsBundle(true, false, false)
		utils.AddFile("mta.yaml", testMtaYml)
		utils.AddFile("test.mtar", []byte("contentsOfMtar"))
		utils.cpe[".pipeline/commonPipelineEnvironment/mtarFilePath"] = "test.mtar"
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(utils, &uploader, &options)
		assert.NoError(t, err, "expected mta.yaml project upload to work")

		assert.Equal(t, "0.3.0", uploader.GetArtifactsVersion())
		assert.Equal(t, "artifact.id", uploader.GetArtifactsID())

		artifacts := uploader.uploadedArtifacts
		if assert.Equal(t, 2, len(artifacts)) {
			assert.Equal(t, "mta.yaml", artifacts[0].File)
			assert.Equal(t, "yaml", artifacts[0].Type)

			assert.Equal(t, "test.mtar", artifacts[1].File)
			assert.Equal(t, "mtar", artifacts[1].Type)
		}
	})
	t.Run("Test uploading mta.yml project works", func(t *testing.T) {
		t.Parallel()
		utils := newMockUtilsBundle(true, false, false)
		utils.AddFile("mta.yml", testMtaYml)
		utils.AddFile("test.mtar", []byte("contentsOfMtar"))
		utils.cpe[".pipeline/commonPipelineEnvironment/mtarFilePath"] = "test.mtar"
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(utils, &uploader, &options)
		assert.NoError(t, err, "expected mta.yml project upload to work")

		assert.Equal(t, "0.3.0", uploader.GetArtifactsVersion())
		assert.Equal(t, "artifact.id", uploader.GetArtifactsID())

		artifacts := uploader.uploadedArtifacts
		if assert.Equal(t, 2, len(artifacts)) {
			assert.Equal(t, "mta.yml", artifacts[0].File)
			assert.Equal(t, "yaml", artifacts[0].Type)

			assert.Equal(t, "test.mtar", artifacts[1].File)
			assert.Equal(t, "mtar", artifacts[1].Type)
		}
	})
}

func TestUploadArtifacts(t *testing.T) {
	t.Parallel()
	t.Run("Uploading MTA project fails without info", func(t *testing.T) {
		t.Parallel()
		utils := newMockUtilsBundle(false, true, false)
		uploader := mockUploader{}
		options := createOptions()

		err := uploadArtifacts(utils, &uploader, &options, false)
		assert.EqualError(t, err, "no group ID was provided, or could be established from project files")
	})
	t.Run("Uploading MTA project fails without any artifacts", func(t *testing.T) {
		t.Parallel()
		utils := newMockUtilsBundle(false, true, false)
		uploader := mockUploader{}
		options := createOptions()

		_ = uploader.SetInfo(options.GroupID, "some.id", "3.0")

		err := uploadArtifacts(utils, &uploader, &options, false)
		assert.EqualError(t, err, "no artifacts to upload")
	})
	t.Run("Uploading MTA project fails for unknown reasons", func(t *testing.T) {
		t.Parallel()
		utils := newMockUtilsBundle(false, true, false)

		// Configure mocked execRunner to fail
		utils.ShouldFailOnCommand = map[string]error{}
		utils.ShouldFailOnCommand["mvn"] = fmt.Errorf("failed")

		uploader := mockUploader{}
		options := createOptions()
		_ = uploader.SetInfo(options.GroupID, "some.id", "3.0")
		_ = uploader.AddArtifact(nexus.ArtifactDescription{
			File: "mta.yaml",
			Type: "yaml",
		})
		_ = uploader.AddArtifact(nexus.ArtifactDescription{
			File: "artifact.mtar",
			Type: "yaml",
		})

		err := uploadArtifacts(utils, &uploader, &options, false)
		assert.EqualError(t, err, "uploading artifacts for ID 'some.id' failed: failed to run executable, command: '[mvn -Durl=http:// -DgroupId=my.group.id -Dversion=3.0 -DartifactId=some.id -Dfile=mta.yaml -Dpackaging=yaml -DgeneratePom=false -Dfiles=artifact.mtar -Dclassifiers= -Dtypes=yaml -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn --batch-mode "+deployGoal+"]', error: failed")
	})
	t.Run("Uploading bundle generates correct maven parameters", func(t *testing.T) {
		t.Parallel()
		utils := newMockUtilsBundle(false, true, false)
		uploader := mockUploader{}
		options := createOptions()

		_ = uploader.SetRepoURL("localhost:8081", "nexus3", "maven-releases", "npm-repo")
		_ = uploader.SetInfo(options.GroupID, "my.artifact", "4.0")
		_ = uploader.AddArtifact(nexus.ArtifactDescription{
			File: "mta.yaml",
			Type: "yaml",
		})
		_ = uploader.AddArtifact(nexus.ArtifactDescription{
			File: "pom.yml",
			Type: "pom",
		})

		err := uploadArtifacts(utils, &uploader, &options, false)
		assert.NoError(t, err, "expected upload as two bundles to work")
		assert.Equal(t, 1, len(utils.Calls))

		expectedParameters1 := []string{
			"-Durl=http://localhost:8081/repository/maven-releases/",
			"-DgroupId=my.group.id",
			"-Dversion=4.0",
			"-DartifactId=my.artifact",
			"-Dfile=mta.yaml",
			"-Dpackaging=yaml",
			"-DgeneratePom=false",
			"-Dfiles=pom.yml",
			"-Dclassifiers=",
			"-Dtypes=pom",
			"-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn",
			"--batch-mode",
			deployGoal}
		assert.Equal(t, len(expectedParameters1), len(utils.Calls[0].Params))
		assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: expectedParameters1}, utils.Calls[0])
	})
}

func TestRunNexusUpload(t *testing.T) {
	t.Parallel()
	t.Run("uploading without any repos fails step", func(t *testing.T) {
		t.Parallel()
		utils := newMockUtilsBundle(false, false, true)
		utils.AddFile("package.json", testPackageJson)
		uploader := mockUploader{}
		options := nexusUploadOptions{
			Url: "localhost:8081",
		}

		err := runNexusUpload(utils, &uploader, &options)
		assert.EqualError(t, err, "none of the parameters 'mavenRepository' and 'npmRepository' are configured, or 'format' should be set if the 'url' already contains the repository ID")
	})
	t.Run("Test uploading simple npm project", func(t *testing.T) {
		t.Parallel()
		utils := newMockUtilsBundle(false, false, true)
		utils.AddFile("package.json", testPackageJson)
		uploader := mockUploader{}
		options := createOptions()
		options.Username = "admin"
		options.Password = "admin123"

		err := runNexusUpload(utils, &uploader, &options)
		assert.NoError(t, err, "expected npm upload to work")

		assert.Equal(t, "localhost:8081/repository/npm-repo/", uploader.GetNpmRepoURL())

		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"publish"}}, utils.Calls[0])
		assert.Equal(t, []string{"npm_config_registry=http://localhost:8081/repository/npm-repo/", "npm_config_email=project-piper@no-reply.com", "npm_config__auth=YWRtaW46YWRtaW4xMjM="}, utils.Env)
	})
}

func TestUploadMavenProjects(t *testing.T) {
	t.Parallel()
	t.Run("Uploading Maven project fails due to missing pom.xml", func(t *testing.T) {
		t.Parallel()
		utils := newMockUtilsBundle(false, true, false)
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(utils, &uploader, &options)
		assert.EqualError(t, err, "pom.xml not found")
		assert.Equal(t, 0, len(uploader.uploadedArtifacts))
	})
	t.Run("Test uploading Maven project with POM packaging works", func(t *testing.T) {
		t.Parallel()
		utils := newMockUtilsBundle(false, true, false)
		utils.setProperty("pom.xml", "project.version", "1.0")
		utils.setProperty("pom.xml", "project.groupId", "com.mycompany.app")
		utils.setProperty("pom.xml", "project.artifactId", "my-app")
		utils.setProperty("pom.xml", "project.packaging", "pom")
		utils.setProperty("pom.xml", "project.build.finalName", "my-app-1.0")
		utils.AddFile("pom.xml", testPomXml)
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(utils, &uploader, &options)
		assert.NoError(t, err, "expected Maven upload to work")
		assert.Equal(t, "1.0", uploader.GetArtifactsVersion())
		assert.Equal(t, "my-app", uploader.GetArtifactsID())

		artifacts := uploader.uploadedArtifacts
		if assert.Equal(t, 1, len(artifacts)) {
			assert.Equal(t, "pom.xml", artifacts[0].File)
			assert.Equal(t, "pom", artifacts[0].Type)
		}
	})
	t.Run("Test uploading Maven project with JAR packaging fails without main target", func(t *testing.T) {
		t.Parallel()
		utils := newMockUtilsBundle(false, true, false)
		utils.setProperty("pom.xml", "project.version", "1.0")
		utils.setProperty("pom.xml", "project.groupId", "com.mycompany.app")
		utils.setProperty("pom.xml", "project.artifactId", "my-app")
		utils.setProperty("pom.xml", "project.packaging", "jar")
		utils.setProperty("pom.xml", "project.build.finalName", "my-app-1.0")
		utils.AddFile("pom.xml", testPomXml)
		utils.AddDir("target")
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(utils, &uploader, &options)
		assert.EqualError(t, err, "target artifact not found for packaging 'jar'")
		assert.Equal(t, 0, len(uploader.uploadedArtifacts))
	})
	t.Run("Test uploading Maven project with JAR packaging works", func(t *testing.T) {
		t.Parallel()
		utils := newMockUtilsBundle(false, true, false)
		utils.setProperty("pom.xml", "project.version", "1.0")
		utils.setProperty("pom.xml", "project.groupId", "com.mycompany.app")
		utils.setProperty("pom.xml", "project.artifactId", "my-app")
		utils.setProperty("pom.xml", "project.packaging", "jar")
		utils.setProperty("pom.xml", "project.build.finalName", "my-app-1.0")
		utils.AddFile("pom.xml", testPomXml)
		utils.AddFile(filepath.Join("target", "my-app-1.0.jar"), []byte("contentsOfJar"))
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(utils, &uploader, &options)
		assert.NoError(t, err, "expected Maven upload to work")

		assert.Equal(t, "1.0", uploader.GetArtifactsVersion())
		assert.Equal(t, "my-app", uploader.GetArtifactsID())

		artifacts := uploader.uploadedArtifacts
		if assert.Equal(t, 2, len(artifacts)) {
			assert.Equal(t, "pom.xml", artifacts[0].File)
			assert.Equal(t, "pom", artifacts[0].Type)

			assert.Equal(t, filepath.Join("target", "my-app-1.0.jar"), artifacts[1].File)
			assert.Equal(t, "jar", artifacts[1].Type)
		}
	})
	t.Run("Test uploading Maven project with fall-back to JAR packaging works", func(t *testing.T) {
		t.Parallel()
		utils := newMockUtilsBundle(false, true, false)
		utils.setProperty("pom.xml", "project.version", "1.0")
		utils.setProperty("pom.xml", "project.groupId", "com.mycompany.app")
		utils.setProperty("pom.xml", "project.artifactId", "my-app")
		utils.setProperty("pom.xml", "project.packaging", "<empty>")
		utils.setProperty("pom.xml", "project.build.finalName", "my-app-1.0")
		utils.AddFile("pom.xml", testPomXml)
		utils.AddFile(filepath.Join("target", "my-app-1.0.jar"), []byte("contentsOfJar"))
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(utils, &uploader, &options)
		assert.NoError(t, err, "expected Maven upload to work")
		assert.Equal(t, "1.0", uploader.GetArtifactsVersion())
		assert.Equal(t, "my-app", uploader.GetArtifactsID())

		artifacts := uploader.uploadedArtifacts
		if assert.Equal(t, 2, len(artifacts)) {
			assert.Equal(t, "pom.xml", artifacts[0].File)
			assert.Equal(t, "pom", artifacts[0].Type)

			assert.Equal(t, filepath.Join("target", "my-app-1.0.jar"), artifacts[1].File)
			assert.Equal(t, "jar", artifacts[1].Type)
		}
	})
	t.Run("Test uploading Maven project with fall-back to group id from parameters works", func(t *testing.T) {
		t.Parallel()
		utils := newMockUtilsBundle(false, true, false)
		utils.setProperty("pom.xml", "project.version", "1.0")
		utils.setProperty("pom.xml", "project.artifactId", "my-app")
		utils.setProperty("pom.xml", "project.packaging", "pom")
		utils.setProperty("pom.xml", "project.build.finalName", "my-app-1.0")
		utils.AddFile("pom.xml", testPomXml)
		uploader := mockUploader{}
		options := createOptions()
		options.GroupID = "awesome.group"

		err := runNexusUpload(utils, &uploader, &options)
		assert.NoError(t, err, "expected Maven upload to work")

		assert.Equal(t, "localhost:8081/repository/maven-releases/",
			uploader.GetMavenRepoURL())
		assert.Equal(t, "1.0", uploader.GetArtifactsVersion())
		assert.Equal(t, "my-app", uploader.GetArtifactsID())

		artifacts := uploader.uploadedArtifacts
		if assert.Equal(t, 1, len(artifacts)) {
			assert.Equal(t, "pom.xml", artifacts[0].File)
			assert.Equal(t, "pom", artifacts[0].Type)
		}
	})
	t.Run("Test uploading Maven project with fall-back for finalBuildName works", func(t *testing.T) {
		t.Parallel()
		utils := newMockUtilsBundle(false, true, false)
		utils.setProperty("pom.xml", "project.version", "1.0")
		utils.setProperty("pom.xml", "project.groupId", "awesome.group")
		utils.setProperty("pom.xml", "project.artifactId", "my-app")
		utils.setProperty("pom.xml", "project.packaging", "jar")
		utils.AddFile("pom.xml", testPomXml)
		utils.AddFile(filepath.Join("target", "my-app-1.0.jar"), []byte("contentsOfJar"))
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(utils, &uploader, &options)
		assert.NoError(t, err, "expected Maven upload to work")

		assert.Equal(t, "localhost:8081/repository/maven-releases/",
			uploader.GetMavenRepoURL())
		assert.Equal(t, "1.0", uploader.GetArtifactsVersion())
		assert.Equal(t, "my-app", uploader.GetArtifactsID())

		artifacts := uploader.uploadedArtifacts
		if assert.Equal(t, 2, len(artifacts)) {
			assert.Equal(t, "pom.xml", artifacts[0].File)
			assert.Equal(t, "pom", artifacts[0].Type)
			assert.Equal(t, filepath.Join("target", "my-app-1.0.jar"), artifacts[1].File)
			assert.Equal(t, "jar", artifacts[1].Type)
		}
	})
	t.Run("Test uploading Maven project with application module and finalName works", func(t *testing.T) {
		t.Parallel()
		utils := newMockUtilsBundle(false, true, false)
		utils.setProperty("pom.xml", "project.version", "1.0")
		utils.setProperty("pom.xml", "project.groupId", "com.mycompany.app")
		utils.setProperty("pom.xml", "project.artifactId", "my-app")
		utils.setProperty("pom.xml", "project.packaging", "pom")
		utils.setProperty("pom.xml", "project.build.finalName", "my-app-1.0")
		utils.setProperty("application/pom.xml", "project.version", "1.0")
		utils.setProperty("application/pom.xml", "project.groupId", "com.mycompany.app")
		utils.setProperty("application/pom.xml", "project.artifactId", "my-app-app")
		utils.setProperty("application/pom.xml", "project.packaging", "war")
		utils.setProperty("application/pom.xml", "project.build.finalName", "final-artifact")
		utils.setProperty("integration-tests/pom.xml", "project.version", "1.0")
		utils.setProperty("integration-tests/pom.xml", "project.groupId", "com.mycompany.app")
		utils.setProperty("integration-tests/pom.xml", "project.artifactId", "my-app-app-integration-tests")
		utils.setProperty("integration-tests/pom.xml", "project.packaging", "jar")
		utils.setProperty("integration-tests/pom.xml", "project.build.finalName", "final-artifact")
		utils.setProperty("unit-tests/pom.xml", "project.version", "1.0")
		utils.setProperty("unit-tests/pom.xml", "project.groupId", "com.mycompany.app")
		utils.setProperty("unit-tests/pom.xml", "project.artifactId", "my-app-app-unit-tests")
		utils.setProperty("unit-tests/pom.xml", "project.packaging", "jar")
		utils.setProperty("unit-tests/pom.xml", "project.build.finalName", "final-artifact")
		utils.setProperty("performance-tests/pom.xml", "project.version", "1.0")
		utils.setProperty("performance-tests/pom.xml", "project.groupId", "com.mycompany.app")
		utils.setProperty("performance-tests/pom.xml", "project.artifactId", "my-app-app")
		utils.setProperty("performance-tests/pom.xml", "project.packaging", "")
		utils.AddFile("pom.xml", testPomXml)
		utils.AddFile(filepath.Join("application", "pom.xml"), testPomXml)
		utils.AddFile("application/target/final-artifact.war", []byte("contentsOfJar"))
		utils.AddFile("application/target/final-artifact-classes.jar", []byte("contentsOfClassesJar"))
		utils.AddFile("integration-tests/pom.xml", testPomXml)
		utils.AddFile("integration-tests/target/final-artifact-integration-tests.jar", []byte("contentsOfJar"))
		utils.AddFile("unit-tests/pom.xml", testPomXml)
		utils.AddFile("unit-tests/target/final-artifact-unit-tests.jar", []byte("contentsOfJar"))
		utils.AddFile("performance-tests/pom.xml", testPomXml)
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(utils, &uploader, &options)
		assert.NoError(t, err, "expected upload of maven project with application module to succeed")
		assert.Equal(t, "1.0", uploader.GetArtifactsVersion())
		assert.Equal(t, "my-app", uploader.GetArtifactsID())

		artifacts := uploader.uploadedArtifacts
		if assert.Equal(t, 4, len(artifacts)) {
			assert.Equal(t, filepath.Join("application", "pom.xml"), artifacts[0].File)
			assert.Equal(t, "pom", artifacts[0].Type)

			assert.Equal(t, filepath.Join("application", "target", "final-artifact.war"), artifacts[1].File)
			assert.Equal(t, "war", artifacts[1].Type)

			assert.Equal(t, filepath.Join("application", "target", "final-artifact-classes.jar"), artifacts[2].File)
			assert.Equal(t, "jar", artifacts[2].Type)

			assert.Equal(t, "pom.xml", artifacts[3].File)
			assert.Equal(t, "pom", artifacts[3].Type)

		}
		if assert.Equal(t, 2, len(utils.Calls)) {
			expectedParameters1 := []string{
				"-Durl=http://localhost:8081/repository/maven-releases/",
				"-DgroupId=com.mycompany.app",
				"-Dversion=1.0",
				"-DartifactId=my-app-app",
				"-Dfile=" + filepath.Join("application", "pom.xml"),
				"-Dpackaging=pom",
				"-Dfiles=" + filepath.Join("application", "target", "final-artifact.war") + "," + filepath.Join("application", "target", "final-artifact-classes.jar"),
				"-Dclassifiers=,classes",
				"-Dtypes=war,jar",
				"-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn",
				"--batch-mode",
				deployGoal}
			assert.Equal(t, len(expectedParameters1), len(utils.Calls[0].Params))
			assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: expectedParameters1}, utils.Calls[0])

			expectedParameters2 := []string{
				"-Durl=http://localhost:8081/repository/maven-releases/",
				"-DgroupId=com.mycompany.app",
				"-Dversion=1.0",
				"-DartifactId=my-app",
				"-Dfile=pom.xml",
				"-Dpackaging=pom",
				"-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn",
				"--batch-mode",
				deployGoal}
			assert.Equal(t, len(expectedParameters2), len(utils.Calls[1].Params))
			assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: expectedParameters2}, utils.Calls[1])
		}
	})
	t.Run("Write credentials settings", func(t *testing.T) {
		t.Parallel()
		utils := newMockUtilsBundle(false, true, false)
		utils.setProperty("pom.xml", "project.version", "1.0")
		utils.setProperty("pom.xml", "project.groupId", "com.mycompany.app")
		utils.setProperty("pom.xml", "project.artifactId", "my-app")
		utils.setProperty("pom.xml", "project.packaging", "pom")
		utils.setProperty("pom.xml", "project.build.finalName", "my-app-1.0")
		utils.AddFile("pom.xml", testPomXml)
		uploader := mockUploader{}
		options := createOptions()
		options.Username = "admin"
		options.Password = "admin123"

		err := runNexusUpload(utils, &uploader, &options)
		assert.NoError(t, err, "expected Maven upload to work")

		assert.Equal(t, 1, len(utils.Calls))
		dir, _ := os.Getwd()
		absoluteSettingsPath := filepath.Join(dir, settingsPath)
		expectedParameters1 := []string{
			"--settings",
			absoluteSettingsPath,
			"-Durl=http://localhost:8081/repository/maven-releases/",
			"-DgroupId=com.mycompany.app",
			"-Dversion=1.0",
			"-DartifactId=my-app",
			"-DrepositoryId=" + settingsServerID,
			"-Dfile=pom.xml",
			"-Dpackaging=pom",
			"-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn",
			"--batch-mode",
			deployGoal}
		assert.Equal(t, len(expectedParameters1), len(utils.Calls[0].Params))
		assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: expectedParameters1}, utils.Calls[0])

		expectedEnv := []string{"NEXUS_username=admin", "NEXUS_password=admin123"}
		assert.Equal(t, 2, len(utils.Env))
		assert.Equal(t, expectedEnv, utils.Env)

		assert.False(t, utils.HasFile(settingsPath))
		assert.True(t, utils.HasRemovedFile(settingsPath))
	})
}

func TestSetupNexusCredentialsSettingsFile(t *testing.T) {
	t.Parallel()
	utils := newMockUtilsBundle(false, true, false)
	options := nexusUploadOptions{Username: "admin", Password: "admin123"}
	mavenOptions := maven.ExecuteOptions{}
	settingsPath, err := setupNexusCredentialsSettingsFile(utils, &options, &mavenOptions)

	assert.NoError(t, err, "expected setting up credentials settings.xml to work")
	assert.Equal(t, 0, len(utils.Calls))
	expectedEnv := []string{"NEXUS_username=admin", "NEXUS_password=admin123"}
	assert.Equal(t, 2, len(utils.Env))
	assert.Equal(t, expectedEnv, utils.Env)

	assert.True(t, settingsPath != "")
	assert.True(t, utils.HasFile(settingsPath))
}
