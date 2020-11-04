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
	"github.com/pkg/errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const toolKubectl = "kubectl"
const toolHelm = "helm"

type iGitopsUpdateDeploymentGitUtils interface {
	CommitSingleFile(filePath, commitMessage, author string) (plumbing.Hash, error)
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

func (g *gitopsUpdateDeploymentGitUtils) CommitSingleFile(filePath, commitMessage, author string) (plumbing.Hash, error) {
	return gitUtil.CommitSingleFile(filePath, commitMessage, author, g.worktree)
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

func gitopsUpdateDeployment(config gitopsUpdateDeploymentOptions, _ *telemetry.CustomData) {
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
	err := checkRequiredFieldsForDeployTool(config)
	if err != nil {
		return err
	}

	temporaryFolder, err := fileUtils.TempDir(".", "temp-")
	if err != nil {
		return errors.Wrap(err, "failed to create temporary directory")
	}

	defer func() {
		err = fileUtils.RemoveAll(temporaryFolder)
		if err != nil {
			log.Entry().WithError(err).Error("error during temporary directory deletion")
		}
	}()

	err = cloneRepositoryAndChangeBranch(config, gitUtils, temporaryFolder)
	if err != nil {
		return errors.Wrap(err, "repository could not get prepared")
	}

	filePath := filepath.Join(temporaryFolder, config.FilePath)

	var outputBytes []byte
	if config.Tool == toolKubectl {
		outputBytes, err = executeKubectl(config, command, outputBytes, filePath)
		if err != nil {
			return errors.Wrap(err, "error on kubectl execution")
		}
	} else if config.Tool == toolHelm {
		outputBytes, err = runHelmCommand(command, config)
		if err != nil {
			return errors.Wrap(err, "failed to apply helm command")
		}
	} else {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.New("tool " + config.Tool + " is not supported")
	}

	err = fileUtils.FileWrite(filePath, outputBytes, 0755)
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

func checkRequiredFieldsForDeployTool(config *gitopsUpdateDeploymentOptions) error {
	if config.Tool == toolHelm {
		err := checkRequiredFieldsForHelm(config)
		if err != nil {
			return errors.Wrap(err, "missing required fields for helm")
		}
		logNotRequiredButFilledFieldForHelm(config)
	} else if config.Tool == toolKubectl {
		err := checkRequiredFieldsForKubectl(config)
		if err != nil {
			return errors.Wrap(err, "missing required fields for kubectl")
		}
		logNotRequiredButFilledFieldForKubectl(config)
	}

	return nil
}

func checkRequiredFieldsForHelm(config *gitopsUpdateDeploymentOptions) error {
	var missingParameters []string
	if config.ChartPath == "" {
		missingParameters = append(missingParameters, "chartPath")
	}
	if config.DeploymentName == "" {
		missingParameters = append(missingParameters, "deploymentName")
	}
	if len(missingParameters) > 0 {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Errorf("the following parameters are necessary for helm: %v", missingParameters)
	}
	return nil
}

func checkRequiredFieldsForKubectl(config *gitopsUpdateDeploymentOptions) error {
	var missingParameters []string
	if config.ContainerName == "" {
		missingParameters = append(missingParameters, "containerName")
	}
	if len(missingParameters) > 0 {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Errorf("the following parameters are necessary for kubectl: %v", missingParameters)
	}
	return nil
}

func logNotRequiredButFilledFieldForHelm(config *gitopsUpdateDeploymentOptions) {
	if config.ContainerName != "" {
		log.Entry().Info("containerName is not used for helm and can be removed")
	}
}

func logNotRequiredButFilledFieldForKubectl(config *gitopsUpdateDeploymentOptions) {
	if config.ChartPath != "" {
		log.Entry().Info("chartPath is not used for kubectl and can be removed")
	}
	if len(config.HelmValues) > 0 {
		log.Entry().Info("helmValues is not used for kubectl and can be removed")
	}
	if len(config.DeploymentName) > 0 {
		log.Entry().Info("deploymentName is not used for kubectl and can be removed")
	}
}

func cloneRepositoryAndChangeBranch(config *gitopsUpdateDeploymentOptions, gitUtils iGitopsUpdateDeploymentGitUtils, temporaryFolder string) error {
	err := gitUtils.PlainClone(config.Username, config.Password, config.ServerURL, temporaryFolder)
	if err != nil {
		return errors.Wrap(err, "failed to plain clone repository")
	}

	err = gitUtils.ChangeBranch(config.BranchName)
	if err != nil {
		return errors.Wrap(err, "failed to change branch")
	}
	return nil
}

func executeKubectl(config *gitopsUpdateDeploymentOptions, command gitopsUpdateDeploymentExecRunner, outputBytes []byte, filePath string) ([]byte, error) {
	registryImage, err := buildRegistryPlusImage(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to apply kubectl command")
	}
	patchString := "{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"" + config.ContainerName + "\",\"image\":\"" + registryImage + "\"}]}}}}"

	outputBytes, err = runKubeCtlCommand(command, patchString, filePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to apply kubectl command")
	}
	return outputBytes, nil
}

func buildRegistryPlusImage(config *gitopsUpdateDeploymentOptions) (string, error) {
	registryURL := config.ContainerRegistryURL
	if registryURL == "" {
		return config.ContainerImageNameTag, nil
	}

	url, err := docker.ContainerRegistryFromURL(registryURL)
	if err != nil {
		return "", errors.Wrap(err, "registry URL could not be extracted")
	}
	if url != "" {
		url = url + "/"
	}
	return url + config.ContainerImageNameTag, nil
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
	err := command.RunExecutable(toolKubectl, kubeParams...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to apply kubectl command")
	}
	return kubectlOutput.Bytes(), nil
}

func runHelmCommand(runner gitopsUpdateDeploymentExecRunner, config *gitopsUpdateDeploymentOptions) ([]byte, error) {
	var helmOutput = bytes.Buffer{}
	runner.Stdout(&helmOutput)

	registryImage, imageTag, err := buildRegistryPlusImageAndTagSeparately(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to extract registry URL, image name, and image tag")
	}
	helmParams := []string{
		"template",
		config.DeploymentName,
		filepath.Join(".", config.ChartPath),
		"--set=image.repository=" + registryImage,
		"--set=image.tag=" + imageTag,
	}

	for _, value := range config.HelmValues {
		helmParams = append(helmParams, "--values", value)
	}

	err = runner.RunExecutable(toolHelm, helmParams...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute helm command")
	}
	return helmOutput.Bytes(), nil
}

// buildRegistryPlusImageAndTagSeparately combines the registry together with the image name. Handles the tag separately.
// Tag is defined by everything on the right hand side of the colon sign. This looks weird for sha container versions but works for helm.
func buildRegistryPlusImageAndTagSeparately(config *gitopsUpdateDeploymentOptions) (string, string, error) {
	registryURL := config.ContainerRegistryURL
	url := ""
	if registryURL != "" {
		containerURL, err := docker.ContainerRegistryFromURL(registryURL)
		if err != nil {
			return "", "", errors.Wrap(err, "registry URL could not be extracted")
		}
		if containerURL != "" {
			containerURL = containerURL + "/"
		}
		url = containerURL
	}

	imageNameTag := config.ContainerImageNameTag
	var imageName, imageTag string
	if strings.Contains(imageNameTag, ":") {
		split := strings.Split(imageNameTag, ":")
		if split[0] == "" {
			log.SetErrorCategory(log.ErrorConfiguration)
			return "", "", errors.New("image name could not be extracted")
		}
		if split[1] == "" {
			log.SetErrorCategory(log.ErrorConfiguration)
			return "", "", errors.New("tag could not be extracted")
		}
		imageName = split[0]
		imageTag = split[1]
		return url + imageName, imageTag, nil
	}

	log.SetErrorCategory(log.ErrorConfiguration)
	return "", "", errors.New("image name and tag could not be extracted")

}

func commitAndPushChanges(config *gitopsUpdateDeploymentOptions, gitUtils iGitopsUpdateDeploymentGitUtils) (plumbing.Hash, error) {
	commitMessage := config.CommitMessage

	if commitMessage == "" {
		commitMessage = defaultCommitMessage(config)
	}

	commit, err := gitUtils.CommitSingleFile(config.FilePath, commitMessage, config.Username)
	if err != nil {
		return [20]byte{}, errors.Wrap(err, "committing changes failed")
	}

	err = gitUtils.PushChangesToRepository(config.Username, config.Password)
	if err != nil {
		return [20]byte{}, errors.Wrap(err, "pushing changes failed")
	}

	return commit, nil
}

func defaultCommitMessage(config *gitopsUpdateDeploymentOptions) string {
	image, tag, _ := buildRegistryPlusImageAndTagSeparately(config)
	commitMessage := fmt.Sprintf("Updated %v to version %v", image, tag)
	return commitMessage
}
