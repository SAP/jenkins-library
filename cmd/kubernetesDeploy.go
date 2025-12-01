package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/SAP/jenkins-library/pkg/docker"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/kubernetes"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/cli/values"
)

func kubernetesDeploy(config kubernetesDeployOptions, telemetryData *telemetry.CustomData) {
	customTLSCertificateLinks := []string{}
	utils := kubernetes.NewDeployUtilsBundle(customTLSCertificateLinks)

	// error situations stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runKubernetesDeploy(config, telemetryData, utils, log.Writer())
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runKubernetesDeploy(config kubernetesDeployOptions, telemetryData *telemetry.CustomData, utils kubernetes.DeployUtils, stdout io.Writer) error {
	telemetryData.DeployTool = config.DeployTool

	if config.DeployTool == "helm" || config.DeployTool == "helm3" {
		err := runHelmDeploy(config, utils, stdout)
		// download and execute teardown script
		if len(config.TeardownScript) > 0 {
			log.Entry().Debugf("start running teardownScript script %v", config.TeardownScript)
			if scriptErr := downloadAndExecuteExtensionScript(config.TeardownScript, config.GithubToken, utils); scriptErr != nil {
				if err != nil {
					err = fmt.Errorf("failed to download/run teardownScript script: %v: %w", fmt.Sprint(scriptErr), err)
				} else {
					err = scriptErr
				}
			}
			log.Entry().Debugf("finished running teardownScript script %v", config.TeardownScript)
		}
		return err
	} else if config.DeployTool == "kubectl" {
		return runKubectlDeploy(config, utils, stdout)
	}
	return fmt.Errorf("Failed to execute deployments")
}

func runHelmDeploy(config kubernetesDeployOptions, utils kubernetes.DeployUtils, stdout io.Writer) error {
	if len(config.ChartPath) <= 0 {
		return fmt.Errorf("chart path has not been set, please configure chartPath parameter")
	}
	if len(config.DeploymentName) <= 0 {
		return fmt.Errorf("deployment name has not been set, please configure deploymentName parameter")
	}

	// download and execute setup script
	if len(config.SetupScript) > 0 {
		log.Entry().Debugf("start running setup script %v", config.SetupScript)
		if err := downloadAndExecuteExtensionScript(config.SetupScript, config.GithubToken, utils); err != nil {
			return fmt.Errorf("failed to download/run setup setup script: %w", err)
		}
		log.Entry().Debugf("finished running setup script %v", config.SetupScript)
	}

	_, containerRegistry, err := splitRegistryURL(config.ContainerRegistryURL)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Container registry url '%v' incorrect", config.ContainerRegistryURL)
	}

	helmValues, err := defineDeploymentValues(config, containerRegistry)
	if err != nil {
		return errors.Wrap(err, "failed to process deployment values")
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
	utils.SetEnv(helmEnv)
	utils.Stdout(stdout)

	if config.DeployTool == "helm" {
		initParams := []string{"init", "--client-only"}
		if err := utils.RunExecutable("helm", initParams...); err != nil {
			log.Entry().WithError(err).Fatal("Helm init call failed")
		}
	}

	if len(config.ContainerRegistryUser) == 0 && len(config.ContainerRegistryPassword) == 0 {
		log.Entry().Info("No/incomplete container registry credentials provided: skipping secret creation")
		if len(config.ContainerRegistrySecret) > 0 {
			log.Entry().Debugf("Using existing container registry secret: %v", config.ContainerRegistrySecret)
			helmValues.add("imagePullSecrets[0].name", config.ContainerRegistrySecret)
		// } else {
		// 	log.Entry().Debugf("Using Docker config.json file at '%v' to create kubernetes secret", config.DockerConfigJSON)
		// 	err, kubeSecretParams := defineKubeSecretParams(config, containerRegistry, utils)
		// 	// show dockerConfig contents in debug log (without secrets)
		// 	dockerConfigContent, err := utils.FileRead(config.DockerConfigJSON)
		// 	if err != nil {
		// 		log.Entry().Warningf("failed to read Docker config.json: %v", err)
		// 		return err, []string{}
		// 	}
		// 	log.Entry().Debugf("Using Docker config.json content: %v", string(dockerConfigContent))
		// }
		} else {
			var dockerRegistrySecret bytes.Buffer
			utils.Stdout(&dockerRegistrySecret)
			config.InsecureSkipTLSVerify = true // Currently CA certificate handling is not supported for helm deployments
			err, kubeSecretParams := defineKubeSecretParams(config, containerRegistry, utils)
			if err != nil {
				log.Entry().WithError(err).Fatal("parameter definition for creating registry secret failed")
			}
			log.Entry().Infof("Calling kubectl create secret --dry-run=true ...")
			// print dockerconfigjson file contents in debug log (without secrets)
			log.Entry().Infof("Using Docker config.json content: %v", dockerRegistrySecret.String())
			log.Entry().Infofs("kubectl parameters %v", kubeSecretParams)
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
			helmValues.add("secret.name", config.ContainerRegistrySecret)
			helmValues.add("secret.dockerconfigjson", dockerRegistrySecretData.Data.DockerConfJSON)
			helmValues.add("imagePullSecrets[0].name", config.ContainerRegistrySecret)
		}
	}

	// Deprecated functionality
	// only for backward compatible handling of ingress.hosts
	// this requires an adoption of the default ingress.yaml template
	// Due to the way helm is implemented it is currently not possible to overwrite a part of a list:
	// see: https://github.com/helm/helm/issues/5711#issuecomment-636177594
	// Recommended way is to use a custom values file which contains the appropriate data
	for i, h := range config.IngressHosts {
		helmValues.add(fmt.Sprintf("ingress.hosts[%v]", i), h)
	}

	upgradeParams := []string{
		"upgrade",
		config.DeploymentName,
		config.ChartPath,
	}

	for _, v := range config.HelmValues {
		upgradeParams = append(upgradeParams, "--values", v)
	}

	err = helmValues.mapValues()
	if err != nil {
		return errors.Wrap(err, "failed to map values using 'valuesMapping' configuration")
	}

	upgradeParams = append(
		upgradeParams,
		"--install",
		"--namespace", config.Namespace,
		"--set", strings.Join(helmValues.marshal(), ","),
	)

	if config.ForceUpdates {
		upgradeParams = append(upgradeParams, "--force")
	}

	if config.DeployTool == "helm" {
		upgradeParams = append(upgradeParams, "--wait", "--timeout", strconv.Itoa(config.HelmDeployWaitSeconds))
	}

	if config.DeployTool == "helm3" {
		upgradeParams = append(upgradeParams, "--wait", "--timeout", fmt.Sprintf("%vs", config.HelmDeployWaitSeconds))
	}

	if !config.KeepFailedDeployments {
		upgradeParams = append(upgradeParams, "--atomic")
	}

	if len(config.KubeContext) > 0 {
		upgradeParams = append(upgradeParams, "--kube-context", config.KubeContext)
	}

	if config.RenderSubchartNotes {
		upgradeParams = append(upgradeParams, "--render-subchart-notes")
	}

	if len(config.AdditionalParameters) > 0 {
		upgradeParams = append(upgradeParams, config.AdditionalParameters...)
	}

	utils.Stdout(stdout)
	log.Entry().Info("Calling helm upgrade ...")
	log.Entry().Debugf("Helm parameters %v", upgradeParams)
	if err := utils.RunExecutable("helm", upgradeParams...); err != nil {
		log.Entry().WithError(err).Fatal("Helm upgrade call failed")
	}

	// download and execute verification script
	if len(config.VerificationScript) > 0 {
		log.Entry().Debugf("start running verification script %v", config.VerificationScript)
		if err := downloadAndExecuteExtensionScript(config.VerificationScript, config.GithubToken, utils); err != nil {
			return fmt.Errorf("failed to download/run verification script: %w", err)
		}
		log.Entry().Debugf("finished running verification script %v", config.VerificationScript)
	}

	testParams := []string{
		"test",
		config.DeploymentName,
		"--namespace", config.Namespace,
	}

	if len(config.KubeContext) > 0 {
		testParams = append(testParams, "--kube-context", config.KubeContext)
	}

	if config.DeployTool == "helm" {
		testParams = append(testParams, "--timeout", strconv.Itoa(config.HelmTestWaitSeconds))
	}

	if config.DeployTool == "helm3" {
		testParams = append(testParams, "--timeout", fmt.Sprintf("%vs", config.HelmTestWaitSeconds))
	}

	if config.ShowTestLogs {
		testParams = append(
			testParams,
			"--logs",
		)
	}

	if config.RunHelmTests {
		if err := utils.RunExecutable("helm", testParams...); err != nil {
			log.Entry().WithError(err).Fatal("Helm test call failed")
		}
	}

	return nil
}

func runKubectlDeploy(config kubernetesDeployOptions, utils kubernetes.DeployUtils, stdout io.Writer) error {
	_, containerRegistry, err := splitRegistryURL(config.ContainerRegistryURL)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Container registry url '%v' incorrect", config.ContainerRegistryURL)
	}

	kubeParams := []string{
		fmt.Sprintf("--namespace=%v", config.Namespace),
	}

	log.Entry().Debugf("Running kubectl with InsecureSkipTLSVerify: %v", config.InsecureSkipTLSVerify)

	// Add CA certificate if provided
	if len(config.CACertificate) > 0 && !config.InsecureSkipTLSVerify {
		kubeParams = append(kubeParams, fmt.Sprintf("--certificate-authority=%v", config.CACertificate))
		log.Entry().Debugf("Running kubectl with CACertificate: %v", config.CACertificate)
	}

	kubeParams = append(kubeParams, "--insecure-skip-tls-verify="+strconv.FormatBool(config.InsecureSkipTLSVerify))
	if config.InsecureSkipTLSVerify {
		log.Entry().Warn("Skipping TLS verification check. Please note that this action poses security concerns.")
	}

	if len(config.KubeConfig) > 0 {
		log.Entry().Info("Using KUBECONFIG environment for authentication.")
		kubeEnv := []string{fmt.Sprintf("KUBECONFIG=%v", config.KubeConfig)}
		utils.SetEnv(kubeEnv)
		if len(config.KubeContext) > 0 {
			kubeParams = append(kubeParams, fmt.Sprintf("--context=%v", config.KubeContext))
		}

	} else {
		log.Entry().Info("Using --token parameter for authentication.")
		kubeParams = append(kubeParams, fmt.Sprintf("--server=%v", config.APIServer))
		kubeParams = append(kubeParams, fmt.Sprintf("--token=%v", config.KubeToken))
	}

	utils.Stdout(stdout)

	if len(config.ContainerRegistryUser) == 0 && len(config.ContainerRegistryPassword) == 0 {
		log.Entry().Info("No/incomplete container registry credentials provided: skipping secret creation")
	} else {
		err, kubeSecretParams := defineKubeSecretParams(config, containerRegistry, utils)
		if err != nil {
			log.Entry().WithError(err).Fatal("parameter definition for creating registry secret failed")
		}
		var dockerRegistrySecret bytes.Buffer
		utils.Stdout(&dockerRegistrySecret)
		log.Entry().Infof("Creating container registry secret '%v'", config.ContainerRegistrySecret)
		kubeSecretParams = append(kubeSecretParams, kubeParams...)
		log.Entry().Debugf("Running kubectl with following parameters: %v", kubeSecretParams)
		if err := utils.RunExecutable("kubectl", kubeSecretParams...); err != nil {
			log.Entry().WithError(err).Fatal("Creating container registry secret failed")
		}

		var dockerRegistrySecretData map[string]interface{}

		if err := json.Unmarshal(dockerRegistrySecret.Bytes(), &dockerRegistrySecretData); err != nil {
			log.Entry().WithError(err).Fatal("Reading docker registry secret json failed")
		}

		// write the json output to a file
		tmpFolder := getTempDirForKubeCtlJSON()
		defer os.RemoveAll(tmpFolder) // clean up
		jsonData, _ := json.Marshal(dockerRegistrySecretData)
		if err := os.WriteFile(filepath.Join(tmpFolder, "secret.json"), jsonData, 0777); err != nil {
			log.Entry().WithError(err).Warning("failed to write secret")
		}

		kubeSecretApplyParams := []string{"apply", "-f", filepath.Join(tmpFolder, "secret.json")}
		if err := utils.RunExecutable("kubectl", kubeSecretApplyParams...); err != nil {
			log.Entry().WithError(err).Fatal("Creating container registry secret failed")
		}

	}

	appTemplate, err := utils.FileRead(config.AppTemplate)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Error when reading appTemplate '%v'", config.AppTemplate)
	}

	values, err := defineDeploymentValues(config, containerRegistry)
	if err != nil {
		return errors.Wrap(err, "failed to process deployment values")
	}
	err = values.mapValues()
	if err != nil {
		return errors.Wrap(err, "failed to map values using 'valuesMapping' configuration")
	}

	re := regexp.MustCompile(`image:[ ]*<image-name>`)
	placeholderFound := re.Match(appTemplate)

	if placeholderFound {
		log.Entry().Warn("image placeholder '<image-name>' is deprecated and does not support multi-image replacement, please use Helm-like template syntax '{{ .Values.image.[image-name].reposotory }}:{{ .Values.image.[image-name].tag }}")
		if values.singleImage {
			// Update image name in deployment yaml, expects placeholder like 'image: <image-name>'
			appTemplate = []byte(re.ReplaceAllString(string(appTemplate), fmt.Sprintf("image: %s:%s", values.get("image.repository"), values.get("image.tag"))))
		} else {
			return fmt.Errorf("multi-image replacement not supported for single image placeholder")
		}
	}

	buf := bytes.NewBufferString("")
	tpl, err := template.New("appTemplate").Parse(string(appTemplate))
	if err != nil {
		return errors.Wrap(err, "failed to parse app-template file")
	}
	err = tpl.Execute(buf, values.asHelmValues())
	if err != nil {
		return errors.Wrap(err, "failed to render app-template file")
	}

	err = utils.FileWrite(config.AppTemplate, buf.Bytes(), 0700)
	if err != nil {
		return errors.Wrapf(err, "Error when updating appTemplate '%v'", config.AppTemplate)
	}

	kubeParams = append(kubeParams, config.DeployCommand, "--filename", config.AppTemplate)
	if config.ForceUpdates && config.DeployCommand == "replace" {
		kubeParams = append(kubeParams, "--force")
	}

	if len(config.AdditionalParameters) > 0 {
		kubeParams = append(kubeParams, config.AdditionalParameters...)
	}
	if err := utils.RunExecutable("kubectl", kubeParams...); err != nil {
		log.Entry().Debugf("Running kubectl with following parameters: %v", kubeParams)
		log.Entry().WithError(err).Fatal("Deployment with kubectl failed.")
	}
	return nil
}

type deploymentValues struct {
	mapping     map[string]interface{}
	singleImage bool
	values      []struct {
		key, value string
	}
}

func (dv *deploymentValues) add(key, value string) {
	dv.values = append(dv.values, struct {
		key   string
		value string
	}{
		key:   key,
		value: value,
	})
}

func (dv deploymentValues) get(key string) string {
	for _, item := range dv.values {
		if item.key == key {
			return item.value
		}
	}

	return ""
}

func (dv *deploymentValues) mapValues() error {
	var keys []string
	for k := range dv.mapping {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, dst := range keys {
		srcString, ok := dv.mapping[dst].(string)
		if !ok {
			return fmt.Errorf("invalid path '%#v' is used for valuesMapping, only strings are supported", dv.mapping[dst])
		}
		if val := dv.get(srcString); val != "" {
			dv.add(dst, val)
		} else {
			escapedSrcString := strings.ReplaceAll(srcString, "-", "_")
			log.Entry().Debugf("property '%s' not found, trying with escaped version '%s'", srcString, escapedSrcString)
			if val := dv.get(escapedSrcString); val != "" {
				dv.add(dst, val)
			} else {
				return fmt.Errorf("can not map '%s: %s', %s is not set", dst, srcString, srcString)
			}
		}
	}

	return nil
}

func (dv deploymentValues) marshal() []string {
	var result []string
	for _, item := range dv.values {
		result = append(result, fmt.Sprintf("%s=%s", item.key, item.value))
	}
	return result
}

func (dv *deploymentValues) asHelmValues() map[string]interface{} {
	valuesOpts := values.Options{
		Values: dv.marshal(),
	}
	mergedValues, err := valuesOpts.MergeValues(nil)
	if err != nil {
		log.Entry().WithError(err).Fatal("failed to process deployment values")
	}
	return map[string]interface{}{
		"Values": mergedValues,
	}
}

func createKey(replacer *strings.Replacer, parts ...string) string {
	escapedParts := make([]string, 0, len(parts))

	for _, part := range parts {
		escapedParts = append(escapedParts, replacer.Replace(part))
	}
	return strings.Join(escapedParts, ".")
}

func createGoKey(parts ...string) string {
	return createKey(strings.NewReplacer(".", "_", "-", "_"), parts...)
}

func createHelmKey(parts ...string) string {
	return createKey(strings.NewReplacer(".", "_"), parts...)
}

func getTempDirForKubeCtlJSON() string {
	tmpFolder, err := os.MkdirTemp(".", "temp-")
	if err != nil {
		log.Entry().WithError(err).WithField("path", tmpFolder).Debug("creating temp directory failed")
	}
	return tmpFolder
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

func defineKubeSecretParams(config kubernetesDeployOptions, containerRegistry string, utils kubernetes.DeployUtils) (error, []string) {
	targetPath := ""
	if len(config.DockerConfigJSON) > 0 {
		// first enhance config.json with additional pipeline-related credentials if they have been provided
		if len(containerRegistry) > 0 && len(config.ContainerRegistryUser) > 0 && len(config.ContainerRegistryPassword) > 0 {
			var err error
			targetPath, err = docker.CreateDockerConfigJSON(containerRegistry, config.ContainerRegistryUser, config.ContainerRegistryPassword, "", config.DockerConfigJSON, utils)
			if err != nil {
				log.Entry().Warningf("failed to update Docker config.json: %v", err)
				return err, []string{}
			}
		} else {
			log.Entry().Debugf("Using Docker config.json file at '%v' to create kubernetes secret", config.DockerConfigJSON)
			targetPath = config.DockerConfigJSON
			// show dockerConfig contents in debug log (without secrets)
			dockerConfigContent, err := utils.FileRead(config.DockerConfigJSON)
			if err != nil {
				log.Entry().Warningf("failed to read Docker config.json: %v", err)
				return err, []string{}
			}
			log.Entry().Debugf("Using Docker config.json content: %v", string(dockerConfigContent))
		}
	} else {
		return fmt.Errorf("no docker config json file found to update credentials '%v'", config.DockerConfigJSON), []string{}
	}
	return nil, []string{
		"create",
		"secret",
		"generic",
		config.ContainerRegistrySecret,
		fmt.Sprintf("--from-file=.dockerconfigjson=%v", targetPath),
		"--type=kubernetes.io/dockerconfigjson",
		"--insecure-skip-tls-verify=" + strconv.FormatBool(config.InsecureSkipTLSVerify),
		"--dry-run=client",
		"--output=json",
	}
}

func defineDeploymentValues(config kubernetesDeployOptions, containerRegistry string) (*deploymentValues, error) {
	var err error
	var useDigests bool
	dv := &deploymentValues{
		mapping: config.ValuesMapping,
	}
	if len(config.ImageNames) > 0 {
		if len(config.ImageNames) != len(config.ImageNameTags) {
			log.SetErrorCategory(log.ErrorConfiguration)
			return nil, fmt.Errorf("number of imageNames and imageNameTags must be equal")
		}
		if len(config.ImageDigests) > 0 {
			if len(config.ImageDigests) != len(config.ImageNameTags) {
				log.SetErrorCategory(log.ErrorConfiguration)
				return nil, fmt.Errorf("number of imageDigests and imageNameTags must be equal")
			}

			useDigests = true
		}
		for i, key := range config.ImageNames {
			name, tag, err := splitFullImageName(config.ImageNameTags[i])
			if err != nil {
				log.Entry().WithError(err).Fatalf("Container image '%v' incorrect", config.ImageNameTags[i])
			}

			if useDigests {
				tag = fmt.Sprintf("%s@%s", tag, config.ImageDigests[i])
			}

			dv.add(createGoKey("image", key, "repository"), fmt.Sprintf("%v/%v", containerRegistry, name))
			dv.add(createGoKey("image", key, "tag"), tag)
			// usable for subcharts:
			dv.add(createGoKey(key, "image", "repository"), fmt.Sprintf("%v/%v", containerRegistry, name))
			dv.add(createGoKey(key, "image", "tag"), tag)
			dv.add(createHelmKey(key, "image", "repository"), fmt.Sprintf("%v/%v", containerRegistry, name))
			dv.add(createHelmKey(key, "image", "tag"), tag)

			if len(config.ImageNames) == 1 {
				dv.singleImage = true
				dv.add("image.repository", fmt.Sprintf("%v/%v", containerRegistry, name))
				dv.add("image.tag", tag)
			}
		}
	} else {
		// support either image or containerImageName and containerImageTag
		containerImageName := ""
		containerImageTag := ""
		dv.singleImage = true

		if len(config.Image) > 0 {
			containerImageName, containerImageTag, err = splitFullImageName(config.Image)
			if err != nil {
				log.Entry().WithError(err).Fatalf("Container image '%v' incorrect", config.Image)
			}
		} else if len(config.ContainerImageName) > 0 && len(config.ContainerImageTag) > 0 {
			containerImageName = config.ContainerImageName
			containerImageTag = config.ContainerImageTag
		} else {
			return nil, fmt.Errorf("image information not given - please either set image or containerImageName and containerImageTag")
		}
		dv.add("image.repository", fmt.Sprintf("%v/%v", containerRegistry, containerImageName))
		dv.add("image.tag", containerImageTag)

		dv.add(createGoKey("image", containerImageName, "repository"), fmt.Sprintf("%v/%v", containerRegistry, containerImageName))
		dv.add(createGoKey("image", containerImageName, "tag"), containerImageTag)
	}

	return dv, nil
}

func downloadAndExecuteExtensionScript(script, githubToken string, utils kubernetes.DeployUtils) error {
	setupScript, err := piperhttp.DownloadExecutable(githubToken, utils, utils, script)
	if err != nil {
		return fmt.Errorf("failed to download script %v: %w", script, err)
	}
	err = utils.RunExecutable(setupScript)
	if err != nil {
		return fmt.Errorf("failed to execute script %v: %w", script, err)
	}
	return nil
}
