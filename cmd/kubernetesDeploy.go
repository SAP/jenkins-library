package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils/shell"
)

func kubernetesDeploy(myKubernetesDeployOptions kubernetesDeployOptions) error {
	c := command.Command{}
	// reroute stderr output to logging framework, stdout will be used for command interactions
	c.Stderr(log.Entry().Writer())
	runKubernetesDeploy(myKubernetesDeployOptions, &c)
	return nil
}

func runKubernetesDeploy(myKubernetesDeployOptions kubernetesDeployOptions, command envExecRunner) {
	if myKubernetesDeployOptions.DeployTool == "helm" {
		runHelmDeploy(myKubernetesDeployOptions, command)
	} else {
		runKubectlDeploy(myKubernetesDeployOptions, command)
	}
}

func runHelmDeploy(myKubernetesDeployOptions kubernetesDeployOptions, command envExecRunner) {
	_, containerRegistry, err := splitRegistryURL(myKubernetesDeployOptions.ContainerRegistryURL)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Container registry url '%v' incorrect", myKubernetesDeployOptions.ContainerRegistryURL)
	}
	containerImageName, containerImageTag, err := splitFullImageName(myKubernetesDeployOptions.Image)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Container image '%v' incorrect", myKubernetesDeployOptions.Image)
	}
	helmLogFields := map[string]interface{}{}
	helmLogFields["Chart Path"] = myKubernetesDeployOptions.ChartPath
	helmLogFields["Namespace"] = myKubernetesDeployOptions.Namespace
	helmLogFields["Deployment Name"] = myKubernetesDeployOptions.DeploymentName
	helmLogFields["Context"] = myKubernetesDeployOptions.KubeContext
	helmLogFields["Kubeconfig"] = myKubernetesDeployOptions.KubeConfig
	log.Entry().WithFields(helmLogFields).Debug("Calling Helm")

	helmEnv := []string{fmt.Sprintf("KUBECONFIG=%v", myKubernetesDeployOptions.KubeConfig)}
	log.Entry().Debugf("TILLER_NAMESPACE=%v", myKubernetesDeployOptions.TillerNamespace)
	if len(myKubernetesDeployOptions.TillerNamespace) > 0 {
		helmEnv = append(helmEnv, fmt.Sprintf("TILLER_NAMESPACE=%v", myKubernetesDeployOptions.TillerNamespace))
	}
	log.Entry().Debugf("Helm Env: %v", helmEnv)
	command.Env(helmEnv)

	initParams := []string{"init", "--client-only"}
	if err := command.RunExecutable("helm", initParams...); err != nil {
		log.Entry().WithError(err).Fatal("Helm init called failed")
	}

	var dockerRegistrySecret bytes.Buffer
	command.Stdout(&dockerRegistrySecret)
	kubeParams := []string{
		"--insecure-skip-tls-verify=true",
		"create",
		"secret",
		"docker-registry",
		"regsecret",
		fmt.Sprintf("--docker-server=%v", containerRegistry),
		fmt.Sprintf("--docker-username=%v", shell.WrapInQuotes(myKubernetesDeployOptions.ContainerRegistryUser)),
		fmt.Sprintf("--docker-password=%v", shell.WrapInQuotes(myKubernetesDeployOptions.ContainerRegistryPassword)),
		"--dry-run=true",
		"--output=json",
	}
	log.Entry().Infof("Calling kubectl with parameters %v", kubeParams)
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

	ingressHosts := ""
	for i, h := range myKubernetesDeployOptions.IngressHosts {
		ingressHosts += fmt.Sprintf(",ingress.hosts[%v]=%v", i, h)
	}

	upgradeParams := []string{
		"upgrade",
		myKubernetesDeployOptions.DeploymentName,
		myKubernetesDeployOptions.ChartPath,
		"--install",
		"--force",
		"--namespace",
		myKubernetesDeployOptions.Namespace,
		"--wait",
		"--timeout",
		strconv.Itoa(myKubernetesDeployOptions.HelmDeployWaitSeconds),
		"--set",
		fmt.Sprintf("image.repository=%v/%v,image.tag=%v,secret.dockerconfigjson=%v%v", containerRegistry, containerImageName, containerImageTag, dockerRegistrySecretData.Data.DockerConfJSON, ingressHosts),
	}

	if len(myKubernetesDeployOptions.KubeContext) > 0 {
		upgradeParams = append(upgradeParams, "--kube-context", myKubernetesDeployOptions.KubeContext)
	}

	if len(myKubernetesDeployOptions.AdditionalParameters) > 0 {
		upgradeParams = append(upgradeParams, myKubernetesDeployOptions.AdditionalParameters...)
	}

	log.Entry().Infof("Calling helm with parameters %v", upgradeParams)
	command.RunExecutable("helm", upgradeParams...)
	if err := command.RunExecutable("helm", upgradeParams...); err != nil {
		log.Entry().WithError(err).Fatal("Helm upgrade call failed")
	}

}

func runKubectlDeploy(myKubernetesDeployOptions kubernetesDeployOptions, command envExecRunner) {

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
