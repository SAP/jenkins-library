//go:build unit

package cnbutils_test

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/cnbutils"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestCreateEnvFiles(t *testing.T) {
	t.Run("successfully writes environment files", func(t *testing.T) {
		mockUtils := &cnbutils.MockUtils{
			FilesMock: &mock.FilesMock{},
		}

		envVars := map[string]interface{}{
			"FOO":     "BAR",
			"BAR":     "BAZ",
			"COMPLEX": "{\"foo\": \"bar=3\"}",
		}

		err := cnbutils.CreateEnvFiles(mockUtils, "/tmp/platform", envVars)

		assert.NoError(t, err)
		assert.True(t, mockUtils.HasWrittenFile("/tmp/platform/env/FOO"))
		assert.True(t, mockUtils.HasWrittenFile("/tmp/platform/env/BAR"))
		assert.True(t, mockUtils.HasWrittenFile("/tmp/platform/env/COMPLEX"))

		result1, err := mockUtils.FileRead("/tmp/platform/env/FOO")
		assert.NoError(t, err)
		assert.Equal(t, "BAR", string(result1))

		result2, err := mockUtils.FileRead("/tmp/platform/env/BAR")
		assert.NoError(t, err)
		assert.Equal(t, "BAZ", string(result2))

		result3, err := mockUtils.FileRead("/tmp/platform/env/COMPLEX")
		assert.NoError(t, err)
		assert.Equal(t, "{\"foo\": \"bar=3\"}", string(result3))
	})

	t.Run("raises an error if unable to write to a file", func(t *testing.T) {
		mockUtils := &cnbutils.MockUtils{
			FilesMock: &mock.FilesMock{
				FileWriteErrors: map[string]error{
					"/tmp/platform/env/FOO": fmt.Errorf("unable to create dir"),
				},
			},
		}

		err := cnbutils.CreateEnvFiles(mockUtils, "/tmp/platform", map[string]interface{}{"FOO": "BAR"})
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "unable to create dir")
	})
}
