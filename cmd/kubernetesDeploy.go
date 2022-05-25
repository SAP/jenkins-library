package cmd

import (
	"fmt"
	"io"

	"github.com/SAP/jenkins-library/pkg/kubernetes"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func kubernetesDeploy(config kubernetesDeployOptions, telemetryData *telemetry.CustomData) {
	kubernetesConfig := kubernetes.KubernetesOptions{
		AdditionalParameters:       config.AdditionalParameters,
		APIServer:                  config.APIServer,
		AppTemplate:                config.AppTemplate,
		ChartPath:                  config.ChartPath,
		ContainerRegistryPassword:  config.ContainerRegistryPassword,
		ContainerImageName:         config.ContainerImageName,
		ContainerImageTag:          config.ContainerImageTag,
		ContainerRegistryURL:       config.ContainerRegistryURL,
		ContainerRegistryUser:      config.ContainerRegistryUser,
		ContainerRegistrySecret:    config.ContainerRegistrySecret,
		CreateDockerRegistrySecret: config.CreateDockerRegistrySecret,
		DeploymentName:             config.DeploymentName,
		DeployTool:                 config.DeployTool,
		ForceUpdates:               config.ForceUpdates,
		HelmDeployWaitSeconds:      config.HelmDeployWaitSeconds,
		HelmValues:                 config.HelmValues,
		ValuesMapping:              config.ValuesMapping,
		Image:                      config.Image,
		ImageNames:                 config.ImageNames,
		ImageNameTags:              config.ImageNameTags,
		ImageDigests:               config.ImageDigests,
		IngressHosts:               config.IngressHosts,
		KeepFailedDeployments:      config.KeepFailedDeployments,
		RunHelmTests:               config.RunHelmTests,
		ShowTestLogs:               config.ShowTestLogs,
		KubeConfig:                 config.KubeConfig,
		KubeContext:                config.KubeContext,
		KubeToken:                  config.KubeToken,
		Namespace:                  config.Namespace,
		TillerNamespace:            config.TillerNamespace,
		DockerConfigJSON:           config.DockerConfigJSON,
		DeployCommand:              config.DeployCommand,
	}

	customTLSCertificateLinks := []string{}
	utils := kubernetes.NewDeployUtilsBundle(customTLSCertificateLinks)

	// error situations stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runKubernetesDeploy(kubernetesConfig, telemetryData, utils, log.Writer())
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runKubernetesDeploy(config kubernetes.KubernetesOptions, telemetryData *telemetry.CustomData, utils kubernetes.DeployUtils, stdout io.Writer) error {
	telemetryData.Custom1Label = "deployTool"
	telemetryData.Custom1 = config.DeployTool

	if config.DeployTool == "helm" || config.DeployTool == "helm3" {
		return kubernetes.RunHelmDeploy(config, utils, stdout)
	} else if config.DeployTool == "kubectl" {
		return kubernetes.RunKubectlDeploy(config, utils, stdout)
	}
	return fmt.Errorf("Failed to execute deployments")
}
