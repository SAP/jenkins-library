package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/pkg/errors"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/docker"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func kanikoExecute(config kanikoExecuteOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *kanikoExecuteCommonPipelineEnvironment) {
	// for command execution use Command
	c := command.Command{
		ErrorCategoryMapping: map[string][]string{
			log.ErrorConfiguration.String(): {
				"unsupported status code 401",
			},
		},
	}

	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	client := &piperhttp.Client{}

	fileUtils := &piperutils.Files{}

	err := runKanikoExecute(&config, telemetryData, commonPipelineEnvironment, &c, client, fileUtils)
	if err != nil {
		log.Entry().WithError(err).Fatal("Kaniko execution failed")
	}
}

func runKanikoExecute(config *kanikoExecuteOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *kanikoExecuteCommonPipelineEnvironment, execRunner command.ExecRunner, httpClient piperhttp.Sender, fileUtils piperutils.FileUtils) error {
	// backward compatibility for parameter ContainerBuildOptions
	if len(config.ContainerBuildOptions) > 0 {
		config.BuildOptions = strings.Split(config.ContainerBuildOptions, " ")
		log.Entry().Warning("Parameter containerBuildOptions is deprecated, please use buildOptions instead.")
		telemetryData.Custom1Label = "ContainerBuildOptions"
		telemetryData.Custom1 = config.ContainerBuildOptions
	}

	// prepare kaniko container for running with proper Docker config.json and custom certificates
	// custom certificates will be downloaded and appended to ca-certificates.crt file used in container
	prepCommand := strings.Split(config.ContainerPreparationCommand, " ")
	if err := execRunner.RunExecutable(prepCommand[0], prepCommand[1:]...); err != nil {
		return errors.Wrap(err, "failed to initialize Kaniko container")
	}

	err := certificateUpdate(config.CustomTLSCertificateLinks, httpClient, fileUtils)
	if err != nil {
		return errors.Wrap(err, "failed to update certificates")
	}

	if !piperutils.ContainsString(config.BuildOptions, "--destination") {
		dest := []string{"--no-push"}
		if len(config.ContainerImage) > 0 {
			containerRegistry, err := docker.ContainerRegistryFromImage(config.ContainerImage)
			if err != nil {
				return errors.Wrapf(err, "invalid registry part in image %v", config.ContainerImage)
			}
			// errors are already caught with previous call to docker.ContainerRegistryFromImage
			containerImageNameTag, _ := docker.ContainerImageNameTagFromImage(config.ContainerImage)
			dest = []string{"--destination", config.ContainerImage}
			commonPipelineEnvironment.container.registryURL = fmt.Sprintf("https://%v", containerRegistry)
			commonPipelineEnvironment.container.imageNameTag = containerImageNameTag
		}
		if len(config.ContainerRegistryURL) > 0 && len(config.ContainerImageName) > 0 && len(config.ContainerImageTag) > 0 {
			containerRegistry, err := docker.ContainerRegistryFromURL(config.ContainerRegistryURL)
			if err != nil {
				return errors.Wrapf(err, "failed to read registry url %v", config.ContainerRegistryURL)
			}
			containerImageTag := fmt.Sprintf("%v:%v", config.ContainerImageName, strings.ReplaceAll(config.ContainerImageTag, "+", "-"))
			dest = []string{"--destination", fmt.Sprintf("%v/%v", containerRegistry, containerImageTag)}
			commonPipelineEnvironment.container.registryURL = config.ContainerRegistryURL
			commonPipelineEnvironment.container.imageNameTag = containerImageTag
		}
		config.BuildOptions = append(config.BuildOptions, dest...)
	}

	dockerConfig := []byte(`{"auths":{}}`)
	if len(config.DockerConfigJSON) > 0 {
		dockerConfig, err = fileUtils.FileRead(config.DockerConfigJSON)
		if err != nil {
			return errors.Wrapf(err, "failed to read file '%v'", config.DockerConfigJSON)
		}
	}

	if err := fileUtils.FileWrite("/kaniko/.docker/config.json", dockerConfig, 0644); err != nil {
		return errors.Wrap(err, "failed to write file '/kaniko/.docker/config.json'")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "failed to get current working directory")
	}
	kanikoOpts := []string{"--dockerfile", config.DockerfilePath, "--context", cwd}
	kanikoOpts = append(kanikoOpts, config.BuildOptions...)

	err = execRunner.RunExecutable("/kaniko/executor", kanikoOpts...)
	if err != nil {
		return errors.Wrap(err, "execution of '/kaniko/executor' failed")
	}
	return nil
}

func certificateUpdate(certLinks []string, httpClient piperhttp.Sender, fileUtils piperutils.FileUtils) error {
	caCertsFile := "/kaniko/ssl/certs/ca-certificates.crt"
	caCerts, err := fileUtils.FileRead(caCertsFile)
	if err != nil {
		return errors.Wrapf(err, "failed to load file '%v'", caCertsFile)
	}
	for _, link := range certLinks {
		response, err := httpClient.SendRequest(http.MethodGet, link, nil, nil, nil)
		if err != nil {
			return errors.Wrap(err, "failed to load certificate from url")
		}

		content, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return errors.Wrap(err, "error reading response")
		}
		response.Body.Close()
		content = append(content, []byte("\n")...)
		caCerts = append(caCerts, content...)
	}
	err = fileUtils.FileWrite(caCertsFile, caCerts, 0644)
	if err != nil {
		return errors.Wrapf(err, "failed to update file '%v'", caCertsFile)
	}
	return nil
}
