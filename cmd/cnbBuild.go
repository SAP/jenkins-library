package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/docker"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/pkg/errors"
)

type cnbBuildUtils interface {
	command.ExecRunner

	FileExists(filename string) (bool, error)
	FileRead(path string) ([]byte, error)
	FileWrite(path string, content []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	Getwd() (string, error)
	Glob(pattern string) (matches []string, err error)
	Copy(src, dest string) (int64, error)
}

type cnbBuildUtilsBundle struct {
	*command.Command
	*piperutils.Files
}

func newCnbBuildUtils() cnbBuildUtils {
	utils := cnbBuildUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
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

func runCnbBuild(config *cnbBuildOptions, telemetryData *telemetry.CustomData, utils cnbBuildUtils, commonPipelineEnvironment *cnbBuildCommonPipelineEnvironment) error {
	var err error

	dockerConfig := &configfile.ConfigFile{}
	dockerConfigJSON := []byte(`{"auths":{}}`)
	if len(config.DockerConfigJSON) > 0 {
		dockerConfigJSON, err = utils.FileRead(config.DockerConfigJSON)
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

	err = utils.RunExecutable("/cnb/lifecycle/detector")
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		return errors.Wrap(err, "execution of '/cnb/lifecycle/detector' failed")
	}

	err = utils.RunExecutable("/cnb/lifecycle/builder")
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		return errors.Wrap(err, "execution of '/cnb/lifecycle/builder' failed")
	}

	utils.AppendEnv([]string{fmt.Sprintf("CNB_REGISTRY_AUTH=%s", string(cnbRegistryAuth))})
	err = utils.RunExecutable("/cnb/lifecycle/exporter", fmt.Sprintf("%s:%s", containerImage, containerImageTag), fmt.Sprintf("%s:latest", containerImage))
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		return errors.Wrap(err, "execution of '/cnb/lifecycle/exporter' failed")
	}

	return nil
}
