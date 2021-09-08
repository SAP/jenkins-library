// Package cnbutils provides utility functions to interact with Buildpacks
package cnbutils

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/SAP/jenkins-library/pkg/docker"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

type BuildPackMetadata struct {
	ID          string    `json:"id,omitempty"`
	Name        string    `json:"name,omitempty"`
	Version     string    `json:"version,omitempty"`
	Description string    `json:"description,omitempty"`
	Homepage    string    `json:"homepage,omitempty"`
	Keywords    []string  `json:"keywords,omitempty"`
	Licenses    []License `json:"licenses,omitempty"`
}

type License struct {
	Type string `json:"type"`
	URI  string `json:"uri"`
}

func DownloadBuildpacks(path string, bpacks []string, dClient docker.Client, fileutils *piperutils.Files) (Order, error) {
	var order Order
	for _, bpack := range bpacks {
		var bpackMeta BuildPackMetadata
		tempDir, err := fileutils.TempDir("", filepath.Base(bpack))
		if err != nil {
			return Order{}, fmt.Errorf("failed to create temp directory, error: %s", err.Error())
		}
		defer fileutils.RemoveAll(tempDir)

		log.Entry().Infof("Downloading buildpack '%s' to %s", bpack, tempDir)
		img, err := dClient.DownloadImageToPath(bpack, tempDir)
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
		order.Order = append(order.Order, OrderEntry{
			Group: []BuildpackRef{{
				ID:       bpackMeta.ID,
				Version:  bpackMeta.Version,
				Optional: false,
			}},
		})

		err = copyBuildPack(filepath.Join(tempDir, "cnb/buildpacks"), path, fileutils)
		if err != nil {
			return Order{}, err
		}
	}

	order.Futils = fileutils

	return order, nil
}

func copyBuildPack(src, dst string, futils *piperutils.Files) error {
	buildpacks, err := futils.Glob(filepath.Join(src, "*"))
	if err != nil {
		return fmt.Errorf("failed to read directory: %s, error: %s", src, err.Error())
	}

	for _, buildpack := range buildpacks {
		versions, err := futils.Glob(filepath.Join(buildpack, "*"))
		if err != nil {
			return fmt.Errorf("failed to read directory: %s, error: %s", buildpack, err.Error())
		}
		for _, srcVersionPath := range versions {
			destVersionPath := filepath.Join(dst, strings.ReplaceAll(srcVersionPath, src, ""))

			exists, err := futils.DirExists(destVersionPath)
			if err != nil {
				return fmt.Errorf("failed to check if directory exists: '%s', error: '%s'", destVersionPath, err.Error())
			}
			if exists {
				futils.RemoveAll(destVersionPath)
			}

			if err := futils.MkdirAll(filepath.Dir(destVersionPath), 0755); err != nil {
				return fmt.Errorf("failed to create directory: '%s', error: '%s'", filepath.Dir(destVersionPath), err.Error())
			}

			err = futils.FileRename(srcVersionPath, destVersionPath)
			if err != nil {
				return fmt.Errorf("failed to move '%s' to '%s', error: %s", srcVersionPath, destVersionPath, err.Error())
			}
		}
	}
	return nil
}
