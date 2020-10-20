package cmd

import (
	"errors"
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

func TestBuildRegistryPlusImage(t *testing.T) {
	t.Parallel()
	t.Run("build full image", func(t *testing.T) {
		registryImage, err := buildRegistryPlusImage(&gitopsUpdateDeploymentOptions{
			ContainerRegistryURL: "https://myregistry.com/registry/containers",
			ContainerImage:       "myFancyContainer:1337",
		})
		assert.NoError(t, err)
		assert.Equal(t, "myregistry.com/myFancyContainer:1337", registryImage)
	})

	t.Run("without registry", func(t *testing.T) {
		registryImage, err := buildRegistryPlusImage(&gitopsUpdateDeploymentOptions{
			ContainerRegistryURL: "",
			ContainerImage:       "myFancyContainer:1337",
		})
		assert.NoError(t, err)
		assert.Equal(t, "myFancyContainer:1337", registryImage)
	})
	t.Run("without faulty URL", func(t *testing.T) {
		_, err := buildRegistryPlusImage(&gitopsUpdateDeploymentOptions{
			ContainerRegistryURL: "//myregistry.com/registry/containers",
			ContainerImage:       "myFancyContainer:1337",
		})
		assert.Error(t, err)
		assert.EqualError(t, err, "registry URL could not be extracted: invalid registry url")
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
		var c gitopsUpdateDeploymentExecRunner = &runnerMock

		err := runGitopsUpdateDeployment(configuration, c, gitUtilsMock, piperutils.Files{})
		assert.NoError(t, err)
		assert.Equal(t, configuration.BranchName, gitUtilsMock.changedBranch)
		assert.Equal(t, expectedYaml, gitUtilsMock.savedFile)
		assert.Equal(t, "kubectl", runnerMock.executable)
		assert.Equal(t, "patch", runnerMock.kubectlParams[0])
		assert.Equal(t, "--local", runnerMock.kubectlParams[1])
		assert.Equal(t, "--output=yaml", runnerMock.kubectlParams[2])
		assert.Equal(t, `--patch={"spec":{"template":{"spec":{"containers":[{"name":"myContainer","image":"myregistry.com/myFancyContainer:1337"}]}}}}`, runnerMock.kubectlParams[3])
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

		err := runGitopsUpdateDeployment(configuration, nil, gitUtilsMock, piperutils.Files{})
		assert.EqualError(t, err, "failed to apply kubectl command: registry URL could not be extracted: invalid registry url")
	})

	t.Run("error on plane clone", func(t *testing.T) {
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

		err := runGitopsUpdateDeployment(configuration, nil, &gitUtilsMockErrorClone{}, piperutils.Files{})
		assert.EqualError(t, err, "failed to plain clone repository: error on clone")
	})

	t.Run("error on temp dir creation", func(t *testing.T) {
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

		err := runGitopsUpdateDeployment(configuration, nil, &gitopsUpdateDeploymentGitUtils{}, filesMockErrorTempDirCreation{})
		assert.EqualError(t, err, "failed to create temporary directory: error appeared")
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

func (gitOpsExecRunnerMock) Stderr(io.Writer) {
	panic("implement me")
}

func (e *gitOpsExecRunnerMock) RunExecutable(executable string, params ...string) error {
	e.executable = executable
	e.kubectlParams = params
	_, err := e.out.Write([]byte(expectedYaml))
	return err
}

type filesMockErrorTempDirCreation struct{}

func (c filesMockErrorTempDirCreation) FileWrite(string, []byte, os.FileMode) error {
	panic("implement me")
}

func (filesMockErrorTempDirCreation) TempDir(string, string) (name string, err error) {
	return "", errors.New("error appeared")
}

func (filesMockErrorTempDirCreation) RemoveAll(string) error {
	panic("implement me")
}

type gitUtilsMockErrorClone struct{}

func (gitUtilsMockErrorClone) CommitSingleFile(string, string) (plumbing.Hash, error) {
	panic("implement me")
}

func (gitUtilsMockErrorClone) PushChangesToRepository(string, string) error {
	panic("implement me")
}

func (gitUtilsMockErrorClone) PlainClone(string, string, string, string) error {
	return errors.New("error on clone")
}

func (gitUtilsMockErrorClone) ChangeBranch(string) error {
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

func (v *validGitUtilsMock) CommitSingleFile(string, string) (plumbing.Hash, error) {
	matches, _ := piperutils.Files{}.Glob("*/dir1/dir2/depl.yaml")
	fileRead, _ := piperutils.Files{}.FileRead(matches[0])
	v.savedFile = string(fileRead)
	return [20]byte{123}, nil
}

func (validGitUtilsMock) PushChangesToRepository(string, string) error {
	return nil
}

func (validGitUtilsMock) PlainClone(_, _, _, directory string) error {
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
