package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar"

	"github.com/google/go-github/v32/github"
	"github.com/google/uuid"

	"github.com/piper-validation/fortify-client-go/models"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/fortify"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/versioning"

	piperGithub "github.com/SAP/jenkins-library/pkg/github"
)

type pullRequestService interface {
	ListPullRequestsWithCommit(ctx context.Context, owner, repo, sha string, opts *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error)
}

type fortifyExecRunner interface {
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	SetDir(d string)
	RunExecutable(e string, p ...string) error
}

const checkString = "<---CHECK FORTIFY---"
const classpathFileName = "fortify-execute-scan-cp.txt"

func fortifyExecuteScan(config fortifyExecuteScanOptions, telemetryData *telemetry.CustomData, influx *fortifyExecuteScanInflux) {
	auditStatus := map[string]string{}
	sys := fortify.NewSystemInstance(config.ServerURL, config.APIEndpoint, config.AuthToken, time.Minute*15)
	c := &command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())

	artifact, err := determineArtifact(config, c)
	if err != nil {
		log.Entry().WithError(err).Fatal()
	}

	reports, err := runFortifyScan(config, sys, c, artifact, telemetryData, influx, auditStatus)
	piperutils.PersistReportsAndLinks("fortifyExecuteScan", config.ModulePath, reports, nil)
	if err != nil {
		log.Entry().WithError(err).Fatal("Fortify scan and check failed")
	}
}

func determineArtifact(config fortifyExecuteScanOptions, c *command.Command) (versioning.Artifact, error) {
	versioningOptions := versioning.Options{
		M2Path:              config.M2Path,
		GlobalSettingsFile:  config.GlobalSettingsFile,
		ProjectSettingsFile: config.ProjectSettingsFile,
	}

	artifact, err := versioning.GetArtifact(config.BuildTool, config.BuildDescriptorFile, &versioningOptions, c)
	if err != nil {
		return nil, fmt.Errorf("Unable to get artifact from descriptor %v: %w", config.BuildDescriptorFile, err)
	}
	return artifact, nil
}

func runFortifyScan(config fortifyExecuteScanOptions, sys fortify.System, command fortifyExecRunner, artifact versioning.Artifact, telemetryData *telemetry.CustomData, influx *fortifyExecuteScanInflux, auditStatus map[string]string) ([]piperutils.Path, error) {
	var reports []piperutils.Path
	log.Entry().Debugf("Running Fortify scan against SSC at %v", config.ServerURL)
	coordinates, err := artifact.GetCoordinates()
	if err != nil {
		return reports, fmt.Errorf("unable to get project coordinates from descriptor %v: %w", config.BuildDescriptorFile, err)
	}
	log.Entry().Debugf("determined project coordinates %v", coordinates)
	fortifyProjectName, fortifyProjectVersion := versioning.DetermineProjectCoordinates(config.ProjectName, config.VersioningModel, coordinates)
	project, err := sys.GetProjectByName(fortifyProjectName, config.AutoCreate, fortifyProjectVersion)
	if err != nil {
		return reports, fmt.Errorf("Failed to load project %v: %w", fortifyProjectName, err)
	}
	projectVersion, err := sys.GetProjectVersionDetailsByProjectIDAndVersionName(project.ID, fortifyProjectVersion, config.AutoCreate, fortifyProjectName)
	if err != nil {
		return reports, fmt.Errorf("Failed to load project version %v: %w", fortifyProjectVersion, err)
	}

	if len(config.PullRequestName) > 0 {
		fortifyProjectVersion = config.PullRequestName
		projectVersion, err := sys.LookupOrCreateProjectVersionDetailsForPullRequest(project.ID, projectVersion, fortifyProjectVersion)
		if err != nil {
			return reports, fmt.Errorf("Failed to lookup / create project version for pull request %v: %w", fortifyProjectVersion, err)
		}
		log.Entry().Debugf("Looked up / created project version with ID %v for PR %v", projectVersion.ID, fortifyProjectVersion)
	} else {
		prID := determinePullRequestMerge(config)
		if len(prID) > 0 {
			log.Entry().Debugf("Determined PR ID '%v' for merge check", prID)
			pullRequestProjectName := fmt.Sprintf("PR-%v", prID)
			err = sys.MergeProjectVersionStateOfPRIntoMaster(config.FprDownloadEndpoint, config.FprUploadEndpoint, project.ID, projectVersion.ID, pullRequestProjectName)
			if err != nil {
				return reports, fmt.Errorf("Failed to merge project version state for pull request %v into project version %v of project %v: %w", pullRequestProjectName, fortifyProjectVersion, project.ID, err)
			}
		}
	}

	filterSet, err := sys.GetFilterSetOfProjectVersionByTitle(projectVersion.ID, config.FilterSetTitle)
	if filterSet == nil || err != nil {
		return reports, fmt.Errorf("Failed to load filter set with title %v", config.FilterSetTitle)
	}

	if config.VerifyOnly {
		log.Entry().Infof("Starting audit status check on project %v with version %v and project version ID %v", fortifyProjectName, fortifyProjectVersion, projectVersion.ID)
		return reports, verifyFFProjectCompliance(config, sys, project, projectVersion, filterSet, influx, auditStatus)
	}

	log.Entry().Infof("Scanning and uploading to project %v with version %v and projectVersionId %v", fortifyProjectName, fortifyProjectVersion, projectVersion.ID)
	buildLabel := fmt.Sprintf("%v/repos/%v/%v/commits/%v", config.GithubAPIURL, config.Owner, config.Repository, config.CommitID)

	// Create sourceanalyzer command based on configuration
	buildID := uuid.New().String()
	command.SetDir(config.ModulePath)
	os.MkdirAll(fmt.Sprintf("%v/%v", config.ModulePath, "target"), os.ModePerm)

	if config.UpdateRulePack {
		err := command.RunExecutable("fortifyupdate", "-acceptKey", "-acceptSSLCertificate", "-url", config.ServerURL)
		if err != nil {
			log.Entry().WithError(err).WithField("serverUrl", config.ServerURL).Fatal("Failed to update rule pack")
		}
		err = command.RunExecutable("fortifyupdate", "-acceptKey", "-acceptSSLCertificate", "-showInstalledRules")
		if err != nil {
			log.Entry().WithError(err).WithField("serverUrl", config.ServerURL).Fatal("Failed to fetch details of installed rule pack")
		}
	}

	triggerFortifyScan(config, command, buildID, buildLabel, fortifyProjectName)

	reports = append(reports, piperutils.Path{Target: fmt.Sprintf("%vtarget/fortify-scan.*", config.ModulePath)})
	reports = append(reports, piperutils.Path{Target: fmt.Sprintf("%vtarget/*.fpr", config.ModulePath)})

	var message string
	if config.UploadResults {
		log.Entry().Debug("Uploading results")
		resultFilePath := fmt.Sprintf("%vtarget/result.fpr", config.ModulePath)
		err = sys.UploadResultFile(config.FprUploadEndpoint, resultFilePath, projectVersion.ID)
		message = fmt.Sprintf("Failed to upload result file %v to Fortify SSC at %v", resultFilePath, config.ServerURL)
	} else {
		log.Entry().Debug("Generating XML report")
		xmlReportName := "fortify_result.xml"
		err = command.RunExecutable("ReportGenerator", "-format", "xml", "-f", xmlReportName, "-source", fmt.Sprintf("%vtarget/result.fpr", config.ModulePath))
		message = fmt.Sprintf("Failed to generate XML report %v", xmlReportName)
		if err != nil {
			reports = append(reports, piperutils.Path{Target: fmt.Sprintf("%vfortify_result.xml", config.ModulePath)})
		}
	}
	if err != nil {
		return reports, fmt.Errorf(message+": %w", err)
	}

	log.Entry().Infof("Starting audit status check on project %v with version %v and project version ID %v", fortifyProjectName, fortifyProjectVersion, projectVersion.ID)
	// Ensure latest FPR is processed
	err = verifyScanResultsFinishedUploading(config, sys, projectVersion.ID, buildLabel, filterSet,
		10*time.Second, time.Duration(config.PollingMinutes)*time.Minute)
	if err != nil {
		return reports, err
	}

	return reports, verifyFFProjectCompliance(config, sys, project, projectVersion, filterSet, influx, auditStatus)
}

func verifyFFProjectCompliance(config fortifyExecuteScanOptions, sys fortify.System, project *models.Project, projectVersion *models.ProjectVersion, filterSet *models.FilterSet, influx *fortifyExecuteScanInflux, auditStatus map[string]string) error {
	// Generate report
	if config.Reporting {
		resultURL := []byte(fmt.Sprintf("https://fortify.tools.sap/ssc/html/ssc/version/%v/fix/null/", projectVersion.ID))
		ioutil.WriteFile(fmt.Sprintf("%vtarget/%v-%v.%v", config.ModulePath, *project.Name, *projectVersion.Name, "txt"), resultURL, 0700)

		data, err := generateAndDownloadQGateReport(config, sys, project, projectVersion)
		if err != nil {
			return err
		}
		ioutil.WriteFile(fmt.Sprintf("%vtarget/%v-%v.%v", config.ModulePath, *project.Name, *projectVersion.Name, config.ReportType), data, 0700)
	}

	// Perform audit compliance checks
	issueFilterSelectorSet, err := sys.GetIssueFilterSelectorOfProjectVersionByName(projectVersion.ID, []string{"Analysis", "Folder", "Category"}, nil)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Failed to fetch project version issue filter selector for project version ID %v", projectVersion.ID)
	}
	log.Entry().Debugf("initial filter selector set: %v", issueFilterSelectorSet)
	numberOfViolations := analyseUnauditedIssues(config, sys, projectVersion, filterSet, issueFilterSelectorSet, influx, auditStatus)
	numberOfViolations += analyseSuspiciousExploitable(config, sys, projectVersion, filterSet, issueFilterSelectorSet, influx, auditStatus)

	log.Entry().Infof("Counted %v violations, details: %v", numberOfViolations, auditStatus)

	influx.fortify_data.fields.projectName = *project.Name
	influx.fortify_data.fields.projectVersion = *projectVersion.Name
	influx.fortify_data.fields.violations = numberOfViolations
	if numberOfViolations > 0 {
		return errors.New("fortify scan failed, the project is not compliant. For details check the archived report")
	}
	return nil
}

func analyseUnauditedIssues(config fortifyExecuteScanOptions, sys fortify.System, projectVersion *models.ProjectVersion, filterSet *models.FilterSet, issueFilterSelectorSet *models.IssueFilterSelectorSet, influx *fortifyExecuteScanInflux, auditStatus map[string]string) int {
	log.Entry().Info("Analyzing unaudited issues")
	reducedFilterSelectorSet := sys.ReduceIssueFilterSelectorSet(issueFilterSelectorSet, []string{"Folder"}, nil)
	fetchedIssueGroups, err := sys.GetProjectIssuesByIDAndFilterSetGroupedBySelector(projectVersion.ID, "", filterSet.GUID, reducedFilterSelectorSet)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Failed to fetch project version issue groups with filter set %v and selector %v for project version ID %v", filterSet, issueFilterSelectorSet, projectVersion.ID)
	}
	overallViolations := 0
	for _, issueGroup := range fetchedIssueGroups {
		overallViolations += getIssueDeltaFor(config, sys, issueGroup, projectVersion.ID, filterSet, issueFilterSelectorSet, influx, auditStatus)
	}
	return overallViolations
}

func getIssueDeltaFor(config fortifyExecuteScanOptions, sys fortify.System, issueGroup *models.ProjectVersionIssueGroup, projectVersionID int64, filterSet *models.FilterSet, issueFilterSelectorSet *models.IssueFilterSelectorSet, influx *fortifyExecuteScanInflux, auditStatus map[string]string) int {
	totalMinusAuditedDelta := 0
	group := ""
	total := 0
	audited := 0
	if issueGroup != nil {
		group = *issueGroup.ID
		total = int(*issueGroup.TotalCount)
		audited = int(*issueGroup.AuditedCount)
	}
	groupTotalMinusAuditedDelta := total - audited
	if groupTotalMinusAuditedDelta > 0 {
		reducedFilterSelectorSet := sys.ReduceIssueFilterSelectorSet(issueFilterSelectorSet, []string{"Folder", "Analysis"}, []string{group})
		folderSelector := sys.GetFilterSetByDisplayName(reducedFilterSelectorSet, "Folder")
		if folderSelector == nil {
			log.Entry().Fatal("folder selector not found")
		}
		analysisSelector := sys.GetFilterSetByDisplayName(reducedFilterSelectorSet, "Analysis")

		auditStatus[group] = fmt.Sprintf("%v total : %v audited", total, audited)

		if strings.Contains(config.MustAuditIssueGroups, group) {
			totalMinusAuditedDelta += groupTotalMinusAuditedDelta
			if group == "Corporate Security Requirements" {
				influx.fortify_data.fields.corporateTotal = total
				influx.fortify_data.fields.corporateAudited = audited
			}
			if group == "Audit All" {
				influx.fortify_data.fields.auditAllTotal = total
				influx.fortify_data.fields.auditAllAudited = audited
			}
			log.Entry().Errorf("[projectVersionId %v]: Unaudited %v detected, count %v", projectVersionID, group, totalMinusAuditedDelta)
			logIssueURL(config, projectVersionID, folderSelector, analysisSelector)
		}

		if strings.Contains(config.SpotAuditIssueGroups, group) {
			log.Entry().Infof("Analyzing %v", config.SpotAuditIssueGroups)
			filter := fmt.Sprintf("%v:%v", folderSelector.EntityType, folderSelector.SelectorOptions[0].Value)
			fetchedIssueGroups, err := sys.GetProjectIssuesByIDAndFilterSetGroupedBySelector(projectVersionID, filter, filterSet.GUID, sys.ReduceIssueFilterSelectorSet(issueFilterSelectorSet, []string{"Category"}, nil))
			if err != nil {
				log.Entry().WithError(err).Fatalf("Failed to fetch project version issue groups with filter %v, filter set %v and selector %v for project version ID %v", filter, filterSet, issueFilterSelectorSet, projectVersionID)
			}
			totalMinusAuditedDelta += getSpotIssueCount(config, sys, fetchedIssueGroups, projectVersionID, filterSet, reducedFilterSelectorSet, influx, auditStatus)
		}
	}
	return totalMinusAuditedDelta
}

func getSpotIssueCount(config fortifyExecuteScanOptions, sys fortify.System, spotCheckCategories []*models.ProjectVersionIssueGroup, projectVersionID int64, filterSet *models.FilterSet, issueFilterSelectorSet *models.IssueFilterSelectorSet, influx *fortifyExecuteScanInflux, auditStatus map[string]string) int {
	overallDelta := 0
	overallIssues := 0
	overallIssuesAudited := 0
	for _, issueGroup := range spotCheckCategories {
		group := ""
		total := 0
		audited := 0
		if issueGroup != nil {
			group = *issueGroup.ID
			total = int(*issueGroup.TotalCount)
			audited = int(*issueGroup.AuditedCount)
		}
		flagOutput := ""

		if ((total <= config.SpotCheckMinimum || config.SpotCheckMinimum < 0) && audited != total) || (total > config.SpotCheckMinimum && audited < config.SpotCheckMinimum) {
			currentDelta := config.SpotCheckMinimum - audited
			if config.SpotCheckMinimum < 0 || config.SpotCheckMinimum > total {
				currentDelta = total - audited
			}
			if currentDelta > 0 {
				filterSelectorFolder := sys.GetFilterSetByDisplayName(issueFilterSelectorSet, "Folder")
				filterSelectorAnalysis := sys.GetFilterSetByDisplayName(issueFilterSelectorSet, "Analysis")
				overallDelta += currentDelta
				log.Entry().Errorf("[projectVersionId %v]: %v unaudited spot check issues detected in group %v", projectVersionID, currentDelta, group)
				logIssueURL(config, projectVersionID, filterSelectorFolder, filterSelectorAnalysis)
				flagOutput = checkString
			}
		}

		overallIssues += total
		overallIssuesAudited += audited

		auditStatus[group] = fmt.Sprintf("%v total : %v audited %v", total, audited, flagOutput)
	}

	influx.fortify_data.fields.spotChecksTotal = overallIssues
	influx.fortify_data.fields.spotChecksAudited = overallIssuesAudited
	influx.fortify_data.fields.spotChecksGap = overallDelta

	return overallDelta
}

func analyseSuspiciousExploitable(config fortifyExecuteScanOptions, sys fortify.System, projectVersion *models.ProjectVersion, filterSet *models.FilterSet, issueFilterSelectorSet *models.IssueFilterSelectorSet, influx *fortifyExecuteScanInflux, auditStatus map[string]string) int {
	log.Entry().Info("Analyzing suspicious and exploitable issues")
	reducedFilterSelectorSet := sys.ReduceIssueFilterSelectorSet(issueFilterSelectorSet, []string{"Analysis"}, []string{})
	fetchedGroups, err := sys.GetProjectIssuesByIDAndFilterSetGroupedBySelector(projectVersion.ID, "", filterSet.GUID, reducedFilterSelectorSet)

	suspiciousCount := 0
	exploitableCount := 0
	for _, issueGroup := range fetchedGroups {
		if *issueGroup.ID == "3" {
			suspiciousCount = int(*issueGroup.TotalCount)
		} else if *issueGroup.ID == "4" {
			exploitableCount = int(*issueGroup.TotalCount)
		}
	}

	result := 0
	if (suspiciousCount > 0 && config.ConsiderSuspicious) || exploitableCount > 0 {
		result = result + suspiciousCount + exploitableCount
		log.Entry().Errorf("[projectVersionId %v]: %v suspicious and %v exploitable issues detected", projectVersion.ID, suspiciousCount, exploitableCount)
		log.Entry().Errorf("%v/html/ssc/index.jsp#!/version/%v/fix?issueGrouping=%v_%v&issueFilters=%v_%v", config.ServerURL, projectVersion.ID, reducedFilterSelectorSet.GroupBySet[0].EntityType, reducedFilterSelectorSet.GroupBySet[0].Value, reducedFilterSelectorSet.FilterBySet[0].EntityType, reducedFilterSelectorSet.FilterBySet[0].Value)
	}
	issueStatistics, err := sys.GetIssueStatisticsOfProjectVersion(projectVersion.ID)
	if err != nil {
		log.Entry().WithError(err).Errorf("Failed to fetch project version statistics for project version ID %v", projectVersion.ID)
	}
	auditStatus["Suspicious"] = fmt.Sprintf("%v", suspiciousCount)
	auditStatus["Exploitable"] = fmt.Sprintf("%v", exploitableCount)
	suppressedCount := *issueStatistics[0].SuppressedCount
	if suppressedCount > 0 {
		auditStatus["Suppressed"] = fmt.Sprintf("WARNING: Detected %v suppressed issues which could violate audit compliance!!!", suppressedCount)
	}
	influx.fortify_data.fields.suspicious = suspiciousCount
	influx.fortify_data.fields.exploitable = exploitableCount
	influx.fortify_data.fields.suppressed = int(suppressedCount)

	return result
}

func logIssueURL(config fortifyExecuteScanOptions, projectVersionID int64, folderSelector, analysisSelector *models.IssueFilterSelector) {
	url := fmt.Sprintf("%v/html/ssc/index.jsp#!/version/%v/fix", config.ServerURL, projectVersionID)
	if len(folderSelector.SelectorOptions) > 0 {
		url += fmt.Sprintf("?issueFilters=%v_%v:%v",
			folderSelector.EntityType,
			folderSelector.Value,
			folderSelector.SelectorOptions[0].Value)
	} else {
		log.Entry().Debugf("no 'filter by set' array entries")
	}
	if analysisSelector != nil {
		url += fmt.Sprintf("&issueFilters=%v_%v:",
			analysisSelector.EntityType,
			analysisSelector.Value)
	} else {
		log.Entry().Debugf("no second entry in 'filter by set' array")
	}
	log.Entry().Error(url)
}

func generateAndDownloadQGateReport(config fortifyExecuteScanOptions, sys fortify.System, project *models.Project, projectVersion *models.ProjectVersion) ([]byte, error) {
	log.Entry().Infof("Generating report with template ID %v", config.ReportTemplateID)
	report, err := sys.GenerateQGateReport(project.ID, projectVersion.ID, int64(config.ReportTemplateID), *project.Name, *projectVersion.Name, config.ReportType)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to generate Q-Gate report")
	}
	log.Entry().Debugf("Triggered report generation of report ID %v", report.ID)
	status := report.Status
	for status == "PROCESSING" || status == "SCHED_PROCESSING" {
		time.Sleep(10 * time.Second)
		report, err = sys.GetReportDetails(report.ID)
		if err != nil {
			return []byte{}, fmt.Errorf("Failed to fetch Q-Gate report generation status: %w", err)
		}
		status = report.Status
	}
	data, err := sys.DownloadReportFile(config.ReportDownloadEndpoint, report.ID)
	if err != nil {
		return []byte{}, fmt.Errorf("Failed to download Q-Gate Report: %w", err)
	}
	return data, nil
}

var errProcessing = errors.New("artifact still processing")

func checkArtifactStatus(config fortifyExecuteScanOptions, projectVersionID int64, filterSet *models.FilterSet, artifact *models.Artifact, retries int, pollingDelay, timeout time.Duration) error {
	if "PROCESSING" == artifact.Status || "SCHED_PROCESSING" == artifact.Status {
		pollingTime := time.Duration(retries) * pollingDelay
		if pollingTime >= timeout {
			return fmt.Errorf("terminating after %v since artifact for Project Version %v is still in status %v", timeout, projectVersionID, artifact.Status)
		}
		log.Entry().Infof("Most recent artifact uploaded on %v of Project Version %v is still in status %v...", artifact.UploadDate, projectVersionID, artifact.Status)
		time.Sleep(pollingDelay)
		return errProcessing
	}
	if "REQUIRE_AUTH" == artifact.Status {
		// verify no manual issue approval needed
		return fmt.Errorf("There are artifacts that require manual approval for Project Version %v\n%v/html/ssc/index.jsp#!/version/%v/artifacts?filterSet=%v", projectVersionID, config.ServerURL, projectVersionID, filterSet.GUID)
	}
	if "ERROR_PROCESSING" == artifact.Status {
		return fmt.Errorf("There are artifacts that failed processing for Project Version %v\n%v/html/ssc/index.jsp#!/version/%v/artifacts?filterSet=%v", projectVersionID, config.ServerURL, projectVersionID, filterSet.GUID)
	}
	return nil
}

func verifyScanResultsFinishedUploading(config fortifyExecuteScanOptions, sys fortify.System, projectVersionID int64, buildLabel string, filterSet *models.FilterSet, pollingDelay, timeout time.Duration) error {
	log.Entry().Debug("Verifying scan results have finished uploading and processing")
	var artifacts []*models.Artifact
	var relatedUpload *models.Artifact
	var err error
	retries := 0
	for relatedUpload == nil {
		artifacts, err = sys.GetArtifactsOfProjectVersion(projectVersionID)
		log.Entry().Debugf("Received %v artifacts for project version ID %v", len(artifacts), projectVersionID)
		if err != nil {
			return fmt.Errorf("failed to fetch artifacts of project version ID %v", projectVersionID)
		}
		if len(artifacts) == 0 {
			return fmt.Errorf("no uploaded artifacts for assessment detected for project version with ID %v", projectVersionID)
		}
		latest := artifacts[0]
		err = checkArtifactStatus(config, projectVersionID, filterSet, latest, retries, pollingDelay, timeout)
		if err != nil {
			if err == errProcessing {
				retries++
				continue
			}
			return err
		}
		relatedUpload = findArtifactByBuildLabel(artifacts, buildLabel)
		if relatedUpload == nil {
			log.Entry().Warn("Unable to identify artifact based on the build label, will consider most recent artifact as related to the scan")
			relatedUpload = artifacts[0]
		}
	}

	differenceInSeconds := calculateTimeDifferenceToLastUpload(relatedUpload.UploadDate, projectVersionID)
	// Use the absolute value for checking the time difference
	if differenceInSeconds > float64(60*config.DeltaMinutes) {
		return errors.New("no recent upload detected on Project Version")
	}
	for _, upload := range artifacts {
		if upload.Status == "ERROR_PROCESSING" {
			log.Entry().Warn("Previous uploads detected that failed processing, please ensure that your scans are properly configured")
			break
		}
	}
	return nil
}

func findArtifactByBuildLabel(artifacts []*models.Artifact, buildLabel string) *models.Artifact {
	if len(buildLabel) == 0 {
		return nil
	}
	for _, artifact := range artifacts {
		if len(buildLabel) > 0 && artifact.Embed != nil && artifact.Embed.Scans != nil && len(artifact.Embed.Scans) > 0 {
			scan := artifact.Embed.Scans[0]
			if scan != nil && strings.HasSuffix(scan.BuildLabel, buildLabel) {
				return artifact
			}
		}
	}
	return nil
}

func calculateTimeDifferenceToLastUpload(uploadDate models.Iso8601MilliDateTime, projectVersionID int64) float64 {
	log.Entry().Infof("Last upload on project version %v happened on %v", projectVersionID, uploadDate)
	uploadDateAsTime := time.Time(uploadDate)
	duration := time.Since(uploadDateAsTime)
	log.Entry().Debugf("Difference duration is %v", duration)
	absoluteSeconds := math.Abs(duration.Seconds())
	log.Entry().Infof("Difference since %v in seconds is %v", uploadDateAsTime, absoluteSeconds)
	return absoluteSeconds
}

func executeTemplatedCommand(command fortifyExecRunner, cmdTemplate []string, context map[string]string) {
	for index, cmdTemplatePart := range cmdTemplate {
		result, err := piperutils.ExecuteTemplate(cmdTemplatePart, context)
		if err != nil {
			log.Entry().WithError(err).Fatalf("Failed to transform template for command fragment: %v", cmdTemplatePart)
		}
		cmdTemplate[index] = result
	}
	err := command.RunExecutable(cmdTemplate[0], cmdTemplate[1:]...)
	if err != nil {
		log.Entry().WithError(err).WithField("command", cmdTemplate).Fatal("Failed to execute command")
	}
}

func autoresolvePipClasspath(executable string, parameters []string, file string, command fortifyExecRunner) string {
	// redirect stdout and create cp file from command output
	outfile, err := os.Create(file)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to create classpath file")
	}
	defer outfile.Close()
	command.Stdout(outfile)
	err = command.RunExecutable(executable, parameters...)
	if err != nil {
		log.Entry().WithError(err).WithField("command", fmt.Sprintf("%v with parameters %v", executable, parameters)).Fatal("Failed to run classpath autodetection command")
	}
	command.Stdout(log.Entry().Writer())
	return readClasspathFile(file)
}

func autoresolveMavenClasspath(config fortifyExecuteScanOptions, file string, command fortifyExecRunner) string {
	if filepath.IsAbs(file) {
		log.Entry().Warnf("Passing an absolute path for -Dmdep.outputFile results in the classpath only for the last module in multi-module maven projects.")
	}
	executeOptions := maven.ExecuteOptions{
		PomPath:             config.BuildDescriptorFile,
		ProjectSettingsFile: config.ProjectSettingsFile,
		GlobalSettingsFile:  config.GlobalSettingsFile,
		M2Path:              config.M2Path,
		Goals:               []string{"dependency:build-classpath"},
		Defines:             []string{fmt.Sprintf("-Dmdep.outputFile=%v", file), "-DincludeScope=compile"},
		ReturnStdout:        false,
	}
	_, err := maven.Execute(&executeOptions, command)
	if err != nil {
		log.Entry().WithError(err).Warn("failed to determine classpath using Maven")
	}
	return readAllClasspathFiles(file)
}

// readAllClasspathFiles tests whether the passed file is an absolute path. If not, it will glob for
// all files under the current directory with the given file name and concatenate their contents.
// Otherwise it will return the contents pointed to by the absolute path.
func readAllClasspathFiles(file string) string {
	var paths []string
	if filepath.IsAbs(file) {
		paths = []string{file}
	} else {
		paths, _ = doublestar.Glob(filepath.Join("**", file))
		log.Entry().Debugf("Concatenating the class paths from %v", paths)
	}
	var contents string
	const separator = ":"
	for _, path := range paths {
		contents += separator + readClasspathFile(path)
	}
	return removeDuplicates(contents, separator)
}

func readClasspathFile(file string) string {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Entry().WithError(err).Warnf("failed to read classpath from file '%v'", file)
	}
	result := strings.TrimSpace(string(data))
	if len(result) == 0 {
		log.Entry().Warnf("classpath from file '%v' was empty", file)
	}
	return result
}

func removeDuplicates(contents, separator string) string {
	if separator == "" || contents == "" {
		return contents
	}
	entries := strings.Split(contents, separator)
	entrySet := map[string]struct{}{}
	contents = ""
	for _, entry := range entries {
		if entry == "" {
			continue
		}
		_, contained := entrySet[entry]
		if !contained {
			entrySet[entry] = struct{}{}
			contents += entry + separator
		}
	}
	if contents != "" {
		// Remove trailing "separator"
		contents = contents[:len(contents)-len(separator)]
	}
	return contents
}

func triggerFortifyScan(config fortifyExecuteScanOptions, command fortifyExecRunner, buildID, buildLabel, buildProject string) {
	var err error = nil
	// Do special Python related prep
	pipVersion := "pip3"
	if config.PythonVersion != "python3" {
		pipVersion = "pip2"
	}

	classpath := ""
	if config.BuildTool == "maven" {
		if config.AutodetectClasspath {
			classpath = autoresolveMavenClasspath(config, classpathFileName, command)
		}
		config.Translate, err = populateMavenTranslate(&config, classpath)
		if err != nil {
			log.Entry().WithError(err).Warnf("failed to apply src ('%s') or exclude ('%s') parameter", config.Src, config.Exclude)
		}
	}
	if config.BuildTool == "pip" {
		if config.AutodetectClasspath {
			separator := getSeparator()
			script := fmt.Sprintf("import sys;p=sys.path;p.remove('');print('%v'.join(p))", separator)
			classpath = autoresolvePipClasspath(config.PythonVersion, []string{"-c", script}, classpathFileName, command)
		}
		// install the dev dependencies
		if len(config.PythonRequirementsFile) > 0 {
			context := map[string]string{}
			cmdTemplate := []string{pipVersion, "install", "--user", "-r", config.PythonRequirementsFile}
			cmdTemplate = append(cmdTemplate, tokenize(config.PythonRequirementsInstallSuffix)...)
			executeTemplatedCommand(command, cmdTemplate, context)
		}

		executeTemplatedCommand(command, tokenize(config.PythonInstallCommand), map[string]string{"Pip": pipVersion})

		config.Translate, err = populatePipTranslate(&config, classpath)
		if err != nil {
			log.Entry().WithError(err).Warnf("failed to apply pythonAdditionalPath ('%s') or src ('%s') parameter", config.PythonAdditionalPath, config.Src)
		}

	}

	translateProject(&config, command, buildID, classpath)

	scanProject(&config, command, buildID, buildLabel, buildProject)
}

func populatePipTranslate(config *fortifyExecuteScanOptions, classpath string) (string, error) {
	if len(config.Translate) > 0 {
		return config.Translate, nil
	}

	var translateList []map[string]interface{}
	translateList = append(translateList, make(map[string]interface{}))

	separator := getSeparator()

	translateList[0]["pythonPath"] = classpath + separator +
		getSuppliedOrDefaultListAsString(config.PythonAdditionalPath, []string{}, separator)
	translateList[0]["src"] = getSuppliedOrDefaultListAsString(
		config.Src, []string{"./**/*"}, ":")
	translateList[0]["exclude"] = getSuppliedOrDefaultListAsString(
		config.Exclude, []string{"./**/tests/**/*", "./**/setup.py"}, separator)

	translateJSON, err := json.Marshal(translateList)

	return string(translateJSON), err
}

func populateMavenTranslate(config *fortifyExecuteScanOptions, classpath string) (string, error) {
	if len(config.Translate) > 0 {
		return config.Translate, nil
	}

	var translateList []map[string]interface{}
	translateList = append(translateList, make(map[string]interface{}))
	translateList[0]["classpath"] = classpath

	setTranslateEntryIfNotEmpty(translateList[0], "src", ":", config.Src,
		[]string{"**/*.xml", "**/*.html", "**/*.jsp", "**/*.js", "**/src/main/resources/**/*", "**/src/main/java/**/*"})

	setTranslateEntryIfNotEmpty(translateList[0], "exclude", getSeparator(), config.Exclude, []string{})

	translateJSON, err := json.Marshal(translateList)

	return string(translateJSON), err
}

func translateProject(config *fortifyExecuteScanOptions, command fortifyExecRunner, buildID, classpath string) {
	var translateList []map[string]string
	json.Unmarshal([]byte(config.Translate), &translateList)
	log.Entry().Debugf("Translating with options: %v", translateList)
	for _, translate := range translateList {
		if len(classpath) > 0 {
			translate["autoClasspath"] = classpath
		}
		handleSingleTranslate(config, command, buildID, translate)
	}
}

func handleSingleTranslate(config *fortifyExecuteScanOptions, command fortifyExecRunner, buildID string, t map[string]string) {
	if t != nil {
		log.Entry().Debugf("Handling translate config %v", t)
		translateOptions := []string{
			"-verbose",
			"-64",
			"-b",
			buildID,
		}
		translateOptions = append(translateOptions, tokenize(config.Memory)...)
		translateOptions = appendToOptions(config, translateOptions, t)
		log.Entry().Debugf("Running sourceanalyzer translate command with options %v", translateOptions)
		err := command.RunExecutable("sourceanalyzer", translateOptions...)
		if err != nil {
			log.Entry().WithError(err).WithField("translateOptions", translateOptions).Fatal("failed to execute sourceanalyzer translate command")
		}
	} else {
		log.Entry().Debug("Skipping translate with nil value")
	}
}

func scanProject(config *fortifyExecuteScanOptions, command fortifyExecRunner, buildID, buildLabel, buildProject string) {
	var scanOptions = []string{
		"-verbose",
		"-64",
		"-b",
		buildID,
		"-scan",
	}
	scanOptions = append(scanOptions, tokenize(config.Memory)...)
	if config.QuickScan {
		scanOptions = append(scanOptions, "-quick")
	}
	if len(buildLabel) > 0 {
		scanOptions = append(scanOptions, "-build-label", buildLabel)
	}
	if len(buildProject) > 0 {
		scanOptions = append(scanOptions, "-build-project", buildProject)
	}
	scanOptions = append(scanOptions, "-logfile", "target/fortify-scan.log", "-f", "target/result.fpr")

	err := command.RunExecutable("sourceanalyzer", scanOptions...)
	if err != nil {
		log.Entry().WithError(err).WithField("scanOptions", scanOptions).Fatal("failed to execute sourceanalyzer scan command")
	}
}

func determinePullRequestMerge(config fortifyExecuteScanOptions) string {
	ctx, client, err := piperGithub.NewClient(config.GithubToken, config.GithubAPIURL, "")
	if err == nil {
		result, err := determinePullRequestMergeGithub(ctx, config, client.PullRequests)
		if err != nil {
			log.Entry().WithError(err).Warn("Failed to get PR metadata via GitHub client")
		} else {
			return result
		}
	}

	log.Entry().Infof("Trying to determine PR ID in commit message: %v", config.CommitMessage)
	r, _ := regexp.Compile(config.PullRequestMessageRegex)
	matches := r.FindSubmatch([]byte(config.CommitMessage))
	if matches != nil && len(matches) > 1 {
		return string(matches[config.PullRequestMessageRegexGroup])
	}
	return ""
}

func determinePullRequestMergeGithub(ctx context.Context, config fortifyExecuteScanOptions, pullRequestServiceInstance pullRequestService) (string, error) {
	options := github.PullRequestListOptions{State: "closed", Sort: "updated", Direction: "desc"}
	prList, _, err := pullRequestServiceInstance.ListPullRequestsWithCommit(ctx, config.Owner, config.Repository, config.CommitID, &options)
	if err == nil && len(prList) > 0 {
		return fmt.Sprintf("%v", prList[0].GetNumber()), nil
	}
	return "", err
}

func appendToOptions(config *fortifyExecuteScanOptions, options []string, t map[string]string) []string {
	switch config.BuildTool {
	case "windows":
		if len(t["aspnetcore"]) > 0 {
			options = append(options, "-aspnetcore")
		}
		if len(t["dotNetCoreVersion"]) > 0 {
			options = append(options, "-dotnet-core-version", t["dotNetCoreVersion"])
		}
		if len(t["libDirs"]) > 0 {
			options = append(options, "-libdirs", t["libDirs"])
		}

	case "maven":
		if len(t["autoClasspath"]) > 0 {
			options = append(options, "-cp", t["autoClasspath"])
		} else if len(t["classpath"]) > 0 {
			options = append(options, "-cp", t["classpath"])
		} else {
			log.Entry().Debugf("no field 'autoClasspath' or 'classpath' in map or both empty")
		}
		if len(t["extdirs"]) > 0 {
			options = append(options, "-extdirs", t["extdirs"])
		}
		if len(t["javaBuildDir"]) > 0 {
			options = append(options, "-java-build-dir", t["javaBuildDir"])
		}
		if len(t["source"]) > 0 {
			options = append(options, "-source", t["source"])
		}
		if len(t["jdk"]) > 0 {
			options = append(options, "-jdk", t["jdk"])
		}
		if len(t["sourcepath"]) > 0 {
			options = append(options, "-sourcepath", t["sourcepath"])
		}

	case "pip":
		if len(t["autoClasspath"]) > 0 {
			options = append(options, "-python-path", t["autoClasspath"])
		} else if len(t["pythonPath"]) > 0 {
			options = append(options, "-python-path", t["pythonPath"])
		}
		if len(t["djangoTemplatDirs"]) > 0 {
			options = append(options, "-django-template-dirs", t["djangoTemplatDirs"])
		}

	default:
		return options
	}

	if len(t["exclude"]) > 0 {
		options = append(options, "-exclude", t["exclude"])
	}
	return append(options, strings.Split(t["src"], ":")...)
}

func getSuppliedOrDefaultList(suppliedList, defaultList []string) []string {
	if len(suppliedList) > 0 {
		return suppliedList
	}
	return defaultList
}

func getSuppliedOrDefaultListAsString(suppliedList, defaultList []string, separator string) string {
	effectiveList := getSuppliedOrDefaultList(suppliedList, defaultList)
	return strings.Join(effectiveList, separator)
}

// setTranslateEntryIfNotEmpty builds a string from either the user-supplied list, or the default list,
// by joining the entries with the given separator. If the resulting string is not empty, it will be
// placed as an entry in the provided map under the given key.
func setTranslateEntryIfNotEmpty(translate map[string]interface{}, key, separator string, suppliedList, defaultList []string) {
	value := getSuppliedOrDefaultListAsString(suppliedList, defaultList, separator)
	if value != "" {
		translate[key] = value
	}
}

// getSeparator returns the separator string depending on the host platform. This assumes that
// Piper executes the Fortify command line tools within the same OS platform as it is running on itself.
func getSeparator() string {
	if runtime.GOOS == "windows" {
		return ";"
	}
	return ":"
}
