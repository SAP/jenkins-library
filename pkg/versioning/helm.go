package versioning

import (
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	"helm.sh/helm/v3/pkg/chart"
)

// JSONfile defines an artifact using a json file for versioning
type HelmChart struct {
	path             string
	metadata         chart.Metadata
	utils            Utils
	updateAppVersion bool
}

func (h *HelmChart) init() error {
	if h.utils == nil {
		return fmt.Errorf("no file utils provided")
	}
	if len(h.path) == 0 {
		charts, err := h.utils.Glob("**/Chart.yaml")
		if len(charts) == 0 || err != nil {
			return fmt.Errorf("failed to find a helm chart file")
		}
		// use first chart which can be found
		h.path = charts[0]
	}

	if len(h.metadata.Version) == 0 {
		content, err := h.utils.FileRead(h.path)
		if err != nil {
			return fmt.Errorf("failed to read file '%v': %w", h.path, err)
		}

		err = yaml.Unmarshal(content, &h.metadata)
		if err != nil {
			return fmt.Errorf("helm chart content invalid '%v': %w", h.path, err)
		}
	}

	return nil
}

// VersioningScheme returns the relevant versioning scheme
func (h *HelmChart) VersioningScheme() string {
	return "semver2"
}

// GetVersion returns the current version of the artifact with a JSON-based build descriptor
func (h *HelmChart) GetVersion() (string, error) {
	if err := h.init(); err != nil {
		return "", fmt.Errorf("failed to init helm chart versioning: %w", err)
	}

	return h.metadata.Version, nil
}

// SetVersion updates the version of the artifact with a JSON-based build descriptor
func (h *HelmChart) SetVersion(version string) error {
	if err := h.init(); err != nil {
		return fmt.Errorf("failed to init helm chart versioning: %w", err)
	}

	h.metadata.Version = version
	if h.updateAppVersion {
		// k8s does not allow a plus sign in labels
		h.metadata.AppVersion = strings.ReplaceAll(version, "+", "_")
	}

	content, err := yaml.Marshal(h.metadata)
	if err != nil {
		return fmt.Errorf("failed to create chart content for '%v': %w", h.path, err)
	}
	err = h.utils.FileWrite(h.path, content, 666)
	if err != nil {
		return fmt.Errorf("failed to write file '%v': %w", h.path, err)
	}

	return nil
}

// GetCoordinates returns the coordinates
func (h *HelmChart) GetCoordinates() (Coordinates, error) {
	result := Coordinates{}
	projectVersion, err := h.GetVersion()
	if err != nil {
		return result, err
	}

	result.ArtifactID = h.metadata.Name
	result.Version = projectVersion
	result.GroupID = h.metadata.Home

	return result, nil
}
