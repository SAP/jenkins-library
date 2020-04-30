package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/nexus"
	"github.com/bmatcuk/doublestar"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

type mockUtilsBundle struct {
	mta          bool
	maven        bool
	npm          bool
	files        map[string][]byte
	removedFiles map[string][]byte
	properties   map[string]map[string]string
	cpe          map[string]string
	execRunner   mock.ExecMockRunner
}

func newMockUtilsBundle(usesMta, usesMaven, usesNpm bool) mockUtilsBundle {
	utils := mockUtilsBundle{mta: usesMta, maven: usesMaven, npm: usesNpm}
	utils.files = map[string][]byte{}
	utils.removedFiles = map[string][]byte{}
	utils.properties = map[string]map[string]string{}
	utils.cpe = map[string]string{}
	return utils
}

func (m *mockUtilsBundle) usesMta() bool {
	return m.mta
}

func (m *mockUtilsBundle) usesMaven() bool {
	return m.maven
}

func (m *mockUtilsBundle) usesNpm() bool {
	return m.npm
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

func (m *mockUtilsBundle) fileWrite(path string, content []byte, _ os.FileMode) error {
	m.files[path] = content
	return nil
}

func (m *mockUtilsBundle) fileRemove(path string) {
	contents := m.files[path]
	m.files[path] = nil
	if contents != nil {
		m.removedFiles[path] = contents
	}
}

func (m *mockUtilsBundle) dirExists(path string) (bool, error) {
	for file := range m.files {
		dir := filepath.Dir(file)
		if dir == path {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockUtilsBundle) getEnvParameter(path, name string) string {
	path = path + "/" + name
	return m.cpe[path]
}

func (m *mockUtilsBundle) getExecRunner() execRunner {
	return &m.execRunner
}

func (m *mockUtilsBundle) setProperty(pomFile, expression, value string) {
	pom := m.properties[pomFile]
	if pom == nil {
		pom = map[string]string{}
		m.properties[pomFile] = pom
	}
	pom[expression] = value
}

func (m *mockUtilsBundle) evaluate(pomFile, expression string) (string, error) {
	pom := m.properties[pomFile]
	if pom == nil {
		return "", fmt.Errorf("pom file '%s' not found", pomFile)
	}
	value := pom[expression]
	if value == "<empty>" {
		return "", nil
	}
	if value == "" {
		return "", fmt.Errorf("property '%s' not found in '%s'", expression, pomFile)
	}
	return value, nil
}

type byLen []string

func (a byLen) Len() int {
	return len(a)
}

func (a byLen) Less(i, j int) bool {
	return len(a[i]) < len(a[j])
}

func (a byLen) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (m *mockUtilsBundle) glob(pattern string) ([]string, error) {
	var matches []string
	for path := range m.files {
		matched, _ := doublestar.Match(pattern, path)
		if matched {
			matches = append(matches, path)
		}
	}
	// The order in m.files is not deterministic, this would result in flaky tests.
	sort.Sort(byLen(matches))
	return matches, nil
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
	t.Run("Uploading MTA project without groupId parameter fails", func(t *testing.T) {
		utils := newMockUtilsBundle(true, false, false)
		utils.files["mta.yaml"] = testMtaYml
		utils.cpe[".pipeline/commonPipelineEnvironment/mtarFilePath"] = "test.mtar"
		uploader := mockUploader{}
		options := createOptions()
		options.GroupID = ""

		err := runNexusUpload(&utils, &uploader, &options)
		assert.EqualError(t, err, "the 'groupId' parameter needs to be provided for MTA projects")
		assert.Equal(t, 0, len(uploader.GetArtifacts()))
		assert.Equal(t, 0, len(uploader.uploadedArtifacts))
	})
	t.Run("Uploading MTA project without artifactId parameter works", func(t *testing.T) {
		utils := newMockUtilsBundle(true, false, false)
		utils.files["mta.yaml"] = testMtaYml
		utils.files["test.mtar"] = []byte("contentsOfMtar")
		utils.cpe[".pipeline/commonPipelineEnvironment/mtarFilePath"] = "test.mtar"
		uploader := mockUploader{}
		options := createOptions()
		options.ArtifactID = ""

		err := runNexusUpload(&utils, &uploader, &options)
		if assert.NoError(t, err) {
			assert.Equal(t, 2, len(uploader.uploadedArtifacts))
			assert.Equal(t, "test", uploader.GetArtifactsID())
		}
	})
	t.Run("Uploading MTA project fails due to missing yaml file", func(t *testing.T) {
		utils := newMockUtilsBundle(true, false, false)
		utils.cpe[".pipeline/commonPipelineEnvironment/mtarFilePath"] = "test.mtar"
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(&utils, &uploader, &options)
		assert.EqualError(t, err, "could not read from required project descriptor file 'mta.yml'")
		assert.Equal(t, 0, len(uploader.GetArtifacts()))
		assert.Equal(t, 0, len(uploader.uploadedArtifacts))
	})
	t.Run("Uploading MTA project fails due to garbage YAML content", func(t *testing.T) {
		utils := newMockUtilsBundle(true, false, false)
		utils.files["mta.yaml"] = []byte("garbage")
		utils.cpe[".pipeline/commonPipelineEnvironment/mtarFilePath"] = "test.mtar"
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(&utils, &uploader, &options)
		assert.EqualError(t, err,
			"failed to parse contents of the project descriptor file 'mta.yaml'")
		assert.Equal(t, 0, len(uploader.GetArtifacts()))
		assert.Equal(t, 0, len(uploader.uploadedArtifacts))
	})
	t.Run("Uploading MTA project fails due invalid version in YAML content", func(t *testing.T) {
		utils := newMockUtilsBundle(true, false, false)
		utils.files["mta.yaml"] = []byte(testMtaYmlNoVersion)
		utils.cpe[".pipeline/commonPipelineEnvironment/mtarFilePath"] = "test.mtar"
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(&utils, &uploader, &options)
		assert.EqualError(t, err,
			"the project descriptor file 'mta.yaml' has an invalid version: version must not be empty")
		assert.Equal(t, 0, len(uploader.GetArtifacts()))
		assert.Equal(t, 0, len(uploader.uploadedArtifacts))
	})
	t.Run("Test uploading mta.yaml project fails due to missing mtar file", func(t *testing.T) {
		utils := newMockUtilsBundle(true, false, false)
		utils.files["mta.yaml"] = testMtaYml
		utils.cpe[".pipeline/commonPipelineEnvironment/mtarFilePath"] = "test.mtar"
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(&utils, &uploader, &options)
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
		utils := newMockUtilsBundle(true, false, false)
		utils.files["mta.yaml"] = testMtaYml
		utils.files["test.mtar"] = []byte("contentsOfMtar")
		utils.cpe[".pipeline/commonPipelineEnvironment/mtarFilePath"] = "test.mtar"
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(&utils, &uploader, &options)
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
		utils := newMockUtilsBundle(true, false, false)
		utils.files["mta.yml"] = testMtaYml
		utils.files["test.mtar"] = []byte("contentsOfMtar")
		utils.cpe[".pipeline/commonPipelineEnvironment/mtarFilePath"] = "test.mtar"
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(&utils, &uploader, &options)
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
	t.Run("Uploading MTA project fails without info", func(t *testing.T) {
		utils := newMockUtilsBundle(false, true, false)
		uploader := mockUploader{}
		options := createOptions()

		err := uploadArtifacts(&utils, &uploader, &options, false)
		assert.EqualError(t, err, "no group ID was provided, or could be established from project files")
	})
	t.Run("Uploading MTA project fails without any artifacts", func(t *testing.T) {
		utils := newMockUtilsBundle(false, true, false)
		uploader := mockUploader{}
		options := createOptions()

		_ = uploader.SetInfo(options.GroupID, "some.id", "3.0")

		err := uploadArtifacts(&utils, &uploader, &options, false)
		assert.EqualError(t, err, "no artifacts to upload")
	})
	t.Run("Uploading MTA project fails for unknown reasons", func(t *testing.T) {
		utils := newMockUtilsBundle(false, true, false)

		// Configure mocked execRunner to fail
		utils.execRunner.ShouldFailOnCommand = map[string]error{}
		utils.execRunner.ShouldFailOnCommand["mvn"] = fmt.Errorf("failed")

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

		err := uploadArtifacts(&utils, &uploader, &options, false)
		assert.EqualError(t, err, "uploading artifacts for ID 'some.id' failed: failed to run executable, command: '[mvn -Durl=http:// -DgroupId=my.group.id -Dversion=3.0 -DartifactId=some.id -Dfile=mta.yaml -Dpackaging=yaml -DgeneratePom=false -Dfiles=artifact.mtar -Dclassifiers= -Dtypes=yaml -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn --batch-mode "+deployGoal+"]', error: failed")
	})
	t.Run("Uploading bundle generates correct maven parameters", func(t *testing.T) {
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

		err := uploadArtifacts(&utils, &uploader, &options, false)
		assert.NoError(t, err, "expected upload as two bundles to work")
		assert.Equal(t, 1, len(utils.execRunner.Calls))

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
		assert.Equal(t, len(expectedParameters1), len(utils.execRunner.Calls[0].Params))
		assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: expectedParameters1}, utils.execRunner.Calls[0])
	})
}

func TestUploadNpmProjects(t *testing.T) {
	t.Run("Test uploading simple npm project", func(t *testing.T) {
		utils := newMockUtilsBundle(false, false, true)
		utils.files["package.json"] = testPackageJson
		uploader := mockUploader{}
		options := createOptions()
		options.User = "admin"
		options.Password = "admin123"

		err := runNexusUpload(&utils, &uploader, &options)
		assert.NoError(t, err, "expected npm upload to work")

		assert.Equal(t, "localhost:8081/repository/npm-repo/", uploader.GetNpmRepoURL())

		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"publish"}}, utils.execRunner.Calls[0])
		assert.Equal(t, []string{"npm_config_registry=http://localhost:8081/repository/npm-repo/", "npm_config_email=project-piper@no-reply.com", "npm_config__auth=YWRtaW46YWRtaW4xMjM="}, utils.execRunner.Env)
	})
}

func TestUploadMavenProjects(t *testing.T) {
	t.Run("Uploading Maven project fails due to missing pom.xml", func(t *testing.T) {
		utils := newMockUtilsBundle(false, true, false)
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(&utils, &uploader, &options)
		assert.EqualError(t, err, "pom.xml not found")
		assert.Equal(t, 0, len(uploader.uploadedArtifacts))
	})
	t.Run("Test uploading Maven project with POM packaging works", func(t *testing.T) {
		utils := newMockUtilsBundle(false, true, false)
		utils.setProperty("pom.xml", "project.version", "1.0")
		utils.setProperty("pom.xml", "project.groupId", "com.mycompany.app")
		utils.setProperty("pom.xml", "project.artifactId", "my-app")
		utils.setProperty("pom.xml", "project.packaging", "pom")
		utils.setProperty("pom.xml", "project.build.finalName", "my-app-1.0")
		utils.files["pom.xml"] = testPomXml
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(&utils, &uploader, &options)
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
		utils := newMockUtilsBundle(false, true, false)
		utils.setProperty("pom.xml", "project.version", "1.0")
		utils.setProperty("pom.xml", "project.groupId", "com.mycompany.app")
		utils.setProperty("pom.xml", "project.artifactId", "my-app")
		utils.setProperty("pom.xml", "project.packaging", "jar")
		utils.setProperty("pom.xml", "project.build.finalName", "my-app-1.0")
		utils.files["pom.xml"] = testPomXml
		utils.files["target/dummy"] = []byte("contentsOfJar") // causes "target" folder to exist
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(&utils, &uploader, &options)
		assert.EqualError(t, err, "target artifact not found for packaging 'jar'")
		assert.Equal(t, 0, len(uploader.uploadedArtifacts))
	})
	t.Run("Test uploading Maven project with JAR packaging works", func(t *testing.T) {
		utils := newMockUtilsBundle(false, true, false)
		utils.setProperty("pom.xml", "project.version", "1.0")
		utils.setProperty("pom.xml", "project.groupId", "com.mycompany.app")
		utils.setProperty("pom.xml", "project.artifactId", "my-app")
		utils.setProperty("pom.xml", "project.packaging", "jar")
		utils.setProperty("pom.xml", "project.build.finalName", "my-app-1.0")
		utils.files["pom.xml"] = testPomXml
		utils.files["target/my-app-1.0.jar"] = []byte("contentsOfJar")
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(&utils, &uploader, &options)
		assert.NoError(t, err, "expected Maven upload to work")

		assert.Equal(t, "1.0", uploader.GetArtifactsVersion())
		assert.Equal(t, "my-app", uploader.GetArtifactsID())

		artifacts := uploader.uploadedArtifacts
		if assert.Equal(t, 2, len(artifacts)) {
			assert.Equal(t, "pom.xml", artifacts[0].File)
			assert.Equal(t, "pom", artifacts[0].Type)

			assert.Equal(t, "target/my-app-1.0.jar", artifacts[1].File)
			assert.Equal(t, "jar", artifacts[1].Type)
		}
	})
	t.Run("Test uploading Maven project with fall-back to JAR packaging works", func(t *testing.T) {
		utils := newMockUtilsBundle(false, true, false)
		utils.setProperty("pom.xml", "project.version", "1.0")
		utils.setProperty("pom.xml", "project.groupId", "com.mycompany.app")
		utils.setProperty("pom.xml", "project.artifactId", "my-app")
		utils.setProperty("pom.xml", "project.packaging", "<empty>")
		utils.setProperty("pom.xml", "project.build.finalName", "my-app-1.0")
		utils.files["pom.xml"] = testPomXml
		utils.files["target/my-app-1.0.jar"] = []byte("contentsOfJar")
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(&utils, &uploader, &options)
		assert.NoError(t, err, "expected Maven upload to work")
		assert.Equal(t, "1.0", uploader.GetArtifactsVersion())
		assert.Equal(t, "my-app", uploader.GetArtifactsID())

		artifacts := uploader.uploadedArtifacts
		if assert.Equal(t, 2, len(artifacts)) {
			assert.Equal(t, "pom.xml", artifacts[0].File)
			assert.Equal(t, "pom", artifacts[0].Type)

			assert.Equal(t, "target/my-app-1.0.jar", artifacts[1].File)
			assert.Equal(t, "jar", artifacts[1].Type)
		}
	})
	t.Run("Test uploading Maven project with fall-back to group id from parameters works", func(t *testing.T) {
		utils := newMockUtilsBundle(false, true, false)
		utils.setProperty("pom.xml", "project.version", "1.0")
		utils.setProperty("pom.xml", "project.artifactId", "my-app")
		utils.setProperty("pom.xml", "project.packaging", "pom")
		utils.setProperty("pom.xml", "project.build.finalName", "my-app-1.0")
		utils.files["pom.xml"] = testPomXml
		uploader := mockUploader{}
		options := createOptions()
		options.GroupID = "awesome.group"

		err := runNexusUpload(&utils, &uploader, &options)
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
		utils := newMockUtilsBundle(false, true, false)
		utils.setProperty("pom.xml", "project.version", "1.0")
		utils.setProperty("pom.xml", "project.groupId", "awesome.group")
		utils.setProperty("pom.xml", "project.artifactId", "my-app")
		utils.setProperty("pom.xml", "project.packaging", "jar")
		utils.files["pom.xml"] = testPomXml
		utils.files["target/my-app-1.0.jar"] = []byte("contentsOfJar")
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(&utils, &uploader, &options)
		assert.NoError(t, err, "expected Maven upload to work")

		assert.Equal(t, "localhost:8081/repository/maven-releases/",
			uploader.GetMavenRepoURL())
		assert.Equal(t, "1.0", uploader.GetArtifactsVersion())
		assert.Equal(t, "my-app", uploader.GetArtifactsID())

		artifacts := uploader.uploadedArtifacts
		if assert.Equal(t, 2, len(artifacts)) {
			assert.Equal(t, "pom.xml", artifacts[0].File)
			assert.Equal(t, "pom", artifacts[0].Type)
			assert.Equal(t, "target/my-app-1.0.jar", artifacts[1].File)
			assert.Equal(t, "jar", artifacts[1].Type)
		}
	})
	t.Run("Test uploading Maven project with application module and finalName works", func(t *testing.T) {
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
		utils.files["pom.xml"] = testPomXml
		utils.files["application/pom.xml"] = testPomXml
		utils.files["application/target/final-artifact.war"] = []byte("contentsOfJar")
		utils.files["application/target/final-artifact-classes.jar"] = []byte("contentsOfClassesJar")
		utils.files["integration-tests/pom.xml"] = testPomXml
		utils.files["integration-tests/target/final-artifact-integration-tests.jar"] = []byte("contentsOfJar")
		utils.files["unit-tests/pom.xml"] = testPomXml
		utils.files["unit-tests/target/final-artifact-unit-tests.jar"] = []byte("contentsOfJar")
		utils.files["performance-tests/pom.xml"] = testPomXml
		uploader := mockUploader{}
		options := createOptions()

		err := runNexusUpload(&utils, &uploader, &options)
		assert.NoError(t, err, "expected upload of maven project with application module to succeed")
		assert.Equal(t, "1.0", uploader.GetArtifactsVersion())
		assert.Equal(t, "my-app-app", uploader.GetArtifactsID())

		artifacts := uploader.uploadedArtifacts
		if assert.Equal(t, 4, len(artifacts)) {
			assert.Equal(t, "pom.xml", artifacts[0].File)
			assert.Equal(t, "pom", artifacts[0].Type)

			assert.Equal(t, "application/pom.xml", artifacts[1].File)
			assert.Equal(t, "pom", artifacts[1].Type)

			assert.Equal(t, "application/target/final-artifact.war", artifacts[2].File)
			assert.Equal(t, "war", artifacts[2].Type)

			assert.Equal(t, "application/target/final-artifact-classes.jar", artifacts[3].File)
			assert.Equal(t, "jar", artifacts[3].Type)
		}
		if assert.Equal(t, 2, len(utils.execRunner.Calls)) {
			expectedParameters1 := []string{
				"-Durl=http://localhost:8081/repository/maven-releases/",
				"-DgroupId=com.mycompany.app",
				"-Dversion=1.0",
				"-DartifactId=my-app",
				"-Dfile=pom.xml",
				"-Dpackaging=pom",
				"-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn",
				"--batch-mode",
				deployGoal}
			assert.Equal(t, len(expectedParameters1), len(utils.execRunner.Calls[0].Params))
			assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: expectedParameters1}, utils.execRunner.Calls[0])

			expectedParameters2 := []string{
				"-Durl=http://localhost:8081/repository/maven-releases/",
				"-DgroupId=com.mycompany.app",
				"-Dversion=1.0",
				"-DartifactId=my-app-app",
				"-Dfile=application/pom.xml",
				"-Dpackaging=pom",
				"-Dfiles=application/target/final-artifact.war,application/target/final-artifact-classes.jar",
				"-Dclassifiers=,classes",
				"-Dtypes=war,jar",
				"-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn",
				"--batch-mode",
				deployGoal}
			assert.Equal(t, len(expectedParameters2), len(utils.execRunner.Calls[1].Params))
			assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: expectedParameters2}, utils.execRunner.Calls[1])
		}
	})
	t.Run("Write credentials settings", func(t *testing.T) {
		utils := newMockUtilsBundle(false, true, false)
		utils.setProperty("pom.xml", "project.version", "1.0")
		utils.setProperty("pom.xml", "project.groupId", "com.mycompany.app")
		utils.setProperty("pom.xml", "project.artifactId", "my-app")
		utils.setProperty("pom.xml", "project.packaging", "pom")
		utils.setProperty("pom.xml", "project.build.finalName", "my-app-1.0")
		utils.files["pom.xml"] = testPomXml
		uploader := mockUploader{}
		options := createOptions()
		options.User = "admin"
		options.Password = "admin123"

		err := runNexusUpload(&utils, &uploader, &options)
		assert.NoError(t, err, "expected Maven upload to work")

		assert.Equal(t, 1, len(utils.execRunner.Calls))
		expectedParameters1 := []string{
			"--settings",
			settingsPath,
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
		assert.Equal(t, len(expectedParameters1), len(utils.execRunner.Calls[0].Params))
		assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: expectedParameters1}, utils.execRunner.Calls[0])

		expectedEnv := []string{"NEXUS_username=admin", "NEXUS_password=admin123"}
		assert.Equal(t, 2, len(utils.execRunner.Env))
		assert.Equal(t, expectedEnv, utils.execRunner.Env)

		assert.Nil(t, utils.files[settingsPath])
		assert.NotNil(t, utils.removedFiles[settingsPath])
	})
}

func TestSetupNexusCredentialsSettingsFile(t *testing.T) {
	utils := newMockUtilsBundle(false, true, false)
	options := nexusUploadOptions{User: "admin", Password: "admin123"}
	mavenOptions := maven.ExecuteOptions{}
	settingsPath, err := setupNexusCredentialsSettingsFile(&utils, &options, &mavenOptions)

	assert.NoError(t, err, "expected setting up credentials settings.xml to work")
	assert.Equal(t, 0, len(utils.execRunner.Calls))
	expectedEnv := []string{"NEXUS_username=admin", "NEXUS_password=admin123"}
	assert.Equal(t, 2, len(utils.execRunner.Env))
	assert.Equal(t, expectedEnv, utils.execRunner.Env)

	assert.True(t, settingsPath != "")
	assert.NotNil(t, utils.files[settingsPath])
}
