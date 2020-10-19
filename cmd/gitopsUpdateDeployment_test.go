package cmd

import (
	"errors"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
	"io"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildRegistryPlusImage(t *testing.T) {
	t.Parallel()
	t.Run("build full image", func(t *testing.T) {
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
		assert.NoError(t, err)
		assert.Equal(t, "myregistry.com/myFancyContainer:1337", registryImage)
	})

	t.Run("without registry", func(t *testing.T) {
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
		assert.NoError(t, err)
		assert.Equal(t, "myFancyContainer:1337", registryImage)
	})
}

func TestRunGitopsUpdateDeployment(t *testing.T) {
	t.Parallel()
	t.Run("successful run", func(t *testing.T) {
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

		gitUtilsMock := &validGitUtilsMock{}

		runnerMock := gitOpsExecRunnerMock{}
		var c gitopsExecRunner = &runnerMock

		err := runGitopsUpdateDeployment(configuration, c, gitUtilsMock, piperutils.Files{})
		assert.NoError(t, err)
		assert.Equal(t, configuration.BranchName, gitUtilsMock.changedBranch)
		assert.Equal(t, expectedYaml, gitUtilsMock.savedFile)
		assert.Equal(t, "kubectl", runnerMock.executable)
		assert.Equal(t, "patch", runnerMock.kubectlParams[0])
		assert.Equal(t, "--local", runnerMock.kubectlParams[1])
		assert.Equal(t, "--output=yaml", runnerMock.kubectlParams[2])
		assert.Equal(t, "--patch={\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"myContainer\",\"image\":\"myregistry.com/myFancyContainer:1337\"}]}}}}", runnerMock.kubectlParams[3])
		assert.True(t, strings.Contains(runnerMock.kubectlParams[4], filepath.Join("dir1/dir2/depl.yaml")))
	})

	t.Run("invalid URL", func(t *testing.T) {
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

		gitUtilsMock := &validGitUtilsMock{}

		var c gitopsExecRunner

		err := runGitopsUpdateDeployment(configuration, c, gitUtilsMock, piperutils.Files{})
		assert.EqualError(t, err, "invalid registry url")
	})

	t.Run("error on plane clone", func(t *testing.T) {
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

		err := runGitopsUpdateDeployment(configuration, c, &gitUtilsMockErrorClone{}, piperutils.Files{})
		assert.EqualError(t, err, "error on clone")
	})

	t.Run("error on temp dir creation", func(t *testing.T) {
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

		err := runGitopsUpdateDeployment(configuration, c, &gitUtilsRuntime{}, filesMockErrorTempDirCreation{})
		assert.EqualError(t, err, "error appeared")
	})
}

type gitOpsExecRunnerMock struct {
	out           io.Writer
	kubectlParams []string
	executable    string
}

func (e *gitOpsExecRunnerMock) Stdout(out io.Writer) {
	e.out = out
}

func (gitOpsExecRunnerMock) Stderr(err io.Writer) {
	panic("implement me")
}

func (e *gitOpsExecRunnerMock) RunExecutable(executable string, params ...string) error {
	e.executable = executable
	e.kubectlParams = params
	_, err := e.out.Write([]byte(expectedYaml))
	return err
}

type filesMockErrorTempDirCreation struct{}

func (filesMockErrorTempDirCreation) TempDir(dir, pattern string) (name string, err error) {
	return "", errors.New("error appeared")
}

func (filesMockErrorTempDirCreation) RemoveAll(path string) error {
	panic("implement me")
}

type gitUtilsMockErrorClone struct{}

func (gitUtilsMockErrorClone) CommitSingleFile(filePath, commitMessage string) (plumbing.Hash, error) {
	panic("implement me")
}

func (gitUtilsMockErrorClone) PushChangesToRepository(username, password string) error {
	panic("implement me")
}

func (gitUtilsMockErrorClone) PlainClone(username, password, serverUrl, directory string) error {
	return errors.New("error on clone")
}

func (gitUtilsMockErrorClone) ChangeBranch(branchName string) error {
	panic("implement me")
}

func (gitUtilsMockErrorClone) GetWorktree() (*git.Worktree, error) {
	panic("implement me")
}

type validGitUtilsMock struct {
	savedFile     string
	changedBranch string
}

func (validGitUtilsMock) GetWorktree() (*git.Worktree, error) {
	return nil, nil
}

func (v *validGitUtilsMock) ChangeBranch(branchName string) error {
	v.changedBranch = branchName
	return nil
}

func (v *validGitUtilsMock) CommitSingleFile(filePath, commitMessage string) (plumbing.Hash, error) {
	matches, _ := piperutils.Files{}.Glob("*/dir1/dir2/depl.yaml")
	fileRead, _ := piperutils.Files{}.FileRead(matches[0])
	v.savedFile = string(fileRead)
	return [20]byte{123}, nil
}

func (validGitUtilsMock) PushChangesToRepository(username, password string) error {
	return nil
}

func (validGitUtilsMock) PlainClone(username, password, serverUrl, directory string) error {
	filePath := filepath.Join(directory, "dir1/dir2/depl.yaml")
	err := piperutils.Files{}.MkdirAll(filepath.Join(directory, "dir1/dir2"), 0755)
	if err != nil {
		return err
	}
	err = piperutils.Files{}.FileWrite(filePath, []byte(existingYaml), 0755)
	if err != nil {
		return err
	}
	return nil
}

var existingYaml = "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: myFancyApp\n  labels:\n    tier: application\nspec:\n  replicas: 4\n  selector:\n    matchLabels:\n      run: myContainer\n  template:\n    metadata:\n      labels:\n        run: myContainer\n    spec:\n      containers:\n      - image: myregistry.com/myFancyContainer:1336\n        name: myContainer"
var expectedYaml = "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: myFancyApp\n  labels:\n    tier: application\nspec:\n  replicas: 4\n  selector:\n    matchLabels:\n      run: myContainer\n  template:\n    metadata:\n      labels:\n        run: myContainer\n    spec:\n      containers:\n      - image: myregistry.com/myFancyContainer:1337\n        name: myContainer"
