package docker

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"

	pkgutil "github.com/GoogleContainerTools/container-diff/pkg/util"
	"github.com/google/go-containerregistry/pkg/legacy/tarball"
	"github.com/google/go-containerregistry/pkg/name"
)

// AuthEntry defines base64 encoded username:password required inside a Docker config.json
type AuthEntry struct {
	Auth string `json:"auth,omitempty"`
}

// CreateDockerConfigJSON creates / updates a Docker config.json with registry credentials
func CreateDockerConfigJSON(registryURL, username, password, targetPath, configPath string, utils piperutils.FileUtils) (string, error) {

	if len(targetPath) == 0 {
		targetPath = configPath
	}

	dockerConfig := map[string]interface{}{}
	if exists, _ := utils.FileExists(configPath); exists {
		dockerConfigContent, err := utils.FileRead(configPath)
		if err != nil {
			return "", fmt.Errorf("failed to read file '%v': %w", configPath, err)
		}

		err = json.Unmarshal(dockerConfigContent, &dockerConfig)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal json file '%v': %w", configPath, err)
		}
	}

	credentialsBase64 := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v:%v", username, password)))
	dockerAuth := AuthEntry{Auth: credentialsBase64}

	if dockerConfig["auths"] == nil {
		dockerConfig["auths"] = map[string]AuthEntry{registryURL: dockerAuth}
	} else {
		authEntries, ok := dockerConfig["auths"].(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("failed to read authentication entries from file '%v': format invalid", configPath)
		}
		authEntries[registryURL] = dockerAuth
		dockerConfig["auths"] = authEntries
	}

	jsonResult, err := json.Marshal(dockerConfig)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Docker config.json: %w", err)
	}

	//always create the target path directories if any before writing
	err = utils.MkdirAll(filepath.Dir(targetPath), 0777)
	if err != nil {
		return "", fmt.Errorf("failed to create directory path for the Docker config.json file %v:%w", targetPath, err)
	}
	err = utils.FileWrite(targetPath, jsonResult, 0666)
	if err != nil {
		return "", fmt.Errorf("failed to write Docker config.json: %w", err)
	}

	return targetPath, nil
}

// Client defines an docker client object
type Client struct {
	imageName     string
	registryURL   string
	localPath     string
	includeLayers bool
}

// ClientOptions defines the options to be set on the client
type ClientOptions struct {
	ImageName     string
	RegistryURL   string
	LocalPath     string
	IncludeLayers bool
}

//Download interface for download an image to a local path
type Download interface {
	GetImageSource() (string, error)
	DownloadImageToPath(imageSource, filePath string) (pkgutil.Image, error)
	TarImage(writer io.Writer, image pkgutil.Image) error
}

// SetOptions sets options used for the docker client
func (c *Client) SetOptions(options ClientOptions) {
	c.imageName = options.ImageName
	c.registryURL = options.RegistryURL
	c.includeLayers = options.IncludeLayers
	c.localPath = options.LocalPath
}

const (
	daemonPrefix = "daemon://"
	remotePrefix = "remote://"
)

//GetImageSource get the image source from client attributes (localPath, imageName, registryURL)
func (c *Client) GetImageSource() (string, error) {

	imageSource := c.imageName

	if len(c.registryURL) > 0 && len(c.localPath) <= 0 {
		registry := c.registryURL

		url, err := url.Parse(c.registryURL)
		if err != nil {
			return "", fmt.Errorf("failed to parse registryURL %v: %w", c.registryURL, err)
		}

		//remove protocol from registryURL to get registry
		if len(url.Scheme) > 0 {
			registry = strings.Replace(c.registryURL, fmt.Sprintf("%v://", url.Scheme), "", 1)
		}

		if strings.HasSuffix(registry, "/") {
			imageSource = fmt.Sprintf("%v%v%v", remotePrefix, registry, c.imageName)
		} else {
			imageSource = fmt.Sprintf("%v%v/%v", remotePrefix, registry, c.imageName)
		}
	} else if len(c.localPath) > 0 {
		imageSource = c.localPath
		if !pkgutil.IsTar(c.localPath) {
			imageSource = fmt.Sprintf("%v%v", daemonPrefix, c.localPath)
		}
	}

	if len(imageSource) <= 0 {
		return imageSource, fmt.Errorf("no image found for the parameters: (Name: %v, Registry: %v, local Path: %v)", c.imageName, c.registryURL, c.localPath)
	}

	return imageSource, nil
}

//DownloadImageToPath download the image to the specified path
func (c *Client) DownloadImageToPath(imageSource, filePath string) (pkgutil.Image, error) {

	return pkgutil.GetImage(imageSource, c.includeLayers, filePath)
}

//TarImage write a tar from the given image
func (c *Client) TarImage(writer io.Writer, image pkgutil.Image) error {

	reference, err := name.ParseReference(image.Digest.String(), name.WeakValidation)
	if err != nil {
		return err
	}
	err = tarball.Write(reference, image.Image, writer)
	if err != nil {
		return err
	}
	return nil
}

// ImageListWithFilePath compiles container image names based on all Dockerfiles found, considering excludes
// according to following search pattern: **/Dockerfile*
// Return value contains a map with image names and file path
// Examples for image names with imageName testImage
// * Dockerfile: `imageName`
// * sub1/Dockerfile: `imageName-sub1`
// * sub2/Dockerfile_proxy: `imageName-sub2-proxy`
func ImageListWithFilePath(imageName string, excludes []string, trimDir string, utils piperutils.FileUtils) (map[string]string, error) {

	imageList := map[string]string{}

	pattern := "**/Dockerfile*"

	matches, err := utils.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return imageList, fmt.Errorf("failed to retrieve Dockerfiles")
	}

	for _, dockerfilePath := range matches {
		// make sure that the path we have is relative
		// ToDo: needs rework
		//dockerfilePath = strings.ReplaceAll(dockerfilePath, cwd, ".")

		if piperutils.ContainsString(excludes, dockerfilePath) {
			log.Entry().Infof("Discard %v since it is in the exclude list %v", dockerfilePath, excludes)
			continue
		}

		if dockerfilePath == "Dockerfile" {
			imageList[imageName] = dockerfilePath
		} else {
			var finalName string
			if base := filepath.Base(dockerfilePath); base == "Dockerfile" {
				subName := strings.ReplaceAll(filepath.Dir(dockerfilePath), string(filepath.Separator), "-")
				if len(trimDir) > 0 {
					// allow to remove trailing sub directories
					// example .ci/app/Dockerfile
					// with trimDir = .ci/ imagename would only contain app part.
					subName = strings.TrimPrefix(subName, strings.ReplaceAll(trimDir, "/", "-"))
					// make sure that subName does not start with a - (e.g. due not configuring trailing slash for trimDir)
					subName = strings.TrimPrefix(subName, "-")
				}
				finalName = fmt.Sprintf("%v-%v", imageName, subName)
			} else {
				parts := strings.FieldsFunc(base, func(separator rune) bool {
					return separator == []rune("-")[0] || separator == []rune("_")[0]
				})
				if len(parts) == 1 {
					return imageList, fmt.Errorf("wrong format of Dockerfile, must be inside a sub-folder or contain a separator")
				}
				parts[0] = imageName
				finalName = strings.Join(parts, "-")
			}

			imageList[finalName] = dockerfilePath
		}
	}

	return imageList, nil
}
