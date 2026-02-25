package cnbutils

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"errors"

	"github.com/SAP/jenkins-library/pkg/piperenv"
)

type TargetImage struct {
	ContainerImageName string
	ContainerImageTag  string
	ContainerRegistry  *url.URL
}

func GetTargetImage(imageRegistry, imageName, imageTag, projectID, envRootPath string) (*TargetImage, error) {
	if imageRegistry == "" || imageTag == "" {
		return nil, errors.New("containerRegistryUrl and containerImageTag must be present")
	}

	targetImage := &TargetImage{
		ContainerImageTag: strings.ReplaceAll(imageTag, "+", "-"),
	}

	if matched, _ := regexp.MatchString("^(http|https)://.*", imageRegistry); !matched {
		imageRegistry = fmt.Sprintf("https://%s", imageRegistry)
	}

	url, err := url.ParseRequestURI(imageRegistry)
	if err != nil {
		return nil, fmt.Errorf("invalid registry url: %w", err)
	}
	targetImage.ContainerRegistry = url

	cpePath := filepath.Join(envRootPath, "commonPipelineEnvironment")
	gitRepository := piperenv.GetResourceParameter(cpePath, "git", "repository")
	githubRepository := piperenv.GetResourceParameter(cpePath, "github", "repository")

	if imageName != "" {
		targetImage.ContainerImageName = imageName
	} else if projectID != "" {
		name := strings.ReplaceAll(projectID, ".", "-")
		targetImage.ContainerImageName = name
	} else if gitRepository != "" {
		targetImage.ContainerImageName = strings.ReplaceAll(gitRepository, ".", "-")
	} else if githubRepository != "" {
		targetImage.ContainerImageName = strings.ReplaceAll(githubRepository, ".", "-")
	} else {
		return nil, errors.New("failed to derive default for image name")
	}

	return targetImage, nil
}
