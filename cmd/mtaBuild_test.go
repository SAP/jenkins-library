package cmd

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
