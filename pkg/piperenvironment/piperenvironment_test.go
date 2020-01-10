package piperenvironment

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Do we still need this?
/*
func TestSetArtifactVersion(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal("Failed to create temporary directory")
	}

	// clean up tmp dir
	defer os.RemoveAll(dir)

	err = SetArtifactVersion(dir, "1.0.0")
	assert.NoError(t, err, "Error occured but none expected")

	_, err = os.Stat(filepath.Join(dir, fileArtifactVersion))
	assert.NoError(t, err, "Expected file does not exist")

	assert.Equal(t, "1.0.0", GetArtifactVersion(dir))
}

func TestSetGitCommitID(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal("Failed to create temporary directory")
	}

	// clean up tmp dir
	defer os.RemoveAll(dir)

	err = SetGitCommitID(dir, "commitSHA")
	assert.NoError(t, err, "Error occured but none expected")

	_, err = os.Stat(filepath.Join(dir, folderGit, fileCommitID))
	assert.NoError(t, err, "Expected file does not exist")

	assert.Equal(t, "commitSHA", GetGitCommitID(dir))
}

func TestSetGitBranch(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal("Failed to create temporary directory")
	}

	// clean up tmp dir
	defer os.RemoveAll(dir)

	err = SetGitBranch(dir, "testBranch")
	assert.NoError(t, err, "Error occured but none expected")

	_, err = os.Stat(filepath.Join(dir, folderGit, fileBranch))
	assert.NoError(t, err, "Expected file does not exist")

	assert.Equal(t, "testBranch", GetGitBranch(dir))
}

func TestSetGitCommitMessage(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal("Failed to create temporary directory")
	}

	// clean up tmp dir
	defer os.RemoveAll(dir)

	err = SetGitCommitMessage(dir, "testCommitMessage")
	assert.NoError(t, err, "Error occured but none expected")

	_, err = os.Stat(filepath.Join(dir, folderGit, fileCommitMessage))
	assert.NoError(t, err, "Expected file does not exist")

	assert.Equal(t, "testCommitMessage", GetGitCommitMessage(dir))
}

func TestSetGithubOwner(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal("Failed to create temporary directory")
	}

	// clean up tmp dir
	defer os.RemoveAll(dir)

	err = SetGithubOwner(dir, "testOwner")
	assert.NoError(t, err, "Error occured but none expected")

	_, err = os.Stat(filepath.Join(dir, folderGithub, fileOwner))
	assert.NoError(t, err, "Expected file does not exist")

	assert.Equal(t, "testOwner", GetGithubOwner(dir))
}

func TestSetGithubRepository(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal("Failed to create temporary directory")
	}

	// clean up tmp dir
	defer os.RemoveAll(dir)

	err = SetGithubRepository(dir, "testRepo")
	assert.NoError(t, err, "Error occured but none expected")

	_, err = os.Stat(filepath.Join(dir, folderGithub, fileRepo))
	assert.NoError(t, err, "Expected file does not exist")

	assert.Equal(t, "testRepo", GetGithubRepository(dir))
}

func TestSetCustomParameter(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal("Failed to create temporary directory")
	}

	// clean up tmp dir
	defer os.RemoveAll(dir)

	err = SetCustomParameter(dir, "testParam", "testVal")

	assert.NoError(t, err, "Error occured but none expected")
	assert.Equal(t, "testVal", GetCustomParameter(dir, "testParam"))
}
*/

func TestSetParameter(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal("Failed to create temporary directory")
	}

	// clean up tmp dir
	defer os.RemoveAll(dir)

	err = SetParameter(dir, "testParam", "testVal")

	assert.NoError(t, err, "Error occured but none expected")
	assert.Equal(t, "testVal", GetParameter(dir, "testParam"))
}

func TestReadFromDisk(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal("Failed to create temporary directory")
	}

	// clean up tmp dir
	defer os.RemoveAll(dir)

	assert.Equal(t, "", GetParameter(dir, "testParamNotExistingYet"))
}
