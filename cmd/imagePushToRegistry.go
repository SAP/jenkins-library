package cmd

import (
	"context"
	"fmt"
	"regexp"
	"strings"

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
	if !config.PushLocalDockerImage && !config.UseImageNameTags {
		if len(config.TargetImages) == 0 {
			config.TargetImages = mapSourceTargetImages(config.SourceImages)
		}
		if len(config.TargetImages) != len(config.SourceImages) {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.New("configuration error: please configure targetImage and sourceImage properly")
		}
	}

	if config.UseImageNameTags {
		if len(config.TargetImageNameTags) > 0 && len(config.TargetImageNameTags) != len(config.SourceImageNameTags) {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.New("configuration error: please configure targetImageNameTags and sourceImageNameTags properly")
		}
	}

	// Docker image tags don't allow plus signs in tags, thus replacing with dash
	config.SourceImageTag = strings.ReplaceAll(config.SourceImageTag, "+", "-")
	config.TargetImageTag = strings.ReplaceAll(config.TargetImageTag, "+", "-")
	re := regexp.MustCompile(`^https?://`)
	config.SourceRegistryURL = re.ReplaceAllString(config.SourceRegistryURL, "")
	config.TargetRegistryURL = re.ReplaceAllString(config.TargetRegistryURL, "")

	log.Entry().Debug("Handling destination registry credentials")
	if err := handleCredentialsForPrivateRegistry(config.DockerConfigJSON, config.TargetRegistryURL, config.TargetRegistryUser, config.TargetRegistryPassword, utils); err != nil {
		return errors.Wrap(err, "failed to handle credentials for target registry")
	}

	if config.PushLocalDockerImage {
		if err := pushLocalImageToTargetRegistry(config, utils); err != nil {
			return errors.Wrapf(err, "failed to push local image to %q", config.TargetRegistryURL)
		}
		return nil
	}

	log.Entry().Debug("Handling source registry credentials")
	if err := handleCredentialsForPrivateRegistry(config.DockerConfigJSON, config.SourceRegistryURL, config.SourceRegistryUser, config.SourceRegistryPassword, utils); err != nil {
		return errors.Wrap(err, "failed to handle credentials for source registry")
	}

	if config.UseImageNameTags {
		if err := pushImageNameTagsToTargetRegistry(config, utils); err != nil {
			return errors.Wrapf(err, "failed to push imageNameTags to target registry")
		}
		return nil
	}

	if err := copyImages(config, utils); err != nil {
		return errors.Wrap(err, "failed to copy images")
	}

	return nil
}

func handleCredentialsForPrivateRegistry(dockerConfigJsonPath, registry, username, password string, utils imagePushToRegistryUtils) error {
	if len(dockerConfigJsonPath) == 0 {
		if len(registry) == 0 || len(username) == 0 || len(password) == 0 {
			return errors.New("docker credentials not provided")
		}

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
	g.SetLimit(10)
	platform := config.TargetArchitecture

	for _, sourceImage := range config.SourceImages {
		sourceImage := sourceImage
		src := fmt.Sprintf("%s/%s:%s", config.SourceRegistryURL, sourceImage, config.SourceImageTag)

		targetImage, ok := config.TargetImages[sourceImage].(string)
		if !ok {
			return fmt.Errorf("incorrect name of target image: %v", config.TargetImages[sourceImage])
		}

		if config.TargetImageTag != "" {
			g.Go(func() error {
				dst := fmt.Sprintf("%s/%s:%s", config.TargetRegistryURL, targetImage, config.TargetImageTag)
				log.Entry().Infof("Copying %s to %s...", src, dst)
				if err := utils.CopyImage(ctx, src, dst, platform); err != nil {
					return err
				}
				log.Entry().Infof("Copying %s to %s... Done", src, dst)
				return nil
			})
		}

		if config.TagLatest {
			g.Go(func() error {
				dst := fmt.Sprintf("%s/%s", config.TargetRegistryURL, config.TargetImages[sourceImage])
				log.Entry().Infof("Copying %s to %s...", src, dst)
				if err := utils.CopyImage(ctx, src, dst, platform); err != nil {
					return err
				}
				log.Entry().Infof("Copying %s to %s... Done", src, dst)
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
	g.SetLimit(10)
	platform := config.TargetArchitecture

	log.Entry().Infof("Loading local image...")
	img, err := utils.LoadImage(ctx, config.LocalDockerImagePath)
	if err != nil {
		return err
	}
	log.Entry().Infof("Loading local image... Done")

	for _, trgImage := range config.TargetImages {
		trgImage := trgImage
		targetImage, ok := trgImage.(string)
		if !ok {
			return fmt.Errorf("incorrect name of target image: %v", trgImage)
		}

		if config.TargetImageTag != "" {
			g.Go(func() error {
				dst := fmt.Sprintf("%s/%s:%s", config.TargetRegistryURL, targetImage, config.TargetImageTag)
				log.Entry().Infof("Pushing %s...", dst)
				if err := utils.PushImage(ctx, img, dst, platform); err != nil {
					return err
				}
				log.Entry().Infof("Pushing %s... Done", dst)
				return nil
			})
		}

		if config.TagLatest {
			g.Go(func() error {
				dst := fmt.Sprintf("%s/%s", config.TargetRegistryURL, targetImage)
				log.Entry().Infof("Pushing %s...", dst)
				if err := utils.PushImage(ctx, img, dst, platform); err != nil {
					return err
				}
				log.Entry().Infof("Pushing %s... Done", dst)
				return nil
			})
		}
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func pushImageNameTagsToTargetRegistry(config *imagePushToRegistryOptions, utils imagePushToRegistryUtils) error {
	g, ctx := errgroup.WithContext(context.Background())
	g.SetLimit(10)

	for i, sourceImageNameTag := range config.SourceImageNameTags {
		src := fmt.Sprintf("%s/%s", config.SourceRegistryURL, sourceImageNameTag)

		dst := ""
		if len(config.TargetImageNameTags) == 0 {
			dst = fmt.Sprintf("%s/%s", config.TargetRegistryURL, sourceImageNameTag)
		} else {
			dst = fmt.Sprintf("%s/%s", config.TargetRegistryURL, config.TargetImageNameTags[i])
		}

		g.Go(func() error {
			log.Entry().Infof("Copying %s to %s...", src, dst)
			if err := utils.CopyImage(ctx, src, dst, ""); err != nil {
				return err
			}
			log.Entry().Infof("Copying %s to %s... Done", src, dst)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func mapSourceTargetImages(sourceImages []string) map[string]any {
	targetImages := make(map[string]any, len(sourceImages))
	for _, sourceImage := range sourceImages {
		targetImages[sourceImage] = sourceImage
	}

	return targetImages
}
