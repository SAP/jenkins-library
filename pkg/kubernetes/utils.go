package kubernetes

import (
	"fmt"
	"io"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/docker"
	"github.com/SAP/jenkins-library/pkg/piperutils"

	"github.com/SAP/jenkins-library/pkg/log"
)

// DeployUtils interface
type DeployUtils interface {
	SetEnv(env []string)
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	RunExecutable(e string, p ...string) error

	piperutils.FileUtils
}

// deployUtilsBundle struct  for utils
type deployUtilsBundle struct {
	*command.Command
	*piperutils.Files
}

// NewDeployUtilsBundle initialize using deployUtilsBundle struct
func NewDeployUtilsBundle() DeployUtils {
	utils := deployUtilsBundle{
		Command: &command.Command{
			ErrorCategoryMapping: map[string][]string{
				log.ErrorConfiguration.String(): {
					"Error: Get * no such host",
					"Error: path * not found",
					"Error: rendered manifests contain a resource that already exists.",
					"Error: unknown flag",
					"Error: UPGRADE FAILED: * failed to replace object: * is invalid",
					"Error: UPGRADE FAILED: * failed to create resource: * is invalid",
					"Error: UPGRADE FAILED: an error occurred * not found",
					"Error: UPGRADE FAILED: query: failed to query with labels:",
					"Invalid value: \"\": field is immutable",
				},
				log.ErrorCustom.String(): {
					"Error: release * failed, * timed out waiting for the condition",
				},
			},
		},
		Files: &piperutils.Files{},
	}
	// reroute stderr output to logging framework, stdout will be used for command interactions
	utils.Stderr(log.Writer())
	return &utils
}

func getContainerInfo(config HelmExecuteOptions) (map[string]string, error) {
	var err error
	containerRegistry, err := docker.ContainerRegistryFromURL(config.ContainerRegistryURL)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Container registry url '%v' incorrect", config.ContainerRegistryURL)
	}

	//support either image or containerImageName and containerImageTag
	containerInfo := map[string]string{
		"containerImageName": "",
		"containerImageTag":  "",
		"containerRegistry":  containerRegistry,
	}

	if len(config.Image) > 0 {
		ref, err := docker.ContainerImageNameTagFromImage(config.Image)
		if err != nil {
			log.Entry().WithError(err).Fatalf("Container image '%v' incorrect", config.Image)
		}
		parts := strings.Split(ref, ":")
		containerInfo["containerImageName"] = parts[0]
		containerInfo["containerImageTag"] = parts[1]
	} else if len(config.ContainerImageName) > 0 && len(config.ContainerImageTag) > 0 {
		containerInfo["containerImageName"] = config.ContainerImageName
		containerInfo["containerImageTag"] = config.ContainerImageTag
	} else {
		return nil, fmt.Errorf("image information not given - please either set image or containerImageName and containerImageTag")
	}

	return containerInfo, nil
}
