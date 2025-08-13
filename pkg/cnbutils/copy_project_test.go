package cnbutils_test

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/cnbutils"
	"github.com/SAP/jenkins-library/pkg/mock"
	ignore "github.com/sabhiram/go-gitignore"
	"github.com/stretchr/testify/assert"
)

func TestCopyProject(t *testing.T) {
	t.Run("copy project with following symlinks", func(t *testing.T) {
		mockUtils := &cnbutils.MockUtils{
			FilesMock: &mock.FilesMock{},
		}
		mockUtils.AddFile("workdir/src/test.yaml", []byte(""))
		mockUtils.AddFile("workdir/src/subdir1/test2.yaml", []byte(""))
		mockUtils.AddFile("workdir/src/subdir1/subdir2/test3.yaml", []byte(""))

		mockUtils.AddDir("workdir/apps")
		mockUtils.AddFile("workdir/apps/foo.yaml", []byte(""))
		mockUtils.Symlink("workdir/apps", "/workdir/src/apps")

		err := cnbutils.CopyProject("workdir/src", "/dest", ignore.CompileIgnoreLines([]string{"**"}...), nil, mockUtils, true)
		assert.NoError(t, err)
		assert.True(t, mockUtils.HasCopiedFile("workdir/src/test.yaml", "/dest/test.yaml"))
		assert.True(t, mockUtils.HasCopiedFile("workdir/src/subdir1/test2.yaml", "/dest/subdir1/test2.yaml"))
		assert.True(t, mockUtils.HasCopiedFile("workdir/src/subdir1/subdir2/test3.yaml", "/dest/subdir1/subdir2/test3.yaml"))
		assert.True(t, mockUtils.HasCopiedFile("workdir/src/apps", "/dest/apps"))
	})

	t.Run("copy project without following symlinks", func(t *testing.T) {
		mockUtils := &cnbutils.MockUtils{
			FilesMock: &mock.FilesMock{},
		}
		mockUtils.AddFile("workdir/src/test.yaml", []byte(""))
		mockUtils.AddFile("workdir/src/subdir1/test2.yaml", []byte(""))
		mockUtils.AddFile("workdir/src/subdir1/subdir2/test3.yaml", []byte(""))

		mockUtils.AddDir("workdir/apps")
		mockUtils.AddFile("workdir/apps/foo.yaml", []byte(""))
		mockUtils.Symlink("workdir/apps", "/workdir/src/apps")

		err := cnbutils.CopyProject("workdir/src", "/dest", ignore.CompileIgnoreLines([]string{"**/*.yaml"}...), nil, mockUtils, false)
		assert.NoError(t, err)
		assert.True(t, mockUtils.HasCopiedFile("workdir/src/test.yaml", "/dest/test.yaml"))
		assert.True(t, mockUtils.HasCopiedFile("workdir/src/subdir1/test2.yaml", "/dest/subdir1/test2.yaml"))
		assert.True(t, mockUtils.HasCopiedFile("workdir/src/subdir1/subdir2/test3.yaml", "/dest/subdir1/subdir2/test3.yaml"))
		assert.True(t, mockUtils.HasCreatedSymlink("workdir/apps", "/workdir/src/apps"))
	})
}
