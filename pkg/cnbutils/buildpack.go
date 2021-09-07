// Package cnbutils provides utility functions to interact with Buildpacks
package cnbutils

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	"github.com/SAP/jenkins-library/pkg/docker"
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

func DownloadBuildpacks(path string, bpacks []string, dClient docker.Client) (Order, error) {
	var order Order
	for _, bpack := range bpacks {
		var bpackMeta BuildPackMetadata

		tempDir, err := ioutil.TempDir("", bpack)
		if err != nil {
			return Order{}, err
		}
		defer os.RemoveAll(tempDir)

		img, err := dClient.DownloadImageToPath(bpack, tempDir)
		if err != nil {
			return Order{}, err
		}

		imgConf, err := img.Image.ConfigFile()
		if err != nil {
			return Order{}, err
		}

		err = json.Unmarshal([]byte(imgConf.Config.Labels["io.buildpacks.buildpackage.metadata"]), &bpackMeta)
		if err != nil {
			return Order{}, err
		}

		order.Order = append(order.Order, OrderEntry{
			Group: []BuildpackRef{{
				ID:       bpackMeta.ID,
				Version:  bpackMeta.Version,
				Optional: false,
			}},
		})

		err = copyBuildPack(fmt.Sprintf("%s/cnb/buildpacks", tempDir), path)
		if err != nil {
			return Order{}, err
		}
	}

	return order, nil
}

func copyBuildPack(src, dst string) error {
	buildpacks, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, buildpack := range buildpacks {
		versions, _ := os.ReadDir(filepath.Join(src, buildpack.Name()))
		for _, version := range versions {
			srcVersionPath := filepath.Join(src, buildpack.Name(), version.Name())
			destVersionPath := filepath.Join(dst, buildpack.Name(), version.Name())

			if exists(destVersionPath) {
				os.RemoveAll(destVersionPath)
			}
			if err := os.MkdirAll(destVersionPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory: '%s', error: '%s'", destVersionPath, err.Error())
			}

			err = copyDirectory(srcVersionPath, destVersionPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func copyDirectory(scrDir, dest string) error {
	entries, err := ioutil.ReadDir(scrDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(scrDir, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}

		stat, ok := fileInfo.Sys().(*syscall.Stat_t)
		if !ok {
			return fmt.Errorf("failed to get raw syscall.Stat_t data for '%s'", sourcePath)
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory: '%s', error: '%s'", destPath, err.Error())
			}
			if err := copyDirectory(sourcePath, destPath); err != nil {
				return err
			}
		default:
			if err := copy(sourcePath, destPath); err != nil {
				return err
			}
		}

		if err := os.Lchown(destPath, int(stat.Uid), int(stat.Gid)); err != nil {
			return err
		}
	}
	return nil
}

func copy(srcFile, dstFile string) error {
	out, err := os.Create(dstFile)
	if err != nil {
		return err
	}

	defer out.Close()

	in, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer in.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return nil
}

func exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	return true
}
