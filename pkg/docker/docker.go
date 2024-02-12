package docker

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/configfile"
	cranecmd "github.com/google/go-containerregistry/cmd/crane/cmd"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/pkg/errors"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

// AuthEntry defines base64 encoded username:password required inside a Docker config.json
type AuthEntry struct {
	Auth string `json:"auth,omitempty"`
}

// MergeDockerConfigJSON merges two docker config.json files.
func MergeDockerConfigJSON(sourcePath, targetPath string, utils piperutils.FileUtils) error {
	if exists, _ := utils.FileExists(sourcePath); !exists {
		return fmt.Errorf("source dockerConfigJSON file %q does not exist", sourcePath)
	}

	sourceReader, err := utils.Open(sourcePath)
	if err != nil {
		return errors.Wrapf(err, "failed to open file %q", sourcePath)
	}
	defer sourceReader.Close()

	sourceConfig, err := config.LoadFromReader(sourceReader)
	if err != nil {
		return errors.Wrapf(err, "failed to read file %q", sourcePath)
	}

	var targetConfig *configfile.ConfigFile
	if exists, _ := utils.FileExists(targetPath); !exists {
		log.Entry().Warnf("target dockerConfigJSON file %q does not exist, creating a new one", sourcePath)
		targetConfig = configfile.New(targetPath)
	} else {
		targetReader, err := utils.Open(targetPath)
		if err != nil {
			return errors.Wrapf(err, "failed to open file %q", targetReader)
		}
		defer targetReader.Close()
		targetConfig, err = config.LoadFromReader(targetReader)
		if err != nil {
			return errors.Wrapf(err, "failed to read file %q", targetPath)
		}
	}

	for registry, auth := range sourceConfig.GetAuthConfigs() {
		targetConfig.AuthConfigs[registry] = auth
	}

	buf := bytes.NewBuffer(nil)
	err = targetConfig.SaveToWriter(buf)
	if err != nil {
		return errors.Wrapf(err, "failed to save file %q", targetPath)
	}

	err = utils.MkdirAll(filepath.Dir(targetPath), 0777)
	if err != nil {
		return fmt.Errorf("failed to create directory path for the file %q: %w", targetPath, err)
	}
	err = utils.FileWrite(targetPath, buf.Bytes(), 0666)
	if err != nil {
		return fmt.Errorf("failed to write %q: %w", targetPath, err)
	}

	return nil
}

// CreateDockerConfigJSON creates / updates a Docker config.json with registry credentials
func CreateDockerConfigJSON(registryURL, username, password, targetPath, configPath string, utils piperutils.FileUtils) (string, error) {

	if len(targetPath) == 0 {
		targetPath = configPath
	}

	dockerConfigContent := []byte{}
	dockerConfig := map[string]interface{}{}
	if exists, err := utils.FileExists(configPath); exists {
		dockerConfigContent, err = utils.FileRead(configPath)
		if err != nil {
			return "", fmt.Errorf("failed to read file '%v': %w", configPath, err)
		}

		err = json.Unmarshal(dockerConfigContent, &dockerConfig)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal json file '%v': %w", configPath, err)
		}
	}

	if registryURL == "" || password == "" || username == "" {
		if err := fileWrite(targetPath, dockerConfigContent, utils); err != nil {
			return "", err
		}
		return targetPath, nil
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

	if err := fileWrite(targetPath, jsonResult, utils); err != nil {
		return "", err
	}

	return targetPath, nil
}

func fileWrite(path string, content []byte, utils piperutils.FileUtils) error {
	err := utils.MkdirAll(filepath.Dir(path), 0777)
	if err != nil {
		return fmt.Errorf("failed to create directory path for the Docker config.json file %v:%w", path, err)
	}
	err = utils.FileWrite(path, content, 0666)
	if err != nil {
		return fmt.Errorf("failed to write Docker config.json: %w", err)
	}

	return nil
}

// Client defines an docker client object
type Client struct {
	imageName     string
	registryURL   string
	localPath     string
	includeLayers bool
	imageFormat   string
}

// ClientOptions defines the options to be set on the client
type ClientOptions struct {
	ImageName   string
	RegistryURL string
	LocalPath   string
	ImageFormat string
}

// Download interface for download an image to a local path
type Download interface {
	DownloadImage(imageSource, targetFile string) (v1.Image, error)
	DownloadImageContent(imageSource, targetDir string) (v1.Image, error)
	GetRemoteImageInfo(string) (v1.Image, error)
}

// SetOptions sets options used for the docker client
func (c *Client) SetOptions(options ClientOptions) {
	c.imageName = options.ImageName
	c.registryURL = options.RegistryURL
	c.localPath = options.LocalPath
	c.imageFormat = options.ImageFormat
}

// DownloadImageContent downloads the image content into the given targetDir. Returns with an error if the targetDir doesnt exist
func (c *Client) DownloadImageContent(imageSource, targetDir string) (v1.Image, error) {
	if fileInfo, err := os.Stat(targetDir); err != nil {
		return nil, err
	} else if !fileInfo.IsDir() {
		return nil, fmt.Errorf("specified target is not a directory: %s", targetDir)
	}

	noOpts := []crane.Option{}

	imageRef, err := c.getImageRef(imageSource)
	if err != nil {
		return nil, err
	}

	img, err := crane.Pull(imageRef.Name(), noOpts...)
	if err != nil {
		return nil, err
	}

	tmpFile, err := os.CreateTemp(".", ".piper-download-")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())

	args := []string{imageRef.Name(), tmpFile.Name()}

	exportCmd := cranecmd.NewCmdExport(&noOpts)
	exportCmd.SetArgs(args)

	if err := exportCmd.Execute(); err != nil {
		return nil, err
	}

	return img, piperutils.Untar(tmpFile.Name(), targetDir, 0)
}

// DownloadImage downloads the image and saves it as tar at the given path
func (c *Client) DownloadImage(imageSource, targetFile string) (v1.Image, error) {
	noOpts := []crane.Option{}

	imageRef, err := c.getImageRef(imageSource)
	if err != nil {
		return nil, err
	}

	img, err := crane.Pull(imageRef.Name(), noOpts...)
	if err != nil {
		return nil, err
	}

	tmpFile, err := os.CreateTemp(".", ".piper-download-")
	if err != nil {
		return nil, err
	}

	craneCmd := cranecmd.NewCmdPull(&noOpts)
	craneCmd.SetOut(log.Writer())
	craneCmd.SetErr(log.Writer())
	args := []string{imageRef.Name(), tmpFile.Name(), "--format=" + c.imageFormat}
	craneCmd.SetArgs(args)

	if err := craneCmd.Execute(); err != nil {
		defer os.Remove(tmpFile.Name())
		return nil, err
	}

	if err := os.Rename(tmpFile.Name(), targetFile); err != nil {
		defer os.Remove(tmpFile.Name())
		return nil, err
	}

	return img, nil
}

// GetRemoteImageInfo retrieves information about the image (e.g. digest) without actually downoading it
func (c *Client) GetRemoteImageInfo(imageSource string) (v1.Image, error) {
	ref, err := c.getImageRef(imageSource)
	if err != nil {
		return nil, errors.Wrap(err, "parsing image reference")
	}

	return remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
}

func (c *Client) getImageRef(image string) (name.Reference, error) {
	opts := []name.Option{}
	registry := ""

	if len(c.registryURL) > 0 {
		re := regexp.MustCompile(`(?i)^https?://`)
		registry = re.ReplaceAllString(c.registryURL, "")
		opts = append(opts, name.WithDefaultRegistry(registry))
	}

	return name.ParseReference(path.Join(registry, image), opts...)
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
		// dockerfilePath = strings.ReplaceAll(dockerfilePath, cwd, ".")

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
