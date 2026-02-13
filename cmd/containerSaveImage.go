package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	piperDocker "github.com/SAP/jenkins-library/pkg/docker"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"

	"github.com/pkg/errors"
)

func containerSaveImage(config containerSaveImageOptions, telemetryData *telemetry.CustomData) {
	var cachePath = "./cache"

	fileUtils := piperutils.Files{}

	dClientOptions := piperDocker.ClientOptions{ImageName: config.ContainerImage, RegistryURL: config.ContainerRegistryURL, LocalPath: config.FilePath, ImageFormat: config.ImageFormat}
	dClient := &piperDocker.Client{}
	dClient.SetOptions(dClientOptions)

	_, err := runContainerSaveImage(&config, telemetryData, cachePath, "", dClient, fileUtils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runContainerSaveImage(config *containerSaveImageOptions, telemetryData *telemetry.CustomData, cachePath, rootPath string, dClient piperDocker.Download, fileUtils piperutils.FileUtils) (string, error) {
	if err := correctContainerDockerConfigEnvVar(config, fileUtils); err != nil {
		return "", err
	}

	tarfilePath := config.FilePath

	if len(tarfilePath) == 0 {
		tarfilePath = filenameFromContainer(rootPath, config.ContainerImage)
	} else {
		tarfilePath = filepath.Join(rootPath, tarfilePath)
		// tarfilePath is passed as project name that will not consist of the .tar extension hence adding the extension and replacing spaces with _
		if fileExtension := filepath.Ext(tarfilePath); fileExtension != ".tar" {
			tarfilePath = fmt.Sprintf("%s.tar", tarfilePath)
		}
	}

	log.Entry().Infof("Downloading '%s' to '%s' with pass '%s' and user '%s'", config.ContainerImage, tarfilePath, config.ContainerRegistryPassword, config.ContainerRegistryUser)
	if _, err := dClient.DownloadImage(config.ContainerImage, tarfilePath); err != nil {
		return "", errors.Wrap(err, "failed to download docker image")
	}

	return tarfilePath, nil
}

func filenameFromContainer(rootPath, containerImage string) string {
	re := regexp.MustCompile("[^a-zA-Z0-9-]")

	return filepath.Join(rootPath, fmt.Sprintf("%s.tar", re.ReplaceAllString(containerImage, "_")))
}

func correctContainerDockerConfigEnvVar(config *containerSaveImageOptions, utils piperutils.FileUtils) error {
	dockerConfigDir, err := utils.TempDir("", "docker")

	if err != nil {
		return errors.Wrap(err, "unable to create docker config dir")
	}

	dockerConfigFile := fmt.Sprintf("%s/%s", dockerConfigDir, "config.json")

	if len(config.DockerConfigJSON) > 0 {
		log.Entry().Infof("Docker credentials configuration: %v", config.DockerConfigJSON)

		if exists, _ := utils.FileExists(config.DockerConfigJSON); exists {
			if _, err = utils.Copy(config.DockerConfigJSON, dockerConfigFile); err != nil {
				return errors.Wrap(err, "unable to copy docker config")
			}
		}
	} else {
		log.Entry().Info("Docker credentials configuration: NONE")
	}

	if len(config.ContainerRegistryURL) > 0 && len(config.ContainerRegistryUser) > 0 && len(config.ContainerRegistryPassword) > 0 {
		if _, err = piperDocker.CreateDockerConfigJSON(config.ContainerRegistryURL, config.ContainerRegistryUser, config.ContainerRegistryPassword, dockerConfigFile, dockerConfigFile, utils); err != nil {
			log.Entry().Warningf("failed to update Docker config.json: %v", err)
		}
	}

	os.Setenv("DOCKER_CONFIG", dockerConfigDir)

	return nil
}
