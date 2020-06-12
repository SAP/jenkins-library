package mock

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestFilesMockFileExists(t *testing.T) {
	t.Parallel()
	t.Run("no init", func(t *testing.T) {
		files := FilesMock{}
		path := filepath.Join("some", "path")
		exists, err := files.FileExists(path)
		assert.NoError(t, err)
		assert.False(t, exists)
	})
	t.Run("file exists after AddFile()", func(t *testing.T) {
		files := FilesMock{}
		path := filepath.Join("some", "path")
		files.AddFile(path, []byte("dummy content"))
		exists, err := files.FileExists(path)
		assert.NoError(t, err)
		assert.True(t, exists)
	})
	t.Run("path is a directory after AddDir()", func(t *testing.T) {
		files := FilesMock{}
		path := filepath.Join("some", "path")
		files.AddDir(path)
		exists, err := files.FileExists(path)
		assert.NoError(t, err)
		assert.False(t, exists)
	})
	t.Run("file exists after changing current dir", func(t *testing.T) {
		files := FilesMock{}
		path := filepath.Join("some", "path")
		files.AddFile(path, []byte("dummy content"))
		err := files.Chdir("some")
		assert.NoError(t, err)
		exists, err := files.FileExists("path")
		assert.NoError(t, err)
		assert.True(t, exists)
	})
	t.Run("file does not exist after changing current dir", func(t *testing.T) {
		files := FilesMock{}
		path := filepath.Join("some", "path")
		files.AddFile(path, []byte("dummy content"))
		err := files.Chdir("some")
		assert.NoError(t, err)
		exists, err := files.FileExists(path)
		assert.EqualError(t, err, "'"+path+"': file does not exist")
		assert.False(t, exists)
	})
}

func TestFilesMockDirExists(t *testing.T) {
	t.Parallel()
	t.Run("no init", func(t *testing.T) {
		files := FilesMock{}
		path := filepath.Join("some", "path")
		exists, err := files.DirExists(path)
		assert.NoError(t, err)
		assert.False(t, exists)
	})
	t.Run("dir exists after AddDir()", func(t *testing.T) {
		files := FilesMock{}
		path := filepath.Join("some", "path")
		files.AddDir(path)
		exists, err := files.DirExists(path)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("parent dirs exists after AddFile()", func(t *testing.T) {
		files := FilesMock{}
		path := filepath.Join("path", "to", "some", "file")
		files.AddFile(path, []byte("dummy content"))
		testDirs := []string{
			"path",
			filepath.Join("path", "to"),
			filepath.Join("path", "to", "some"),
		}
		for _, dir := range testDirs {
			exists, err := files.DirExists(dir)
			assert.NoError(t, err)
			assert.True(t, exists, "Should exist: '%s'", dir)
		}
	})
	t.Run("invalid sub-folders do not exist after AddFile()", func(t *testing.T) {
		files := FilesMock{}
		path := filepath.Join("path", "to", "some", "file")
		files.AddFile(path, []byte("dummy content"))
		testDirs := []string{
			"to",
			filepath.Join("to", "some"),
			"some",
			filepath.Join("path", "to", "so"),
		}
		for _, dir := range testDirs {
			exists, err := files.DirExists(dir)
			assert.NoError(t, err)
			assert.False(t, exists, "Should not exist: '%s'", dir)
		}
	})
}

func TestFilesMockCopy(t *testing.T) {
	t.Parallel()
	t.Run("copy a previously added file successfully", func(t *testing.T) {
		files := FilesMock{}
		src := filepath.Join("some", "file")
		content := []byte("dummy content")
		files.AddFile(src, content)
		dst := filepath.Join("another", "file")
		length, err := files.Copy(src, dst)
		assert.NoError(t, err)
		assert.Equal(t, length, int64(len(content)))
	})
	t.Run("fail to copy non-existing file", func(t *testing.T) {
		files := FilesMock{}
		src := filepath.Join("some", "file")
		dst := filepath.Join("another", "file")
		length, err := files.Copy(src, dst)
		assert.EqualError(t, err, "cannot copy '"+src+"': file does not exist")
		assert.Equal(t, length, int64(0))
	})
	t.Run("fail to copy folder", func(t *testing.T) {
		files := FilesMock{}
		src := filepath.Join("some", "file")
		files.AddDir(src)
		dst := filepath.Join("another", "file")
		length, err := files.Copy(src, dst)
		assert.EqualError(t, err, "cannot copy '"+src+"': file does not exist")
		assert.Equal(t, length, int64(0))
	})
}

func TestFilesMockMkdirAll(t *testing.T) {
}

func TestFilesMockGetwd(t *testing.T) {
	t.Parallel()
	t.Run("test root", func(t *testing.T) {
		files := FilesMock{}
		dir, err := files.Getwd()
		assert.NoError(t, err)
		assert.Equal(t, string(os.PathSeparator), dir)
	})
	t.Run("test sub folder", func(t *testing.T) {
		files := FilesMock{}
		dirChain := []string{"some", "deep", "folder"}
		files.AddDir(filepath.Join(dirChain...))
		for _, element := range dirChain {
			err := files.Chdir(element)
			assert.NoError(t, err)
		}
		dir, err := files.Getwd()
		assert.NoError(t, err)
		assert.Equal(t, string(os.PathSeparator)+filepath.Join(dirChain...), dir)
	})
}
