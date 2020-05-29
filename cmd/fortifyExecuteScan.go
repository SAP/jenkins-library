package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/v28/github"
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

const checkString = "<---CHECK FORTIFY---"
const classpathFileName = "cp.txt"

func fortifyExecuteScan(config fortifyExecuteScanOptions, telemetryData *telemetry.CustomData, influx *fortifyExecuteScanInflux) {
	auditStatus := map[string]string{}
	sys := fortify.NewSystemInstance(config.ServerURL, config.APIEndpoint, config.AuthToken, time.Second*30)
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())
	err := runFortifyScan(config, sys, &c, telemetryData, influx, auditStatus)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Fortify scan and check failed")
	}
}

func runFortifyScan(config fortifyExecuteScanOptions, sys fortify.System, command execRunner, telemetryData *telemetry.CustomData, influx *fortifyExecuteScanInflux, auditStatus map[string]string) error {
	log.Entry().Debugf("Running Fortify scan against SSC at %v", config.ServerURL)
	artifact, err := versioning.GetArtifact(config.BuildTool, config.BuildDescriptorFile, &versioning.Options{}, command)
	if err != nil {
		return fmt.Errorf("unable to get artifact from descriptor %v: %w", config.BuildDescriptorFile, err)
	}
	gav, err := artifact.GetCoordinates()
	if err != nil {
		return fmt.Errorf("unable to get project coordinates from descriptor %v: %w", config.BuildDescriptorFile, err)
	}
	log.Entry().Debugf("determined project coordinates %v", gav)
	fortifyProjectName, fortifyProjectVersion := versioning.DetermineProjectCoordinates(config.ProjectName, config.DefaultVersioningModel, gav)
	project, err := sys.GetProjectByName(fortifyProjectName, config.AutoCreate, fortifyProjectVersion)
	if err != nil {
		return fmt.Errorf("Failed to load project %v: %w", fortifyProjectName, err)
	}
	projectVersion, err := sys.GetProjectVersionDetailsByProjectIDAndVersionName(project.ID, fortifyProjectVersion, config.AutoCreate, fortifyProjectName)
	if err != nil {
		return fmt.Errorf("Failed to load project version %v: %w", fortifyProjectVersion, err)
	}

	if len(config.PullRequestName) > 0 {
		fortifyProjectVersion = config.PullRequestName
		projectVersion, err := sys.LookupOrCreateProjectVersionDetailsForPullRequest(project.ID, projectVersion, fortifyProjectVersion)
		if err != nil {
			return fmt.Errorf("Failed to lookup / create project version for pull request %v: %w", fortifyProjectVersion, err)
		}
		log.Entry().Debugf("Looked up / created project version with ID %v for PR %v", projectVersion.ID, fortifyProjectVersion)
	} else {
		prID := determinePullRequestMerge(config)
		if len(prID) > 0 {
			log.Entry().Debugf("Determined PR ID '%v' for merge check", prID)
			pullRequestProjectName := fmt.Sprintf("PR-%v", prID)
			err = sys.MergeProjectVersionStateOfPRIntoMaster(config.FprDownloadEndpoint, config.FprUploadEndpoint, project.ID, projectVersion.ID, pullRequestProjectName)
			if err != nil {
				return fmt.Errorf("Failed to merge project version state for pull request %v into project version %v of project %v: %w", pullRequestProjectName, fortifyProjectVersion, project.ID, err)
			}
		}
	}

	log.Entry().Debugf("Scanning and uploading to project %v with version %v and projectVersionId %v", fortifyProjectName, fortifyProjectVersion, projectVersion.ID)
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

	var reports []piperutils.Path
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
	piperutils.PersistReportsAndLinks("fortifyExecuteScan", config.ModulePath, reports, nil)
	if err != nil {
		return fmt.Errorf(message+": %w", err)
	}

	log.Entry().Debugf("Starting audit status check on project %v with version %v and project version ID %v", fortifyProjectName, fortifyProjectVersion, projectVersion.ID)
	filterSet, err := sys.GetFilterSetOfProjectVersionByTitle(projectVersion.ID, config.FilterSetTitle)
	if filterSet == nil || err != nil {
		return fmt.Errorf("Failed to load filter set with title %v", config.FilterSetTitle)
	}

	// Ensure latest FPR is processed
	err = verifyScanResultsFinishedUploading(config, sys, projectVersion.ID, buildLabel, filterSet, 0)
	if err != nil {
		return err
	}

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

	influx.fortify_data.fields.projectName = fortifyProjectName
	influx.fortify_data.fields.projectVersion = fortifyProjectVersion
	influx.fortify_data.fields.violations = fmt.Sprintf("%v", numberOfViolations)
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
				influx.fortify_data.fields.corporateTotal = fmt.Sprintf("%v", total)
				influx.fortify_data.fields.corporateAudited = fmt.Sprintf("%v", audited)
			}
			if group == "Audit All" {
				influx.fortify_data.fields.auditAllTotal = fmt.Sprintf("%v", total)
				influx.fortify_data.fields.auditAllAudited = fmt.Sprintf("%v", audited)
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

	influx.fortify_data.fields.spotChecksTotal = fmt.Sprintf("%v", overallIssues)
	influx.fortify_data.fields.spotChecksAudited = fmt.Sprintf("%v", overallIssuesAudited)
	influx.fortify_data.fields.spotChecksGap = fmt.Sprintf("%v", overallDelta)

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
	influx.fortify_data.fields.suspicious = fmt.Sprintf("%v", suspiciousCount)
	influx.fortify_data.fields.exploitable = fmt.Sprintf("%v", exploitableCount)
	influx.fortify_data.fields.suppressed = fmt.Sprintf("%v", suppressedCount)

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
	for status != "Complete" && status != "Error Processing" {
		time.Sleep(10 * time.Second)
		report, err = sys.GetReportDetails(report.ID)
		if err != nil {
			return []byte{}, fmt.Errorf("Failed to fetch Q-Gate report generation status: %w", err)
		}
		status = report.Status
	}
	data, err := sys.DownloadReportFile(config.ReportDownloadEndpoint, projectVersion.ID)
	if err != nil {
		return []byte{}, fmt.Errorf("Failed to download Q-Gate Report: %w", err)
	}
	return data, nil
}

func checkArtifactStatus(config fortifyExecuteScanOptions, sys fortify.System, projectVersionID int64, buildLabel string, filterSet *models.FilterSet, artifact *models.Artifact, numInvokes int) error {
	numInvokes++
	if "PROCESSING" == artifact.Status || "SCHED_PROCESSING" == artifact.Status {
		if numInvokes >= (config.PollingMinutes * 6) {
			return fmt.Errorf("Terminating after %v minutes since artifact for Project Version %v is still in status %v", config.PollingMinutes, projectVersionID, artifact.Status)
		}
		log.Entry().Infof("Most recent artifact uploaded on %v of Project Version %v is still in status %v...", artifact.UploadDate, projectVersionID, artifact.Status)
		time.Sleep(10 * time.Second)
		return verifyScanResultsFinishedUploading(config, sys, projectVersionID, buildLabel, filterSet, numInvokes)
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

func verifyScanResultsFinishedUploading(config fortifyExecuteScanOptions, sys fortify.System, projectVersionID int64, buildLabel string, filterSet *models.FilterSet, numInvokes int) error {
	log.Entry().Debug("Verifying scan results have finished uploading and processing")
	var artifacts []*models.Artifact
	var relatedUpload *models.Artifact
	for relatedUpload == nil {
		artifacts, err := sys.GetArtifactsOfProjectVersion(projectVersionID)
		log.Entry().Debugf("Received %v artifacts for project version ID %v", len(artifacts), projectVersionID)
		if err != nil {
			log.Entry().WithError(err).Fatalf("Failed to fetch artifacts of project version ID %v", projectVersionID)
		}
		if len(artifacts) > 0 {
			latest := artifacts[0]
			if err := checkArtifactStatus(config, sys, projectVersionID, buildLabel, filterSet, latest, numInvokes); err != nil {
				return err
			}
			notFound := true
			for _, artifact := range artifacts {
				if len(buildLabel) > 0 && artifact.Embed != nil && artifact.Embed.Scans != nil && len(artifact.Embed.Scans) > 0 {
					scan := artifact.Embed.Scans[0]
					if notFound && scan != nil && strings.HasSuffix(scan.BuildLabel, buildLabel) {
						relatedUpload = artifact
						notFound = false
					}
				}
			}
		} else {
			return fmt.Errorf("No uploaded artifacts for assessment detected for project version with ID %v", projectVersionID)
		}
		if relatedUpload == nil {
			log.Entry().Warn("Unable to identify artifact based on the build label, will consider most recent artifact as related to the scan")
			relatedUpload = artifacts[0]
		}
	}

	differenceInSeconds := calculateTimeDifferenceToLastUpload(relatedUpload.UploadDate, projectVersionID)
	// Use the absolute value for checking the time difference
	if differenceInSeconds > float64(60*config.DeltaMinutes) {
		return errors.New("No recent upload detected on Project Version")
	}
	warn := false
	for _, upload := range artifacts {
		if upload.Status == "ERROR_PROCESSING" {
			warn = true
		}
	}
	if warn {
		log.Entry().Warn("Previous uploads detected that failed processing, please ensure that your scans are properly configured")
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

func executeTemplatedCommand(command execRunner, cmdTemplate []string, context map[string]string) {
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

func autoresolvePipClasspath(executable string, parameters []string, file string, command execRunner) string {
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

func autoresolveMavenClasspath(pomFilePath, file string, command execRunner) string {
	executeOptions := maven.ExecuteOptions{
		PomPath:      pomFilePath,
		Goals:        []string{"dependency:build-classpath"},
		Defines:      []string{fmt.Sprintf("-Dmdep.outputFile=%v", file), "-DincludeScope=compile"},
		ReturnStdout: false,
	}
	_, err := maven.Execute(&executeOptions, command)
	if err != nil {
		log.Entry().WithError(err).Warn("failed to determine classpath using Maven")
	}
	return readClasspathFile(file)
}

func readClasspathFile(file string) string {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Entry().WithError(err).Warnf("failed to read classpath from file '%v'", file)
	}
	return strings.TrimSpace(string(data))
}

func triggerFortifyScan(config fortifyExecuteScanOptions, command execRunner, buildID, buildLabel, buildProject string) {
	var err error = nil
	// Do special Python related prep
	pipVersion := "pip3"
	if config.PythonVersion != "python3" {
		pipVersion = "pip2"
	}

	classpath := ""
	if config.BuildTool == "maven" {
		if config.AutodetectClasspath {
			classpath = autoresolveMavenClasspath(config.BuildDescriptorFile, classpathFileName, command)
		}
		config.Translate, err = populateMavenTranslate(&config, classpath)
		if err != nil {
			log.Entry().WithError(err).Warnf("failed to apply src ('%s') or exclude ('%s') parameter", config.Src, config.Exclude)
		}
	}
	if config.BuildTool == "pip" {
		if config.AutodetectClasspath {
			classpath = autoresolvePipClasspath(config.PythonVersion, []string{"-c", "import sys;p=sys.path;p.remove('');print(';'.join(p))"}, classpathFileName, command)
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
			log.Entry().WithError(err).Warnf("failed to apply pythonAdditionalPath ('%s') or pythonIncludes ('%s') parameter", config.PythonAdditionalPath, config.PythonIncludes)
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

	translateList[0]["pythonPath"] = classpath + ";" + config.PythonAdditionalPath
	translateList[0]["pythonIncludes"] = config.PythonIncludes
	translateList[0]["pythonExcludes"] = strings.ReplaceAll(config.PythonExcludes, "-exclude ", "")

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

	if len(config.Src) > 0 {
		translateList[0]["src"] = config.Src
	}
	if len(config.Exclude) > 0 {
		translateList[0]["exclude"] = config.Exclude
	}

	translateJSON, err := json.Marshal(translateList)

	return string(translateJSON), err
}

func translateProject(config *fortifyExecuteScanOptions, command execRunner, buildID, classpath string) {
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

func handleSingleTranslate(config *fortifyExecuteScanOptions, command execRunner, buildID string, t map[string]string) {
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

func scanProject(config *fortifyExecuteScanOptions, command execRunner, buildID, buildLabel, buildProject string) {
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
	if config.BuildTool == "windows" {
		if len(t["aspnetcore"]) > 0 {
			options = append(options, "-aspnetcore")
		}
		if len(t["dotNetCoreVersion"]) > 0 {
			options = append(options, "-dotnet-core-version", t["dotNetCoreVersion"])
		}
		if len(t["exclude"]) > 0 {
			options = append(options, "-exclude", t["exclude"])
		}
		if len(t["libDirs"]) > 0 {
			options = append(options, "-libdirs", t["libDirs"])
		}
		return append(options, tokenize(t["src"])...)
	}
	if config.BuildTool == "maven" {
		if len(t["autoClasspath"]) > 0 {
			options = append(options, "-cp", t["autoClasspath"])
		} else if len(t["classpath"]) > 0 {
			options = append(options, "-cp", t["classpath"])
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
		if len(t["exclude"]) > 0 {
			options = append(options, "-exclude", t["exclude"])
		}
		return append(options, tokenize(t["src"])...)
	}
	if config.BuildTool == "pip" {
		if len(t["autoClasspath"]) > 0 {
			options = append(options, "-python-path", t["autoClasspath"])
		} else if len(t["pythonPath"]) > 0 {
			options = append(options, "-python-path", t["pythonPath"])
		}
		if len(t["djangoTemplatDirs"]) > 0 {
			options = append(options, "-django-template-dirs", t["djangoTemplatDirs"])
		}
		if len(t["pythonExcludes"]) > 0 {
			options = append(options, "-exclude", t["pythonExcludes"])
		}
		return append(options, t["pythonIncludes"])
	}
	return options
}
