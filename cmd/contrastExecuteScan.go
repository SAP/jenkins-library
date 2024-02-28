package cmd

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/contrast"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type contrastExecuteScanUtils interface {
	command.ExecRunner
	piperutils.FileUtils
}

type contrastExecuteScanUtilsBundle struct {
	*command.Command
	*piperutils.Files
}

func newContrastExecuteScanUtils() contrastExecuteScanUtils {
	utils := contrastExecuteScanUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func contrastExecuteScan(config contrastExecuteScanOptions, telemetryData *telemetry.CustomData) {
	utils := newContrastExecuteScanUtils()

	reports, err := runContrastExecuteScan(&config, telemetryData, utils)
	piperutils.PersistReportsAndLinks("contrastExecuteScan", "./", utils, reports, nil)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func validateConfigs(config *contrastExecuteScanOptions) error {
	validations := map[string]string{
		"server":         config.Server,
		"organizationId": config.OrganizationID,
		"applicationId":  config.ApplicationID,
		"userApiKey":     config.UserAPIKey,
		"username":       config.Username,
		"serviceKey":     config.ServiceKey,
	}

	for k, v := range validations {
		if v == "" {
			return fmt.Errorf("%s is empty", k)
		}
	}

	if !strings.HasPrefix(config.Server, "https://") {
		config.Server = "https://" + config.Server
	}

	return nil
}

func runContrastExecuteScan(config *contrastExecuteScanOptions, telemetryData *telemetry.CustomData, utils contrastExecuteScanUtils) (reports []piperutils.Path, err error) {
	err = validateConfigs(config)
	if err != nil {
		log.Entry().Errorf("config is invalid: %v", err)
		return nil, err
	}

	auth := getAuth(config)
	appAPIUrl, appUIUrl := getApplicationUrls(config)

	contrastInstance := contrast.NewContrastInstance(appAPIUrl, config.UserAPIKey, auth)
	appInfo, err := contrastInstance.GetAppInfo(appUIUrl, config.Server)
	if err != nil {
		log.Entry().Errorf("error while getting app info")
		return nil, err
	}

	findings, err := contrastInstance.GetVulnerabilities()
	if err != nil {
		log.Entry().Errorf("error while getting vulns")
		return nil, err
	}

	contrastAudit := contrast.ContrastAudit{
		ToolName:       "contrast",
		ApplicationUrl: appInfo.Url,
		ScanResults:    findings,
	}
	paths, err := contrast.WriteJSONReport(contrastAudit, "./")
	if err != nil {
		log.Entry().Errorf("error while writing json report")
		return nil, err
	}
	reports = append(reports, paths...)

	if config.CheckForCompliance {
		for _, results := range findings {
			if results.ClassificationName == "Audit All" {
				unaudited := results.Total - results.Audited
				if unaudited > config.VulnerabilityThresholdTotal {
					msg := fmt.Sprintf("Your application %v in organization %v is not compliant. Total unaudited issues are %v which is greater than the VulnerabilityThresholdTotal count %v",
						config.ApplicationID, config.OrganizationID, unaudited, config.VulnerabilityThresholdTotal)
					return reports, fmt.Errorf(msg)
				}
			}
		}
	}

	toolRecordFileName, err := contrast.CreateAndPersistToolRecord(utils, appInfo, "./")
	if err != nil {
		log.Entry().Warning("TR_CONTRAST: Failed to create toolrecord file ...", err)
	} else {
		reports = append(reports, piperutils.Path{Target: toolRecordFileName})
	}

	return reports, nil
}

func getApplicationUrls(config *contrastExecuteScanOptions) (string, string) {
	appURL := fmt.Sprintf("%s/api/v4/organizations/%s/applications/%s", config.Server, config.OrganizationID, config.ApplicationID)
	guiURL := fmt.Sprintf("%s/Contrast/static/ng/index.html#/%s/applications/%s", config.Server, config.OrganizationID, config.ApplicationID)

	return appURL, guiURL
}

func getAuth(config *contrastExecuteScanOptions) string {
	return base64.StdEncoding.EncodeToString([]byte(config.Username + ":" + config.ServiceKey))
}
