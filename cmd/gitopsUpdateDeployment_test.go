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

func TestErrorOnTempDir(t *testing.T) {
	var c gitopsExecRunner
	var configuration = &gitopsUpdateDeploymentOptions{
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

	err := runGitopsUpdateDeployment(configuration, c, gitUtil.TheGitUtils{}, filesMockErrorTempDirCreation{})
	assert.Equal(t, errors.New("error appeared"), err)
}

func TestErrorGitPlainClone(t *testing.T) {
	var c gitopsExecRunner
	var configuration = &gitopsUpdateDeploymentOptions{
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

	err := runGitopsUpdateDeployment(configuration, c, gitUtilsMockErrorClone{}, piperutils.Files{})
	assert.Equal(t, errors.New("error on clone"), err)
}

func TestErrorOnInvalidURL(t *testing.T) {
	var configuration = &gitopsUpdateDeploymentOptions{
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

	gitUtilsMock := validGitUtilsMock{
		configuration: configuration,
		test:          t,
	}

	var c gitopsExecRunner

	err := runGitopsUpdateDeployment(configuration, c, gitUtilsMock, piperutils.Files{})
	assert.Equal(t, errors.New("invalid registry url"), err)
}

func TestBuildRegistryPlusImage(t *testing.T) {
	registryImage, err := buildRegistryPlusImage(&gitopsUpdateDeploymentOptions{
		BranchName:           "main",
		CommitMessage:        "This is the commit message",
		ServerURL:            "https://github.com",
		Username:             "admin3",
		Password:             "validAccessToken",
		FilePath:             "dir1/dir2/depl.yaml",
		ContainerName:        "myContainer",
		ContainerRegistryURL: "https://myregistry.com/registry/containers",
		ContainerImage:       "myFancyContainer:1337",
	})
	assert.Nil(t, err)
	assert.Equal(t, "myregistry.com/myFancyContainer:1337", registryImage)
}

func TestBuildRegistryPlusImageWithoutRegistry(t *testing.T) {
	registryImage, err := buildRegistryPlusImage(&gitopsUpdateDeploymentOptions{
		BranchName:           "main",
		CommitMessage:        "This is the commit message",
		ServerURL:            "https://github.com",
		Username:             "admin3",
		Password:             "validAccessToken",
		FilePath:             "dir1/dir2/depl.yaml",
		ContainerName:        "myContainer",
		ContainerRegistryURL: "",
		ContainerImage:       "myFancyContainer:1337",
	})
	assert.Nil(t, err)
	assert.Equal(t, "myFancyContainer:1337", registryImage)
}

func TestRunGitopsUpdateDeployment(t *testing.T) {
	var configuration = &gitopsUpdateDeploymentOptions{
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

	gitUtilsMock := validGitUtilsMock{
		configuration: configuration,
		test:          t,
	}

	var c gitopsExecRunner = &gitOpsExecRunnerMock{
		test: t,
	}

	err := runGitopsUpdateDeployment(configuration, c, gitUtilsMock, piperutils.Files{})
	assert.NoError(t, err)
}

type gitOpsExecRunnerMock struct {
	out  io.Writer
	test *testing.T
}

func (e *gitOpsExecRunnerMock) Stdout(out io.Writer) {
	e.out = out
}

func (gitOpsExecRunnerMock) Stderr(err io.Writer) {
	panic("implement me")
}

func (e *gitOpsExecRunnerMock) RunExecutable(executable string, params ...string) error {
	assert.Equal(e.test, "kubectl", executable)
	assert.Equal(e.test, "patch", params[0])
	assert.Equal(e.test, "--local", params[1])
	assert.Equal(e.test, "--output=yaml", params[2])
	assert.Equal(e.test, "--patch={\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"myContainer\",\"image\":\"myregistry.com/myFancyContainer:1337\"}]}}}}", params[3])
	assert.True(e.test, strings.Contains(params[4], filepath.Join("dir1/dir2/depl.yaml")))
	_, err := e.out.Write([]byte(expectedYaml))
	assert.NoError(e.test, err)
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

type validGitUtilsMock struct {
	configuration *gitopsUpdateDeploymentOptions
	test          *testing.T
}

func (validGitUtilsMock) GetWorktree(repository gitUtil.UtilsRepository) (gitUtil.UtilsWorkTree, error) {
	return nil, nil
}

func (v validGitUtilsMock) ChangeBranch(branchName string, worktree gitUtil.UtilsWorkTree) error {
	assert.Equal(v.test, v.configuration.BranchName, branchName)
	return nil
}

func (v validGitUtilsMock) CommitSingleFile(filePath, commitMessage string, worktree gitUtil.UtilsWorkTree) (plumbing.Hash, error) {
	matches, _ := piperutils.Files{}.Glob("*/dir1/dir2/depl.yaml")
	fileRead, _ := piperutils.Files{}.FileRead(matches[0])
	assert.Equal(v.test, expectedYaml, string(fileRead))
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
