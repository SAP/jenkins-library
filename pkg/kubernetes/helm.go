package kubernetes

import (
	"fmt"
	"io"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

// HelmDeployUtils interface
type HelmDeployUtils interface {
	SetEnv(env []string)
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	RunExecutable(e string, p ...string) error

	piperutils.FileUtils
}

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
	RunHelmDirect() error
}

// HelmExecute struct
type HelmExecute struct {
	utils   HelmDeployUtils
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

// deployUtilsBundle struct  for utils
type deployUtilsBundle struct {
	*command.Command
	*piperutils.Files
}

// NewExecutor
func NewHelmExecutor(config HelmExecuteOptions, utils HelmDeployUtils, verbose bool, stdout io.Writer) HelmExecutor {
	return &HelmExecute{
		config:  config,
		utils:   utils,
		verbose: verbose,
		stdout:  stdout,
	}
}

// NewDeployUtilsBundle initialize using deployUtilsBundle struct
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

	containerInfo, err := getContainerInfo(h.config)
	if err != nil {
		return fmt.Errorf("failed to execute deployments")
	}
	secretsData, err := getSecretsData(h.config, h.utils, containerInfo)
	if err != nil {
		return fmt.Errorf("failed to execute deployments")
	}

	helmParams := []string{
		"upgrade",
		h.config.DeploymentName,
		h.config.ChartPath,
	}

	for _, v := range h.config.HelmValues {
		helmParams = append(helmParams, "--values", v)
	}

	helmParams = append(
		helmParams,
		"--install",
		"--namespace", h.config.Namespace,
		"--set",
		fmt.Sprintf("image.repository=%v/%v,image.tag=%v%v", containerInfo["containerRegistry"], containerInfo["containerImageName"],
			containerInfo["containerImageTag"], secretsData),
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

//RunHelmDirect is used to run helm command directly via flag
func (h *HelmExecute) RunHelmDirect() error {
	err := h.runHelmInit()
	if err != nil {
		return fmt.Errorf("failed to execute helm command: %v", err)
	}

	if err := h.RunHelmAdd(); err != nil {
		return fmt.Errorf("failed to execute helm command: %v", err)
	}

	helmParams := []string{}

	if len(h.config.AdditionalParameters) > 0 {
		helmParams = append(helmParams, h.config.AdditionalParameters...)
	} else {
		return fmt.Errorf("helm command is not presented via flag or config yaml")
	}

	h.utils.Stdout(h.stdout)
	log.Entry().Info("Calling helm command ...")
	log.Entry().Debugf("Helm parameters: %v", helmParams)
	if err := h.utils.RunExecutable("helm", helmParams...); err != nil {
		log.Entry().WithError(err).Fatal("Helm command call failed")
	}

	return nil
}
