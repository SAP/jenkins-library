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

type dockerUtils interface {
	CreateDockerConfigJSON(registry, username, password, targetPath, configPath string, utils piperutils.FileUtils) (string, error)
	MergeDockerConfigJSON(sourcePath, targetPath string, utils piperutils.FileUtils) error
	LoadImage(src string) (v1.Image, error)
	PushImage(im v1.Image, dest string) error
	CopyImage(src, dest string) error
}

type dockerUtilsBundle struct{}

func (d *dockerUtilsBundle) CreateDockerConfigJSON(registry, username, password, targetPath, configPath string, utils piperutils.FileUtils) (string, error) {
	return docker.CreateDockerConfigJSON(registry, username, password, targetPath, configPath, utils)
}

func (d *dockerUtilsBundle) MergeDockerConfigJSON(sourcePath, targetPath string, utils piperutils.FileUtils) error {
	return docker.MergeDockerConfigJSON(sourcePath, targetPath, utils)
}

func (d *dockerUtilsBundle) LoadImage(src string) (v1.Image, error) {
	return docker.LoadImage(src)
}

func (d *dockerUtilsBundle) PushImage(im v1.Image, dest string) error {
	return docker.PushImage(im, dest)
}

func (d *imagePushToRegistryUtilsBundle) CopyImage(src, dest string) error {
	return docker.CopyImage(src, dest)
}

type imagePushToRegistryUtils interface {
	command.ExecRunner
	piperutils.FileUtils
	dockerUtils

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The imagePushToRegistryUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type imagePushToRegistryUtilsBundle struct {
	*command.Command
	*piperutils.Files
	*dockerUtilsBundle

	// Embed more structs as necessary to implement methods or interfaces you add to imagePushToRegistryUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// imagePushToRegistryUtilsBundle and forward to the implementation of the dependency.
}

func newImagePushToRegistryUtils() imagePushToRegistryUtils {
	utils := imagePushToRegistryUtilsBundle{
		Command:           &command.Command{},
		Files:             &piperutils.Files{},
		dockerUtilsBundle: &dockerUtilsBundle{},
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
	if config.TargetImage == "" {
		config.TargetImage = config.SourceImage
	}

	re := regexp.MustCompile(`^https?://`)
	sourceRegistry := re.ReplaceAllString(config.SourceRegistryURL, "")
	targetRegistry := re.ReplaceAllString(config.TargetRegistryURL, "")
	src := fmt.Sprintf("%s/%s", sourceRegistry, config.SourceImage)
	dst := fmt.Sprintf("%s/%s", targetRegistry, config.TargetImage)

	err := handleCredentialsForPrivateRegistry(config.DockerConfigJSON, sourceRegistry, config.SourceRegistryUser, config.SourceRegistryPassword, utils)
	if err != nil {
		return errors.Wrap(err, "failed to handle credentials for source registry")
	}

	err = handleCredentialsForPrivateRegistry(config.DockerConfigJSON, targetRegistry, config.TargetRegistryUser, config.TargetRegistryPassword, utils)
	if err != nil {
		return errors.Wrap(err, "failed to handle credentials for target registry")
	}

	if len(config.LocalDockerImagePath) > 0 {
		if err := pushLocalImageToTargetRegistry(config.LocalDockerImagePath, dst, utils); err != nil {
			return errors.Wrapf(err, "failed to push local image to %q", targetRegistry)
		}
		return nil
	}

	if err := copyImage(src, dst, utils); err != nil {
		return errors.Wrapf(err, "failed to copy image from %q to %q", sourceRegistry, targetRegistry)
	}

	return nil
}

func handleCredentialsForPrivateRegistry(dockerConfigJsonPath, registry, username, password string, utils imagePushToRegistryUtils) error {
	if len(dockerConfigJsonPath) == 0 && (len(registry) == 0 || len(username) == 0 || len(password) == 0) {
		return nil
	}

	if len(dockerConfigJsonPath) == 0 {
		_, err := utils.CreateDockerConfigJSON(registry, username, password, "", targetDockerConfigPath, utils)
		if err != nil {
			return errors.Wrap(err, "failed to create new docker config")
		}
		return nil
	}

	_, err := utils.CreateDockerConfigJSON(registry, username, password, targetDockerConfigPath, dockerConfigJsonPath, utils)
	if err != nil {
		return errors.Wrapf(err, "failed to update docker config %q", dockerConfigJsonPath)
	}

	err = utils.MergeDockerConfigJSON(targetDockerConfigPath, dockerConfigJsonPath, utils)
	if err != nil {
		return errors.Wrapf(err, "failed to merge docker config files")
	}

	return nil
}

func pushLocalImageToTargetRegistry(localDockerImagePath, targetRegistry string, utils imagePushToRegistryUtils) error {
	img, err := utils.LoadImage(localDockerImagePath)
	if err != nil {
		return err
	}
	return utils.PushImage(img, targetRegistry)
}

func copyImage(sourceRegistry, targetRegistry string, utils imagePushToRegistryUtils) error {
	return utils.CopyImage(sourceRegistry, targetRegistry)
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
