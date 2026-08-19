package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"errors"

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

	// Create unified Contrast client for both sync and async operations
	appAPIUrl, appUIUrl := getApplicationUrls(config)
	client := contrast.NewClient(config.UserAPIKey, config.ServiceKey, config.Username, config.OrganizationID, config.Server, appAPIUrl)

	// Pre-flight agent setup checks
	routeCoveragePct, err := checkAgentSetup(client, config)
	if err != nil {
		log.Entry().Errorf("Agent setup check failed: %v", err)
		return nil, err
	}

	// Generate SARIF report if boolean flag is set
	if config.GenerateSarif {
		sarifReports, sarifErr := generateSarifReport(config, utils, client)
		if sarifErr != nil {
			log.Entry().WithError(sarifErr).Error("Failed to generate SARIF report")
			return nil, sarifErr
		}
		reports = append(reports, sarifReports...)
	}

	// Generate PDF report if boolean flag is set
	if config.GeneratePdf {
		pdfReports, pdfErr := generatePdfReport(config, utils, client)
		if pdfErr != nil {
			log.Entry().WithError(pdfErr).Error("Failed to generate PDF report")
			return nil, pdfErr
		}
		reports = append(reports, pdfReports...)
	}
	appInfo, err := client.GetAppInfo(appUIUrl, config.Server)
	if err != nil {
		log.Entry().Errorf("error while getting app info")
		return nil, err
	}

	findings, err := client.GetVulnerabilities()
	if err != nil {
		log.Entry().Errorf("error while getting vulns")
		return nil, err
	}

	contrastAudit := contrast.ContrastAudit{
		ToolName:       "contrast",
		ApplicationUrl: appInfo.Url,
		ScanResults:    findings,
	}
	jsonData, err := json.Marshal(contrastAudit)
	if err != nil {
		log.Entry().Errorf("error while marshaling json report: %v", err)
		return nil, err
	}
	paths, err := contrast.SaveReportFile(utils, "piper_contrast_report.json", "Contrast JSON Compliance Report", jsonData)
	if err != nil {
		log.Entry().Errorf("error while writing json report: %v", err)
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
					return reports, errors.New(msg)
				}
			}
		}
	}

	toolRecordFileName, err := contrast.CreateAndPersistToolRecord(utils, appInfo, "./", routeCoveragePct)
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

// generateSarifReport generates a SARIF report using the Contrast API
func generateSarifReport(config *contrastExecuteScanOptions, utils contrastExecuteScanUtils, client *contrast.Client) ([]piperutils.Path, error) {
	log.Entry().Info("Starting SARIF report generation...")

	data, err := client.GenerateSarifReport(config.ApplicationID)
	if err != nil {
		return nil, fmt.Errorf("SARIF generation failed: %w", err)
	}

	return contrast.SaveReportFile(utils, "piper_contrast.sarif", "Contrast SARIF Report", data)
}

// generatePdfReport generates a PDF attestation report using the Contrast API
func generatePdfReport(config *contrastExecuteScanOptions, utils contrastExecuteScanUtils, client *contrast.Client) ([]piperutils.Path, error) {
	log.Entry().Info("Starting PDF report generation...")

	data, err := client.GeneratePdfReport(config.ApplicationID)
	if err != nil {
		return nil, fmt.Errorf("PDF generation failed: %w", err)
	}

	return contrast.SaveReportFile(utils, "piper_contrast_attestation.pdf", "Contrast PDF Attestation Report", data)
}

// checkAgentSetup performs agent pre-flight checks.
// Hard fails if no agents are connected to the application.
// Warns if all agents have been inactive beyond the configured threshold.
func checkAgentSetup(client *contrast.Client, config *contrastExecuteScanOptions) (*float64, error) {
	servers, err := client.GetServers(config.ApplicationID)
	if err != nil {
		// Non-fatal: don't break the pipeline on a transient API error.
		log.Entry().Warnf("Could not retrieve server list: %v", err)
		return nil, nil
	}

	// Check 1: Are there any servers?
	if len(servers) == 0 {
		return nil, fmt.Errorf("application %s has no agents connected in organization %s. "+
			"Please finish your Contrast agent setup before running this step.",
			config.ApplicationID, config.OrganizationID)
	}
	log.Entry().Infof("Agent check: %d server(s) connected to the application.", len(servers))

	// Check 2: When was the last server active?
	if contrast.MinAgentInactivityThresholdDays > 0 {
		latestActivity := int64(0)
		for _, srv := range servers {
			if srv.LastActivity > latestActivity {
				latestActivity = srv.LastActivity
			}
		}
		if latestActivity > 0 {
			lastSeen := time.UnixMilli(latestActivity)
			inactiveSince := time.Since(lastSeen)
			threshold := time.Duration(contrast.MinAgentInactivityThresholdDays) * 24 * time.Hour
			if inactiveSince > threshold {
				log.Entry().Warnf("Agent activity check: most recent server activity was %s ago (%s). "+
					"No agent has been active in the last %d day(s). "+
					"Results may be incomplete — consider restarting your agent.",
					inactiveSince.Round(time.Hour),
					lastSeen.Format(time.RFC3339),
					contrast.MinAgentInactivityThresholdDays,
				)
			} else {
				log.Entry().Infof("Agent activity check: last active %s ago.", inactiveSince.Round(time.Minute))
			}
		}
	}

	routeCoveragePct := checkRouteCoverage(client, config)

	return routeCoveragePct, nil
}

func checkRouteCoverage(client *contrast.Client, config *contrastExecuteScanOptions) *float64 {
	coverage, err := client.GetRouteCoverage(config.ApplicationID)
	if err != nil {
		log.Entry().Warnf("Could not retrieve route coverage: %v", err)
		return nil
	}
	if coverage.DiscoveredCount == 0 {
		return nil
	}
	exercisedPct := float64(coverage.ExercisedCount) / float64(coverage.DiscoveredCount) * 100
	if exercisedPct < contrast.MinRouteCoveragePct {
		log.Entry().Warnf("Route coverage check: only %.1f%% of discovered routes have been exercised (%d/%d). "+
			"Security findings may be incomplete — consider increasing test coverage.",
			exercisedPct, coverage.ExercisedCount, coverage.DiscoveredCount)
	} else {
		log.Entry().Infof("Route coverage check: %.1f%% of routes exercised (%d/%d).",
			exercisedPct, coverage.ExercisedCount, coverage.DiscoveredCount)
	}
	return &exercisedPct
}
