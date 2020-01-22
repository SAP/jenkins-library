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
)

func kubernetesDeploy(myKubernetesDeployOptions kubernetesDeployOptions) error {
	c := command.Command{}
	// reroute stderr output to logging framework, stdout will be used for command interactions
	c.Stderr(log.Entry().Writer())
	runKubernetesDeploy(myKubernetesDeployOptions, &c, log.Entry().Writer())
	return nil
}

func runKubernetesDeploy(myKubernetesDeployOptions kubernetesDeployOptions, command envExecRunner, stdout io.Writer) {
	if myKubernetesDeployOptions.DeployTool == "helm" {
		runHelmDeploy(myKubernetesDeployOptions, command, stdout)
	} else {
		runKubectlDeploy(myKubernetesDeployOptions, command)
	}
}

func runHelmDeploy(myKubernetesDeployOptions kubernetesDeployOptions, command envExecRunner, stdout io.Writer) {
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
	command.Stdout(stdout)

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
		fmt.Sprintf("--docker-username=%v", myKubernetesDeployOptions.ContainerRegistryUser),
		fmt.Sprintf("--docker-password=%v", myKubernetesDeployOptions.ContainerRegistryPassword),
		"--dry-run=true",
		"--output=json",
	}
	log.Entry().Infof("Calling kubectl create secret --dry-run=true ...")
	log.Entry().Debugf("kubectl parameters %v", kubeParams)
	if err := command.RunExecutable("kubectl", kubeParams...); err != nil {
		log.Entry().WithError(err).Fatal("Retrieving Docker config via kubectl failed")
	}
	log.Entry().Debugf("Secret created: %v", string(dockerRegistrySecret.Bytes()))

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

	command.Stdout(stdout)
	log.Entry().Info("Calling helm upgrade ...")
	log.Entry().Debugf("Helm parameters %v", upgradeParams)
	command.RunExecutable("helm", upgradeParams...)
	if err := command.RunExecutable("helm", upgradeParams...); err != nil {
		log.Entry().WithError(err).Fatal("Helm upgrade call failed")
	}

}

func runKubectlDeploy(myKubernetesDeployOptions kubernetesDeployOptions, command envExecRunner) {
	_, containerRegistry, err := splitRegistryURL(myKubernetesDeployOptions.ContainerRegistryURL)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Container registry url '%v' incorrect", myKubernetesDeployOptions.ContainerRegistryURL)
	}

	kubeParams := []string{
		"--insecure-skip-tls-verify=true",
		fmt.Sprintf("--namespace=%v", myKubernetesDeployOptions.Namespace),
	}

	if len(myKubernetesDeployOptions.KubeConfig) > 0 {
		log.Entry().Info("Using KUBECONFIG environment for authentication.")
		kubeEnv := []string{fmt.Sprintf("KUBECONFIG=%v", myKubernetesDeployOptions.KubeConfig)}
		command.Env(kubeEnv)
		if len(myKubernetesDeployOptions.KubeContext) > 0 {
			kubeParams = append(kubeParams, fmt.Sprintf("--context=%v", myKubernetesDeployOptions.KubeContext))
		}

	} else {
		log.Entry().Info("Using --token parameter for authentication.")
		kubeParams = append(kubeParams, fmt.Sprintf("--server=%v", myKubernetesDeployOptions.APIServer))
		kubeParams = append(kubeParams, fmt.Sprintf("--token=%v", myKubernetesDeployOptions.KubeToken))
	}

	if myKubernetesDeployOptions.CreateDockerRegistrySecret {
		if len(myKubernetesDeployOptions.ContainerRegistryUser)+len(myKubernetesDeployOptions.ContainerRegistryPassword) == 0 {
			log.Entry().Fatal("Cannot create Container registry secret without proper registry username/password")
		}
		// first check if secret already exists
		kubeCheckParams := append(kubeParams, "get", "secret", myKubernetesDeployOptions.ContainerRegistrySecret)
		if err := command.RunExecutable("kubectl", kubeCheckParams...); err != nil {
			log.Entry().Infof("Registry secret '%v' does not exist, let's create it ...", myKubernetesDeployOptions.ContainerRegistrySecret)
			kubeSecretParams := append(
				kubeParams,
				"create",
				"secret",
				"docker-registry",
				myKubernetesDeployOptions.ContainerRegistrySecret,
				fmt.Sprintf("--docker-server=%v", containerRegistry),
				fmt.Sprintf("--docker-username=%v", myKubernetesDeployOptions.ContainerRegistryUser),
				fmt.Sprintf("--docker-password=%v", myKubernetesDeployOptions.ContainerRegistryPassword),
			)
			log.Entry().Infof("Creating container registry secret '%v'", myKubernetesDeployOptions.ContainerRegistrySecret)
			log.Entry().Debugf("Running kubectl with following parameters: %v", kubeSecretParams)
			if err := command.RunExecutable("kubectl", kubeSecretParams...); err != nil {
				log.Entry().WithError(err).Fatal("Creating container registry secret failed")
			}
		}
	}

	appTemplate, err := ioutil.ReadFile(myKubernetesDeployOptions.AppTemplate)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Error when reading appTemplate '%v'", myKubernetesDeployOptions.AppTemplate)
	}

	// Update image name in deployment yaml, expects placeholder like 'image: <image-name>'
	re := regexp.MustCompile(`image:[ ]*<image-name>`)
	appTemplate = []byte(re.ReplaceAllString(string(appTemplate), fmt.Sprintf("image: %v/%v", containerRegistry, myKubernetesDeployOptions.Image)))

	err = ioutil.WriteFile(myKubernetesDeployOptions.AppTemplate, appTemplate, 0700)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Error when updating appTemplate '%v'", myKubernetesDeployOptions.AppTemplate)
	}

	kubeApplyParams := append(kubeParams, "apply", "--filename", myKubernetesDeployOptions.AppTemplate)
	if len(myKubernetesDeployOptions.AdditionalParameters) > 0 {
		kubeApplyParams = append(kubeApplyParams, myKubernetesDeployOptions.AdditionalParameters...)
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
