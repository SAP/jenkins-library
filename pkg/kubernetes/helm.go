package kubernetes

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
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
	RunHelmPublish() error
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
	PackageVersion            string   `json:"packageVersion,omitempty"`
	AppVersion                string   `json:"appVersion,omitempty"`
	DependencyUpdate          bool     `json:"dependencyUpdate,omitempty"`
	DumpLogs                  bool     `json:"dumpLogs,omitempty"`
	FilterTest                string   `json:"filterTest,omitempty"`
	TargetRepositoryURL       string   `json:"targetRepositoryURL,omitempty"`
	TargetRepositoryName      string   `json:"targetRepositoryName,omitempty"`
	TargetRepositoryUser      string   `json:"targetRepositoryUser,omitempty"`
	TargetRepositoryPassword  string   `json:"targetRepositoryPassword,omitempty"`
	HelmCommand               string   `json:"helmCommand,omitempty"`
	CustomTLSCertificateLinks []string `json:"customTlsCertificateLinks,omitempty"`
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
	}
	helmParams = append(helmParams, h.config.TargetRepositoryName)
	helmParams = append(helmParams, h.config.TargetRepositoryURL)
	if h.verbose {
		helmParams = append(helmParams, "--debug")
	}

	if err := h.runHelmCommand(helmParams); err != nil {
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

	if err := h.runHelmCommand(helmParams); err != nil {
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

	if h.verbose {
		helmParamsDryRun := helmParams
		helmParamsDryRun = append(helmParamsDryRun, "--dry-run")
		if err := h.runHelmCommand(helmParamsDryRun); err != nil {
			log.Entry().WithError(err).Error("Helm install --dry-run call failed")
		}
	}

	if err := h.runHelmCommand(helmParams); err != nil {
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
	if h.verbose {
		helmParams = append(helmParams, "--debug")
	}

	if h.verbose {
		helmParamsDryRun := helmParams
		helmParamsDryRun = append(helmParamsDryRun, "--dry-run")
		if err := h.runHelmCommand(helmParamsDryRun); err != nil {
			log.Entry().WithError(err).Error("Helm uninstall --dry-run call failed")
		}
	}

	if err := h.runHelmCommand(helmParams); err != nil {
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

	if err := h.runHelmCommand(helmParams); err != nil {
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

	if err := h.runHelmCommand(helmParams); err != nil {
		log.Entry().WithError(err).Fatal("Helm test call failed")
	}

	return nil
}

//RunHelmPublish is used to upload a chart to a registry
func (h *HelmExecute) RunHelmPublish() error {
	err := h.runHelmInit()
	if err != nil {
		return fmt.Errorf("failed to execute deployments: %v", err)
	}

	if len(h.config.TargetRepositoryURL) == 0 {
		return fmt.Errorf("there's no target repository for helm chart publishing configured")
	}

	repoClientOptions := piperhttp.ClientOptions{
		Username:     h.config.TargetRepositoryUser,
		Password:     h.config.TargetRepositoryPassword,
		TrustedCerts: h.config.CustomTLSCertificateLinks,
	}

	h.utils.SetOptions(repoClientOptions)

	binary := fmt.Sprintf("%v", h.config.DeploymentName+"-"+h.config.PackageVersion+".tgz")

	targetPath := fmt.Sprintf("%v/%s", h.config.DeploymentName, binary)

	separator := "/"

	if strings.HasSuffix(h.config.TargetRepositoryURL, "/") {
		separator = ""
	}

	targetURL := fmt.Sprintf("%s%s%s", h.config.TargetRepositoryURL, separator, targetPath)

	log.Entry().Infof("publishing artifact: %s", targetURL)

	response, err := h.utils.UploadRequest(http.MethodPut, targetURL, binary, "", nil, nil, "binary")
	if err != nil {
		return fmt.Errorf("couldn't upload artifact: %w", err)
	}

	if !(response.StatusCode == 200 || response.StatusCode == 201) {
		return fmt.Errorf("couldn't upload artifact, received status code %d", response.StatusCode)
	}

	return nil
}

func (h *HelmExecute) runHelmCommand(helmParams []string) error {

	h.utils.Stdout(h.stdout)
	log.Entry().Infof("Calling helm %v ...", h.config.HelmCommand)
	log.Entry().Debugf("Helm parameters: %v", helmParams)
	if err := h.utils.RunExecutable("helm", helmParams...); err != nil {
		log.Entry().WithError(err).Fatalf("Helm %v call failed", h.config.HelmCommand)
		return err
	}

	return nil
}
