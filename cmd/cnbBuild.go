package cmd

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/buildsettings"
	"github.com/SAP/jenkins-library/pkg/certutils"
	"github.com/SAP/jenkins-library/pkg/cnbutils"
	"github.com/SAP/jenkins-library/pkg/cnbutils/bindings"
	"github.com/SAP/jenkins-library/pkg/cnbutils/privacy"
	"github.com/SAP/jenkins-library/pkg/cnbutils/project"
	"github.com/SAP/jenkins-library/pkg/cnbutils/project/metadata"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/docker"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"

	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/imdario/mergo"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	ignore "github.com/sabhiram/go-gitignore"
)

const (
	creatorPath  = "/cnb/lifecycle/creator"
	platformPath = "/tmp/platform"
)

type pathEnum string

const (
	pathEnumRoot    = pathEnum("root")
	pathEnumFolder  = pathEnum("folder")
	pathEnumArchive = pathEnum("archive")
)

type cnbBuildUtilsBundle struct {
	*command.Command
	*piperutils.Files
	*docker.Client
}

type cnbBuildTelemetry struct {
	dockerImage string
	Version     int                     `json:"version"`
	Data        []cnbBuildTelemetryData `json:"data"`
}

type cnbBuildTelemetryData struct {
	ImageTag          string                                 `json:"imageTag"`
	AdditionalTags    []string                               `json:"additionalTags"`
	BindingKeys       []string                               `json:"bindingKeys"`
	Path              pathEnum                               `json:"path"`
	BuildEnv          cnbBuildTelemetryDataBuildEnv          `json:"buildEnv"`
	Buildpacks        cnbBuildTelemetryDataBuildpacks        `json:"buildpacks"`
	ProjectDescriptor cnbBuildTelemetryDataProjectDescriptor `json:"projectDescriptor"`
	BuildTool         string                                 `json:"buildTool"`
	Builder           string                                 `json:"builder"`
}

type cnbBuildTelemetryDataBuildEnv struct {
	KeysFromConfig            []string               `json:"keysFromConfig"`
	KeysFromProjectDescriptor []string               `json:"keysFromProjectDescriptor"`
	KeysOverall               []string               `json:"keysOverall"`
	JVMVersion                string                 `json:"jvmVersion"`
	KeyValues                 map[string]interface{} `json:"keyValues"`
}

type cnbBuildTelemetryDataBuildpacks struct {
	FromConfig            []string `json:"FromConfig"`
	FromProjectDescriptor []string `json:"FromProjectDescriptor"`
	Overall               []string `json:"overall"`
}

type cnbBuildTelemetryDataProjectDescriptor struct {
	Used        bool `json:"used"`
	IncludeUsed bool `json:"includeUsed"`
	ExcludeUsed bool `json:"excludeUsed"`
}

func processConfigs(main cnbBuildOptions, multipleImages []map[string]interface{}) ([]cnbBuildOptions, error) {
	var result []cnbBuildOptions

	if len(multipleImages) == 0 {
		result = append(result, main)
		return result, nil
	}

	for _, conf := range multipleImages {
		var structuredConf cnbBuildOptions
		err := mapstructure.Decode(conf, &structuredConf)
		if err != nil {
			return nil, err
		}

		err = mergo.Merge(&structuredConf, main)
		if err != nil {
			return nil, err
		}

		result = append(result, structuredConf)
	}

	return result, nil
}

func setCustomBuildpacks(bpacks []string, dockerCreds string, utils cnbutils.BuildUtils) (string, string, error) {
	buildpacksPath := "/tmp/buildpacks"
	orderPath := "/tmp/buildpacks/order.toml"
	newOrder, err := cnbutils.DownloadBuildpacks(buildpacksPath, bpacks, dockerCreds, utils)
	if err != nil {
		return "", "", err
	}

	err = newOrder.Save(orderPath)
	if err != nil {
		return "", "", err
	}

	return buildpacksPath, orderPath, nil
}

func newCnbBuildUtils() cnbutils.BuildUtils {
	utils := cnbBuildUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
		Client:  &docker.Client{},
	}
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func cnbBuild(config cnbBuildOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *cnbBuildCommonPipelineEnvironment) {
	utils := newCnbBuildUtils()

	client := &piperhttp.Client{}

	err := callCnbBuild(&config, telemetryData, utils, commonPipelineEnvironment, client)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func isBuilder(utils cnbutils.BuildUtils) error {
	exists, err := utils.FileExists(creatorPath)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("binary '%s' not found", creatorPath)
	}

	return nil
}

func isZip(path string) bool {
	r, err := zip.OpenReader(path)

	switch {
	case err == nil:
		_ = r.Close()
		return true
	case err == zip.ErrFormat:
		return false
	default:
		return false
	}
}

func cleanDir(dir string, utils cnbutils.BuildUtils) error {
	dirContent, err := utils.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return err
	}

	for _, obj := range dirContent {
		err = utils.RemoveAll(obj)
		if err != nil {
			return err
		}
	}

	return nil
}

func extractZip(source, target string) error {
	if isZip(source) {
		log.Entry().Infof("Extracting archive '%s' to '%s'", source, target)
		_, err := piperutils.Unzip(source, target)
		if err != nil {
			log.SetErrorCategory(log.ErrorBuild)
			return errors.Wrapf(err, "Extracting archive '%s' to '%s' failed", source, target)
		}
	} else {
		log.SetErrorCategory(log.ErrorBuild)
		return errors.New("application path must be a directory or zip")
	}

	return nil
}

func prepareDockerConfig(source string, utils cnbutils.BuildUtils) (string, error) {
	if filepath.Base(source) != "config.json" {
		log.Entry().Debugf("Renaming docker config file from '%s' to 'config.json'", filepath.Base(source))

		newPath := filepath.Join(filepath.Dir(source), "config.json")
		alreadyExists, err := utils.FileExists(newPath)
		if err != nil {
			return "", err
		}
		if alreadyExists {
			return newPath, nil
		}

		err = utils.FileRename(source, newPath)
		if err != nil {
			return "", err
		}

		return newPath, nil
	}

	return source, nil
}

func linkTargetFolder(utils cnbutils.BuildUtils, source, target string) error {
	var err error
	linkPath := filepath.Join(target, "target")
	targetPath := filepath.Join(source, "target")
	if ok, _ := utils.DirExists(targetPath); !ok {
		err = utils.MkdirAll(targetPath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	if ok, _ := utils.DirExists(linkPath); ok {
		err = utils.RemoveAll(linkPath)
		if err != nil {
			return err
		}
	}

	return utils.Symlink(targetPath, linkPath)
}

func (config *cnbBuildOptions) mergeEnvVars(vars map[string]interface{}) {
	if config.BuildEnvVars == nil {
		config.BuildEnvVars = vars

		return
	}

	for k, v := range vars {
		_, exists := config.BuildEnvVars[k]

		if !exists {
			config.BuildEnvVars[k] = v
		}
	}
}

func (config *cnbBuildOptions) resolvePath(utils cnbutils.BuildUtils) (pathEnum, string, error) {
	pwd, err := utils.Getwd()
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		return "", "", errors.Wrap(err, "failed to get current working directory")
	}

	if config.Path == "" {
		return pathEnumRoot, pwd, nil
	}
	matches, err := utils.Glob(config.Path)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return "", "", errors.Wrapf(err, "Failed to resolve glob for '%s'", config.Path)
	}
	numMatches := len(matches)
	if numMatches != 1 {
		log.SetErrorCategory(log.ErrorConfiguration)
		return "", "", errors.Errorf("Failed to resolve glob for '%s', matching %d file(s)", config.Path, numMatches)
	}
	source, err := utils.Abs(matches[0])
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return "", "", errors.Wrapf(err, "Failed to resolve absolute path for '%s'", matches[0])
	}

	dir, err := utils.DirExists(source)
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		return "", "", errors.Wrapf(err, "Checking file info '%s' failed", source)
	}

	if dir {
		return pathEnumFolder, source, nil
	} else {
		return pathEnumArchive, source, nil
	}
}

func addConfigTelemetryData(utils cnbutils.BuildUtils, data *cnbBuildTelemetryData, dockerImage string, config *cnbBuildOptions) {
	var bindingKeys []string
	for k := range config.Bindings {
		bindingKeys = append(bindingKeys, k)
	}
	data.ImageTag = config.ContainerImageTag
	data.AdditionalTags = config.AdditionalTags
	data.BindingKeys = bindingKeys
	data.Path, _, _ = config.resolvePath(utils) // ignore error here, telemetry problems should not fail the build

	configKeys := data.BuildEnv.KeysFromConfig
	overallKeys := data.BuildEnv.KeysOverall
	for key := range config.BuildEnvVars {
		configKeys = append(configKeys, key)
		overallKeys = append(overallKeys, key)
	}
	data.BuildEnv.KeysFromConfig = configKeys
	data.BuildEnv.KeysOverall = overallKeys

	buildTool, _ := getBuildToolFromStageConfig("cnbBuild") // ignore error here, telemetry problems should not fail the build
	data.BuildTool = buildTool

	data.Buildpacks.FromConfig = privacy.FilterBuildpacks(config.Buildpacks)

	data.Builder = privacy.FilterBuilder(dockerImage)
}

func addProjectDescriptorTelemetryData(data *cnbBuildTelemetryData, descriptor project.Descriptor) {
	descriptorKeys := data.BuildEnv.KeysFromProjectDescriptor
	overallKeys := data.BuildEnv.KeysOverall
	for key := range descriptor.EnvVars {
		descriptorKeys = append(descriptorKeys, key)
		overallKeys = append(overallKeys, key)
	}
	data.BuildEnv.KeysFromProjectDescriptor = descriptorKeys
	data.BuildEnv.KeysOverall = overallKeys

	data.Buildpacks.FromProjectDescriptor = privacy.FilterBuildpacks(descriptor.Buildpacks)

	data.ProjectDescriptor.Used = true
	data.ProjectDescriptor.IncludeUsed = descriptor.Include != nil
	data.ProjectDescriptor.ExcludeUsed = descriptor.Exclude != nil
}

func callCnbBuild(config *cnbBuildOptions, telemetryData *telemetry.CustomData, utils cnbutils.BuildUtils, commonPipelineEnvironment *cnbBuildCommonPipelineEnvironment, httpClient piperhttp.Sender) error {
	stepName := "cnbBuild"
	cnbTelemetry := &cnbBuildTelemetry{
		Version: 3,
	}

	dockerImage, err := GetDockerImageValue(stepName)
	if err != nil {
		log.Entry().Warnf("failed to retrieve dockerImage configuration: '%v'", err)
	}

	cnbTelemetry.dockerImage = dockerImage

	cnbBuildConfig := buildsettings.BuildOptions{
		DockerImage:       dockerImage,
		BuildSettingsInfo: config.BuildSettingsInfo,
	}
	log.Entry().Debugf("creating build settings information...")
	buildSettingsInfo, err := buildsettings.CreateBuildSettingsInfo(&cnbBuildConfig, stepName)
	if err != nil {
		log.Entry().Warnf("failed to create build settings info: %v", err)
	}
	commonPipelineEnvironment.custom.buildSettingsInfo = buildSettingsInfo

	mergedConfigs, err := processConfigs(*config, config.MultipleImages)
	if err != nil {
		return errors.Wrap(err, "failed to process config")
	}
	for _, c := range mergedConfigs {
		err = runCnbBuild(&c, cnbTelemetry, utils, commonPipelineEnvironment, httpClient)
		if err != nil {
			return err
		}
	}

	telemetryData.Custom1Label = "cnbBuildStepData"
	customData, err := json.Marshal(cnbTelemetry)
	if err != nil {
		return errors.Wrap(err, "failed to marshal custom telemetry data")
	}
	telemetryData.Custom1 = string(customData)
	return nil
}

func runCnbBuild(config *cnbBuildOptions, cnbTelemetry *cnbBuildTelemetry, utils cnbutils.BuildUtils, commonPipelineEnvironment *cnbBuildCommonPipelineEnvironment, httpClient piperhttp.Sender) error {
	err := cleanDir("/layers", utils)
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		return errors.Wrap(err, "failed to clean up layers folder /layers")
	}

	err = cleanDir(platformPath, utils)
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		return errors.Wrap(err, fmt.Sprintf("failed to clean up platform folder %s", platformPath))
	}

	customTelemetryData := cnbBuildTelemetryData{}
	addConfigTelemetryData(utils, &customTelemetryData, cnbTelemetry.dockerImage, config)

	err = isBuilder(utils)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Wrap(err, "the provided dockerImage is not a valid builder")
	}

	include := ignore.CompileIgnoreLines("**/*")
	exclude := ignore.CompileIgnoreLines("piper", ".pipeline", ".git")

	projDescPath, err := project.ResolvePath(config.ProjectDescriptor, config.Path, utils)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Wrap(err, "failed to check if project descriptor exists")
	}

	var projectID string
	if projDescPath != "" {
		descriptor, err := project.ParseDescriptor(projDescPath, utils, httpClient)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.Wrapf(err, "failed to parse %s", projDescPath)
		}
		addProjectDescriptorTelemetryData(&customTelemetryData, *descriptor)

		config.mergeEnvVars(descriptor.EnvVars)

		if (config.Buildpacks == nil || len(config.Buildpacks) == 0) && len(descriptor.Buildpacks) > 0 {
			config.Buildpacks = descriptor.Buildpacks
		}

		if descriptor.Exclude != nil {
			exclude = descriptor.Exclude
		}

		if descriptor.Include != nil {
			include = descriptor.Include
		}

		projectID = descriptor.ProjectID
	}

	targetImage, err := cnbutils.GetTargetImage(config.ContainerRegistryURL, config.ContainerImageName, config.ContainerImageTag, projectID, GeneralConfig.EnvRootPath)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Wrap(err, "failed to retrieve target image configuration")
	}
	customTelemetryData.Buildpacks.Overall = privacy.FilterBuildpacks(config.Buildpacks)
	customTelemetryData.BuildEnv.KeyValues = privacy.FilterEnv(config.BuildEnvVars)
	cnbTelemetry.Data = append(cnbTelemetry.Data, customTelemetryData)

	if commonPipelineEnvironment.container.imageNameTag == "" {
		commonPipelineEnvironment.container.registryURL = fmt.Sprintf("%s://%s", targetImage.ContainerRegistry.Scheme, targetImage.ContainerRegistry.Host)
		commonPipelineEnvironment.container.imageNameTag = fmt.Sprintf("%v:%v", targetImage.ContainerImageName, targetImage.ContainerImageTag)
	}
	commonPipelineEnvironment.container.imageNameTags = append(commonPipelineEnvironment.container.imageNameTags, fmt.Sprintf("%v:%v", targetImage.ContainerImageName, targetImage.ContainerImageTag))
	commonPipelineEnvironment.container.imageNames = append(commonPipelineEnvironment.container.imageNames, targetImage.ContainerImageName)

	if config.BuildEnvVars != nil && len(config.BuildEnvVars) > 0 {
		log.Entry().Infof("Setting custom environment variables: '%v'", config.BuildEnvVars)
		err = cnbutils.CreateEnvFiles(utils, platformPath, config.BuildEnvVars)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.Wrap(err, "failed to write environment variables to files")
		}
	}

	err = bindings.ProcessBindings(utils, httpClient, platformPath, config.Bindings)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Wrap(err, "failed process bindings")
	}

	dockerConfigFile := ""
	if len(config.DockerConfigJSON) > 0 {
		dockerConfigFile, err = prepareDockerConfig(config.DockerConfigJSON, utils)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.Wrapf(err, "failed to rename DockerConfigJSON file '%v'", config.DockerConfigJSON)
		}
	}

	pathType, source, err := config.resolvePath(utils)
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		return errors.Wrapf(err, "could not resolve path")
	}

	target := "/workspace"
	err = cleanDir(target, utils)
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		return errors.Wrapf(err, "failed to clean up target folder %s", target)
	}

	if pathType != pathEnumArchive {
		err = cnbutils.CopyProject(source, target, include, exclude, utils)
		if err != nil {
			log.SetErrorCategory(log.ErrorBuild)
			return errors.Wrapf(err, "Copying  '%s' into '%s' failed", source, target)
		}
	} else {
		err = extractZip(source, target)
		if err != nil {
			log.SetErrorCategory(log.ErrorBuild)
			return errors.Wrapf(err, "Copying  '%s' into '%s' failed", source, target)
		}
	}

	if ok, _ := utils.FileExists(filepath.Join(target, "pom.xml")); ok {
		err = linkTargetFolder(utils, source, target)
		if err != nil {
			log.SetErrorCategory(log.ErrorBuild)
			return err
		}
	}

	metadata.WriteProjectMetadata(GeneralConfig.EnvRootPath, utils)

	var buildpacksPath = "/cnb/buildpacks"
	var orderPath = "/cnb/order.toml"

	if config.Buildpacks != nil && len(config.Buildpacks) > 0 {
		log.Entry().Infof("Setting custom buildpacks: '%v'", config.Buildpacks)
		buildpacksPath, orderPath, err = setCustomBuildpacks(config.Buildpacks, dockerConfigFile, utils)
		defer utils.RemoveAll(buildpacksPath)
		defer utils.RemoveAll(orderPath)
		if err != nil {
			log.SetErrorCategory(log.ErrorBuild)
			return errors.Wrapf(err, "Setting custom buildpacks: %v", config.Buildpacks)
		}
	}

	cnbRegistryAuth, err := cnbutils.GenerateCnbAuth(dockerConfigFile, utils)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Wrap(err, "failed to generate CNB_REGISTRY_AUTH")
	}

	if len(config.CustomTLSCertificateLinks) > 0 {
		caCertificates := "/tmp/ca-certificates.crt"
		_, err := utils.Copy("/etc/ssl/certs/ca-certificates.crt", caCertificates)
		if err != nil {
			return errors.Wrap(err, "failed to copy certificates")
		}
		err = certutils.CertificateUpdate(config.CustomTLSCertificateLinks, httpClient, utils, caCertificates)
		if err != nil {
			return errors.Wrap(err, "failed to update certificates")
		}
		utils.AppendEnv([]string{fmt.Sprintf("SSL_CERT_FILE=%s", caCertificates)})
	} else {
		log.Entry().Info("skipping certificates update")
	}

	utils.AppendEnv([]string{fmt.Sprintf("CNB_REGISTRY_AUTH=%s", cnbRegistryAuth)})
	utils.AppendEnv([]string{"CNB_PLATFORM_API=0.8"})

	creatorArgs := []string{
		"-no-color",
		"-buildpacks", buildpacksPath,
		"-order", orderPath,
		"-platform", platformPath,
		"-skip-restore",
	}

	if GeneralConfig.Verbose {
		creatorArgs = append(creatorArgs, "-log-level", "debug")
	}

	containerImage := path.Join(targetImage.ContainerRegistry.Host, targetImage.ContainerImageName)
	for _, tag := range config.AdditionalTags {
		target := fmt.Sprintf("%s:%s", containerImage, tag)
		if !piperutils.ContainsString(creatorArgs, target) {
			creatorArgs = append(creatorArgs, "-tag", target)
		}
	}

	creatorArgs = append(creatorArgs, fmt.Sprintf("%s:%s", containerImage, targetImage.ContainerImageTag))
	err = utils.RunExecutable(creatorPath, creatorArgs...)
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		return errors.Wrapf(err, "execution of '%s' failed", creatorArgs)
	}

	digest, err := cnbutils.DigestFromReport(utils)
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		return errors.Wrap(err, "failed to read image digest")
	}
	commonPipelineEnvironment.container.imageDigest = digest
	commonPipelineEnvironment.container.imageDigests = append(commonPipelineEnvironment.container.imageDigests, digest)

	if len(config.PreserveFiles) > 0 {
		if pathType != pathEnumArchive {
			err = cnbutils.CopyProject(target, source, ignore.CompileIgnoreLines(config.PreserveFiles...), nil, utils)
			if err != nil {
				log.SetErrorCategory(log.ErrorBuild)
				return errors.Wrapf(err, "failed to preserve files using glob '%s'", config.PreserveFiles)
			}
		} else {
			log.Entry().Warnf("skipping preserving files because the source '%s' is an archive", source)
		}
	}

	return nil
}
