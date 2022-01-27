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
	if err := runHelmExecute(config, utils, log.Writer()); err != nil {
		log.Entry().WithError(err).Fatalf("step execution failed: %v", err)
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
		DeployCommand:         config.DeployCommand,
		HelmDeployWaitSeconds: config.HelmDeployWaitSeconds,
		DryRun:                config.DryRun,
		PackageVersion:        config.PackageVersion,
		AppVersion:            config.AppVersion,
		DependencyUpdate:      config.DependencyUpdate,
		HelmValues:            config.HelmValues,
		FilterTest:            config.FilterTest,
		DumpLogs:              config.DumpLogs,
		ChartRepo:             config.ChartRepo,
		HelmRegistryUser:      config.HelmRegistryUser,
	}
	switch config.DeployCommand {
	case "upgrade":
		if err := kubernetes.RunHelmUpgrade(helmConfig, utils, stdout); err != nil {
			return fmt.Errorf("failed to execute upgrade: %v", err)
		}
	case "lint":
		if err := kubernetes.RunHelmLint(helmConfig, utils, stdout); err != nil {
			return fmt.Errorf("failed to execute helm lint: %v", err)
		}
	case "install":
		if err := kubernetes.RunHelmInstall(helmConfig, utils, stdout); err != nil {
			return fmt.Errorf("failed to execute helm install: %v", err)
		}
	case "test":
		if err := kubernetes.RunHelmTest(helmConfig, utils, stdout); err != nil {
			return fmt.Errorf("failed to execute helm test: %v", err)
		}
	case "uninstall":
		if err := kubernetes.RunHelmUninstall(helmConfig, utils, stdout); err != nil {
			return fmt.Errorf("failed to execute helm uninstall: %v", err)
		}
	case "package":
		if err := kubernetes.RunHelmPackage(helmConfig, utils, stdout); err != nil {
			return fmt.Errorf("failed to execute helm package: %v", err)
		}
	}
	return nil
}
