package cmd

import (
	"errors"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
)

type mtaBuildTestUtilsBundle struct {
	*mock.ExecMockRunner
	*mock.FilesMock
	projectSettingsFile            string
	globalSettingsFile             string
	registryUsedInSetNpmRegistries string
}

func (m *mtaBuildTestUtilsBundle) SetNpmRegistries(defaultNpmRegistry string) error {
	m.registryUsedInSetNpmRegistries = defaultNpmRegistry
	return nil
}

func (m *mtaBuildTestUtilsBundle) InstallAllDependencies(defaultNpmRegistry string) error {
	return errors.New("Test should not install dependencies.") //TODO implement test
}

func (m *mtaBuildTestUtilsBundle) DownloadAndCopySettingsFiles(globalSettingsFile string, projectSettingsFile string) error {
	m.projectSettingsFile = projectSettingsFile
	m.globalSettingsFile = globalSettingsFile
	return nil
}

func (m *mtaBuildTestUtilsBundle) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	return errors.New("Test should not download files.")
}

func newMtaBuildTestUtilsBundle() *mtaBuildTestUtilsBundle {
	utilsBundle := mtaBuildTestUtilsBundle{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return &utilsBundle
}

func TestMarBuild(t *testing.T) {

	cpe := mtaBuildCommonPipelineEnvironment{}

	t.Run("Application name not set", func(t *testing.T) {

		utilsMock := newMtaBuildTestUtilsBundle()
		options := mtaBuildOptions{}

		err := runMtaBuild(options, &cpe, utilsMock)

		assert.NotNil(t, err)
		assert.Equal(t, "'mta.yaml' not found in project sources and 'applicationName' not provided as parameter - cannot generate 'mta.yaml' file", err.Error())

	})

	t.Run("Provide default npm registry", func(t *testing.T) {

		utilsMock := newMtaBuildTestUtilsBundle()
		options := mtaBuildOptions{ApplicationName: "myApp", Platform: "CF", DefaultNpmRegistry: "https://example.org/npm", MtarName: "myName"}

		utilsMock.AddFile("package.json", []byte("{\"name\": \"myName\", \"version\": \"1.2.3\"}"))

		err := runMtaBuild(options, &cpe, utilsMock)

		assert.Nil(t, err)

		assert.Equal(t, "https://example.org/npm", utilsMock.registryUsedInSetNpmRegistries)
	})

	t.Run("Package json does not exist", func(t *testing.T) {

		utilsMock := newMtaBuildTestUtilsBundle()

		options := mtaBuildOptions{ApplicationName: "myApp"}

		err := runMtaBuild(options, &cpe, utilsMock)

		assert.NotNil(t, err)

		assert.Equal(t, "package.json file does not exist", err.Error())

	})

	t.Run("Write yaml file", func(t *testing.T) {

		utilsMock := newMtaBuildTestUtilsBundle()

		options := mtaBuildOptions{ApplicationName: "myApp", Platform: "CF", MtarName: "myName"}

		utilsMock.AddFile("package.json", []byte("{\"name\": \"myName\", \"version\": \"1.2.3\"}"))

		err := runMtaBuild(options, &cpe, utilsMock)

		assert.Nil(t, err)

		type MtaResult struct {
			Version    string
			ID         string `yaml:"ID,omitempty"`
			Parameters map[string]string
			Modules    []struct {
				Name       string
				Type       string
				Parameters map[string]interface{}
			}
		}

		assert.True(t, utilsMock.HasWrittenFile("mta.yaml"))

		var result MtaResult
		mtaContent, _ := utilsMock.FileRead("mta.yaml")
		yaml.Unmarshal(mtaContent, &result)

		assert.Equal(t, "myName", result.ID)
		assert.Equal(t, "1.2.3", result.Version)
		assert.Equal(t, "myApp", result.Modules[0].Name)
		assert.Regexp(t, "^1\\.2\\.3-[\\d]{14}$", result.Modules[0].Parameters["version"])
		assert.Equal(t, "myApp", result.Modules[0].Parameters["name"])

	})

	t.Run("Dont write mta yaml file when already present no timestamp placeholder", func(t *testing.T) {

		utilsMock := newMtaBuildTestUtilsBundle()

		options := mtaBuildOptions{ApplicationName: "myApp"}

		utilsMock.AddFile("package.json", []byte("{\"name\": \"myName\", \"version\": \"1.2.3\"}"))
		utilsMock.AddFile("mta.yaml", []byte("already there"))

		_ = runMtaBuild(options, &cpe, utilsMock)

		assert.False(t, utilsMock.HasWrittenFile("mta.yaml"))
	})

	t.Run("Write mta yaml file when already present with timestamp placeholder", func(t *testing.T) {

		utilsMock := newMtaBuildTestUtilsBundle()

		options := mtaBuildOptions{ApplicationName: "myApp"}

		utilsMock.AddFile("package.json", []byte("{\"name\": \"myName\", \"version\": \"1.2.3\"}"))
		utilsMock.AddFile("mta.yaml", []byte("already there with-${timestamp}"))

		_ = runMtaBuild(options, &cpe, utilsMock)

		assert.True(t, utilsMock.HasWrittenFile("mta.yaml"))
	})

	t.Run("Mta build mbt toolset", func(t *testing.T) {

		utilsMock := newMtaBuildTestUtilsBundle()

		cpe.mtarFilePath = ""

		options := mtaBuildOptions{ApplicationName: "myApp", Platform: "CF", MtarName: "myName.mtar"}

		utilsMock.AddFile("package.json", []byte("{\"name\": \"myName\", \"version\": \"1.2.3\"}"))

		err := runMtaBuild(options, &cpe, utilsMock)

		assert.Nil(t, err)

		if assert.Len(t, utilsMock.Calls, 1) {
			assert.Equal(t, "mbt", utilsMock.Calls[0].Exec)
			assert.Equal(t, []string{"build", "--mtar", "myName.mtar", "--platform", "CF", "--target", "./"}, utilsMock.Calls[0].Params)
		}
		assert.Equal(t, "myName.mtar", cpe.mtarFilePath)
	})

	t.Run("M2Path related tests", func(t *testing.T) {
		t.Run("Mta build mbt toolset with m2Path", func(t *testing.T) {

			utilsMock := newMtaBuildTestUtilsBundle()
			utilsMock.CurrentDir = "root_folder/workspace"
			cpe.mtarFilePath = ""

			options := mtaBuildOptions{ApplicationName: "myApp", Platform: "CF", MtarName: "myName.mtar", M2Path: ".pipeline/local_repo"}

			utilsMock.AddFile("mta.yaml", []byte("ID: \"myNameFromMtar\""))

			err := runMtaBuild(options, &cpe, utilsMock)

			assert.Nil(t, err)
			assert.Contains(t, utilsMock.Env, filepath.FromSlash("MAVEN_OPTS=-Dmaven.repo.local=/root_folder/workspace/.pipeline/local_repo"))
		})
	})

	t.Run("Settings file releatd tests", func(t *testing.T) {

		t.Run("Copy global settings file", func(t *testing.T) {

			utilsMock := newMtaBuildTestUtilsBundle()
			utilsMock.AddFile("mta.yaml", []byte("ID: \"myNameFromMtar\""))

			options := mtaBuildOptions{ApplicationName: "myApp", GlobalSettingsFile: "/opt/maven/settings.xml", Platform: "CF", MtarName: "myName"}

			err := runMtaBuild(options, &cpe, utilsMock)

			assert.Nil(t, err)

			assert.Equal(t, "/opt/maven/settings.xml", utilsMock.globalSettingsFile)
			assert.Equal(t, "", utilsMock.projectSettingsFile)
		})

		t.Run("Copy project settings file", func(t *testing.T) {

			utilsMock := newMtaBuildTestUtilsBundle()
			utilsMock.AddFile("mta.yaml", []byte("ID: \"myNameFromMtar\""))

			options := mtaBuildOptions{ApplicationName: "myApp", ProjectSettingsFile: "/my/project/settings.xml", Platform: "CF", MtarName: "myName"}

			err := runMtaBuild(options, &cpe, utilsMock)

			assert.Nil(t, err)

			assert.Equal(t, "/my/project/settings.xml", utilsMock.projectSettingsFile)
			assert.Equal(t, "", utilsMock.globalSettingsFile)
		})
	})
}
