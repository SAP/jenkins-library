package cmd

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/kubernetes"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func helmExecute(config helmExecuteOptions, telemetryData *telemetry.CustomData) {
	helmConfig := kubernetes.HelmExecuteOptions{
		ChartPath:                     config.ChartPath,
		DeploymentName:                config.DeploymentName,
		Image:                         config.Image,
		Namespace:                     config.Namespace,
		KubeContext:                   config.KubeContext,
		KubeConfig:                    config.KubeConfig,
		HelmDeployWaitSeconds:         config.HelmDeployWaitSeconds,
		PackageVersion:                config.PackageVersion,
		AppVersion:                    config.AppVersion,
		DependencyUpdate:              config.DependencyUpdate,
		HelmValues:                    config.HelmValues,
		FilterTest:                    config.FilterTest,
		DumpLogs:                      config.DumpLogs,
		TargetChartRepositoryURL:      config.TargetChartRepositoryURL,
		TargetChartRepositoryName:     config.TargetChartRepositoryName,
		TargetChartRepositoryUser:     config.TargetChartRepositoryUser,
		TargetChartRepositoryPassword: config.TargetChartRepositoryPassword,
		HelmCommand:                   config.HelmCommand,
		CustomTLSCertificateLinks:     config.CustomTLSCertificateLinks,
	}

	utils := kubernetes.NewDeployUtilsBundle(helmConfig.CustomTLSCertificateLinks)

	helmExecutor := kubernetes.NewHelmExecutor(helmConfig, utils, GeneralConfig.Verbose, log.Writer())

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	if err := runHelmExecute(config.HelmCommand, helmExecutor); err != nil {
		log.Entry().WithError(err).Fatalf("step execution failed: %v", err)
	}
}

func runHelmExecute(helmCommand string, helmExecutor kubernetes.HelmExecutor) error {
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
	case "publish":
		if err := helmExecutor.RunHelmPublish(); err != nil {
			return fmt.Errorf("failed to execute helm publish: %v", err)
		}
	default:
		if err := runHelmExecuteDefault(helmCommand, helmExecutor); err != nil {
			return fmt.Errorf("failed to execute helm command: %v", err)
		}
	}

	return nil
}

func runHelmExecuteDefault(helmCommand string, helmExecutor kubernetes.HelmExecutor) error {
	if err := helmExecutor.RunHelmLint(); err != nil {
		return fmt.Errorf("failed to execute helm lint: %v", err)
	}
	if err := helmExecutor.RunHelmPackage(); err != nil {
		return fmt.Errorf("failed to execute helm package: %v", err)
	}
	if err := helmExecutor.RunHelmPublish(); err != nil {
		return fmt.Errorf("failed to execute helm publish: %v", err)
	}

	return nil
}
