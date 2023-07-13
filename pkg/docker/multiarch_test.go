//go:build unit
// +build unit

package docker

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestIsBinfmtMiscSupportedByHost(t *testing.T) {
	t.Run("returns true - binfmt_misc supported by host", func(t *testing.T) {
		utils := mock.FilesMock{}
		utils.AddDir("/proc/sys/fs/binfmt_misc")

		b, err := IsBinfmtMiscSupportedByHost(&utils)

		if assert.NoError(t, err) {
			assert.True(t, b)
		}
	})

	t.Run("returns false - binfmt_misc not supported by host", func(t *testing.T) {
		utils := mock.FilesMock{}

		b, err := IsBinfmtMiscSupportedByHost(&utils)

		if assert.NoError(t, err) {
			assert.False(t, b)
		}
	})
}
