//go:build unit
// +build unit

package mock

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	t.Run("root folder always exists", func(t *testing.T) {
		files := FilesMock{}
		exists, err := files.DirExists(string(os.PathSeparator))
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
	t.Run("absolute dir exists after AddDir()", func(t *testing.T) {
		files := FilesMock{}
		path := filepath.Join("some", "path")
		files.AddDir(path)
		err := files.Chdir("some")
		assert.NoError(t, err)
		exists, err := files.DirExists(string(os.PathSeparator) + path)
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
	t.Run("dir still exists after removing last file", func(t *testing.T) {
		files := FilesMock{}
		dir := filepath.Join("path", "to")
		file := filepath.Join(dir, "file")
		files.AddFile(file, []byte("dummy content"))
		err := files.FileRemove(file)
		assert.NoError(t, err)
		exists, err := files.DirExists(dir)
		assert.NoError(t, err)
		assert.True(t, exists)
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

func TestFilesMockFileRemove(t *testing.T) {
	t.Parallel()
	t.Run("fail to remove non-existing file", func(t *testing.T) {
		files := FilesMock{}
		path := filepath.Join("some", "file")
		err := files.FileRemove(path)
		assert.EqualError(t, err, "the file '"+path+"' does not exist: file does not exist")
		assert.False(t, files.HasRemovedFile(path))
	})
	t.Run("fail to remove non-empty directory", func(t *testing.T) {
		files := FilesMock{}
		path := filepath.Join("dir", "file")
		files.AddFile(path, []byte("dummy content"))
		err := files.FileRemove("dir")
		assert.Error(t, err)
	})
	t.Run("fail to remove non-empty directory also when it was explicitly added", func(t *testing.T) {
		files := FilesMock{}
		path := filepath.Join("dir", "file")
		files.AddFile(path, []byte("dummy content"))
		files.AddDir("dir")
		err := files.FileRemove("dir")
		assert.Error(t, err)
	})
	t.Run("succeed to remove empty directory when it was explicitly added", func(t *testing.T) {
		files := FilesMock{}
		files.AddDir("dir")
		err := files.FileRemove("dir")
		assert.NoError(t, err)
	})
	t.Run("removing chain of entries works", func(t *testing.T) {
		files := FilesMock{}
		path := filepath.Join("path", "to", "file")
		files.AddFile(path, []byte("dummy content"))
		assert.NoError(t, files.FileRemove(filepath.Join("path", "to", "file")))
		assert.NoError(t, files.FileRemove(filepath.Join("path", "to")))
		assert.NoError(t, files.FileRemove(filepath.Join("path")))
		assert.Len(t, files.files, 0)
	})
	t.Run("removing entry from current dir works", func(t *testing.T) {
		files := FilesMock{}
		path := filepath.Join("path", "to", "file")
		files.AddFile(path, []byte("dummy content"))

		err := files.Chdir("path")
		assert.NoError(t, err)

		assert.NoError(t, files.FileRemove(filepath.Join("to", "file")))
		assert.NoError(t, files.FileRemove(filepath.Join("to")))

		err = files.Chdir("/")
		assert.NoError(t, err)

		assert.NoError(t, files.FileRemove(filepath.Join("path")))
		assert.Len(t, files.files, 0)
	})
	t.Run("track removing a file", func(t *testing.T) {
		files := FilesMock{}
		path := filepath.Join("some", "file")
		files.AddFile(path, []byte("dummy content"))
		assert.True(t, files.HasFile(path))
		err := files.FileRemove(path)
		assert.NoError(t, err)
		assert.False(t, files.HasFile(path))
		assert.True(t, files.HasRemovedFile(path))
	})
}

func TestFilesMockFileRename(t *testing.T) {
	t.Parallel()
	t.Run("fail to rename non-existing file (no init)", func(t *testing.T) {
		files := FilesMock{}
		oldPath := filepath.Join("foo", "bar")
		newPath := filepath.Join("foo", "baz")
		err := files.FileRename(oldPath, newPath)
		assert.EqualError(t, err, "the file '"+oldPath+"' does not exist: file does not exist")
	})
	t.Run("fail to rename non-existing file", func(t *testing.T) {
		files := FilesMock{}
		files.AddDir("triggers/initialization")
		oldPath := filepath.Join("foo", "bar")
		newPath := filepath.Join("foo", "baz")
		err := files.FileRename(oldPath, newPath)
		assert.EqualError(t, err,
			"renaming file '"+oldPath+"' is not supported, since it does not exist, or is not a leaf-entry")
	})
	t.Run("fail to rename non-existing file, even no-op", func(t *testing.T) {
		files := FilesMock{}
		files.AddDir("triggers/initialization")
		oldPath := filepath.Join("foo", "bar")
		newPath := oldPath
		err := files.FileRename(oldPath, newPath)
		assert.EqualError(t, err,
			"renaming file '"+oldPath+"' is not supported, since it does not exist, or is not a leaf-entry")
	})
	t.Run("success to rename dir (no-op)", func(t *testing.T) {
		files := FilesMock{}
		oldPath := filepath.Join("foo", "bar")
		files.AddDir(oldPath)
		newPath := oldPath
		err := files.FileRename(oldPath, newPath)
		assert.NoError(t, err)
		assert.True(t, files.HasFile(newPath))
	})
	t.Run("success to rename dir", func(t *testing.T) {
		files := FilesMock{}
		oldPath := filepath.Join("foo", "bar")
		files.AddDir(oldPath)
		newPath := filepath.Join("foo", "baz")
		err := files.FileRename(oldPath, newPath)
		assert.NoError(t, err)
		assert.True(t, files.HasFile(newPath))
		assert.False(t, files.HasFile(oldPath))
	})
	t.Run("success to rename file", func(t *testing.T) {
		files := FilesMock{}
		oldPath := filepath.Join("foo", "bar")
		files.AddFile(oldPath, []byte("dummy contents"))
		newPath := filepath.Join("foo", "baz")
		err := files.FileRename(oldPath, newPath)
		assert.NoError(t, err)
		assert.True(t, files.HasFile(newPath))
		assert.False(t, files.HasFile(oldPath))
	})
	t.Run("fail to rename file, already exists", func(t *testing.T) {
		files := FilesMock{}
		oldPath := filepath.Join("foo", "bar")
		newPath := filepath.Join("foo", "baz")
		files.AddFile(oldPath, []byte("dummy contents"))
		files.AddFile(newPath, []byte("dummy contents"))
		err := files.FileRename(oldPath, newPath)
		assert.EqualError(t, err, "cannot rename '"+oldPath+"', target path '"+newPath+"' already exists")
		assert.True(t, files.HasFile(newPath))
		assert.True(t, files.HasFile(oldPath))
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

func TestFilesMockGlob(t *testing.T) {
	t.Parallel()

	files := FilesMock{}
	content := []byte("dummy content")
	files.AddFile(filepath.Join("dir", "foo.xml"), content)
	files.AddFile(filepath.Join("dir", "another", "foo.xml"), content)
	files.AddFile(filepath.Join("dir", "baz"), content)
	files.AddFile(filepath.Join("baz.xml"), content)

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

func TestStat(t *testing.T) {
	t.Parallel()
	t.Run("non existing file (no init)", func(t *testing.T) {
		// init
		files := FilesMock{}
		// test
		_, err := files.Stat("doesNotExist.txt")
		assert.EqualError(t, err, "stat doesNotExist.txt: no such file or directory")
	})
	t.Run("non existing file", func(t *testing.T) {
		// init
		files := FilesMock{}
		files.AddFile("tmp/dummy.txt", []byte("Hello SAP"))
		// test
		_, err := files.Stat("doesNotExist.txt")
		assert.EqualError(t, err, "stat doesNotExist.txt: no such file or directory")
	})
	t.Run("check file info", func(t *testing.T) {
		// init
		files := FilesMock{}
		files.AddFile("tmp/dummy.txt", []byte("Hello SAP"))
		// test
		info, err := files.Stat("tmp/dummy.txt")
		// assert
		if assert.NoError(t, err) {
			// only the base name is returned.
			assert.Equal(t, "dummy.txt", info.Name())
			assert.False(t, info.IsDir())
			// if not specified otherwise the 644 file mode is used.
			assert.Equal(t, defaultFileMode, info.Mode())
		}
	})
	t.Run("check implicit dir", func(t *testing.T) {
		// init
		files := FilesMock{}
		files.AddFile("tmp/dummy.txt", []byte("Hello SAP"))
		// test
		info, err := files.Stat("tmp")
		// assert
		if assert.NoError(t, err) {
			assert.True(t, info.IsDir())
			assert.Equal(t, defaultDirMode, info.Mode())
		}
	})
	t.Run("check explicit dir", func(t *testing.T) {
		// init
		files := FilesMock{}
		explicitMode := os.FileMode(0700)
		files.AddDirWithMode("bin", explicitMode)
		// test
		info, err := files.Stat("bin")
		// assert
		if assert.NoError(t, err) {
			assert.True(t, info.IsDir())
			assert.Equal(t, explicitMode, info.Mode())
		}
	})
}

func TestChmod(t *testing.T) {
	files := FilesMock{}
	files.AddFileWithMode("tmp/log.txt", []byte("build failed"), 0777)

	t.Run("non existing file", func(t *testing.T) {
		err := files.Chmod("does/not/exist", 0400)
		assert.EqualError(t, err, "chmod: does/not/exist: No such file or directory")
	})
	t.Run("chmod for file", func(t *testing.T) {
		err := files.Chmod("tmp/log.txt", 0645)
		if assert.NoError(t, err) {
			info, e := files.Stat("tmp/log.txt")
			if assert.NoError(t, e) {
				assert.Equal(t, os.FileMode(0645), info.Mode())
			}
		}
	})
	t.Run("chmod for directory", func(t *testing.T) {
		err := files.Chmod("tmp", 0766)
		if assert.NoError(t, err) {
			info, e := files.Stat("tmp")
			if assert.NoError(t, e) {
				assert.Equal(t, os.FileMode(0766), info.Mode())
			}
		}
	})
}

func TestRelativePaths(t *testing.T) {
	t.Parallel()
	t.Run("files are not mixed up", func(t *testing.T) {
		files := FilesMock{}
		files.AddFile("path/to/file", []byte("content"))
		files.AddFile("file", []byte("root-content"))

		err := files.Chdir("path")
		if assert.NoError(t, err) {
			exists, _ := files.FileExists("file")
			assert.False(t, exists)

			err := files.Chdir("to")
			if assert.NoError(t, err) {
				content, err := files.FileRead("file")
				if assert.NoError(t, err) {
					assert.Equal(t, []byte("content"), content, "should not read root file")
				}
			}
		}
	})
	t.Run("root folder exists after change dir", func(t *testing.T) {
		files := FilesMock{}
		files.AddFile("path/to/file", []byte("content"))

		errChdirInto := files.Chdir("path")
		assert.NoError(t, errChdirInto)

		exists, err := files.DirExists("/")
		assert.NoError(t, err)
		assert.True(t, exists, "the root folder should exist no matter the current dir")
	})
	t.Run("current folder always exists", func(t *testing.T) {
		files := FilesMock{}
		files.AddFile("path/to/file", []byte("content"))

		exists, err := files.DirExists(".")
		assert.NoError(t, err)
		assert.True(t, exists, "the current folder should exist")

		errChdirInto := files.Chdir("path")
		assert.NoError(t, errChdirInto)

		exists, err = files.DirExists("./")
		assert.NoError(t, err)
		assert.True(t, exists, "the current folder should exist after changing into it")
	})
	t.Run("chmod works on implicit, relative dir", func(t *testing.T) {
		files := FilesMock{}
		files.AddFile("path/to/file", []byte("content"))

		errChdirInto := files.Chdir("path")
		errChmod := files.Chmod("to", 0700)
		errChdirBack := files.Chdir("/")

		assert.NoError(t, errChdirInto)
		assert.NoError(t, errChmod)
		assert.NoError(t, errChdirBack)

		fileInfo, err := files.Stat("path/to")
		if assert.NoError(t, err) {
			assert.Equal(t, os.FileMode(0700), fileInfo.Mode())
		}
	})
}

func TestOpen(t *testing.T) {
	filePath := filepath.Join("some", "file")
	t.Parallel()
	t.Run("no init without O_CREATE", func(t *testing.T) {
		// init
		files := FilesMock{}
		// test
		file, err := files.OpenFile(filePath, 0, 0)
		// assert
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "does not exist")
			assert.Nil(t, file)
		}
	})
	t.Run("no init with O_CREATE", func(t *testing.T) {
		// init
		files := FilesMock{}
		// test
		file, err := files.OpenFile(filePath, os.O_CREATE, 0644)
		// assert
		if assert.NoError(t, err) && assert.NotNil(t, file) {
			assert.Equal(t, &files, file.files)
			assert.Equal(t, files.Separator+filePath, file.absPath)
			assert.NotNil(t, file.content)
		}
	})
	t.Run("content is replaced without O_APPEND", func(t *testing.T) {
		// init
		files := FilesMock{}
		files.AddFile(filePath, []byte("initial-content"))
		// test
		file, _ := files.OpenFile(filePath, os.O_CREATE, 0644)
		written, err := file.WriteString("hello")
		if assert.NoError(t, err) {
			assert.Equal(t, written, len("hello"))
			content, err := files.FileRead(filePath)
			if assert.NoError(t, err) {
				assert.Equal(t, []byte("hello"), content)
			}
		}
	})
	t.Run("content is truncated with O_TRUNC an nothing written", func(t *testing.T) {
		// init
		files := FilesMock{}
		files.AddFile(filePath, []byte("initial-content"))
		// test
		file, err := files.OpenFile(filePath, os.O_CREATE|os.O_TRUNC, 0644)
		require.NoError(t, err)
		err = file.Close()
		assert.NoError(t, err)
		content, err := files.FileRead(filePath)
		if assert.NoError(t, err) {
			assert.Len(t, content, 0)
		}
	})
	t.Run("content is appended", func(t *testing.T) {
		// init
		files := FilesMock{}
		files.AddFile(filePath, []byte("initial-content"))
		// test
		file, _ := files.OpenFile(filePath, os.O_APPEND, 0644)
		written1, err1 := file.WriteString("-hel")
		written2, err2 := file.WriteString("lo")
		if assert.NoError(t, err1) && assert.NoError(t, err2) {
			assert.Equal(t, written1+written2, len("-hello"))
			content, err := files.FileRead(filePath)
			if assert.NoError(t, err) {
				assert.Equal(t, []byte("initial-content-hello"), content)
			}
		}
	})
	t.Run("cannot write to closed file", func(t *testing.T) {
		// init
		files := FilesMock{}
		files.AddFile(filePath, []byte("initial-content"))
		// test
		file, _ := files.OpenFile(filePath, os.O_APPEND, 0644)
		_, err := file.WriteString("-hello")
		assert.NoError(t, err)
		err = file.Close()
		assert.NoError(t, err)
		_, err = file.WriteString("-more")
		assert.Error(t, err)
		content, err := files.FileRead(filePath)
		if assert.NoError(t, err) {
			assert.Equal(t, []byte("initial-content-hello"), content)
		}
	})
}

func TestFilesMockTempDir(t *testing.T) {
	t.Parallel()
	t.Run("creates a temp dir without a pattern", func(t *testing.T) {
		files := FilesMock{}
		dir, err := files.TempDir("", "")
		assert.NoError(t, err)
		assert.Equal(t, "/tmp", dir)
		ok, err := files.DirExists("/tmp")
		assert.NoError(t, err)
		assert.True(t, ok)
	})
	t.Run("creates a temp dir with a pattern", func(t *testing.T) {
		files := FilesMock{}
		dir, err := files.TempDir("", "pattern")
		assert.NoError(t, err)
		assert.Equal(t, "/tmp/patterntest", dir)
		ok, err := files.DirExists("/tmp/patterntest")
		assert.NoError(t, err)
		assert.True(t, ok)
	})
}

func TestFilesMockSymlink(t *testing.T) {
	t.Parallel()
	t.Run("creates a symlink", func(t *testing.T) {
		files := FilesMock{}
		files.AddDir("/backup")
		assert.NoError(t, files.Symlink("/folder", "/backup/folder"))

		assert.True(t, files.HasCreatedSymlink("/folder", "/backup/folder"))
	})

	t.Run("fails if parent directory doesn't exist", func(t *testing.T) {
		files := FilesMock{}
		err := files.Symlink("/non/existent/folder", "/symbolic/link")
		assert.Error(t, err)
		assert.Equal(t, "failed to create symlink: parent directory /symbolic doesn't exist", err.Error())
	})

	t.Run("fails if FileWriteError is specified", func(t *testing.T) {
		expectedErr := errors.New("test")
		files := FilesMock{
			FileWriteErrors: map[string]error{
				"/symbolic/link": expectedErr,
			},
		}
		err := files.Symlink("/non/existent/folder", "/symbolic/link")
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}

func TestFilesMockCreateArchive(t *testing.T) {
	t.Parallel()
	t.Run("creates an archive with the provided content", func(t *testing.T) {
		files := FilesMock{}
		a, err := files.CreateArchive(map[string][]byte{
			"filename": []byte("file content"),
		})
		assert.NoError(t, err)

		buf := bytes.NewBuffer(a)
		zr, err := gzip.NewReader(buf)
		assert.NoError(t, err)
		defer zr.Close()

		tr := tar.NewReader(zr)

		for {
			f, err := tr.Next()
			if err == io.EOF {
				break
			}
			assert.NoError(t, err)
			assert.Equal(t, "filename", f.Name)

			content, err := io.ReadAll(tr)
			assert.NoError(t, err)
			assert.Equal(t, "file content", string(content))
		}
	})

	t.Run("fails if the content is empty", func(t *testing.T) {
		files := FilesMock{}
		a, err := files.CreateArchive(nil)
		assert.Error(t, err)
		assert.Equal(t, "mock archive content must not be empty", err.Error())
		assert.Nil(t, a)
	})
}
