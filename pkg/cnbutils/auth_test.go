//go:build unit

package cnbutils_test

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/cnbutils"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestGenerateCnbAuth(t *testing.T) {
	var mockUtils = &cnbutils.MockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}

	t.Run("successfully generates cnb auth env variable", func(t *testing.T) {
		mockUtils.AddFile("/test/valid_config.json", []byte("{\"auths\":{\"https://example.com/\":{\"auth\":\"dXNlcm5hbWU6cGFzc3dvcmQ=\"}}}"))
		auth, err := cnbutils.ParseDockerConfig("/test/valid_config.json", mockUtils)
		assert.NoError(t, err)

		authString, err := auth.ToCNBString()
		assert.NoError(t, err)
		assert.Equal(t, "{\"example.com\":\"Basic dXNlcm5hbWU6cGFzc3dvcmQ=\"}", authString)
		assert.True(t, auth.AuthExistsForImage("example.com/foo/bar:123"))
		assert.False(t, auth.AuthExistsForImage("docker.io/foo/bar"))
	})

	t.Run("successfully generates cnb auth env variable from username and password", func(t *testing.T) {
		mockUtils.AddFile("/test/valid_config.json", []byte("{\"auths\":{\"example.com\":{\"username\":\"username\",\"password\":\"password\"}}}"))
		auth, err := cnbutils.ParseDockerConfig("/test/valid_config.json", mockUtils)
		assert.NoError(t, err)

		authString, err := auth.ToCNBString()
		assert.NoError(t, err)
		assert.Equal(t, "{\"example.com\":\"Basic dXNlcm5hbWU6cGFzc3dvcmQ=\"}", authString)
		assert.True(t, auth.AuthExistsForImage("example.com/foo/bar:123"))
		assert.False(t, auth.AuthExistsForImage("docker.io/foo/bar"))
	})

	t.Run("skips registry with empty credentials", func(t *testing.T) {
		mockUtils.AddFile("/test/valid_config.json", []byte("{\"auths\":{\"example.com\":{}}}"))
		auth, err := cnbutils.ParseDockerConfig("/test/valid_config.json", mockUtils)
		assert.NoError(t, err)

		authString, err := auth.ToCNBString()
		assert.NoError(t, err)
		assert.Equal(t, "{}", authString)
		assert.False(t, auth.AuthExistsForImage("example.com/foo/bar:123"))
		assert.False(t, auth.AuthExistsForImage("docker.io/foo/bar"))
	})

	t.Run("successfully generates cnb auth env variable if docker config is not present", func(t *testing.T) {
		auth, err := cnbutils.ParseDockerConfig("", mockUtils)
		assert.NoError(t, err)

		authString, err := auth.ToCNBString()
		assert.NoError(t, err)
		assert.Equal(t, "{}", authString)
		assert.False(t, auth.AuthExistsForImage("example.com/foo/bar:123"))
		assert.False(t, auth.AuthExistsForImage("docker.io/foo/bar"))
	})

	t.Run("fails if file not found", func(t *testing.T) {
		_, err := cnbutils.ParseDockerConfig("/not/found", mockUtils)
		assert.Error(t, err)
		assert.Equal(t, "could not read '/not/found'", err.Error())
	})

	t.Run("fails if file is invalid json", func(t *testing.T) {
		mockUtils.AddFile("/test/invalid_config.json", []byte("not a json"))
		_, err := cnbutils.ParseDockerConfig("/test/invalid_config.json", mockUtils)
		assert.Error(t, err)
		assert.Equal(t, "invalid character 'o' in literal null (expecting 'u')", err.Error())
	})
}
