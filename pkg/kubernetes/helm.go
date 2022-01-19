package kubernetes

import (
	"fmt"
	"io"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

type HelmDeployUtils interface {
	SetEnv(env []string)
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	RunExecutable(e string, p ...string) error

	piperutils.FileUtils
}

type HelmExecuteOptions struct {
	AdditionalParameters      []string `json:"additionalParameters,omitempty"`
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
	KeepFailedDeployments     bool     `json:"keepFailedDeployments,omitempty"`
	KubeConfig                string   `json:"kubeConfig,omitempty"`
	KubeContext               string   `json:"kubeContext,omitempty"`
	Namespace                 string   `json:"namespace,omitempty"`
	DockerConfigJSON          string   `json:"dockerConfigJSON,omitempty"`
	DeployCommand             string   `json:"deployCommand,omitempty"`
	DryRun                    bool     `json:"dryRun,omitempty"`
	PackageVersion            string   `json:"packageVersion,omitempty"`
	AppVersion                string   `json:"appVersion,omitempty"`
	DependencyUpdate          bool     `json:"dependencyUpdate,omitempty"`
	DumpLogs                  bool     `json:"dumpLogs,omitempty"`
	FilterTest                string   `json:"filterTest,omitempty"`
}

type deployUtilsBundle struct {
	*command.Command
	*piperutils.Files
}

func NewDeployUtilsBundle() HelmDeployUtils {
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

func runHelmInit(config HelmExecuteOptions, utils HelmDeployUtils, stdout io.Writer) error {

	if len(config.ChartPath) <= 0 {
		return fmt.Errorf("chart path has not been set, please configure chartPath parameter")
	}
	if len(config.DeploymentName) <= 0 {
		return fmt.Errorf("deployment name has not been set, please configure deploymentName parameter")
	}

	helmLogFields := map[string]interface{}{}
	helmLogFields["Chart Path"] = config.ChartPath
	helmLogFields["Namespace"] = config.Namespace
	helmLogFields["Deployment Name"] = config.DeploymentName
	helmLogFields["Context"] = config.KubeContext
	helmLogFields["Kubeconfig"] = config.KubeConfig
	log.Entry().WithFields(helmLogFields).Debug("Calling Helm")

	helmEnv := []string{fmt.Sprintf("KUBECONFIG=%v", config.KubeConfig)}

	log.Entry().Debugf("Helm SetEnv: %v", helmEnv)
	utils.SetEnv(helmEnv)
	utils.Stdout(stdout)

	return nil
}

func RunHelmUpgrade(config HelmExecuteOptions, utils HelmDeployUtils, stdout io.Writer) error {
	err := runHelmInit(config, utils, stdout)
	if err != nil {
		return fmt.Errorf("failed to execute deployments: %v", err)
	}

	containerInfo, err := getContainerInfo(config)
	if err != nil {
		return fmt.Errorf("failed to execute deployments")
	}
	secretsData, err := getSecretsData(config, utils, containerInfo)
	if err != nil {
		return fmt.Errorf("failed to execute deployments")
	}

	helmParams := []string{
		"upgrade",
		config.DeploymentName,
		config.ChartPath,
	}

	for _, v := range config.HelmValues {
		helmParams = append(helmParams, "--values", v)
	}

	helmParams = append(
		helmParams,
		"--install",
		"--namespace", config.Namespace,
		"--set",
		fmt.Sprintf("image.repository=%v/%v,image.tag=%v%v", containerInfo["containerRegistry"], containerInfo["containerImageName"],
			containerInfo["containerImageTag"], secretsData),
	)

	if config.ForceUpdates {
		helmParams = append(helmParams, "--force")
	}

	helmParams = append(helmParams, "--wait", "--timeout", fmt.Sprintf("%vs", config.HelmDeployWaitSeconds))

	if !config.KeepFailedDeployments {
		helmParams = append(helmParams, "--atomic")
	}

	if len(config.AdditionalParameters) > 0 {
		helmParams = append(helmParams, config.AdditionalParameters...)
	}

	utils.Stdout(stdout)
	log.Entry().Info("Calling helm upgrade ...")
	log.Entry().Debugf("Helm parameters: %v", helmParams)
	if err := utils.RunExecutable("helm", helmParams...); err != nil {
		log.Entry().WithError(err).Fatal("Helm upgrade call failed")
	}

	return nil
}

// ToDo RunHelmInstall
func RunHelmLint() {

}

func RunHelmInstall(config HelmExecuteOptions, utils HelmDeployUtils, stdout io.Writer) error {
	err := runHelmInit(config, utils, stdout)
	if err != nil {
		return fmt.Errorf("failed to execute deployments: %v", err)
	}

	helmParams := []string{
		"install",
		config.DeploymentName,
		config.ChartPath,
	}
	helmParams = append(helmParams, "--namespace", config.Namespace)
	helmParams = append(helmParams, "--create-namespace")
	if !config.KeepFailedDeployments {
		helmParams = append(helmParams, "--atomic")
	}
	if config.DryRun {
		helmParams = append(helmParams, "--dry-run")
	}
	helmParams = append(helmParams, "--wait", "--timeout", fmt.Sprintf("%vs", config.HelmDeployWaitSeconds))
	for _, v := range config.HelmValues {
		helmParams = append(helmParams, "--values", v)
	}
	if len(config.AdditionalParameters) > 0 {
		helmParams = append(helmParams, config.AdditionalParameters...)
	}

	utils.Stdout(stdout)
	log.Entry().Info("Calling helm install ...")
	log.Entry().Debugf("Helm parameters: %v", helmParams)
	if err := utils.RunExecutable("helm", helmParams...); err != nil {
		log.Entry().WithError(err).Fatal("Helm install call failed")
	}

	return nil
}

func RunHelmUninstall(config HelmExecuteOptions, utils HelmDeployUtils, stdout io.Writer) error {
	err := runHelmInit(config, utils, stdout)
	if err != nil {
		return fmt.Errorf("failed to execute deployments: %v", err)
	}

	helmParams := []string{
		"uninstall",
		config.DeploymentName,
	}
	helmParams = append(helmParams, "--namespace", config.Namespace)
	helmParams = append(helmParams, "--wait", "--timeout", fmt.Sprintf("%vs", config.HelmDeployWaitSeconds))
	if config.DryRun {
		helmParams = append(helmParams, "--dry-run")
	}

	utils.Stdout(stdout)
	log.Entry().Info("Calling helm uninstall ...")
	log.Entry().Debugf("Helm parameters: %v", helmParams)
	if err := utils.RunExecutable("helm", helmParams...); err != nil {
		log.Entry().WithError(err).Fatal("Helm uninstall call failed")
	}

	return nil
}

func RunHelmPackage(config HelmExecuteOptions, utils HelmDeployUtils, stdout io.Writer) error {
	err := runHelmInit(config, utils, stdout)
	if err != nil {
		return fmt.Errorf("failed to execute deployments: %v", err)
	}

	helmParams := []string{
		"package",
		config.ChartPath,
	}
	if len(config.PackageVersion) > 0 {
		helmParams = append(helmParams, "--version", config.PackageVersion)
	}
	if config.DependencyUpdate {
		helmParams = append(helmParams, "--dependency-update")
	}
	if len(config.AppVersion) > 0 {
		helmParams = append(helmParams, "--app-version", config.AppVersion)
	}

	utils.Stdout(stdout)
	log.Entry().Info("Calling helm package ...")
	log.Entry().Debugf("Helm parameters: %v", helmParams)
	if err := utils.RunExecutable("helm", helmParams...); err != nil {
		log.Entry().WithError(err).Fatal("Helm package call failed")
	}

	return nil
}

// ToDo RunHelmTest
func RunHelmTest(config HelmExecuteOptions, utils HelmDeployUtils, stdout io.Writer) error {
	err := runHelmInit(config, utils, stdout)
	if err != nil {
		return fmt.Errorf("failed to execute deployments: %v", err)
	}

	helmParams := []string{
		"test",
		config.ChartPath,
	}
	if len(config.FilterTest) > 0 {
		helmParams = append(helmParams, "--filter", config.FilterTest)
	}
	if config.DumpLogs {
		helmParams = append(helmParams, "--logs")
	}

	utils.Stdout(stdout)
	log.Entry().Info("Calling helm test ...")
	log.Entry().Debugf("Helm parameters: %v", helmParams)
	if err := utils.RunExecutable("helm", helmParams...); err != nil {
		log.Entry().WithError(err).Fatal("Helm test call failed")
	}

	return nil
}
