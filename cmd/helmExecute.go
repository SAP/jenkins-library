package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/kubernetes"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/versioning"
)

func helmExecute(config helmExecuteOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *helmExecuteCommonPipelineEnvironment) {
	helmConfig := kubernetes.HelmExecuteOptions{
		AdditionalParameters:      config.AdditionalParameters,
		ChartPath:                 config.ChartPath,
		Image:                     config.Image,
		Namespace:                 config.Namespace,
		KubeContext:               config.KubeContext,
		KeepFailedDeployments:     config.KeepFailedDeployments,
		KubeConfig:                config.KubeConfig,
		HelmDeployWaitSeconds:     config.HelmDeployWaitSeconds,
		DockerConfigJSON:          config.DockerConfigJSON,
		AppVersion:                config.AppVersion,
		Dependency:                config.Dependency,
		PackageDependencyUpdate:   config.PackageDependencyUpdate,
		HelmValues:                config.HelmValues,
		FilterTest:                config.FilterTest,
		DumpLogs:                  config.DumpLogs,
		TargetRepositoryURL:       config.TargetRepositoryURL,
		TargetRepositoryName:      config.TargetRepositoryName,
		TargetRepositoryUser:      config.TargetRepositoryUser,
		TargetRepositoryPassword:  config.TargetRepositoryPassword,
		SourceRepositoryName:      config.SourceRepositoryName,
		SourceRepositoryURL:       config.SourceRepositoryURL,
		SourceRepositoryUser:      config.SourceRepositoryUser,
		SourceRepositoryPassword:  config.SourceRepositoryPassword,
		HelmCommand:               config.HelmCommand,
		CustomTLSCertificateLinks: config.CustomTLSCertificateLinks,
		Version:                   config.Version,
		PublishVersion:            config.Version,
		RenderSubchartNotes:       config.RenderSubchartNotes,
	}

	utils := kubernetes.NewDeployUtilsBundle(helmConfig.CustomTLSCertificateLinks)

	artifactOpts := versioning.Options{
		VersioningScheme: "library",
	}

	buildDescriptorFile := ""
	if helmConfig.ChartPath != "" {
		buildDescriptorFile = filepath.Join(helmConfig.ChartPath, "Chart.yaml")
	}

	artifact, err := versioning.GetArtifact("helm", buildDescriptorFile, &artifactOpts, utils)
	if err != nil {
		log.Entry().WithError(err).Fatalf("getting artifact information failed: %v", err)
	}
	artifactInfo, err := artifact.GetCoordinates()
	if err != nil {
		log.Entry().WithError(err).Fatalf("getting artifact coordinates failed: %v", err)
	}

	helmConfig.DeploymentName = artifactInfo.ArtifactID

	if len(helmConfig.PublishVersion) == 0 {
		helmConfig.PublishVersion = artifactInfo.Version
	}

	helmExecutor := kubernetes.NewHelmExecutor(helmConfig, utils, GeneralConfig.Verbose, log.Writer())

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	if err := runHelmExecute(config, helmExecutor, utils, commonPipelineEnvironment); err != nil {
		log.Entry().WithError(err).Fatalf("step execution failed: %v", err)
	}
}

func runHelmExecute(config helmExecuteOptions, helmExecutor kubernetes.HelmExecutor, utils fileHandler, commonPipelineEnvironment *helmExecuteCommonPipelineEnvironment) error {
	if config.RenderValuesTemplate {
		err := parseAndRenderCPETemplate(config, GeneralConfig.EnvRootPath, utils)
		if err != nil {
			log.Entry().WithError(err).Fatalf("failed to parse/render template: %v", err)
		}
	}
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
	case "dependency":
		if err := helmExecutor.RunHelmDependency(); err != nil {
			return fmt.Errorf("failed to execute helm dependency: %v", err)
		}
	case "publish":
		targetURL, err := helmExecutor.RunHelmPublish()
		if err != nil {
			return fmt.Errorf("failed to execute helm publish: %v", err)
		}
		commonPipelineEnvironment.custom.helmChartURL = targetURL
	default:
		if err := runHelmExecuteDefault(config, helmExecutor, commonPipelineEnvironment); err != nil {
			return err
		}
	}

	return nil
}

func runHelmExecuteDefault(config helmExecuteOptions, helmExecutor kubernetes.HelmExecutor, commonPipelineEnvironment *helmExecuteCommonPipelineEnvironment) error {
	if len(config.Dependency) > 0 {
		if err := helmExecutor.RunHelmDependency(); err != nil {
			return fmt.Errorf("failed to execute helm dependency: %v", err)
		}
	}

	if err := helmExecutor.RunHelmLint(); err != nil {
		return fmt.Errorf("failed to execute helm lint: %v", err)
	}

	if config.Publish {
		targetURL, err := helmExecutor.RunHelmPublish()
		if err != nil {
			return fmt.Errorf("failed to execute helm publish: %v", err)
		}
		commonPipelineEnvironment.custom.helmChartURL = targetURL
	}

	return nil
}

// parseAndRenderCPETemplate allows to parse and render a template which contains references to the CPE
func parseAndRenderCPETemplate(config helmExecuteOptions, rootPath string, utils fileHandler) error {
	cpe := piperenv.CPEMap{}
	err := cpe.LoadFromDisk(path.Join(rootPath, "commonPipelineEnvironment"))
	if err != nil {
		return fmt.Errorf("failed to load values from commonPipelineEnvironment: %v", err)
	}

	valueFiles := []string{}
	defaultValueFile := fmt.Sprintf("%s/%s", config.ChartPath, "values.yaml")
	defaultValueFileExists, err := utils.FileExists(defaultValueFile)
	if err != nil {
		return err
	}

	if defaultValueFileExists {
		valueFiles = append(valueFiles, defaultValueFile)
	} else {
		if len(config.HelmValues) == 0 {
			return fmt.Errorf("no value file to proccess, please provide value file(s)")
		}
	}
	valueFiles = append(valueFiles, config.HelmValues...)

	for _, valueFile := range valueFiles {
		cpeTemplate, err := utils.FileRead(valueFile)
		if err != nil {
			return fmt.Errorf("failed to read file: %v", err)
		}
		generated, err := cpe.ParseTemplateWithDelimiter(string(cpeTemplate), config.TemplateStartDelimiter, config.TemplateEndDelimiter)
		if err != nil {
			return fmt.Errorf("failed to parse template: %v", err)
		}
		err = utils.FileWrite(valueFile, generated.Bytes(), 0700)
		if err != nil {
			return fmt.Errorf("failed to update file: %v", err)
		}
	}

	return nil
}

type fileHandler interface {
	FileExists(string) (bool, error)
	FileRead(string) ([]byte, error)
	FileWrite(string, []byte, os.FileMode) error
}
