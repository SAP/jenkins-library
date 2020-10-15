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
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"path/filepath"
)

type GitopsGitUtils interface {
	CommitSingleFile(filePath, commitMessage string, repository *git.Repository) (plumbing.Hash, error)
	PushChangesToRepository(username, password string, repository *git.Repository) error
	PlainClone(username, password, serverUrl, directory string) (*git.Repository, error)
	ChangeBranch(branchName string, repository *git.Repository) error
}

type GitopsFileUtils interface {
	TempDir(dir, pattern string) (name string, err error)
	RemoveAll(path string) error
}

var gitUtilities GitopsGitUtils = gitUtil.TheGitUtils{}
var fileUtilities GitopsFileUtils = piperutils.Files{}

func gitopsUpdateDeployment(config gitopsUpdateDeploymentOptions, telemetryData *telemetry.CustomData) {
	// for command execution use Command
	var c command.ExecRunner = &command.Command{}
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

func runGitopsUpdateDeployment(config *gitopsUpdateDeploymentOptions, command command.ExecRunner) error {
	temporaryFolder, tempDirError := fileUtilities.TempDir(".", "temp-")
	if tempDirError != nil {
		log.Entry().WithError(tempDirError).Error("Failed to create temporary directory")
		return tempDirError
	}

	defer fileUtilities.RemoveAll(temporaryFolder)

	repository, gitCloneError := gitUtilities.PlainClone(config.Username, config.Password, config.ServerURL, temporaryFolder)
	if gitCloneError != nil {
		return gitCloneError
	}

	changeBranchError := gitUtilities.ChangeBranch(config.BranchName, repository)
	if changeBranchError != nil {
		return changeBranchError
	}

	registryImage, buildRegistryError := BuildRegistryPlusImage(config)
	if buildRegistryError != nil {
		return buildRegistryError
	}
	patchString := "{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"" + config.ContainerName + "\",\"image\":\"" + registryImage + "\"}]}}}}"

	filePath := filepath.Join(temporaryFolder, config.FilePath)

	kubectlOutputBytes, err := RunKubeCtlCommand(command, patchString, filePath)
	if err != nil {
		return err
	}

	fileWriteError := piperutils.Files{}.FileWrite(filePath, kubectlOutputBytes, 0755)
	if fileWriteError != nil {
		log.Entry().WithError(fileWriteError).Error("Failing write file step")
		return fileWriteError
	}

	commit, commitError := CommitAndPushChanges(config, repository)
	if commitError != nil {
		return commitError
	}

	log.Entry().Infof("Changes committed with %s", commit.String())

	return nil
}

func RunKubeCtlCommand(command command.ExecRunner, patchString string, filePath string) ([]byte, error) {
	var kubectlOutput = bytes.Buffer{}
	command.Stdout(&kubectlOutput)

	kubeParams := []string{
		fmt.Sprint("patch"),
		fmt.Sprint("--local"),
		fmt.Sprint("--output=yaml"),
		fmt.Sprintf("--patch=%v", patchString),
		fmt.Sprintf("--filename=%v", filePath),
	}
	kubectlError := command.RunExecutable("kubectl", kubeParams...)
	if kubectlError != nil {
		log.Entry().WithError(kubectlError).Error("Failed to apply kubectl command")
		return nil, kubectlError
	}
	return kubectlOutput.Bytes(), nil
}

func BuildRegistryPlusImage(config *gitopsUpdateDeploymentOptions) (string, error) {
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

func CommitAndPushChanges(config *gitopsUpdateDeploymentOptions, repository *git.Repository) (plumbing.Hash, error) {
	commit, commitError := gitUtilities.CommitSingleFile(config.FilePath, config.CommitMessage, repository)
	if commitError != nil {
		return [20]byte{}, commitError
	}

	pushError := gitUtilities.PushChangesToRepository(config.Username, config.Password, repository)
	if pushError != nil {
		return [20]byte{}, pushError
	}

	return commit, nil
}
