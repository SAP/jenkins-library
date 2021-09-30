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

	"github.com/SAP/jenkins-library/pkg/cnbutils"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/docker"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/pkg/errors"
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

	err := runCnbBuild(&config, telemetryData, utils, commonPipelineEnvironment)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func isIgnored(find string) bool {
	return strings.HasSuffix(find, "piper") || strings.Contains(find, ".pipeline")
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

func copyProject(source, target string, utils cnbutils.BuildUtils) error {
	sourceFiles, _ := utils.Glob(path.Join(source, "**"))
	for _, sourceFile := range sourceFiles {
		if !isIgnored(sourceFile) {
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
				_, err = utils.Copy(sourceFile, target)
				if err != nil {
					log.SetErrorCategory(log.ErrorBuild)
					return errors.Wrapf(err, "Copying '%s' to '%s' failed", sourceFile, target)
				}
			}

		} else {
			log.Entry().Debugf("Filtered out '%s'", sourceFile)
		}
	}
	return nil
}

func copyFile(source, target string, utils cnbutils.BuildUtils) error {

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

func runCnbBuild(config *cnbBuildOptions, telemetryData *telemetry.CustomData, utils cnbutils.BuildUtils, commonPipelineEnvironment *cnbBuildCommonPipelineEnvironment) error {
	var err error

	exists, err := isBuilder(utils)

	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Wrapf(err, "failed to check if dockerImage is a valid builder")
	}
	if !exists {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.New("the provided dockerImage is not a valid builder")
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
		if filepath.Base(config.DockerConfigJSON) != "config.json" {
			log.Entry().Debugf("Renaming docker config file from '%s' to 'config.json'", filepath.Base(config.DockerConfigJSON))

			dockerConfigFile = filepath.Join(filepath.Dir(config.DockerConfigJSON), "config.json")
			err = utils.FileRename(config.DockerConfigJSON, dockerConfigFile)
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return errors.Wrapf(err, "failed to rename DockerConfigJSON file '%v'", config.DockerConfigJSON)
			}
		} else {
			dockerConfigFile = config.DockerConfigJSON
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
		err = copyProject(source, target, utils)
		if err != nil {
			log.SetErrorCategory(log.ErrorBuild)
			return errors.Wrapf(err, "Copying  '%s' into '%s' failed", source, target)
		}
	} else {
		err = copyFile(source, target, utils)
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

	if len(config.ContainerRegistryURL) > 0 && len(config.ContainerImageName) > 0 && len(config.ContainerImageTag) > 0 {
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
	} else {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.New("containerRegistryUrl, containerImageName and containerImageTag must be present")
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

	utils.AppendEnv([]string{fmt.Sprintf("CNB_REGISTRY_AUTH=%s", string(cnbRegistryAuth))})
	err = utils.RunExecutable(exporterPath, fmt.Sprintf("%s:%s", containerImage, containerImageTag), fmt.Sprintf("%s:latest", containerImage))
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		return errors.Wrapf(err, "execution of '%s' failed", exporterPath)
	}

	return nil
}
