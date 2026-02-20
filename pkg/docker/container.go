package docker

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	containerName "github.com/google/go-containerregistry/pkg/name"
)

// ContainerRegistryFromURL provides the registry part of a complete registry url including the port
func ContainerRegistryFromURL(registryURL string) (string, error) {
	u, err := url.ParseRequestURI(registryURL)
	if err != nil {
		return "", fmt.Errorf("invalid registry url: %w", err)
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
		return "", fmt.Errorf("failed to parse image name: %w", err)
	}
	return ref.Context().RegistryStr(), nil
}

// ContainerImageNameTagFromImage provides the name & tag part of a full image name
func ContainerImageNameTagFromImage(fullImage string) (string, error) {
	ref, err := containerName.ParseReference(strings.ToLower(fullImage))
	if err != nil {
		return "", fmt.Errorf("failed to parse image name: %w", err)
	}
	registryOnly := fmt.Sprintf("%v/", ref.Context().RegistryStr())
	return strings.ReplaceAll(fullImage, registryOnly, ""), nil
}

// ContainerImageNameFromImage returns the image name of a given docker reference
func ContainerImageNameFromImage(fullImage string) (string, error) {
	imageNameTag, err := ContainerImageNameTagFromImage(fullImage)

	if err != nil {
		return "", err
	}

	r := regexp.MustCompile(`([^:@]+)`)
	m := r.FindStringSubmatch(imageNameTag)

	return m[0], nil
}
