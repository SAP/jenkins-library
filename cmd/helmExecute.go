package cmd

import (
	"fmt"
	"io"

	"github.com/SAP/jenkins-library/pkg/helm"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func helmExecute(config helmExecuteOptions, telemetryData *telemetry.CustomData) {
	helmExecuteOption := helm.ExecutorOptions{}
	helmExecutor := helm.NewExecutor(helmExecuteOption)
	// utils := helm.NewHelmDeployUtilsBundle()

	if config.DeployTool == "helm" || config.DeployTool == "helm3" {
		err := runHelmExecute(helmExecutor, &config, log.Writer())
		if err != nil {
			log.Entry().WithError(err).Fatal("step execution failed")
		}
	} else {
		fmt.Errorf("Failed to execute deployments")
	}
}

func runHelmExecute(helmExecutor helm.Executor, config *helmExecuteOptions, stdout io.Writer) error {
	if len(config.ChartPath) <= 0 {
		return fmt.Errorf("chart path has not been set, please configure chartPath parameter")
	}
	if len(config.DeploymentName) <= 0 {
		return fmt.Errorf("deployment name has not been set, please configure deploymentName parameter")
	}

	_, containerRegistry, err := helm.SplitRegistryURL(config.ContainerRegistryURL)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Container registry url '%v' incorrect", config.ContainerRegistryURL)
	}
	//support either image or containerImageName and containerImageTag
	containerImageName := ""
	containerImageTag := ""

	if len(config.Image) > 0 {
		containerImageName, containerImageTag, err = helm.SplitFullImageName(config.Image)
		if err != nil {
			log.Entry().WithError(err).Fatalf("Container image '%v' incorrect", config.Image)
		}
	} else if len(config.ContainerImageName) > 0 && len(config.ContainerImageTag) > 0 {
		containerImageName = config.ContainerImageName
		containerImageTag = config.ContainerImageTag
	} else {
		return fmt.Errorf("image information not given - please either set image or containerImageName and containerImageTag")
	}

	helmConfig := helm.HelmExecuteOptions{
		AdditionalParameters: config.AdditionalParameters,
	}

	err = helmExecutor.RunHelmLint(containerRegistry, containerImageName, containerImageTag, helmConfig, stdout)
	if err != nil {
		log.Entry().Warnf("failed to execute helm lint: %v", err)
	}

	return nil
}
