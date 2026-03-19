package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"

	"github.com/SAP/jenkins-library/pkg/build"
	"github.com/SAP/jenkins-library/pkg/buildsettings"
	"github.com/SAP/jenkins-library/pkg/certutils"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/docker"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/syft"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/versioning"
	"github.com/moby/buildkit/util/purl"
	purlParser "github.com/package-url/packageurl-go"
)

const (
	buildkitDockerConfigDir  = "/root/.docker"
	buildkitDockerConfigPath = "/root/.docker/config.json"
	buildkitTLSCertPath      = "/etc/ssl/certs/ca-certificates.crt"
	buildkitDaemonJSONPath   = "/etc/docker/daemon.json"
)

func buildkitExecute(config buildkitExecuteOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *buildkitExecuteCommonPipelineEnvironment) {
	c := command.Command{
		ErrorCategoryMapping: map[string][]string{
			log.ErrorConfiguration.String(): {
				"unsupported status code 401",
			},
		},
		StepName: "buildkitExecute",
	}
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	client := &piperhttp.Client{}
	fileUtils := &piperutils.Files{}

	err := runBuildkitExecute(&config, telemetryData, commonPipelineEnvironment, &c, client, fileUtils)
	if err != nil {
		log.Entry().WithError(err).Fatal("BuildKit execution failed")
	}
}

func runBuildkitExecute(config *buildkitExecuteOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *buildkitExecuteCommonPipelineEnvironment, execRunner command.ExecRunner, httpClient piperhttp.Sender, fileUtils piperutils.FileUtils) error {
	// backward compatibility for parameter ContainerBuildOptions
	if len(config.ContainerBuildOptions) > 0 {
		config.BuildOptions = strings.Split(config.ContainerBuildOptions, " ")
		log.Entry().Warning("Parameter containerBuildOptions is deprecated, please use buildOptions instead.")
		telemetryData.ContainerBuildOptions = config.ContainerBuildOptions
	}

	// run container preparation command if configured
	if len(config.ContainerPreparationCommand) > 0 {
		prepCommand := strings.Split(config.ContainerPreparationCommand, " ")
		if err := execRunner.RunExecutable(prepCommand[0], prepCommand[1:]...); err != nil {
			return fmt.Errorf("failed to run container preparation command: %w", err)
		}
	}

	if len(config.CustomTLSCertificateLinks) > 0 {
		err := certutils.CertificateUpdate(config.CustomTLSCertificateLinks, httpClient, fileUtils, buildkitTLSCertPath)
		if err != nil {
			return fmt.Errorf("failed to update certificates: %w", err)
		}
	} else {
		log.Entry().Info("skipping updation of certificates")
	}

	// Docker config handling
	dockerConfig := []byte(`{"auths":{}}`)

	if len(config.DockerConfigJSON) > 0 {
		var err error
		dockerConfig, err = fileUtils.FileRead(config.DockerConfigJSON)
		if err != nil {
			return fmt.Errorf("failed to read existing docker config json at '%v': %w", config.DockerConfigJSON, err)
		}
	}

	if len(config.DockerConfigJSON) > 0 && len(config.ContainerRegistryURL) > 0 && len(config.ContainerRegistryPassword) > 0 && len(config.ContainerRegistryUser) > 0 {
		targetConfigJSON, err := docker.CreateDockerConfigJSON(config.ContainerRegistryURL, config.ContainerRegistryUser, config.ContainerRegistryPassword, "", config.DockerConfigJSON, fileUtils)
		if err != nil {
			return fmt.Errorf("failed to update existing docker config json file '%v': %w", config.DockerConfigJSON, err)
		}
		dockerConfig, err = fileUtils.FileRead(targetConfigJSON)
		if err != nil {
			return fmt.Errorf("failed to read enhanced file '%v': %w", config.DockerConfigJSON, err)
		}
	} else if len(config.DockerConfigJSON) == 0 && len(config.ContainerRegistryURL) > 0 && len(config.ContainerRegistryPassword) > 0 && len(config.ContainerRegistryUser) > 0 {
		targetConfigJSON, err := docker.CreateDockerConfigJSON(config.ContainerRegistryURL, config.ContainerRegistryUser, config.ContainerRegistryPassword, "", buildkitDockerConfigPath, fileUtils)
		if err != nil {
			return fmt.Errorf("failed to create new docker config json at %s: %w", buildkitDockerConfigPath, err)
		}
		dockerConfig, err = fileUtils.FileRead(targetConfigJSON)
		if err != nil {
			return fmt.Errorf("failed to read new docker config file at %s: %w", buildkitDockerConfigPath, err)
		}
	}

	if err := fileUtils.FileWrite(buildkitDockerConfigPath, dockerConfig, 0644); err != nil {
		return fmt.Errorf("failed to write file '%s': %w", buildkitDockerConfigPath, err)
	}

	// Registry mirrors: write daemon.json for Docker daemon
	if len(config.RegistryMirrors) > 0 {
		daemonJSON, err := json.Marshal(map[string]interface{}{
			"registry-mirrors": config.RegistryMirrors,
		})
		if err != nil {
			return fmt.Errorf("failed to marshal daemon.json: %w", err)
		}
		if err := fileUtils.FileWrite(buildkitDaemonJSONPath, daemonJSON, 0644); err != nil {
			return fmt.Errorf("failed to write '%s': %w", buildkitDaemonJSONPath, err)
		}
		log.Entry().Infof("Registry mirrors configured: %v", config.RegistryMirrors)
	}

	// Build settings
	log.Entry().Debugf("preparing build settings information...")
	stepName := "buildkitExecute"
	dockerImage, err := GetDockerImageValue(stepName)
	if err != nil {
		return fmt.Errorf("failed to retrieve dockerImage configuration: %w", err)
	}

	buildkitConfig := buildsettings.BuildOptions{
		DockerImage:       dockerImage,
		BuildSettingsInfo: config.BuildSettingsInfo,
	}

	log.Entry().Debugf("creating build settings information...")
	buildSettingsInfo, err := buildsettings.CreateBuildSettingsInfo(&buildkitConfig, stepName)
	if err != nil {
		log.Entry().Warnf("failed to create build settings info: %v", err)
	}
	commonPipelineEnvironment.custom.buildSettingsInfo = buildSettingsInfo

	switch {
	case config.ContainerMultiImageBuild:
		log.Entry().Debugf("Multi-image build activated for image name '%v'", config.ContainerImageName)

		if config.ContainerRegistryURL == "" {
			return fmt.Errorf("empty ContainerRegistryURL")
		}
		if config.ContainerImageName == "" {
			return fmt.Errorf("empty ContainerImageName")
		}
		if config.ContainerImageTag == "" {
			return fmt.Errorf("empty ContainerImageTag")
		}

		containerRegistry, err := docker.ContainerRegistryFromURL(config.ContainerRegistryURL)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return fmt.Errorf("failed to read registry url %v: %w", config.ContainerRegistryURL, err)
		}

		commonPipelineEnvironment.container.registryURL = config.ContainerRegistryURL

		// Docker image tags don't allow plus signs in tags, thus replacing with dash
		containerImageTag := strings.ReplaceAll(config.ContainerImageTag, "+", "-")

		imageListWithFilePath, err := docker.ImageListWithFilePath(config.ContainerImageName, config.ContainerMultiImageBuildExcludes, config.ContainerMultiImageBuildTrimDir, fileUtils)
		if err != nil {
			return fmt.Errorf("failed to identify image list for multi image build: %w", err)
		}
		if len(imageListWithFilePath) == 0 {
			return fmt.Errorf("no docker files to process, please check exclude list")
		}
		for image, file := range imageListWithFilePath {
			log.Entry().Debugf("Building image '%v' using file '%v'", image, file)
			containerImageNameAndTag := fmt.Sprintf("%v:%v", image, containerImageTag)
			dest := fmt.Sprintf("%v/%v", containerRegistry, containerImageNameAndTag)
			if err = runBuildkit(config, file, dest, true, ".", execRunner, fileUtils, commonPipelineEnvironment); err != nil {
				return fmt.Errorf("failed to build image '%v' using '%v': %w", image, file, err)
			}
			commonPipelineEnvironment.container.imageNames = append(commonPipelineEnvironment.container.imageNames, image)
			commonPipelineEnvironment.container.imageNameTags = append(commonPipelineEnvironment.container.imageNameTags, containerImageNameAndTag)
		}

		// for compatibility reasons also fill single imageNameTag field with "root" image
		if len(imageListWithFilePath[config.ContainerImageName]) > 0 {
			containerImageNameAndTag := fmt.Sprintf("%v:%v", config.ContainerImageName, containerImageTag)
			commonPipelineEnvironment.container.imageNameTag = containerImageNameAndTag
		}
		if config.CreateBOM {
			err := syft.GenerateSBOM(config.SyftDownloadURL, buildkitDockerConfigDir, execRunner, fileUtils, httpClient, commonPipelineEnvironment.container.registryURL, commonPipelineEnvironment.container.imageNameTags)
			if err != nil {
				return err
			}
		}

		if config.CreateBuildArtifactsMetadata {
			err := buildkitCreateDockerBuildArtifactMetadata(commonPipelineEnvironment.container.imageNameTags, commonPipelineEnvironment)
			if err != nil {
				return err
			}
		}
		return nil

	case config.MultipleImages != nil:
		log.Entry().Debugf("multipleImages build activated")
		parsedMultipleImages, err := buildkitParseMultipleImages(config.MultipleImages)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return fmt.Errorf("failed to parse multipleImages param: %w", err)
		}

		for _, entry := range parsedMultipleImages {
			switch {
			case entry.ContextSubPath == "":
				return fmt.Errorf("multipleImages: empty contextSubPath")
			case entry.ContainerImageName != "":
				containerRegistry, err := docker.ContainerRegistryFromURL(config.ContainerRegistryURL)
				if err != nil {
					log.SetErrorCategory(log.ErrorConfiguration)
					return fmt.Errorf("multipleImages: failed to read registry url %v: %w", config.ContainerRegistryURL, err)
				}

				if entry.ContainerImageTag == "" {
					if config.ContainerImageTag == "" {
						return fmt.Errorf("both multipleImages containerImageTag and config.containerImageTag are empty")
					}
					entry.ContainerImageTag = config.ContainerImageTag
				}
				containerImageTag := strings.ReplaceAll(entry.ContainerImageTag, "+", "-")
				containerImageNameAndTag := fmt.Sprintf("%v:%v", entry.ContainerImageName, containerImageTag)

				log.Entry().Debugf("multipleImages: image build '%v'", entry.ContainerImageName)

				dockerfilePath := config.DockerfilePath
				if entry.DockerfilePath != "" {
					dockerfilePath = entry.DockerfilePath
				}

				dest := fmt.Sprintf("%v/%v", containerRegistry, containerImageNameAndTag)
				if err = runBuildkit(config, dockerfilePath, dest, true, entry.ContextSubPath, execRunner, fileUtils, commonPipelineEnvironment); err != nil {
					return fmt.Errorf("multipleImages: failed to build image '%v' using '%v': %w", entry.ContainerImageName, config.DockerfilePath, err)
				}

				commonPipelineEnvironment.container.imageNameTags = append(commonPipelineEnvironment.container.imageNameTags, containerImageNameAndTag)
				commonPipelineEnvironment.container.imageNames = append(commonPipelineEnvironment.container.imageNames, entry.ContainerImageName)

			case entry.ContainerImage != "":
				containerImageName, err := docker.ContainerImageNameFromImage(entry.ContainerImage)
				if err != nil {
					log.SetErrorCategory(log.ErrorConfiguration)
					return fmt.Errorf("invalid name part in image %v: %w", entry.ContainerImage, err)
				}
				containerImageNameTag, err := docker.ContainerImageNameTagFromImage(entry.ContainerImage)
				if err != nil {
					log.SetErrorCategory(log.ErrorConfiguration)
					return fmt.Errorf("invalid tag part in image %v: %w", entry.ContainerImage, err)
				}

				log.Entry().Debugf("multipleImages: image build '%v'", containerImageName)

				dockerfilePath := config.DockerfilePath
				if entry.DockerfilePath != "" {
					dockerfilePath = entry.DockerfilePath
				}

				if err = runBuildkit(config, dockerfilePath, entry.ContainerImage, true, entry.ContextSubPath, execRunner, fileUtils, commonPipelineEnvironment); err != nil {
					return fmt.Errorf("multipleImages: failed to build image '%v' using '%v': %w", containerImageName, config.DockerfilePath, err)
				}

				commonPipelineEnvironment.container.imageNameTags = append(commonPipelineEnvironment.container.imageNameTags, containerImageNameTag)
				commonPipelineEnvironment.container.imageNames = append(commonPipelineEnvironment.container.imageNames, containerImageName)
			default:
				return fmt.Errorf("multipleImages: either containerImageName or containerImage must be filled")
			}
		}

		containerImageTag := strings.ReplaceAll(config.ContainerImageTag, "+", "-")
		containerImageNameAndTag := fmt.Sprintf("%v:%v", config.ContainerImageName, containerImageTag)
		commonPipelineEnvironment.container.imageNameTag = containerImageNameAndTag
		commonPipelineEnvironment.container.registryURL = config.ContainerRegistryURL

		if config.CreateBOM {
			err := syft.GenerateSBOM(config.SyftDownloadURL, buildkitDockerConfigDir, execRunner, fileUtils, httpClient, commonPipelineEnvironment.container.registryURL, commonPipelineEnvironment.container.imageNameTags)
			if err != nil {
				return err
			}
		}

		if config.CreateBuildArtifactsMetadata {
			err := buildkitCreateDockerBuildArtifactMetadata(commonPipelineEnvironment.container.imageNameTags, commonPipelineEnvironment)
			if err != nil {
				return err
			}
		}
		return nil

	case buildkitHasDestination(config.BuildOptions):
		log.Entry().Infof("Running BuildKit build with destination defined via buildOptions: %v", config.BuildOptions)

		for i, o := range config.BuildOptions {
			if o == "-t" && i+1 < len(config.BuildOptions) {
				destination := config.BuildOptions[i+1]

				containerRegistry, err := docker.ContainerRegistryFromImage(destination)
				if err != nil {
					log.SetErrorCategory(log.ErrorConfiguration)
					return fmt.Errorf("invalid registry part in image %v: %w", destination, err)
				}
				if commonPipelineEnvironment.container.registryURL == "" {
					commonPipelineEnvironment.container.registryURL = fmt.Sprintf("https://%v", containerRegistry)
				}

				containerImageName, _ := docker.ContainerImageNameFromImage(destination)
				containerImageNameTag, _ := docker.ContainerImageNameTagFromImage(destination)

				if commonPipelineEnvironment.container.imageNameTag == "" {
					commonPipelineEnvironment.container.imageNameTag = containerImageNameTag
				}
				commonPipelineEnvironment.container.imageNameTags = append(commonPipelineEnvironment.container.imageNameTags, containerImageNameTag)
				commonPipelineEnvironment.container.imageNames = append(commonPipelineEnvironment.container.imageNames, containerImageName)
			}
		}

		// -t is already in buildOptions, so pass empty destination; push is true since user specified -t
		if err := runBuildkit(config, config.DockerfilePath, "", true, ".", execRunner, fileUtils, commonPipelineEnvironment); err != nil {
			return err
		}

	case config.ContainerRegistryURL != "" && config.ContainerImageName != "" && config.ContainerImageTag != "":
		log.Entry().Debugf("Single image build for image name '%v'", config.ContainerImageName)

		containerRegistry, err := docker.ContainerRegistryFromURL(config.ContainerRegistryURL)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return fmt.Errorf("failed to read registry url %v: %w", config.ContainerRegistryURL, err)
		}

		containerImageTag := strings.ReplaceAll(config.ContainerImageTag, "+", "-")
		containerImageNameAndTag := fmt.Sprintf("%v:%v", config.ContainerImageName, containerImageTag)

		commonPipelineEnvironment.container.registryURL = config.ContainerRegistryURL
		commonPipelineEnvironment.container.imageNameTag = containerImageNameAndTag
		commonPipelineEnvironment.container.imageNameTags = append(commonPipelineEnvironment.container.imageNameTags, containerImageNameAndTag)
		commonPipelineEnvironment.container.imageNames = append(commonPipelineEnvironment.container.imageNames, config.ContainerImageName)

		dest := fmt.Sprintf("%v/%v", containerRegistry, containerImageNameAndTag)
		if err = runBuildkit(config, config.DockerfilePath, dest, true, ".", execRunner, fileUtils, commonPipelineEnvironment); err != nil {
			return err
		}

	case config.ContainerImage != "":
		log.Entry().Debugf("Single image build for image '%v'", config.ContainerImage)

		containerRegistry, err := docker.ContainerRegistryFromImage(config.ContainerImage)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return fmt.Errorf("invalid registry part in image %v: %w", config.ContainerImage, err)
		}

		containerImageName, _ := docker.ContainerImageNameFromImage(config.ContainerImage)
		containerImageNameTag, _ := docker.ContainerImageNameTagFromImage(config.ContainerImage)

		commonPipelineEnvironment.container.registryURL = fmt.Sprintf("https://%v", containerRegistry)
		commonPipelineEnvironment.container.imageNameTag = containerImageNameTag
		commonPipelineEnvironment.container.imageNameTags = append(commonPipelineEnvironment.container.imageNameTags, containerImageNameTag)
		commonPipelineEnvironment.container.imageNames = append(commonPipelineEnvironment.container.imageNames, containerImageName)

		if err = runBuildkit(config, config.DockerfilePath, config.ContainerImage, true, ".", execRunner, fileUtils, commonPipelineEnvironment); err != nil {
			return err
		}

	default:
		// no-push: build without pushing
		if err := runBuildkit(config, config.DockerfilePath, "", false, ".", execRunner, fileUtils, commonPipelineEnvironment); err != nil {
			return err
		}
	}

	if config.CreateBOM {
		err := syft.GenerateSBOM(config.SyftDownloadURL, buildkitDockerConfigDir, execRunner, fileUtils, httpClient, commonPipelineEnvironment.container.registryURL, commonPipelineEnvironment.container.imageNameTags)
		if err != nil {
			return err
		}
	}
	if config.CreateBuildArtifactsMetadata {
		err := buildkitCreateDockerBuildArtifactMetadata(commonPipelineEnvironment.container.imageNameTags, commonPipelineEnvironment)
		if err != nil {
			return err
		}
	}

	return nil
}

func runBuildkit(config *buildkitExecuteOptions, dockerfilePath string, destination string, push bool, contextSubPath string, execRunner command.ExecRunner, fileUtils piperutils.FileUtils, commonPipelineEnvironment *buildkitExecuteCommonPipelineEnvironment) error {
	cwd, err := fileUtils.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	buildContext := cwd
	if contextSubPath != "" && contextSubPath != "." {
		buildContext = filepath.Join(cwd, contextSubPath)
	}

	buildkitOpts := []string{"buildx", "build", "--file", dockerfilePath}

	if destination != "" {
		buildkitOpts = append(buildkitOpts, "-t", destination)
	}
	if push {
		buildkitOpts = append(buildkitOpts, "--push")
	}

	tmpDir, err := fileUtils.TempDir("", "*-buildkitExecute")
	if err != nil {
		return fmt.Errorf("failed to create tmp dir for buildkitExecute: %w", err)
	}

	metadataFilePath := fmt.Sprintf("%s/metadata.json", tmpDir)

	if config.ReadImageDigest {
		buildkitOpts = append(buildkitOpts, "--metadata-file", metadataFilePath)
	}

	if GeneralConfig.Verbose {
		buildkitOpts = append(buildkitOpts, "--progress=plain")
	}

	buildkitOpts = append(buildkitOpts, config.BuildOptions...)

	// build context must be last argument
	buildkitOpts = append(buildkitOpts, buildContext)

	err = execRunner.RunExecutable("docker", buildkitOpts...)
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		return fmt.Errorf("execution of 'docker buildx build' failed: %w", err)
	}

	if config.ReadImageDigest {
		digest, err := extractDigestFromMetadata(metadataFilePath, fileUtils)
		if err != nil {
			log.Entry().Warnf("failed to extract image digest: %v", err)
		} else if digest != "" {
			log.Entry().Debugf("image digest: %s", digest)
			commonPipelineEnvironment.container.imageDigest = digest
			commonPipelineEnvironment.container.imageDigests = append(commonPipelineEnvironment.container.imageDigests, digest)
		}
	}

	return nil
}

func extractDigestFromMetadata(metadataFilePath string, fileUtils piperutils.FileUtils) (string, error) {
	exists, err := fileUtils.FileExists(metadataFilePath)
	if err != nil || !exists {
		return "", fmt.Errorf("metadata file '%s' not found", metadataFilePath)
	}

	data, err := fileUtils.FileRead(metadataFilePath)
	if err != nil {
		return "", fmt.Errorf("error reading metadata file: %w", err)
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(data, &metadata); err != nil {
		return "", fmt.Errorf("error parsing metadata JSON: %w", err)
	}

	digest, ok := metadata["containerimage.digest"]
	if !ok {
		return "", fmt.Errorf("'containerimage.digest' not found in metadata")
	}

	digestStr, ok := digest.(string)
	if !ok {
		return "", fmt.Errorf("'containerimage.digest' is not a string")
	}

	return digestStr, nil
}

// buildkitHasDestination checks if buildOptions contains a -t flag (BuildKit equivalent of --destination)
func buildkitHasDestination(buildOptions []string) bool {
	for _, o := range buildOptions {
		if o == "-t" {
			return true
		}
	}
	return false
}

// Duplicated helpers from kanikoExecute (to be refactored to shared package later)

type buildkitMultipleImageConf struct {
	ContextSubPath     string `json:"contextSubPath,omitempty"`
	DockerfilePath     string `json:"dockerfilePath,omitempty"`
	ContainerImageName string `json:"containerImageName,omitempty"`
	ContainerImageTag  string `json:"containerImageTag,omitempty"`
	ContainerImage     string `json:"containerImage,omitempty"`
}

func buildkitParseMultipleImages(src []map[string]interface{}) ([]buildkitMultipleImageConf, error) {
	var result []buildkitMultipleImageConf

	for _, conf := range src {
		var structuredConf buildkitMultipleImageConf
		if err := mapstructure.Decode(conf, &structuredConf); err != nil {
			return nil, err
		}
		result = append(result, structuredConf)
	}

	return result, nil
}

func buildkitCreateDockerBuildArtifactMetadata(containerImageNameTags []string, commonPipelineEnvironment *buildkitExecuteCommonPipelineEnvironment) error {
	buildCoordinates := []versioning.Coordinates{}

	pattern := "bom*.xml"

	files, err := filepath.Glob(pattern)
	if err != nil || len(files) == 0 {
		log.Entry().Warnf("no sbom files for build not creating build artifact metadata")
		return nil
	}

	for _, file := range files {
		parentComponent := piperutils.GetComponent(file)
		parentComponentName := parentComponent.Name
		parentComponentVersion := parentComponent.Version

		constructedPurl, err := purl.RefToPURL("docker", fmt.Sprintf("%s:%s", parentComponentName, parentComponentVersion), nil)
		if err != nil {
			log.Entry().Warnf("unable to create purl from reference")
			return nil
		}

		registry, name, version, err := buildkitParsePurl(constructedPurl)
		if err != nil {
			log.Entry().Warnf("unable to parse purl creating build artifact metadata")
			return nil
		}

		constructedPurl, err = purl.RefToPURL("docker", fmt.Sprintf("%s:%s", name, version), nil)
		if err != nil {
			log.Entry().Warnf("unable to create purl from reference")
			return nil
		}

		log.Entry().Debugf("purl is %s", constructedPurl)
		imageNameTag := buildkitFindImageNameTagInPurl(containerImageNameTags, fmt.Sprintf("%s:%s", name, version))
		var coordinate versioning.Coordinates
		if imageNameTag != "" {
			coordinate.ArtifactID = name
			coordinate.BuildPath = filepath.Dir(file)
			coordinate.Version = version
			coordinate.GroupID = ""
			coordinate.PURL = constructedPurl
			coordinate.URL = registry
		} else {
			log.Entry().Warnf("unable to find imageNameTag in purl, not creating build artifact metadata for :%s", file)
			return nil
		}

		err = piperutils.UpdatePurl(file, constructedPurl)
		if err != nil {
			log.Entry().Warnf("unable to update purl in sbom file, hence not creating build artifact metadata for :%s due to err %v", file, err)
			return nil
		}
		buildCoordinates = append(buildCoordinates, coordinate)
	}

	if len(buildCoordinates) > 0 {
		var buildArtifacts build.BuildArtifacts
		buildArtifacts.Coordinates = buildCoordinates
		jsonResult, _ := json.Marshal(buildArtifacts)
		commonPipelineEnvironment.custom.dockerBuildArtifacts = string(jsonResult)
	}

	return nil
}

func buildkitParsePurl(purlStr string) (registry, name, version string, err error) {
	p, err := purlParser.FromString(purlStr)
	if err != nil {
		return "", "", "", err
	}

	namespace := p.Namespace
	if namespace == "" {
		registry = "docker.io"
	} else {
		nsParts := strings.Split(namespace, "/")
		if strings.Contains(nsParts[0], ".") {
			registry = nsParts[0]
		} else {
			registry = "docker.io"
		}
	}

	name = p.Name
	version = p.Version
	return
}

func buildkitFindImageNameTagInPurl(containerImageNameTags []string, purlReference string) string {
	for _, entry := range containerImageNameTags {
		if entry == purlReference {
			log.Entry().Debugf("found image name tag %s in purlReference %s", entry, purlReference)
			return entry
		}
	}

	for _, entry := range containerImageNameTags {
		if strings.HasSuffix(entry, purlReference) {
			log.Entry().Debugf("found suffix match: %s for purlReference %s", entry, purlReference)
			return entry
		}
	}

	log.Entry().Warnf("unable to find image name tag in purlReference '%s' from tags: %v", purlReference, containerImageNameTags)
	return ""
}
