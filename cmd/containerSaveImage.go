package cmd

import (
	"os"
	"path/filepath"
	"strings"

	piperDocker "github.com/SAP/jenkins-library/pkg/docker"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"

	"github.com/pkg/errors"
)

func containerSaveImage(config containerSaveImageOptions, telemetryData *telemetry.CustomData) {
	var cachePath = "./cache"

	dClientOptions := piperDocker.ClientOptions{ImageName: config.ContainerImage, RegistryURL: config.ContainerRegistryURL, LocalPath: config.FilePath, IncludeLayers: config.IncludeLayers}
	dClient := &piperDocker.Client{}
	dClient.SetOptions(dClientOptions)

	_, err := runContainerSaveImage(&config, telemetryData, cachePath, "", dClient)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runContainerSaveImage(config *containerSaveImageOptions, telemetryData *telemetry.CustomData, cachePath, rootPath string, dClient piperDocker.Download) (string, error) {
	err := os.RemoveAll(cachePath)
	if err != nil {
		return "", errors.Wrap(err, "failed to prepare cache")
	}

	err = os.Mkdir(cachePath, 0755)
	if err != nil {
		return "", errors.Wrap(err, "failed to create cache")
	}

	// ensure that download cache is cleaned up at the end
	defer os.RemoveAll(cachePath)

	imageSource, err := dClient.GetImageSource()
	if err != nil {
		return "", errors.Wrap(err, "failed to get docker image source")
	}
	image, err := dClient.DownloadImageToPath(imageSource, cachePath)
	if err != nil {
		return "", errors.Wrap(err, "failed to download docker image")
	}

	tarfilePath := config.FilePath
	if len(tarfilePath) == 0 {
		tarfilePath = filenameFromContainer(rootPath, config.ContainerImage)
	} else {
		tarfilePath = filenameFromContainer(rootPath, tarfilePath)
	}

	tarFile, err := os.Create(tarfilePath)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create %v for docker image", tarfilePath)
	}
	defer tarFile.Close()

	if err := os.Chmod(tarfilePath, 0644); err != nil {
		return "", errors.Wrapf(err, "failed to adapt permissions on %v", tarfilePath)
	}

	err = dClient.TarImage(tarFile, image)
	if err != nil {
		return "", errors.Wrap(err, "failed to tar container image")
	}

	return tarfilePath, nil
}

func filenameFromContainer(rootPath, containerImage string) string {
	return filepath.Join(rootPath, strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(containerImage, "/", "_"), ":", "_"), ".", "_")+".tar")
}
