package cmd

import (
	"fmt"
	"io"

	"github.com/SAP/jenkins-library/pkg/kubernetes"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func helmExecute(config helmExecuteOptions, telemetryData *telemetry.CustomData) {
	utils := kubernetes.NewDeployUtilsBundle()

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	if config.DeployTool == "helm" || config.DeployTool == "helm3" {
		err := runHelmExecute(config, utils, log.Writer())
		if err != nil {
			log.Entry().WithError(err).Fatal("step execution failed")
		}
	} else {
		log.Entry().Error("Failed to execute deployments since '%v' tool is not a helm", config.DeployTool)
	}
}

func runHelmExecute(config helmExecuteOptions, utils kubernetes.HelmDeployUtils, stdout io.Writer) error {
	helmConfig := kubernetes.HelmExecuteOptions{
		ChartPath:            config.ChartPath,
		DeploymentName:       config.DeploymentName,
		ContainerRegistryURL: config.ContainerRegistryURL,
		Image:                config.Image,
		ContainerImageName:   config.ContainerImageName,
		ContainerImageTag:    config.ContainerImageTag,
		Namespace:            config.Namespace,
		KubeContext:          config.KubeContext,
		KubeConfig:           config.KubeConfig,
		DeployTool:           config.DeployTool,
		TillerNamespace:      config.TillerNamespace,
	}

	if config.DeployCommand == "upgrade" {
		err := kubernetes.RunHelmUpgrade(helmConfig, utils, stdout)
		if err != nil {
			return fmt.Errorf("failed to execute deployments")
		}
	}

	// ToDo: helm lint
	if config.DeployCommand == "lint" {
		kubernetes.RunHelmLint()
	}

	// ToDo: helm install
	if config.DeployCommand == "install" {
		kubernetes.RunHelmInstall()
	}

	// ToDo: helm test
	if config.DeployCommand == "test" {
		kubernetes.RunHelmTest()
	}

	// ToDo: helm delete
	if config.DeployCommand == "delete" {
		kubernetes.RunHelmDelete()
	}

	// ToDo: helm package
	if config.DeployCommand == "package" {
		kubernetes.RunHelmPackage()
	}

	return nil
}
