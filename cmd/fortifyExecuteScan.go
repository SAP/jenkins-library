package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/piper-validation/fortify-client-go/models"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/fortify"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

var auditStatus map[string]string

const checkString = "<---CHECK FORTIFY---"

func fortifyExecuteScan(config fortifyExecuteScanOptions, telemetryData *telemetry.CustomData, influx *fortifyExecuteScanInflux) error {
	auditStatus = map[string]string{}
	sys := fortify.NewSystemInstance(config.ServerURL, config.APIEndpoint, config.AuthToken, time.Second*30)
	c := command.Command{}
	// reroute command output to loging framework
	// also log stdout as Karma reports into it
	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())
	c.Env(os.Environ())
	return runFortifyScan(config, sys, &c, telemetryData, influx)
}

func runFortifyScan(config fortifyExecuteScanOptions, sys fortify.System, command execRunner, telemetryData *telemetry.CustomData, influx *fortifyExecuteScanInflux) error {
	log.Entry().Debugf("Running Fortify scan against SSC at %v", config.ServerURL)
	var gav piperutils.BuildDescriptor
	var err error
	if config.ScanType == "maven" {
		gav, err = piperutils.GetMavenCoordinates(config.BuildDescriptorFile)
	}
	if config.ScanType == "pip" {
		gav, err = piperutils.GetPipCoordinates(config.BuildDescriptorFile)
	}
	if err != nil {
		log.Entry().Warnf("Unable to load project coordinates from descriptor %v: %v", config.BuildDescriptorFile, err)
	}
	fortifyProjectName, fortifyProjectVersion := piperutils.DetermineProjectCoordinates(config.ProjectName, config.ProjectVersion, gav)
	project, err := sys.GetProjectByName(fortifyProjectName)
	if err != nil {
		log.Entry().Fatalf("Failed to load project %v: %v", fortifyProjectName, err)
	}
	projectVersion, err := sys.GetProjectVersionDetailsByProjectIDAndVersionName(project.ID, fortifyProjectVersion)
	if err != nil {
		log.Entry().Fatalf("Failed to load project version %v: %v", fortifyProjectVersion, err)
	}
	if len(config.PullRequestName) > 0 {
		fortifyProjectVersion = config.PullRequestName
		projectVersion, err := sys.LookupOrCreateProjectVersionDetailsForPullRequest(project.ID, projectVersion, fortifyProjectVersion)
		if err != nil {
			log.Entry().Fatalf("Failed to lookup / create project version for pull request %v: %v", fortifyProjectVersion, err)
		}
		log.Entry().Debugf("Looked up / created project version with ID %v for PR %v", projectVersion.ID, fortifyProjectVersion)
	} else {
		prID := determinePullRequestMerge(config)
		if len(prID) > 0 {
			log.Entry().Debugf("Determined PR identifier %v for merge check", prID)
			err = sys.MergeProjectVersionStateOfPRIntoMaster(config.FprDownloadEndpoint, config.FprUploadEndpoint, project.ID, projectVersion.ID, fmt.Sprintf("PR-%v", prID))
			if err != nil {
				log.Entry().Fatalf("Failed to merge project version state for pull request %v: %v", fortifyProjectVersion, err)
			}
		}
	}

	log.Entry().Debugf("Scanning and uploading to project %v with version %v and projectVersionId %v", fortifyProjectName, fortifyProjectVersion, projectVersion.ID)
	repoURL := strings.ReplaceAll(config.RepoURL, ".git", "")
	buildLabel := fmt.Sprintf("%v/commit/%v", repoURL, config.CommitID)

	// Create sourceanalyzer / maven command based on configuration
	if config.ScanType == "maven" {
		// Create and execute special maven command

	} else {
		buildID := uuid.New().String()
		command.Dir(config.ModulePath)
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

		triggerFortifyScan(config, command, buildID, buildLabel)
	}

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
		log.Entry().WithError(err).Fatal(message)
	}

	log.Entry().Debugf("Starting audit status check on project %v with version %v and project version ID %v", fortifyProjectName, fortifyProjectVersion, projectVersion.ID)
	filterSet, err := sys.GetFilterSetOfProjectVersionByTitle(projectVersion.ID, config.FilterSetTitle)
	if filterSet == nil || err != nil {
		log.Entry().Fatalf("Failed to load filter set with title %v", config.FilterSetTitle)
	}

	// Ensure latest FPR is processed
	verifyScanResultsFinishedUploading(config, sys, projectVersion.ID, buildLabel, filterSet, 0)

	// Generate report
	if config.Reporting {
		generateAndDownloadQGateReport(config, sys, project, projectVersion)
	}

	// Perform audit compliance checks
	issueFilterSelectorSet, err := sys.GetIssueFilterSelectorOfProjectVersionByName(projectVersion.ID, []string{"Analysis", "Folder", "Category"}, nil)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Failed to fetch project version issue filter selector for project version ID %v", projectVersion.ID)
	}
	numberOfViolations := analyseUnauditedIssues(config, sys, projectVersion, filterSet, issueFilterSelectorSet, influx)
	numberOfViolations += analyseSuspiciousExploitable(config, sys, projectVersion, filterSet, issueFilterSelectorSet, influx)

	auditStatusOutput := auditStatus
	log.Entry().Infof("Counted %v violations, details: %v", numberOfViolations, auditStatusOutput)

	influx.fortify_data.fields.projectName = fortifyProjectName
	influx.fortify_data.fields.projectVersion = fortifyProjectVersion
	influx.fortify_data.fields.violations = fmt.Sprintf("%v", numberOfViolations)
	if numberOfViolations > 0 {
		return errors.New("fortify scan failed, the project is not compliant. For details check the archived report")
	}

	return nil
}

func analyseUnauditedIssues(config fortifyExecuteScanOptions, sys fortify.System, projectVersion *models.ProjectVersion, filterSet *models.FilterSet, issueFilterSelectorSet *models.IssueFilterSelectorSet, influx *fortifyExecuteScanInflux) int {
	log.Entry().Info("Analyzing unaudited issues")
	reducedFilterSelectorSet := sys.ReduceIssueFilterSelectorSet(issueFilterSelectorSet, []string{"Folder"}, nil)
	fetchedIssueGroups, err := sys.GetProjectIssuesByIDAndFilterSetGroupedBySelector(projectVersion.ID, "", filterSet.GUID, reducedFilterSelectorSet)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Failed to fetch project version issue groups with filter set %v and selector %v for project version ID %v", filterSet, issueFilterSelectorSet, projectVersion.ID)
	}
	overallViolations := 0
	for _, issueGroup := range fetchedIssueGroups {
		overallViolations += getIssueDeltaFor(config, sys, issueGroup, projectVersion.ID, filterSet, issueFilterSelectorSet, influx)
	}
	return overallViolations
}

func getIssueDeltaFor(config fortifyExecuteScanOptions, sys fortify.System, issueGroup *models.ProjectVersionIssueGroup, projectVersionID int64, filterSet *models.FilterSet, issueFilterSelectorSet *models.IssueFilterSelectorSet, influx *fortifyExecuteScanInflux) int {
	totalMinusAuditedDelta := 0
	group := ""
	total := 0
	audited := 0
	if issueGroup != nil {
		group = fmt.Sprintf("%v", issueGroup.ID)
		total = int(*issueGroup.TotalCount)
		audited = int(*issueGroup.AuditedCount)
	}
	if group != "Optional" {
		groupTotalMinusAuditedDelta := total - audited
		if groupTotalMinusAuditedDelta > 0 {
			switch group {
			case "Corporate Security Requirements":
				auditStatus[group] = fmt.Sprintf("%v total : %v audited", total, audited)
				influx.fortify_data.fields.corporateTotal = fmt.Sprintf("%v", total)
				influx.fortify_data.fields.corporateAudited = fmt.Sprintf("%v", audited)
				totalMinusAuditedDelta += groupTotalMinusAuditedDelta

				reducedFilterSelectorSet := sys.ReduceIssueFilterSelectorSet(issueFilterSelectorSet, []string{"Folder", "Analysis"}, []string{"Corporate Security Requirements"})
				log.Entry().Errorf("[projectVersionId %v]: Unaudited corporate issues detected, count %v", projectVersionID, totalMinusAuditedDelta)
				log.Entry().Errorf("%v/html/ssc/index.jsp#!/version/%v/fix?issueFilters=%v_%v:%v&issueFilters=%v_%v:", config.ServerURL, projectVersionID, reducedFilterSelectorSet.FilterBySet[0].EntityType, reducedFilterSelectorSet.FilterBySet[0].GUID, reducedFilterSelectorSet.FilterBySet[0].SelectorOptions[0].GUID, reducedFilterSelectorSet.FilterBySet[1].EntityType, reducedFilterSelectorSet.FilterBySet[1].GUID)
				break
			case "Audit All":
				auditStatus[group] = fmt.Sprintf("%v total : %v audited", total, audited)
				influx.fortify_data.fields.auditAllTotal = fmt.Sprintf("%v", total)
				influx.fortify_data.fields.auditAllAudited = fmt.Sprintf("%v", audited)
				totalMinusAuditedDelta += groupTotalMinusAuditedDelta

				reducedFilterSelectorSet := sys.ReduceIssueFilterSelectorSet(issueFilterSelectorSet, []string{"Folder", "Analysis"}, []string{"Audit All"})
				log.Entry().Errorf("[projectVersionId %v]: Unaudited audit all issues detected, count %v", projectVersionID, totalMinusAuditedDelta)
				log.Entry().Errorf("%v/html/ssc/index.jsp#!/version/%v/fix?issueFilters=%v_%v:%v&issueFilters=%v_%v:", config.ServerURL, projectVersionID, reducedFilterSelectorSet.FilterBySet[0].EntityType, reducedFilterSelectorSet.FilterBySet[0].GUID, reducedFilterSelectorSet.FilterBySet[0].SelectorOptions[0].GUID, reducedFilterSelectorSet.FilterBySet[1].EntityType, reducedFilterSelectorSet.FilterBySet[1].GUID)
				break

			case "Spot Checks of Each Category":
				log.Entry().Info("Analyzing spot check issues")
				reducedFilterSelectorSet := sys.ReduceIssueFilterSelectorSet(issueFilterSelectorSet, []string{"Folder", "Analysis"}, []string{"Spot Checks of Each Category"})
				filter := fmt.Sprintf("%v:%v", reducedFilterSelectorSet.FilterBySet[0].EntityType, reducedFilterSelectorSet.FilterBySet[0].SelectorOptions[0].GUID)
				fetchedIssueGroups, err := sys.GetProjectIssuesByIDAndFilterSetGroupedBySelector(projectVersionID, filter, filterSet.GUID, sys.ReduceIssueFilterSelectorSet(issueFilterSelectorSet, []string{"Category"}, nil))
				if err != nil {
					log.Entry().WithError(err).Fatalf("Failed to fetch project version issue groups with filter %v, filter set %v and selector %v for project version ID %v", filter, filterSet, issueFilterSelectorSet, projectVersionID)
				}
				totalMinusAuditedDelta += getSpotIssueCount(config, fetchedIssueGroups, projectVersionID, filterSet, reducedFilterSelectorSet, influx)
				break
			}
		}
	}
	return totalMinusAuditedDelta
}

func getSpotIssueCount(config fortifyExecuteScanOptions, spotCheckCategories []*models.ProjectVersionIssueGroup, projectVersionID int64, filterSet *models.FilterSet, issueFilterSelectorSet *models.IssueFilterSelectorSet, influx *fortifyExecuteScanInflux) int {
	overallDelta := 0
	overallIssues := 0
	overallIssuesAudited := 0
	for _, issueGroup := range spotCheckCategories {
		group := ""
		total := 0
		audited := 0
		if issueGroup != nil {
			group = fmt.Sprintf("%v", issueGroup.ID)
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
				overallDelta += currentDelta
				log.Entry().Errorf("[projectVersionId %v]: %v unaudited spot check issues detected in group %v", projectVersionID, currentDelta, group)
				log.Entry().Errorf("%v/html/ssc/index.jsp#!/version/%v/fix?issueFilters=%v_%v:%v&issueFilters=%v_%v:", config.ServerURL, projectVersionID, issueFilterSelectorSet.FilterBySet[0].EntityType, issueFilterSelectorSet.FilterBySet[0].GUID, issueFilterSelectorSet.FilterBySet[0].SelectorOptions[0].GUID, issueFilterSelectorSet.FilterBySet[1].EntityType, issueFilterSelectorSet.FilterBySet[1].GUID)
				flagOutput = checkString
			}
		}

		auditStatus[group] = fmt.Sprintf("%v total : %v audited %v", total, audited, flagOutput)
	}

	influx.fortify_data.fields.spotChecksTotal = fmt.Sprintf("%v", overallIssues)
	influx.fortify_data.fields.spotChecksAudited = fmt.Sprintf("%v", overallIssuesAudited)

	return overallDelta
}

func analyseSuspiciousExploitable(config fortifyExecuteScanOptions, sys fortify.System, projectVersion *models.ProjectVersion, filterSet *models.FilterSet, issueFilterSelectorSet *models.IssueFilterSelectorSet, influx *fortifyExecuteScanInflux) int {
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
		log.Entry().Errorf("%v/html/ssc/index.jsp#!/version/%v/fix?issueGrouping=%v_%v&issueFilters=%v_%v", config.ServerURL, projectVersion.ID, reducedFilterSelectorSet.GroupBySet[0].EntityType, reducedFilterSelectorSet.GroupBySet[0].GUID, reducedFilterSelectorSet.FilterBySet[0].EntityType, reducedFilterSelectorSet.FilterBySet[0].GUID)
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

func generateAndDownloadQGateReport(config fortifyExecuteScanOptions, sys fortify.System, project *models.Project, projectVersion *models.ProjectVersion) {
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
			log.Entry().WithError(err).Fatal("Failed to fetch Q-Gate report generation status")
		}
		status = report.Status
	}
	data, err := sys.DownloadReportFile(config.ReportDownloadEndpoint, projectVersion.ID)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to download Q-Gate Report")
	}
	ioutil.WriteFile(fmt.Sprintf("%vtarget/%v-%v.%v", config.ModulePath, *project.Name, *projectVersion.Name, config.ReportType), data, 0x700)
}

func checkArtifactStatus(config fortifyExecuteScanOptions, sys fortify.System, projectVersionID int64, buildLabel string, filterSet *models.FilterSet, artifact *models.Artifact, numInvokes int) bool {
	numInvokes++
	if "PROCESSING" == artifact.Status || "SCHED_PROCESSING" == artifact.Status {
		if numInvokes >= (config.PollingMinutes * 6) {
			log.Entry().Fatalf("Terminating after %v minutes since artifact for Project Version %v is still in status %v", config.PollingMinutes, projectVersionID, artifact.Status)
		}
		log.Entry().Infof("Most recent artifact uploaded on %v of Project Version %v is still in status %v...", artifact.UploadDate, projectVersionID, artifact.Status)
		time.Sleep(10 * time.Second)
		verifyScanResultsFinishedUploading(config, sys, projectVersionID, buildLabel, filterSet, numInvokes)
		return true
	}
	if "REQUIRE_AUTH" == artifact.Status {
		// verify no manual issue approval needed
		log.Entry().Warnf("There are artifacts that require manual approval for Project Version %v\n%v/html/ssc/index.jsp#!/version/%v/artifacts?filterSet=%v", projectVersionID, config.ServerURL, projectVersionID, filterSet.GUID)
	}
	if "ERROR_PROCESSING" == artifact.Status {
		log.Entry().Warnf("There are artifacts that failed processing for Project Version %v\n%v/html/ssc/index.jsp#!/version/%v/artifacts?filterSet=%v", projectVersionID, config.ServerURL, projectVersionID, filterSet.GUID)
	}
	return false
}

func verifyScanResultsFinishedUploading(config fortifyExecuteScanOptions, sys fortify.System, projectVersionID int64, buildLabel string, filterSet *models.FilterSet, numInvokes int) {
	log.Entry().Debug("Verifying scan results have finished uploading and processing")
	var artifacts []*models.Artifact
	var relatedUpload *models.Artifact
	for relatedUpload == nil {
		artifacts, err := sys.GetArtifactsOfProjectVersion(projectVersionID)
		log.Entry().Debugf("Recieved %v artifacts for project version ID %v", len(artifacts), projectVersionID)
		if err != nil {
			log.Entry().WithError(err).Fatalf("Failed to fetch artifacts of project version ID %v", projectVersionID)
		}
		if len(artifacts) > 0 {
			latest := artifacts[0]
			if checkArtifactStatus(config, sys, projectVersionID, buildLabel, filterSet, latest, numInvokes) {
				return
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
			log.Entry().Fatalf("No uploaded artifacts for assessment detected for project version with ID %v", projectVersionID)
		}
		if relatedUpload == nil {
			log.Entry().Warn("Unable to identify artifact based on the build label, will consider most recent artifact as related to the scan")
			relatedUpload = artifacts[0]
		}
	}

	differenceInSeconds := calculateTimeDifferenceToLastUpload(config, relatedUpload.UploadDate, projectVersionID)
	// Use the absolute value for checking the time difference
	if differenceInSeconds > float64(60*config.DeltaMinutes) {
		log.Entry().Fatal("No recent upload detected on Project Version")
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
}

func calculateTimeDifferenceToLastUpload(config fortifyExecuteScanOptions, uploadDate models.Iso8601MilliDateTime, projectVersionID int64) float64 {
	log.Entry().Infof("Last upload on project version %v happened on %v", projectVersionID, uploadDate)
	uploadDateAsTime := time.Time(uploadDate)
	duration := time.Since(uploadDateAsTime)
	log.Entry().Debugf("Difference duration is %v", duration)
	absoluteSeconds := math.Abs(duration.Seconds())
	log.Entry().Infof("Difference since %v in seconds is %v", uploadDateAsTime, absoluteSeconds)
	return absoluteSeconds
}

func triggerFortifyScan(config fortifyExecuteScanOptions, command execRunner, buildID, buildLabel string) {
	if config.ScanType == "pip" {
		// Do special Python related prep
		pipVersion := "pip3"
		if config.PythonVersion != "python3" {
			pipVersion = "pip2"
		}
		installCommand, err := piperutils.ExecuteTemplate(config.PythonInstallCommand, map[string]string{"Pip": pipVersion})
		if err != nil {
			log.Entry().WithError(err).Fatalf("Failed to execute template for PythonInstallCommand: %v", config.PythonInstallCommand)
		}
		installCommandTokens := tokenize(installCommand)
		err = command.RunExecutable(installCommandTokens[0], installCommandTokens[1:]...)
		if err != nil {
			log.Entry().WithError(err).WithField("command", config.PythonInstallCommand).Fatal("Failed to execute python install command")
		}

		if len(config.Translate) == 0 {
			buf := new(bytes.Buffer)
			command.Stdout(buf)
			err := command.RunExecutable(config.PythonVersion, "-c", "import sys;p=sys.path;p.remove('');print(';'.join(p))")
			command.Stdout(log.Entry().Writer())

			config.Translate = `[{"pythonPath":"`
			if err == nil {
				config.Translate += strings.TrimSpace(buf.String())
				config.Translate += ";"
			}
			config.Translate += config.PythonAdditionalPath
			config.Translate += `","pythonIncludes":"`
			config.Translate += config.PythonIncludes
			config.Translate += `","pythonExcludes":"`
			config.Translate += strings.ReplaceAll(config.PythonExcludes, "-exclude ", "")
			config.Translate += `"}]`
		}
	}

	translateProject(config, command, buildID)

	scanProject(config, command, buildID, buildLabel)
}

func translateProject(config fortifyExecuteScanOptions, command execRunner, buildID string) {
	log.Entry().Debugf("Translate options are %v", config.Translate)
	var translateList []map[string]string
	json.Unmarshal([]byte(config.Translate), &translateList)
	for _, translate := range translateList {
		handleSingleTranslate(config, command, buildID, translate)
	}
}

func handleSingleTranslate(config fortifyExecuteScanOptions, command execRunner, buildID string, t map[string]string) {
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

func scanProject(config fortifyExecuteScanOptions, command execRunner, buildID, buildLabel string) {
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
	scanOptions = append(scanOptions, "-logfile", "target/fortify-scan.log", "-f", "target/result.fpr")

	err := command.RunExecutable("sourceanalyzer", scanOptions...)
	if err != nil {
		log.Entry().WithError(err).WithField("scanOptions", scanOptions).Fatal("failed to execute sourceanalyzer scan command")
	}
}

func determinePullRequestMerge(config fortifyExecuteScanOptions) string {
	log.Entry().Debugf("Retrieved commit message %v", config.CommitMessage)
	r, _ := regexp.Compile(config.PullRequestMessageRegex)
	matches := r.FindSubmatch([]byte(config.CommitMessage))
	if matches != nil && len(matches) > 1 {
		return string(matches[config.PullRequestMessageRegexGroup])
	}
	return ""
}

func appendToOptions(config fortifyExecuteScanOptions, options []string, t map[string]string) []string {
	if config.ScanType == "windows" {
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
		return append(options, t["src"])
	}
	if config.ScanType == "java" {
		if len(t["classpath"]) > 0 {
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
		return append(options, t["src"])
	}
	if config.ScanType == "pip" {
		if len(t["pythonPath"]) > 0 {
			options = append(options, "-python-path", t["pythonPath"])
		}
		if len(t["pythonExcludes"]) > 0 {
			options = append(options, "-exclude", t["pythonExcludes"])
		}
		return append(options, t["pythonIncludes"])
	}
	return options
}
