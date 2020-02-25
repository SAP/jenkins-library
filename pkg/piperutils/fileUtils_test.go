package piperutils

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestFileExists(t *testing.T) {
	dir, err := ioutil.TempDir("", "dir")
	if err != nil {
		t.Fatal("Failed to create temporary workspace directory")
	}
	// clean up tmp dir
	defer os.RemoveAll(dir)

	result, err := FileExists(dir)
	assert.NoError(t, err, "Didn't expert error but got one")
	assert.Equal(t, false, result, "Expected false but got true")

	file, err := ioutil.TempFile(dir, "testFile")
	assert.NoError(t, err, "Didn't expert error but got one")
	result, err = FileExists(file.Name())
	assert.NoError(t, err, "Didn't expert error but got one")
	assert.Equal(t, true, result, "Expected true but got false")
}

func TestCopy(t *testing.T) {
	dir, err := ioutil.TempDir("", "dir2")
	file := filepath.Join(dir, "testFile")
	err = ioutil.WriteFile(file, []byte{byte(1), byte(2), byte(3)}, 0700)
	if err != nil {
		t.Fatal("Failed to create temporary workspace directory")
	}
	// clean up tmp dir
	defer os.RemoveAll(dir)

	result, err := Copy(file, filepath.Join(dir, "testFile2"), false)
	assert.NoError(t, err, "Didn't expert error but got one")
	assert.Equal(t, int64(3), result, "Expected true but got false")
}

func TestCopyIntoNestedFoldersWhichDoesNotExist(t *testing.T) {
	dir, err := ioutil.TempDir("", "dir2")
	file := filepath.Join(dir, "testFile")
	err = ioutil.WriteFile(file, []byte{byte(1), byte(2), byte(3)}, 0700)
	if err != nil {
		t.Fatal("Failed to create temporary workspace directory")
	}
	// clean up tmp dir
	defer os.RemoveAll(dir)

	result, err := Copy(file, filepath.Join(filepath.Join(filepath.Join(dir, "1"), "2"), "testFile2"), true)
	assert.NoError(t, err, "Didn't expert error but got one")
	assert.Equal(t, int64(3), result, "Expected true but got false")
}

func TestCopyIntoNestedFoldersWhichDoesNotExistDontCreateMissingFolders(t *testing.T) {
	dir, err := ioutil.TempDir("", "dir2")
	file := filepath.Join(dir, "testFile")
	err = ioutil.WriteFile(file, []byte{byte(1), byte(2), byte(3)}, 0700)
	if err != nil {
		t.Fatal("Failed to create temporary workspace directory")
	}
	// clean up tmp dir
	defer os.RemoveAll(dir)

	dest := filepath.Join(filepath.Join(filepath.Join(dir, "1"), "2"), "testFile2")
	_, err = Copy(file, dest, false)
	if assert.Error(t, err) {
		_, err := os.Stat(dest)
		assert.True(t, os.IsNotExist(err))
		assert.Equal(t, fmt.Sprintf("Parent folder for file '%s/1/2/3/testFile2' does not exist, createMissingDirectories was 'false'", dir), err.Error())
	}
}
