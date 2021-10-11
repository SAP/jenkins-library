// Package cnbutils provides utility functions to interact with Buildpacks
package cnbutils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
)

type BuildPackMetadata struct {
	ID          string    `toml:"id,omitempty" json:"id,omitempty" yaml:"id,omitempty"`
	Name        string    `toml:"name,omitempty" json:"name,omitempty" yaml:"name,omitempty"`
	Version     string    `toml:"version,omitempty" json:"version,omitempty" yaml:"version,omitempty"`
	Description string    `toml:"description,omitempty" json:"description,omitempty" yaml:"description,omitempty"`
	Homepage    string    `toml:"homepage,omitempty" json:"homepage,omitempty" yaml:"homepage,omitempty"`
	Keywords    []string  `toml:"keywords,omitempty" json:"keywords,omitempty" yaml:"keywords,omitempty"`
	Licenses    []License `toml:"licenses,omitempty" json:"licenses,omitempty" yaml:"licenses,omitempty"`
}

type License struct {
	Type string `toml:"type" json:"type"`
	URI  string `toml:"uri" json:"uri"`
}

func DownloadBuildpacks(path string, bpacks []string, dockerCreds string, utils BuildUtils) (Order, error) {

	if dockerCreds != "" {
		os.Setenv("DOCKER_CONFIG", filepath.Dir(dockerCreds))
	}

	var orderEntry OrderEntry
	order := Order{
		Utils: utils,
	}

	for _, bpack := range bpacks {
		var bpackMeta BuildPackMetadata
		tempDir, err := utils.TempDir("", filepath.Base(bpack))
		if err != nil {
			return Order{}, fmt.Errorf("failed to create temp directory, error: %s", err.Error())
		}
		defer utils.RemoveAll(tempDir)

		log.Entry().Infof("Downloading buildpack '%s' to %s", bpack, tempDir)
		img, err := utils.DownloadImageToPath(bpack, tempDir)
		if err != nil {
			return Order{}, fmt.Errorf("failed download buildpack image '%s', error: %s", bpack, err.Error())
		}

		imgConf, err := img.Image.ConfigFile()
		if err != nil {
			return Order{}, fmt.Errorf("failed to read '%s' image config, error: %s", bpack, err.Error())
		}

		err = json.Unmarshal([]byte(imgConf.Config.Labels["io.buildpacks.buildpackage.metadata"]), &bpackMeta)
		if err != nil {
			return Order{}, fmt.Errorf("failed unmarshal '%s' image label, error: %s", bpack, err.Error())
		}
		log.Entry().Debugf("Buildpack metadata: '%v'", bpackMeta)
		orderEntry.Group = append(orderEntry.Group, bpackMeta)

		err = copyBuildPack(filepath.Join(tempDir, "cnb/buildpacks"), path, utils)
		if err != nil {
			return Order{}, err
		}
	}

	order.Order = []OrderEntry{orderEntry}

	return order, nil
}

func copyBuildPack(src, dst string, utils BuildUtils) error {
	buildpacks, err := utils.Glob(filepath.Join(src, "*"))
	if err != nil {
		return fmt.Errorf("failed to read directory: %s, error: %s", src, err.Error())
	}

	for _, buildpack := range buildpacks {
		versions, err := utils.Glob(filepath.Join(buildpack, "*"))
		if err != nil {
			return fmt.Errorf("failed to read directory: %s, error: %s", buildpack, err.Error())
		}
		for _, srcVersionPath := range versions {
			destVersionPath := filepath.Join(dst, strings.ReplaceAll(srcVersionPath, src, ""))

			exists, err := utils.FileExists(destVersionPath)
			if err != nil {
				return fmt.Errorf("failed to check if directory exists: '%s', error: '%s'", destVersionPath, err.Error())
			}
			if exists {
				utils.RemoveAll(destVersionPath)
			}

			if err := utils.MkdirAll(filepath.Dir(destVersionPath), 0755); err != nil {
				return fmt.Errorf("failed to create directory: '%s', error: '%s'", filepath.Dir(destVersionPath), err.Error())
			}

			err = utils.FileRename(srcVersionPath, destVersionPath)
			if err != nil {
				return fmt.Errorf("failed to move '%s' to '%s', error: %s", srcVersionPath, destVersionPath, err.Error())
			}
		}
	}
	return nil
}
