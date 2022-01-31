package cmd

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/kubernetes"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func helmExecute(config helmExecuteOptions, telemetryData *telemetry.CustomData) {
	utils := kubernetes.NewDeployUtilsBundle()

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

	helmExecutor := kubernetes.NewHelmExecutor(helmConfig, utils, GeneralConfig.Verbose, log.Writer())

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	if err := runHelmExecute(config.HelmCommand, config.AdditionalParameters, helmExecutor); err != nil {
		log.Entry().WithError(err).Fatalf("step execution failed: %v", err)
	}
}

func runHelmExecute(helmCommand string, additionalParameters []string, helmExecutor kubernetes.HelmExecutor) error {

	if helmCommand == "" && len(additionalParameters) == 0 {
		return fmt.Errorf("helm command is not presented")
	}

	switch helmCommand {
	case "upgrade":
		if err := helmExecutor.RunHelmUpgrade(); err != nil {
			return fmt.Errorf("failed to execute upgrade: %v", err)
		}
	case "lint":
		if err := helmExecutor.RunHelmLint(); err != nil {
			return fmt.Errorf("failed to execute helm lint: %v", err)
		}
	case "install":
		if err := helmExecutor.RunHelmInstall(); err != nil {
			return fmt.Errorf("failed to execute helm install: %v", err)
		}
	case "test":
		if err := helmExecutor.RunHelmTest(); err != nil {
			return fmt.Errorf("failed to execute helm test: %v", err)
		}
	case "uninstall":
		if err := helmExecutor.RunHelmUninstall(); err != nil {
			return fmt.Errorf("failed to execute helm uninstall: %v", err)
		}
	case "package":
		if err := helmExecutor.RunHelmPackage(); err != nil {
			return fmt.Errorf("failed to execute helm package: %v", err)
		}
	case "push":
		if err := helmExecutor.RunHelmPush(); err != nil {
			return fmt.Errorf("failed to execute helm package: %v", err)
		}
	default:
		fmt.Printf("Helm command will be executed directly")
		if err := helmExecutor.RunHelmDirect(); err != nil {
			return fmt.Errorf("failed to execute helm command directly: %v", err)
		}
	}

	return nil
}
