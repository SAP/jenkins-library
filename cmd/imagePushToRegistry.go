package cmd

import (
	"context"
	"fmt"
	"regexp"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

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
	LoadImage(ctx context.Context, src string) (v1.Image, error)
	PushImage(ctx context.Context, im v1.Image, dest, platform string) error
	CopyImage(ctx context.Context, src, dest, platform string) error
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
		Command: &command.Command{
			StepName: "imagePushToRegistry",
		},
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
		return errors.New("configuration error: please configure targetImage and sourceImage properly")
	}

	re := regexp.MustCompile(`^https?://`)
	config.SourceRegistryURL = re.ReplaceAllString(config.SourceRegistryURL, "")
	config.TargetRegistryURL = re.ReplaceAllString(config.TargetRegistryURL, "")

	log.Entry().Debug("Handling source registry credentials")
	if err := handleCredentialsForPrivateRegistry(config.DockerConfigJSON, config.SourceRegistryURL, config.SourceRegistryUser, config.SourceRegistryPassword, utils); err != nil {
		return errors.Wrap(err, "failed to handle credentials for source registry")
	}

	log.Entry().Debug("Handling destination registry credentials")
	if err := handleCredentialsForPrivateRegistry(config.DockerConfigJSON, config.TargetRegistryURL, config.TargetRegistryUser, config.TargetRegistryPassword, utils); err != nil {
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
		return errors.New("docker credentials not provided")
	}

	if len(dockerConfigJsonPath) == 0 {
		if _, err := docker.CreateDockerConfigJSON(registry, username, password, "", targetDockerConfigPath, utils); err != nil {
			return errors.Wrap(err, "failed to create new docker config")
		}
		return nil
	}

	if _, err := docker.CreateDockerConfigJSON(registry, username, password, targetDockerConfigPath, dockerConfigJsonPath, utils); err != nil {
		return errors.Wrapf(err, "failed to update docker config %q", dockerConfigJsonPath)
	}

	if err := docker.MergeDockerConfigJSON(targetDockerConfigPath, dockerConfigJsonPath, utils); err != nil {
		return errors.Wrapf(err, "failed to merge docker config files")
	}

	return nil
}

func copyImages(config *imagePushToRegistryOptions, utils imagePushToRegistryUtils) error {
	g, ctx := errgroup.WithContext(context.Background())
	platform := config.TargetArchitecture

	for i := 0; i < len(config.SourceImages); i++ {
		src := fmt.Sprintf("%s/%s", config.SourceRegistryURL, config.SourceImages[i])
		dst := fmt.Sprintf("%s/%s", config.TargetRegistryURL, config.TargetImages[i])

		g.Go(func() error {
			log.Entry().Infof("Copying %s to %s...", src, dst)
			if err := utils.CopyImage(ctx, src, dst, platform); err != nil {
				return err
			}
			log.Entry().Infof("Copying %s to %s... Done", src, dst)
			return nil
		})

		if config.TagLatest {
			g.Go(func() error {
				// imageName is repository + image, e.g test.registry/testImage
				imageName := parseDockerImageName(dst)
				log.Entry().Infof("Copying %s to %s...", src, imageName)
				if err := utils.CopyImage(ctx, src, imageName, platform); err != nil {
					return err
				}
				log.Entry().Infof("Copying %s to %s... Done", src, imageName)
				return nil
			})
		}
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func pushLocalImageToTargetRegistry(config *imagePushToRegistryOptions, utils imagePushToRegistryUtils) error {
	g, ctx := errgroup.WithContext(context.Background())
	platform := config.TargetArchitecture

	log.Entry().Infof("Loading local image...")
	img, err := utils.LoadImage(ctx, config.LocalDockerImagePath)
	if err != nil {
		return err
	}
	log.Entry().Infof("Loading local image... Done")

	for i := 0; i < len(config.TargetImages); i++ {
		i := i // https://golang.org/doc/faq#closures_and_goroutines
		dst := fmt.Sprintf("%s/%s", config.TargetRegistryURL, config.TargetImages[i])

		g.Go(func() error {
			log.Entry().Infof("Pushing %s...", dst)
			if err := utils.PushImage(ctx, img, dst, platform); err != nil {
				return err
			}
			log.Entry().Infof("Pushing %s... Done", dst)
			return nil
		})

		if config.TagLatest {
			g.Go(func() error {
				// imageName is repository + image, e.g test.registry/testImage
				imageName := parseDockerImageName(dst)
				log.Entry().Infof("Pushing %s...", imageName)
				if err := utils.PushImage(ctx, img, imageName, platform); err != nil {
					return err
				}
				log.Entry().Infof("Pushing %s... Done", imageName)
				return nil
			})
		}
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func parseDockerImageName(image string) string {
	re := regexp.MustCompile(`^(.*?)(?::([^:/]+))?$`)
	matches := re.FindStringSubmatch(image)
	if len(matches) > 1 {
		fmt.Println(matches[0], matches[1], matches[2])
		return matches[1]
	}

	return image
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
