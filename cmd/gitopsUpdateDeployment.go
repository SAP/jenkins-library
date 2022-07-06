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
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pkg/errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const toolKubectl = "kubectl"
const toolHelm = "helm"
const toolKustomize = "kustomize"

type iGitopsUpdateDeploymentGitUtils interface {
	CommitFiles(filePaths []string, commitMessage, author string) (plumbing.Hash, error)
	PushChangesToRepository(username, password string, force *bool) error
	PlainClone(username, password, serverURL, directory string) error
	ChangeBranch(branchName string) error
}

type gitopsUpdateDeploymentFileUtils interface {
	TempDir(dir, pattern string) (name string, err error)
	RemoveAll(path string) error
	FileWrite(path string, content []byte, perm os.FileMode) error
	Glob(pattern string) ([]string, error)
}

type gitopsUpdateDeploymentExecRunner interface {
	RunExecutable(executable string, params ...string) error
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	SetDir(dir string)
}

type gitopsUpdateDeploymentGitUtils struct {
	worktree   *git.Worktree
	repository *git.Repository
}

func (g *gitopsUpdateDeploymentGitUtils) CommitFiles(filePaths []string, commitMessage, author string) (plumbing.Hash, error) {
	for _, path := range filePaths {
		_, err := g.worktree.Add(path)

		if err != nil {
			return [20]byte{}, errors.Wrap(err, "failed to add file to git")
		}
	}

	commit, err := g.worktree.Commit(commitMessage, &git.CommitOptions{
		All:    true,
		Author: &object.Signature{Name: author, When: time.Now()},
	})
	if err != nil {
		return [20]byte{}, errors.Wrap(err, "failed to commit file")
	}

	return commit, nil
}

func (g *gitopsUpdateDeploymentGitUtils) PushChangesToRepository(username, password string, force *bool) error {
	return gitUtil.PushChangesToRepository(username, password, force, g.repository)
}

func (g *gitopsUpdateDeploymentGitUtils) PlainClone(username, password, serverURL, directory string) error {
	var err error
	g.repository, err = gitUtil.PlainClone(username, password, serverURL, directory)
	if err != nil {
		return errors.Wrapf(err, "plain clone failed '%s'", serverURL)
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
	temporaryFolder = regexp.MustCompile(`^./`).ReplaceAllString(temporaryFolder, "")
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
	if config.Tool == toolHelm {
		filePath = filepath.Join(temporaryFolder, config.ChartPath)
	}

	allFiles, err := fileUtils.Glob(filePath)
	if err != nil {
		return errors.Wrap(err, "unable to expand globbing pattern")
	} else if len(allFiles) == 0 {
		return errors.New("no matching files found for provided globbing pattern")
	}
	command.SetDir("./")

	var outputBytes []byte
	for _, currentFile := range allFiles {
		if config.Tool == toolKubectl {
			outputBytes, err = executeKubectl(config, command, currentFile)
			if err != nil {
				return errors.Wrap(err, "error on kubectl execution")
			}
		} else if config.Tool == toolHelm {

			out, err := runHelmCommand(command, config, currentFile)
			if err != nil {
				return errors.Wrap(err, "failed to apply helm command")
			}
			// join all helm outputs into the same "FilePath"
			outputBytes = append(outputBytes, []byte("---\n")...)
			outputBytes = append(outputBytes, out...)
			currentFile = filepath.Join(temporaryFolder, config.FilePath)

		} else if config.Tool == toolKustomize {
			_, err = runKustomizeCommand(command, config, currentFile)
			if err != nil {
				return errors.Wrap(err, "failed to apply kustomize command")
			}
			outputBytes = nil
		} else {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.New("tool " + config.Tool + " is not supported")
		}

		if outputBytes != nil {
			err = fileUtils.FileWrite(currentFile, outputBytes, 0755)
			if err != nil {
				return errors.Wrap(err, "failed to write file")
			}
		}
	}
	if config.Tool == toolHelm {
		// helm only creates one output file.
		allFiles = []string{config.FilePath}
	} else {
		// git expects the file path relative to its root:
		for i := range allFiles {
			allFiles[i] = strings.ReplaceAll(allFiles[i], temporaryFolder+"/", "")
		}
	}

	commit, err := commitAndPushChanges(config, gitUtils, allFiles)
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
	} else if config.Tool == toolKustomize {
		err := checkRequiredFieldsForKustomize(config)
		if err != nil {
			return errors.Wrap(err, "missing required fields for kustomize")
		}
		logNotRequiredButFilledFieldForKustomize(config)
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

func checkRequiredFieldsForKustomize(config *gitopsUpdateDeploymentOptions) error {
	var missingParameters []string
	if config.FilePath == "" {
		missingParameters = append(missingParameters, "filePath")
	}
	if config.DeploymentName == "" {
		missingParameters = append(missingParameters, "deploymentName")
	}
	if len(missingParameters) > 0 {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Errorf("the following parameters are necessary for kustomize: %v", missingParameters)
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
func logNotRequiredButFilledFieldForKustomize(config *gitopsUpdateDeploymentOptions) {
	if config.ChartPath != "" {
		log.Entry().Info("chartPath is not used for kubectl and can be removed")
	}
	if len(config.HelmValues) > 0 {
		log.Entry().Info("helmValues is not used for kubectl and can be removed")
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

func executeKubectl(config *gitopsUpdateDeploymentOptions, command gitopsUpdateDeploymentExecRunner, filePath string) ([]byte, error) {
	var outputBytes []byte
	registryImage, err := buildRegistryPlusImage(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to apply kubectl command")
	}
	patchString := "{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"" + config.ContainerName + "\",\"image\":\"" + registryImage + "\"}]}}}}"

	log.Entry().Infof("[kubectl] updating '%s'", filePath)
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

func runHelmCommand(command gitopsUpdateDeploymentExecRunner, config *gitopsUpdateDeploymentOptions, filePath string) ([]byte, error) {
	var helmOutput = bytes.Buffer{}
	command.Stdout(&helmOutput)

	registryImage, imageTag, err := buildRegistryPlusImageAndTagSeparately(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to extract registry URL, image name, and image tag")
	}
	helmParams := []string{
		"template",
		config.DeploymentName,
		filePath,
		"--set=image.repository=" + registryImage,
		"--set=image.tag=" + imageTag,
	}

	for _, value := range config.HelmValues {
		helmParams = append(helmParams, "--values", value)
	}

	log.Entry().Infof("[helmn] updating '%s'", filePath)
	err = command.RunExecutable(toolHelm, helmParams...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute helm command")
	}
	return helmOutput.Bytes(), nil
}

func runKustomizeCommand(command gitopsUpdateDeploymentExecRunner, config *gitopsUpdateDeploymentOptions, filePath string) ([]byte, error) {
	var kustomizeOutput = bytes.Buffer{}
	command.Stdout(&kustomizeOutput)
	registryImage, imageTag, err := buildRegistryPlusImageAndTagSeparately(config)

	kustomizeParams := []string{
		"edit",
		"set",
		"image",
		config.DeploymentName + "=" + registryImage + ":" + imageTag,
	}

	command.SetDir(filepath.Dir(filePath))

	log.Entry().Infof("[kustomize] updating '%s'", filePath)
	err = command.RunExecutable(toolKustomize, kustomizeParams...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute kustomize command")
	}

	return kustomizeOutput.Bytes(), nil
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

func commitAndPushChanges(config *gitopsUpdateDeploymentOptions, gitUtils iGitopsUpdateDeploymentGitUtils, filePaths []string) (plumbing.Hash, error) {
	commitMessage := config.CommitMessage

	if commitMessage == "" {
		commitMessage = defaultCommitMessage(config)
	}

	commit, err := gitUtils.CommitFiles(filePaths, commitMessage, config.Username)
	if err != nil {
		return [20]byte{}, errors.Wrap(err, "committing changes failed")
	}

	err = gitUtils.PushChangesToRepository(config.Username, config.Password, &config.ForcePush)
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
