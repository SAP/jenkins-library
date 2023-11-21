package cmd

import (
	"fmt"
	"regexp"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/docker"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

const (
	targetDockerConfigPath = "/root/.docker/config.json"
)

type imagePushToRegistryUtils interface {
	command.ExecRunner

	FileExists(filename string) (bool, error)

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
	fileUtils := &piperutils.Files{}

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runImagePushToRegistry(&config, telemetryData, utils, fileUtils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runImagePushToRegistry(config *imagePushToRegistryOptions, telemetryData *telemetry.CustomData, utils imagePushToRegistryUtils, fileUtils piperutils.FileUtils) error {
	re := regexp.MustCompile(`^https?://`)
	sourceRegistry := re.ReplaceAllString(config.SourceRegistryURL, "")
	targetRegistry := re.ReplaceAllString(config.TargetRegistryURL, "")

	err := handleCredentialsForPrivateRegistries(config.DockerConfigJSON, sourceRegistry, config.SourceRegistryUser, config.SourceRegistryPassword, fileUtils)
	if err != nil {
		return fmt.Errorf("failed to handle registry credentials for source registry: %w", err)
	}

	err = handleCredentialsForPrivateRegistries(config.DockerConfigJSON, targetRegistry, config.TargetRegistryUser, config.TargetRegistryPassword, fileUtils)
	if err != nil {
		return fmt.Errorf("failed to handle registry credentials for target registry: %w", err)
	}

	if len(config.LocalDockerImagePath) > 0 {
		err = pushLocalImageToTargetRegistry(config.LocalDockerImagePath, config.TargetRegistryURL)
		if err != nil {
			return fmt.Errorf("failed to push to local image to registry: %w", err)
		}
	} else {
		src := fmt.Sprintf("%s/%s", sourceRegistry, config.SourceImageNameTag)
		dst := fmt.Sprintf("%s/%s", targetRegistry, config.SourceImageNameTag)
		err = copyImage(src, dst)
		if err != nil {
			return fmt.Errorf("failed to copy image from %v to %v with err: %w", config.SourceRegistryURL, config.TargetRegistryURL, err)
		}
	}

	log.Entry().WithField("LogField", "Log field content").Info("This is just a demo for a simple step.")

	// Example of calling methods from external dependencies directly on utils:
	// exists, err := utils.FileExists("file.txt")
	// if err != nil {
	// It is good practice to set an error category.
	// Most likely you want to do this at the place where enough context is known.
	// log.SetErrorCategory(log.ErrorConfiguration)
	// Always wrap non-descriptive errors to enrich them with context for when they appear in the log:
	// return fmt.Errorf("failed to check for important file: %w", err)
	// }
	// if !exists {
	// log.SetErrorCategory(log.ErrorConfiguration)
	// return fmt.Errorf("cannot run without important file")
	// }

	return nil

}

func handleCredentialsForPrivateRegistries(dockerConfigJsonPath string, registryURL string, username string, password string, fileUtils piperutils.FileUtils) error {
	if len(dockerConfigJsonPath) == 0 && (len(registryURL) == 0 || len(username) == 0 || len(password) == 0) {
		return nil
	}

	if len(dockerConfigJsonPath) == 0 {
		_, err := docker.CreateDockerConfigJSON(registryURL, username, password, "", targetDockerConfigPath, fileUtils)
		if err != nil {
			return errors.Wrap(err, "failed to create new docker config json at .docker/config.json")
		}
		log.Entry().Debug("Docker config has been created/updated")
		return nil
	}

	_, err := docker.CreateDockerConfigJSON(registryURL, username, password, targetDockerConfigPath, dockerConfigJsonPath, fileUtils)
	if err != nil {
		return errors.Wrapf(err, "failed to update existing docker config json file '%v'", dockerConfigJsonPath)
	}

	err = docker.MergeDockerConfigJSON(targetDockerConfigPath, dockerConfigJsonPath, fileUtils)
	if err != nil {
		return errors.Wrapf(err, "failed to merge docker config files '%v'", dockerConfigJsonPath)
	}
	log.Entry().Debug("Docker config has been created/updated")

	return nil
}

func pushLocalImageToTargetRegistry(localDockerImagePath string, targetRegistryURL string) error {
	img, err := docker.LoadImage(localDockerImagePath)
	if err != nil {
		return err
	}
	return docker.PushImage(img, targetRegistryURL)
}

func copyImage(sourceRegistry string, targetRegistry string) error {
	return docker.CopyImage(sourceRegistry, targetRegistry)
}

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

// https://f61e79198081-20231024-082222420-132.staging.repositories.cloud.sap
// password: ufrZ9a1Q4db0Lbv
// username: K3BGjVT5b1MvsZh
// sourceImage: azure-demo-k8s-go:0.1.0-20231024082209-cbe4e4e9e1f58f748fc8970671a59b4474b38ba8
