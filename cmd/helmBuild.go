package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/build"
	"github.com/SAP/jenkins-library/pkg/buildsettings"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/docker"
	"github.com/SAP/jenkins-library/pkg/helm"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/kubernetes"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/syft"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/versioning"
)

// helmDockerConfigDir is the directory syft uses to authenticate against the
// container registry when scanning referenced images.
const helmDockerConfigDir = "/root/.docker"

func helmBuild(config helmBuildOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *helmBuildCommonPipelineEnvironment) {
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

	// dependencies required for SBOM generation (syft): a command runner to
	// execute the syft binary, an http client to download it, and file utils.
	execRunner := &command.Command{}
	execRunner.Stdout(log.Writer())
	execRunner.Stderr(log.Writer())
	httpClient := &piperhttp.Client{}
	fileUtils := &piperutils.Files{}

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	if err := runHelmBuild(config, helmExecutor, utils, commonPipelineEnvironment, execRunner, fileUtils, httpClient, artifactInfo); err != nil {
		log.Entry().WithError(err).Fatalf("step execution failed: %v", err)
	}
}

func runHelmBuild(config helmBuildOptions, helmExecutor kubernetes.HelmExecutor, utils fileHandler, commonPipelineEnvironment *helmBuildCommonPipelineEnvironment, execRunner command.ExecRunner, fileUtils piperutils.FileUtils, httpClient piperhttp.Sender, artifact ...versioning.Coordinates) error {
	var artifactInfo versioning.Coordinates
	if len(artifact) > 0 {
		artifactInfo = artifact[0]
	}
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
		if config.CreateBOM {
			generateSBOMs(config, helmExecutor, execRunner, fileUtils, httpClient)
		}
		if config.CreateBuildArtifactsMetadata {
			if err := createHelmBuildArtifactsMetadata(artifactInfo, targetURL, commonPipelineEnvironment); err != nil {
				return err
			}
		}
	default:
		if err := runHelmBuildDefault(config, helmExecutor, commonPipelineEnvironment, execRunner, fileUtils, httpClient, artifactInfo); err != nil {
			return err
		}
	}

	// buildSettingsInfo is written only on the success path — a failed helm run
	// produces no meaningful build artifact, so there is nothing to report to
	// downstream compliance steps.
	log.Entry().Debugf("creating build settings information...")
	dockerImage, err := GetDockerImageValue("helmBuild")
	if err != nil {
		log.Entry().Warnf("failed to retrieve dockerImage configuration: %v", err)
	}
	buildSettingsInfo, err := buildsettings.CreateBuildSettingsInfo(&buildsettings.BuildOptions{
		DockerImage:       dockerImage,
		Publish:           config.Publish,
		CreateBOM:         config.CreateBOM,
		BuildSettingsInfo: config.BuildSettingsInfo,
	}, "helmBuild")
	if err != nil {
		log.Entry().Warnf("failed to create build settings info: %v", err)
	}
	commonPipelineEnvironment.custom.buildSettingsInfo = buildSettingsInfo

	return nil
}

func runHelmBuildDefault(config helmBuildOptions, helmExecutor kubernetes.HelmExecutor, commonPipelineEnvironment *helmBuildCommonPipelineEnvironment, execRunner command.ExecRunner, fileUtils piperutils.FileUtils, httpClient piperhttp.Sender, artifact versioning.Coordinates) error {
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
		if config.CreateBOM {
			generateSBOMs(config, helmExecutor, execRunner, fileUtils, httpClient)
		}
		if config.CreateBuildArtifactsMetadata {
			if err := createHelmBuildArtifactsMetadata(artifact, targetURL, commonPipelineEnvironment); err != nil {
				return err
			}
		}
	}

	return nil
}

// generateSBOMs produces both SBOMs for the published chart, sharing a single
// discovered image set so the chart BOM and the container BOMs describe the
// same images. Both are best-effort: a failure is logged but never fails the
// step.
func generateSBOMs(config helmBuildOptions, helmExecutor kubernetes.HelmExecutor, execRunner command.ExecRunner, fileUtils piperutils.FileUtils, httpClient piperhttp.Sender) {
	images := discoverImages(config, helmExecutor)

	if err := generateContainerSBOMs(config, images, execRunner, fileUtils, httpClient); err != nil {
		log.Entry().Warnf("container SBOM generation failed: %v", err)
	}

	if err := helm.GenerateChartSBOM(config.ChartPath, "bom-helm.xml", images, fileUtils); err != nil {
		log.Entry().Warnf("chart SBOM generation failed: %v", err)
	} else {
		log.Entry().Infof("helm SBOM: generated chart SBOM bom-helm.xml for %s", config.ChartPath)
	}
}

// generateContainerSBOMs generates a CycloneDX SBOM (bom-docker-<N>.xml) for
// each of the given container images, using Syft. The registry is derived from
// each full image reference.
func generateContainerSBOMs(config helmBuildOptions, imageNameTags []string, execRunner command.ExecRunner, fileUtils piperutils.FileUtils, httpClient piperhttp.Sender) error {
	if len(imageNameTags) == 0 {
		log.Entry().Info("helm SBOM: no container images available, skipping syft scan")
		return nil
	}

	registryURL, err := docker.ContainerRegistryFromImage(imageNameTags[0])
	if err != nil {
		log.Entry().Infof("helm SBOM: failed to derive registry from image %q: %v", imageNameTags[0], err)
		return fmt.Errorf("failed to derive registry from image %q: %w", imageNameTags[0], err)
	}

	images := make([]string, 0, len(imageNameTags))
	for _, image := range imageNameTags {
		nameTag, err := docker.ContainerImageNameTagFromImage(image)
		if err != nil {
			log.Entry().Infof("helm SBOM: failed to parse image %q: %v", image, err)
			return fmt.Errorf("failed to parse image %q: %w", image, err)
		}
		images = append(images, nameTag)
	}

	if err := syft.GenerateSBOM(config.SyftDownloadURL, helmDockerConfigDir, execRunner, fileUtils, httpClient, registryURL, images); err != nil {
		return err
	}

	// Syft omits the PURL for the root/parent component of an image BOM
	// (anchore/syft#1408); without it the BOM fails CycloneDX validation with
	// "Purl is missing for the root component!". Inject a clean docker PURL.
	return injectContainerBOMPurls()
}

// discoverImages returns the container images referenced by the chart. It
// prefers images rendered by `helm template`; if templating fails or yields no
// images, it falls back to the containerImageNameTags CPE list.
func discoverImages(config helmBuildOptions, helmExecutor kubernetes.HelmExecutor) []string {
	manifests, err := helmExecutor.RunHelmTemplate()
	if err != nil {
		log.Entry().Warnf("helm SBOM: helm template failed, falling back to CPE image list: %v", err)
		return config.ContainerImageNameTags
	}

	images, err := kubernetes.ExtractImagesFromManifests(manifests)
	if err != nil {
		log.Entry().Warnf("helm SBOM: failed to parse rendered manifests, falling back to CPE image list: %v", err)
		return config.ContainerImageNameTags
	}

	if len(images) == 0 {
		return config.ContainerImageNameTags
	}
	return images
}

// injectContainerBOMPurls sets a clean, registry-free docker PURL on the root
// component of every bom-docker-*.xml produced by Syft. Syft does not emit a
// PURL for the parent component (anchore/syft#1408), which makes the BOM fail
// CycloneDX validation. Best-effort: individual failures are logged, not fatal.
func injectContainerBOMPurls() error {
	files, err := filepath.Glob("bom-docker-*.xml")
	if err != nil || len(files) == 0 {
		log.Entry().Debug("helm SBOM: no bom-docker-*.xml files to update with a root PURL")
		return nil
	}

	for _, file := range files {
		component := piperutils.GetComponent(file)
		if component.Name == "" {
			log.Entry().Warnf("helm SBOM: unable to read root component from %s, skipping PURL injection", file)
			continue
		}

		// Build a clean, registry-free docker PURL (e.g. pkg:docker/nginx@1.25)
		// for the root component — shared with kaniko (see
		// piperutils.BuildRegistryFreeDockerPurl).
		constructedPurl, _, _, _, err := piperutils.BuildRegistryFreeDockerPurl(component.Name, component.Version)
		if err != nil {
			log.Entry().Warnf("helm SBOM: %v for %s", err, file)
			continue
		}

		if err := piperutils.UpdatePurl(file, constructedPurl); err != nil {
			log.Entry().Warnf("helm SBOM: unable to update root purl in %s: %v", file, err)
		}
	}

	return nil
}

func createHelmBuildArtifactsMetadata(artifact versioning.Coordinates, targetURL string, commonPipelineEnvironment *helmBuildCommonPipelineEnvironment) error {
	// defensive: need chart name + version to build a coordinate
	if len(artifact.ArtifactID) == 0 || len(artifact.Version) == 0 {
		log.Entry().Warnf("missing chart name or version, not creating helm build artifact metadata")
		return nil
	}
	coordinate := versioning.Coordinates{
		ArtifactID: artifact.ArtifactID,
		Version:    artifact.Version,
		URL:        targetURL,
		PURL:       fmt.Sprintf("pkg:helm/%s@%s", artifact.ArtifactID, artifact.Version),
	}
	var buildArtifacts build.BuildArtifacts
	buildArtifacts.Coordinates = []versioning.Coordinates{coordinate}
	jsonResult, err := json.Marshal(buildArtifacts)
	if err != nil {
		log.Entry().Warnf("unable to marshal helm build artifact metadata: %v", err)
		return nil
	}
	commonPipelineEnvironment.custom.helmBuildArtifacts = string(jsonResult)
	log.Entry().Infof("helm build artifact metadata created for %s@%s", coordinate.ArtifactID, coordinate.Version)
	return nil
}

// parseAndRenderCPETemplate allows to parse and render a template which contains references to the CPE
func parseAndRenderCPETemplate(config helmBuildOptions, rootPath string, utils fileHandler) error {
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
