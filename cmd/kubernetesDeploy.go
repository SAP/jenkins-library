package cmd

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/kubernetes"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func kubernetesDeploy(config kubernetesDeployOptions, telemetryData *telemetry.CustomData) {
	kubernetesConfig := kubernetes.KubernetesOptions{
		ExecOpts: kubernetes.ExecuteOptions{
			AdditionalParameters:      config.AdditionalParameters,
			AppTemplate:               config.AppTemplate,
			ChartPath:                 config.ChartPath,
			ContainerRegistryPassword: config.ContainerRegistryPassword,
			ContainerImageName:        config.ContainerImageName,
			ContainerImageTag:         config.ContainerImageTag,
			ContainerRegistryURL:      config.ContainerRegistryURL,
			ContainerRegistryUser:     config.ContainerRegistryUser,
			ContainerRegistrySecret:   config.ContainerRegistrySecret,
			DeploymentName:            config.DeploymentName,
			ForceUpdates:              config.ForceUpdates,
			HelmDeployWaitSeconds:     config.HelmDeployWaitSeconds,
			HelmValues:                config.HelmValues,
			ValuesMapping:             config.ValuesMapping,
			Image:                     config.Image,
			ImageNames:                config.ImageNames,
			ImageNameTags:             config.ImageNameTags,
			ImageDigests:              config.ImageDigests,
			KeepFailedDeployments:     config.KeepFailedDeployments,
			KubeConfig:                config.KubeConfig,
			KubeContext:               config.KubeContext,
			Namespace:                 config.Namespace,
			DockerConfigJSON:          config.DockerConfigJSON,
		},
		APIServer:                  config.APIServer,
		CreateDockerRegistrySecret: config.CreateDockerRegistrySecret,
		DeployTool:                 config.DeployTool,
		IngressHosts:               config.IngressHosts,
		RunHelmTests:               config.RunHelmTests,
		ShowTestLogs:               config.ShowTestLogs,
		KubeToken:                  config.KubeToken,
		TillerNamespace:            config.TillerNamespace,
		DeployCommand:              config.DeployCommand,
	}

	customTLSCertificateLinks := []string{}
	utils := kubernetes.NewDeployUtilsBundle(customTLSCertificateLinks)

	kubernetesDeploy := kubernetes.NewKubernetesDeploy(kubernetesConfig, utils, GeneralConfig.Verbose, log.Writer())

	// error situations stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runKubernetesDeploy(kubernetesConfig, telemetryData, kubernetesDeploy)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runKubernetesDeploy(config kubernetes.KubernetesOptions, telemetryData *telemetry.CustomData, kubernetesDeploy kubernetes.KubernetesDeploy) error {
	telemetryData.Custom1Label = "deployTool"
	telemetryData.Custom1 = config.DeployTool

	if config.DeployTool == "helm" || config.DeployTool == "helm3" {
		return kubernetesDeploy.RunHelmDeploy()
	} else if config.DeployTool == "kubectl" {
		return kubernetesDeploy.RunKubectlDeploy()
	}
	return fmt.Errorf("Failed to execute deployments")
}
