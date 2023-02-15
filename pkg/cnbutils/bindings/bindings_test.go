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

	t.Run("DEPRECATED: writes bindings to files", func(t *testing.T) {
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
			"new": map[string]interface{}{
				"type": "mixin",
				"data": []map[string]interface{}{
					{
						"key":     "file",
						"content": "content",
					},
				},
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

	t.Run("writes bindings to files", func(t *testing.T) {
		var utils = mockUtils()
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, "http://test-url.com/binding", httpmock.NewStringResponder(200, "from url content"))
		client := &piperhttp.Client{}
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
		err := bindings.ProcessBindings(utils, client, "/tmp/platform", map[string]interface{}{
			"a": map[string]interface{}{
				"type": "test",
				"data": []map[string]interface{}{
					{
						"key":     "inline.yaml",
						"content": "my inline content",
					},
					{
						"key":  "from-file.yaml",
						"file": "/tmp/somefile.yaml",
					},
					{
						"key":     "from-url.yaml",
						"fromUrl": "http://test-url.com/binding",
					},
				},
			},
			"b": map[string]interface{}{
				"type": "test2",
				"data": []map[string]interface{}{
					{
						"key":     "inline2.yaml",
						"content": "my inline content2",
					},
				},
			},
		})

		if assert.NoError(t, err) {
			if assert.True(t, utils.HasFile("/tmp/platform/bindings/a/type")) {
				content, err := utils.FileRead("/tmp/platform/bindings/a/type")
				if assert.NoError(t, err) {
					assert.Equal(t, string(content), "test")
				}

				if assert.True(t, utils.HasFile("/tmp/platform/bindings/a/inline.yaml")) {
					content, err = utils.FileRead("/tmp/platform/bindings/a/inline.yaml")
					if assert.NoError(t, err) {
						assert.Equal(t, string(content), "my inline content")
					}
				}

				assert.True(t, utils.HasCopiedFile("/tmp/somefile.yaml", "/tmp/platform/bindings/a/from-file.yaml"))

				if assert.True(t, utils.HasFile("/tmp/platform/bindings/a/from-url.yaml")) {
					content, err := utils.FileRead("/tmp/platform/bindings/a/from-url.yaml")
					if assert.NoError(t, err) {
						assert.Equal(t, string(content), "from url content")
					}
				}
			}
			if assert.True(t, utils.HasFile("/tmp/platform/bindings/b/type")) {
				content, err := utils.FileRead("/tmp/platform/bindings/b/type")
				if assert.NoError(t, err) {
					assert.Equal(t, string(content), "test2")
				}

				if assert.True(t, utils.HasFile("/tmp/platform/bindings/b/inline2.yaml")) {
					content, err = utils.FileRead("/tmp/platform/bindings/b/inline2.yaml")
					if assert.NoError(t, err) {
						assert.Equal(t, string(content), "my inline content2")
					}
				}
			}
		}
	})

	t.Run("fails if the name being invalid", func(t *testing.T) {
		var utils = mockUtils()
		err := bindings.ProcessBindings(utils, &piperhttp.Client{}, "/tmp/platform", map[string]interface{}{
			"..": map[string]interface{}{
				"type": "inline",
				"data": []map[string]interface{}{
					{
						"key":     "inline.yaml",
						"content": "my inline content",
					},
				},
			},
		})

		if assert.Error(t, err) {
			assert.Equal(t, "invalid binding name: '..'", err.Error())
		}
	})

	t.Run("fails if the key being invalid", func(t *testing.T) {
		var utils = mockUtils()
		err := bindings.ProcessBindings(utils, &piperhttp.Client{}, "/tmp/platform", map[string]interface{}{
			"my-binding": map[string]interface{}{
				"type": "inline",
				"data": []map[string]interface{}{
					{
						"key":     "test/test.yaml",
						"content": "my inline content",
					},
				},
			},
		})

		if assert.Error(t, err) {
			assert.Equal(t, "failed to validate binding 'my-binding': invalid key: 'test/test.yaml'", err.Error())
		}
	})

	t.Run("fails if both content and file being specified", func(t *testing.T) {
		var utils = mockUtils()
		err := bindings.ProcessBindings(utils, &piperhttp.Client{}, "/tmp/platform", map[string]interface{}{
			"my-binding": map[string]interface{}{
				"type": "both",
				"data": []map[string]interface{}{
					{
						"key":     "test.yaml",
						"content": "my inline content",
						"file":    "/tmp/somefile.yaml",
					},
				},
			},
		})

		if assert.Error(t, err) {
			assert.Equal(t, "failed to validate binding 'my-binding': only one of 'content', 'file' or 'fromUrl' can be set", err.Error())
		}
	})

	t.Run("fails if no content or file being specified", func(t *testing.T) {
		var utils = mockUtils()
		err := bindings.ProcessBindings(utils, &piperhttp.Client{}, "/tmp/platform", map[string]interface{}{
			"my-binding": map[string]interface{}{
				"type": "none",
				"data": []map[string]interface{}{{"key": "test.yaml"}},
			},
		})

		if assert.Error(t, err) {
			assert.Equal(t, "failed to validate binding 'my-binding': one of 'file', 'content' or 'fromUrl' properties must be specified", err.Error())
		}
	})

	t.Run("fails if binding has no data specified", func(t *testing.T) {
		var utils = mockUtils()
		err := bindings.ProcessBindings(utils, &piperhttp.Client{}, "/tmp/platform", map[string]interface{}{
			"my-binding": map[string]interface{}{
				"type": "none",
				"data": []map[string]interface{}{},
			},
		})

		if assert.Error(t, err) {
			assert.Equal(t, "empty binding: 'my-binding'", err.Error())
		}
	})

	t.Run("fails if binding is not a map", func(t *testing.T) {
		var utils = mockUtils()
		err := bindings.ProcessBindings(utils, &piperhttp.Client{}, "/tmp/platform", map[string]interface{}{
			"my-binding": 42,
		})

		if assert.Error(t, err) {
			assert.Equal(t, "error while reading bindings: could not process binding 'my-binding'", err.Error())
		}
	})

	t.Run("fails if binding is an invalid map", func(t *testing.T) {
		var utils = mockUtils()
		err := bindings.ProcessBindings(utils, &piperhttp.Client{}, "/tmp/platform", map[string]interface{}{
			"my-binding": map[string]interface{}{
				"typo": "test",
				"data": []map[string]interface{}{{
					"key": "test.yaml",
				}},
			},
		})

		if assert.Error(t, err) {
			assert.Equal(t, "error while reading bindings: could not process binding 'my-binding'", err.Error())
		}
	})
}
