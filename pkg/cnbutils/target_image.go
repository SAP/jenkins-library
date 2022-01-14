package cnbutils

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/SAP/jenkins-library/pkg/docker"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/pkg/errors"
)

type TargetImage struct {
	ContainerImageName   string
	ContainerImageTag    string
	ContainerRegistryURL string
}

func GetTargetImage(imageRegistry, imageName, imageTag, envRootPath, projectID string) (*TargetImage, error) {
	if imageRegistry == "" || imageTag == "" {
		return nil, errors.New("containerRegistryUrl and containerImageTag must be present")
	}

	targetImage := &TargetImage{
		ContainerImageTag: strings.ReplaceAll(imageTag, "+", "-"),
	}

	if matched, _ := regexp.MatchString("^(http|https)://.*", imageRegistry); matched {
		containerRegistry, err := docker.ContainerRegistryFromURL(imageRegistry)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read containerRegistryUrl %s", imageRegistry)
		}
		targetImage.ContainerRegistryURL = containerRegistry
	} else {
		targetImage.ContainerRegistryURL = fmt.Sprintf("https://%v", imageRegistry)
	}

	cpePath := filepath.Join(envRootPath, "commonPipelineEnvironment")
	gitRepository := piperenv.GetResourceParameter(cpePath, "git", "repository")

	if imageName != "" {
		targetImage.ContainerImageName = imageName
	} else if projectID != "" {
		name := strings.ReplaceAll(projectID, ".", "-") // Sanitize image name
		targetImage.ContainerImageName = name
	} else if gitRepository != "" {
		targetImage.ContainerImageName = gitRepository // Sanitize image name?
	} else {
		return nil, errors.New("failed to derive default for image name")
	}

	return targetImage, nil
}
