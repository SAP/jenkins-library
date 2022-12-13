// Package cnbutils provides utility functions to interact with Buildpacks
package cnbutils

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
)

const bpCacheDir = "/tmp/buildpacks_cache"

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

	err := utils.MkdirAll(bpCacheDir, os.ModePerm)
	if err != nil {
		return Order{}, errors.Wrap(err, "failed to create temp directory for buildpack cache")
	}

	for _, bpack := range bpacks {
		var bpackMeta BuildPackMetadata
		imageInfo, err := utils.GetRemoteImageInfo(bpack)
		if err != nil {
			return Order{}, errors.Wrap(err, "failed to get remote image info of buildpack")
		}
		hash, err := imageInfo.Digest()
		if err != nil {
			return Order{}, errors.Wrap(err, "failed to get image digest")
		}
		cacheDir := filepath.Join(bpCacheDir, hash.String())

		cacheExists, err := utils.DirExists(cacheDir)
		if err != nil {
			return Order{}, errors.Wrapf(err, "failed to check if cache dir '%s' exists", cacheDir)
		}

		if cacheExists {
			log.Entry().Infof("Using cached buildpack '%s'", bpack)
		} else {
			err := utils.MkdirAll(cacheDir, os.ModePerm)
			if err != nil {
				return Order{}, errors.Wrap(err, "failed to create temp directory for buildpack cache")
			}

			log.Entry().Infof("Downloading buildpack '%s' to %s", bpack, cacheDir)
			img, err := utils.DownloadImageContent(bpack, cacheDir)
			if err != nil {
				return Order{}, errors.Wrapf(err, "failed download buildpack image '%s'", bpack)
			}
			imageInfo = img
		}

		imgConf, err := imageInfo.ConfigFile()
		if err != nil {
			return Order{}, errors.Wrapf(err, "failed to read '%s' image config", bpack)
		}

		err = json.Unmarshal([]byte(imgConf.Config.Labels["io.buildpacks.buildpackage.metadata"]), &bpackMeta)
		if err != nil {
			return Order{}, errors.Wrapf(err, "failed unmarshal '%s' image label", bpack)
		}
		log.Entry().Debugf("Buildpack metadata: '%v'", bpackMeta)
		orderEntry.Group = append(orderEntry.Group, bpackMeta)

		err = CopyProject(filepath.Join(cacheDir, "cnb/buildpacks"), path, nil, nil, utils)
		if err != nil {
			return Order{}, err
		}
	}

	order.Order = []OrderEntry{orderEntry}

	return order, nil
}
