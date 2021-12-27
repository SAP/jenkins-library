package helm

import (
	"fmt"
	"io"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

type Executor interface {
	RunHelmLint(containerRegistry, containerImageName, containerImageTag string, HelmExecuteOptions, stdout io.Writer) error
}

// Execute struct holds utils to enable mocking and common parameters
type Execute struct {
	Utils   Utils
	Options ExecutorOptions
}

type ExecRunner interface {
	SetEnv(env []string)
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	RunExecutable(e string, p ...string) error
}

type ExecutorOptions struct {
	ExecRunner ExecRunner
}

// NewExecutor instantiates Execute struct and sets executeOptions
func NewExecutor(executorOptions ExecutorOptions) Utils {
	utils := utilsBundle{
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

type Utils interface {
	piperutils.FileUtils
	GetExecRunner() ExecRunner
}

// GetExecRunner returns an execRunner if it's not yet initialized
func (u *utilsBundle) GetExecRunner() ExecRunner {
	if u.execRunner == nil {
		u.execRunner = &command.Command{}
		u.execRunner.Stdout(log.Writer())
		u.execRunner.Stderr(log.Writer())
	}
	return u.execRunner
}

type utilsBundle struct {
	*command.Command
	*piperutils.Files
	execRunner ExecRunner
}

type HelmExecuteOptions struct {
	AdditionalParameters      []string `json:"additionalParameters,omitempty"`
	APIServer                 string   `json:"apiServer,omitempty"`
	ChartPath                 string   `json:"chartPath,omitempty"`
	ContainerRegistryPassword string   `json:"containerRegistryPassword,omitempty"`
	ContainerImageName        string   `json:"containerImageName,omitempty"`
	ContainerImageTag         string   `json:"containerImageTag,omitempty"`
	ContainerRegistryURL      string   `json:"containerRegistryUrl,omitempty"`
	ContainerRegistryUser     string   `json:"containerRegistryUser,omitempty"`
	ContainerRegistrySecret   string   `json:"containerRegistrySecret,omitempty"`
	DeploymentName            string   `json:"deploymentName,omitempty"`
	DeployTool                string   `json:"deployTool,omitempty" validate:"possible-values=helm helm3"`
	ForceUpdates              bool     `json:"forceUpdates,omitempty"`
	HelmDeployWaitSeconds     int      `json:"helmDeployWaitSeconds,omitempty"`
	HelmValues                []string `json:"helmValues,omitempty"`
	Image                     string   `json:"image,omitempty"`
	IngressHosts              []string `json:"ingressHosts,omitempty"`
	KeepFailedDeployments     bool     `json:"keepFailedDeployments,omitempty"`
	KubeConfig                string   `json:"kubeConfig,omitempty"`
	KubeContext               string   `json:"kubeContext,omitempty"`
	Namespace                 string   `json:"namespace,omitempty"`
	TillerNamespace           string   `json:"tillerNamespace,omitempty"`
	DockerConfigJSON          string   `json:"dockerConfigJSON,omitempty"`
}

// ToDo RunHelmUpgrade
func RunHelmUpgrade() {

}

// ToDo RunHelmInstall
func RunHelmInstall() {

}

// ToDo RunHelmLint
func (exec *Execute) RunHelmLint(containerRegistry, containerImageName, containerImageTag string, config HelmExecuteOptions, stdout io.Writer) error {
	execRunner := exec.Utils.GetExecRunner()
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

	execRunner.SetEnv(helmEnv)
	execRunner.Stdout(stdout)

	if config.DeployTool == "helm" {
		initParams := []string{"init", "--client-only"}
		if err := execRunner.RunExecutable("helm", initParams...); err != nil {
			log.Entry().WithError(err).Fatal("Helm init call failed")
		}
	}

	// var secretsData string
	// if len(config.DockerConfigJSON) == 0 && (len(config.ContainerRegistryUser) == 0 || len(config.ContainerRegistryPassword) == 0) {
	// 	log.Entry().Info("No/incomplete container registry credentials and no docker config.json file provided: skipping secret creation")
	// 	if len(config.ContainerRegistrySecret) > 0 {
	// 		secretsData = fmt.Sprintf(",imagePullSecrets[0].name=%v", config.ContainerRegistrySecret)
	// 	}
	// } else {
	// 	var dockerRegistrySecret bytes.Buffer
	// 	utils.Stdout(&dockerRegistrySecret)
	// 	kubeSecretParams := helmDefineKubeSecretParams(config, containerRegistry, utils)
	// 	log.Entry().Infof("Calling kubectl create secret --dry-run=true ...")
	// 	log.Entry().Debugf("kubectl parameters %v", kubeSecretParams)
	// 	if err := utils.RunExecutable("kubectl", kubeSecretParams...); err != nil {
	// 		log.Entry().WithError(err).Fatal("Retrieving Docker config via kubectl failed")
	// 	}

	// 	var dockerRegistrySecretData struct {
	// 		Kind string `json:"kind"`
	// 		Data struct {
	// 			DockerConfJSON string `json:".dockerconfigjson"`
	// 		} `json:"data"`
	// 		Type string `json:"type"`
	// 	}
	// 	if err := json.Unmarshal(dockerRegistrySecret.Bytes(), &dockerRegistrySecretData); err != nil {
	// 		log.Entry().WithError(err).Fatal("Reading docker registry secret json failed")
	// 	}
	// 	// make sure that secret is hidden in log output
	// 	log.RegisterSecret(dockerRegistrySecretData.Data.DockerConfJSON)

	// 	log.Entry().Debugf("Secret created: %v", string(dockerRegistrySecret.Bytes()))

	// 	// pass secret in helm default template way and in Piper backward compatible way
	// 	secretsData = fmt.Sprintf(",secret.name=%v,secret.dockerconfigjson=%v,imagePullSecrets[0].name=%v", config.ContainerRegistrySecret, dockerRegistrySecretData.Data.DockerConfJSON, config.ContainerRegistrySecret)
	// }

	// // Deprecated functionality
	// // only for backward compatible handling of ingress.hosts
	// // this requires an adoption of the default ingress.yaml template
	// // Due to the way helm is implemented it is currently not possible to overwrite a part of a list:
	// // see: https://github.com/helm/helm/issues/5711#issuecomment-636177594
	// // Recommended way is to use a custom values file which contains the appropriate data
	// ingressHosts := ""
	// for i, h := range config.IngressHosts {
	// 	ingressHosts += fmt.Sprintf(",ingress.hosts[%v]=%v", i, h)
	// }

	// upgradeParams := []string{
	// 	"upgrade",
	// 	config.DeploymentName,
	// 	config.ChartPath,
	// }

	// for _, v := range config.HelmValues {
	// 	upgradeParams = append(upgradeParams, "--values", v)
	// }

	// upgradeParams = append(
	// 	upgradeParams,
	// 	"--install",
	// 	"--namespace", config.Namespace,
	// 	"--set",
	// 	fmt.Sprintf("image.repository=%v/%v,image.tag=%v%v%v", containerRegistry, containerImageName, containerImageTag, secretsData, ingressHosts),
	// )

	// if config.ForceUpdates {
	// 	upgradeParams = append(upgradeParams, "--force")
	// }

	// if config.DeployTool == "helm" {
	// 	upgradeParams = append(upgradeParams, "--wait", "--timeout", strconv.Itoa(config.HelmDeployWaitSeconds))
	// }

	// if config.DeployTool == "helm3" {
	// 	upgradeParams = append(upgradeParams, "--wait", "--timeout", fmt.Sprintf("%vs", config.HelmDeployWaitSeconds))
	// }

	// if !config.KeepFailedDeployments {
	// 	upgradeParams = append(upgradeParams, "--atomic")
	// }

	// if len(config.KubeContext) > 0 {
	// 	upgradeParams = append(upgradeParams, "--kube-context", config.KubeContext)
	// }

	// if len(config.AdditionalParameters) > 0 {
	// 	upgradeParams = append(upgradeParams, config.AdditionalParameters...)
	// }

	// utils.Stdout(stdout)
	// log.Entry().Info("Calling helm upgrade ...")
	// log.Entry().Debugf("Helm parameters %v", upgradeParams)
	// if err := utils.RunExecutable("helm", upgradeParams...); err != nil {
	// 	log.Entry().WithError(err).Fatal("Helm upgrade call failed")
	// }
	return nil
}

func SplitRegistryURL(registryURL string) (protocol, registry string, err error) {
	parts := strings.Split(registryURL, "://")
	if len(parts) != 2 || len(parts[1]) == 0 {
		return "", "", fmt.Errorf("Failed to split registry url '%v'", registryURL)
	}
	return parts[0], parts[1], nil
}

func SplitFullImageName(image string) (imageName, tag string, err error) {
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

// ToDo RunHelmTest
func RunHelmTest() {

}

// ToDo RunHelmDelete
func RunHelmDelete() {

}

// ToDo RunHelmPackage
func RunHelmPackage() {

}

// func helmDefineKubeSecretParams(config helmExecuteOptions, containerRegistry string, utils helmDeployUtils) []string {
// 	kubeSecretParams := []string{
// 		"create",
// 		"secret",
// 	}
// 	if config.DeployTool == "helm" || config.DeployTool == "helm3" {
// 		kubeSecretParams = append(
// 			kubeSecretParams,
// 			"--insecure-skip-tls-verify=true",
// 			"--dry-run=true",
// 			"--output=json",
// 		)
// 	}

// 	if len(config.DockerConfigJSON) > 0 {
// 		// first enhance config.json with additional pipeline-related credentials if they have been provided
// 		if len(containerRegistry) > 0 && len(config.ContainerRegistryUser) > 0 && len(config.ContainerRegistryPassword) > 0 {
// 			var err error
// 			_, err = piperDocker.CreateDockerConfigJSON(containerRegistry, config.ContainerRegistryUser, config.ContainerRegistryPassword, "", config.DockerConfigJSON, utils)
// 			if err != nil {
// 				log.Entry().Warningf("failed to update Docker config.json: %v", err)
// 			}
// 		}

// 		return append(
// 			kubeSecretParams,
// 			"generic",
// 			config.ContainerRegistrySecret,
// 			fmt.Sprintf("--from-file=.dockerconfigjson=%v", config.DockerConfigJSON),
// 			"--type=kubernetes.io/dockerconfigjson",
// 		)
// 	}
// 	return append(
// 		kubeSecretParams,
// 		"docker-registry",
// 		config.ContainerRegistrySecret,
// 		fmt.Sprintf("--docker-server=%v", containerRegistry),
// 		fmt.Sprintf("--docker-username=%v", config.ContainerRegistryUser),
// 		fmt.Sprintf("--docker-password=%v", config.ContainerRegistryPassword),
// 	)
// }
