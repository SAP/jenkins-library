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

func DownloadBuildpacks(path string, bpacks []string, dockerCreds string, utils BuildUtils) error {
	if dockerCreds != "" {
		os.Setenv("DOCKER_CONFIG", filepath.Dir(dockerCreds))
	}

	err := utils.MkdirAll(bpCacheDir, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "failed to create temp directory for buildpack cache")
	}

	for _, bpack := range bpacks {
		imageInfo, err := utils.GetRemoteImageInfo(bpack)
		if err != nil {
			return errors.Wrap(err, "failed to get remote image info of buildpack")
		}
		hash, err := imageInfo.Digest()
		if err != nil {
			return errors.Wrap(err, "failed to get image digest")
		}
		cacheDir := filepath.Join(bpCacheDir, hash.String())

		cacheExists, err := utils.DirExists(cacheDir)
		if err != nil {
			return errors.Wrapf(err, "failed to check if cache dir '%s' exists", cacheDir)
		}

		if cacheExists {
			log.Entry().Infof("Using cached buildpack '%s'", bpack)
		} else {
			err := utils.MkdirAll(cacheDir, os.ModePerm)
			if err != nil {
				return errors.Wrap(err, "failed to create temp directory for buildpack cache")
			}

			log.Entry().Infof("Downloading buildpack '%s' to %s", bpack, cacheDir)
			_, err = utils.DownloadImageContent(bpack, cacheDir)
			if err != nil {
				return errors.Wrapf(err, "failed download buildpack image '%s'", bpack)
			}
		}

		matches, err := utils.Glob(filepath.Join(cacheDir, "cnb/buildpacks/*"))
		if err != nil {
			return err
		}

		for _, match := range matches {
			err = CreateVersionSymlinks(path, match, utils)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func GetMetadata(bpacks []string, utils BuildUtils) ([]BuildPackMetadata, error) {
	var metadata []BuildPackMetadata

	for _, bpack := range bpacks {
		var bpackMeta BuildPackMetadata
		imageInfo, err := utils.GetRemoteImageInfo(bpack)
		if err != nil {
			return nil, err
		}

		imgConf, err := imageInfo.ConfigFile()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read '%s' image config", bpack)
		}

		err = json.Unmarshal([]byte(imgConf.Config.Labels["io.buildpacks.buildpackage.metadata"]), &bpackMeta)
		if err != nil {
			return nil, err
		}
		metadata = append(metadata, bpackMeta)
	}

	return metadata, nil
}

func CreateVersionSymlinks(basePath, buildpackDir string, utils BuildUtils) error {
	newBuildpackPath := filepath.Join(basePath, filepath.Base(buildpackDir))
	err := utils.MkdirAll(newBuildpackPath, os.ModePerm)
	if err != nil {
		return err
	}

	versions, err := utils.Glob(filepath.Join(buildpackDir, "*"))
	if err != nil {
		return err
	}

	for _, version := range versions {
		newVersionPath := filepath.Join(newBuildpackPath, filepath.Base(version))
		exists, err := utils.DirExists(newVersionPath)
		if err != nil {
			return err
		}

		if !exists {
			err = utils.Symlink(version, newVersionPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
