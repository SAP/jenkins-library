package cmd

import (
	"archive/zip"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/SAP/jenkins-library/pkg/certutils"
	"github.com/SAP/jenkins-library/pkg/cnbutils"
	"github.com/SAP/jenkins-library/pkg/cnbutils/bindings"
	"github.com/SAP/jenkins-library/pkg/cnbutils/project"
	"github.com/SAP/jenkins-library/pkg/cnbutils/project/metadata"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/docker"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
	ignore "github.com/sabhiram/go-gitignore"
)

const (
	creatorPath  = "/cnb/lifecycle/creator"
	platformPath = "/tmp/platform"
)

type cnbBuildUtilsBundle struct {
	*command.Command
	*piperutils.Files
	*docker.Client
}

func setCustomBuildpacks(bpacks []string, dockerCreds string, utils cnbutils.BuildUtils) (string, string, error) {
	buildpacksPath := "/tmp/buildpacks"
	orderPath := "/tmp/buildpacks/order.toml"
	newOrder, err := cnbutils.DownloadBuildpacks(buildpacksPath, bpacks, dockerCreds, utils)
	if err != nil {
		return "", "", err
	}

	err = newOrder.Save(orderPath)
	if err != nil {
		return "", "", err
	}

	return buildpacksPath, orderPath, nil
}

func newCnbBuildUtils() cnbutils.BuildUtils {
	utils := cnbBuildUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
		Client:  &docker.Client{},
	}
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func cnbBuild(config cnbBuildOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *cnbBuildCommonPipelineEnvironment) {
	utils := newCnbBuildUtils()

	client := &piperhttp.Client{}

	err := runCnbBuild(&config, telemetryData, utils, commonPipelineEnvironment, client)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func isIgnored(find string, include, exclude *ignore.GitIgnore) bool {
	if exclude != nil {
		filtered := exclude.MatchesPath(find)

		if filtered {
			log.Entry().Debugf("%s matches exclude pattern, ignoring", find)
			return true
		}
	}

	if include != nil {
		filtered := !include.MatchesPath(find)

		if filtered {
			log.Entry().Debugf("%s doesn't match include pattern, ignoring", find)
			return true
		} else {
			log.Entry().Debugf("%s matches include pattern", find)
			return false
		}
	}

	return false
}

func isBuilder(utils cnbutils.BuildUtils) error {
	exists, err := utils.FileExists(creatorPath)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("binary '%s' not found", creatorPath)
	}

	return nil
}

func isZip(path string) bool {
	r, err := zip.OpenReader(path)

	switch {
	case err == nil:
		_ = r.Close()
		return true
	case err == zip.ErrFormat:
		return false
	default:
		return false
	}
}

func copyFile(source, target string, utils cnbutils.BuildUtils) error {
	targetDir := filepath.Dir(target)

	exists, err := utils.DirExists(targetDir)
	if err != nil {
		return err
	}

	if !exists {
		log.Entry().Debugf("Creating directory %s", targetDir)
		err = utils.MkdirAll(targetDir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	_, err = utils.Copy(source, target)
	return err
}

func copyProject(source, target string, include, exclude *ignore.GitIgnore, utils cnbutils.BuildUtils) error {
	sourceFiles, _ := utils.Glob(path.Join(source, "**"))
	for _, sourceFile := range sourceFiles {
		relPath, err := filepath.Rel(source, sourceFile)
		if err != nil {
			log.SetErrorCategory(log.ErrorBuild)
			return errors.Wrapf(err, "Calculating relative path for '%s' failed", sourceFile)
		}
		if !isIgnored(relPath, include, exclude) {
			target := path.Join(target, strings.ReplaceAll(sourceFile, source, ""))
			dir, err := utils.DirExists(sourceFile)
			if err != nil {
				log.SetErrorCategory(log.ErrorBuild)
				return errors.Wrapf(err, "Checking file info '%s' failed", target)
			}

			if dir {
				err = utils.MkdirAll(target, os.ModePerm)
				if err != nil {
					log.SetErrorCategory(log.ErrorBuild)
					return errors.Wrapf(err, "Creating directory '%s' failed", target)
				}
			} else {
				log.Entry().Debugf("Copying '%s' to '%s'", sourceFile, target)
				err = copyFile(sourceFile, target, utils)
				if err != nil {
					log.SetErrorCategory(log.ErrorBuild)
					return errors.Wrapf(err, "Copying '%s' to '%s' failed", sourceFile, target)
				}
			}

		}
	}
	return nil
}

func extractZip(source, target string) error {
	if isZip(source) {
		log.Entry().Infof("Extracting archive '%s' to '%s'", source, target)
		_, err := piperutils.Unzip(source, target)
		if err != nil {
			log.SetErrorCategory(log.ErrorBuild)
			return errors.Wrapf(err, "Extracting archive '%s' to '%s' failed", source, target)
		}
	} else {
		log.SetErrorCategory(log.ErrorBuild)
		return errors.New("application path must be a directory or zip")
	}

	return nil
}

func prepareDockerConfig(source string, utils cnbutils.BuildUtils) (string, error) {
	if filepath.Base(source) != "config.json" {
		log.Entry().Debugf("Renaming docker config file from '%s' to 'config.json'", filepath.Base(source))

		newPath := filepath.Join(filepath.Dir(source), "config.json")
		err := utils.FileRename(source, newPath)
		if err != nil {
			return "", err
		}

		return newPath, nil
	}

	return source, nil
}

func linkTargetFolder(utils cnbutils.BuildUtils, source, target string) error {
	var err error
	linkPath := filepath.Join(target, "target")
	targetPath := filepath.Join(source, "target")
	if ok, _ := utils.DirExists(targetPath); !ok {
		err = utils.MkdirAll(targetPath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	if ok, _ := utils.DirExists(linkPath); ok {
		err = utils.RemoveAll(linkPath)
		if err != nil {
			return err
		}
	}

	return utils.Symlink(targetPath, linkPath)
}

func (c *cnbBuildOptions) mergeEnvVars(vars map[string]interface{}) {
	if c.BuildEnvVars == nil {
		c.BuildEnvVars = vars

		return
	}

	for k, v := range vars {
		_, exists := c.BuildEnvVars[k]

		if !exists {
			c.BuildEnvVars[k] = v
		}
	}
}

func runCnbBuild(config *cnbBuildOptions, telemetryData *telemetry.CustomData, utils cnbutils.BuildUtils, commonPipelineEnvironment *cnbBuildCommonPipelineEnvironment, httpClient piperhttp.Sender) error {
	var err error

	err = isBuilder(utils)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Wrap(err, "the provided dockerImage is not a valid builder")
	}

	include := ignore.CompileIgnoreLines("**/*")
	exclude := ignore.CompileIgnoreLines("piper", ".pipeline")

	projDescExists, err := utils.FileExists(config.ProjectDescriptor)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Wrap(err, "failed to check if project descriptor exists")
	}

	var projectID string
	if projDescExists {
		descriptor, err := project.ParseDescriptor(config.ProjectDescriptor, utils, httpClient)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.Wrapf(err, "failed to parse %s", config.ProjectDescriptor)
		}

		config.mergeEnvVars(descriptor.EnvVars)

		if (config.Buildpacks == nil || len(config.Buildpacks) == 0) && len(descriptor.Buildpacks) > 0 {
			config.Buildpacks = descriptor.Buildpacks
		}

		if descriptor.Exclude != nil {
			exclude = descriptor.Exclude
		}

		if descriptor.Include != nil {
			include = descriptor.Include
		}

		projectID = descriptor.ProjectID
	}

	targetImage, err := cnbutils.GetTargetImage(config.ContainerRegistryURL, config.ContainerImageName, config.ContainerImageTag, GeneralConfig.EnvRootPath, projectID)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Wrap(err, "failed to retrieve target image configuration")
	}

	commonPipelineEnvironment.container.registryURL = targetImage.ContainerRegistryURL
	commonPipelineEnvironment.container.imageNameTag = fmt.Sprintf("%v:%v", targetImage.ContainerImageName, targetImage.ContainerImageTag)

	if config.BuildEnvVars != nil && len(config.BuildEnvVars) > 0 {
		log.Entry().Infof("Setting custom environment variables: '%v'", config.BuildEnvVars)
		err = cnbutils.CreateEnvFiles(utils, platformPath, config.BuildEnvVars)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.Wrap(err, "failed to write environment variables to files")
		}
	}

	err = bindings.ProcessBindings(utils, httpClient, platformPath, config.Bindings)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Wrap(err, "failed process bindings")
	}

	dockerConfigFile := ""
	if len(config.DockerConfigJSON) > 0 {
		dockerConfigFile, err = prepareDockerConfig(config.DockerConfigJSON, utils)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.Wrapf(err, "failed to rename DockerConfigJSON file '%v'", config.DockerConfigJSON)
		}
	}

	target := "/workspace"
	pwd, err := utils.Getwd()
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		return errors.Wrap(err, "failed to get current working directory")
	}

	var source string
	if config.Path != "" {
		source = config.Path
	} else {
		source = pwd
	}

	dir, err := utils.DirExists(source)
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		return errors.Wrapf(err, "Checking file info '%s' failed", source)
	}

	if dir {
		err = copyProject(source, target, include, exclude, utils)
		if err != nil {
			log.SetErrorCategory(log.ErrorBuild)
			return errors.Wrapf(err, "Copying  '%s' into '%s' failed", source, target)
		}
	} else {
		err = extractZip(source, target)
		if err != nil {
			log.SetErrorCategory(log.ErrorBuild)
			return errors.Wrapf(err, "Copying  '%s' into '%s' failed", source, target)
		}
	}

	if ok, _ := utils.FileExists(filepath.Join(target, "pom.xml")); ok {
		err = linkTargetFolder(utils, pwd, target)
		if err != nil {
			log.SetErrorCategory(log.ErrorBuild)
			return err
		}
	}

	metadata.WriteProjectMetadata(GeneralConfig.EnvRootPath, utils)

	var buildpacksPath = "/cnb/buildpacks"
	var orderPath = "/cnb/order.toml"

	if config.Buildpacks != nil && len(config.Buildpacks) > 0 {
		log.Entry().Infof("Setting custom buildpacks: '%v'", config.Buildpacks)
		buildpacksPath, orderPath, err = setCustomBuildpacks(config.Buildpacks, dockerConfigFile, utils)
		defer utils.RemoveAll(buildpacksPath)
		defer utils.RemoveAll(orderPath)
		if err != nil {
			log.SetErrorCategory(log.ErrorBuild)
			return errors.Wrapf(err, "Setting custom buildpacks: %v", config.Buildpacks)
		}
	}

	cnbRegistryAuth, err := cnbutils.GenerateCnbAuth(dockerConfigFile, utils)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Wrap(err, "failed to generate CNB_REGISTRY_AUTH")
	}

	if dockerConfigFile != "" {
		err = utils.FileRemove(dockerConfigFile)
		if err != nil {
			log.SetErrorCategory(log.ErrorBuild)
			return errors.Wrap(err, "failed to remove docker config.json file")
		}
	}

	if len(config.CustomTLSCertificateLinks) > 0 {
		caCertificates := "/tmp/ca-certificates.crt"
		_, err := utils.Copy("/etc/ssl/certs/ca-certificates.crt", caCertificates)
		if err != nil {
			return errors.Wrap(err, "failed to copy certificates")
		}
		err = certutils.CertificateUpdate(config.CustomTLSCertificateLinks, httpClient, utils, caCertificates)
		if err != nil {
			return errors.Wrap(err, "failed to update certificates")
		}
		utils.AppendEnv([]string{fmt.Sprintf("SSL_CERT_FILE=%s", caCertificates)})
	} else {
		log.Entry().Info("skipping certificates update")
	}

	utils.AppendEnv([]string{fmt.Sprintf("CNB_REGISTRY_AUTH=%s", cnbRegistryAuth)})
	utils.AppendEnv([]string{"CNB_PLATFORM_API=0.8"})

	creatorArgs := []string{
		"-no-color",
		"-buildpacks", buildpacksPath,
		"-order", orderPath,
		"-platform", platformPath,
	}

	containerImage := path.Join(targetImage.ContainerRegistryURL, targetImage.ContainerImageName)
	for _, tag := range config.AdditionalTags {
		target := fmt.Sprintf("%s:%s", containerImage, tag)
		if !piperutils.ContainsString(creatorArgs, target) {
			creatorArgs = append(creatorArgs, "-tag", target)
		}
	}

	creatorArgs = append(creatorArgs, fmt.Sprintf("%s:%s", containerImage, targetImage.ContainerImageTag))
	err = utils.RunExecutable(creatorPath, creatorArgs...)
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		return errors.Wrapf(err, "execution of '%s' failed", creatorArgs)
	}

	return nil
}
