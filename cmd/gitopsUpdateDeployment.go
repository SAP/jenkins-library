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
	"io"
	"path/filepath"
)

type gitopsGitUtils interface {
	CommitSingleFile(filePath, commitMessage string) (plumbing.Hash, error)
	PushChangesToRepository(username, password string) error
	PlainClone(username, password, serverURL, directory string) error
	ChangeBranch(branchName string) error
}

type gitopsFileUtils interface {
	TempDir(dir, pattern string) (name string, err error)
	RemoveAll(path string) error
}

type gitopsExecRunner interface {
	RunExecutable(executable string, params ...string) error
	Stdout(out io.Writer)
	Stderr(err io.Writer)
}

type gitUtilsRuntime struct {
	worktree   *git.Worktree
	repository *git.Repository
}

func (g *gitUtilsRuntime) CommitSingleFile(filePath, commitMessage string) (plumbing.Hash, error) {
	return gitUtil.CommitSingleFile(filePath, commitMessage, g.worktree)
}

func (g *gitUtilsRuntime) PushChangesToRepository(username, password string) error {
	return gitUtil.PushChangesToRepository(username, password, g.repository)
}

func (g *gitUtilsRuntime) PlainClone(username, password, serverURL, directory string) error {
	var err error
	g.repository, err = gitUtil.PlainClone(username, password, serverURL, directory)
	if err != nil {
		return err
	}
	g.worktree, err = g.repository.Worktree()
	return err
}

func (g *gitUtilsRuntime) ChangeBranch(branchName string) error {
	return gitUtil.ChangeBranch(branchName, g.worktree)
}

func gitopsUpdateDeployment(config gitopsUpdateDeploymentOptions, telemetryData *telemetry.CustomData) {
	// for command execution use Command
	var c gitopsExecRunner = &command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	// for http calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runGitopsUpdateDeployment(&config, c, &gitUtilsRuntime{}, piperutils.Files{})
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runGitopsUpdateDeployment(config *gitopsUpdateDeploymentOptions, command gitopsExecRunner, gitopsUpdateDeploymentGitUtilities gitopsGitUtils, gitopsUpdateDeploymentFileUtilities gitopsFileUtils) error {
	temporaryFolder, err := gitopsUpdateDeploymentFileUtilities.TempDir(".", "temp-")
	if err != nil {
		log.Entry().WithError(err).Error("Failed to create temporary directory")
		return err
	}

	defer gitopsUpdateDeploymentFileUtilities.RemoveAll(temporaryFolder)

	err = gitopsUpdateDeploymentGitUtilities.PlainClone(config.Username, config.Password, config.ServerURL, temporaryFolder)
	if err != nil {
		return err
	}

	err = gitopsUpdateDeploymentGitUtilities.ChangeBranch(config.BranchName)
	if err != nil {
		return err
	}

	registryImage, err := buildRegistryPlusImage(config)
	if err != nil {
		return err
	}
	patchString := "{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"" + config.ContainerName + "\",\"image\":\"" + registryImage + "\"}]}}}}"

	filePath := filepath.Join(temporaryFolder, config.FilePath)

	kubectlOutputBytes, err := runKubeCtlCommand(command, patchString, filePath)
	if err != nil {
		return err
	}

	err = piperutils.Files{}.FileWrite(filePath, kubectlOutputBytes, 0755)
	if err != nil {
		log.Entry().WithError(err).Error("Failing write file step")
		return err
	}

	commit, err := commitAndPushChanges(config, gitopsUpdateDeploymentGitUtilities)
	if err != nil {
		return err
	}

	log.Entry().Infof("Changes committed with %s", commit.String())

	return nil
}

func runKubeCtlCommand(command gitopsExecRunner, patchString string, filePath string) ([]byte, error) {
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
		log.Entry().WithError(err).Error("Failed to apply kubectl command")
		return nil, err
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
		log.Entry().WithError(err).Error("registry URL could not be extracted")
		return "", err
	}
	if url != "" {
		url = url + "/"
	}
	return url + config.ContainerImage, nil
}

func commitAndPushChanges(config *gitopsUpdateDeploymentOptions, gitopsUpdateDeploymentGitUtilities gitopsGitUtils) (plumbing.Hash, error) {
	commit, err := gitopsUpdateDeploymentGitUtilities.CommitSingleFile(config.FilePath, config.CommitMessage)
	if err != nil {
		return [20]byte{}, err
	}

	err = gitopsUpdateDeploymentGitUtilities.PushChangesToRepository(config.Username, config.Password)
	if err != nil {
		return [20]byte{}, err
	}

	return commit, nil
}
