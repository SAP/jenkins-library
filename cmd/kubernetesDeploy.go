package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func kubernetesDeploy(config kubernetesDeployOptions, telemetryData *telemetry.CustomData) {
	c := command.Command{
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
	}
	// reroute stderr output to logging framework, stdout will be used for command interactions
	c.Stderr(log.Writer())
	runKubernetesDeploy(config, &c, log.Writer())
}

func runKubernetesDeploy(config kubernetesDeployOptions, command command.ExecRunner, stdout io.Writer) {
	if config.DeployTool == "helm" || config.DeployTool == "helm3" {
		runHelmDeploy(config, command, stdout)
	} else {
		runKubectlDeploy(config, command)
	}
}

func runHelmDeploy(config kubernetesDeployOptions, command command.ExecRunner, stdout io.Writer) {
	_, containerRegistry, err := splitRegistryURL(config.ContainerRegistryURL)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Container registry url '%v' incorrect", config.ContainerRegistryURL)
	}
	containerImageName, containerImageTag, err := splitFullImageName(config.Image)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Container image '%v' incorrect", config.Image)
	}
	helmLogFields := map[string]interface{}{}
	helmLogFields["Chart Path"] = config.ChartPath
	helmLogFields["Namespace"] = config.Namespace
	helmLogFields["Deployment Name"] = config.DeploymentName
	helmLogFields["Context"] = config.KubeContext
	helmLogFields["Kubeconfig"] = config.KubeConfig
	log.Entry().WithFields(helmLogFields).Debug("Calling Helm")

	helmEnv := []string{fmt.Sprintf("KUBECONFIG=%v", config.KubeConfig)}
	if config.DeployTool == "helm" && len(config.TillerNamespace) > 0 {
		helmEnv = append(helmEnv, fmt.Sprintf("TILLER_NAMESPACE=%v", config.TillerNamespace))
	}
	log.Entry().Debugf("Helm SetEnv: %v", helmEnv)
	command.SetEnv(helmEnv)
	command.Stdout(stdout)

	if config.DeployTool == "helm" {
		initParams := []string{"init", "--client-only"}
		if err := command.RunExecutable("helm", initParams...); err != nil {
			log.Entry().WithError(err).Fatal("Helm init call failed")
		}
	}

	var secretsData string
	if len(config.ContainerRegistryUser) == 0 || len(config.ContainerRegistryPassword) == 0 {
		log.Entry().Info("No container registry credentials provided or credentials incomplete: skipping secret creation")
		if len(config.ContainerRegistrySecret) > 0 {
			secretsData = fmt.Sprintf(",imagePullSecrets[0].name=%v", config.ContainerRegistrySecret)
		}
	} else {
		var dockerRegistrySecret bytes.Buffer
		command.Stdout(&dockerRegistrySecret)
		kubeParams := []string{
			"--insecure-skip-tls-verify=true",
			"create",
			"secret",
			"docker-registry",
			config.ContainerRegistrySecret,
			fmt.Sprintf("--docker-server=%v", containerRegistry),
			fmt.Sprintf("--docker-username=%v", config.ContainerRegistryUser),
			fmt.Sprintf("--docker-password=%v", config.ContainerRegistryPassword),
			"--dry-run=true",
			"--output=json",
		}
		log.Entry().Infof("Calling kubectl create secret --dry-run=true ...")
		log.Entry().Debugf("kubectl parameters %v", kubeParams)
		if err := command.RunExecutable("kubectl", kubeParams...); err != nil {
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

		log.Entry().Debugf("Secret created: %v", string(dockerRegistrySecret.Bytes()))

		// pass secret in helm default template way and in Piper backward compatible way
		secretsData = fmt.Sprintf(",secret.name=%v,secret.dockerconfigjson=%v,imagePullSecrets[0].name=%v", config.ContainerRegistrySecret, dockerRegistrySecretData.Data.DockerConfJSON, config.ContainerRegistrySecret)
	}

	// Deprecated functionality
	// only for backward compatible handling of ingress.hosts
	// this requires an adoption of the default ingress.yaml template
	// Due to the way helm is implemented it is currently not possible to overwrite a part of a list:
	// see: https://github.com/helm/helm/issues/5711#issuecomment-636177594
	// Recommended way is to use a custom values file which contains the appropriate data
	ingressHosts := ""
	for i, h := range config.IngressHosts {
		ingressHosts += fmt.Sprintf(",ingress.hosts[%v]=%v", i, h)
	}

	upgradeParams := []string{
		"upgrade",
		config.DeploymentName,
		config.ChartPath,
	}

	for _, v := range config.HelmValues {
		upgradeParams = append(upgradeParams, "--values", v)
	}

	upgradeParams = append(
		upgradeParams,
		"--install",
		"--force",
		"--namespace", config.Namespace,
		"--set",
		fmt.Sprintf("image.repository=%v/%v,image.tag=%v%v%v", containerRegistry, containerImageName, containerImageTag, secretsData, ingressHosts),
	)

	if config.DeployTool == "helm" {
		upgradeParams = append(upgradeParams, "--wait", "--timeout", strconv.Itoa(config.HelmDeployWaitSeconds))
	}

	if config.DeployTool == "helm3" {
		upgradeParams = append(upgradeParams, "--atomic", "--timeout", fmt.Sprintf("%vs", config.HelmDeployWaitSeconds))
	}

	if len(config.KubeContext) > 0 {
		upgradeParams = append(upgradeParams, "--kube-context", config.KubeContext)
	}

	if len(config.AdditionalParameters) > 0 {
		upgradeParams = append(upgradeParams, config.AdditionalParameters...)
	}

	command.Stdout(stdout)
	log.Entry().Info("Calling helm upgrade ...")
	log.Entry().Debugf("Helm parameters %v", upgradeParams)
	if err := command.RunExecutable("helm", upgradeParams...); err != nil {
		log.Entry().WithError(err).Fatal("Helm upgrade call failed")
	}

}

func runKubectlDeploy(config kubernetesDeployOptions, command command.ExecRunner) {
	_, containerRegistry, err := splitRegistryURL(config.ContainerRegistryURL)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Container registry url '%v' incorrect", config.ContainerRegistryURL)
	}

	kubeParams := []string{
		"--insecure-skip-tls-verify=true",
		fmt.Sprintf("--namespace=%v", config.Namespace),
	}

	if len(config.KubeConfig) > 0 {
		log.Entry().Info("Using KUBECONFIG environment for authentication.")
		kubeEnv := []string{fmt.Sprintf("KUBECONFIG=%v", config.KubeConfig)}
		command.SetEnv(kubeEnv)
		if len(config.KubeContext) > 0 {
			kubeParams = append(kubeParams, fmt.Sprintf("--context=%v", config.KubeContext))
		}

	} else {
		log.Entry().Info("Using --token parameter for authentication.")
		kubeParams = append(kubeParams, fmt.Sprintf("--server=%v", config.APIServer))
		kubeParams = append(kubeParams, fmt.Sprintf("--token=%v", config.KubeToken))
	}

	if config.CreateDockerRegistrySecret {
		if len(config.ContainerRegistryUser)+len(config.ContainerRegistryPassword) == 0 {
			log.Entry().Fatal("Cannot create Container registry secret without proper registry username/password")
		}
		// first check if secret already exists
		kubeCheckParams := append(kubeParams, "get", "secret", config.ContainerRegistrySecret)
		if err := command.RunExecutable("kubectl", kubeCheckParams...); err != nil {
			log.Entry().Infof("Registry secret '%v' does not exist, let's create it ...", config.ContainerRegistrySecret)
			kubeSecretParams := append(
				kubeParams,
				"create",
				"secret",
				"docker-registry",
				config.ContainerRegistrySecret,
				fmt.Sprintf("--docker-server=%v", containerRegistry),
				fmt.Sprintf("--docker-username=%v", config.ContainerRegistryUser),
				fmt.Sprintf("--docker-password=%v", config.ContainerRegistryPassword),
			)
			log.Entry().Infof("Creating container registry secret '%v'", config.ContainerRegistrySecret)
			log.Entry().Debugf("Running kubectl with following parameters: %v", kubeSecretParams)
			if err := command.RunExecutable("kubectl", kubeSecretParams...); err != nil {
				log.Entry().WithError(err).Fatal("Creating container registry secret failed")
			}
		}
	}

	appTemplate, err := ioutil.ReadFile(config.AppTemplate)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Error when reading appTemplate '%v'", config.AppTemplate)
	}

	// Update image name in deployment yaml, expects placeholder like 'image: <image-name>'
	re := regexp.MustCompile(`image:[ ]*<image-name>`)
	appTemplate = []byte(re.ReplaceAllString(string(appTemplate), fmt.Sprintf("image: %v/%v", containerRegistry, config.Image)))

	err = ioutil.WriteFile(config.AppTemplate, appTemplate, 0700)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Error when updating appTemplate '%v'", config.AppTemplate)
	}

	kubeApplyParams := append(kubeParams, "apply", "--filename", config.AppTemplate)
	if len(config.AdditionalParameters) > 0 {
		kubeApplyParams = append(kubeApplyParams, config.AdditionalParameters...)
	}

	if err := command.RunExecutable("kubectl", kubeApplyParams...); err != nil {
		log.Entry().Debugf("Running kubectl with following parameters: %v", kubeApplyParams)
		log.Entry().WithError(err).Fatal("Deployment with kubectl failed.")
	}
}

func splitRegistryURL(registryURL string) (protocol, registry string, err error) {
	parts := strings.Split(registryURL, "://")
	if len(parts) != 2 || len(parts[1]) == 0 {
		return "", "", fmt.Errorf("Failed to split registry url '%v'", registryURL)
	}
	return parts[0], parts[1], nil
}

func splitFullImageName(image string) (imageName, tag string, err error) {
	parts := strings.Split(image, ":")
	switch len(parts) {
	case 0:
		return "", "", fmt.Errorf("Failed to split image name '%v'", image)
	case 1:
		if len(parts[0]) > 0 {
			return parts[0], "", nil
		}
		return "", "", fmt.Errorf("Failed to split image name '%v'", image)
	case 2:
		return parts[0], parts[1], nil
	}
	return "", "", fmt.Errorf("Failed to split image name '%v'", image)
}
