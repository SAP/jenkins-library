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
		files.AddFile(path, []byte("dummy content"), 0644)
		exists, err := files.FileExists(path)
		assert.NoError(t, err)
		assert.True(t, exists)
	})
	t.Run("path is a directory after AddDir()", func(t *testing.T) {
		files := FilesMock{}
		path := filepath.Join("some", "path")
		files.AddDir(path, 0755)
		exists, err := files.FileExists(path)
		assert.NoError(t, err)
		assert.False(t, exists)
	})
	t.Run("file exists after changing current dir", func(t *testing.T) {
		files := FilesMock{}
		path := filepath.Join("some", "path")
		files.AddFile(path, []byte("dummy content"), 0644)
		err := files.Chdir("some")
		assert.NoError(t, err)
		exists, err := files.FileExists("path")
		assert.NoError(t, err)
		assert.True(t, exists)
	})
	t.Run("file does not exist after changing current dir", func(t *testing.T) {
		files := FilesMock{}
		path := filepath.Join("some", "path")
		files.AddFile(path, []byte("dummy content"), 0644)
		err := files.Chdir("some")
		assert.NoError(t, err)
		exists, err := files.FileExists(path)
		assert.NoError(t, err)
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
		files.AddDir(path, 0755)
		exists, err := files.DirExists(path)
		assert.NoError(t, err)
		assert.True(t, exists)
	})
	t.Run("absolute dir exists after AddDir()", func(t *testing.T) {
		files := FilesMock{}
		path := filepath.Join("some", "path")
		files.AddDir(path, 0755)
		err := files.Chdir("some")
		assert.NoError(t, err)
		exists, err := files.DirExists(string(os.PathSeparator) + path)
		assert.NoError(t, err)
		assert.True(t, exists)
	})
	t.Run("parent dirs exists after AddFile()", func(t *testing.T) {
		files := FilesMock{}
		path := filepath.Join("path", "to", "some", "file")
		files.AddFile(path, []byte("dummy content"), 0644)
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
		files.AddFile(path, []byte("dummy content"), 0644)
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
		files.AddFile(src, content, 0644)
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
		files.AddDir(src, 0755)
		dst := filepath.Join("another", "file")
		length, err := files.Copy(src, dst)
		assert.EqualError(t, err, "cannot copy '"+src+"': file does not exist")
		assert.Equal(t, length, int64(0))
	})
}

func TestFilesMockFileRemove(t *testing.T) {
	t.Parallel()
	t.Run("fail to remove non-existing file", func(t *testing.T) {
		files := FilesMock{}
		path := filepath.Join("some", "file")
		err := files.FileRemove(path)
		assert.EqualError(t, err, "the file '"+path+"' does not exist: file does not exist")
		assert.False(t, files.HasRemovedFile(path))
	})
	t.Run("track removing a file", func(t *testing.T) {
		files := FilesMock{}
		path := filepath.Join("some", "file")
		files.AddFile(path, []byte("dummy content"), 0644)
		assert.True(t, files.HasFile(path))
		err := files.FileRemove(path)
		assert.NoError(t, err)
		assert.False(t, files.HasFile(path))
		assert.True(t, files.HasRemovedFile(path))
	})
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
		files.AddDir(filepath.Join(dirChain...), 0755)
		for _, element := range dirChain {
			err := files.Chdir(element)
			assert.NoError(t, err)
		}
		dir, err := files.Getwd()
		assert.NoError(t, err)
		assert.Equal(t, string(os.PathSeparator)+filepath.Join(dirChain...), dir)
	})
}

func TestFilesMockGlob(t *testing.T) {
	t.Parallel()

	files := FilesMock{}
	content := []byte("dummy content")
	files.AddFile(filepath.Join("dir", "foo.xml"), content, 0644)
	files.AddFile(filepath.Join("dir", "another", "foo.xml"), content, 0644)
	files.AddFile(filepath.Join("dir", "baz"), content, 0644)
	files.AddFile(filepath.Join("baz.xml"), content, 0644)

	t.Run("one match in root-dir", func(t *testing.T) {
		matches, err := files.Glob("*.xml")
		assert.NoError(t, err)
		if assert.Len(t, matches, 1) {
			assert.Equal(t, matches[0], "baz.xml")
		}
	})
	t.Run("three matches in all levels", func(t *testing.T) {
		matches, err := files.Glob("**/*.xml")
		assert.NoError(t, err)
		if assert.Len(t, matches, 3) {
			assert.Equal(t, matches[0], "baz.xml")
			assert.Equal(t, matches[1], filepath.Join("dir", "another", "foo.xml"))
			assert.Equal(t, matches[2], filepath.Join("dir", "foo.xml"))
		}
	})
	t.Run("match only in sub-dir", func(t *testing.T) {
		matches, err := files.Glob("*/*.xml")
		assert.NoError(t, err)
		if assert.Len(t, matches, 1) {
			assert.Equal(t, matches[0], filepath.Join("dir", "foo.xml"))
		}
	})
	t.Run("match for two levels", func(t *testing.T) {
		matches, err := files.Glob("*/*/*.xml")
		assert.NoError(t, err)
		if assert.Len(t, matches, 1) {
			assert.Equal(t, matches[0], filepath.Join("dir", "another", "foo.xml"))
		}
	})
	t.Run("match prefix", func(t *testing.T) {
		matches, err := files.Glob("**/baz*")
		assert.NoError(t, err)
		if assert.Len(t, matches, 2) {
			assert.Equal(t, matches[0], filepath.Join("baz.xml"))
			assert.Equal(t, matches[1], filepath.Join("dir", "baz"))
		}
	})
	t.Run("no matches", func(t *testing.T) {
		matches, err := files.Glob("**/*bar*")
		assert.NoError(t, err)
		assert.Len(t, matches, 0)
	})
}
