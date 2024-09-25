//go:build unit
// +build unit

package cnbutils_test

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/cnbutils"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestDigestFromReport(t *testing.T) {
	t.Run("return a digest from the report.toml", func(t *testing.T) {
		mockUtils := &cnbutils.MockUtils{
			FilesMock: &mock.FilesMock{},
		}
		mockUtils.AddFile("/layers/report.toml", []byte(`[build]
[image]
digest = "sha256:52eac630560210e5ae13eb10797c4246d6f02d425f32b9430ca00bde697c79ec"
`))
		digest, err := cnbutils.DigestFromReport(mockUtils)
		assert.NoError(t, err)
		assert.Equal(t, "sha256:52eac630560210e5ae13eb10797c4246d6f02d425f32b9430ca00bde697c79ec", digest)
	})

	t.Run("fails if digest is empty", func(t *testing.T) {
		mockUtils := &cnbutils.MockUtils{
			ExecMockRunner: &mock.ExecMockRunner{},
			FilesMock:      &mock.FilesMock{},
		}
		mockUtils.AddFile("/layers/report.toml", []byte(``))

		digest, err := cnbutils.DigestFromReport(mockUtils)
		assert.Empty(t, digest)
		assert.EqualError(t, err, "image digest is empty")
	})

	t.Run("fails to unmarshal corrupted file", func(t *testing.T) {
		mockUtils := &cnbutils.MockUtils{
			ExecMockRunner: &mock.ExecMockRunner{},
			FilesMock:      &mock.FilesMock{},
		}
		mockUtils.AddFile("/layers/report.toml", []byte(`{}`))

		digest, err := cnbutils.DigestFromReport(mockUtils)
		assert.Empty(t, digest)
		assert.EqualError(t, err, "toml: line 1: expected '.' or '=', but got '{' instead")
	})
}
