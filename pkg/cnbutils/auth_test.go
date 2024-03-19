//go:build unit
// +build unit

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
		mockUtils.AddFile("/test/valid_config.json", []byte(`{"auths":{"example.com":{"auth":"dXNlcm5hbWU6cGFzc3dvcmQ="}}}`))
		credentials := cnbutils.NewCredentials(mockUtils)
		auth, err := credentials.GenerateCredentials("/test/valid_config.json")
		assert.NoError(t, err)
		assert.Equal(t, `{"example.com":"Basic dXNlcm5hbWU6cGFzc3dvcmQ="}`, auth)
	})

	t.Run("successfully generates cnb auth env variable from username and password", func(t *testing.T) {
		mockUtils.AddFile("/test/valid_config.json", []byte(`{"auths":{"example.com":{"username":"username","password":"password"}}}`))
		credentials := cnbutils.NewCredentials(mockUtils)
		auth, err := credentials.GenerateCredentials("/test/valid_config.json")
		assert.NoError(t, err)
		assert.Equal(t, `{"example.com":"Basic dXNlcm5hbWU6cGFzc3dvcmQ="}`, auth)
	})

	t.Run("skips registry with empty credentials", func(t *testing.T) {
		mockUtils.AddFile("/test/valid_config.json", []byte(`{"auths":{"example.com":{}}}`))
		credentials := cnbutils.NewCredentials(mockUtils)
		auth, err := credentials.GenerateCredentials("/test/valid_config.json")
		assert.NoError(t, err)
		assert.Equal(t, "{}", auth)
	})

	t.Run("successfully generates cnb auth env variable if docker config is not present", func(t *testing.T) {
		credentials := cnbutils.NewCredentials(mockUtils)
		auth, err := credentials.GenerateCredentials("")
		assert.NoError(t, err)
		assert.Equal(t, "{}", auth)
	})

	t.Run("fails if file not found", func(t *testing.T) {
		credentials := cnbutils.NewCredentials(mockUtils)
		_, err := credentials.GenerateCredentials("/not/found")
		assert.Error(t, err)
		assert.Equal(t, "could not read '/not/found'", err.Error())
	})

	t.Run("fails if file is invalid json", func(t *testing.T) {
		mockUtils.AddFile("/test/invalid_config.json", []byte(`not a json`))
		credentials := cnbutils.NewCredentials(mockUtils)
		_, err := credentials.GenerateCredentials("/test/invalid_config.json")
		assert.Error(t, err)
		assert.Equal(t, "invalid character 'o' in literal null (expecting 'u')", err.Error())
	})

	t.Run("validate finds present registry", func(t *testing.T) {
		mockUtils.AddFile("/test/valid_config.json", []byte(`{"auths":{"example.com":{"auth":"dXNlcm5hbWU6cGFzc3dvcmQ="}}}`))
		credentials := cnbutils.NewCredentials(mockUtils)
		_, err := credentials.GenerateCredentials("/test/valid_config.json")
		assert.NoError(t, err)
		assert.True(t, credentials.Validate("example.com"))
		assert.False(t, credentials.Validate("missing.com"))

	})

	t.Run("validate handles invalid credentials", func(t *testing.T) {
		mockUtils.AddFile("/test/invalid_config.json", []byte(`{"foo":"bar"}`))
		credentials := cnbutils.NewCredentials(mockUtils)
		_, err := credentials.GenerateCredentials("/test/invalid_config.json")
		assert.NoError(t, err)
		assert.False(t, credentials.Validate("example.com"))
	})

	t.Run("validate ignors registry on localhost", func(t *testing.T) {
		mockUtils.AddFile("/test/valid_config.json", []byte(`{"auths":{}}`))
		credentials := cnbutils.NewCredentials(mockUtils)
		_, err := credentials.GenerateCredentials("/test/valid_config.json")
		assert.NoError(t, err)
		assert.True(t, credentials.Validate("localhost:5000"))
	})
}
