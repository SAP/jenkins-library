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
		Image:                         config.Image,
		Namespace:                     config.Namespace,
		KubeContext:                   config.KubeContext,
		KubeConfig:                    config.KubeConfig,
		HelmDeployWaitSeconds:         config.HelmDeployWaitSeconds,
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

	helmChart := config.ChartPath + "Chart.yaml"
	nameChart, packageVersion, err := kubernetes.GetChartInfo(helmChart, utils)
	if err != nil {
		log.Entry().WithError(err).Fatalf("failed to get version in Chart.yaml: %v", err)
	}
	helmConfig.DeploymentName = nameChart
	helmConfig.PackageVersion = packageVersion

	helmExecutor := kubernetes.NewHelmExecutor(helmConfig, utils, GeneralConfig.Verbose, log.Writer())

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	if err := runHelmExecute(config, helmExecutor); err != nil {
		log.Entry().WithError(err).Fatalf("step execution failed: %v", err)
	}
}

func runHelmExecute(config helmExecuteOptions, helmExecutor kubernetes.HelmExecutor) error {
	switch config.HelmCommand {
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
		if err := runHelmExecuteDefault(config, helmExecutor); err != nil {
			return err
		}
	}

	return nil
}

func runHelmExecuteDefault(config helmExecuteOptions, helmExecutor kubernetes.HelmExecutor) error {
	if config.LintFlag {
		if err := helmExecutor.RunHelmLint(); err != nil {
			return fmt.Errorf("failed to execute helm lint: %v", err)
		}
	}

	if config.PackageFlag {
		if err := helmExecutor.RunHelmPackage(); err != nil {
			return fmt.Errorf("failed to execute helm package: %v", err)
		}
	}

	if config.PublishFlag {
		if err := helmExecutor.RunHelmPublish(); err != nil {
			return fmt.Errorf("failed to execute helm publish: %v", err)
		}
	}

	return nil
}
