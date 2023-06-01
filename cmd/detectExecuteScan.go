package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	bd "github.com/SAP/jenkins-library/pkg/blackduck"
	"github.com/SAP/jenkins-library/pkg/command"
	piperGithub "github.com/SAP/jenkins-library/pkg/github"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/reporting"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/toolrecord"
	"github.com/SAP/jenkins-library/pkg/versioning"

	"github.com/google/go-github/v45/github"
	"github.com/pkg/errors"
)

type detectUtils interface {
	piperutils.FileUtils

	GetExitCode() int
	GetOsEnv() []string
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	SetDir(dir string)
	SetEnv(env []string)
	RunExecutable(e string, p ...string) error
	RunShell(shell, script string) error

	DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error

	GetIssueService() *github.IssuesService
	GetSearchService() *github.SearchService
	GetProvider() orchestrator.OrchestratorSpecificConfigProviding
}

type detectUtilsBundle struct {
	*command.Command
	*piperutils.Files
	*piperhttp.Client
	issues   *github.IssuesService
	search   *github.SearchService
	provider orchestrator.OrchestratorSpecificConfigProviding
}

func (d *detectUtilsBundle) GetIssueService() *github.IssuesService {
	return d.issues
}

func (d *detectUtilsBundle) GetSearchService() *github.SearchService {
	return d.search
}

func (d *detectUtilsBundle) GetProvider() orchestrator.OrchestratorSpecificConfigProviding {
	return d.provider
}

type blackduckSystem struct {
	Client bd.Client
}

func newDetectUtils(client *github.Client) detectUtils {
	utils := detectUtilsBundle{
		Command: &command.Command{
			ErrorCategoryMapping: map[string][]string{
				log.ErrorCompliance.String(): {
					"FAILURE_POLICY_VIOLATION - Detect found policy violations.",
				},
				log.ErrorConfiguration.String(): {
					"FAILURE_CONFIGURATION - Detect was unable to start due to issues with it's configuration.",
					"FAILURE_DETECTOR - Detect had one or more detector failures while extracting dependencies. Check that all projects build and your environment is configured correctly.",
					"FAILURE_SCAN - Detect was unable to run the signature scanner against your source. Check your configuration.",
				},
				log.ErrorInfrastructure.String(): {
					"FAILURE_PROXY_CONNECTIVITY - Detect was unable to use the configured proxy. Check your configuration and connection.",
					"FAILURE_BLACKDUCK_CONNECTIVITY - Detect was unable to connect to Black Duck. Check your configuration and connection.",
					"FAILURE_POLARIS_CONNECTIVITY - Detect was unable to connect to Polaris. Check your configuration and connection.",
				},
				log.ErrorService.String(): {
					"FAILURE_TIMEOUT - Detect could not wait for actions to be completed on Black Duck. Check your Black Duck server or increase your timeout.",
					"FAILURE_DETECTOR_REQUIRED - Detect did not run all of the required detectors. Fix detector issues or disable required detectors.",
					"FAILURE_BLACKDUCK_VERSION_NOT_SUPPORTED - Detect attempted an operation that was not supported by your version of Black Duck. Ensure your Black Duck is compatible with this version of detect.",
					"FAILURE_BLACKDUCK_FEATURE_ERROR - Detect encountered an error while attempting an operation on Black Duck. Ensure your Black Duck is compatible with this version of detect.",
					"FAILURE_GENERAL_ERROR - Detect encountered a known error, details of the error are provided.",
					"FAILURE_UNKNOWN_ERROR - Detect encountered an unknown error.",
				},
			},
		},
		Files:  &piperutils.Files{},
		Client: &piperhttp.Client{},
	}
	if client != nil {
		utils.issues = client.Issues
		utils.search = client.Search
	}
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())

	provider, err := orchestrator.NewOrchestratorSpecificConfigProvider()
	if err != nil {
		log.Entry().WithError(err).Warning(err)
		provider = &orchestrator.UnknownOrchestratorConfigProvider{}
	}

	utils.provider = provider

	return &utils
}

func newBlackduckSystem(config detectExecuteScanOptions) *blackduckSystem {
	sys := blackduckSystem{
		Client: bd.NewClient(config.Token, config.ServerURL, &piperhttp.Client{}),
	}
	return &sys
}

func detectExecuteScan(config detectExecuteScanOptions, _ *telemetry.CustomData, influx *detectExecuteScanInflux) {
	influx.step_data.fields.detect = false

	ctx, client, err := piperGithub.NewClient(config.GithubToken, config.GithubAPIURL, "", config.CustomTLSCertificateLinks)
	if err != nil {
		log.Entry().WithError(err).Warning("Failed to get GitHub client")
	}
	utils := newDetectUtils(client)
	if err := runDetect(ctx, config, utils, influx); err != nil {
		log.Entry().
			WithError(err).
			Fatal("failed to execute detect scan")
	}

	influx.step_data.fields.detect = true
}

func runDetect(ctx context.Context, config detectExecuteScanOptions, utils detectUtils, influx *detectExecuteScanInflux) error {
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
	err = utils.Chmod("detect.sh", 0o700)
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

	blackduckSystem := newBlackduckSystem(config)

	args := []string{"./detect.sh"}
	args, err = addDetectArgs(args, config, utils, blackduckSystem)
	if err != nil {
		return err
	}
	script := strings.Join(args, " ")

	envs := []string{"BLACKDUCK_SKIP_PHONE_HOME=true"}
	envs = append(envs, config.CustomEnvironmentVariables...)

	utils.SetDir(".")
	utils.SetEnv(envs)

	err = utils.RunShell("/bin/bash", script)
	reportingErr := postScanChecksAndReporting(ctx, config, influx, utils, blackduckSystem)
	if reportingErr != nil {
		if strings.Contains(reportingErr.Error(), "License Policy Violations found") {
			log.Entry().Errorf("License Policy Violations found")
			log.SetErrorCategory(log.ErrorCompliance)
			if err == nil && !piperutils.ContainsStringPart(config.FailOn, "NONE") {
				err = errors.New("License Policy Violations found")
			}
		} else {
			log.Entry().Warnf("Failed to generate reports: %v", reportingErr)
		}
	}
	if err != nil {
		// Setting error category based on exit code
		mapErrorCategory(utils.GetExitCode())
		if log.GetErrorCategory() == log.ErrorCompliance && !config.FailOnSevereVulnerabilities {
			err = nil
			log.Entry().Infof("policy violation(s) found - step will only create data but not fail due to setting failOnSevereVulnerabilities: false")
		} else {
			// Error code mapping with more human readable text
			err = errors.Wrapf(err, exitCodeMapping(utils.GetExitCode()))
		}
	}
	// create Toolrecord file
	toolRecordFileName, toolRecordErr := createToolRecordDetect(utils, "./", config, blackduckSystem)
	if toolRecordErr != nil {
		// do not fail until the framework is well established
		log.Entry().Warning("TR_DETECT: Failed to create toolrecord file "+toolRecordFileName, err)
	}
	return err
}

// Get proper error category
func mapErrorCategory(exitCodeKey int) {
	switch exitCodeKey {
	case 0:
		// In case detect exits successfully, we rely on the function 'postScanChecksAndReporting' to determine the error category
		// hence this method doesnt need to set an error category or go to 'default' case
		break
	case 1:
		log.SetErrorCategory(log.ErrorInfrastructure)
	case 2:
		log.SetErrorCategory(log.ErrorService)
	case 3:
		log.SetErrorCategory(log.ErrorCompliance)
	case 4:
		log.SetErrorCategory(log.ErrorInfrastructure)
	case 5:
		log.SetErrorCategory(log.ErrorConfiguration)
	case 6:
		log.SetErrorCategory(log.ErrorConfiguration)
	case 7:
		log.SetErrorCategory(log.ErrorConfiguration)
	case 9:
		log.SetErrorCategory(log.ErrorService)
	case 10:
		log.SetErrorCategory(log.ErrorService)
	case 11:
		log.SetErrorCategory(log.ErrorService)
	case 12:
		log.SetErrorCategory(log.ErrorInfrastructure)
	case 99:
		log.SetErrorCategory(log.ErrorService)
	case 100:
		log.SetErrorCategory(log.ErrorUndefined)
	default:
		log.SetErrorCategory(log.ErrorUndefined)
	}
}

// Exit codes/error code mapping
func exitCodeMapping(exitCodeKey int) string {
	exitCodes := map[int]string{
		0:   "Detect Scan completed successfully",
		1:   "FAILURE_BLACKDUCK_CONNECTIVITY => Detect was unable to connect to Black Duck. Check your configuration and connection.",
		2:   "FAILURE_TIMEOUT => Detect could not wait for actions to be completed on Black Duck. Check your Black Duck server or increase your timeout.",
		3:   "FAILURE_POLICY_VIOLATION => Detect found policy violations.",
		4:   "FAILURE_PROXY_CONNECTIVITY => Detect was unable to use the configured proxy. Check your configuration and connection.",
		5:   "FAILURE_DETECTOR => Detect had one or more detector failures while extracting dependencies. Check that all projects build and your environment is configured correctly.",
		6:   "FAILURE_SCAN => Detect was unable to run the signature scanner against your source. Check your configuration.",
		7:   "FAILURE_CONFIGURATION => Detect was unable to start because of a configuration issue. Check and fix your configuration.",
		9:   "FAILURE_DETECTOR_REQUIRED => Detect did not run all of the required detectors. Fix detector issues or disable required detectors.",
		10:  "FAILURE_BLACKDUCK_VERSION_NOT_SUPPORTED => Detect attempted an operation that was not supported by your version of Black Duck. Ensure your Black Duck is compatible with this version of detect.",
		11:  "FAILURE_BLACKDUCK_FEATURE_ERROR => Detect encountered an error while attempting an operation on Black Duck. Ensure your Black Duck is compatible with this version of detect.",
		12:  "FAILURE_POLARIS_CONNECTIVITY => Detect was unable to connect to Polaris. Check your configuration and connection.",
		99:  "FAILURE_GENERAL_ERROR => Detect encountered a known error, details of the error are provided.",
		100: "FAILURE_UNKNOWN_ERROR => Detect encountered an unknown error.",
	}

	if _, isKeyExists := exitCodes[exitCodeKey]; isKeyExists {
		return exitCodes[exitCodeKey]
	}

	return "[" + strconv.Itoa(exitCodeKey) + "]: Not known exit code key"
}

func getDetectScript(config detectExecuteScanOptions, utils detectUtils) error {
	if config.ScanOnChanges {
		log.Entry().Infof("Using Detect Rescan script")
		return utils.DownloadFile("https://raw.githubusercontent.com/blackducksoftware/detect_rescan/master/detect_rescan.sh", "detect.sh", nil, nil)
	}

	log.Entry().Infof("Downloading Detect7")
	return utils.DownloadFile("https://detect.synopsys.com/detect7.sh", "detect.sh", nil, nil)
}

func addDetectArgs(args []string, config detectExecuteScanOptions, utils detectUtils, sys *blackduckSystem) ([]string, error) {
	detectVersionName := getVersionName(config)
	// Split on spaces, the scanPropeties, so that each property is available as a single string
	// instead of all properties being part of a single string
	config.ScanProperties = piperutils.SplitAndTrim(config.ScanProperties, " ")

	if config.BuildTool == "mta" {

		if !checkIfArgumentIsInScanProperties(config, "detect.detector.search.depth") {
			args = append(args, "--detect.detector.search.depth=100")
		}

		if !checkIfArgumentIsInScanProperties(config, "detect.detector.search.continue") {
			args = append(args, "--detect.detector.search.continue=true")
		}

	}

	if len(config.ExcludedDirectories) != 0 && !checkIfArgumentIsInScanProperties(config, "detect.excluded.directories") {
		args = append(args, fmt.Sprintf("--detect.excluded.directories=%s", strings.Join(config.ExcludedDirectories, ",")))
	}

	if config.ScanOnChanges {
		args = append(args, "--report")
		config.Unmap = false
	}

	if config.MinScanInterval > 0 {
		//Unmap doesnt work well with min-scan-interval and should be removed
		config.Unmap = false
		args = append(args, fmt.Sprintf("--detect.blackduck.signature.scanner.arguments='--min-scan-interval=%d'", config.MinScanInterval))
	}

	if config.Unmap {
		if !piperutils.ContainsString(config.ScanProperties, "--detect.project.codelocation.unmap=true") {
			args = append(args, "--detect.project.codelocation.unmap=true")
		}
		config.ScanProperties, _ = piperutils.RemoveAll(config.ScanProperties, "--detect.project.codelocation.unmap=false")
	} else {
		// When unmap is set to false, any occurances of unmap=true from scanProperties must be removed
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
		args = append(args, "--detect.source.path='.'")
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

	// to exclude dependency types for npm
	if len(config.NpmDependencyTypesExcluded) > 0 && !checkIfArgumentIsInScanProperties(config, "detect.npm.dependency.types.excluded") {
		args = append(args, fmt.Sprintf("--detect.npm.dependency.types.excluded=%v", strings.ToUpper(strings.Join(config.NpmDependencyTypesExcluded, ","))))
	}

	// A space-separated list of additional arguments that Detect will add at then end of the npm ls command line
	if len(config.NpmArguments) > 0 && !checkIfArgumentIsInScanProperties(config, "detect.npm.arguments") {
		args = append(args, fmt.Sprintf("--detect.npm.arguments=%v", strings.ToUpper(strings.Join(config.NpmArguments, " "))))
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

	// rapid scan on pull request
	if utils.GetProvider().IsPullRequest() {
		log.Entry().Debug("pull request detected")
		args = append(args, "--detect.blackduck.scan.mode='RAPID'")
		_, err := sys.Client.GetProjectVersion(config.ProjectName, config.Version)
		if err == nil {
			args = append(args, "--detect.blackduck.rapid.compare.mode='BOM_COMPARE_STRICT'")
		}
		args = append(args, "--detect.cleanup=false")
		args = append(args, "--detect.output.path='report'")
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

func checkIfArgumentIsInScanProperties(config detectExecuteScanOptions, argumentName string) bool {
	for _, argument := range config.ScanProperties {
		if strings.Contains(argument, argumentName) {
			return true
		}
	}

	return false
}

func createVulnerabilityReport(config detectExecuteScanOptions, vulns *bd.Vulnerabilities, influx *detectExecuteScanInflux, sys *blackduckSystem) reporting.ScanReport {
	versionName := getVersionName(config)
	versionUrl, _ := sys.Client.GetProjectVersionLink(config.ProjectName, versionName)
	scanReport := reporting.ScanReport{
		ReportTitle: "BlackDuck Security Vulnerability Report",
		Subheaders: []reporting.Subheader{
			{Description: "BlackDuck Project Name ", Details: config.ProjectName},
			{Description: "BlackDuck Project Version ", Details: fmt.Sprintf("<a href='%v'>%v</a>", versionUrl, versionName)},
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

	vulnItems := vulns.Items
	sort.Slice(vulnItems, func(i, j int) bool {
		return vulnItems[i].OverallScore > vulnItems[j].OverallScore
	})

	for _, vuln := range vulnItems {
		row := reporting.ScanRow{}
		row.AddColumn(vuln.VulnerabilityWithRemediation.VulnerabilityName, 0)
		row.AddColumn(vuln.VulnerabilityWithRemediation.Severity, 0)

		var scoreStyle reporting.ColumnStyle = reporting.Yellow
		if isMajorVulnerability(vuln) {
			scoreStyle = reporting.Red
		}
		if !isActiveVulnerability(vuln) {
			scoreStyle = reporting.Grey
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

func isActiveVulnerability(v bd.Vulnerability) bool {
	if v.Ignored {
		return false
	}
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
	if v.Ignored {
		return false
	}
	switch v.VulnerabilityWithRemediation.Severity {
	case "CRITICAL":
		return true
	case "HIGH":
		return true
	default:
		return false
	}
}

func postScanChecksAndReporting(ctx context.Context, config detectExecuteScanOptions, influx *detectExecuteScanInflux, utils detectUtils, sys *blackduckSystem) error {

	if utils.GetProvider().IsPullRequest() {
		issueNumber, err := strconv.Atoi(utils.GetProvider().GetPullRequestConfig().Key)
		if err != nil {
			log.Entry().Warning("Can not get issue number ", err)
			return nil
		}
		commentBody, err := reporting.RapidScanResult("./report")
		if err != nil {
			log.Entry().Warning("Couldn't read file of report of rapid scan, error: ", err)
			return nil
		}
		_, _, err = utils.GetIssueService().CreateComment(ctx,
			config.Owner,
			config.Repository,
			issueNumber,
			&github.IssueComment{
				Body: &commentBody,
			})
		if err != nil {
			log.Entry().Warning("Can send request to github ", err)
			return nil
		}

		return nil
	}

	errorsOccured := []string{}
	vulns, err := getVulnerabilitiesWithComponents(config, influx, sys)
	if err != nil {
		return errors.Wrap(err, "failed to fetch vulnerabilities")
	}

	if config.CreateResultIssue && len(config.GithubToken) > 0 && len(config.GithubAPIURL) > 0 && len(config.Owner) > 0 && len(config.Repository) > 0 {
		log.Entry().Debugf("Creating result issues for %v alert(s)", len(vulns.Items))
		issueDetails := make([]reporting.IssueDetail, len(vulns.Items))
		piperutils.CopyAtoB(vulns.Items, issueDetails)
		gh := reporting.GitHub{
			Owner:         &config.Owner,
			Repository:    &config.Repository,
			Assignees:     &config.Assignees,
			IssueService:  utils.GetIssueService(),
			SearchService: utils.GetSearchService(),
		}
		if err := gh.UploadMultipleReports(ctx, &issueDetails); err != nil {
			errorsOccured = append(errorsOccured, fmt.Sprint(err))
		}
	}

	projectVersion, err := sys.Client.GetProjectVersion(config.ProjectName, config.Version)

	var projectLink string
	if projectVersion != nil {
		projectLink = projectVersion.Href
	}

	sarif := bd.CreateSarifResultFile(vulns, config.ProjectName, config.Version, projectLink)
	paths, err := bd.WriteSarifFile(sarif, utils)
	if err != nil {
		errorsOccured = append(errorsOccured, fmt.Sprint(err))
	}

	scanReport := createVulnerabilityReport(config, vulns, influx, sys)
	vulnerabilityReportPaths, err := bd.WriteVulnerabilityReports(scanReport, utils)
	if err != nil {
		errorsOccured = append(errorsOccured, fmt.Sprint(err))
	}
	paths = append(paths, vulnerabilityReportPaths...)

	policyStatus, err := getPolicyStatus(config, influx, sys)
	if err != nil {
		errorsOccured = append(errorsOccured, fmt.Sprint(err))
	}
	policyReport := createPolicyStatusReport(config, policyStatus, influx, sys)
	policyReportPaths, err := writePolicyStatusReports(policyReport, config, utils)
	if err != nil {
		errorsOccured = append(errorsOccured, fmt.Sprint(err))
	}
	paths = append(paths, policyReportPaths...)

	piperutils.PersistReportsAndLinks("detectExecuteScan", "", utils, paths, nil)
	if err != nil {
		errorsOccured = append(errorsOccured, fmt.Sprint(err))
	}

	err, violationCount := writeIpPolicyJson(config, utils, paths, sys)
	if err != nil {
		errorsOccured = append(errorsOccured, fmt.Sprint(err))
	}

	if violationCount > 0 {
		log.SetErrorCategory(log.ErrorCompliance)
		errorsOccured = append(errorsOccured, "License Policy Violations found")
	}

	if len(errorsOccured) > 0 {
		return fmt.Errorf(strings.Join(errorsOccured, ": "))
	}

	return nil
}

func getVulnerabilitiesWithComponents(config detectExecuteScanOptions, influx *detectExecuteScanInflux, sys *blackduckSystem) (*bd.Vulnerabilities, error) {
	detectVersionName := getVersionName(config)
	components, err := sys.Client.GetComponents(config.ProjectName, detectVersionName)
	if err != nil {
		return nil, err
	}
	// create component lookup map to interconnect vulnerability and component
	keyFormat := "%v/%v"
	componentLookup := map[string]*bd.Component{}
	for i := 0; i < len(components.Items); i++ {
		componentLookup[fmt.Sprintf(keyFormat, components.Items[i].Name, components.Items[i].Version)] = &components.Items[i]
	}

	vulns, err := sys.Client.GetVulnerabilities(config.ProjectName, detectVersionName)
	if err != nil {
		return nil, err
	}

	majorVulns := 0
	activeVulns := 0
	for index, vuln := range vulns.Items {
		if isActiveVulnerability(vuln) {
			activeVulns++
			if isMajorVulnerability(vuln) {
				majorVulns++
			}
		}
		component := componentLookup[fmt.Sprintf(keyFormat, vuln.Name, vuln.Version)]
		if component != nil && len(component.Name) > 0 {
			vulns.Items[index].Component = component
		} else {
			vulns.Items[index].Component = &bd.Component{Name: vuln.Name, Version: vuln.Version}
		}
	}
	influx.detect_data.fields.vulnerabilities = activeVulns
	influx.detect_data.fields.major_vulnerabilities = majorVulns
	influx.detect_data.fields.minor_vulnerabilities = activeVulns - majorVulns
	influx.detect_data.fields.components = components.TotalCount

	return vulns, nil
}

func getPolicyStatus(config detectExecuteScanOptions, influx *detectExecuteScanInflux, sys *blackduckSystem) (*bd.PolicyStatus, error) {
	policyStatus, err := sys.Client.GetPolicyStatus(config.ProjectName, getVersionName(config))
	if err != nil {
		return nil, err
	}

	totalViolations := 0
	for _, level := range policyStatus.SeverityLevels {
		totalViolations += level.Value
	}
	influx.detect_data.fields.policy_violations = totalViolations

	return policyStatus, nil
}

func createPolicyStatusReport(config detectExecuteScanOptions, policyStatus *bd.PolicyStatus, influx *detectExecuteScanInflux, sys *blackduckSystem) reporting.ScanReport {
	versionName := getVersionName(config)
	versionUrl, _ := sys.Client.GetProjectVersionLink(config.ProjectName, versionName)
	policyReport := reporting.ScanReport{
		ReportTitle: "BlackDuck Policy Violations Report",
		Subheaders: []reporting.Subheader{
			{Description: "BlackDuck project name ", Details: config.ProjectName},
			{Description: "BlackDuck project version name", Details: fmt.Sprintf("<a href='%v'>%v</a>", versionUrl, versionName)},
		},
		Overview: []reporting.OverviewRow{
			{Description: "Overall Policy Violation Status", Details: policyStatus.OverallStatus},
			{Description: "Total Number of Policy Vioaltions", Details: fmt.Sprint(influx.detect_data.fields.policy_violations)},
		},
		SuccessfulScan: influx.detect_data.fields.policy_violations > 0,
		ReportTime:     time.Now(),
	}

	detailTable := reporting.ScanDetailTable{
		Headers: []string{
			"Policy Severity Level", "Number of Components in Violation",
		},
		WithCounter: false,
	}

	for _, level := range policyStatus.SeverityLevels {
		row := reporting.ScanRow{}
		row.AddColumn(level.Name, 0)
		row.AddColumn(level.Value, 0)
		detailTable.Rows = append(detailTable.Rows, row)
	}
	policyReport.DetailTable = detailTable

	return policyReport
}

func writePolicyStatusReports(scanReport reporting.ScanReport, config detectExecuteScanOptions, utils detectUtils) ([]piperutils.Path, error) {
	reportPaths := []piperutils.Path{}

	htmlReport, _ := scanReport.ToHTML()
	htmlReportPath := "piper_detect_policy_violation_report.html"
	if err := utils.FileWrite(htmlReportPath, htmlReport, 0o666); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return reportPaths, errors.Wrapf(err, "failed to write html report")
	}
	reportPaths = append(reportPaths, piperutils.Path{Name: "BlackDuck Policy Violation Report", Target: htmlReportPath})

	jsonReport, _ := scanReport.ToJSON()
	if exists, _ := utils.DirExists(reporting.StepReportDirectory); !exists {
		err := utils.MkdirAll(reporting.StepReportDirectory, 0o777)
		if err != nil {
			return reportPaths, errors.Wrap(err, "failed to create reporting directory")
		}
	}
	if err := utils.FileWrite(filepath.Join(reporting.StepReportDirectory, fmt.Sprintf("detectExecuteScan_policy_%v.json", fmt.Sprintf("%v", time.Now()))), jsonReport, 0o666); err != nil {
		return reportPaths, errors.Wrapf(err, "failed to write json report")
	}

	return reportPaths, nil
}

func writeIpPolicyJson(config detectExecuteScanOptions, utils detectUtils, paths []piperutils.Path, sys *blackduckSystem) (error, int) {
	components, err := sys.Client.GetComponentsWithLicensePolicyRule(config.ProjectName, getVersionName(config))
	if err != nil {
		return errors.Wrapf(err, "failed to get License Policy Violations"), 0
	}

	violationCount := getActivePolicyViolations(components)
	violations := struct {
		PolicyViolations int      `json:"policyViolations"`
		Reports          []string `json:"reports"`
	}{
		PolicyViolations: violationCount,
		Reports:          []string{},
	}

	for _, path := range paths {
		violations.Reports = append(violations.Reports, path.Target)
	}
	if files, err := utils.Glob("**/*BlackDuck_RiskReport.pdf"); err == nil && len(files) > 0 {
		// there should only be one RiskReport thus only taking the first one
		_, reportFile := filepath.Split(files[0])
		violations.Reports = append(violations.Reports, reportFile)
	}

	violationContent, err := json.Marshal(violations)
	if err != nil {
		return fmt.Errorf("failed to marshal policy violation data: %w", err), violationCount
	}

	err = utils.FileWrite("blackduck-ip.json", violationContent, 0o666)
	if err != nil {
		return fmt.Errorf("failed to write policy violation report: %w", err), violationCount
	}
	return nil, violationCount
}

func getActivePolicyViolations(components *bd.Components) int {
	if components.TotalCount == 0 {
		return 0
	}
	activeViolations := 0
	for _, component := range components.Items {
		if isActivePolicyViolation(component.PolicyStatus) {
			activeViolations++
		}
	}
	return activeViolations
}

func isActivePolicyViolation(status string) bool {
	return status == "IN_VIOLATION"
}

// create toolrecord file for detectExecute
func createToolRecordDetect(utils detectUtils, workspace string, config detectExecuteScanOptions, sys *blackduckSystem) (string, error) {
	record := toolrecord.New(utils, workspace, "detectExecute", config.ServerURL)
	project, err := sys.Client.GetProject(config.ProjectName)
	if err != nil {
		return "", fmt.Errorf("TR_DETECT: GetProject failed %v", err)
	}
	metadata := project.Metadata
	projectURL := metadata.Href
	if projectURL == "" {
		return "", fmt.Errorf("TR_DETECT: no project URL")
	}
	// project UUID comes as last part of the URL
	parts := strings.Split(projectURL, "/")
	projectId := parts[len(parts)-1]
	if projectId == "" {
		return "", fmt.Errorf("TR_DETECT: no project id in %v", projectURL)
	}
	err = record.AddKeyData("project",
		projectId,
		config.ProjectName,
		projectURL)
	if err != nil {
		return "", err
	}
	projectVersionName := getVersionName(config)
	projectVersion, err := sys.Client.GetProjectVersion(config.ProjectName, projectVersionName)
	if err != nil {
		return "", err
	}
	projectVersionUrl := projectVersion.Href
	if projectVersionUrl == "" {
		return "", fmt.Errorf("TR_DETECT: no projectversion URL")
	}
	// projectVersion UUID comes as last part of the URL
	vparts := strings.Split(projectVersionUrl, "/")
	projectVersionId := vparts[len(vparts)-1]
	if projectVersionId == "" {
		return "", fmt.Errorf("TR_DETECT: no projectversion id in %v", projectVersionUrl)
	}

	err = record.AddKeyData("version",
		projectVersionId,
		projectVersion.Name,
		projectVersion.Href)
	if err != nil {
		return "", err
	}
	_ = record.AddContext("DetectTools", config.DetectTools)
	err = record.Persist()
	if err != nil {
		return "", err
	}
	return record.GetFileName(), nil
}
