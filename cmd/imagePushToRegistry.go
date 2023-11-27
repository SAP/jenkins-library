package cmd

import (
	"fmt"
	"regexp"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/pkg/errors"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/docker"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

const (
	targetDockerConfigPath = "/root/.docker/config.json"
)

type dockerImageUtils interface {
	LoadImage(src string) (v1.Image, error)
	PushImage(im v1.Image, dest string) error
	CopyImage(src, dest string) error
}

type imagePushToRegistryUtils interface {
	command.ExecRunner
	piperutils.FileUtils
	dockerImageUtils

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The imagePushToRegistryUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type imagePushToRegistryUtilsBundle struct {
	*command.Command
	*piperutils.Files
	dockerImageUtils

	// Embed more structs as necessary to implement methods or interfaces you add to imagePushToRegistryUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// imagePushToRegistryUtilsBundle and forward to the implementation of the dependency.
}

func newImagePushToRegistryUtils() imagePushToRegistryUtils {
	utils := imagePushToRegistryUtilsBundle{
		Command:          &command.Command{},
		Files:            &piperutils.Files{},
		dockerImageUtils: &docker.CraneUtilsBundle{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func imagePushToRegistry(config imagePushToRegistryOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newImagePushToRegistryUtils()

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runImagePushToRegistry(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runImagePushToRegistry(config *imagePushToRegistryOptions, telemetryData *telemetry.CustomData, utils imagePushToRegistryUtils) error {
	if len(config.TargetImages) == 0 {
		config.TargetImages = config.SourceImages
	}

	if len(config.TargetImages) != len(config.SourceImages) {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.New("err")
	}

	re := regexp.MustCompile(`^https?://`)
	config.SourceRegistryURL = re.ReplaceAllString(config.SourceRegistryURL, "")
	config.TargetRegistryURL = re.ReplaceAllString(config.TargetRegistryURL, "")

	err := handleCredentialsForPrivateRegistry(config.DockerConfigJSON, config.SourceRegistryURL, config.SourceRegistryUser, config.SourceRegistryPassword, utils)
	if err != nil {
		return errors.Wrap(err, "failed to handle credentials for source registry")
	}

	err = handleCredentialsForPrivateRegistry(config.DockerConfigJSON, config.TargetRegistryURL, config.TargetRegistryUser, config.TargetRegistryPassword, utils)
	if err != nil {
		return errors.Wrap(err, "failed to handle credentials for target registry")
	}

	if len(config.LocalDockerImagePath) > 0 {
		if err := pushLocalImageToTargetRegistry(config, utils); err != nil {
			return errors.Wrapf(err, "failed to push local image to %q", config.TargetRegistryURL)
		}
		return nil
	}

	if err := copyImages(config, utils); err != nil {
		return errors.Wrap(err, "failed to copy images")
	}

	return nil
}

func handleCredentialsForPrivateRegistry(dockerConfigJsonPath, registry, username, password string, utils imagePushToRegistryUtils) error {
	if len(dockerConfigJsonPath) == 0 && (len(registry) == 0 || len(username) == 0 || len(password) == 0) {
		return nil
	}

	if len(dockerConfigJsonPath) == 0 {
		_, err := docker.CreateDockerConfigJSON(registry, username, password, "", targetDockerConfigPath, utils)
		if err != nil {
			return errors.Wrap(err, "failed to create new docker config")
		}
		return nil
	}

	_, err := docker.CreateDockerConfigJSON(registry, username, password, targetDockerConfigPath, dockerConfigJsonPath, utils)
	if err != nil {
		return errors.Wrapf(err, "failed to update docker config %q", dockerConfigJsonPath)
	}

	if err := docker.MergeDockerConfigJSON(targetDockerConfigPath, dockerConfigJsonPath, utils); err != nil {
		return errors.Wrapf(err, "failed to merge docker config files")
	}

	return nil
}

func copyImages(config *imagePushToRegistryOptions, utils imagePushToRegistryUtils) error {
	for i := 0; i < len(config.SourceImages); i++ {
		src := fmt.Sprintf("%s/%s", config.SourceRegistryURL, config.SourceImages[i])
		dst := fmt.Sprintf("%s/%s", config.TargetRegistryURL, config.TargetImages[i])

		if err := copyImage(src, dst, config.TagLatest, utils); err != nil {
			return err
		}
	}

	return nil
}

func copyImage(src, dst string, tagLatest bool, utils imagePushToRegistryUtils) error {
	if tagLatest {
		// imageName is repository + image, e.g test.registry/testImage
		imageName, _ := parseDockerImage(dst)
		if err := utils.CopyImage(src, imageName); err != nil {
			return err
		}
	}

	return utils.CopyImage(src, dst)
}

func pushLocalImageToTargetRegistry(config *imagePushToRegistryOptions, utils imagePushToRegistryUtils) error {
	img, err := utils.LoadImage(config.LocalDockerImagePath)
	if err != nil {
		return err
	}

	for i := 0; i < len(config.TargetImages); i++ {
		dst := fmt.Sprintf("%s/%s", config.TargetRegistryURL, config.TargetImages[i])
		if err := utils.PushImage(img, dst); err != nil {
			return err
		}

		if config.TagLatest {
			// imageName is repository + image, e.g test.registry/testImage
			imageName, _ := parseDockerImage(dst)
			if err := utils.PushImage(img, imageName); err != nil {
				return err
			}
		}
	}

	return nil
}

func parseDockerImage(image string) (string, string) {
	re := regexp.MustCompile(`^(.*?)(?::([^:/]+))?$`)

	matches := re.FindStringSubmatch(image)
	if len(matches) > 1 {
		return matches[1], matches[2]
	}

	return image, ""
}

// ???
func skopeoMoveImage(sourceImageFullName string, sourceRegistryUser string, sourceRegistryPassword string, targetImageFullName string, targetRegistryUser string, targetRegistryPassword string, utils imagePushToRegistryUtils) error {
	skopeoRunParameters := []string{
		"copy",
		"--multi-arch=all",
		"--src-tls-verify=false",
	}
	if len(sourceRegistryUser) > 0 && len(sourceRegistryPassword) > 0 {
		skopeoRunParameters = append(skopeoRunParameters, fmt.Sprintf("--src-creds=%s:%s", sourceRegistryUser, sourceRegistryPassword))
	}
	skopeoRunParameters = append(skopeoRunParameters, "--src-tls-verify=false")
	if len(targetRegistryUser) > 0 && len(targetRegistryPassword) > 0 {
		skopeoRunParameters = append(skopeoRunParameters, fmt.Sprintf("--dest-creds=%s:%s", targetRegistryUser, targetRegistryPassword))

	}
	skopeoRunParameters = append(skopeoRunParameters, fmt.Sprintf("docker://%s docker://%s", sourceImageFullName, targetImageFullName))
	err := utils.RunExecutable("skopeo", skopeoRunParameters...)
	if err != nil {
		return err
	}
	return nil
}
