package piperutils

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	result, err := Copy(file, filepath.Join(dir, "testFile2"))
	assert.NoError(t, err, "Didn't expert error but got one")
	assert.Equal(t, int64(3), result, "Expected true but got false")
}

func TestFindFiles(t *testing.T) {

	workdir, err := ioutil.TempDir("", "findFiles")
	defer os.RemoveAll(workdir)
	require.NoError(t, err)
	_, err = os.Create(workdir + "/temp.txt")
	require.NoError(t, err)
	err = os.MkdirAll(workdir+"/1/2/3", 0775)
	require.NoError(t, err)
	_, err = os.Create(workdir + "/1/2/3/my.txt")
	require.NoError(t, err)
	_, err = os.Create(workdir + "/1/2/this.one.not")
	require.NoError(t, err)

	files, err := FindFiles(workdir, ".*/2/.*\\.txt$")

	assert.ElementsMatch(t, []string{workdir + "/1/2/3/my.txt", workdir + "/temp.txt"}, files)
}
