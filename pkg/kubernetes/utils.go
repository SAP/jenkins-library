package kubernetes

import (
	"errors"
	"fmt"
	"io"

	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"gopkg.in/yaml.v3"

	"github.com/SAP/jenkins-library/pkg/log"
)

// DeployUtils interface
type DeployUtils interface {
	SetEnv(env []string)
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	RunExecutable(e string, p ...string) error

	piperutils.FileUtils
	piperhttp.Uploader
	piperhttp.Downloader
}

// deployUtilsBundle struct  for utils
type deployUtilsBundle struct {
	*command.Command
	*piperutils.Files
	*piperhttp.Client
}

// NewDeployUtilsBundle initialize using deployUtilsBundle struct
func NewDeployUtilsBundle(customTLSCertificateLinks []string) DeployUtils {
	httpClientOptions := piperhttp.ClientOptions{}

	if len(customTLSCertificateLinks) > 0 {
		httpClientOptions.TransportSkipVerification = false
		httpClientOptions.TrustedCerts = customTLSCertificateLinks
	}

	httpClient := piperhttp.Client{}
	httpClient.SetOptions(httpClientOptions)

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
			StepName: "helmExecute",
		},
		Files:  &piperutils.Files{},
		Client: &piperhttp.Client{},
	}
	// reroute stderr output to logging framework, stdout will be used for command interactions
	utils.Stderr(log.Writer())
	return &utils
}

// GetChartInfo is used to get name and version of helm chart
func GetChartInfo(chartYamlFile string, utils DeployUtils) (string, string, error) {

	var result map[string]any
	p, err := utils.FileRead(chartYamlFile)
	if err != nil {
		return "", "", fmt.Errorf("file couldn't read: %w", err)
	}
	err = yaml.Unmarshal(p, &result)
	if err != nil {
		return "", "", fmt.Errorf("failed unmarshal: %w", err)
	}

	name, ok := result["name"].(string)
	if !ok || len(name) == 0 {
		return "", "", errors.New("name not found in chart yaml file (or wrong type)")
	}

	version, ok := result["version"].(string)
	if !ok || len(name) == 0 {
		return "", "", errors.New("version not found in chart yaml file (or wrong type)")
	}

	return name, version, nil
}
