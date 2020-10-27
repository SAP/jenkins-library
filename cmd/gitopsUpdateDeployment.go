package cmd

import (
	"bytes"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/docker"
	gitUtil "github.com/SAP/jenkins-library/pkg/git"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/pkg/errors"
	"io"
	"os"
	"path/filepath"
)

type iGitopsUpdateDeploymentGitUtils interface {
	CommitSingleFile(filePath, commitMessage string) (plumbing.Hash, error)
	PushChangesToRepository(username, password string) error
	PlainClone(username, password, serverURL, directory string) error
	ChangeBranch(branchName string) error
}

type gitopsUpdateDeploymentFileUtils interface {
	TempDir(dir, pattern string) (name string, err error)
	RemoveAll(path string) error
	FileWrite(path string, content []byte, perm os.FileMode) error
}

type gitopsUpdateDeploymentExecRunner interface {
	RunExecutable(executable string, params ...string) error
	Stdout(out io.Writer)
	Stderr(err io.Writer)
}

type gitopsUpdateDeploymentGitUtils struct {
	worktree   *git.Worktree
	repository *git.Repository
}

func (g *gitopsUpdateDeploymentGitUtils) CommitSingleFile(filePath, commitMessage string) (plumbing.Hash, error) {
	return gitUtil.CommitSingleFile(filePath, commitMessage, g.worktree)
}

func (g *gitopsUpdateDeploymentGitUtils) PushChangesToRepository(username, password string) error {
	return gitUtil.PushChangesToRepository(username, password, g.repository)
}

func (g *gitopsUpdateDeploymentGitUtils) PlainClone(username, password, serverURL, directory string) error {
	var err error
	g.repository, err = gitUtil.PlainClone(username, password, serverURL, directory)
	if err != nil {
		return errors.Wrap(err, "plain clone failed")
	}
	g.worktree, err = g.repository.Worktree()
	return errors.Wrap(err, "failed to retrieve worktree")
}

func (g *gitopsUpdateDeploymentGitUtils) ChangeBranch(branchName string) error {
	return gitUtil.ChangeBranch(branchName, g.worktree)
}

func gitopsUpdateDeployment(config gitopsUpdateDeploymentOptions, telemetryData *telemetry.CustomData) {
	// for command execution use Command
	var c gitopsUpdateDeploymentExecRunner = &command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	// for http calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runGitopsUpdateDeployment(&config, c, &gitopsUpdateDeploymentGitUtils{}, piperutils.Files{})
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runGitopsUpdateDeployment(config *gitopsUpdateDeploymentOptions, command gitopsUpdateDeploymentExecRunner, gitUtils iGitopsUpdateDeploymentGitUtils, fileUtils gitopsUpdateDeploymentFileUtils) error {
	temporaryFolder, err := fileUtils.TempDir(".", "temp-")
	if err != nil {
		return errors.Wrap(err, "failed to create temporary directory")
	}

	defer fileUtils.RemoveAll(temporaryFolder)

	err = gitUtils.PlainClone(config.Username, config.Password, config.ServerURL, temporaryFolder)
	if err != nil {
		return errors.Wrap(err, "failed to plain clone repository")
	}

	err = gitUtils.ChangeBranch(config.BranchName)
	if err != nil {
		return errors.Wrap(err, "failed to change branch")
	}

	registryImage, err := buildRegistryPlusImage(config)
	if err != nil {
		return errors.Wrap(err, "failed to apply kubectl command")
	}
	patchString := "{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"" + config.ContainerName + "\",\"image\":\"" + registryImage + "\"}]}}}}"

	filePath := filepath.Join(temporaryFolder, config.FilePath)

	kubectlOutputBytes, err := runKubeCtlCommand(command, patchString, filePath)
	if err != nil {
		return errors.Wrap(err, "failed to apply kubectl command")
	}

	err = fileUtils.FileWrite(filePath, kubectlOutputBytes, 0755)
	if err != nil {
		return errors.Wrap(err, "failed to write file")
	}

	commit, err := commitAndPushChanges(config, gitUtils)
	if err != nil {
		return errors.Wrap(err, "failed to commit and push changes")
	}

	log.Entry().Infof("Changes committed with %s", commit.String())

	return nil
}

func runKubeCtlCommand(command gitopsUpdateDeploymentExecRunner, patchString string, filePath string) ([]byte, error) {
	var kubectlOutput = bytes.Buffer{}
	command.Stdout(&kubectlOutput)

	kubeParams := []string{
		"patch",
		"--local",
		"--output=yaml",
		"--patch=" + patchString,
		"--filename=" + filePath,
	}
	err := command.RunExecutable("kubectl", kubeParams...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to apply kubectl command")
	}
	return kubectlOutput.Bytes(), nil
}

func buildRegistryPlusImage(config *gitopsUpdateDeploymentOptions) (string, error) {
	registryURL := config.ContainerRegistryURL
	if registryURL == "" {
		return config.ContainerImage, nil
	}

	url, err := docker.ContainerRegistryFromURL(registryURL)
	if err != nil {
		return "", errors.Wrap(err, "registry URL could not be extracted")
	}
	if url != "" {
		url = url + "/"
	}
	return url + config.ContainerImage, nil
}

func commitAndPushChanges(config *gitopsUpdateDeploymentOptions, gitUtils iGitopsUpdateDeploymentGitUtils) (plumbing.Hash, error) {
	commit, err := gitUtils.CommitSingleFile(config.FilePath, config.CommitMessage)
	if err != nil {
		return [20]byte{}, errors.Wrap(err, "committing changes failed")
	}

	err = gitUtils.PushChangesToRepository(config.Username, config.Password)
	if err != nil {
		return [20]byte{}, errors.Wrap(err, "pushing changes failed")
	}

	return commit, nil
}
