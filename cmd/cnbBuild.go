package cmd

import (
	"archive/zip"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/buildpacks"
	"github.com/SAP/jenkins-library/pkg/buildsettings"
	"github.com/SAP/jenkins-library/pkg/certutils"
	"github.com/SAP/jenkins-library/pkg/cnbutils"
	"github.com/SAP/jenkins-library/pkg/cnbutils/bindings"
	"github.com/SAP/jenkins-library/pkg/cnbutils/project"
	"github.com/SAP/jenkins-library/pkg/cnbutils/project/metadata"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/docker"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/syft"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/imdario/mergo"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	ignore "github.com/sabhiram/go-gitignore"
)

const (
	creatorPath        = "/cnb/lifecycle/creator"
	platformPath       = "/tmp/platform"
	platformAPIVersion = "0.12"
)

type cnbBuildUtilsBundle struct {
	*command.Command
	*piperutils.Files
	*docker.Client
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

func setCustomBuildpacks(bpacks, preBuildpacks, postBuildpacks []string, dockerCreds string, utils cnbutils.BuildUtils) (string, string, error) {
	buildpacksPath := "/tmp/buildpacks"
	orderPath := "/tmp/buildpacks/order.toml"
	err := cnbutils.DownloadBuildpacks(buildpacksPath, append(bpacks, append(preBuildpacks, postBuildpacks...)...), dockerCreds, utils)
	if err != nil {
		return "", "", err
	}

	if len(bpacks) == 0 && (len(postBuildpacks) > 0 || len(preBuildpacks) > 0) {
		matches, err := utils.Glob("/cnb/buildpacks/*")
		if err != nil {
			return "", "", err
		}

		for _, match := range matches {
			err = cnbutils.CreateVersionSymlinks(buildpacksPath, match, utils)
			if err != nil {
				return "", "", err
			}
		}
	}

	newOrder, err := cnbutils.CreateOrder(bpacks, preBuildpacks, postBuildpacks, dockerCreds, utils)
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
		Command: &command.Command{
			StepName: "cnbBuild",
		},
		Files:  &piperutils.Files{},
		Client: &docker.Client{},
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

func ensureDockerConfig(config *cnbBuildOptions, utils cnbutils.BuildUtils) error {
	newFile := "/tmp/config.json"
	if config.DockerConfigJSON == "" {
		config.DockerConfigJSON = newFile

		return utils.FileWrite(config.DockerConfigJSON, []byte("{}"), os.ModePerm)
	}

	log.Entry().Debugf("Copying docker config file from '%s' to '%s'", config.DockerConfigJSON, newFile)
	_, err := utils.Copy(config.DockerConfigJSON, newFile)
	if err != nil {
		return err
	}

	err = utils.Chmod(newFile, 0644)
	if err != nil {
		return err
	}

	config.DockerConfigJSON = newFile
	return nil
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

func (config *cnbBuildOptions) resolvePath(utils cnbutils.BuildUtils) (buildpacks.PathEnum, string, error) {
	pwd, err := utils.Getwd()
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		return "", "", errors.Wrap(err, "failed to get current working directory")
	}

	if config.Path == "" {
		return buildpacks.PathEnumRoot, pwd, nil
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
		return buildpacks.PathEnumFolder, source, nil
	} else {
		return buildpacks.PathEnumArchive, source, nil
	}
}

func callCnbBuild(config *cnbBuildOptions, telemetryData *telemetry.CustomData, utils cnbutils.BuildUtils, commonPipelineEnvironment *cnbBuildCommonPipelineEnvironment, httpClient piperhttp.Sender) error {
	stepName := "cnbBuild"

	err := isBuilder(utils)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Wrap(err, "the provided dockerImage is not a valid builder")
	}

	telemetry := buildpacks.NewTelemetry(telemetryData)
	dockerImage, err := GetDockerImageValue(stepName)
	if err != nil {
		log.Entry().Warnf("failed to retrieve dockerImage configuration: '%v'", err)
	}
	telemetry.WithBuilder(dockerImage)

	buildTool, _ := getBuildToolFromStageConfig("cnbBuild")
	telemetry.WithBuildTool(buildTool)

	cnbBuildConfig := buildsettings.BuildOptions{
		CreateBOM:         config.CreateBOM,
		DockerImage:       dockerImage,
		BuildSettingsInfo: config.BuildSettingsInfo,
	}
	log.Entry().Debugf("creating build settings information...")
	buildSettingsInfo, err := buildsettings.CreateBuildSettingsInfo(&cnbBuildConfig, stepName)
	if err != nil {
		log.Entry().Warnf("failed to create build settings info: %v", err)
	}
	commonPipelineEnvironment.custom.buildSettingsInfo = buildSettingsInfo

	err = ensureDockerConfig(config, utils)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Wrapf(err, "failed to create/rename DockerConfigJSON file")
	}

	if config.DockerConfigJSONCPE != "" {
		log.Entry().Debugf("merging docker config file '%s' into '%s'", config.DockerConfigJSONCPE, config.DockerConfigJSON)
		err = docker.MergeDockerConfigJSON(config.DockerConfigJSONCPE, config.DockerConfigJSON, utils)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.Wrapf(err, "failed to merge DockerConfigJSON files")
		}
	}

	mergedConfigs, err := processConfigs(*config, config.MultipleImages)
	if err != nil {
		return errors.Wrap(err, "failed to process config")
	}

	buildSummary := cnbutils.NewBuildSummary(dockerImage, utils)
	for _, c := range mergedConfigs {
		imageSummary := &cnbutils.ImageSummary{}
		err = runCnbBuild(&c, telemetry, imageSummary, utils, commonPipelineEnvironment, httpClient)
		if err != nil {
			return err
		}
		buildSummary.Images = append(buildSummary.Images, imageSummary)
	}

	buildSummary.Print()

	if config.CreateBOM {
		err = syft.GenerateSBOM(config.SyftDownloadURL, filepath.Dir(config.DockerConfigJSON), utils, utils, httpClient, commonPipelineEnvironment.container.registryURL, commonPipelineEnvironment.container.imageNameTags)
		if err != nil {
			log.SetErrorCategory(log.ErrorCompliance)
			return errors.Wrap(err, "failed to create BOM file")
		}
	}

	return nil
}

func runCnbBuild(config *cnbBuildOptions, telemetry *buildpacks.Telemetry, imageSummary *cnbutils.ImageSummary, utils cnbutils.BuildUtils, commonPipelineEnvironment *cnbBuildCommonPipelineEnvironment, httpClient piperhttp.Sender) error {
	telemetry.WithRunImage(config.RunImage)

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

	tempdir, err := utils.TempDir("", "cnbBuild-")
	if err != nil {
		return errors.Wrap(err, "failed to create tempdir")
	}
	defer utils.RemoveAll(tempdir)

	uid, gid, err := cnbutils.CnbUserInfo()
	if err != nil {
		return errors.Wrap(err, "failed to get user information")
	}

	err = utils.Chown(tempdir, uid, gid)
	if err != nil {
		return errors.Wrap(err, "failed to change tempdir ownership")
	}

	if config.BuildEnvVars == nil {
		config.BuildEnvVars = map[string]interface{}{}
	}
	config.BuildEnvVars["TMPDIR"] = tempdir

	include := ignore.CompileIgnoreLines("**/*")
	exclude := ignore.CompileIgnoreLines("piper", ".pipeline", ".git")

	projDescPath, err := project.ResolvePath(config.ProjectDescriptor, config.Path, utils)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Wrap(err, "failed to check if project descriptor exists")
	}
	imageSummary.ProjectDescriptor = projDescPath

	var projectID string
	if projDescPath != "" {
		descriptor, err := project.ParseDescriptor(projDescPath, utils, httpClient)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.Wrapf(err, "failed to parse %s", projDescPath)
		}

		config.mergeEnvVars(descriptor.EnvVars)

		if len(config.Buildpacks) == 0 {
			config.Buildpacks = descriptor.Buildpacks
		}

		if len(config.PreBuildpacks) == 0 {
			config.PreBuildpacks = descriptor.PreBuildpacks
		}

		if len(config.PostBuildpacks) == 0 {
			config.PostBuildpacks = descriptor.PostBuildpacks
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

	if commonPipelineEnvironment.container.imageNameTag == "" {
		commonPipelineEnvironment.container.registryURL = fmt.Sprintf("%s://%s", targetImage.ContainerRegistry.Scheme, targetImage.ContainerRegistry.Host)
		commonPipelineEnvironment.container.imageNameTag = fmt.Sprintf("%v:%v", targetImage.ContainerImageName, targetImage.ContainerImageTag)
	}
	commonPipelineEnvironment.container.imageNameTags = append(commonPipelineEnvironment.container.imageNameTags, fmt.Sprintf("%v:%v", targetImage.ContainerImageName, targetImage.ContainerImageTag))
	imageNameAlias := targetImage.ContainerImageName
	if config.ContainerImageAlias != "" {
		imageNameAlias = config.ContainerImageAlias
	}
	commonPipelineEnvironment.container.imageNames = append(commonPipelineEnvironment.container.imageNames, imageNameAlias)

	if config.ExpandBuildEnvVars {
		config.BuildEnvVars = expandEnvVars(config.BuildEnvVars)
	}

	if config.BuildEnvVars != nil && len(config.BuildEnvVars) > 0 {
		log.Entry().Infof("Setting custom environment variables: '%v'", config.BuildEnvVars)
		imageSummary.AddEnv(config.BuildEnvVars)
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

	if pathType != buildpacks.PathEnumArchive {
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

	if err := utils.Chown(target, uid, gid); err != nil {
		return err
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
	var orderPath = cnbutils.DefaultOrderPath

	if len(config.Buildpacks) > 0 || len(config.PreBuildpacks) > 0 || len(config.PostBuildpacks) > 0 {
		log.Entry().Infof("Setting custom buildpacks: '%v'", config.Buildpacks)
		log.Entry().Infof("Pre-buildpacks: '%v'", config.PreBuildpacks)
		log.Entry().Infof("Post-buildpacks: '%v'", config.PostBuildpacks)
		buildpacksPath, orderPath, err = setCustomBuildpacks(config.Buildpacks, config.PreBuildpacks, config.PostBuildpacks, config.DockerConfigJSON, utils)
		defer func() { _ = utils.RemoveAll(buildpacksPath) }()
		defer func() { _ = utils.RemoveAll(orderPath) }()
		if err != nil {
			log.SetErrorCategory(log.ErrorBuild)
			return errors.Wrapf(err, "Setting custom buildpacks: %v", config.Buildpacks)
		}
	}

	cnbRegistryAuth, err := cnbutils.GenerateCnbAuth(config.DockerConfigJSON, utils)
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
	utils.AppendEnv([]string{fmt.Sprintf("CNB_PLATFORM_API=%s", platformAPIVersion)})

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

	if config.RunImage != "" {
		creatorArgs = append(creatorArgs, "-run-image", config.RunImage)
	}

	if config.DefaultProcess != "" {
		creatorArgs = append(creatorArgs, "-process-type", config.DefaultProcess)
	}

	containerImage := path.Join(targetImage.ContainerRegistry.Host, targetImage.ContainerImageName)
	for _, tag := range config.AdditionalTags {
		target := fmt.Sprintf("%s:%s", containerImage, tag)
		if !piperutils.ContainsString(creatorArgs, target) {
			creatorArgs = append(creatorArgs, "-tag", target)
		}
	}

	creatorArgs = append(creatorArgs, fmt.Sprintf("%s:%s", containerImage, targetImage.ContainerImageTag))
	attr := getSysProcAttr(uid, gid)

	err = utils.RunExecutableWithAttrs(creatorPath, attr, creatorArgs...)
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
	imageSummary.ImageRef = fmt.Sprintf("%s@%s", containerImage, digest)

	if len(config.PreserveFiles) > 0 {
		if pathType != buildpacks.PathEnumArchive {
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

func expandEnvVars(envVars map[string]any) map[string]any {
	expandedEnvVars := map[string]any{}
	for key, value := range envVars {
		valueString, valueIsString := value.(string)
		if valueIsString {
			expandedEnvVars[key] = os.ExpandEnv(valueString)
		} else {
			expandedEnvVars[key] = value
		}
	}
	return expandedEnvVars
}
