package docker

import (
	"fmt"
	"net/url"
	"strings"

	containerName "github.com/google/go-containerregistry/pkg/name"
	"github.com/pkg/errors"
)

// ContainerRegistryFromURL provides the registry part of a complete registry url including the port
func ContainerRegistryFromURL(registryURL string) (string, error) {
	u, err := url.ParseRequestURI(registryURL)
	if err != nil {
		return "", errors.Wrap(err, "invalid registry url")
	}
	if len(u.Host) == 0 {
		return "", fmt.Errorf("invalid registry url")
	}
	return u.Host, nil
}

// ContainerRegistryFromImage provides the registry part of a full image name
func ContainerRegistryFromImage(fullImage string) (string, error) {
	ref, err := containerName.ParseReference(strings.ToLower(fullImage))
	if err != nil {
		return "", errors.Wrap(err, "failed to parse image name")
	}
	return ref.Context().RegistryStr(), nil
}

// ContainerImageNameTagFromImage provides the name & tag part of a full image name
func ContainerImageNameTagFromImage(fullImage string) (string, error) {
	ref, err := containerName.ParseReference(strings.ToLower(fullImage))
	if err != nil {
		return "", errors.Wrap(err, "failed to parse image name")
	}
	registryOnly := fmt.Sprintf("%v/", ref.Context().RegistryStr())
	return strings.ReplaceAll(fullImage, registryOnly, ""), nil
}

// ContainerImageNameFromImage provides the name without tag part of a full image name
func ContainerImageNameFromImage(fullImage string) (string, error) {
	imageNameTag, err := ContainerImageNameTagFromImage(fullImage)
	if err != nil {
		return "", errors.Wrap(err, "could not extract image name with tag from full image")
	}
	return strings.Split(imageNameTag, ":")[0], nil
}

// ContainerImageTagFromImage provides the tag part of a full image name
func ContainerImageTagFromImage(fullImage string) (string, error) {
	imageNameTag, err := ContainerImageNameTagFromImage(fullImage)
	if err != nil {
		return "", errors.Wrap(err, "could not extract image name with tag from full image")
	}
	splitting := strings.Split(imageNameTag, ":")
	if len(splitting) < 2 || splitting[1] == "" {
		return "", errors.New("could not extract tag from full image")
	}
	return splitting[1], nil
}
