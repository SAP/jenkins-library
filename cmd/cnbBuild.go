package cmd

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/SAP/jenkins-library/pkg/certutils"
	"github.com/SAP/jenkins-library/pkg/cnbutils"
	"github.com/SAP/jenkins-library/pkg/cnbutils/project"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/docker"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/pkg/errors"
	ignore "github.com/sabhiram/go-gitignore"
)

const (
	detectorPath = "/cnb/lifecycle/detector"
	builderPath  = "/cnb/lifecycle/builder"
	exporterPath = "/cnb/lifecycle/exporter"
)

type cnbBuildUtilsBundle struct {
	*command.Command
	*piperutils.Files
	*docker.Client
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

	err := runCnbBuild(&config, telemetryData, utils, commonPipelineEnvironment, client)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func isIgnored(find string, include, exclude *ignore.GitIgnore) bool {
	if exclude != nil {
		filtered := exclude.MatchesPath(find)

		if filtered {
			log.Entry().Debugf("%s matches exclude pattern, ignoring", find)
			return true
		}
	}

	if include != nil {
		filtered := !include.MatchesPath(find)

		if filtered {
			log.Entry().Debugf("%s doesn't match include pattern, ignoring", find)
			return true
		} else {
			log.Entry().Debugf("%s matches include pattern", find)
			return false
		}
	}

	return false
}

func isDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

func isBuilder(utils cnbutils.BuildUtils) (bool, error) {
	for _, path := range []string{detectorPath, builderPath, exporterPath} {
		exists, err := utils.FileExists(path)
		if err != nil || !exists {
			return exists, err
		}
	}
	return true, nil
}

func isZip(path string) bool {
	r, err := zip.OpenReader(path)

	switch {
	case err == nil:
		r.Close()
		return true
	case err == zip.ErrFormat:
		return false
	default:
		return false
	}
}

func copyFile(source, target string, utils cnbutils.BuildUtils) error {
	targetDir := filepath.Dir(target)

	exists, err := utils.DirExists(targetDir)
	if err != nil {
		return err
	}

	if !exists {
		log.Entry().Debugf("Creating directory %s", targetDir)
		err = utils.MkdirAll(targetDir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	_, err = utils.Copy(source, target)
	return err
}

func copyProject(source, target string, include, exclude *ignore.GitIgnore, utils cnbutils.BuildUtils) error {
	sourceFiles, _ := utils.Glob(path.Join(source, "**"))
	for _, sourceFile := range sourceFiles {
		if !isIgnored(sourceFile, include, exclude) {
			target := path.Join(target, strings.ReplaceAll(sourceFile, source, ""))
			dir, err := isDir(sourceFile)
			if err != nil {
				log.SetErrorCategory(log.ErrorBuild)
				return errors.Wrapf(err, "Checking file info '%s' failed", target)
			}

			if dir {
				err = utils.MkdirAll(target, os.ModePerm)
				if err != nil {
					log.SetErrorCategory(log.ErrorBuild)
					return errors.Wrapf(err, "Creating directory '%s' failed", target)
				}
			} else {
				log.Entry().Debugf("Copying '%s' to '%s'", sourceFile, target)
				err = copyFile(sourceFile, target, utils)
				if err != nil {
					log.SetErrorCategory(log.ErrorBuild)
					return errors.Wrapf(err, "Copying '%s' to '%s' failed", sourceFile, target)
				}
			}

		}
	}
	return nil
}

func extractZip(source, target string, utils cnbutils.BuildUtils) error {

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
		err := utils.FileRename(source, newPath)
		if err != nil {
			return "", err
		}

		return newPath, nil
	}

	return source, nil
}

func runCnbBuild(config *cnbBuildOptions, telemetryData *telemetry.CustomData, utils cnbutils.BuildUtils, commonPipelineEnvironment *cnbBuildCommonPipelineEnvironment, httpClient piperhttp.Sender) error {
	var err error

	exists, err := isBuilder(utils)

	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Wrap(err, "failed to check if dockerImage is a valid builder")
	}
	if !exists {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.New("the provided dockerImage is not a valid builder")
	}

	include := ignore.CompileIgnoreLines("**/*")
	exclude := ignore.CompileIgnoreLines("piper", ".pipeline")

	projDescExists, err := utils.FileExists(config.ProjectDescriptor)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Wrap(err, "failed to check if project descriptor exists")
	}

	if projDescExists {
		descriptor, err := project.ParseDescriptor(config.ProjectDescriptor, utils, httpClient)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.Wrapf(err, "failed to parse %s", config.ProjectDescriptor)
		}

		if (config.Buildpacks == nil || len(config.Buildpacks) == 0) && len(descriptor.Buildpacks) > 0 {
			config.Buildpacks = descriptor.Buildpacks
		}

		if (config.BuildEnvVars == nil || len(config.BuildEnvVars) == 0) && len(descriptor.EnvVars) > 0 {
			config.BuildEnvVars = descriptor.EnvVars
		}

		if descriptor.Exclude != nil {
			exclude = descriptor.Exclude
		}

		if descriptor.Include != nil {
			include = descriptor.Include
		}
	}

	platformPath := "/platform"
	if config.BuildEnvVars != nil && len(config.BuildEnvVars) > 0 {
		log.Entry().Infof("Setting custom environment variables: '%v'", config.BuildEnvVars)
		platformPath = "/tmp/platform"
		err = cnbutils.CreateEnvFiles(utils, platformPath, config.BuildEnvVars)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.Wrap(err, "failed to write environment variables to files")
		}
	}

	dockerConfigFile := ""
	dockerConfig := &configfile.ConfigFile{}
	dockerConfigJSON := []byte(`{"auths":{}}`)
	if len(config.DockerConfigJSON) > 0 {
		dockerConfigFile, err = prepareDockerConfig(config.DockerConfigJSON, utils)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.Wrapf(err, "failed to rename DockerConfigJSON file '%v'", config.DockerConfigJSON)
		}
		dockerConfigJSON, err = utils.FileRead(dockerConfigFile)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.Wrapf(err, "failed to read DockerConfigJSON file '%v'", config.DockerConfigJSON)
		}
	}

	err = json.Unmarshal(dockerConfigJSON, dockerConfig)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Wrapf(err, "failed to parse DockerConfigJSON file '%v'", config.DockerConfigJSON)
	}

	auth := map[string]string{}
	for registry, value := range dockerConfig.AuthConfigs {
		auth[registry] = fmt.Sprintf("Basic %s", value.Auth)
	}

	cnbRegistryAuth, err := json.Marshal(auth)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Wrap(err, "failed to marshal DockerConfigJSON")
	}

	target := "/workspace"
	source, err := utils.Getwd()
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		return errors.Wrap(err, "failed to get current working directory")
	}

	if len(config.Path) > 0 {
		source = config.Path
	}

	dir, err := isDir(source)
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		return errors.Wrapf(err, "Checking file info '%s' failed", target)
	}

	if dir {
		err = copyProject(source, target, include, exclude, utils)
		if err != nil {
			log.SetErrorCategory(log.ErrorBuild)
			return errors.Wrapf(err, "Copying  '%s' into '%s' failed", source, target)
		}
	} else {
		err = extractZip(source, target, utils)
		if err != nil {
			log.SetErrorCategory(log.ErrorBuild)
			return errors.Wrapf(err, "Copying  '%s' into '%s' failed", source, target)
		}
	}

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

	var containerImage string
	var containerImageTag string

	if len(config.ContainerRegistryURL) == 0 || len(config.ContainerImageName) == 0 || len(config.ContainerImageTag) == 0 {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.New("containerRegistryUrl, containerImageName and containerImageTag must be present")
	}

	var containerRegistry string
	if matched, _ := regexp.MatchString("^(http|https)://.*", config.ContainerRegistryURL); matched {
		containerRegistry, err = docker.ContainerRegistryFromURL(config.ContainerRegistryURL)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.Wrapf(err, "failed to read containerRegistryUrl %s", config.ContainerRegistryURL)
		}
	} else {
		containerRegistry = config.ContainerRegistryURL
	}

	containerImage = fmt.Sprintf("%s/%s", containerRegistry, config.ContainerImageName)
	containerImageTag = strings.ReplaceAll(config.ContainerImageTag, "+", "-")
	commonPipelineEnvironment.container.registryURL = config.ContainerRegistryURL
	commonPipelineEnvironment.container.imageNameTag = containerImage

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
		log.Entry().Info("skipping updation of certificates")
	}

	err = utils.RunExecutable(detectorPath, "-buildpacks", buildpacksPath, "-order", orderPath, "-platform", platformPath)
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		return errors.Wrapf(err, "execution of '%s' failed", detectorPath)
	}

	err = utils.RunExecutable(builderPath, "-buildpacks", buildpacksPath, "-platform", platformPath)
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		return errors.Wrapf(err, "execution of '%s' failed", builderPath)
	}

	targets := []string{
		fmt.Sprintf("%s:%s", containerImage, containerImageTag),
	}

	for _, tag := range config.AdditionalTags {
		target := fmt.Sprintf("%s:%s", containerImage, tag)
		if !piperutils.ContainsString(targets, target) {
			targets = append(targets, target)
		}
	}

	utils.AppendEnv([]string{fmt.Sprintf("CNB_REGISTRY_AUTH=%s", string(cnbRegistryAuth))})
	err = utils.RunExecutable(exporterPath, targets...)
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		return errors.Wrapf(err, "execution of '%s' failed", exporterPath)
	}

	return nil
}
