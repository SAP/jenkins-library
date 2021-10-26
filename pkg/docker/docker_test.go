package docker

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"

	"github.com/stretchr/testify/assert"
)

func TestCreateDockerConfigJSON(t *testing.T) {
	t.Parallel()
	t.Run("success - new file", func(t *testing.T) {
		utilsMock := mock.FilesMock{}
		configFile, err := CreateDockerConfigJSON("https://test.server.url", "testUser", "testPassword", "", &utilsMock)
		assert.NoError(t, err)

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
		configFile, err := CreateDockerConfigJSON("https://test.server.url", "testUser", "testPassword", existingConfigFilePath, &utilsMock)
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
		configFile, err := CreateDockerConfigJSON("https://test.server.url", "testUser", "testPassword", existingConfigFilePath, &utilsMock)
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

		_, err := CreateDockerConfigJSON("https://test.server.url", "testUser", "testPassword", existingConfigFilePath, &utilsMock)

		assert.Error(t, err)
		assert.Contains(t, fmt.Sprint(err), "failed to read file '.docker/config.json'")

	})

	t.Run("error - config file unmarshal", func(t *testing.T) {
		utilsMock := mock.FilesMock{}
		existingConfig := `{`
		existingConfigFilePath := ".docker/config.json"
		utilsMock.AddFile(existingConfigFilePath, []byte(existingConfig))
		_, err := CreateDockerConfigJSON("https://test.server.url", "testUser", "testPassword", existingConfigFilePath, &utilsMock)

		assert.Error(t, err)
		assert.Contains(t, fmt.Sprint(err), "failed to unmarshal json file '.docker/config.json'")
	})

	t.Run("error - config file write", func(t *testing.T) {
		utilsMock := mock.FilesMock{}
		utilsMock.FileWriteError = fmt.Errorf("write error")
		_, err := CreateDockerConfigJSON("https://test.server.url", "testUser", "testPassword", "", &utilsMock)

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
