package piperutils

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileExists(t *testing.T) {
	runInTempDir(t, "testing dir returns false", "dir", func(t *testing.T) {
		err := os.Mkdir("test", 0777)
		if err != nil {
			t.Fatal("failed to create test dir in temporary dir")
		}
		result, err := FileExists("test")
		assert.NoError(t, err)
		assert.False(t, result)
	})
	runInTempDir(t, "testing file returns true", "dir", func(t *testing.T) {
		file, err := ioutil.TempFile("", "testFile")
		assert.NoError(t, err)
		result, err := FileExists(file.Name())
		assert.NoError(t, err)
		assert.True(t, result)
	})
}

func TestDirExists(t *testing.T) {
	runInTempDir(t, "testing dir exists", "dir-exists", func(t *testing.T) {
		err := os.Mkdir("test", 0777)
		if err != nil {
			t.Fatal("failed to create test dir in temporary dir")
		}
		files := Files{}

		result, err := files.DirExists("test")
		assert.NoError(t, err)
		assert.True(t, result, "created folder should exist")

		result, err = files.DirExists(".")
		assert.NoError(t, err)
		assert.True(t, result, "current directory should exist")

		result, err = files.DirExists(string(os.PathSeparator))
		assert.NoError(t, err)
		assert.True(t, result, "root directory should exist")
	})
}

func TestCopy(t *testing.T) {
	runInTempDir(t, "copying file succeeds", "dir2", func(t *testing.T) {
		file := "testFile"
		err := ioutil.WriteFile(file, []byte{byte(1), byte(2), byte(3)}, 0700)
		if err != nil {
			t.Fatal("Failed to create temporary workspace directory")
		}

		result, err := Copy(file, "testFile2")
		assert.NoError(t, err, "Didn't expert error but got one")
		assert.Equal(t, int64(3), result, "Expected true but got false")
	})
	runInTempDir(t, "copying directory fails", "dir3", func(t *testing.T) {
		src := filepath.Join("some", "file")
		dst := filepath.Join("another", "file")

		err := os.MkdirAll(src, 0777)
		if err != nil {
			t.Fatal("Failed to create test directory")
		}

		files := Files{}
		exists, err := files.DirExists(src)
		assert.NoError(t, err)
		assert.True(t, exists)

		length, err := files.Copy(src, dst)
		assert.EqualError(t, err, "Source file '"+src+"' does not exist")
		assert.Equal(t, length, int64(0))
	})
}

func runInTempDir(t *testing.T, nameOfRun, tempDirPattern string, run func(t *testing.T)) {
	dir, err := ioutil.TempDir("", tempDirPattern)
	if err != nil {
		t.Fatal("Failed to create temporary directory")
	}
	oldCWD, _ := os.Getwd()
	_ = os.Chdir(dir)
	// clean up tmp dir
	defer func() {
		_ = os.Chdir(oldCWD)
		_ = os.RemoveAll(dir)
	}()

	t.Run(nameOfRun, run)
}

func TestExcludeFiles(t *testing.T) {
	t.Parallel()
	t.Run("nil slices", func(t *testing.T) {
		t.Parallel()
		filtered, err := ExcludeFiles(nil, nil)
		assert.NoError(t, err)
		assert.Len(t, filtered, 0)
	})
	t.Run("empty excludes", func(t *testing.T) {
		t.Parallel()
		files := []string{"file"}
		filtered, err := ExcludeFiles(files, nil)
		assert.NoError(t, err)
		assert.Equal(t, files, filtered)
	})
	t.Run("direct match", func(t *testing.T) {
		t.Parallel()
		files := []string{"file"}
		filtered, err := ExcludeFiles(files, files)
		assert.NoError(t, err)
		assert.Len(t, filtered, 0)
	})
	t.Run("two direct matches", func(t *testing.T) {
		t.Parallel()
		files := []string{"a", "b"}
		filtered, err := ExcludeFiles(files, files)
		assert.NoError(t, err)
		assert.Len(t, filtered, 0)
	})
	t.Run("one direct exclude matches", func(t *testing.T) {
		t.Parallel()
		files := []string{"a", "b"}
		filtered, err := ExcludeFiles(files, []string{"b"})
		assert.NoError(t, err)
		assert.Equal(t, []string{"a"}, filtered)
	})
	t.Run("no glob matches", func(t *testing.T) {
		t.Parallel()
		files := []string{"a", "b"}
		filtered, err := ExcludeFiles(files, []string{"*/a", "b/*"})
		assert.NoError(t, err)
		assert.Equal(t, []string{"a", "b"}, filtered)
	})
	t.Run("two globs match", func(t *testing.T) {
		t.Parallel()
		files := []string{"path/to/a", "b"}
		filtered, err := ExcludeFiles(files, []string{"**/a", "**/b"})
		assert.NoError(t, err)
		assert.Len(t, filtered, 0)
	})
}

func TestUntar(t *testing.T) {
	runInTempDir(t, "untar a file at level 0", "dir", func(t *testing.T) {
		//Create a temp file
		file, err := ioutil.TempFile(".", "testFile")
		filename := file.Name()
		assert.NoError(t, err)
		//Create an archive
		cmd := exec.Command("tar", "-czf", "test.tgz", filename)
		if err := cmd.Run(); err != nil {
			t.Fatal("Failed to create dummy tar archive")
		}
		result, err := FileExists("test.tgz")
		assert.NoError(t, err)
		assert.True(t, result)
		//Remove the temp file
		if err := file.Close(); err != nil {
			t.Fatal("Failed to close temp file")
		}
		os.Remove(filename)
		result, err = FileExists(filename)
		assert.NoError(t, err)
		assert.False(t, result)
		//Untar
		err = Untar("test.tgz", ".", 0)
		//Check presence of file
		result, err = FileExists(filename)
		assert.NoError(t, err)
		assert.True(t, result)
	})

	runInTempDir(t, "untar a file in a folder at level 0", "dir", func(t *testing.T) {
		//Create a temp dir
		err := os.Mkdir("test", 0777)
		if err != nil {
			t.Fatal("failed to create test dir in temporary dir")
		}
		//Create a temp file
		file, err := ioutil.TempFile("./test", "testFile")
		filename := file.Name()
		assert.NoError(t, err)
		//Create an archive
		cmd := exec.Command("tar", "-czf", "test.tgz", filename)
		if err := cmd.Run(); err != nil {
			t.Fatal("Failed to create dummy tar archive")
		}
		result, err := FileExists("test.tgz")
		assert.NoError(t, err)
		assert.True(t, result)
		//Remove the temp file
		if err := file.Close(); err != nil {
			t.Fatal("Failed to close temp file")
		}
		os.Remove(filename)
		result, err = FileExists(filename)
		assert.NoError(t, err)
		assert.False(t, result)
		//Test dir has been removed
		os.Remove("test")
		_, err = os.Stat("test")
		assert.True(t, os.IsNotExist(err))
		//Untar
		err = Untar("test.tgz", ".", 0)
		//Check presence of file: filename has the whole path so should be true
		result, err = FileExists(filename)
		assert.NoError(t, err)
		assert.True(t, result)
	})

	runInTempDir(t, "untar a file at level 1", "dir", func(t *testing.T) {
		//Create a temp dir
		err := os.Mkdir("test", 0777)
		if err != nil {
			t.Fatal("failed to create test dir in temporary dir")
		}
		//Create a temp file
		file, err := ioutil.TempFile("./test", "testFile")
		filename := file.Name()
		assert.NoError(t, err)
		//Create an archive
		cmd := exec.Command("tar", "-czf", "test.tgz", filename)
		if err := cmd.Run(); err != nil {
			t.Fatal("Failed to create dummy tar archive")
		}
		result, err := FileExists("test.tgz")
		assert.NoError(t, err)
		assert.True(t, result)
		//Remove the temp file
		if err := file.Close(); err != nil {
			t.Fatal("Failed to close temp file")
		}
		os.Remove(filename)
		result, err = FileExists(filename)
		assert.NoError(t, err)
		assert.False(t, result)
		//Test dir has been removed
		os.Remove("test")
		_, err = os.Stat("test")
		assert.True(t, os.IsNotExist(err))
		//Untar
		err = Untar("test.tgz", ".", 1)
		//Check presence of file: filename has the whole path so should be false
		result, err = FileExists(filename)
		assert.NoError(t, err)
		assert.False(t, result)
		//Split filename, retry
		result, err = FileExists(strings.Split(filename, string(os.PathSeparator))[1])
		assert.NoError(t, err)
		assert.True(t, result)
	})

	runInTempDir(t, "untar a non-tar file: error", "dir", func(t *testing.T) {
		//Create a temp file
		file, err := ioutil.TempFile(".", "testFile")
		assert.NoError(t, err)
		//Untar
		err = Untar(file.Name(), ".", 0)
		assert.EqualError(t, err, "requires gzip-compressed body: EOF")
	})

	runInTempDir(t, "untar a level 0 file at level 1: error", "dir", func(t *testing.T) {
		//Create a temp file
		file, err := ioutil.TempFile(".", "testFile")
		filename := file.Name()
		assert.NoError(t, err)
		//Create an archive
		cmd := exec.Command("tar", "-czf", "test.tgz", filename)
		if err := cmd.Run(); err != nil {
			t.Fatal("Failed to create dummy tar archive")
		}
		result, err := FileExists("test.tgz")
		assert.NoError(t, err)
		assert.True(t, result)
		//Untar
		err = Untar("test.tgz", ".", 1)
		//Check presence of file
		assert.EqualError(t, err, "files \""+filename+"\" in tarball archive not under level 1")
	})
}
