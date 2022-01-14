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
	if config.DeployTool == "helm3" {
		err := runHelmExecute(config, utils, log.Writer())
		if err != nil {
			log.Entry().WithError(err).Fatal("step execution failed")
		}
	} else if config.DeployTool == "helm" {
		log.Entry().Error("Failed to execute deployments since '%v' is not support deployment via helmExecute step, due to helm2 is deprecated", config.DeployTool)
	} else {
		log.Entry().Error("Failed to execute deployments since '%v' tool is not a helm3.", config.DeployTool)
	}
}

func runHelmExecute(config helmExecuteOptions, utils kubernetes.HelmDeployUtils, stdout io.Writer) error {
	helmConfig := kubernetes.HelmExecuteOptions{
		ChartPath:             config.ChartPath,
		DeploymentName:        config.DeploymentName,
		ContainerRegistryURL:  config.ContainerRegistryURL,
		Image:                 config.Image,
		ContainerImageName:    config.ContainerImageName,
		ContainerImageTag:     config.ContainerImageTag,
		Namespace:             config.Namespace,
		KubeContext:           config.KubeContext,
		KubeConfig:            config.KubeConfig,
		DeployTool:            config.DeployTool,
		DeployCommand:         config.DeployCommand,
		HelmDeployWaitSeconds: config.HelmDeployWaitSeconds,
		DryRun:                config.DryRun,
	}

	switch config.DeployCommand {
	case "upgrade":
		err := kubernetes.RunHelmUpgrade(helmConfig, utils, stdout)
		if err != nil {
			return fmt.Errorf("failed to execute deployments")
		}
	case "lint":
		kubernetes.RunHelmLint()
	case "install":
		err := kubernetes.RunHelmInstall(helmConfig, utils, stdout)
		if err != nil {
			return fmt.Errorf("failed to execute helm install")
		}
	case "test":
		kubernetes.RunHelmTest()
	case "uninstall":
		err := kubernetes.RunHelmUninstall(helmConfig, utils, stdout)
		if err != nil {
			return fmt.Errorf("failed to execute helm uninstall")
		}
	case "package":
		kubernetes.RunHelmPackage()
	default:
		log.Entry().Error("Command '%v' is not supported Piper tool", config.DeployCommand)
	}

	return nil
}
