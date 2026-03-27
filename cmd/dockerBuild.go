package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/shlex"
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
	dockerBuildTLSCertPath    = "/etc/ssl/certs/ca-certificates.crt"
	dockerBuildDaemonJSONPath = "/etc/docker/daemon.json"
)

func dockerBuildGetDockerConfigDir() string {
	home := os.Getenv("HOME")
	if home == "" {
		home = "/root"
	}
	return filepath.Join(home, ".docker")
}

func dockerBuildGetDockerConfigPath() string {
	return filepath.Join(dockerBuildGetDockerConfigDir(), "config.json")
}

func dockerBuild(config dockerBuildOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *dockerBuildCommonPipelineEnvironment) {
	c := command.Command{
		ErrorCategoryMapping: map[string][]string{
			log.ErrorConfiguration.String(): {
				"unsupported status code 401",
			},
		},
		StepName: "dockerBuild",
	}
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	client := &piperhttp.Client{}
	fileUtils := &piperutils.Files{}

	err := runDockerBuild(&config, telemetryData, commonPipelineEnvironment, &c, client, fileUtils)
	if err != nil {
		log.Entry().WithError(err).Fatal("docker build execution failed")
	}
}

func runDockerBuild(config *dockerBuildOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *dockerBuildCommonPipelineEnvironment, execRunner command.ExecRunner, httpClient piperhttp.Sender, fileUtils piperutils.FileUtils) error {
	// backward compatibility for parameter ContainerBuildOptions
	if len(config.ContainerBuildOptions) > 0 {
		parsedOpts, err := shlex.Split(config.ContainerBuildOptions)
		if err != nil {
			return fmt.Errorf("failed to parse containerBuildOptions: %w", err)
		}
		config.BuildOptions = parsedOpts
		log.Entry().Warning("Parameter containerBuildOptions is deprecated, please use buildOptions instead.")
		telemetryData.ContainerBuildOptions = config.ContainerBuildOptions
	}

	// run container preparation command if configured
	if len(config.ContainerPreparationCommand) > 0 {
		prepCommand, err := shlex.Split(config.ContainerPreparationCommand)
		if err != nil {
			return fmt.Errorf("failed to parse container preparation command: %w", err)
		}
		if err := execRunner.RunExecutable(prepCommand[0], prepCommand[1:]...); err != nil {
			return fmt.Errorf("failed to run container preparation command: %w", err)
		}
	}

	if len(config.CustomTLSCertificateLinks) > 0 {
		err := certutils.CertificateUpdate(config.CustomTLSCertificateLinks, httpClient, fileUtils, dockerBuildTLSCertPath)
		if err != nil {
			return fmt.Errorf("failed to update certificates: %w", err)
		}
	} else {
		log.Entry().Info("no custom TLS certificates configured, skipping")
	}

	// Docker config handling
	dockerConfigPath := dockerBuildGetDockerConfigPath()
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
		targetConfigJSON, err := docker.CreateDockerConfigJSON(config.ContainerRegistryURL, config.ContainerRegistryUser, config.ContainerRegistryPassword, "", dockerConfigPath, fileUtils)
		if err != nil {
			return fmt.Errorf("failed to create new docker config json at %s: %w", dockerConfigPath, err)
		}
		dockerConfig, err = fileUtils.FileRead(targetConfigJSON)
		if err != nil {
			return fmt.Errorf("failed to read new docker config file at %s: %w", dockerConfigPath, err)
		}
	}

	if err := fileUtils.FileWrite(dockerConfigPath, dockerConfig, 0600); err != nil {
		return fmt.Errorf("failed to write file '%s': %w", dockerConfigPath, err)
	}

	// Registry mirrors: write daemon.json for Docker daemon
	if len(config.RegistryMirrors) > 0 {
		daemonJSON, err := json.Marshal(map[string]interface{}{
			"registry-mirrors": config.RegistryMirrors,
		})
		if err != nil {
			return fmt.Errorf("failed to marshal daemon.json: %w", err)
		}
		if err := fileUtils.FileWrite(dockerBuildDaemonJSONPath, daemonJSON, 0644); err != nil {
			return fmt.Errorf("failed to write '%s': %w", dockerBuildDaemonJSONPath, err)
		}
		log.Entry().Infof("Registry mirrors configured: %v", config.RegistryMirrors)
	}

	// Build settings
	log.Entry().Debugf("preparing build settings information...")
	stepName := "dockerBuild"
	dockerImage, err := GetDockerImageValue(stepName)
	if err != nil {
		return fmt.Errorf("failed to retrieve dockerImage configuration: %w", err)
	}

	dockerBuildConfig := buildsettings.BuildOptions{
		DockerImage:       dockerImage,
		BuildSettingsInfo: config.BuildSettingsInfo,
	}

	log.Entry().Debugf("creating build settings information...")
	buildSettingsInfo, err := buildsettings.CreateBuildSettingsInfo(&dockerBuildConfig, stepName)
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
			if err = runDockerBuildExec(config, file, dest, true, ".", execRunner, fileUtils, commonPipelineEnvironment); err != nil {
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
			err := syft.GenerateSBOM(config.SyftDownloadURL, dockerBuildGetDockerConfigDir(), execRunner, fileUtils, httpClient, commonPipelineEnvironment.container.registryURL, commonPipelineEnvironment.container.imageNameTags)
			if err != nil {
				return err
			}
		}

		if config.CreateBuildArtifactsMetadata {
			err := dockerBuildCreateArtifactMetadata(commonPipelineEnvironment.container.imageNameTags, commonPipelineEnvironment)
			if err != nil {
				return err
			}
		}
		return nil

	case config.MultipleImages != nil:
		log.Entry().Debugf("multipleImages build activated")
		parsedMultipleImages, err := dockerBuildParseMultipleImages(config.MultipleImages)
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
				if err = runDockerBuildExec(config, dockerfilePath, dest, true, entry.ContextSubPath, execRunner, fileUtils, commonPipelineEnvironment); err != nil {
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

				if err = runDockerBuildExec(config, dockerfilePath, entry.ContainerImage, true, entry.ContextSubPath, execRunner, fileUtils, commonPipelineEnvironment); err != nil {
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
			err := syft.GenerateSBOM(config.SyftDownloadURL, dockerBuildGetDockerConfigDir(), execRunner, fileUtils, httpClient, commonPipelineEnvironment.container.registryURL, commonPipelineEnvironment.container.imageNameTags)
			if err != nil {
				return err
			}
		}

		if config.CreateBuildArtifactsMetadata {
			err := dockerBuildCreateArtifactMetadata(commonPipelineEnvironment.container.imageNameTags, commonPipelineEnvironment)
			if err != nil {
				return err
			}
		}
		return nil

	case dockerBuildHasDestination(config.BuildOptions):
		log.Entry().Infof("Running docker build with destination defined via buildOptions: %v", config.BuildOptions)

		for i, o := range config.BuildOptions {
			if (o == "-t" || o == "--tag") && i+1 < len(config.BuildOptions) {
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
		if err := runDockerBuildExec(config, config.DockerfilePath, "", true, ".", execRunner, fileUtils, commonPipelineEnvironment); err != nil {
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
		if err = runDockerBuildExec(config, config.DockerfilePath, dest, true, ".", execRunner, fileUtils, commonPipelineEnvironment); err != nil {
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

		if err = runDockerBuildExec(config, config.DockerfilePath, config.ContainerImage, true, ".", execRunner, fileUtils, commonPipelineEnvironment); err != nil {
			return err
		}

	default:
		// no-push: build without pushing
		if err := runDockerBuildExec(config, config.DockerfilePath, "", false, ".", execRunner, fileUtils, commonPipelineEnvironment); err != nil {
			return err
		}
	}

	if config.CreateBOM {
		if len(commonPipelineEnvironment.container.imageNameTags) == 0 {
			log.Entry().Warn("skipping BOM creation: no container image was pushed")
		} else {
			err := syft.GenerateSBOM(config.SyftDownloadURL, dockerBuildGetDockerConfigDir(), execRunner, fileUtils, httpClient, commonPipelineEnvironment.container.registryURL, commonPipelineEnvironment.container.imageNameTags)
			if err != nil {
				return err
			}
		}
	}
	if config.CreateBuildArtifactsMetadata {
		if len(commonPipelineEnvironment.container.imageNameTags) == 0 {
			log.Entry().Warn("skipping build artifacts metadata creation: no container image was pushed")
		} else {
			err := dockerBuildCreateArtifactMetadata(commonPipelineEnvironment.container.imageNameTags, commonPipelineEnvironment)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func runDockerBuildExec(config *dockerBuildOptions, dockerfilePath string, destination string, push bool, contextSubPath string, execRunner command.ExecRunner, fileUtils piperutils.FileUtils, commonPipelineEnvironment *dockerBuildCommonPipelineEnvironment) error {
	cwd, err := fileUtils.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	buildContext := cwd
	if contextSubPath != "" && contextSubPath != "." {
		buildContext = filepath.Join(cwd, contextSubPath)
	}

	buildOpts := []string{"buildx", "build", "--file", dockerfilePath}

	if destination != "" {
		buildOpts = append(buildOpts, "-t", destination)
	}
	if push {
		buildOpts = append(buildOpts, "--push")
	}

	tmpDir, err := fileUtils.TempDir("", "*-dockerBuild")
	if err != nil {
		return fmt.Errorf("failed to create tmp dir for dockerBuild: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	metadataFilePath := fmt.Sprintf("%s/metadata.json", tmpDir)

	if config.ReadImageDigest {
		buildOpts = append(buildOpts, "--metadata-file", metadataFilePath)
	}

	if GeneralConfig.Verbose {
		buildOpts = append(buildOpts, "--progress=plain")
	}

	buildOpts = append(buildOpts, config.BuildOptions...)

	// build context must be last argument
	buildOpts = append(buildOpts, buildContext)

	err = execRunner.RunExecutable("docker", buildOpts...)
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

// dockerBuildHasDestination checks if buildOptions contains a -t or --tag flag
func dockerBuildHasDestination(buildOptions []string) bool {
	for _, o := range buildOptions {
		if o == "-t" || o == "--tag" {
			return true
		}
	}
	return false
}

// Duplicated helpers from kanikoExecute (to be refactored to shared package later)

type dockerBuildMultipleImageConf struct {
	ContextSubPath     string `json:"contextSubPath,omitempty"`
	DockerfilePath     string `json:"dockerfilePath,omitempty"`
	ContainerImageName string `json:"containerImageName,omitempty"`
	ContainerImageTag  string `json:"containerImageTag,omitempty"`
	ContainerImage     string `json:"containerImage,omitempty"`
}

func dockerBuildParseMultipleImages(src []map[string]interface{}) ([]dockerBuildMultipleImageConf, error) {
	var result []dockerBuildMultipleImageConf

	for _, conf := range src {
		var structuredConf dockerBuildMultipleImageConf
		if err := mapstructure.Decode(conf, &structuredConf); err != nil {
			return nil, err
		}
		result = append(result, structuredConf)
	}

	return result, nil
}

func dockerBuildCreateArtifactMetadata(containerImageNameTags []string, commonPipelineEnvironment *dockerBuildCommonPipelineEnvironment) error {
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

		registry, name, version, err := dockerBuildParsePurl(constructedPurl)
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
		imageNameTag := dockerBuildFindImageNameTagInPurl(containerImageNameTags, fmt.Sprintf("%s:%s", name, version))
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
		jsonResult, err := json.Marshal(buildArtifacts)
		if err != nil {
			log.Entry().Warnf("failed to marshal build artifacts metadata: %v", err)
			return nil
		}
		commonPipelineEnvironment.custom.dockerBuildArtifacts = string(jsonResult)
	}

	return nil
}

func dockerBuildParsePurl(purlStr string) (registry, name, version string, err error) {
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

func dockerBuildFindImageNameTagInPurl(containerImageNameTags []string, purlReference string) string {
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
