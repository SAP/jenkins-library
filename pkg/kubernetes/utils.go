package kubernetes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/docker"
	"github.com/SAP/jenkins-library/pkg/piperutils"

	"github.com/SAP/jenkins-library/pkg/log"
)

// HelmDeployUtils interface
type DeployUtils interface {
	SetEnv(env []string)
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	RunExecutable(e string, p ...string) error

	piperutils.FileUtils
}

// deployUtilsBundle struct  for utils
type deployUtilsBundle struct {
	*command.Command
	*piperutils.Files
}

// NewDeployUtilsBundle initialize using deployUtilsBundle struct
func NewDeployUtilsBundle() DeployUtils {
	utils := deployUtilsBundle{
		Command: &command.Command{
			ErrorCategoryMapping: map[string][]string{
				log.ErrorConfiguration.String(): {
					"Error: Get * no such host",
					"Error: path * not found",
					"Error: rendered manifests contain a resource that already exists.",
					"Error: unknown flag",
					"Error: UPGRADE FAILED: * failed to replace object: * is invalid",
					"Error: UPGRADE FAILED: * failed to create resource: * is invalid",
					"Error: UPGRADE FAILED: an error occurred * not found",
					"Error: UPGRADE FAILED: query: failed to query with labels:",
					"Invalid value: \"\": field is immutable",
				},
				log.ErrorCustom.String(): {
					"Error: release * failed, * timed out waiting for the condition",
				},
			},
		},
		Files: &piperutils.Files{},
	}
	// reroute stderr output to logging framework, stdout will be used for command interactions
	utils.Stderr(log.Writer())
	return &utils
}

func getContainerInfo(config HelmExecuteOptions) (map[string]string, error) {
	var err error
	containerRegistry, err := docker.ContainerRegistryFromURL(config.ContainerRegistryURL)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Container registry url '%v' incorrect", config.ContainerRegistryURL)
	}

	//support either image or containerImageName and containerImageTag
	containerInfo := map[string]string{
		"containerImageName": "",
		"containerImageTag":  "",
		"containerRegistry":  containerRegistry,
	}

	if len(config.Image) > 0 {
		ref, err := docker.ContainerImageNameTagFromImage(config.Image)
		if err != nil {
			log.Entry().WithError(err).Fatalf("Container image '%v' incorrect", config.Image)
		}
		parts := strings.Split(ref, ":")
		containerInfo["containerImageName"] = parts[0]
		containerInfo["containerImageTag"] = parts[1]
	} else if len(config.ContainerImageName) > 0 && len(config.ContainerImageTag) > 0 {
		containerInfo["containerImageName"] = config.ContainerImageName
		containerInfo["containerImageTag"] = config.ContainerImageTag
	} else {
		return nil, fmt.Errorf("image information not given - please either set image or containerImageName and containerImageTag")
	}

	return containerInfo, nil
}

func getSecretsData(config HelmExecuteOptions, utils DeployUtils, containerInfo map[string]string) (string, error) {

	var secretsData string
	if len(config.DockerConfigJSON) == 0 && (len(config.ContainerRegistryUser) == 0 || len(config.ContainerRegistryPassword) == 0) {
		log.Entry().Info("No/incomplete container registry credentials and no docker config.json file provided: skipping secret creation")
		if len(config.ContainerRegistrySecret) > 0 {
			secretsData = fmt.Sprintf(",imagePullSecrets[0].name=%v", config.ContainerRegistrySecret)
		}
	} else {
		var dockerRegistrySecret bytes.Buffer
		utils.Stdout(&dockerRegistrySecret)
		kubeSecretParams := defineKubeSecretParams(config, containerInfo["containerRegistry"], utils)
		log.Entry().Infof("Calling kubectl create secret --dry-run=true ...")
		log.Entry().Debugf("kubectl parameters %v", kubeSecretParams)
		if err := utils.RunExecutable("kubectl", kubeSecretParams...); err != nil {
			log.Entry().WithError(err).Fatal("Retrieving Docker config via kubectl failed")
		}

		var dockerRegistrySecretData struct {
			Kind string `json:"kind"`
			Data struct {
				DockerConfJSON string `json:".dockerconfigjson"`
			} `json:"data"`
			Type string `json:"type"`
		}
		if err := json.Unmarshal(dockerRegistrySecret.Bytes(), &dockerRegistrySecretData); err != nil {
			log.Entry().WithError(err).Fatal("Reading docker registry secret json failed")
		}
		// make sure that secret is hidden in log output
		log.RegisterSecret(dockerRegistrySecretData.Data.DockerConfJSON)

		log.Entry().Debugf("Secret created: %v", dockerRegistrySecret.String())

		// pass secret in helm default template way and in Piper backward compatible way
		secretsData = fmt.Sprintf(",secret.name=%v,secret.dockerconfigjson=%v,imagePullSecrets[0].name=%v", config.ContainerRegistrySecret, dockerRegistrySecretData.Data.DockerConfJSON, config.ContainerRegistrySecret)
	}

	return secretsData, nil
}

func defineKubeSecretParams(config HelmExecuteOptions, containerRegistry string, utils DeployUtils) []string {
	kubeSecretParams := []string{
		"create",
		"secret",
	}
	kubeSecretParams = append(
		kubeSecretParams,
		"--insecure-skip-tls-verify=true",
		"--dry-run=true",
		"--output=json",
	)

	if len(config.DockerConfigJSON) > 0 {
		// first enhance config.json with additional pipeline-related credentials if they have been provided
		if len(containerRegistry) > 0 && len(config.ContainerRegistryUser) > 0 && len(config.ContainerRegistryPassword) > 0 {
			var err error
			_, err = docker.CreateDockerConfigJSON(containerRegistry, config.ContainerRegistryUser, config.ContainerRegistryPassword, "", config.DockerConfigJSON, utils)
			if err != nil {
				log.Entry().Warningf("failed to update Docker config.json: %v", err)
			}
		}

		return append(
			kubeSecretParams,
			"generic",
			config.ContainerRegistrySecret,
			fmt.Sprintf("--from-file=.dockerconfigjson=%v", config.DockerConfigJSON),
			"--type=kubernetes.io/dockerconfigjson",
		)
	}
	return append(
		kubeSecretParams,
		"docker-registry",
		config.ContainerRegistrySecret,
		fmt.Sprintf("--docker-server=%v", containerRegistry),
		fmt.Sprintf("--docker-username=%v", config.ContainerRegistryUser),
		fmt.Sprintf("--docker-password=%v", config.ContainerRegistryPassword),
	)
}
