package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

	// negative case
	options := ClientOptions{ImageName: "abc", RegistryURL: " http: //aa.bb"}
	client.SetOptions(options)
	_, err := client.GetImageSource()
	assert.NotNil(t, err)
}
