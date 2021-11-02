package bindings_test

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/cnbutils"
	"github.com/SAP/jenkins-library/pkg/cnbutils/bindings"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestProcessBindings(t *testing.T) {

	var mockUtils = func() cnbutils.MockUtils {
		var utils = cnbutils.MockUtils{
			FilesMock: &mock.FilesMock{},
		}
		utils.AddFile("/tmp/somefile.yaml", []byte("some file content"))
		return utils
	}

	t.Run("writes bindings to files", func(t *testing.T) {
		var utils = mockUtils()
		err := bindings.ProcessBindings(utils, "/tmp/platform", map[string]interface{}{
			"a": map[string]interface{}{
				"secret":  "inline.yaml",
				"type":    "inline",
				"content": "my inline content",
			},
			"b": map[string]interface{}{
				"secret": "from-file.yaml",
				"type":   "file",
				"file":   "/tmp/somefile.yaml",
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
		}
	})

	t.Run("fails with the name being invalid", func(t *testing.T) {
		var utils = mockUtils()
		err := bindings.ProcessBindings(utils, "/tmp/platform", map[string]interface{}{
			"..": map[string]interface{}{
				"secret":  "inline.yaml",
				"type":    "inline",
				"content": "my inline content",
			},
		})

		if assert.Error(t, err) {
			assert.Equal(t, "invalid binding name: ..", err.Error())
		}
	})

	t.Run("fails with the secret being invalid", func(t *testing.T) {
		var utils = mockUtils()
		err := bindings.ProcessBindings(utils, "/tmp/platform", map[string]interface{}{
			"binding": map[string]interface{}{
				"secret":  "test/test.yaml",
				"type":    "inline",
				"content": "my inline content",
			},
		})

		if assert.Error(t, err) {
			assert.Equal(t, "invalid secret name: test/test.yaml", err.Error())
		}
	})

	t.Run("fails with both content and file being specified", func(t *testing.T) {
		var utils = mockUtils()
		err := bindings.ProcessBindings(utils, "/tmp/platform", map[string]interface{}{
			"binding": map[string]interface{}{
				"secret":  "test.yaml",
				"type":    "both",
				"content": "my inline content",
				"file":    "/tmp/somefile.yaml",
			},
		})

		if assert.Error(t, err) {
			assert.Equal(t, "either 'file' or 'content' property must be specified for binding", err.Error())
		}
	})

	t.Run("fails with no content or file being specified", func(t *testing.T) {
		var utils = mockUtils()
		err := bindings.ProcessBindings(utils, "/tmp/platform", map[string]interface{}{
			"binding": map[string]interface{}{
				"secret": "test.yaml",
				"type":   "none",
			},
		})

		if assert.Error(t, err) {
			assert.Equal(t, "either 'file' or 'content' property must be specified for binding", err.Error())
		}
	})

	t.Run("fails with not a map", func(t *testing.T) {
		var utils = mockUtils()
		err := bindings.ProcessBindings(utils, "/tmp/platform", map[string]interface{}{
			"binding": 42,
		})

		if assert.Error(t, err) {
			assert.Equal(t, "1 error(s) decoding:\n\n* '[binding]' expected a map, got 'int'", err.Error())

		}
	})

	t.Run("fails with invalid map", func(t *testing.T) {
		var utils = mockUtils()
		err := bindings.ProcessBindings(utils, "/tmp/platform", map[string]interface{}{
			"test": map[string]interface{}{
				"secret": "test.yaml",
				"typo":   "test",
			},
		})

		if assert.Error(t, err) {
			assert.Equal(t, "1 error(s) decoding:\n\n* '[test]' has invalid keys: typo", err.Error())
		}
	})
}
