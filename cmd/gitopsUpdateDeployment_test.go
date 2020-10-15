package cmd

import (
	"errors"
	gitUtil "github.com/SAP/jenkins-library/pkg/git"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var commonOptions = gitopsUpdateDeploymentOptions{
	BranchName:           "main",
	CommitMessage:        "This is the commit message",
	ServerURL:            "https://github.com",
	Username:             "admin3",
	Password:             "validAccessToken",
	FilePath:             "dir1/dir2/depl.yaml",
	ContainerName:        "myContainer",
	ContainerRegistryURL: "https://myregistry.com/registry/containers",
	ContainerImage:       "myFancyContainer:1337",
}
var commonOptionsNoRegistry = gitopsUpdateDeploymentOptions{
	BranchName:           "main",
	CommitMessage:        "This is the commit message",
	ServerURL:            "https://github.com",
	Username:             "admin3",
	Password:             "validAccessToken",
	FilePath:             "dir1/dir2/depl.yaml",
	ContainerName:        "myContainer",
	ContainerRegistryURL: "",
	ContainerImage:       "myFancyContainer:1337",
}
var invalidURLOptions = gitopsUpdateDeploymentOptions{
	BranchName:           "main",
	CommitMessage:        "This is the commit message",
	ServerURL:            "https://github.com",
	Username:             "admin3",
	Password:             "validAccessToken",
	FilePath:             "dir1/dir2/depl.yaml",
	ContainerName:        "myContainer",
	ContainerRegistryURL: "//myregistry.com/registry/containers",
	ContainerImage:       "myFancyContainer:1337",
}

var test *testing.T
var configuration *gitopsUpdateDeploymentOptions

func TestErrorOnTempDir(t *testing.T) {
	test = t

	defer func() {
		fileUtilities = piperutils.Files{}
	}()

	fileUtilities = FilesMockErrorTempDirCreation{}

	var c GitopsExecRunner
	configuration = &commonOptions

	err := runGitopsUpdateDeployment(configuration, c)
	assert.Equal(t, errors.New("error appeared"), err)
}

func TestErrorGitPlainClone(t *testing.T) {
	test = t

	defer func() {
		gitUtilities = gitUtil.TheGitUtils{}
	}()

	gitUtilities = GitUtilsMockErrorClone{}

	var c GitopsExecRunner
	configuration = &commonOptions

	err := runGitopsUpdateDeployment(configuration, c)
	assert.Equal(t, errors.New("error on clone"), err)
}

func TestErrorOnInvalidURL(t *testing.T) {
	test = t

	defer func() {
		gitUtilities = gitUtil.TheGitUtils{}
	}()

	gitUtilities = ValidGitUtilsMock{}

	var c GitopsExecRunner
	configuration = &invalidURLOptions

	err := runGitopsUpdateDeployment(configuration, c)
	assert.Equal(t, errors.New("invalid registry url"), err)
}

func TestBuildRegistryPlusImage(t *testing.T) {
	test = t
	registryImage, err := BuildRegistryPlusImage(&commonOptions)
	assert.Nil(t, err)
	assert.Equal(t, "myregistry.com/myFancyContainer:1337", registryImage)
}

func TestBuildRegistryPlusImageWithoutRegistry(t *testing.T) {
	test = t
	registryImage, err := BuildRegistryPlusImage(&commonOptionsNoRegistry)
	assert.Nil(t, err)
	assert.Equal(t, "myFancyContainer:1337", registryImage)
}

func TestRunGitopsUpdateDeployment(t *testing.T) {
	test = t
	defer func() {
		gitUtilities = gitUtil.TheGitUtils{}
	}()

	gitUtilities = ValidGitUtilsMock{}

	var c GitopsExecRunner = &ExecRunnerMock{}

	configuration = &commonOptions

	err := runGitopsUpdateDeployment(configuration, c)
	assert.NoError(t, err)
}

type ExecRunnerMock struct {
	out io.Writer
}

func (e *ExecRunnerMock) Stdout(out io.Writer) {
	e.out = out
}

func (ExecRunnerMock) Stderr(err io.Writer) {
	panic("implement me")
}

func (e *ExecRunnerMock) RunExecutable(executable string, params ...string) error {
	assert.Equal(test, "kubectl", executable)
	assert.Equal(test, "patch", params[0])
	assert.Equal(test, "--local", params[1])
	assert.Equal(test, "--output=yaml", params[2])
	assert.Equal(test, "--patch={\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"myContainer\",\"image\":\"myregistry.com/myFancyContainer:1337\"}]}}}}", params[3])
	assert.True(test, strings.Contains(params[4], "dir1\\dir2\\depl.yaml"))
	_, _ = e.out.Write([]byte(expectedYaml))
	return nil
}

type FilesMockErrorTempDirCreation struct{}

func (c FilesMockErrorTempDirCreation) Getwd() (string, error) {
	panic("implement me")
}

func (FilesMockErrorTempDirCreation) Abs(path string) (string, error) {
	panic("implement me")
}

func (FilesMockErrorTempDirCreation) FileExists(filename string) (bool, error) {
	panic("implement me")
}

func (FilesMockErrorTempDirCreation) Copy(src, dest string) (int64, error) {
	panic("implement me")
}

func (FilesMockErrorTempDirCreation) FileRead(path string) ([]byte, error) {
	panic("implement me")
}

func (FilesMockErrorTempDirCreation) FileWrite(path string, content []byte, perm os.FileMode) error {
	panic("implement me")
}

func (FilesMockErrorTempDirCreation) MkdirAll(path string, perm os.FileMode) error {
	panic("implement me")
}

func (FilesMockErrorTempDirCreation) Chmod(path string, mode os.FileMode) error {
	panic("implement me")
}

func (FilesMockErrorTempDirCreation) Glob(pattern string) (matches []string, err error) {
	panic("implement me")
}

func (FilesMockErrorTempDirCreation) TempDir(dir, pattern string) (name string, err error) {
	return "", errors.New("error appeared")
}

func (FilesMockErrorTempDirCreation) RemoveAll(path string) error {
	panic("implement me")
}

type GitUtilsMockErrorClone struct{}

func (c GitUtilsMockErrorClone) CommitSingleFile(filePath, commitMessage string, repository *git.Repository) (plumbing.Hash, error) {
	panic("implement me")
}

func (c GitUtilsMockErrorClone) ChangeBranch(branchName string, repository *git.Repository) error {
	panic("implement me")
}

func (GitUtilsMockErrorClone) PushChangesToRepository(username, password string, repository *git.Repository) error {
	panic("implement me")
}

func (GitUtilsMockErrorClone) PlainClone(username, password, serverUrl, directory string) (*git.Repository, error) {
	return nil, errors.New("error on clone")
}

type ValidGitUtilsMock struct{}

func (m ValidGitUtilsMock) ChangeBranch(branchName string, repository *git.Repository) error {
	assert.Equal(test, configuration.BranchName, branchName)
	return nil
}

func (ValidGitUtilsMock) CommitSingleFile(filePath, commitMessage string, repository *git.Repository) (plumbing.Hash, error) {
	matches, _ := piperutils.Files{}.Glob("*/dir1/dir2/depl.yaml")
	fileRead, _ := piperutils.Files{}.FileRead(matches[0])
	assert.Equal(test, expectedYaml, string(fileRead))
	return [20]byte{123}, nil
}

func (ValidGitUtilsMock) PushChangesToRepository(username, password string, repository *git.Repository) error {
	return nil
}

func (ValidGitUtilsMock) PlainClone(username, password, serverUrl, directory string) (*git.Repository, error) {
	filePath := filepath.Join(directory, "dir1/dir2/depl.yaml")
	err2 := piperutils.Files{}.MkdirAll(filepath.Join(directory, "dir1/dir2"), 0755)
	if err2 != nil {
		return nil, err2
	}
	err := piperutils.Files{}.FileWrite(filePath, []byte(existingYaml), 0755)
	if err != nil {
		return nil, err
	}
	return &git.Repository{}, nil
}

var existingYaml = "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: myFancyApp\n  labels:\n    tier: application\nspec:\n  replicas: 4\n  selector:\n    matchLabels:\n      run: myContainer\n  template:\n    metadata:\n      labels:\n        run: myContainer\n    spec:\n      containers:\n      - image: myregistry.com/myFancyContainer:1336\n        name: myContainer"
var expectedYaml = "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: myFancyApp\n  labels:\n    tier: application\nspec:\n  replicas: 4\n  selector:\n    matchLabels:\n      run: myContainer\n  template:\n    metadata:\n      labels:\n        run: myContainer\n    spec:\n      containers:\n      - image: myregistry.com/myFancyContainer:1337\n        name: myContainer"
