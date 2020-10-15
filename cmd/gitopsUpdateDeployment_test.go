package cmd

import (
	"errors"
	gitUtil "github.com/SAP/jenkins-library/pkg/git"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
	"io"
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

	fileUtilities = filesMockErrorTempDirCreation{}

	var c gitopsExecRunner
	configuration = &commonOptions

	err := runGitopsUpdateDeployment(configuration, c)
	assert.Equal(t, errors.New("error appeared"), err)
}

func TestErrorGitPlainClone(t *testing.T) {
	test = t

	defer func() {
		gitUtilities = gitUtil.TheGitUtils{}
	}()

	gitUtilities = gitUtilsMockErrorClone{}

	var c gitopsExecRunner
	configuration = &commonOptions

	err := runGitopsUpdateDeployment(configuration, c)
	assert.Equal(t, errors.New("error on clone"), err)
}

func TestErrorOnInvalidURL(t *testing.T) {
	test = t

	defer func() {
		gitUtilities = gitUtil.TheGitUtils{}
	}()

	gitUtilities = validGitUtilsMock{}

	var c gitopsExecRunner
	configuration = &invalidURLOptions

	err := runGitopsUpdateDeployment(configuration, c)
	assert.Equal(t, errors.New("invalid registry url"), err)
}

func TestBuildRegistryPlusImage(t *testing.T) {
	test = t
	registryImage, err := buildRegistryPlusImage(&commonOptions)
	assert.Nil(t, err)
	assert.Equal(t, "myregistry.com/myFancyContainer:1337", registryImage)
}

func TestBuildRegistryPlusImageWithoutRegistry(t *testing.T) {
	test = t
	registryImage, err := buildRegistryPlusImage(&commonOptionsNoRegistry)
	assert.Nil(t, err)
	assert.Equal(t, "myFancyContainer:1337", registryImage)
}

func TestRunGitopsUpdateDeployment(t *testing.T) {
	test = t
	defer func() {
		gitUtilities = gitUtil.TheGitUtils{}
	}()

	gitUtilities = validGitUtilsMock{}

	var c gitopsExecRunner = &gitOpsExecRunnerMock{}

	configuration = &commonOptions

	err := runGitopsUpdateDeployment(configuration, c)
	assert.NoError(t, err)
}

type gitOpsExecRunnerMock struct {
	out io.Writer
}

func (e *gitOpsExecRunnerMock) Stdout(out io.Writer) {
	e.out = out
}

func (gitOpsExecRunnerMock) Stderr(err io.Writer) {
	panic("implement me")
}

func (e *gitOpsExecRunnerMock) RunExecutable(executable string, params ...string) error {
	assert.Equal(test, "kubectl", executable)
	assert.Equal(test, "patch", params[0])
	assert.Equal(test, "--local", params[1])
	assert.Equal(test, "--output=yaml", params[2])
	assert.Equal(test, "--patch={\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"myContainer\",\"image\":\"myregistry.com/myFancyContainer:1337\"}]}}}}", params[3])
	assert.True(test, strings.Contains(params[4], filepath.Join("dir1/dir2/depl.yaml")))
	_, err := e.out.Write([]byte(expectedYaml))
	assert.NoError(test, err)
	return nil
}

type filesMockErrorTempDirCreation struct{}

func (filesMockErrorTempDirCreation) TempDir(dir, pattern string) (name string, err error) {
	return "", errors.New("error appeared")
}

func (filesMockErrorTempDirCreation) RemoveAll(path string) error {
	panic("implement me")
}

type gitUtilsMockErrorClone struct{}

func (gitUtilsMockErrorClone) CommitSingleFile(filePath, commitMessage string, worktree gitUtil.UtilsWorkTree) (plumbing.Hash, error) {
	panic("implement me")
}

func (gitUtilsMockErrorClone) PushChangesToRepository(username, password string, repository gitUtil.UtilsRepository) error {
	panic("implement me")
}

func (gitUtilsMockErrorClone) PlainClone(username, password, serverUrl, directory string) (gitUtil.UtilsRepository, error) {
	return nil, errors.New("error on clone")
}

func (gitUtilsMockErrorClone) ChangeBranch(branchName string, worktree gitUtil.UtilsWorkTree) error {
	panic("implement me")
}

func (gitUtilsMockErrorClone) GetWorktree(repository gitUtil.UtilsRepository) (gitUtil.UtilsWorkTree, error) {
	panic("implement me")
}

type validGitUtilsMock struct{}

func (validGitUtilsMock) GetWorktree(repository gitUtil.UtilsRepository) (gitUtil.UtilsWorkTree, error) {
	return nil, nil
}

func (validGitUtilsMock) ChangeBranch(branchName string, worktree gitUtil.UtilsWorkTree) error {
	assert.Equal(test, configuration.BranchName, branchName)
	return nil
}

func (validGitUtilsMock) CommitSingleFile(filePath, commitMessage string, worktree gitUtil.UtilsWorkTree) (plumbing.Hash, error) {
	matches, _ := piperutils.Files{}.Glob("*/dir1/dir2/depl.yaml")
	fileRead, _ := piperutils.Files{}.FileRead(matches[0])
	assert.Equal(test, expectedYaml, string(fileRead))
	return [20]byte{123}, nil
}

func (validGitUtilsMock) PushChangesToRepository(username, password string, repository gitUtil.UtilsRepository) error {
	return nil
}

func (validGitUtilsMock) PlainClone(username, password, serverUrl, directory string) (gitUtil.UtilsRepository, error) {
	filePath := filepath.Join(directory, "dir1/dir2/depl.yaml")
	err2 := piperutils.Files{}.MkdirAll(filepath.Join(directory, "dir1/dir2"), 0755)
	if err2 != nil {
		return nil, err2
	}
	err := piperutils.Files{}.FileWrite(filePath, []byte(existingYaml), 0755)
	if err != nil {
		return nil, err
	}
	return &repositoryMock{}, nil
}

type repositoryMock struct{}

func (repositoryMock) Worktree() (*git.Worktree, error) {
	return nil, nil
}

func (repositoryMock) Push(o *git.PushOptions) error {
	panic("implement me")
}

var existingYaml = "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: myFancyApp\n  labels:\n    tier: application\nspec:\n  replicas: 4\n  selector:\n    matchLabels:\n      run: myContainer\n  template:\n    metadata:\n      labels:\n        run: myContainer\n    spec:\n      containers:\n      - image: myregistry.com/myFancyContainer:1336\n        name: myContainer"
var expectedYaml = "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: myFancyApp\n  labels:\n    tier: application\nspec:\n  replicas: 4\n  selector:\n    matchLabels:\n      run: myContainer\n  template:\n    metadata:\n      labels:\n        run: myContainer\n    spec:\n      containers:\n      - image: myregistry.com/myFancyContainer:1337\n        name: myContainer"
