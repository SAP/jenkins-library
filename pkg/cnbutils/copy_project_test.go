//go:build unit
// +build unit

package cnbutils_test

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/cnbutils"
	"github.com/SAP/jenkins-library/pkg/mock"
	ignore "github.com/sabhiram/go-gitignore"
	"github.com/stretchr/testify/assert"
)

func TestCopyProject(t *testing.T) {
	t.Run("copies file according to doublestart globs", func(t *testing.T) {
		mockUtils := &cnbutils.MockUtils{
			FilesMock: &mock.FilesMock{},
		}
		mockUtils.AddFile("workdir/src/test.yaml", []byte(""))
		mockUtils.AddFile("workdir/src/subdir1/test2.yaml", []byte(""))
		mockUtils.AddFile("workdir/src/subdir1/subdir2/test3.yaml", []byte(""))
		err := cnbutils.CopyProject("workdir/src", "/dest", ignore.CompileIgnoreLines([]string{"**/*.yaml"}...), nil, mockUtils)
		assert.NoError(t, err)
		assert.True(t, mockUtils.HasCopiedFile("workdir/src/test.yaml", "/dest/test.yaml"))
		assert.True(t, mockUtils.HasCopiedFile("workdir/src/subdir1/test2.yaml", "/dest/subdir1/test2.yaml"))
		assert.True(t, mockUtils.HasCopiedFile("workdir/src/subdir1/subdir2/test3.yaml", "/dest/subdir1/subdir2/test3.yaml"))
	})

	t.Run("copies file according to simple globs", func(t *testing.T) {
		mockUtils := &cnbutils.MockUtils{
			FilesMock: &mock.FilesMock{},
		}
		mockUtils.AddFile("src/test.yaml", []byte(""))
		err := cnbutils.CopyProject("src", "/dest", ignore.CompileIgnoreLines([]string{"*.yaml"}...), nil, mockUtils)
		assert.NoError(t, err)
		assert.True(t, mockUtils.HasCopiedFile("src/test.yaml", "/dest/test.yaml"))
	})
}
