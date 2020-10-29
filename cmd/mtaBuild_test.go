package cmd

import (
	"github.com/SAP/jenkins-library/pkg/npm"
	"os"
	"testing"

	"fmt"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
)

func TestMarBuild(t *testing.T) {

	cpe := mtaBuildCommonPipelineEnvironment{}
	httpClient := piperhttp.Client{}

	t.Run("Application name not set", func(t *testing.T) {

		e := mock.ExecMockRunner{}

		options := mtaBuildOptions{}

		fileUtils := MtaTestFileUtilsMock{}

		err := runMtaBuild(options, &cpe, &e, &fileUtils, &httpClient, newNpmExecutor(&e))

		assert.NotNil(t, err)
		assert.Equal(t, "'mta.yaml' not found in project sources and 'applicationName' not provided as parameter - cannot generate 'mta.yaml' file", err.Error())

	})

	t.Run("Provide default npm registry", func(t *testing.T) {

		e := mock.ExecMockRunner{}
		e.StdoutReturn = map[string]string{"npm config get registry": "undefined"}

		options := mtaBuildOptions{ApplicationName: "myApp", MtaBuildTool: "classic", BuildTarget: "CF", DefaultNpmRegistry: "https://example.org/npm", MtarName: "myName"}

		existingFiles := make(map[string]string)
		existingFiles["package.json"] = "{\"name\": \"myName\", \"version\": \"1.2.3\"}"
		fileUtils := MtaTestFileUtilsMock{existingFiles: existingFiles}

		npmExecutor := newNpmExecutor(&e)
		npmExecutor.Options = npm.ExecutorOptions{DefaultNpmRegistry: options.DefaultNpmRegistry}

		err := runMtaBuild(options, &cpe, &e, &fileUtils, &httpClient, npmExecutor)

		assert.Nil(t, err)

		if assert.Len(t, e.Calls, 3) { // the second (unchecked) entry is the mta call
			assert.Equal(t, "npm", e.Calls[1].Exec)
			assert.Equal(t, []string{"config", "set", "registry", "https://example.org/npm"}, e.Calls[1].Params)
		}
	})

	t.Run("Package json does not exist", func(t *testing.T) {

		e := mock.ExecMockRunner{}

		options := mtaBuildOptions{ApplicationName: "myApp"}

		fileUtils := MtaTestFileUtilsMock{}

		err := runMtaBuild(options, &cpe, &e, &fileUtils, &httpClient, newNpmExecutor(&e))

		assert.NotNil(t, err)

		assert.Equal(t, "package.json file does not exist", err.Error())

	})

	t.Run("Write yaml file", func(t *testing.T) {

		e := mock.ExecMockRunner{}

		options := mtaBuildOptions{ApplicationName: "myApp", MtaBuildTool: "classic", BuildTarget: "CF", MtarName: "myName"}

		existingFiles := make(map[string]string)
		existingFiles["package.json"] = "{\"name\": \"myName\", \"version\": \"1.2.3\"}"

		fileUtils := MtaTestFileUtilsMock{existingFiles: existingFiles}

		err := runMtaBuild(options, &cpe, &e, &fileUtils, &httpClient, newNpmExecutor(&e))

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

		assert.NotEmpty(t, fileUtils.writtenFiles["mta.yaml"])

		var result MtaResult
		yaml.Unmarshal([]byte(fileUtils.writtenFiles["mta.yaml"]), &result)

		assert.Equal(t, "myName", result.ID)
		assert.Equal(t, "1.2.3", result.Version)
		assert.Equal(t, "myApp", result.Modules[0].Name)
		assert.Equal(t, result.Modules[0].Parameters["version"], "1.2.3-${timestamp}")
		assert.Equal(t, "myApp", result.Modules[0].Parameters["name"])

	})

	t.Run("Dont write mta yaml file when already present no timestamp placeholder", func(t *testing.T) {

		e := mock.ExecMockRunner{}

		options := mtaBuildOptions{ApplicationName: "myApp", MtaBuildTool: "classic", BuildTarget: "CF"}

		existingFiles := make(map[string]string)
		existingFiles["package.json"] = "{\"name\": \"myName\", \"version\": \"1.2.3\"}"
		existingFiles["mta.yaml"] = "already there"
		fileUtils := MtaTestFileUtilsMock{existingFiles: existingFiles}

		_ = runMtaBuild(options, &cpe, &e, &fileUtils, &httpClient, newNpmExecutor(&e))

		assert.Empty(t, fileUtils.writtenFiles)
	})

	t.Run("Write mta yaml file when already present with timestamp placeholder", func(t *testing.T) {

		e := mock.ExecMockRunner{}

		options := mtaBuildOptions{ApplicationName: "myApp", MtaBuildTool: "classic", BuildTarget: "CF"}

		existingFiles := make(map[string]string)
		existingFiles["package.json"] = "{\"name\": \"myName\", \"version\": \"1.2.3\"}"
		existingFiles["mta.yaml"] = "already there with-${timestamp}"
		fileUtils := MtaTestFileUtilsMock{existingFiles: existingFiles}

		_ = runMtaBuild(options, &cpe, &e, &fileUtils, &httpClient, newNpmExecutor(&e))

		assert.NotEmpty(t, fileUtils.writtenFiles["mta.yaml"])
	})

	t.Run("Test mta build classic toolset", func(t *testing.T) {

		e := mock.ExecMockRunner{}

		options := mtaBuildOptions{ApplicationName: "myApp", MtaBuildTool: "classic", BuildTarget: "CF", MtarName: "myName.mtar"}

		cpe.mtarFilePath = ""

		existingFiles := make(map[string]string)
		existingFiles["package.json"] = "{\"name\": \"myName\", \"version\": \"1.2.3\"}"
		fileUtils := MtaTestFileUtilsMock{existingFiles: existingFiles}

		err := runMtaBuild(options, &cpe, &e, &fileUtils, &httpClient, newNpmExecutor(&e))

		assert.Nil(t, err)

		if assert.Len(t, e.Calls, 2) {
			assert.Equal(t, "java", e.Calls[1].Exec)
			assert.Equal(t, []string{"-jar", "mta.jar", "--mtar", "myName.mtar", "--build-target=CF", "build"}, e.Calls[1].Params)
		}

		assert.Equal(t, "myName.mtar", cpe.mtarFilePath)
	})

	t.Run("Test mta build classic toolset, mtarName from already existing mta.yaml", func(t *testing.T) {

		e := mock.ExecMockRunner{}

		options := mtaBuildOptions{ApplicationName: "myApp", MtaBuildTool: "classic", BuildTarget: "CF"}

		cpe.mtarFilePath = ""

		existingFiles := make(map[string]string)
		existingFiles["package.json"] = "{\"name\": \"myName\", \"version\": \"1.2.3\"}"
		existingFiles["mta.yaml"] = "ID: \"myNameFromMtar\""
		fileUtils := MtaTestFileUtilsMock{existingFiles: existingFiles}

		err := runMtaBuild(options, &cpe, &e, &fileUtils, &httpClient, newNpmExecutor(&e))

		assert.Nil(t, err)

		if assert.Len(t, e.Calls, 2) {
			assert.Equal(t, "java", e.Calls[1].Exec)
			assert.Equal(t, []string{"-jar", "mta.jar", "--mtar", "myNameFromMtar.mtar", "--build-target=CF", "build"}, e.Calls[1].Params)
		}
	})

	t.Run("Test mta build classic toolset with configured mta jar", func(t *testing.T) {

		e := mock.ExecMockRunner{}

		options := mtaBuildOptions{ApplicationName: "myApp", MtaBuildTool: "classic", BuildTarget: "CF", MtaJarLocation: "/opt/sap/mta/lib/mta.jar", MtarName: "myName.mtar"}

		existingFiles := make(map[string]string)
		existingFiles["package.json"] = "{\"name\": \"myName\", \"version\": \"1.2.3\"}"
		fileUtils := MtaTestFileUtilsMock{existingFiles: existingFiles}

		err := runMtaBuild(options, &cpe, &e, &fileUtils, &httpClient, newNpmExecutor(&e))

		assert.Nil(t, err)

		if assert.Len(t, e.Calls, 2) {
			assert.Equal(t, "java", e.Calls[1].Exec)
			assert.Equal(t, []string{"-jar", "/opt/sap/mta/lib/mta.jar", "--mtar", "myName.mtar", "--build-target=CF", "build"}, e.Calls[1].Params)
		}
	})

	t.Run("Mta build mbt toolset", func(t *testing.T) {

		e := mock.ExecMockRunner{}

		cpe.mtarFilePath = ""

		options := mtaBuildOptions{ApplicationName: "myApp", MtaBuildTool: "cloudMbt", Platform: "CF", MtarName: "myName.mtar"}

		existingFiles := make(map[string]string)
		existingFiles["package.json"] = "{\"name\": \"myName\", \"version\": \"1.2.3\"}"
		fileUtils := MtaTestFileUtilsMock{existingFiles: existingFiles}

		err := runMtaBuild(options, &cpe, &e, &fileUtils, &httpClient, newNpmExecutor(&e))

		assert.Nil(t, err)

		if assert.Len(t, e.Calls, 2) {
			assert.Equal(t, "mbt", e.Calls[1].Exec)
			assert.Equal(t, []string{"build", "--mtar", "myName.mtar", "--platform", "CF", "--target", "./"}, e.Calls[1].Params)
		}
		assert.Equal(t, "myName.mtar", cpe.mtarFilePath)
	})

	t.Run("M2Path related tests", func(t *testing.T) {
		t.Run("Mta build mbt toolset with m2Path", func(t *testing.T) {

			e := mock.ExecMockRunner{}

			cpe.mtarFilePath = ""

			options := mtaBuildOptions{ApplicationName: "myApp", MtaBuildTool: "cloudMbt", Platform: "CF", MtarName: "myName.mtar", M2Path: ".pipeline/local_repo"}

			existingFiles := make(map[string]string)
			existingFiles["mta.yaml"] = "ID: \"myNameFromMtar\""
			fileUtils := MtaTestFileUtilsMock{existingFiles: existingFiles}

			err := runMtaBuild(options, &cpe, &e, &fileUtils, &httpClient, newNpmExecutor(&e))

			assert.Nil(t, err)
			assert.Contains(t, e.Env, "MAVEN_OPTS=-Dmaven.repo.local=/root_folder/workspace/.pipeline/local_repo")
		})
	})

	t.Run("Settings file releatd tests", func(t *testing.T) {

		var projectSettingsFile string
		var globalSettingsFile string

		defer func() {
			downloadAndCopySettingsFiles = maven.DownloadAndCopySettingsFiles
		}()

		downloadAndCopySettingsFiles = func(
			globalSettings string,
			projectSettings string,
			fileUtils maven.FileUtils,
			httpClient maven.SettingsDownloadUtils) error {
			projectSettingsFile = projectSettings
			globalSettingsFile = globalSettings
			return nil
		}

		fileUtils := MtaTestFileUtilsMock{}
		fileUtils.existingFiles = make(map[string]string)
		fileUtils.existingFiles["package.json"] = "{\"name\": \"myName\", \"version\": \"1.2.3\"}"

		t.Run("Copy global settings file", func(t *testing.T) {

			defer func() {
				projectSettingsFile = ""
				globalSettingsFile = ""
			}()

			e := mock.ExecMockRunner{}

			options := mtaBuildOptions{ApplicationName: "myApp", GlobalSettingsFile: "/opt/maven/settings.xml", MtaBuildTool: "cloudMbt", Platform: "CF", MtarName: "myName"}

			err := runMtaBuild(options, &cpe, &e, &fileUtils, &httpClient, newNpmExecutor(&e))

			assert.Nil(t, err)

			assert.Equal(t, globalSettingsFile, "/opt/maven/settings.xml")
			assert.Equal(t, projectSettingsFile, "")
		})

		t.Run("Copy project settings file", func(t *testing.T) {

			defer func() {
				projectSettingsFile = ""
				globalSettingsFile = ""
			}()

			e := mock.ExecMockRunner{}

			options := mtaBuildOptions{ApplicationName: "myApp", ProjectSettingsFile: "/my/project/settings.xml", MtaBuildTool: "cloudMbt", Platform: "CF", MtarName: "myName"}

			err := runMtaBuild(options, &cpe, &e, &fileUtils, &httpClient, newNpmExecutor(&e))

			assert.Nil(t, err)

			assert.Equal(t, "/my/project/settings.xml", projectSettingsFile)
			assert.Equal(t, "", globalSettingsFile)
		})
	})
}

type MtaTestFileUtilsMock struct {
	existingFiles map[string]string
	writtenFiles  map[string]string
	copiedFiles   map[string]string
}

func (f *MtaTestFileUtilsMock) FileExists(path string) (bool, error) {

	if _, ok := f.existingFiles[path]; ok {
		return true, nil
	}
	return false, nil
}

func (f *MtaTestFileUtilsMock) Copy(src, dest string) (int64, error) {

	if f.copiedFiles == nil {
		f.copiedFiles = make(map[string]string)
	}
	f.copiedFiles[src] = dest

	return 0, nil
}

func (f *MtaTestFileUtilsMock) FileRead(path string) ([]byte, error) {
	return []byte(f.existingFiles[path]), nil
}

func (f *MtaTestFileUtilsMock) FileWrite(path string, content []byte, perm os.FileMode) error {

	if f.writtenFiles == nil {
		f.writtenFiles = make(map[string]string)
	}

	if _, ok := f.writtenFiles[path]; ok {
		delete(f.writtenFiles, path)
	}
	f.writtenFiles[path] = string(content)
	return nil
}

func (f *MtaTestFileUtilsMock) MkdirAll(path string, perm os.FileMode) error {
	return nil
}

func (f *MtaTestFileUtilsMock) Abs(path string) (string, error) {
	return "/root_folder/workspace/" + path, nil
}

func (f *MtaTestFileUtilsMock) Glob(pattern string) (matches []string, err error) {
	return nil, fmt.Errorf("not implemented. func is only present in order to fullfil the interface contract. Needs to be ajusted in case it gets used.")
}

func (f *MtaTestFileUtilsMock) Chmod(path string, mode os.FileMode) error {
	return fmt.Errorf("not implemented. func is only present in order to fullfil the interface contract. Needs to be ajusted in case it gets used.")
}

func newNpmExecutor(execRunner *mock.ExecMockRunner) *npm.Execute {
	utils := newNpmMockUtilsBundle()
	utils.execRunner = execRunner
	return &npm.Execute{Utils: &utils}
}
