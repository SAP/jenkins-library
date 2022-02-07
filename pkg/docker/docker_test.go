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

func TestGetImageSource(t *testing.T) {

	cases := []struct {
		imageName   string
		registryURL string
		localPath   string
		want        string
	}{
		{"imageName", "", "", "imageName"},
		{"imageName", "", "localPath", "daemon://localPath"},
		{"imageName", "http://registryURL", "", "remote://registryURL/imageName"},
		{"imageName", "https://containerRegistryUrl", "", "remote://containerRegistryUrl/imageName"},
		{"imageName", "registryURL", "", "remote://registryURL/imageName"},
	}

	client := Client{}

	for _, c := range cases {

		options := ClientOptions{ImageName: c.imageName, RegistryURL: c.registryURL, LocalPath: c.localPath}
		client.SetOptions(options)

		got, err := client.GetImageSource()

		assert.Nil(t, err)
		assert.Equal(t, c.want, got)
	}
}

func TestImageListWithFilePath(t *testing.T) {
	t.Parallel()

	imageName := "testImage"

	tt := []struct {
		name          string
		excludes      []string
		fileList      []string
		expected      map[string]string
		expectedError error
	}{
		{name: "Dockerfile only", fileList: []string{"Dockerfile"}, expected: map[string]string{imageName: "Dockerfile"}},
		{name: "Dockerfile in subdir", fileList: []string{"sub/Dockerfile"}, expected: map[string]string{fmt.Sprintf("%v-sub", imageName): filepath.FromSlash("sub/Dockerfile")}},
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

			imageList, err := ImageListWithFilePath(imageName, test.excludes, &fileMock)

			if test.expectedError != nil {
				assert.EqualError(t, err, fmt.Sprint(test.expectedError))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, imageList)
			}
		})
	}
}
