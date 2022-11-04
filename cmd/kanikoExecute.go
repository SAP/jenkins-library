package cmd

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/SAP/jenkins-library/pkg/buildsettings"
	"github.com/SAP/jenkins-library/pkg/certutils"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/pkg/errors"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/docker"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

const syftURL = "https://raw.githubusercontent.com/anchore/syft/main/install.sh"

type kanikoHttpClient interface {
	piperhttp.Sender
	DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error
}

func installSyft(shellRunner command.ShellRunner, fileUtils piperutils.FileUtils, httpClient kanikoHttpClient) error {
	installationScript := "./install.sh"
	err := httpClient.DownloadFile(syftURL, installationScript, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to download syft: %w", err)
	}

	err = fileUtils.Chmod(installationScript, 0777)
	if err != nil {
		return err
	}

	err = shellRunner.RunShell("/busybox/sh", "cat ./install.sh | sh -s -- -b .")
	if err != nil {
		return fmt.Errorf("failed to install syft: %w", err)
	}

	return nil
}

func kanikoExecute(config kanikoExecuteOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *kanikoExecuteCommonPipelineEnvironment) {
	// for command execution use Command
	c := command.Command{
		ErrorCategoryMapping: map[string][]string{
			log.ErrorConfiguration.String(): {
				"unsupported status code 401",
			},
		},
	}

	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	client := &piperhttp.Client{}

	fileUtils := &piperutils.Files{}

	err := runKanikoExecute(&config, telemetryData, commonPipelineEnvironment, &c, &c, client, fileUtils)
	if err != nil {
		log.Entry().WithError(err).Fatal("Kaniko execution failed")
	}
}

func runKanikoExecute(config *kanikoExecuteOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *kanikoExecuteCommonPipelineEnvironment, execRunner command.ExecRunner, shellRunner command.ShellRunner, httpClient kanikoHttpClient, fileUtils piperutils.FileUtils) error {
	binfmtSupported, _ := docker.IsBinfmtMiscSupportedByHost(fileUtils)

	if !binfmtSupported && len(config.TargetArchitectures) > 0 {
		log.Entry().Warning("Be aware that the host doesn't support binfmt_misc and thus multi archtecture docker builds might not be possible")
	}

	// backward compatibility for parameter ContainerBuildOptions
	if len(config.ContainerBuildOptions) > 0 {
		config.BuildOptions = strings.Split(config.ContainerBuildOptions, " ")
		log.Entry().Warning("Parameter containerBuildOptions is deprecated, please use buildOptions instead.")
		telemetryData.Custom1Label = "ContainerBuildOptions"
		telemetryData.Custom1 = config.ContainerBuildOptions
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

	if !piperutils.ContainsString(config.BuildOptions, "--destination") {
		dest := []string{"--no-push"}
		if len(config.ContainerRegistryURL) > 0 && len(config.ContainerImageName) > 0 && len(config.ContainerImageTag) > 0 {
			containerRegistry, err := docker.ContainerRegistryFromURL(config.ContainerRegistryURL)
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return errors.Wrapf(err, "failed to read registry url %v", config.ContainerRegistryURL)
			}

			commonPipelineEnvironment.container.registryURL = config.ContainerRegistryURL

			// Docker image tags don't allow plus signs in tags, thus replacing with dash
			containerImageTag := strings.ReplaceAll(config.ContainerImageTag, "+", "-")

			if config.ContainerMultiImageBuild {
				log.Entry().Debugf("Multi-image build activated for image name '%v'", config.ContainerImageName)
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
					dest = []string{"--destination", fmt.Sprintf("%v/%v", containerRegistry, containerImageNameAndTag)}
					buildOpts := append(config.BuildOptions, dest...)
					err = runKaniko(file, buildOpts, config.ReadImageDigest, execRunner, fileUtils, commonPipelineEnvironment)
					if err != nil {
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
					//Syft for multi image, generates bom-docker-(1/2/3).xml
					shellRunner.AppendEnv([]string{"DOCKER_CONFIG", "/kaniko/.docker"})
					syftInstallErr := installSyft(shellRunner, fileUtils, httpClient)
					if syftInstallErr != nil {
						return syftInstallErr
					}
					for index, eachImageTag := range commonPipelineEnvironment.container.imageNameTags {
						// TrimPrefix needed as syft needs containerRegistry name only
						syftRunErr := shellRunner.RunShell("/busybox/sh", fmt.Sprintf("./syft %s/%s -o cyclonedx-xml=bom-docker-%v.xml", strings.TrimPrefix(commonPipelineEnvironment.container.registryURL, "https://"), eachImageTag, index))
						if syftRunErr != nil {
							return syftRunErr
						}
					}
				}
				return nil
			} else {
				commonPipelineEnvironment.container.imageNames = append(commonPipelineEnvironment.container.imageNames, config.ContainerImageName)
				commonPipelineEnvironment.container.imageNameTags = append(commonPipelineEnvironment.container.imageNameTags, fmt.Sprintf("%v:%v", config.ContainerImageName, containerImageTag))
			}

			log.Entry().Debugf("Single image build for image name '%v'", config.ContainerImageName)
			containerImageNameAndTag := fmt.Sprintf("%v:%v", config.ContainerImageName, containerImageTag)
			dest = []string{"--destination", fmt.Sprintf("%v/%v", containerRegistry, containerImageNameAndTag)}
			commonPipelineEnvironment.container.imageNameTag = containerImageNameAndTag
		} else if len(config.ContainerImage) > 0 {
			log.Entry().Debugf("Single image build for image '%v'", config.ContainerImage)
			containerRegistry, err := docker.ContainerRegistryFromImage(config.ContainerImage)
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return errors.Wrapf(err, "invalid registry part in image %v", config.ContainerImage)
			}
			// errors are already caught with previous call to docker.ContainerRegistryFromImage
			containerImageName, _ := docker.ContainerImageNameFromImage(config.ContainerImage)
			containerImageNameTag, _ := docker.ContainerImageNameTagFromImage(config.ContainerImage)
			dest = []string{"--destination", config.ContainerImage}
			commonPipelineEnvironment.container.registryURL = fmt.Sprintf("https://%v", containerRegistry)
			commonPipelineEnvironment.container.imageNameTag = containerImageNameTag
			commonPipelineEnvironment.container.imageNameTags = append(commonPipelineEnvironment.container.imageNameTags, containerImageNameTag)
			commonPipelineEnvironment.container.imageNames = append(commonPipelineEnvironment.container.imageNames, containerImageName)
		}
		config.BuildOptions = append(config.BuildOptions, dest...)
	} else {
		log.Entry().Infof("Running Kaniko build with destination defined via buildOptions: %v", config.BuildOptions)

		destination := ""

		for i, o := range config.BuildOptions {
			if o == "--destination" && i+1 < len(config.BuildOptions) {
				destination = config.BuildOptions[i+1]
				break
			}
		}

		containerRegistry, err := docker.ContainerRegistryFromImage(destination)

		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.Wrapf(err, "invalid registry part in image %v", destination)
		}

		containerImageName, _ := docker.ContainerImageNameFromImage(destination)
		containerImageNameTag, _ := docker.ContainerImageNameTagFromImage(destination)

		commonPipelineEnvironment.container.registryURL = fmt.Sprintf("https://%v", containerRegistry)
		commonPipelineEnvironment.container.imageNameTag = containerImageNameTag
		commonPipelineEnvironment.container.imageNameTags = append(commonPipelineEnvironment.container.imageNameTags, containerImageNameTag)
		commonPipelineEnvironment.container.imageNames = append(commonPipelineEnvironment.container.imageNames, containerImageName)
	}

	// no support for building multiple containers
	kanikoErr := runKaniko(config.DockerfilePath, config.BuildOptions, config.ReadImageDigest, execRunner, fileUtils, commonPipelineEnvironment)
	if kanikoErr != nil {
		return kanikoErr
	}
	if config.CreateBOM {
		// Syft for single image, generates bom-docker.xml
		shellRunner.AppendEnv([]string{"DOCKER_CONFIG", "/kaniko/.docker"})
		syftInstallErr := installSyft(shellRunner, fileUtils, httpClient)
		if syftInstallErr != nil {
			return syftInstallErr
		}
		// TrimPrefix of https:// is needed as syft needs the containerRegistry name only
		return shellRunner.RunShell("/busybox/sh", fmt.Sprintf("./syft %s/%s -o cyclonedx-xml=bom-docker.xml", strings.TrimPrefix(commonPipelineEnvironment.container.registryURL, "https://"), commonPipelineEnvironment.container.imageNameTag))
	}
	return nil
}

func runKaniko(dockerFilepath string, buildOptions []string, readDigest bool, execRunner command.ExecRunner, fileUtils piperutils.FileUtils, commonPipelineEnvironment *kanikoExecuteCommonPipelineEnvironment) error {
	cwd, err := fileUtils.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	kanikoOpts := []string{"--dockerfile", dockerFilepath, "--context", cwd}
	kanikoOpts = append(kanikoOpts, buildOptions...)

	tmpDir, err := fileUtils.TempDir("", "*-kanikoExecute")
	if err != nil {
		return fmt.Errorf("failed to create tmp dir for kanikoExecute: %w", err)
	}

	digestFilePath := fmt.Sprintf("%s/digest.txt", tmpDir)

	if readDigest {
		kanikoOpts = append(kanikoOpts, "--digest-file", digestFilePath)
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

		commonPipelineEnvironment.container.imageDigest = string(digestStr)
		commonPipelineEnvironment.container.imageDigests = append(commonPipelineEnvironment.container.imageDigests, digestStr)
	}

	return nil
}
