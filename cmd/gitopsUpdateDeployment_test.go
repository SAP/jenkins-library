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
			ContainerRegistryURL:  "https://myregistry.com/registry/containers",
			ContainerImageNameTag: "myFancyContainer:1337",
		})
		assert.NoError(t, err)
		assert.Equal(t, "myregistry.com/myFancyContainer:1337", registryImage)
	})

	t.Run("without registry", func(t *testing.T) {
		registryImage, err := buildRegistryPlusImage(&gitopsUpdateDeploymentOptions{
			ContainerRegistryURL:  "",
			ContainerImageNameTag: "myFancyContainer:1337",
		})
		assert.NoError(t, err)
		assert.Equal(t, "myFancyContainer:1337", registryImage)
	})
	t.Run("without faulty URL", func(t *testing.T) {
		_, err := buildRegistryPlusImage(&gitopsUpdateDeploymentOptions{
			ContainerRegistryURL:  "//myregistry.com/registry/containers",
			ContainerImageNameTag: "myFancyContainer:1337",
		})
		assert.Error(t, err)
		assert.EqualError(t, err, "registry URL could not be extracted: invalid registry url")
	})
}

func TestBuildRegistryPlusImageWithoutTag(t *testing.T) {
	t.Parallel()
	t.Run("build full image", func(t *testing.T) {
		registryImage, tag, err := buildRegistryPlusImageAndTagSeparately(&gitopsUpdateDeploymentOptions{
			ContainerRegistryURL:  "https://myregistry.com/registry/containers",
			ContainerImageNameTag: "myFancyContainer:1337",
		})
		assert.NoError(t, err)
		assert.Equal(t, "myregistry.com/myFancyContainer", registryImage)
		assert.Equal(t, "1337", tag)
	})

	t.Run("without registry", func(t *testing.T) {
		registryImage, tag, err := buildRegistryPlusImageAndTagSeparately(&gitopsUpdateDeploymentOptions{
			ContainerRegistryURL:  "",
			ContainerImageNameTag: "myFancyContainer:1337",
		})
		assert.NoError(t, err)
		assert.Equal(t, "myFancyContainer", registryImage)
		assert.Equal(t, "1337", tag)
	})
	t.Run("without faulty URL", func(t *testing.T) {
		_, _, err := buildRegistryPlusImageAndTagSeparately(&gitopsUpdateDeploymentOptions{
			ContainerRegistryURL:  "//myregistry.com/registry/containers",
			ContainerImageNameTag: "myFancyContainer:1337",
		})
		assert.Error(t, err)
		assert.EqualError(t, err, "registry URL could not be extracted: invalid registry url")
	})
}

func TestRunGitopsUpdateDeploymentWithKubectl(t *testing.T) {
	var validConfiguration = &gitopsUpdateDeploymentOptions{
		BranchName:            "main",
		CommitMessage:         "This is the commit message",
		ServerURL:             "https://github.com",
		Username:              "admin3",
		Password:              "validAccessToken",
		FilePath:              "dir1/dir2/depl.yaml",
		ContainerName:         "myContainer",
		ContainerRegistryURL:  "https://myregistry.com/registry/containers",
		ContainerImageNameTag: "myFancyContainer:1337",
		Tool:                  "kubectl",
	}

	t.Parallel()
	t.Run("successful run", func(t *testing.T) {
		gitUtilsMock := &gitUtilsMock{}
		runnerMock := &gitOpsExecRunnerMock{}

		err := runGitopsUpdateDeployment(validConfiguration, runnerMock, gitUtilsMock, &filesMock{})
		assert.NoError(t, err)
		assert.Equal(t, validConfiguration.BranchName, gitUtilsMock.changedBranch)
		assert.Equal(t, expectedYaml, gitUtilsMock.savedFile)
		assert.Equal(t, "This is the commit message", gitUtilsMock.commitMessage)
		assert.Equal(t, "kubectl", runnerMock.executable)
		assert.Equal(t, "patch", runnerMock.params[0])
		assert.Equal(t, "--local", runnerMock.params[1])
		assert.Equal(t, "--output=yaml", runnerMock.params[2])
		assert.Equal(t, `--patch={"spec":{"template":{"spec":{"containers":[{"name":"myContainer","image":"myregistry.com/myFancyContainer:1337"}]}}}}`, runnerMock.params[3])
		assert.True(t, strings.Contains(runnerMock.params[4], filepath.Join("dir1/dir2/depl.yaml")))
	})

	t.Run("default commit message", func(t *testing.T) {
		var configuration = *validConfiguration
		configuration.CommitMessage = ""

		gitUtilsMock := &gitUtilsMock{}
		runnerMock := &gitOpsExecRunnerMock{}

		err := runGitopsUpdateDeployment(&configuration, runnerMock, gitUtilsMock, &filesMock{})
		assert.NoError(t, err)
		assert.Equal(t, validConfiguration.BranchName, gitUtilsMock.changedBranch)
		assert.Equal(t, expectedYaml, gitUtilsMock.savedFile)
		assert.Equal(t, "Updated myregistry.com/myFancyContainer to version 1337", gitUtilsMock.commitMessage)
		assert.Equal(t, "kubectl", runnerMock.executable)
		assert.Equal(t, "patch", runnerMock.params[0])
		assert.Equal(t, "--local", runnerMock.params[1])
		assert.Equal(t, "--output=yaml", runnerMock.params[2])
		assert.Equal(t, `--patch={"spec":{"template":{"spec":{"containers":[{"name":"myContainer","image":"myregistry.com/myFancyContainer:1337"}]}}}}`, runnerMock.params[3])
		assert.True(t, strings.Contains(runnerMock.params[4], filepath.Join("dir1/dir2/depl.yaml")))
	})

	t.Run("ChartPath not used for kubectl", func(t *testing.T) {
		var configuration = *validConfiguration
		configuration.ChartPath = "chartPath"

		gitUtilsMock := &gitUtilsMock{}
		runnerMock := &gitOpsExecRunnerMock{}

		err := runGitopsUpdateDeployment(&configuration, runnerMock, gitUtilsMock, &filesMock{})
		assert.NoError(t, err)
		assert.Equal(t, configuration.BranchName, gitUtilsMock.changedBranch)
		assert.Equal(t, expectedYaml, gitUtilsMock.savedFile)
		assert.Equal(t, "kubectl", runnerMock.executable)
		assert.Equal(t, "patch", runnerMock.params[0])
		assert.Equal(t, "--local", runnerMock.params[1])
		assert.Equal(t, "--output=yaml", runnerMock.params[2])
		assert.Equal(t, `--patch={"spec":{"template":{"spec":{"containers":[{"name":"myContainer","image":"myregistry.com/myFancyContainer:1337"}]}}}}`, runnerMock.params[3])
		assert.True(t, strings.Contains(runnerMock.params[4], filepath.Join("dir1/dir2/depl.yaml")))
	})

	t.Run("HelmValues not used for kubectl", func(t *testing.T) {
		var configuration = *validConfiguration
		configuration.HelmValues = []string{"HelmValues"}

		gitUtilsMock := &gitUtilsMock{}
		runnerMock := &gitOpsExecRunnerMock{}

		err := runGitopsUpdateDeployment(&configuration, runnerMock, gitUtilsMock, &filesMock{})
		assert.NoError(t, err)
		assert.Equal(t, configuration.BranchName, gitUtilsMock.changedBranch)
		assert.Equal(t, expectedYaml, gitUtilsMock.savedFile)
		assert.Equal(t, "kubectl", runnerMock.executable)
		assert.Equal(t, "patch", runnerMock.params[0])
		assert.Equal(t, "--local", runnerMock.params[1])
		assert.Equal(t, "--output=yaml", runnerMock.params[2])
		assert.Equal(t, `--patch={"spec":{"template":{"spec":{"containers":[{"name":"myContainer","image":"myregistry.com/myFancyContainer:1337"}]}}}}`, runnerMock.params[3])
		assert.True(t, strings.Contains(runnerMock.params[4], filepath.Join("dir1/dir2/depl.yaml")))
	})

	t.Run("DeploymentName not used for kubectl", func(t *testing.T) {
		var configuration = *validConfiguration
		configuration.DeploymentName = "DeploymentName"

		gitUtilsMock := &gitUtilsMock{}
		runnerMock := &gitOpsExecRunnerMock{}

		err := runGitopsUpdateDeployment(&configuration, runnerMock, gitUtilsMock, &filesMock{})
		assert.NoError(t, err)
		assert.Equal(t, configuration.BranchName, gitUtilsMock.changedBranch)
		assert.Equal(t, expectedYaml, gitUtilsMock.savedFile)
		assert.Equal(t, "kubectl", runnerMock.executable)
		assert.Equal(t, "patch", runnerMock.params[0])
		assert.Equal(t, "--local", runnerMock.params[1])
		assert.Equal(t, "--output=yaml", runnerMock.params[2])
		assert.Equal(t, `--patch={"spec":{"template":{"spec":{"containers":[{"name":"myContainer","image":"myregistry.com/myFancyContainer:1337"}]}}}}`, runnerMock.params[3])
		assert.True(t, strings.Contains(runnerMock.params[4], filepath.Join("dir1/dir2/depl.yaml")))
	})

	t.Run("missing ContainerName", func(t *testing.T) {
		var configuration = *validConfiguration
		configuration.ContainerName = ""

		gitUtilsMock := &gitUtilsMock{}
		runnerMock := &gitOpsExecRunnerMock{}

		err := runGitopsUpdateDeployment(&configuration, runnerMock, gitUtilsMock, &filesMock{})
		assert.Error(t, err)
		assert.EqualError(t, err, "missing required fields for kubectl: the following parameters are necessary for kubectl: [containerName]")
	})

	t.Run("error on kubectl execution", func(t *testing.T) {
		runner := &gitOpsExecRunnerMock{failOnRunExecutable: true}

		err := runGitopsUpdateDeployment(validConfiguration, runner, &gitUtilsMock{}, &filesMock{})
		assert.Error(t, err)
		assert.EqualError(t, err, "error on kubectl execution: failed to apply kubectl command: failed to apply kubectl command: error happened")
	})

	t.Run("invalid URL", func(t *testing.T) {
		var configuration = *validConfiguration
		configuration.ContainerRegistryURL = "//myregistry.com/registry/containers"

		err := runGitopsUpdateDeployment(&configuration, &gitOpsExecRunnerMock{}, &gitUtilsMock{}, &filesMock{})
		assert.EqualError(t, err, "error on kubectl execution: failed to apply kubectl command: registry URL could not be extracted: invalid registry url")
	})

	t.Run("error on plain clone", func(t *testing.T) {
		gitUtils := &gitUtilsMock{failOnClone: true}

		err := runGitopsUpdateDeployment(validConfiguration, &gitOpsExecRunnerMock{}, gitUtils, &filesMock{})
		assert.EqualError(t, err, "repository could not get prepared: failed to plain clone repository: error on clone")
	})

	t.Run("error on change branch", func(t *testing.T) {
		gitUtils := &gitUtilsMock{failOnChangeBranch: true}

		err := runGitopsUpdateDeployment(validConfiguration, &gitOpsExecRunnerMock{}, gitUtils, &filesMock{})
		assert.EqualError(t, err, "repository could not get prepared: failed to change branch: error on change branch")
	})

	t.Run("error on commit changes", func(t *testing.T) {
		gitUtils := &gitUtilsMock{failOnCommit: true}

		err := runGitopsUpdateDeployment(validConfiguration, &gitOpsExecRunnerMock{}, gitUtils, &filesMock{})
		assert.EqualError(t, err, "failed to commit and push changes: committing changes failed: error on commit")
	})

	t.Run("error on push commits", func(t *testing.T) {
		gitUtils := &gitUtilsMock{failOnPush: true}

		err := runGitopsUpdateDeployment(validConfiguration, &gitOpsExecRunnerMock{}, gitUtils, &filesMock{})
		assert.EqualError(t, err, "failed to commit and push changes: pushing changes failed: error on push")
	})

	t.Run("error on temp dir creation", func(t *testing.T) {
		fileUtils := &filesMock{failOnCreation: true}

		err := runGitopsUpdateDeployment(validConfiguration, &gitOpsExecRunnerMock{}, &gitUtilsMock{}, fileUtils)
		assert.EqualError(t, err, "failed to create temporary directory: error appeared")
	})

	t.Run("error on file write", func(t *testing.T) {
		fileUtils := &filesMock{failOnWrite: true}

		err := runGitopsUpdateDeployment(validConfiguration, &gitOpsExecRunnerMock{}, &gitUtilsMock{}, fileUtils)
		assert.EqualError(t, err, "failed to write file: error appeared")
	})

	t.Run("error on temp dir deletion", func(t *testing.T) {
		fileUtils := &filesMock{failOnDeletion: true}

		err := runGitopsUpdateDeployment(validConfiguration, &gitOpsExecRunnerMock{}, &gitUtilsMock{}, fileUtils)
		assert.NoError(t, err)
		_ = piperutils.Files{}.RemoveAll(fileUtils.path)
	})
}

func TestRunGitopsUpdateDeploymentWithInvalid(t *testing.T) {
	t.Run("invalid deploy tool is not supported", func(t *testing.T) {
		var configuration = &gitopsUpdateDeploymentOptions{
			BranchName:            "main",
			CommitMessage:         "This is the commit message",
			ServerURL:             "https://github.com",
			Username:              "admin3",
			Password:              "validAccessToken",
			FilePath:              "dir1/dir2/depl.yaml",
			ContainerName:         "myContainer",
			ContainerRegistryURL:  "https://myregistry.com",
			ContainerImageNameTag: "registry/containers/myFancyContainer:1337",
			Tool:                  "invalid",
			ChartPath:             "./helm",
			DeploymentName:        "myFancyDeployment",
			HelmValues:            []string{"./helm/additionalValues.yaml"},
		}

		err := runGitopsUpdateDeployment(configuration, &gitOpsExecRunnerMock{}, &gitUtilsMock{}, &filesMock{})
		assert.Error(t, err)
		assert.EqualError(t, err, "tool invalid is not supported")
	})
}

func TestRunGitopsUpdateDeploymentWithHelm(t *testing.T) {
	var validConfiguration = &gitopsUpdateDeploymentOptions{
		BranchName:            "main",
		CommitMessage:         "This is the commit message",
		ServerURL:             "https://github.com",
		Username:              "admin3",
		Password:              "validAccessToken",
		FilePath:              "dir1/dir2/depl.yaml",
		ContainerRegistryURL:  "https://myregistry.com",
		ContainerImageNameTag: "registry/containers/myFancyContainer:1337",
		Tool:                  "helm",
		ChartPath:             "./helm",
		DeploymentName:        "myFancyDeployment",
		HelmValues:            []string{"./helm/additionalValues.yaml"},
	}

	t.Parallel()
	t.Run("successful run", func(t *testing.T) {
		gitUtilsMock := &gitUtilsMock{}
		runnerMock := &gitOpsExecRunnerMock{}

		err := runGitopsUpdateDeployment(validConfiguration, runnerMock, gitUtilsMock, &filesMock{})
		assert.NoError(t, err)
		assert.Equal(t, validConfiguration.BranchName, gitUtilsMock.changedBranch)
		assert.Equal(t, expectedYaml, gitUtilsMock.savedFile)
		assert.Equal(t, "This is the commit message", gitUtilsMock.commitMessage)
		assert.Equal(t, "helm", runnerMock.executable)
		assert.Equal(t, "template", runnerMock.params[0])
		assert.Equal(t, "myFancyDeployment", runnerMock.params[1])
		assert.Equal(t, filepath.Join(".", "helm"), runnerMock.params[2])
		assert.Equal(t, "--set=image.repository=myregistry.com/registry/containers/myFancyContainer", runnerMock.params[3])
		assert.Equal(t, "--set=image.tag=1337", runnerMock.params[4])
		assert.Equal(t, "--values", runnerMock.params[5])
		assert.Equal(t, "./helm/additionalValues.yaml", runnerMock.params[6])
	})

	t.Run("default commit message", func(t *testing.T) {
		var configuration = *validConfiguration
		configuration.CommitMessage = ""

		gitUtilsMock := &gitUtilsMock{}
		runnerMock := &gitOpsExecRunnerMock{}

		err := runGitopsUpdateDeployment(&configuration, runnerMock, gitUtilsMock, &filesMock{})
		assert.NoError(t, err)
		assert.Equal(t, configuration.BranchName, gitUtilsMock.changedBranch)
		assert.Equal(t, expectedYaml, gitUtilsMock.savedFile)
		assert.Equal(t, "Updated myregistry.com/registry/containers/myFancyContainer to version 1337", gitUtilsMock.commitMessage)
		assert.Equal(t, "helm", runnerMock.executable)
		assert.Equal(t, "template", runnerMock.params[0])
		assert.Equal(t, "myFancyDeployment", runnerMock.params[1])
		assert.Equal(t, filepath.Join(".", "helm"), runnerMock.params[2])
		assert.Equal(t, "--set=image.repository=myregistry.com/registry/containers/myFancyContainer", runnerMock.params[3])
		assert.Equal(t, "--set=image.tag=1337", runnerMock.params[4])
		assert.Equal(t, "--values", runnerMock.params[5])
		assert.Equal(t, "./helm/additionalValues.yaml", runnerMock.params[6])
	})

	t.Run("ContainerName not used for helm", func(t *testing.T) {
		var configuration = *validConfiguration
		configuration.ContainerName = "containerName"

		gitUtilsMock := &gitUtilsMock{}
		runnerMock := &gitOpsExecRunnerMock{}

		err := runGitopsUpdateDeployment(&configuration, runnerMock, gitUtilsMock, &filesMock{})
		assert.NoError(t, err)
		assert.Equal(t, configuration.BranchName, gitUtilsMock.changedBranch)
		assert.Equal(t, expectedYaml, gitUtilsMock.savedFile)
		assert.Equal(t, "helm", runnerMock.executable)
		assert.Equal(t, "template", runnerMock.params[0])
		assert.Equal(t, "myFancyDeployment", runnerMock.params[1])
		assert.Equal(t, filepath.Join(".", "helm"), runnerMock.params[2])
		assert.Equal(t, "--set=image.repository=myregistry.com/registry/containers/myFancyContainer", runnerMock.params[3])
		assert.Equal(t, "--set=image.tag=1337", runnerMock.params[4])
		assert.Equal(t, "--values", runnerMock.params[5])
		assert.Equal(t, "./helm/additionalValues.yaml", runnerMock.params[6])
	})

	t.Run("HelmValues is optional", func(t *testing.T) {
		var configuration = *validConfiguration
		configuration.HelmValues = nil

		gitUtilsMock := &gitUtilsMock{}
		runnerMock := &gitOpsExecRunnerMock{}

		err := runGitopsUpdateDeployment(&configuration, runnerMock, gitUtilsMock, &filesMock{})
		assert.NoError(t, err)
		assert.Equal(t, configuration.BranchName, gitUtilsMock.changedBranch)
		assert.Equal(t, expectedYaml, gitUtilsMock.savedFile)
		assert.Equal(t, "helm", runnerMock.executable)
		assert.Equal(t, "template", runnerMock.params[0])
		assert.Equal(t, "myFancyDeployment", runnerMock.params[1])
		assert.Equal(t, filepath.Join(".", "helm"), runnerMock.params[2])
		assert.Equal(t, "--set=image.repository=myregistry.com/registry/containers/myFancyContainer", runnerMock.params[3])
		assert.Equal(t, "--set=image.tag=1337", runnerMock.params[4])
	})

	t.Run("erroneous URL", func(t *testing.T) {
		var configuration = *validConfiguration
		configuration.ContainerRegistryURL = "://myregistry.com"

		err := runGitopsUpdateDeployment(&configuration, &gitOpsExecRunnerMock{}, &gitUtilsMock{}, &filesMock{})
		assert.Error(t, err)
		assert.EqualError(t, err, `failed to apply helm command: failed to extract registry URL, image name, and image tag: registry URL could not be extracted: invalid registry url: parse "://myregistry.com": missing protocol scheme`)
	})

	t.Run("missing ChartPath", func(t *testing.T) {
		var configuration = *validConfiguration
		configuration.ChartPath = ""

		err := runGitopsUpdateDeployment(&configuration, &gitOpsExecRunnerMock{}, &gitUtilsMock{}, &filesMock{})
		assert.Error(t, err)
		assert.EqualError(t, err, "missing required fields for helm: the following parameters are necessary for helm: [chartPath]")
	})

	t.Run("missing DeploymentName", func(t *testing.T) {
		var configuration = *validConfiguration
		configuration.DeploymentName = ""

		err := runGitopsUpdateDeployment(&configuration, &gitOpsExecRunnerMock{}, &gitUtilsMock{}, &filesMock{})
		assert.Error(t, err)
		assert.EqualError(t, err, "missing required fields for helm: the following parameters are necessary for helm: [deploymentName]")
	})

	t.Run("missing DeploymentName and ChartPath", func(t *testing.T) {
		var configuration = *validConfiguration
		configuration.DeploymentName = ""
		configuration.ChartPath = ""

		err := runGitopsUpdateDeployment(&configuration, &gitOpsExecRunnerMock{}, &gitUtilsMock{}, &filesMock{})
		assert.Error(t, err)
		assert.EqualError(t, err, "missing required fields for helm: the following parameters are necessary for helm: [chartPath deploymentName]")
	})

	t.Run("erroneous tag", func(t *testing.T) {
		var configuration = *validConfiguration
		configuration.ContainerImageNameTag = "registry/containers/myFancyContainer:"

		err := runGitopsUpdateDeployment(&configuration, &gitOpsExecRunnerMock{}, &gitUtilsMock{}, &filesMock{})
		assert.Error(t, err)
		assert.EqualError(t, err, "failed to apply helm command: failed to extract registry URL, image name, and image tag: tag could not be extracted")
	})

	t.Run("erroneous image name", func(t *testing.T) {
		var configuration = *validConfiguration
		configuration.ContainerImageNameTag = ":1.0.1"

		err := runGitopsUpdateDeployment(&configuration, &gitOpsExecRunnerMock{}, &gitUtilsMock{}, &filesMock{})
		assert.Error(t, err)
		assert.EqualError(t, err, "failed to apply helm command: failed to extract registry URL, image name, and image tag: image name could not be extracted")
	})

	t.Run("error on helm execution", func(t *testing.T) {
		runner := &gitOpsExecRunnerMock{failOnRunExecutable: true}

		err := runGitopsUpdateDeployment(validConfiguration, runner, &gitUtilsMock{}, &filesMock{})
		assert.Error(t, err)
		assert.EqualError(t, err, "failed to apply helm command: failed to execute helm command: error happened")
	})

	t.Run("error on plain clone", func(t *testing.T) {
		gitUtils := &gitUtilsMock{failOnClone: true}

		err := runGitopsUpdateDeployment(validConfiguration, &gitOpsExecRunnerMock{}, gitUtils, &filesMock{})
		assert.EqualError(t, err, "repository could not get prepared: failed to plain clone repository: error on clone")
	})

	t.Run("error on change branch", func(t *testing.T) {
		gitUtils := &gitUtilsMock{failOnChangeBranch: true}

		err := runGitopsUpdateDeployment(validConfiguration, &gitOpsExecRunnerMock{}, gitUtils, &filesMock{})
		assert.EqualError(t, err, "repository could not get prepared: failed to change branch: error on change branch")
	})

	t.Run("error on commit changes", func(t *testing.T) {
		gitUtils := &gitUtilsMock{failOnCommit: true}

		err := runGitopsUpdateDeployment(validConfiguration, &gitOpsExecRunnerMock{}, gitUtils, &filesMock{})
		assert.EqualError(t, err, "failed to commit and push changes: committing changes failed: error on commit")
	})

	t.Run("error on push commits", func(t *testing.T) {
		gitUtils := &gitUtilsMock{failOnPush: true}

		err := runGitopsUpdateDeployment(validConfiguration, &gitOpsExecRunnerMock{}, gitUtils, &filesMock{})
		assert.EqualError(t, err, "failed to commit and push changes: pushing changes failed: error on push")
	})

	t.Run("error on temp dir creation", func(t *testing.T) {
		fileUtils := &filesMock{failOnCreation: true}

		err := runGitopsUpdateDeployment(validConfiguration, &gitOpsExecRunnerMock{}, &gitUtilsMock{}, fileUtils)
		assert.EqualError(t, err, "failed to create temporary directory: error appeared")
	})

	t.Run("error on file write", func(t *testing.T) {
		fileUtils := &filesMock{failOnWrite: true}

		err := runGitopsUpdateDeployment(validConfiguration, &gitOpsExecRunnerMock{}, &gitUtilsMock{}, fileUtils)
		assert.EqualError(t, err, "failed to write file: error appeared")
	})

	t.Run("error on temp dir deletion", func(t *testing.T) {
		fileUtils := &filesMock{failOnDeletion: true}

		err := runGitopsUpdateDeployment(validConfiguration, &gitOpsExecRunnerMock{}, &gitUtilsMock{}, fileUtils)
		assert.NoError(t, err)
		_ = piperutils.Files{}.RemoveAll(fileUtils.path)
	})
}

type gitOpsExecRunnerMock struct {
	out                 io.Writer
	params              []string
	executable          string
	failOnRunExecutable bool
}

func (e *gitOpsExecRunnerMock) Stdout(out io.Writer) {
	e.out = out
}

func (gitOpsExecRunnerMock) Stderr(io.Writer) {
	panic("implement me")
}

func (e *gitOpsExecRunnerMock) RunExecutable(executable string, params ...string) error {
	if e.failOnRunExecutable {
		return errors.New("error happened")
	}
	e.executable = executable
	e.params = params
	_, err := e.out.Write([]byte(expectedYaml))
	return err
}

type filesMock struct {
	failOnCreation bool
	failOnDeletion bool
	failOnWrite    bool
	path           string
}

func (f filesMock) FileWrite(path string, content []byte, perm os.FileMode) error {
	if f.failOnWrite {
		return errors.New("error appeared")
	}
	return piperutils.Files{}.FileWrite(path, content, perm)
}

func (f filesMock) TempDir(dir string, pattern string) (name string, err error) {
	if f.failOnCreation {
		return "", errors.New("error appeared")
	}
	return piperutils.Files{}.TempDir(dir, pattern)
}

func (f *filesMock) RemoveAll(path string) error {
	if f.failOnDeletion {
		f.path = path
		return errors.New("error appeared")
	}
	return piperutils.Files{}.RemoveAll(path)
}

type gitUtilsMock struct {
	savedFile          string
	changedBranch      string
	commitMessage      string
	temporaryDirectory string
	failOnClone        bool
	failOnChangeBranch bool
	failOnCommit       bool
	failOnPush         bool
}

func (gitUtilsMock) GetWorktree() (*git.Worktree, error) {
	return nil, nil
}

func (v *gitUtilsMock) ChangeBranch(branchName string) error {
	if v.failOnChangeBranch {
		return errors.New("error on change branch")
	}
	v.changedBranch = branchName
	return nil
}

func (v *gitUtilsMock) CommitSingleFile(_ string, commitMessage string, _ string) (plumbing.Hash, error) {
	if v.failOnCommit {
		return [20]byte{}, errors.New("error on commit")
	}

	v.commitMessage = commitMessage

	matches, _ := piperutils.Files{}.Glob(v.temporaryDirectory + "/dir1/dir2/depl.yaml")
	if len(matches) < 1 {
		return [20]byte{}, errors.New("could not find file")
	}
	fileRead, _ := piperutils.Files{}.FileRead(matches[0])
	v.savedFile = string(fileRead)
	return [20]byte{123}, nil
}

func (v gitUtilsMock) PushChangesToRepository(string, string) error {
	if v.failOnPush {
		return errors.New("error on push")
	}
	return nil
}

func (v *gitUtilsMock) PlainClone(_, _, _, directory string) error {
	if v.failOnClone {
		return errors.New("error on clone")
	}
	v.temporaryDirectory = directory
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
