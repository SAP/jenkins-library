package kubernetes

import (
	"fmt"
	"io"

	"github.com/SAP/jenkins-library/pkg/log"
)

// HelmExecutor is used for mock
type HelmExecutor interface {
	RunHelmAdd() error
	RunHelmUpgrade() error
	RunHelmLint() error
	RunHelmInstall() error
	RunHelmUninstall() error
	RunHelmPackage() error
	RunHelmTest() error
	RunHelmRegistryLogin() error
	RunHelmRegistryLogout() error
	RunHelmPush() error
}

// HelmExecute struct
type HelmExecute struct {
	utils   DeployUtils
	config  HelmExecuteOptions
	verbose bool
	stdout  io.Writer
}

// HelmExecuteOptions struct holds common parameters for functions RunHelm...
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
	ForceUpdates              bool     `json:"forceUpdates,omitempty"`
	HelmDeployWaitSeconds     int      `json:"helmDeployWaitSeconds,omitempty"`
	HelmValues                []string `json:"helmValues,omitempty"`
	Image                     string   `json:"image,omitempty"`
	KeepFailedDeployments     bool     `json:"keepFailedDeployments,omitempty"`
	KubeConfig                string   `json:"kubeConfig,omitempty"`
	KubeContext               string   `json:"kubeContext,omitempty"`
	Namespace                 string   `json:"namespace,omitempty"`
	DockerConfigJSON          string   `json:"dockerConfigJSON,omitempty"`
	DryRun                    bool     `json:"dryRun,omitempty"`
	PackageVersion            string   `json:"packageVersion,omitempty"`
	AppVersion                string   `json:"appVersion,omitempty"`
	DependencyUpdate          bool     `json:"dependencyUpdate,omitempty"`
	DumpLogs                  bool     `json:"dumpLogs,omitempty"`
	FilterTest                string   `json:"filterTest,omitempty"`
	ChartRepo                 string   `json:"chartRepo,omitempty"`
	HelmRegistryUser          string   `json:"helmRegistryUser,omitempty"`
}

// NewHelmExecutor creates HelmExecute instance
func NewHelmExecutor(config HelmExecuteOptions, utils DeployUtils, verbose bool, stdout io.Writer) HelmExecutor {
	return &HelmExecute{
		config:  config,
		utils:   utils,
		verbose: verbose,
		stdout:  stdout,
	}
}

// runHelmInit is used to set up env for executing helm command
func (h *HelmExecute) runHelmInit() error {
	helmLogFields := map[string]interface{}{}
	helmLogFields["Chart Path"] = h.config.ChartPath
	helmLogFields["Namespace"] = h.config.Namespace
	helmLogFields["Deployment Name"] = h.config.DeploymentName
	helmLogFields["Context"] = h.config.KubeContext
	helmLogFields["Kubeconfig"] = h.config.KubeConfig
	log.Entry().WithFields(helmLogFields).Debug("Calling Helm")

	helmEnv := []string{fmt.Sprintf("KUBECONFIG=%v", h.config.KubeConfig)}

	log.Entry().Debugf("Helm SetEnv: %v", helmEnv)
	h.utils.SetEnv(helmEnv)
	h.utils.Stdout(h.stdout)

	return nil
}

// RunHelmAdd is used to add a chart repository
func (h *HelmExecute) RunHelmAdd() error {
	helmParams := []string{
		"repo",
		"add",
		"stable",
	}

	helmParams = append(helmParams, h.config.ChartRepo)
	if h.verbose {
		helmParams = append(helmParams, "--debug")
	}

	h.utils.Stdout(h.stdout)
	log.Entry().Info("Calling helm add ...")
	log.Entry().Debugf("Helm parameters: %v", helmParams)
	if err := h.utils.RunExecutable("helm", helmParams...); err != nil {
		log.Entry().WithError(err).Fatal("Helm add call failed")
	}

	return nil
}

// RunHelmUpgrade is used to upgrade a release
func (h *HelmExecute) RunHelmUpgrade() error {
	err := h.runHelmInit()
	if err != nil {
		return fmt.Errorf("failed to execute deployments: %v", err)
	}

	// containerInfo, err := getContainerInfo(h.config)
	// if err != nil {
	// 	return fmt.Errorf("failed to execute deployments")
	// }
	// secretsData, err := getSecretsData(h.config, h.utils, containerInfo)
	// if err != nil {
	// 	return fmt.Errorf("failed to execute deployments")
	// }

	helmParams := []string{
		"upgrade",
		h.config.DeploymentName,
		h.config.ChartPath,
	}

	if h.verbose {
		helmParams = append(helmParams, "--debug")
	}

	for _, v := range h.config.HelmValues {
		helmParams = append(helmParams, "--values", v)
	}

	helmParams = append(
		helmParams,
		"--install",
		"--namespace", h.config.Namespace,
		// "--set",
		// fmt.Sprintf("image.repository=%v/%v,image.tag=%v", containerInfo["containerRegistry"], containerInfo["containerImageName"],
		// 	containerInfo["containerImageTag"], secretsData),
	)

	if h.config.ForceUpdates {
		helmParams = append(helmParams, "--force")
	}

	helmParams = append(helmParams, "--wait", "--timeout", fmt.Sprintf("%vs", h.config.HelmDeployWaitSeconds))

	if !h.config.KeepFailedDeployments {
		helmParams = append(helmParams, "--atomic")
	}

	if len(h.config.AdditionalParameters) > 0 {
		helmParams = append(helmParams, h.config.AdditionalParameters...)
	}

	h.utils.Stdout(h.stdout)
	log.Entry().Info("Calling helm upgrade ...")
	log.Entry().Debugf("Helm parameters: %v", helmParams)
	if err := h.utils.RunExecutable("helm", helmParams...); err != nil {
		log.Entry().WithError(err).Fatal("Helm upgrade call failed")
	}

	return nil
}

// RunHelmLint is used to examine a chart for possible issues
func (h *HelmExecute) RunHelmLint() error {
	err := h.runHelmInit()
	if err != nil {
		return fmt.Errorf("failed to execute deployments: %v", err)
	}

	helmParams := []string{
		"lint",
		h.config.ChartPath,
	}

	if h.verbose {
		helmParams = append(helmParams, "--debug")
	}

	h.utils.Stdout(h.stdout)
	log.Entry().Info("Calling helm lint ...")
	log.Entry().Debugf("Helm parameters: %v", helmParams)
	if err := h.utils.RunExecutable("helm", helmParams...); err != nil {
		log.Entry().WithError(err).Fatal("Helm lint call failed")
	}

	return nil
}

// RunHelmInstall is used to install a chart
func (h *HelmExecute) RunHelmInstall() error {
	if err := h.runHelmInit(); err != nil {
		return fmt.Errorf("failed to execute deployments: %v", err)
	}

	if err := h.RunHelmAdd(); err != nil {
		return fmt.Errorf("failed to execute deployments: %v", err)
	}

	helmParams := []string{
		"install",
		h.config.DeploymentName,
		h.config.ChartPath,
	}
	helmParams = append(helmParams, "--namespace", h.config.Namespace)
	helmParams = append(helmParams, "--create-namespace")
	if !h.config.KeepFailedDeployments {
		helmParams = append(helmParams, "--atomic")
	}
	if h.config.DryRun {
		helmParams = append(helmParams, "--dry-run")
	}
	helmParams = append(helmParams, "--wait", "--timeout", fmt.Sprintf("%vs", h.config.HelmDeployWaitSeconds))
	for _, v := range h.config.HelmValues {
		helmParams = append(helmParams, "--values", v)
	}
	if len(h.config.AdditionalParameters) > 0 {
		helmParams = append(helmParams, h.config.AdditionalParameters...)
	}
	if h.verbose {
		helmParams = append(helmParams, "--debug")
	}

	h.utils.Stdout(h.stdout)
	log.Entry().Info("Calling helm install ...")
	log.Entry().Debugf("Helm parameters: %v", helmParams)
	if err := h.utils.RunExecutable("helm", helmParams...); err != nil {
		log.Entry().WithError(err).Fatal("Helm install call failed")
	}

	return nil
}

// RunHelmUninstall is used to uninstall a chart
func (h *HelmExecute) RunHelmUninstall() error {
	err := h.runHelmInit()
	if err != nil {
		return fmt.Errorf("failed to execute deployments: %v", err)
	}

	helmParams := []string{
		"uninstall",
		h.config.DeploymentName,
	}
	if len(h.config.Namespace) <= 0 {
		return fmt.Errorf("namespace has not been set, please configure namespace parameter")
	}
	helmParams = append(helmParams, "--namespace", h.config.Namespace)
	if h.config.HelmDeployWaitSeconds > 0 {
		helmParams = append(helmParams, "--wait", "--timeout", fmt.Sprintf("%vs", h.config.HelmDeployWaitSeconds))
	}
	if h.config.DryRun {
		helmParams = append(helmParams, "--dry-run")
	}
	if h.verbose {
		helmParams = append(helmParams, "--debug")
	}

	h.utils.Stdout(h.stdout)
	log.Entry().Info("Calling helm uninstall ...")
	log.Entry().Debugf("Helm parameters: %v", helmParams)
	if err := h.utils.RunExecutable("helm", helmParams...); err != nil {
		log.Entry().WithError(err).Fatal("Helm uninstall call failed")
	}

	return nil
}

// RunHelmPackage is used to package a chart directory into a chart archive
func (h *HelmExecute) RunHelmPackage() error {
	err := h.runHelmInit()
	if err != nil {
		return fmt.Errorf("failed to execute deployments: %v", err)
	}

	helmParams := []string{
		"package",
		h.config.ChartPath,
	}
	if len(h.config.PackageVersion) > 0 {
		helmParams = append(helmParams, "--version", h.config.PackageVersion)
	}
	if h.config.DependencyUpdate {
		helmParams = append(helmParams, "--dependency-update")
	}
	if len(h.config.AppVersion) > 0 {
		helmParams = append(helmParams, "--app-version", h.config.AppVersion)
	}
	if h.verbose {
		helmParams = append(helmParams, "--debug")
	}

	h.utils.Stdout(h.stdout)
	log.Entry().Info("Calling helm package ...")
	log.Entry().Debugf("Helm parameters: %v", helmParams)
	if err := h.utils.RunExecutable("helm", helmParams...); err != nil {
		log.Entry().WithError(err).Fatal("Helm package call failed")
	}

	return nil
}

// RunHelmTest is used to run tests for a release
func (h *HelmExecute) RunHelmTest() error {
	err := h.runHelmInit()
	if err != nil {
		return fmt.Errorf("failed to execute deployments: %v", err)
	}

	helmParams := []string{
		"test",
		h.config.ChartPath,
	}
	if len(h.config.FilterTest) > 0 {
		helmParams = append(helmParams, "--filter", h.config.FilterTest)
	}
	if h.config.DumpLogs {
		helmParams = append(helmParams, "--logs")
	}
	if h.verbose {
		helmParams = append(helmParams, "--debug")
	}

	h.utils.Stdout(h.stdout)
	log.Entry().Info("Calling helm test ...")
	log.Entry().Debugf("Helm parameters: %v", helmParams)
	if err := h.utils.RunExecutable("helm", helmParams...); err != nil {
		log.Entry().WithError(err).Fatal("Helm test call failed")
	}

	return nil
}

// RunHelmRegistryLogin is used to login private registry
func (h *HelmExecute) RunHelmRegistryLogin() error {
	helmParams := []string{
		"registry login",
	}
	helmParams = append(helmParams, "-u", h.config.HelmRegistryUser)
	helmParams = append(helmParams, "localhost:5000")

	h.utils.Stdout(h.stdout)
	log.Entry().Info("Calling helm login ...")
	log.Entry().Debugf("Helm parameters: %v", helmParams)
	if err := h.utils.RunExecutable("helm", helmParams...); err != nil {
		log.Entry().WithError(err).Fatal("Helm push login failed")
	}

	return nil
}

// RunHelmRegistryLogout is logout to login private registry
func (h *HelmExecute) RunHelmRegistryLogout() error {
	helmParams := []string{
		"registry logout",
	}
	helmParams = append(helmParams, "localhost:5000")

	h.utils.Stdout(h.stdout)
	log.Entry().Info("Calling helm logout ...")
	log.Entry().Debugf("Helm parameters: %v", helmParams)
	if err := h.utils.RunExecutable("helm", helmParams...); err != nil {
		log.Entry().WithError(err).Fatal("Helm push logout failed")
	}

	return nil
}

//RunHelmPush is used to upload a chart to a registry
func (h *HelmExecute) RunHelmPush() error {
	err := h.runHelmInit()
	if err != nil {
		return fmt.Errorf("failed to execute deployments: %v", err)
	}

	if err := h.RunHelmRegistryLogin(); err != nil {
		return fmt.Errorf("failed to execute registry login: %v", err)
	}

	helmParams := []string{
		"push",
	}
	helmParams = append(helmParams, fmt.Sprintf("%v", h.config.DeploymentName+h.config.PackageVersion+".tgz"))
	helmParams = append(helmParams, "oci://localhost:5000/helm-charts")

	h.utils.Stdout(h.stdout)
	log.Entry().Info("Calling helm push ...")
	log.Entry().Debugf("Helm parameters: %v", helmParams)
	if err := h.utils.RunExecutable("helm", helmParams...); err != nil {
		log.Entry().WithError(err).Fatal("Helm push call failed")
	}

	if err := h.RunHelmRegistryLogout(); err != nil {
		return fmt.Errorf("failed to execute registry logout: %v", err)
	}

	return nil
}
