package cmd

import (
	"fmt"
	"regexp"

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

type imagePushToRegistryUtils interface {
	command.ExecRunner
	piperutils.FileUtils

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The imagePushToRegistryUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type imagePushToRegistryUtilsBundle struct {
	*command.Command
	*piperutils.Files

	// Embed more structs as necessary to implement methods or interfaces you add to imagePushToRegistryUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// imagePushToRegistryUtilsBundle and forward to the implementation of the dependency.
}

func newImagePushToRegistryUtils() imagePushToRegistryUtils {
	utils := imagePushToRegistryUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
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
	re := regexp.MustCompile(`^https?://`)
	sourceRegistry := re.ReplaceAllString(config.SourceRegistryURL, "")
	targetRegistry := re.ReplaceAllString(config.TargetRegistryURL, "")
	src := fmt.Sprintf("%s/%s", sourceRegistry, config.SourceImage)
	dst := fmt.Sprintf("%s/%s", targetRegistry, config.SourceImage)

	err := handleCredentialsForPrivateRegistries(config.DockerConfigJSON, sourceRegistry, config.SourceRegistryUser, config.SourceRegistryPassword, utils)
	if err != nil {
		return errors.Wrap(err, "failed to handle credentials for source registry")
	}

	err = handleCredentialsForPrivateRegistries(config.DockerConfigJSON, targetRegistry, config.TargetRegistryUser, config.TargetRegistryPassword, utils)
	if err != nil {
		return errors.Wrap(err, "failed to handle credentials for target registry")
	}

	if len(config.LocalDockerImagePath) > 0 {
		if err := pushLocalImageToTargetRegistry(config.LocalDockerImagePath, dst); err != nil {
			return errors.Wrapf(err, "failed to push local image to %q", targetRegistry)
		}
		return nil
	}

	if err := copyImage(src, dst); err != nil {
		return errors.Wrapf(err, "failed to copy image from %q to %q", sourceRegistry, targetRegistry)
	}

	return nil
}

func handleCredentialsForPrivateRegistries(dockerConfigJsonPath, registry, username, password string, utils imagePushToRegistryUtils) error {
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

	err = docker.MergeDockerConfigJSON(targetDockerConfigPath, dockerConfigJsonPath, utils)
	if err != nil {
		return errors.Wrapf(err, "failed to merge docker config files")
	}

	return nil
}

func pushLocalImageToTargetRegistry(localDockerImagePath, targetRegistry string) error {
	img, err := docker.LoadImage(localDockerImagePath)
	if err != nil {
		return err
	}
	return docker.PushImage(img, targetRegistry)
}

func copyImage(sourceRegistry, targetRegistry string) error {
	return docker.CopyImage(sourceRegistry, targetRegistry)
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
