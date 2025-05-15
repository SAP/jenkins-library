//go:build unit
// +build unit

package docker

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"

	"github.com/stretchr/testify/assert"
)

func TestCreateDockerConfigJSON(t *testing.T) {
	t.Parallel()
	t.Run("success - new file", func(t *testing.T) {
		utilsMock := mock.FilesMock{}
		configFile, err := CreateDockerConfigJSON("https://test.server.url", "testUser", "testPassword", "test/config.json", "", &utilsMock)
		assert.NoError(t, err)
		assert.Equal(t, "test/config.json", configFile)

		configFileContent, err := utilsMock.FileRead(configFile)
		assert.NoError(t, err)
		assert.Contains(t, string(configFileContent), `"auth":"dGVzdFVzZXI6dGVzdFBhc3N3b3Jk"`)
	})

	t.Run("success - update file", func(t *testing.T) {
		utilsMock := mock.FilesMock{}
		existingConfig := `{
	"auths": {
			"existing.registry.url:50000": {
					"auth": "Base64Auth"
			}
	},
	"HttpHeaders": {
			"User-Agent": "Docker-Client/18.06.3-ce (linux)"
	}
}`
		existingConfigFilePath := ".docker/config.json"
		utilsMock.AddFile(existingConfigFilePath, []byte(existingConfig))
		configFile, err := CreateDockerConfigJSON("https://test.server.url", "testUser", "testPassword", "", existingConfigFilePath, &utilsMock)
		assert.NoError(t, err)

		configFileContent, err := utilsMock.FileRead(configFile)
		assert.NoError(t, err)
		assert.Contains(t, string(configFileContent), `"existing.registry.url:50000"`)
		assert.Contains(t, string(configFileContent), `"auth":"Base64Auth"`)
		assert.Contains(t, string(configFileContent), `"https://test.server.url"`)
		assert.Contains(t, string(configFileContent), `"auth":"dGVzdFVzZXI6dGVzdFBhc3N3b3Jk"`)
		assert.Contains(t, string(configFileContent), `"User-Agent":"Docker-Client/18.06.3-ce (linux)`)
	})

	t.Run("success - update file with empty auths", func(t *testing.T) {
		utilsMock := mock.FilesMock{}
		existingConfig := `{
	"auths": {},
	"HttpHeaders": {
		"User-Agent": "Docker-Client/18.06.3-ce (linux)"
	}
}`
		existingConfigFilePath := ".docker/config.json"
		utilsMock.AddFile(existingConfigFilePath, []byte(existingConfig))
		configFile, err := CreateDockerConfigJSON("https://test.server.url", "testUser", "testPassword", "", existingConfigFilePath, &utilsMock)
		assert.NoError(t, err)

		configFileContent, err := utilsMock.FileRead(configFile)
		assert.NoError(t, err)
		assert.Contains(t, string(configFileContent), `"auth":"dGVzdFVzZXI6dGVzdFBhc3N3b3Jk"`)
		assert.Contains(t, string(configFileContent), `"User-Agent":"Docker-Client/18.06.3-ce (linux)`)
	})

	t.Run("error - config file read", func(t *testing.T) {
		// FilesMock does not yet provide capability for FileRead errors
		//t.Skip()
		utilsMock := mock.FilesMock{FileReadErrors: map[string]error{".docker/config.json": fmt.Errorf("read error")}}
		existingConfigFilePath := ".docker/config.json"
		utilsMock.AddFile(existingConfigFilePath, []byte("{}"))

		_, err := CreateDockerConfigJSON("https://test.server.url", "testUser", "testPassword", "", existingConfigFilePath, &utilsMock)

		assert.Error(t, err)
		assert.Contains(t, fmt.Sprint(err), "failed to read file '.docker/config.json'")

	})

	t.Run("error - config file unmarshal", func(t *testing.T) {
		utilsMock := mock.FilesMock{}
		existingConfig := `{`
		existingConfigFilePath := ".docker/config.json"
		utilsMock.AddFile(existingConfigFilePath, []byte(existingConfig))
		_, err := CreateDockerConfigJSON("https://test.server.url", "testUser", "testPassword", "", existingConfigFilePath, &utilsMock)

		assert.Error(t, err)
		assert.Contains(t, fmt.Sprint(err), "failed to unmarshal json file '.docker/config.json'")
	})

	t.Run("error - config file write", func(t *testing.T) {
		utilsMock := mock.FilesMock{}
		utilsMock.FileWriteError = fmt.Errorf("write error")
		_, err := CreateDockerConfigJSON("https://test.server.url", "testUser", "testPassword", "", "", &utilsMock)

		assert.Error(t, err)
		assert.Contains(t, fmt.Sprint(err), "failed to write Docker config.json")
	})
}

func TestImageListWithFilePath(t *testing.T) {
	t.Parallel()

	imageName := "testImage"

	tt := []struct {
		name          string
		excludes      []string
		trimDir       string
		fileList      []string
		expected      map[string]string
		expectedError error
	}{
		{name: "Dockerfile only", fileList: []string{"Dockerfile"}, expected: map[string]string{imageName: "Dockerfile"}},
		{name: "Dockerfile in subdir", fileList: []string{"sub/Dockerfile"}, expected: map[string]string{fmt.Sprintf("%v-sub", imageName): filepath.FromSlash("sub/Dockerfile")}},
		{name: "Dockerfile in subdir excluding top dir", fileList: []string{".ci/sub/Dockerfile"}, trimDir: ".ci", expected: map[string]string{fmt.Sprintf("%v-sub", imageName): filepath.FromSlash(".ci/sub/Dockerfile")}},
		{name: "Dockerfiles in multiple subdirs & parent", fileList: []string{"Dockerfile", "sub1/Dockerfile", "sub2/Dockerfile"}, expected: map[string]string{fmt.Sprintf("%v", imageName): filepath.FromSlash("Dockerfile"), fmt.Sprintf("%v-sub1", imageName): filepath.FromSlash("sub1/Dockerfile"), fmt.Sprintf("%v-sub2", imageName): filepath.FromSlash("sub2/Dockerfile")}},
		{name: "Dockerfiles in multiple subdirs & parent - with excludes", excludes: []string{"Dockerfile"}, fileList: []string{"Dockerfile", "sub1/Dockerfile", "sub2/Dockerfile"}, expected: map[string]string{fmt.Sprintf("%v-sub1", imageName): filepath.FromSlash("sub1/Dockerfile"), fmt.Sprintf("%v-sub2", imageName): filepath.FromSlash("sub2/Dockerfile")}},
		{name: "Dockerfiles with extensions", fileList: []string{"Dockerfile_main", "Dockerfile_sub1", "Dockerfile_sub2"}, expected: map[string]string{fmt.Sprintf("%v-main", imageName): filepath.FromSlash("Dockerfile_main"), fmt.Sprintf("%v-sub1", imageName): filepath.FromSlash("Dockerfile_sub1"), fmt.Sprintf("%v-sub2", imageName): filepath.FromSlash("Dockerfile_sub2")}},
		{name: "Dockerfiles with extensions", fileList: []string{"Dockerfile_main", "Dockerfile_sub1", "Dockerfile_sub2"}, expected: map[string]string{fmt.Sprintf("%v-main", imageName): filepath.FromSlash("Dockerfile_main"), fmt.Sprintf("%v-sub1", imageName): filepath.FromSlash("Dockerfile_sub1"), fmt.Sprintf("%v-sub2", imageName): filepath.FromSlash("Dockerfile_sub2")}},
		{name: "No Dockerfile", fileList: []string{"NoDockerFile"}, expectedError: fmt.Errorf("failed to retrieve Dockerfiles")},
		{name: "Incorrect Dockerfile", fileList: []string{"DockerfileNotSupported"}, expectedError: fmt.Errorf("wrong format of Dockerfile, must be inside a sub-folder or contain a separator")},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			fileMock := mock.FilesMock{}
			for _, file := range test.fileList {
				fileMock.AddFile(file, []byte("someContent"))
			}

			imageList, err := ImageListWithFilePath(imageName, test.excludes, test.trimDir, &fileMock)

			if test.expectedError != nil {
				assert.EqualError(t, err, fmt.Sprint(test.expectedError))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, imageList)
			}
		})
	}
}

func TestMergeDockerConfigJSON(t *testing.T) {
	t.Parallel()

	t.Run("success - both files present", func(t *testing.T) {
		sourceFile := "/tmp/source.json"
		targetFile := "/tmp/target.json"
		expectedContent := "{\n\t\"auths\": {\n\t\t\"bar\": {},\n\t\t\"foo\": {\n\t\t\t\"auth\": \"Zm9vOmJhcg==\"\n\t\t}\n\t}\n}"

		utilsMock := mock.FilesMock{}
		utilsMock.AddFile(targetFile, []byte("{\"auths\": {\"foo\": {\"auth\": \"dGVzdDp0ZXN0\"}}}"))
		utilsMock.AddFile(sourceFile, []byte("{\"auths\": {\"bar\": {}, \"foo\": {\"auth\": \"Zm9vOmJhcg==\"}}}"))

		err := MergeDockerConfigJSON(sourceFile, targetFile, &utilsMock)
		assert.NoError(t, err)

		content, err := utilsMock.FileRead(targetFile)
		assert.NoError(t, err)
		assert.Equal(t, expectedContent, string(content))
	})

	t.Run("success - target file is missing", func(t *testing.T) {
		sourceFile := "/tmp/source.json"
		targetFile := "/tmp/target.json"
		expectedContent := "{\n\t\"auths\": {\n\t\t\"bar\": {},\n\t\t\"foo\": {\n\t\t\t\"auth\": \"Zm9vOmJhcg==\"\n\t\t}\n\t}\n}"

		utilsMock := mock.FilesMock{}
		utilsMock.AddFile(sourceFile, []byte("{\"auths\": {\"bar\": {}, \"foo\": {\"auth\": \"Zm9vOmJhcg==\"}}}"))

		err := MergeDockerConfigJSON(sourceFile, targetFile, &utilsMock)
		assert.NoError(t, err)

		content, err := utilsMock.FileRead(targetFile)
		assert.NoError(t, err)
		assert.Equal(t, expectedContent, string(content))
	})

	t.Run("error - source file is missing", func(t *testing.T) {
		utilsMock := mock.FilesMock{}
		err := MergeDockerConfigJSON("missing-file", "also-missing-file", &utilsMock)
		assert.Error(t, err)
		assert.Equal(t, "source dockerConfigJSON file \"missing-file\" does not exist", err.Error())
	})
}
