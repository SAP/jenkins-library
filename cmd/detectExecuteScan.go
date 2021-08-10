package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	bd "github.com/SAP/jenkins-library/pkg/blackduck"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/pkg/errors"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/reporting"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/toolrecord"
	"github.com/SAP/jenkins-library/pkg/versioning"
)

type detectUtils interface {
	Abs(path string) (string, error)
	FileExists(filename string) (bool, error)
	FileRemove(filename string) error
	Copy(src, dest string) (int64, error)
	DirExists(dest string) (bool, error)
	FileRead(path string) ([]byte, error)
	FileWrite(path string, content []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	Chmod(path string, mode os.FileMode) error
	Glob(pattern string) (matches []string, err error)

	Stdout(out io.Writer)
	Stderr(err io.Writer)
	SetDir(dir string)
	SetEnv(env []string)
	RunExecutable(e string, p ...string) error
	RunShell(shell, script string) error

	DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error
}

type detectUtilsBundle struct {
	*command.Command
	*piperutils.Files
	*piperhttp.Client
}

func newDetectUtils() detectUtils {
	utils := detectUtilsBundle{
		Command: &command.Command{
			ErrorCategoryMapping: map[string][]string{
				log.ErrorCompliance.String(): {
					"FAILURE_POLICY_VIOLATION - Detect found policy violations.",
				},
				log.ErrorConfiguration.String(): {
					"FAILURE_CONFIGURATION - Detect was unable to start due to issues with it's configuration.",
				},
			},
		},
		Files:  &piperutils.Files{},
		Client: &piperhttp.Client{},
	}
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func detectExecuteScan(config detectExecuteScanOptions, _ *telemetry.CustomData, influx *detectExecuteScanInflux) {
	influx.step_data.fields.detect = false
	utils := newDetectUtils()
	err := runDetect(config, utils, influx)

	if err != nil {
		log.Entry().
			WithError(err).
			Fatal("failed to execute detect scan")
	}

	influx.step_data.fields.detect = true
	// create Toolrecord file
	toolRecordFileName, err := createToolRecordDetect("./", config)
	if err != nil {
		// do not fail until the framework is well established
		log.Entry().Warning("TR_DETECT: Failed to create toolrecord file "+toolRecordFileName, err)
	}
}

func runDetect(config detectExecuteScanOptions, utils detectUtils, influx *detectExecuteScanInflux) error {
	// detect execution details, see https://synopsys.atlassian.net/wiki/spaces/INTDOCS/pages/88440888/Sample+Synopsys+Detect+Scan+Configuration+Scenarios+for+Black+Duck
	err := getDetectScript(config, utils)
	if err != nil {
		return fmt.Errorf("failed to download 'detect.sh' script: %w", err)
	}
	defer func() {
		err := utils.FileRemove("detect.sh")
		if err != nil {
			log.Entry().Warnf("failed to delete 'detect.sh' script: %v", err)
		}
	}()
	err = utils.Chmod("detect.sh", 0700)
	if err != nil {
		return err
	}

	if config.InstallArtifacts {
		err := maven.InstallMavenArtifacts(&maven.EvaluateOptions{
			M2Path:              config.M2Path,
			ProjectSettingsFile: config.ProjectSettingsFile,
			GlobalSettingsFile:  config.GlobalSettingsFile,
		}, utils)
		if err != nil {
			return err
		}
	}

	args := []string{"./detect.sh"}
	args, err = addDetectArgs(args, config, utils)
	if err != nil {
		return err
	}
	script := strings.Join(args, " ")

	envs := []string{"BLACKDUCK_SKIP_PHONE_HOME=true"}
	envs = append(envs, config.CustomEnvironmentVariables...)

	utils.SetDir(".")
	utils.SetEnv(envs)

	err = utils.RunShell("/bin/bash", script)
	postScanChecksAndReporting(config, influx, utils)
	if err == nil && piperutils.ContainsString(config.FailOn, "BLOCKER") {
		violations := struct {
			PolicyViolations int      `json:"policyViolations"`
			Reports          []string `json:"reports"`
		}{
			PolicyViolations: 0,
			Reports:          []string{},
		}

		if files, err := utils.Glob("**/*BlackDuck_RiskReport.pdf"); err == nil && len(files) > 0 {
			// there should only be one RiskReport thus only taking the first one
			_, reportFile := filepath.Split(files[0])
			violations.Reports = append(violations.Reports, reportFile)
		}

		violationContent, err := json.Marshal(violations)
		if err != nil {
			return fmt.Errorf("failed to marshal policy violation data: %w", err)
		}

		err = utils.FileWrite("blackduck-ip.json", violationContent, 0666)
		if err != nil {
			return fmt.Errorf("failed to write policy violation report: %w", err)
		}
	}
	return err
}

func getDetectScript(config detectExecuteScanOptions, utils detectUtils) error {
	if config.ScanOnChanges {
		return utils.DownloadFile("https://raw.githubusercontent.com/blackducksoftware/detect_rescan/master/detect_rescan.sh", "detect.sh", nil, nil)
	}
	return utils.DownloadFile("https://detect.synopsys.com/detect.sh", "detect.sh", nil, nil)
}

func addDetectArgs(args []string, config detectExecuteScanOptions, utils detectUtils) ([]string, error) {
	detectVersionName := getVersionName(config)
	//Split on spaces, the scanPropeties, so that each property is available as a single string
	//instead of all properties being part of a single string
	config.ScanProperties = piperutils.SplitAndTrim(config.ScanProperties, " ")

	if config.ScanOnChanges {
		args = append(args, "--report")
		config.Unmap = false
	}

	if config.Unmap {
		if !piperutils.ContainsString(config.ScanProperties, "--detect.project.codelocation.unmap=true") {
			args = append(args, fmt.Sprintf("--detect.project.codelocation.unmap=true"))
		}
		config.ScanProperties, _ = piperutils.RemoveAll(config.ScanProperties, "--detect.project.codelocation.unmap=false")
	} else {
		//When unmap is set to false, any occurances of unmap=true from scanProperties must be removed
		config.ScanProperties, _ = piperutils.RemoveAll(config.ScanProperties, "--detect.project.codelocation.unmap=true")
	}

	args = append(args, config.ScanProperties...)

	args = append(args, fmt.Sprintf("--blackduck.url=%v", config.ServerURL))
	args = append(args, fmt.Sprintf("--blackduck.api.token=%v", config.Token))
	// ProjectNames, VersionName, GroupName etc can contain spaces and need to be escaped using double quotes in CLI
	// Hence the string need to be surrounded by \"
	args = append(args, fmt.Sprintf("\"--detect.project.name='%v'\"", config.ProjectName))
	args = append(args, fmt.Sprintf("\"--detect.project.version.name='%v'\"", detectVersionName))

	// Groups parameter is added only when there is atleast one non-empty groupname provided
	if len(config.Groups) > 0 && len(config.Groups[0]) > 0 {
		args = append(args, fmt.Sprintf("\"--detect.project.user.groups='%v'\"", strings.Join(config.Groups, ",")))
	}

	// Atleast 1, non-empty category to fail on must be provided
	if len(config.FailOn) > 0 && len(config.FailOn[0]) > 0 {
		args = append(args, fmt.Sprintf("--detect.policy.check.fail.on.severities=%v", strings.Join(config.FailOn, ",")))
	}

	codelocation := config.CodeLocation
	if len(codelocation) == 0 && len(config.ProjectName) > 0 {
		codelocation = fmt.Sprintf("%v/%v", config.ProjectName, detectVersionName)
	}
	args = append(args, fmt.Sprintf("\"--detect.code.location.name='%v'\"", codelocation))

	if len(config.ScanPaths) > 0 && len(config.ScanPaths[0]) > 0 {
		args = append(args, fmt.Sprintf("--detect.blackduck.signature.scanner.paths=%v", strings.Join(config.ScanPaths, ",")))
	}

	if len(config.DependencyPath) > 0 {
		args = append(args, fmt.Sprintf("--detect.source.path=%v", config.DependencyPath))
	} else {
		args = append(args, fmt.Sprintf("--detect.source.path='.'"))
	}

	if len(config.IncludedPackageManagers) > 0 {
		args = append(args, fmt.Sprintf("--detect.included.detector.types=%v", strings.ToUpper(strings.Join(config.IncludedPackageManagers, ","))))
	}

	if len(config.ExcludedPackageManagers) > 0 {
		args = append(args, fmt.Sprintf("--detect.excluded.detector.types=%v", strings.ToUpper(strings.Join(config.ExcludedPackageManagers, ","))))
	}

	if len(config.MavenExcludedScopes) > 0 {
		args = append(args, fmt.Sprintf("--detect.maven.excluded.scopes=%v", strings.ToLower(strings.Join(config.MavenExcludedScopes, ","))))
	}

	if len(config.DetectTools) > 0 {
		args = append(args, fmt.Sprintf("--detect.tools=%v", strings.Join(config.DetectTools, ",")))
	}

	mavenArgs, err := maven.DownloadAndGetMavenParameters(config.GlobalSettingsFile, config.ProjectSettingsFile, utils)
	if err != nil {
		return nil, err
	}

	if len(config.M2Path) > 0 {
		absolutePath, err := utils.Abs(config.M2Path)
		if err != nil {
			return nil, err
		}
		mavenArgs = append(mavenArgs, fmt.Sprintf("-Dmaven.repo.local=%v", absolutePath))
	}

	if len(mavenArgs) > 0 {
		args = append(args, fmt.Sprintf("\"--detect.maven.build.command='%v'\"", strings.Join(mavenArgs, " ")))
	}

	return args, nil
}

func getVersionName(config detectExecuteScanOptions) string {
	detectVersionName := config.CustomScanVersion
	if len(detectVersionName) > 0 {
		log.Entry().Infof("Using custom version: %v", detectVersionName)
	} else {
		detectVersionName = versioning.ApplyVersioningModel(config.VersioningModel, config.Version)
	}
	return detectVersionName
}

func postScanChecksAndReporting(config detectExecuteScanOptions, influx *detectExecuteScanInflux, utils detectUtils) error {
	vulns, _, err := getVulnsAndComponents(config, influx)
	if err != nil {
		return err
	}
	scanReport := createVulnerabilityReport(config, vulns, influx)
	paths, err := writeVulnerabilityReports(scanReport, config, utils)
	piperutils.PersistReportsAndLinks("detectExecuteScan", "", paths, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to check and report scan results")
	}
	return nil
}

func getVulnsAndComponents(config detectExecuteScanOptions, influx *detectExecuteScanInflux) (*bd.Vulnerabilities, *bd.Components, error) {
	detectVersionName := getVersionName(config)
	bdClient := bd.NewClient(config.Token, config.ServerURL, &piperhttp.Client{})
	vulns, err := bdClient.GetVulnerabilities(config.ProjectName, detectVersionName)
	if err != nil {
		return nil, nil, err
	}

	majorVulns := 0
	activeVulns := 0
	for _, vuln := range vulns.Items {
		if isActiveVulnerability(vuln) {
			activeVulns++
			if isMajorVulnerability(vuln) {
				majorVulns++
			}
		}
	}
	influx.detect_data.fields.vulnerabilities = activeVulns
	influx.detect_data.fields.major_vulnerabilities = majorVulns
	influx.detect_data.fields.minor_vulnerabilities = activeVulns - majorVulns

	components, err := bdClient.GetComponents(config.ProjectName, detectVersionName)
	if err != nil {
		return vulns, nil, err
	}
	influx.detect_data.fields.components = components.TotalCount

	return vulns, components, nil
}

func createVulnerabilityReport(config detectExecuteScanOptions, vulns *bd.Vulnerabilities, influx *detectExecuteScanInflux) reporting.ScanReport {
	scanReport := reporting.ScanReport{
		Title: "BlackDuck Security Vulnerability Report",
		Subheaders: []reporting.Subheader{
			{Description: "BlackDuck Project Name ", Details: config.ProjectName},
			{Description: "BlackDuck Project Version ", Details: getVersionName(config)},
		},
		Overview: []reporting.OverviewRow{
			{Description: "Total number of vulnerabilities ", Details: fmt.Sprint(influx.detect_data.fields.vulnerabilities)},
			{Description: "Total number of Critical/High vulnerabilties ", Details: fmt.Sprint(influx.detect_data.fields.major_vulnerabilities)},
		},
		SuccessfulScan: influx.detect_data.fields.major_vulnerabilities == 0,
		ReportTime:     time.Now(),
	}

	detailTable := reporting.ScanDetailTable{
		NoRowsMessage: "No publicly known vulnerabilities detected",
		Headers: []string{
			"Vulnerability Name",
			"Severity",
			"Overall Score",
			"Base Score",
			"Component Name",
			"Component Version",
			"Description",
			"Status",
		},
		WithCounter:   true,
		CounterHeader: "Entry#",
	}

	for _, vuln := range vulns.Items {
		row := reporting.ScanRow{}
		row.AddColumn(vuln.VulnerabilityWithRemediation.VulnerabilityName, 0)
		row.AddColumn(vuln.VulnerabilityWithRemediation.Severity, 0)

		var scoreStyle reporting.ColumnStyle = reporting.Yellow
		if isMajorVulnerability(vuln) {
			scoreStyle = reporting.Red
		}
		row.AddColumn(vuln.VulnerabilityWithRemediation.OverallScore, scoreStyle)
		row.AddColumn(vuln.VulnerabilityWithRemediation.BaseScore, 0)
		row.AddColumn(vuln.Name, 0)
		row.AddColumn(vuln.Version, 0)
		row.AddColumn(vuln.VulnerabilityWithRemediation.Description, 0)
		row.AddColumn(vuln.VulnerabilityWithRemediation.RemediationStatus, 0)

		detailTable.Rows = append(detailTable.Rows, row)
	}

	scanReport.DetailTable = detailTable
	return scanReport
}

func writeVulnerabilityReports(scanReport reporting.ScanReport, config detectExecuteScanOptions, utils detectUtils) ([]piperutils.Path, error) {
	reportPaths := []piperutils.Path{}

	htmlReport, _ := scanReport.ToHTML()
	htmlReportPath := filepath.Join("blackduck", "piper_detect_vulnerability_report.html")
	if err := utils.FileWrite(htmlReportPath, htmlReport, 0666); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return reportPaths, errors.Wrapf(err, "failed to write html report")
	}
	reportPaths = append(reportPaths, piperutils.Path{Name: "BlackDuck Vulnerability Report", Target: htmlReportPath})

	jsonReport, _ := scanReport.ToJSON()
	if exists, _ := utils.DirExists(reporting.StepReportDirectory); !exists {
		err := utils.MkdirAll(reporting.StepReportDirectory, 0777)
		if err != nil {
			return reportPaths, errors.Wrap(err, "failed to create reporting directory")
		}
	}
	if err := utils.FileWrite(filepath.Join(reporting.StepReportDirectory, fmt.Sprintf("detectExecuteScan_oss_%v.json", fmt.Sprintf("%v", time.Now()))), jsonReport, 0666); err != nil {
		return reportPaths, errors.Wrapf(err, "failed to write json report")
	}

	return reportPaths, nil
}

func isActiveVulnerability(v bd.Vulnerability) bool {
	switch v.VulnerabilityWithRemediation.RemediationStatus {
	case "NEW":
		return true
	case "REMEDIATION_REQUIRED":
		return true
	case "NEEDS_REVIEW":
		return true
	default:
		return false
	}
}

func isMajorVulnerability(v bd.Vulnerability) bool {
	switch v.VulnerabilityWithRemediation.Severity {
	case "CRITICAL":
		return true
	case "HIGH":
		return true
	default:
		return false
	}
}

// create toolrecord file for detect
//
//
func createToolRecordDetect(workspace string, config detectExecuteScanOptions) (string, error) {
	record := toolrecord.New(workspace, "detectExecute", config.ServerURL)

	projectId := ""  // todo needs more research; according to synopsis documentation
	productURL := "" // relevant ids can be found in the logfile
	err := record.AddKeyData("project",
		projectId,
		config.ProjectName,
		productURL)
	if err != nil {
		return "", err
	}
	record.AddContext("DetectTools", config.DetectTools)
	err = record.Persist()
	if err != nil {
		return "", err
	}
	return record.GetFileName(), nil
}
