package bindings_test

import (
	"net/http"
	"testing"

	"github.com/SAP/jenkins-library/pkg/cnbutils"
	"github.com/SAP/jenkins-library/pkg/cnbutils/bindings"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestProcessBindings(t *testing.T) {
	var mockUtils = func() *cnbutils.MockUtils {
		var utils = &cnbutils.MockUtils{
			FilesMock: &mock.FilesMock{},
		}
		utils.AddFile("/tmp/somefile.yaml", []byte("some file content"))
		return utils
	}

	t.Run("writes bindings to files", func(t *testing.T) {
		var utils = mockUtils()
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, "http://test-url.com/binding", httpmock.NewStringResponder(200, "from url content"))
		client := &piperhttp.Client{}
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
		err := bindings.ProcessBindings(utils, client, "/tmp/platform", map[string]interface{}{
			"a": map[string]interface{}{
				"key":     "inline.yaml",
				"type":    "inline",
				"content": "my inline content",
			},
			"b": map[string]interface{}{
				"key":  "from-file.yaml",
				"type": "file",
				"file": "/tmp/somefile.yaml",
			},
			"c": map[string]interface{}{
				"key":     "from-url.yaml",
				"type":    "url",
				"fromUrl": "http://test-url.com/binding",
			},
		})

		if assert.NoError(t, err) {
			if assert.True(t, utils.HasFile("/tmp/platform/bindings/a/inline.yaml")) {
				content, err := utils.FileRead("/tmp/platform/bindings/a/inline.yaml")
				if assert.NoError(t, err) {
					assert.Equal(t, string(content), "my inline content")
				}
			}

			if assert.True(t, utils.HasFile("/tmp/platform/bindings/a/type")) {
				content, err := utils.FileRead("/tmp/platform/bindings/a/type")
				if assert.NoError(t, err) {
					assert.Equal(t, string(content), "inline")
				}
			}

			assert.True(t, utils.HasCopiedFile("/tmp/somefile.yaml", "/tmp/platform/bindings/b/from-file.yaml"))

			if assert.True(t, utils.HasFile("/tmp/platform/bindings/b/type")) {
				content, err := utils.FileRead("/tmp/platform/bindings/b/type")
				if assert.NoError(t, err) {
					assert.Equal(t, string(content), "file")
				}
			}

			if assert.True(t, utils.HasFile("/tmp/platform/bindings/c/type")) {
				content, err := utils.FileRead("/tmp/platform/bindings/c/type")
				if assert.NoError(t, err) {
					assert.Equal(t, string(content), "url")
				}
			}

			if assert.True(t, utils.HasFile("/tmp/platform/bindings/c/from-url.yaml")) {
				content, err := utils.FileRead("/tmp/platform/bindings/c/from-url.yaml")
				if assert.NoError(t, err) {
					assert.Equal(t, string(content), "from url content")
				}
			}
		}
	})

	t.Run("fails with the name being invalid", func(t *testing.T) {
		var utils = mockUtils()
		err := bindings.ProcessBindings(utils, &piperhttp.Client{}, "/tmp/platform", map[string]interface{}{
			"..": map[string]interface{}{
				"key":     "inline.yaml",
				"type":    "inline",
				"content": "my inline content",
			},
		})

		if assert.Error(t, err) {
			assert.Equal(t, "invalid binding name: '..'", err.Error())
		}
	})

	t.Run("fails with the key being invalid", func(t *testing.T) {
		var utils = mockUtils()
		err := bindings.ProcessBindings(utils, &piperhttp.Client{}, "/tmp/platform", map[string]interface{}{
			"binding": map[string]interface{}{
				"key":     "test/test.yaml",
				"type":    "inline",
				"content": "my inline content",
			},
		})

		if assert.Error(t, err) {
			assert.Equal(t, "invalid key: 'test/test.yaml'", err.Error())
		}
	})

	t.Run("fails with both content and file being specified", func(t *testing.T) {
		var utils = mockUtils()
		err := bindings.ProcessBindings(utils, &piperhttp.Client{}, "/tmp/platform", map[string]interface{}{
			"binding": map[string]interface{}{
				"key":     "test.yaml",
				"type":    "both",
				"content": "my inline content",
				"file":    "/tmp/somefile.yaml",
			},
		})

		if assert.Error(t, err) {
			assert.Equal(t, "only one of 'content', 'file' or 'fromUrl' can be set for a binding", err.Error())
		}
	})

	t.Run("fails with no content or file being specified", func(t *testing.T) {
		var utils = mockUtils()
		err := bindings.ProcessBindings(utils, &piperhttp.Client{}, "/tmp/platform", map[string]interface{}{
			"binding": map[string]interface{}{
				"key":  "test.yaml",
				"type": "none",
			},
		})

		if assert.Error(t, err) {
			assert.Equal(t, "one of 'file', 'content' or 'fromUrl' properties must be specified for binding", err.Error())
		}
	})

	t.Run("fails with not a map", func(t *testing.T) {
		var utils = mockUtils()
		err := bindings.ProcessBindings(utils, &piperhttp.Client{}, "/tmp/platform", map[string]interface{}{
			"binding": 42,
		})

		if assert.Error(t, err) {
			assert.Equal(t, "failed to convert map to struct: 1 error(s) decoding:\n\n* '[binding]' expected a map, got 'int'", err.Error())

		}
	})

	t.Run("fails with invalid map", func(t *testing.T) {
		var utils = mockUtils()
		err := bindings.ProcessBindings(utils, &piperhttp.Client{}, "/tmp/platform", map[string]interface{}{
			"test": map[string]interface{}{
				"key":  "test.yaml",
				"typo": "test",
			},
		})

		if assert.Error(t, err) {
			assert.Equal(t, "failed to convert map to struct: 1 error(s) decoding:\n\n* '[test]' has invalid keys: typo", err.Error())
		}
	})
}
