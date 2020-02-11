package cmd

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"os"
	"gopkg.in/yaml.v2"
)

func TestMtaApplicationNameNotSet(t *testing.T) {

	options := mtaBuildOptions{}
	cpe := mtaBuildCommonPipelineEnvironment{}
	e := execMockRunner{}
	fileUtils := MtaTestFileUtilsMock{}

	err := runMtaBuild(options, &cpe, &e, &fileUtils)

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

	err := runMtaBuild(options, &cpe, &e, &fileUtils)

	if err != nil {
		t.Fatalf("Error received but not expected: '%s'", err.Error())
	}

	assert.Equal(t, "npm", e.calls[0].exec)
	assert.Equal(t, []string {"config",  "set", "registry", "https://example.org/npm"}, e.calls[0].params)

}

func TestMtaPackageJsonDoesNotExist(t *testing.T) {

	options := mtaBuildOptions{ApplicationName: "myApp"}
	cpe := mtaBuildCommonPipelineEnvironment{}
	e := execMockRunner{}

	fileUtils := MtaTestFileUtilsMock{}

	err := runMtaBuild(options, &cpe, &e, &fileUtils)

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

	runMtaBuild(options, &cpe, &e, &fileUtils)

	type MtaResult struct {
		Version string
		ID string `yaml:"ID,omitempty"`
		Parameters map[string]string
		Modules []struct {
			Name string
			Type string
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

	runMtaBuild(options, &cpe, &e, &fileUtils)

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

	runMtaBuild(options, &cpe, &e, &fileUtils)

	assert.NotEmpty(t, fileUtils.writtenFiles["mta.yaml"])
}

func TestMtaBuildClassicToolset(t *testing.T) {

	options := mtaBuildOptions{ApplicationName: "myApp", MtaBuildTool: "classic", BuildTarget: "CF"}
	cpe := mtaBuildCommonPipelineEnvironment{}
	e := execMockRunner{}

	existingFiles := make(map[string]string)
	existingFiles["package.json"] = "{\"name\": \"myName\", \"version\": \"1.2.3\"}"
	fileUtils := MtaTestFileUtilsMock{existingFiles: existingFiles}

	err := runMtaBuild(options, &cpe, &e, &fileUtils)

	if err != nil {
		t.Errorf("No Error expected but error received: '%s'", err.Error())
	}

	assert.Equal(t, "java", e.calls[0].exec)
	assert.Equal(t, []string {"-jar", "mta.jar", "--mtar", "myName.mtar", "--build-target=CF"}, e.calls[0].params)
}

func TestMtaBuildMbtToolset(t *testing.T) {

	options := mtaBuildOptions{ApplicationName: "myApp", MtaBuildTool: "cloudMbt", Platform: "CF"}
	cpe := mtaBuildCommonPipelineEnvironment{}
	e := execMockRunner{}

	existingFiles := make(map[string]string)
	existingFiles["package.json"] = "{\"name\": \"myName\", \"version\": \"1.2.3\"}"
	fileUtils := MtaTestFileUtilsMock{existingFiles: existingFiles}

	err := runMtaBuild(options, &cpe, &e, &fileUtils)

	if err != nil {
		t.Fatalf("No Error expected but error received: '%s'", err.Error())

	}

	assert.Equal(t, "mbt", e.calls[0].exec)
	assert.Equal(t, []string {"build", "--mtar", "myName.mtar", "--platform", "CF", "--target", "./"}, e.calls[0].params)

}


type MtaTestFileUtilsMock struct {
	existingFiles map[string]string
	writtenFiles map[string]string
}

func (f *MtaTestFileUtilsMock) FileExists(path string) (bool, error) {

	if _, ok := f.existingFiles[path]; ok {
		return true, nil
	}
	return false, nil
}

func (f *MtaTestFileUtilsMock) Copy(src, dest string) (int64, error) {
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
