package cmd

import (
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"os"
	"testing"
)

func TestMtaApplicationNameNotSet(t *testing.T) {

	options := mtaBuildOptions{}
	cpe := mtaBuildCommonPipelineEnvironment{}
	e := execMockRunner{}
	fileUtils := MtaTestFileUtilsMock{}
	httpClient := piperhttp.Client{}

	err := runMtaBuild(options, &cpe, &e, &fileUtils, &httpClient)

	if err == nil {
		t.Errorf("Error expected but not received.")
	}
	assert.Equal(t, "'mta.yaml' not found in project sources and 'applicationName' not provided as parameter - cannot generate 'mta.yaml' file", err.Error())
}

func TestProvideDefaultNpmRegistry(t *testing.T) {

	options := mtaBuildOptions{ApplicationName: "myApp", MtaBuildTool: "classic", BuildTarget: "CF", DefaultNpmRegistry: "https://example.org/npm"}
	cpe := mtaBuildCommonPipelineEnvironment{}
	e := execMockRunner{}

	existingFiles := make(map[string]string)
	existingFiles["package.json"] = "{\"name\": \"myName\", \"version\": \"1.2.3\"}"
	fileUtils := MtaTestFileUtilsMock{existingFiles: existingFiles}
	httpClient := piperhttp.Client{}

	err := runMtaBuild(options, &cpe, &e, &fileUtils, &httpClient)

	if err != nil {
		t.Fatalf("Error received but not expected: '%s'", err.Error())
	}

	assert.Equal(t, "npm", e.calls[0].exec)
	assert.Equal(t, []string{"config", "set", "registry", "https://example.org/npm"}, e.calls[0].params)

}

func TestMtaPackageJsonDoesNotExist(t *testing.T) {

	options := mtaBuildOptions{ApplicationName: "myApp"}
	cpe := mtaBuildCommonPipelineEnvironment{}
	e := execMockRunner{}

	fileUtils := MtaTestFileUtilsMock{}
	httpClient := piperhttp.Client{}

	err := runMtaBuild(options, &cpe, &e, &fileUtils, &httpClient)

	if err == nil {
		t.Errorf("Error expected but not received.")
	}
	assert.Equal(t, "package.json file does not exist", err.Error())
}

func TestWriteMtaYamlFile(t *testing.T) {

	options := mtaBuildOptions{ApplicationName: "myApp", MtaBuildTool: "classic", BuildTarget: "CF"}
	cpe := mtaBuildCommonPipelineEnvironment{}
	e := execMockRunner{}

	existingFiles := make(map[string]string)
	existingFiles["package.json"] = "{\"name\": \"myName\", \"version\": \"1.2.3\"}"
	fileUtils := MtaTestFileUtilsMock{existingFiles: existingFiles}
	httpClient := piperhttp.Client{}

	runMtaBuild(options, &cpe, &e, &fileUtils, &httpClient)

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
	assert.NotContains(t, "${timestamp}", result.Modules[0].Parameters["version"])
}

func TestDontWriteMtaYamlFileWhenAlreadyPresentNoTimestampPlaceholder(t *testing.T) {

	options := mtaBuildOptions{ApplicationName: "myApp", MtaBuildTool: "classic", BuildTarget: "CF"}
	cpe := mtaBuildCommonPipelineEnvironment{}
	e := execMockRunner{}

	existingFiles := make(map[string]string)
	existingFiles["package.json"] = "{\"name\": \"myName\", \"version\": \"1.2.3\"}"
	existingFiles["mta.yaml"] = "already there"
	fileUtils := MtaTestFileUtilsMock{existingFiles: existingFiles}
	httpClient := piperhttp.Client{}

	runMtaBuild(options, &cpe, &e, &fileUtils, &httpClient)

	assert.Empty(t, fileUtils.writtenFiles)
}

func TestWriteMtaYamlFileWhenAlreadyPresentWithTimestampPlaceholder(t *testing.T) {

	options := mtaBuildOptions{ApplicationName: "myApp", MtaBuildTool: "classic", BuildTarget: "CF"}
	cpe := mtaBuildCommonPipelineEnvironment{}
	e := execMockRunner{}

	existingFiles := make(map[string]string)
	existingFiles["package.json"] = "{\"name\": \"myName\", \"version\": \"1.2.3\"}"
	existingFiles["mta.yaml"] = "already there with-${timestamp}"
	fileUtils := MtaTestFileUtilsMock{existingFiles: existingFiles}
	httpClient := piperhttp.Client{}

	runMtaBuild(options, &cpe, &e, &fileUtils, &httpClient)

	assert.NotEmpty(t, fileUtils.writtenFiles["mta.yaml"])
}

func TestMtaBuildClassicToolset(t *testing.T) {

	options := mtaBuildOptions{ApplicationName: "myApp", MtaBuildTool: "classic", BuildTarget: "CF"}
	cpe := mtaBuildCommonPipelineEnvironment{}
	e := execMockRunner{}

	existingFiles := make(map[string]string)
	existingFiles["package.json"] = "{\"name\": \"myName\", \"version\": \"1.2.3\"}"
	fileUtils := MtaTestFileUtilsMock{existingFiles: existingFiles}
	httpClient := piperhttp.Client{}

	err := runMtaBuild(options, &cpe, &e, &fileUtils, &httpClient)

	if err != nil {
		t.Errorf("No Error expected but error received: '%s'", err.Error())
	}

	assert.Equal(t, "java", e.calls[0].exec)
	assert.Equal(t, []string{"-jar", "mta.jar", "--mtar", "myName.mtar", "--build-target=CF"}, e.calls[0].params)
}

func TestMtaBuildClassicToolsetWithConfiguredMtaJar(t *testing.T) {

	options := mtaBuildOptions{ApplicationName: "myApp", MtaBuildTool: "classic", BuildTarget: "CF", MtaJarLocation: "/opt/sap/mta/lib/mta.jar"}
	cpe := mtaBuildCommonPipelineEnvironment{}
	e := execMockRunner{}

	existingFiles := make(map[string]string)
	existingFiles["package.json"] = "{\"name\": \"myName\", \"version\": \"1.2.3\"}"
	fileUtils := MtaTestFileUtilsMock{existingFiles: existingFiles}
	httpClient := piperhttp.Client{}

	err := runMtaBuild(options, &cpe, &e, &fileUtils, &httpClient)

	if err != nil {
		t.Errorf("No Error expected but error received: '%s'", err.Error())
	}

	assert.Equal(t, "java", e.calls[0].exec)
	assert.Equal(t, []string{"-jar", "/opt/sap/mta/lib/mta.jar", "--mtar", "myName.mtar", "--build-target=CF"}, e.calls[0].params)
}
func TestMtaBuildMbtToolset(t *testing.T) {

	options := mtaBuildOptions{ApplicationName: "myApp", MtaBuildTool: "cloudMbt", Platform: "CF"}
	cpe := mtaBuildCommonPipelineEnvironment{}
	e := execMockRunner{}

	existingFiles := make(map[string]string)
	existingFiles["package.json"] = "{\"name\": \"myName\", \"version\": \"1.2.3\"}"
	fileUtils := MtaTestFileUtilsMock{existingFiles: existingFiles}
	httpClient := piperhttp.Client{}

	err := runMtaBuild(options, &cpe, &e, &fileUtils, &httpClient)

	if err != nil {
		t.Fatalf("No Error expected but error received: '%s'", err.Error())

	}

	assert.Equal(t, "mbt", e.calls[0].exec)
	assert.Equal(t, []string{"build", "--mtar", "myName.mtar", "--platform", "CF", "--target", "./"}, e.calls[0].params)

}

func TestCopyGlobalSettingsFile(t *testing.T) {

	// Revisit: make independent of existing M2_HOME

	options := mtaBuildOptions{GlobalSettingsFile: "/opt/maven/settings.xml", MtaBuildTool: "cloudMbt", Platform: "CF"}
	cpe := mtaBuildCommonPipelineEnvironment{}
	e := execMockRunner{}
	fileUtils := MtaTestFileUtilsMock{}
	fileUtils.existingFiles = make(map[string]string)
	fileUtils.existingFiles["mta.yaml"] = "already there"
	httpClient := piperhttp.Client{}

	err := runMtaBuild(options, &cpe, &e, &fileUtils, &httpClient)

	if err != nil {
		t.Fatalf("ERR: %s" + err.Error())
	}

	assert.NotEmpty(t, fileUtils.copiedFiles["/opt/maven/settings.xml"])
}

func TestCopyProjectSettingsFile(t *testing.T) {

	// Revisit: make independent of existing M2_HOME

	options := mtaBuildOptions{ProjectSettingsFile: "/my/project/settings.xml", MtaBuildTool: "cloudMbt", Platform: "CF"}
	cpe := mtaBuildCommonPipelineEnvironment{}
	e := execMockRunner{}
	fileUtils := MtaTestFileUtilsMock{}
	fileUtils.existingFiles = make(map[string]string)
	fileUtils.existingFiles["mta.yaml"] = "already there"
	httpClient := piperhttp.Client{}

	err := runMtaBuild(options, &cpe, &e, &fileUtils, &httpClient)

	if err != nil {
		t.Fatalf("ERR: %s" + err.Error())
	}

	assert.NotEmpty(t, fileUtils.copiedFiles["/my/project/settings.xml"])
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