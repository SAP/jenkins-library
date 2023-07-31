//go:build unit
// +build unit

package log

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestFatalHookLevels(t *testing.T) {
	hook := FatalHook{}
	assert.Equal(t, []logrus.Level{logrus.FatalLevel}, hook.Levels())
}

func TestFatalHookFire(t *testing.T) {
	workspace := t.TempDir()

	t.Run("with step name", func(t *testing.T) {
		hook := FatalHook{
			Path:          workspace,
			CorrelationID: "https://build.url",
		}
		entry := logrus.Entry{
			Data: logrus.Fields{
				"category": "testCategory",
				"stepName": "testStep",
			},
			Message: "the error message",
		}

		err := hook.Fire(&entry)

		assert.NoError(t, err)
		fileContent, err := os.ReadFile(filepath.Join(workspace, "testStep_errorDetails.json"))
		assert.NoError(t, err)
		assert.NotContains(t, string(fileContent), `"category":"testCategory"`)
		assert.Contains(t, string(fileContent), `"correlationId":"https://build.url"`)
		assert.Contains(t, string(fileContent), `"message":"the error message"`)
	})

	t.Run("no step name", func(t *testing.T) {
		hook := FatalHook{
			Path:          workspace,
			CorrelationID: "https://build.url",
		}
		entry := logrus.Entry{
			Data: logrus.Fields{
				"category": "testCategory",
			},
			Message: "the error message",
		}

		err := hook.Fire(&entry)

		assert.NoError(t, err)
		fileContent, err := os.ReadFile(filepath.Join(workspace, "errorDetails.json"))
		assert.NoError(t, err)
		assert.NotContains(t, string(fileContent), `"category":"testCategory"`)
		assert.Contains(t, string(fileContent), `"correlationId":"https://build.url"`)
		assert.Contains(t, string(fileContent), `"message":"the error message"`)
	})

	t.Run("file exists", func(t *testing.T) {
		hook := FatalHook{}
		entry := logrus.Entry{
			Message: "the new error message",
		}

		err := hook.Fire(&entry)

		assert.NoError(t, err)
		fileContent, err := os.ReadFile(filepath.Join(workspace, "errorDetails.json"))
		assert.NoError(t, err)

		assert.Contains(t, string(fileContent), `"message":"the error message"`)
	})
}
