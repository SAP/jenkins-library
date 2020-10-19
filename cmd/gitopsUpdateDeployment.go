package cmd

import (
	"bytes"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/docker"
	gitUtil "github.com/SAP/jenkins-library/pkg/git"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/go-git/go-git/v5/plumbing"
	"io"
	"path/filepath"
)

type gitopsGitUtils interface {
	CommitSingleFile(filePath, commitMessage string, worktree gitUtil.UtilsWorkTree) (plumbing.Hash, error)
	PushChangesToRepository(username, password string, repository gitUtil.UtilsRepository) error
	PlainClone(username, password, serverURL, directory string) (gitUtil.UtilsRepository, error)
	ChangeBranch(branchName string, worktree gitUtil.UtilsWorkTree) error
	GetWorktree(repository gitUtil.UtilsRepository) (gitUtil.UtilsWorkTree, error)
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

var gitopsUpdateDeploymentGitUtilities gitopsGitUtils = gitUtil.TheGitUtils{}
var gitopsUpdateDeploymentFileUtilities gitopsFileUtils = piperutils.Files{}

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
	err := runGitopsUpdateDeployment(&config, c)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runGitopsUpdateDeployment(config *gitopsUpdateDeploymentOptions, command gitopsExecRunner) error {
	temporaryFolder, err := gitopsUpdateDeploymentFileUtilities.TempDir(".", "temp-")
	if err != nil {
		log.Entry().WithError(err).Error("Failed to create temporary directory")
		return err
	}

	defer gitopsUpdateDeploymentFileUtilities.RemoveAll(temporaryFolder)

	repository, err := gitopsUpdateDeploymentGitUtilities.PlainClone(config.Username, config.Password, config.ServerURL, temporaryFolder)
	if err != nil {
		return err
	}

	worktree, err := repository.Worktree()
	if err != nil {
		return err
	}

	err = gitopsUpdateDeploymentGitUtilities.ChangeBranch(config.BranchName, worktree)
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

	commit, err := commitAndPushChanges(config, repository, worktree)
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
		fmt.Sprint("patch"),
		fmt.Sprint("--local"),
		fmt.Sprint("--output=yaml"),
		fmt.Sprintf("--patch=%v", patchString),
		fmt.Sprintf("--filename=%v", filePath),
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

func commitAndPushChanges(config *gitopsUpdateDeploymentOptions, repository gitUtil.UtilsRepository, worktree gitUtil.UtilsWorkTree) (plumbing.Hash, error) {
	commit, err := gitopsUpdateDeploymentGitUtilities.CommitSingleFile(config.FilePath, config.CommitMessage, worktree)
	if err != nil {
		return [20]byte{}, err
	}

	err = gitopsUpdateDeploymentGitUtilities.PushChangesToRepository(config.Username, config.Password, repository)
	if err != nil {
		return [20]byte{}, err
	}

	return commit, nil
}
