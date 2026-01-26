package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

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

func kanikoExecute(config kanikoExecuteOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *kanikoExecuteCommonPipelineEnvironment) {
	// for command execution use Command
	c := command.Command{
		ErrorCategoryMapping: map[string][]string{
			log.ErrorConfiguration.String(): {
				"unsupported status code 401",
			},
		},
		StepName: "kanikoExecute",
	}

	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	client := &piperhttp.Client{}

	fileUtils := &piperutils.Files{}

	err := runKanikoExecute(&config, telemetryData, commonPipelineEnvironment, &c, client, fileUtils)
	if err != nil {
		log.Entry().WithError(err).Fatal("Kaniko execution failed")
	}
}

func runKanikoExecute(config *kanikoExecuteOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *kanikoExecuteCommonPipelineEnvironment, execRunner command.ExecRunner, httpClient piperhttp.Sender, fileUtils piperutils.FileUtils) error {
	binfmtSupported, _ := docker.IsBinfmtMiscSupportedByHost(fileUtils)

	if !binfmtSupported && len(config.TargetArchitectures) > 0 {
		log.Entry().Warning("Be aware that the host doesn't support binfmt_misc and thus multi archtecture docker builds might not be possible")
	}

	// backward compatibility for parameter ContainerBuildOptions
	if len(config.ContainerBuildOptions) > 0 {
		config.BuildOptions = strings.Split(config.ContainerBuildOptions, " ")
		log.Entry().Warning("Parameter containerBuildOptions is deprecated, please use buildOptions instead.")
		telemetryData.ContainerBuildOptions = config.ContainerBuildOptions
	}

	// prepare kaniko container for running with proper Docker config.json and custom certificates
	// custom certificates will be downloaded and appended to ca-certificates.crt file used in container
	if len(config.ContainerPreparationCommand) > 0 {
		prepCommand := strings.Split(config.ContainerPreparationCommand, " ")
		if err := execRunner.RunExecutable(prepCommand[0], prepCommand[1:]...); err != nil {
			return errors.Wrap(err, "failed to initialize Kaniko container")
		}
	}

	if len(config.CustomTLSCertificateLinks) > 0 {
		err := certutils.CertificateUpdate(config.CustomTLSCertificateLinks, httpClient, fileUtils, "/kaniko/ssl/certs/ca-certificates.crt")
		if err != nil {
			return errors.Wrap(err, "failed to update certificates")
		}
	} else {
		log.Entry().Info("skipping updation of certificates")
	}

	dockerConfig := []byte(`{"auths":{}}`)

	// respect user provided docker config json file
	if len(config.DockerConfigJSON) > 0 {
		var err error
		dockerConfig, err = fileUtils.FileRead(config.DockerConfigJSON)
		if err != nil {
			return errors.Wrapf(err, "failed to read existing docker config json at '%v'", config.DockerConfigJSON)
		}
	}

	// if : user provided docker config json and registry credentials present then enahance the user provided docker provided json with the registry credentials
	// else if : no user provided docker config json then create a new docker config json for kaniko
	if len(config.DockerConfigJSON) > 0 && len(config.ContainerRegistryURL) > 0 && len(config.ContainerRegistryPassword) > 0 && len(config.ContainerRegistryUser) > 0 {
		targetConfigJson, err := docker.CreateDockerConfigJSON(config.ContainerRegistryURL, config.ContainerRegistryUser, config.ContainerRegistryPassword, "", config.DockerConfigJSON, fileUtils)
		if err != nil {
			return errors.Wrapf(err, "failed to update existing docker config json file '%v'", config.DockerConfigJSON)
		}

		dockerConfig, err = fileUtils.FileRead(targetConfigJson)
		if err != nil {
			return errors.Wrapf(err, "failed to read enhanced file '%v'", config.DockerConfigJSON)
		}
	} else if len(config.DockerConfigJSON) == 0 && len(config.ContainerRegistryURL) > 0 && len(config.ContainerRegistryPassword) > 0 && len(config.ContainerRegistryUser) > 0 {
		targetConfigJson, err := docker.CreateDockerConfigJSON(config.ContainerRegistryURL, config.ContainerRegistryUser, config.ContainerRegistryPassword, "", "/kaniko/.docker/config.json", fileUtils)
		if err != nil {
			return errors.Wrap(err, "failed to create new docker config json at /kaniko/.docker/config.json")
		}

		dockerConfig, err = fileUtils.FileRead(targetConfigJson)
		if err != nil {
			return errors.Wrapf(err, "failed to read new docker config file at /kaniko/.docker/config.json")
		}
	}

	if err := fileUtils.FileWrite("/kaniko/.docker/config.json", dockerConfig, 0644); err != nil {
		return errors.Wrap(err, "failed to write file '/kaniko/.docker/config.json'")
	}

	log.Entry().Debugf("preparing build settings information...")
	stepName := "kanikoExecute"
	// ToDo: better testability required. So far retrieval of config is rather non deterministic
	dockerImage, err := GetDockerImageValue(stepName)
	if err != nil {
		return fmt.Errorf("failed to retrieve dockerImage configuration: %w", err)
	}

	kanikoConfig := buildsettings.BuildOptions{
		DockerImage:       dockerImage,
		BuildSettingsInfo: config.BuildSettingsInfo,
	}

	log.Entry().Debugf("creating build settings information...")
	buildSettingsInfo, err := buildsettings.CreateBuildSettingsInfo(&kanikoConfig, stepName)
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
			return errors.Wrapf(err, "failed to read registry url %v", config.ContainerRegistryURL)
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
			buildOpts := append(config.BuildOptions, "--destination", fmt.Sprintf("%v/%v", containerRegistry, containerImageNameAndTag))
			if err = runKaniko(file, buildOpts, config.ReadImageDigest, execRunner, fileUtils, commonPipelineEnvironment); err != nil {
				return fmt.Errorf("failed to build image '%v' using '%v': %w", image, file, err)
			}
			commonPipelineEnvironment.container.imageNames = append(commonPipelineEnvironment.container.imageNames, image)
			commonPipelineEnvironment.container.imageNameTags = append(commonPipelineEnvironment.container.imageNameTags, containerImageNameAndTag)
		}

		// for compatibility reasons also fill single imageNameTag field with "root" image in commonPipelineEnvironment
		// only consider if it has been built
		// ToDo: reconsider and possibly remove at a later point
		if len(imageListWithFilePath[config.ContainerImageName]) > 0 {
			containerImageNameAndTag := fmt.Sprintf("%v:%v", config.ContainerImageName, containerImageTag)
			commonPipelineEnvironment.container.imageNameTag = containerImageNameAndTag
		}
		if config.CreateBOM {
			// Syft for multi image, generates bom-docker-(1/2/3).xml
			err := syft.GenerateSBOM(config.SyftDownloadURL, "/kaniko/.docker", execRunner, fileUtils, httpClient, commonPipelineEnvironment.container.registryURL, commonPipelineEnvironment.container.imageNameTags)
			if err != nil {
				return err
			}
		}

		if config.CreateBuildArtifactsMetadata {
			err := createDockerBuildArtifactMetadata(commonPipelineEnvironment.container.imageNameTags, commonPipelineEnvironment)
			if err != nil {
				return err
			}
		}
		return nil

	case config.MultipleImages != nil:
		log.Entry().Debugf("multipleImages build activated")
		parsedMultipleImages, err := parseMultipleImages(config.MultipleImages)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.Wrap(err, "failed to parse multipleImages param")
		}

		for _, entry := range parsedMultipleImages {
			switch {
			case entry.ContextSubPath == "":
				return fmt.Errorf("multipleImages: empty contextSubPath")
			case entry.ContainerImageName != "":
				containerRegistry, err := docker.ContainerRegistryFromURL(config.ContainerRegistryURL)
				if err != nil {
					log.SetErrorCategory(log.ErrorConfiguration)
					return errors.Wrapf(err, "multipleImages: failed to read registry url %v", config.ContainerRegistryURL)
				}

				if entry.ContainerImageTag == "" {
					if config.ContainerImageTag == "" {
						return fmt.Errorf("both multipleImages containerImageTag and config.containerImageTag are empty")
					}
					entry.ContainerImageTag = config.ContainerImageTag
				}
				// Docker image tags don't allow plus signs in tags, thus replacing with dash
				containerImageTag := strings.ReplaceAll(entry.ContainerImageTag, "+", "-")
				containerImageNameAndTag := fmt.Sprintf("%v:%v", entry.ContainerImageName, containerImageTag)

				log.Entry().Debugf("multipleImages: image build '%v'", entry.ContainerImageName)

				buildOptions := append(config.BuildOptions,
					"--context-sub-path", entry.ContextSubPath,
					"--destination", fmt.Sprintf("%v/%v", containerRegistry, containerImageNameAndTag),
				)

				dockerfilePath := config.DockerfilePath
				if entry.DockerfilePath != "" {
					dockerfilePath = entry.DockerfilePath
				}

				if err = runKaniko(dockerfilePath, buildOptions, config.ReadImageDigest, execRunner, fileUtils, commonPipelineEnvironment); err != nil {
					return fmt.Errorf("multipleImages: failed to build image '%v' using '%v': %w", entry.ContainerImageName, config.DockerfilePath, err)
				}

				commonPipelineEnvironment.container.imageNameTags = append(commonPipelineEnvironment.container.imageNameTags, containerImageNameAndTag)
				commonPipelineEnvironment.container.imageNames = append(commonPipelineEnvironment.container.imageNames, entry.ContainerImageName)

			case entry.ContainerImage != "":
				containerImageName, err := docker.ContainerImageNameFromImage(entry.ContainerImage)
				if err != nil {
					log.SetErrorCategory(log.ErrorConfiguration)
					return errors.Wrapf(err, "invalid name part in image %v", entry.ContainerImage)
				}
				containerImageNameTag, err := docker.ContainerImageNameTagFromImage(entry.ContainerImage)
				if err != nil {
					log.SetErrorCategory(log.ErrorConfiguration)
					return errors.Wrapf(err, "invalid tag part in image %v", entry.ContainerImage)
				}

				log.Entry().Debugf("multipleImages: image build '%v'", containerImageName)

				buildOptions := append(config.BuildOptions,
					"--context-sub-path", entry.ContextSubPath,
					"--destination", entry.ContainerImage,
				)

				dockerfilePath := config.DockerfilePath
				if entry.DockerfilePath != "" {
					dockerfilePath = entry.DockerfilePath
				}

				if err = runKaniko(dockerfilePath, buildOptions, config.ReadImageDigest, execRunner, fileUtils, commonPipelineEnvironment); err != nil {
					return fmt.Errorf("multipleImages: failed to build image '%v' using '%v': %w", containerImageName, config.DockerfilePath, err)
				}

				commonPipelineEnvironment.container.imageNameTags = append(commonPipelineEnvironment.container.imageNameTags, containerImageNameTag)
				commonPipelineEnvironment.container.imageNames = append(commonPipelineEnvironment.container.imageNames, containerImageName)
			default:
				return fmt.Errorf("multipleImages: either containerImageName or containerImage must be filled")
			}
		}

		// Docker image tags don't allow plus signs in tags, thus replacing with dash
		containerImageTag := strings.ReplaceAll(config.ContainerImageTag, "+", "-")

		// for compatibility reasons also fill single imageNameTag field with "root" image in commonPipelineEnvironment
		containerImageNameAndTag := fmt.Sprintf("%v:%v", config.ContainerImageName, containerImageTag)
		commonPipelineEnvironment.container.imageNameTag = containerImageNameAndTag
		commonPipelineEnvironment.container.registryURL = config.ContainerRegistryURL

		if config.CreateBOM {
			// Syft for multi image, generates bom-docker-(1/2/3).xml
			err := syft.GenerateSBOM(config.SyftDownloadURL, "/kaniko/.docker", execRunner, fileUtils, httpClient, commonPipelineEnvironment.container.registryURL, commonPipelineEnvironment.container.imageNameTags)
			if err != nil {
				return err
			}
		}

		if config.CreateBuildArtifactsMetadata {
			err := createDockerBuildArtifactMetadata(commonPipelineEnvironment.container.imageNameTags, commonPipelineEnvironment)
			if err != nil {
				return err
			}
		}
		return nil

	case slices.Contains(config.BuildOptions, "--destination"):
		log.Entry().Infof("Running Kaniko build with destination defined via buildOptions: %v", config.BuildOptions)

		for i, o := range config.BuildOptions {
			if o == "--destination" && i+1 < len(config.BuildOptions) {
				destination := config.BuildOptions[i+1]

				containerRegistry, err := docker.ContainerRegistryFromImage(destination)
				if err != nil {
					log.SetErrorCategory(log.ErrorConfiguration)
					return errors.Wrapf(err, "invalid registry part in image %v", destination)
				}
				if commonPipelineEnvironment.container.registryURL == "" {
					commonPipelineEnvironment.container.registryURL = fmt.Sprintf("https://%v", containerRegistry)
				}

				// errors are already caught with previous call to docker.ContainerRegistryFromImage
				containerImageName, _ := docker.ContainerImageNameFromImage(destination)
				containerImageNameTag, _ := docker.ContainerImageNameTagFromImage(destination)

				if commonPipelineEnvironment.container.imageNameTag == "" {
					commonPipelineEnvironment.container.imageNameTag = containerImageNameTag
				}
				commonPipelineEnvironment.container.imageNameTags = append(commonPipelineEnvironment.container.imageNameTags, containerImageNameTag)
				commonPipelineEnvironment.container.imageNames = append(commonPipelineEnvironment.container.imageNames, containerImageName)
			}
		}

	case config.ContainerRegistryURL != "" && config.ContainerImageName != "" && config.ContainerImageTag != "":
		log.Entry().Debugf("Single image build for image name '%v'", config.ContainerImageName)

		containerRegistry, err := docker.ContainerRegistryFromURL(config.ContainerRegistryURL)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.Wrapf(err, "failed to read registry url %v", config.ContainerRegistryURL)
		}

		// Docker image tags don't allow plus signs in tags, thus replacing with dash
		containerImageTag := strings.ReplaceAll(config.ContainerImageTag, "+", "-")
		containerImageNameAndTag := fmt.Sprintf("%v:%v", config.ContainerImageName, containerImageTag)

		commonPipelineEnvironment.container.registryURL = config.ContainerRegistryURL
		commonPipelineEnvironment.container.imageNameTag = containerImageNameAndTag
		commonPipelineEnvironment.container.imageNameTags = append(commonPipelineEnvironment.container.imageNameTags, containerImageNameAndTag)
		commonPipelineEnvironment.container.imageNames = append(commonPipelineEnvironment.container.imageNames, config.ContainerImageName)
		config.BuildOptions = append(config.BuildOptions, "--destination", fmt.Sprintf("%v/%v", containerRegistry, containerImageNameAndTag))

	case config.ContainerImage != "":
		log.Entry().Debugf("Single image build for image '%v'", config.ContainerImage)

		containerRegistry, err := docker.ContainerRegistryFromImage(config.ContainerImage)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.Wrapf(err, "invalid registry part in image %v", config.ContainerImage)
		}

		// errors are already caught with previous call to docker.ContainerRegistryFromImage
		containerImageName, _ := docker.ContainerImageNameFromImage(config.ContainerImage)
		containerImageNameTag, _ := docker.ContainerImageNameTagFromImage(config.ContainerImage)

		commonPipelineEnvironment.container.registryURL = fmt.Sprintf("https://%v", containerRegistry)
		commonPipelineEnvironment.container.imageNameTag = containerImageNameTag
		commonPipelineEnvironment.container.imageNameTags = append(commonPipelineEnvironment.container.imageNameTags, containerImageNameTag)
		commonPipelineEnvironment.container.imageNames = append(commonPipelineEnvironment.container.imageNames, containerImageName)
		config.BuildOptions = append(config.BuildOptions, "--destination", config.ContainerImage)
	default:
		config.BuildOptions = append(config.BuildOptions, "--no-push")
	}

	if err = runKaniko(config.DockerfilePath, config.BuildOptions, config.ReadImageDigest, execRunner, fileUtils, commonPipelineEnvironment); err != nil {
		return err
	}

	if config.CreateBOM {
		// Syft for single image, generates bom-docker-0.xml
		err := syft.GenerateSBOM(config.SyftDownloadURL, "/kaniko/.docker", execRunner, fileUtils, httpClient, commonPipelineEnvironment.container.registryURL, commonPipelineEnvironment.container.imageNameTags)
		if err != nil {
			return err
		}
	}
	if config.CreateBuildArtifactsMetadata {
		err := createDockerBuildArtifactMetadata(commonPipelineEnvironment.container.imageNameTags, commonPipelineEnvironment)
		if err != nil {
			return err
		}
	}

	return nil
}

func runKaniko(dockerFilepath string, buildOptions []string, readDigest bool, execRunner command.ExecRunner, fileUtils piperutils.FileUtils, commonPipelineEnvironment *kanikoExecuteCommonPipelineEnvironment) error {
	cwd, err := fileUtils.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// kaniko build context needs a proper prefix, for local directory it is 'dir://'
	// for more details see https://github.com/GoogleContainerTools/kaniko#kaniko-build-contexts
	kanikoOpts := []string{"--dockerfile", dockerFilepath, "--context", "dir://" + cwd}
	kanikoOpts = append(kanikoOpts, buildOptions...)

	tmpDir, err := fileUtils.TempDir("", "*-kanikoExecute")
	if err != nil {
		return fmt.Errorf("failed to create tmp dir for kanikoExecute: %w", err)
	}

	digestFilePath := fmt.Sprintf("%s/digest.txt", tmpDir)

	if readDigest {
		kanikoOpts = append(kanikoOpts, "--digest-file", digestFilePath)
	}

	if GeneralConfig.Verbose {
		kanikoOpts = append(kanikoOpts, "--verbosity=debug")
	}

	err = execRunner.RunExecutable("/kaniko/executor", kanikoOpts...)
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		return errors.Wrap(err, "execution of '/kaniko/executor' failed")
	}

	if b, err := fileUtils.FileExists(digestFilePath); err == nil && b {
		digest, err := fileUtils.FileRead(digestFilePath)
		if err != nil {
			return errors.Wrap(err, "error while reading image digest")
		}

		digestStr := string(digest)

		log.Entry().Debugf("image digest: %s", digestStr)

		commonPipelineEnvironment.container.imageDigest = digestStr
		commonPipelineEnvironment.container.imageDigests = append(commonPipelineEnvironment.container.imageDigests, digestStr)
	}

	return nil
}

func createDockerBuildArtifactMetadata(containerImageNameTags []string, commonPipelineEnvironment *kanikoExecuteCommonPipelineEnvironment) error {
	buildCoordinates := []versioning.Coordinates{}

	// for docker the logic will be slighlty different since we need to co-relate the sbom generated to the actual built docker images
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

		// syft sbom do not contain purl for the parent component
		// this is problem since the way we tie back promoted artifact to build
		// is only via the sbom parent component , until the time https://github.com/anchore/syft/issues/1408
		// is fixed we are generating a purl and inserting it into the sbom
		constructedPurl, err := purl.RefToPURL("docker", fmt.Sprintf("%s:%s", parentComponentName, parentComponentVersion), nil)
		if err != nil {
			log.Entry().Warnf("unable to create purl from reference")
			return nil
		}

		// this purl contains the docker registry and we remove that from the final purl
		// and recreate the purl without the registry, and we dont want to expose that
		registry, name, version, err := parsePurl(constructedPurl)
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
		imageNameTag := findImageNameTagInPurl(containerImageNameTags, fmt.Sprintf("%s:%s", name, version))
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

func parsePurl(purlStr string) (registry, name, version string, err error) {
	p, err := purlParser.FromString(purlStr)
	if err != nil {
		return "", "", "", err
	}

	// Split namespace to extract registry
	// E.g., namespace = "ghcr.io/my-org"
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

func findImageNameTagInPurl(containerImageNameTags []string, purlReference string) string {
	// check for exact matches
	for _, entry := range containerImageNameTags {
		if entry == purlReference {
			log.Entry().Debugf("found image name tag %s in purlReference %s", entry, purlReference)
			return entry
		}
	}

	// check for suffix matches
	for _, entry := range containerImageNameTags {
		if strings.HasSuffix(entry, purlReference) {
			log.Entry().Debugf("found suffix match: %s for purlReference %s", entry, purlReference)
			return entry
		}
	}

	log.Entry().Warnf("unable to find image name tag in purlReference '%s' from tags: %v", purlReference, containerImageNameTags)
	return ""
}

type multipleImageConf struct {
	ContextSubPath     string `json:"contextSubPath,omitempty"`
	DockerfilePath     string `json:"dockerfilePath,omitempty"`
	ContainerImageName string `json:"containerImageName,omitempty"`
	ContainerImageTag  string `json:"containerImageTag,omitempty"`
	ContainerImage     string `json:"containerImage,omitempty"`
}

func parseMultipleImages(src []map[string]interface{}) ([]multipleImageConf, error) {
	var result []multipleImageConf

	for _, conf := range src {
		var structuredConf multipleImageConf
		if err := mapstructure.Decode(conf, &structuredConf); err != nil {
			return nil, err
		}

		result = append(result, structuredConf)
	}

	return result, nil
}
