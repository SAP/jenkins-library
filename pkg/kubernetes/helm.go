package kubernetes

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/SAP/jenkins-library/pkg/config"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
)

// HelmExecutor is used for mock
type HelmExecutor interface {
	RunHelmUpgrade() error
	RunHelmLint() error
	RunHelmInstall() error
	RunHelmUninstall() error
	RunHelmTest() error
	RunHelmPublish() (string, error)
	RunHelmDependency() error
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
	Version                   string   `json:"version,omitempty"`
	AppVersion                string   `json:"appVersion,omitempty"`
	PublishVersion            string   `json:"publishVersion,omitempty"`
	Dependency                string   `json:"dependency,omitempty" validate:"possible-values=build list update"`
	PackageDependencyUpdate   bool     `json:"packageDependencyUpdate,omitempty"`
	DumpLogs                  bool     `json:"dumpLogs,omitempty"`
	FilterTest                string   `json:"filterTest,omitempty"`
	TargetRepositoryURL       string   `json:"targetRepositoryURL,omitempty"`
	TargetRepositoryName      string   `json:"targetRepositoryName,omitempty"`
	TargetRepositoryUser      string   `json:"targetRepositoryUser,omitempty"`
	TargetRepositoryPassword  string   `json:"targetRepositoryPassword,omitempty"`
	SourceRepositoryURL       string   `json:"sourceRepositoryURL,omitempty"`
	SourceRepositoryName      string   `json:"sourceRepositoryName,omitempty"`
	SourceRepositoryUser      string   `json:"sourceRepositoryUser,omitempty"`
	SourceRepositoryPassword  string   `json:"sourceRepositoryPassword,omitempty"`
	HelmCommand               string   `json:"helmCommand,omitempty"`
	CustomTLSCertificateLinks []string `json:"customTlsCertificateLinks,omitempty"`
	RenderSubchartNotes       bool     `json:"renderSubchartNotes,omitempty"`
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

// runHelmAdd is used to add a chart repository
func (h *HelmExecute) runHelmAdd(name, url, user, password string) error {
	helmParams := []string{
		"repo",
		"add",
	}
	if len(name) == 0 {
		return fmt.Errorf("there is no RepositoryName value. 'helm repo add' command requires 2 arguments")
	}
	if len(url) == 0 {
		return fmt.Errorf("there is no RepositoryURL value. 'helm repo add' command requires 2 arguments")
	}
	if len(user) != 0 {
		helmParams = append(helmParams, "--username", user)
	}
	if len(password) != 0 {
		helmParams = append(helmParams, "--password", password)
	}
	helmParams = append(helmParams, name)
	helmParams = append(helmParams, url)
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
	}

	if len(h.config.ChartPath) == 0 {
		if err := h.runHelmAdd(h.config.TargetRepositoryName, h.config.TargetRepositoryURL, h.config.TargetRepositoryUser, h.config.TargetRepositoryPassword); err != nil {
			return fmt.Errorf("failed to add a chart repository: %v", err)
		}
		helmParams = append(helmParams, h.config.TargetRepositoryName)
	} else {
		helmParams = append(helmParams, h.config.ChartPath)
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

	if len(h.config.KubeContext) > 0 {
		helmParams = append(helmParams, "--kube-context", h.config.KubeContext)
	}

	if !h.config.KeepFailedDeployments {
		helmParams = append(helmParams, "--atomic")
	}

	if h.config.RenderSubchartNotes {
		helmParams = append(helmParams, "--render-subchart-notes")
	}

	if len(h.config.AdditionalParameters) > 0 {
		helmParams = append(helmParams, expandEnv(h.config.AdditionalParameters)...)
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

	for _, v := range h.config.HelmValues {
		helmParams = append(helmParams, "--values", v)
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

	helmParams := []string{
		"install",
		h.config.DeploymentName,
	}

	if len(h.config.ChartPath) == 0 {
		if err := h.runHelmAdd(h.config.TargetRepositoryName, h.config.TargetRepositoryURL, h.config.TargetRepositoryUser, h.config.TargetRepositoryPassword); err != nil {
			return fmt.Errorf("failed to add a chart repository: %v", err)
		}
		helmParams = append(helmParams, h.config.TargetRepositoryName)
	} else {
		helmParams = append(helmParams, h.config.ChartPath)
	}
	helmParams = append(helmParams, "--namespace", h.config.Namespace)
	helmParams = append(helmParams, "--create-namespace")

	if len(h.config.KubeContext) > 0 {
		helmParams = append(helmParams, "--kube-context", h.config.KubeContext)
	}

	if !h.config.KeepFailedDeployments {
		helmParams = append(helmParams, "--atomic")
	}

	helmParams = append(helmParams, "--wait", "--timeout", fmt.Sprintf("%vs", h.config.HelmDeployWaitSeconds))
	for _, v := range h.config.HelmValues {
		helmParams = append(helmParams, "--values", v)
	}

	if h.config.RenderSubchartNotes {
		helmParams = append(helmParams, "--render-subchart-notes")
	}

	if len(h.config.AdditionalParameters) > 0 {
		helmParams = append(helmParams, expandEnv(h.config.AdditionalParameters)...)
	}

	if h.verbose {
		helmParams = append(helmParams, "--debug")

		helmParamsDryRun := append(helmParams, "--dry-run", "--hide-secret")
		if err := h.runHelmCommand(helmParamsDryRun); err != nil {
			log.Entry().WithError(err).Error("Helm install --dry-run --hide-secret call failed")
		}
	}

	if err := h.runHelmCommand(helmParams); err != nil {
		log.Entry().WithError(err).Fatal("Helm install call failed")
	}

	return nil
}

// RunHelmUninstall is used to uninstall a chart
func (h *HelmExecute) RunHelmUninstall() error {
	if err := h.runHelmInit(); err != nil {
		return fmt.Errorf("failed to execute deployments: %v", err)
	}

	helmParams := []string{
		"uninstall",
		h.config.DeploymentName,
	}
	if len(h.config.Namespace) <= 0 {
		return fmt.Errorf("namespace has not been set, please configure namespace parameter")
	}
	if len(h.config.KubeContext) > 0 {
		helmParams = append(helmParams, "--kube-context", h.config.KubeContext)
	}
	helmParams = append(helmParams, "--namespace", h.config.Namespace)
	if h.config.HelmDeployWaitSeconds > 0 {
		helmParams = append(helmParams, "--wait", "--timeout", fmt.Sprintf("%vs", h.config.HelmDeployWaitSeconds))
	}
	if h.verbose {
		helmParams = append(helmParams, "--debug")

		helmParamsDryRun := append(helmParams, "--dry-run")
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
func (h *HelmExecute) runHelmPackage() error {
	if len(h.config.ChartPath) == 0 {
		return fmt.Errorf("there is no ChartPath value. The chartPath value is mandatory")
	}

	err := h.runHelmInit()
	if err != nil {
		return fmt.Errorf("failed to execute deployments: %v", err)
	}

	helmParams := []string{
		"package",
		h.config.ChartPath,
	}
	if len(h.config.Version) > 0 {
		helmParams = append(helmParams, "--version", h.config.Version)
	}
	if h.config.PackageDependencyUpdate {
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

// RunHelmDependency is used to manage a chart's dependencies
func (h *HelmExecute) RunHelmDependency() error {
	if len(h.config.Dependency) == 0 {
		return fmt.Errorf("there is no dependency value. Possible values are build, list, update")
	}

	if len(h.config.SourceRepositoryName) > 0 && len(h.config.SourceRepositoryURL) > 0 {
		if err := h.runHelmAdd(h.config.SourceRepositoryName, h.config.SourceRepositoryURL, h.config.SourceRepositoryUser, h.config.SourceRepositoryPassword); err != nil {
			log.Entry().WithError(err).Fatal("Helm repo call failed")
		}
	}

	helmParams := []string{
		"dependency",
	}

	helmParams = append(helmParams, h.config.Dependency, h.config.ChartPath)

	if len(h.config.AdditionalParameters) > 0 {
		helmParams = append(helmParams, h.config.AdditionalParameters...)
	}

	if err := h.runHelmCommand(helmParams); err != nil {
		log.Entry().WithError(err).Fatal("Helm dependency call failed")
	}

	dependencyDir := filepath.Join(h.config.ChartPath, "charts")
	exists, err := h.utils.DirExists(dependencyDir)
	if err != nil {
		return fmt.Errorf("failed to get directory information: %v", err)
	}

	if exists {
		if err := h.utils.Chmod(dependencyDir, 0777); err != nil {
			return fmt.Errorf("failed to change permissions: %v", err)
		}
	}

	return nil
}

// RunHelmPublish is used to upload a chart to a registry
func (h *HelmExecute) RunHelmPublish() (string, error) {
	err := h.runHelmInit()
	if err != nil {
		return "", fmt.Errorf("failed to execute deployments: %v", err)
	}

	err = h.runHelmPackage()
	if err != nil {
		return "", fmt.Errorf("failed to execute deployments: %v", err)
	}

	if len(h.config.TargetRepositoryURL) == 0 {
		return "", fmt.Errorf("there's no target repository for helm chart publishing configured")
	}

	repoClientOptions := piperhttp.ClientOptions{
		Username:     h.config.TargetRepositoryUser,
		Password:     h.config.TargetRepositoryPassword,
		TrustedCerts: h.config.CustomTLSCertificateLinks,
	}

	h.utils.SetOptions(repoClientOptions)

	binary := fmt.Sprintf("%s-%s.tgz", h.config.DeploymentName, h.config.PublishVersion)

	separator := "/"

	if strings.HasSuffix(h.config.TargetRepositoryURL, "/") {
		separator = ""
	}

	targetURL := fmt.Sprintf("%s%s%s", h.config.TargetRepositoryURL, separator, binary)

	log.Entry().Infof("publishing artifact: %s", targetURL)

	response, err := h.utils.UploadRequest(http.MethodPut, targetURL, binary, "", nil, nil, "binary")
	if err != nil {
		return "", fmt.Errorf("couldn't upload artifact: %w", err)
	}

	if !(response.StatusCode == 200 || response.StatusCode == 201) {
		return "", fmt.Errorf("couldn't upload artifact, received status code %d", response.StatusCode)
	}

	return targetURL, nil
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

// expandEnv replaces ${var} or $var in params according to the values of the current environment variables
func expandEnv(params []string) []string {
	expanded := []string{}

	for _, param := range params {
		if strings.Contains(param, config.VaultCredentialEnvPrefixDefault) {
			expanded = append(expanded, os.ExpandEnv(param))
		} else {
			expanded = append(expanded, param)
		}
	}

	return expanded
}
