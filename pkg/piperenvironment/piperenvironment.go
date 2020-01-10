package piperenvironment

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// This file contains functions used to read/write pipeline environment data from/to disk.
// The content of the file is the value. For the custom parameters this could for example also be a JSON representation of a more complex value.
// The file representation looks as follows:

// <pipeline env path>/

// <pipeline env path>/artifactVersion

// <pipeline env path>/git/
// <pipeline env path>/git/branch
// <pipeline env path>/git/commitId
// <pipeline env path>/git/commitMessage
// <pipeline env path>/git/repositoryUrl -> TODO: storing function(s) with ssh and https getters

// <pipeline env path>/github/
// <pipeline env path>/github/owner
// <pipeline env path>/github/repository

// <pipeline env path>/custom/
// <pipeline env path>/custom/<parameter>

const fileArtifactVersion = "artifactVersion"

const folderGit = "git"
const fileBranch = "branch"
const fileCommitID = "commitId"
const fileCommitMessage = "commitMessage"
const fileRepositoryURL = "repositoryUrl"

const folderGithub = "github"
const fileOwner = "owner"
const fileRepo = "repository"

const folderCustom = "custom"

//Do we still need this?
/*

// SetArtifactVersion stores the artifact's version in the pipeline environment
func SetArtifactVersion(path, version string) error {
	paramPath := filepath.Join(path, fileArtifactVersion)
	return writeToDisk(paramPath, []byte(version))
}

// GetArtifactVersion reads the artifact's version from the pipeline environment
func GetArtifactVersion(path string) string {
	paramPath := filepath.Join(path, fileArtifactVersion)
	return readFromDisk(paramPath)
}

// SetGitCommitID stores the git commit id in the pipeline environment
func SetGitCommitID(path, owner string) error {
	return setDeepParameter(path, folderGit, fileCommitID, owner)
}

// GetGitCommitID reads the git commit id from the pipeline environment
func GetGitCommitID(path string) string {
	return getDeepParameter(path, folderGit, fileCommitID)
}

// SetGitBranch stores the git branch in the pipeline environment
func SetGitBranch(path, owner string) error {
	return setDeepParameter(path, folderGit, fileBranch, owner)
}

// GetGitBranch reads the git branch from the pipeline environment
func GetGitBranch(path string) string {
	return getDeepParameter(path, folderGit, fileBranch)
}

// SetGitCommitMessage stores the git commit message in the pipeline environment
func SetGitCommitMessage(path, owner string) error {
	return setDeepParameter(path, folderGit, fileCommitMessage, owner)
}

// GetGitCommitMessage reads the git commit message from the pipeline environment
func GetGitCommitMessage(path string) string {
	return getDeepParameter(path, folderGit, fileCommitMessage)
}

// SetGithubOwner stores the github owner in the pipeline environment
func SetGithubOwner(path, owner string) error {
	return setDeepParameter(path, folderGithub, fileOwner, owner)
}

// GetGithubOwner reads the github owner from the pipeline environment
func GetGithubOwner(path string) string {
	return getDeepParameter(path, folderGithub, fileOwner)
}

// SetGithubRepository stores the github repository in the pipeline environment
func SetGithubRepository(path, owner string) error {
	paramPath := filepath.Join(path, folderGithub, fileRepo)
	return writeToDisk(paramPath, []byte(owner))
}

// GetGithubRepository reads the github repository from the pipeline environment
func GetGithubRepository(path string) string {
	paramPath := filepath.Join(path, folderGithub, fileRepo)
	return readFromDisk(paramPath)
}

// SetCustomParameter stores a custom parameter in the pipeline environment
func SetCustomParameter(path, param, value string) error {
	return setDeepParameter(path, folderCustom, param, value)
}

// GetCustomParameter reads a custom parameter from the pipeline environment
func GetCustomParameter(path, param string) string {
	return getDeepParameter(path, folderCustom, param)
}

func setDeepParameter(path, folder, param, value string) error {
	paramPath := filepath.Join(path, folder, param)
	return writeToDisk(paramPath, []byte(value))
}

func getDeepParameter(path, folder, param string) string {
	paramPath := filepath.Join(path, folder, param)
	return readFromDisk(paramPath)
}

*/

// SetParameter sets any parameter in the pipeline environment or another environment stored in the file system
func SetParameter(path, name, value string) error {
	paramPath := filepath.Join(path, name)
	return writeToDisk(paramPath, []byte(value))
}

// GetParameter reads any parameter from the pipeline environment or another environment stored in the file system
func GetParameter(path, name string) string {
	paramPath := filepath.Join(path, name)
	return readFromDisk(paramPath)
}

func writeToDisk(filename string, data []byte) error {

	if _, err := os.Stat(filepath.Dir(filename)); os.IsNotExist(err) {
		os.MkdirAll(filepath.Dir(filename), 0700)
	}

	//ToDo: make sure to not overwrite file but rather add another file? Create error if already existing?
	return ioutil.WriteFile(filename, data, 0700)
}

func readFromDisk(filename string) string {
	//ToDo: if multiple files exist, read from latest file
	v, err := ioutil.ReadFile(filename)
	val := string(v)
	if err != nil {
		val = ""
	}
	return val
}
