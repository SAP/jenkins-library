package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"

	"github.com/bmatcuk/doublestar"

	"github.com/google/go-github/v68/github"
	"github.com/google/uuid"

	"github.com/piper-validation/fortify-client-go/models"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/fortify"
	"github.com/SAP/jenkins-library/pkg/gradle"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/reporting"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/toolrecord"
	"github.com/SAP/jenkins-library/pkg/versioning"

	piperGithub "github.com/SAP/jenkins-library/pkg/github"

	"github.com/pkg/errors"
)

const getClasspathScriptContent = `
gradle.allprojects {
    task getClasspath {
        doLast {
            new File(projectDir, filename).text = sourceSets.main.compileClasspath.asPath
        }
    }
}
`

type pullRequestService interface {
	ListPullRequestsWithCommit(ctx context.Context, owner, repo, sha string, opts *github.ListOptions) ([]*github.PullRequest, *github.Response, error)
}

type fortifyUtils interface {
	maven.Utils
	gradle.Utils
	piperutils.FileUtils

	SetDir(d string)
	GetArtifact(buildTool, buildDescriptorFile string, options *versioning.Options) (versioning.Artifact, error)
	GetIssueService() *github.IssuesService
	GetSearchService() *github.SearchService
}

type fortifyUtilsBundle struct {
	*command.Command
	*piperutils.Files
	*piperhttp.Client
	issues *github.IssuesService
	search *github.SearchService
}

func (f *fortifyUtilsBundle) GetArtifact(buildTool, buildDescriptorFile string, options *versioning.Options) (versioning.Artifact, error) {
	return versioning.GetArtifact(buildTool, buildDescriptorFile, options, f)
}

func (f *fortifyUtilsBundle) CreateIssue(ghCreateIssueOptions *piperGithub.CreateIssueOptions) error {
	_, err := piperGithub.CreateIssue(ghCreateIssueOptions)
	return err
}

func (f *fortifyUtilsBundle) GetIssueService() *github.IssuesService {
	return f.issues
}

func (f *fortifyUtilsBundle) GetSearchService() *github.SearchService {
	return f.search
}

func newFortifyUtilsBundle(client *github.Client) fortifyUtils {
	utils := fortifyUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
		Client:  &piperhttp.Client{},
	}
	if client != nil {
		utils.issues = client.Issues
		utils.search = client.Search
	}
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

const (
	checkString       = "<---CHECK FORTIFY---"
	classpathFileName = "fortify-execute-scan-cp.txt"
)

var execInPath = exec.LookPath

func fortifyExecuteScan(config fortifyExecuteScanOptions, telemetryData *telemetry.CustomData, influx *fortifyExecuteScanInflux) {
	// TODO provide parameter for trusted certs
	ctx, client, err := piperGithub.NewClientBuilder(config.GithubToken, config.GithubAPIURL).Build()
	if err != nil {
		log.Entry().WithError(err).Warning("Failed to get GitHub client")
	}
	auditStatus := map[string]string{}
	sys := fortify.NewSystemInstance(config.ServerURL, config.APIEndpoint, config.AuthToken, config.Proxy, time.Minute*15)
	utils := newFortifyUtilsBundle(client)

	influx.step_data.fields.fortify = false
	reports, err := runFortifyScan(ctx, config, sys, utils, telemetryData, influx, auditStatus)
	piperutils.PersistReportsAndLinks("fortifyExecuteScan", config.ModulePath, utils, reports, nil)
	if err != nil {
		log.Entry().WithError(err).Fatal("Fortify scan and check failed")
	}
	influx.step_data.fields.fortify = true
	// make sure that no specific error category is set in success case
	log.SetErrorCategory(log.ErrorUndefined)
}

func determineArtifact(config fortifyExecuteScanOptions, utils fortifyUtils) (versioning.Artifact, error) {
	versioningOptions := versioning.Options{
		M2Path:              config.M2Path,
		GlobalSettingsFile:  config.GlobalSettingsFile,
		ProjectSettingsFile: config.ProjectSettingsFile,
		Defines:             config.AdditionalMvnParameters,
	}

	artifact, err := utils.GetArtifact(config.BuildTool, config.BuildDescriptorFile, &versioningOptions)
	if err != nil {
		return nil, fmt.Errorf("Unable to get artifact from descriptor %v: %w", config.BuildDescriptorFile, err)
	}
	return artifact, nil
}

func runFortifyScan(ctx context.Context, config fortifyExecuteScanOptions, sys fortify.System, utils fortifyUtils, telemetryData *telemetry.CustomData, influx *fortifyExecuteScanInflux, auditStatus map[string]string) ([]piperutils.Path, error) {
	var reports []piperutils.Path
	log.Entry().Debugf("Running Fortify scan against SSC at %v", config.ServerURL)
	executableList := []string{"fortifyupdate", "sourceanalyzer"}
	for _, exec := range executableList {
		_, err := execInPath(exec)
		if err != nil {
			return reports, fmt.Errorf("Command not found: %v. Please configure a supported docker image or install Fortify SCA on the system.", exec)
		}
	}

	if config.BuildTool == "maven" && config.InstallArtifacts {
		err := maven.InstallMavenArtifacts(&maven.EvaluateOptions{
			M2Path:              config.M2Path,
			ProjectSettingsFile: config.ProjectSettingsFile,
			GlobalSettingsFile:  config.GlobalSettingsFile,
			PomPath:             config.BuildDescriptorFile,
		}, utils)
		if err != nil {
			return reports, fmt.Errorf("Unable to install artifacts: %w", err)
		}
	}

	artifact, err := determineArtifact(config, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal()
	}
	coordinates, err := artifact.GetCoordinates()
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return reports, fmt.Errorf("unable to get project coordinates from descriptor %v: %w", config.BuildDescriptorFile, err)
	}
	log.Entry().Debugf("loaded project coordinates %v from descriptor", coordinates)

	if len(config.Version) > 0 {
		log.Entry().Infof("Resolving product version from default provided '%s' with versioning '%s'", config.Version, config.VersioningModel)
		coordinates.Version = config.Version
	}

	fortifyProjectName, fortifyProjectVersion := versioning.DetermineProjectCoordinatesWithCustomVersion(config.ProjectName, config.VersioningModel, config.CustomScanVersion, coordinates)
	project, err := sys.GetProjectByName(fortifyProjectName, config.AutoCreate, fortifyProjectVersion)
	if err != nil {
		classifyErrorOnLookup(err)
		return reports, fmt.Errorf("Failed to load project %v: %w", fortifyProjectName, err)
	}
	projectVersion, err := sys.GetProjectVersionDetailsByProjectIDAndVersionName(project.ID, fortifyProjectVersion, config.AutoCreate, fortifyProjectName)
	if err != nil {
		classifyErrorOnLookup(err)
		return reports, fmt.Errorf("Failed to load project version %v: %w", fortifyProjectVersion, err)
	}

	if len(config.PullRequestName) > 0 {
		fortifyProjectVersion = config.PullRequestName
		projectVersion, err = sys.LookupOrCreateProjectVersionDetailsForPullRequest(project.ID, projectVersion, fortifyProjectVersion)
		if err != nil {
			classifyErrorOnLookup(err)
			return reports, fmt.Errorf("Failed to lookup / create project version for pull request %v: %w", fortifyProjectVersion, err)
		}
		log.Entry().Debugf("Looked up / created project version with ID %v for PR %v", projectVersion.ID, fortifyProjectVersion)
	} else {
		prID, prAuthor := determinePullRequestMerge(config)
		if prID != "0" {
			log.Entry().Debugf("Determined PR ID '%v' for merge check", prID)
			if len(prAuthor) > 0 && !slices.Contains(config.Assignees, prAuthor) {
				log.Entry().Debugf("Determined PR Author '%v' for result assignment", prAuthor)
				config.Assignees = append(config.Assignees, prAuthor)
			} else {
				log.Entry().Debugf("Unable to determine PR Author, using assignees: %v", config.Assignees)
			}
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

	// create toolrecord file
	// tbd - how to handle verifyOnly
	toolRecordFileName, err := createToolRecordFortify(utils, "./", config, project.ID, fortifyProjectName, projectVersion.ID, fortifyProjectVersion)
	if err != nil {
		// do not fail until the framework is well established
		log.Entry().Warning("TR_FORTIFY: Failed to create toolrecord file ...", err)
	} else {
		reports = append(reports, piperutils.Path{Target: toolRecordFileName})
	}

	if config.VerifyOnly {
		log.Entry().Infof("Starting audit status check on project %v with version %v and project version ID %v", fortifyProjectName, fortifyProjectVersion, projectVersion.ID)
		paths, err := verifyFFProjectCompliance(ctx, config, utils, sys, project, projectVersion, filterSet, influx, auditStatus)
		reports = append(reports, paths...)
		return reports, err
	}

	log.Entry().Infof("Scanning and uploading to project %v with version %v and projectVersionId %v", fortifyProjectName, fortifyProjectVersion, projectVersion.ID)
	buildLabel := fmt.Sprintf("%v/repos/%v/%v/commits/%v", config.GithubAPIURL, config.Owner, config.Repository, config.CommitID)

	// Create sourceanalyzer command based on configuration
	buildID := uuid.New().String()
	utils.SetDir(config.ModulePath)
	if err := os.MkdirAll(fmt.Sprintf("%v/%v", config.ModulePath, "target"), os.ModePerm); err != nil {
		log.Entry().WithError(err).Error("failed to create directory")
	}

	if config.UpdateRulePack {

		fortifyUpdateParams := []string{"-acceptKey", "-acceptSSLCertificate", "-url", config.ServerURL}
		proxyPort, proxyHost := getProxyParams(config.Proxy)
		if proxyHost != "" && proxyPort != "" {
			fortifyUpdateParams = append(fortifyUpdateParams, "-proxyhost", proxyHost, "-proxyport", proxyPort)
		}

		err := utils.RunExecutable("fortifyupdate", fortifyUpdateParams...)
		if err != nil {
			return reports, fmt.Errorf("failed to update rule pack, serverUrl: %v", config.ServerURL)
		}

		err = utils.RunExecutable("fortifyupdate", "-acceptKey", "-acceptSSLCertificate", "-showInstalledRules")
		if err != nil {
			return reports, fmt.Errorf("failed to fetch details of installed rule pack, serverUrl: %v", config.ServerURL)
		}
	}

	err = triggerFortifyScan(config, utils, buildID, buildLabel, fortifyProjectName)
	reports = append(reports, piperutils.Path{Target: fmt.Sprintf("%vtarget/fortify-scan.*", config.ModulePath)})
	reports = append(reports, piperutils.Path{Target: fmt.Sprintf("%vtarget/*.fpr", config.ModulePath)})
	if err != nil {
		return reports, errors.Wrap(err, "failed to scan project")
	}

	var message string
	if config.UploadResults {
		log.Entry().Debug("Uploading results")
		resultFilePath := fmt.Sprintf("%vtarget/result.fpr", config.ModulePath)
		err = sys.UploadResultFile(config.FprUploadEndpoint, resultFilePath, projectVersion.ID)
		message = fmt.Sprintf("Failed to upload result file %v to Fortify SSC at %v", resultFilePath, config.ServerURL)
	} else {
		log.Entry().Debug("Generating XML report")
		xmlReportName := "fortify_result.xml"
		err = utils.RunExecutable("ReportGenerator", "-format", "xml", "-f", xmlReportName, "-source", fmt.Sprintf("%vtarget/result.fpr", config.ModulePath))
		message = fmt.Sprintf("Failed to generate XML report %v", xmlReportName)
		if err != nil {
			reports = append(reports, piperutils.Path{Target: fmt.Sprintf("%vfortify_result.xml", config.ModulePath)})
		}
	}
	if err != nil {
		return reports, fmt.Errorf(message+": %w", err)
	}

	log.Entry().Infof("Ensuring latest FPR is processed for project %v with version %v and project version ID %v", fortifyProjectName, fortifyProjectVersion, projectVersion.ID)
	// Ensure latest FPR is processed
	err = verifyScanResultsFinishedUploading(config, sys, projectVersion.ID, buildLabel, filterSet,
		10*time.Second, time.Duration(config.PollingMinutes)*time.Minute)
	if err != nil {
		return reports, err
	}

	// SARIF conversion done after latest FPR is processed, but before the compliance is checked
	if config.ConvertToSarif {
		resultFilePath := fmt.Sprintf("%vtarget/result.fpr", config.ModulePath)
		log.Entry().Info("Calling conversion to SARIF function.")
		sarif, sarifSimplified, err := fortify.ConvertFprToSarif(sys, projectVersion, resultFilePath, filterSet)
		if err != nil {
			return reports, fmt.Errorf("failed to generate SARIF")
		}
		log.Entry().Debug("Writing simplified sarif file in plain text to disk.")
		paths, err := fortify.WriteSarif(sarifSimplified, "result.sarif")
		if err != nil {
			return reports, fmt.Errorf("failed to write simplified sarif")
		}
		reports = append(reports, paths...)

		log.Entry().Debug("Writing full sarif file to disk and gzip it.")
		paths, err = fortify.WriteGzipSarif(sarif, "result.sarif.gz")
		if err != nil {
			return reports, fmt.Errorf("failed to write gzip sarif")
		}
		reports = append(reports, paths...)
	}

	log.Entry().Infof("Starting audit status check on project %v with version %v and project version ID %v", fortifyProjectName, fortifyProjectVersion, projectVersion.ID)
	paths, err := verifyFFProjectCompliance(ctx, config, utils, sys, project, projectVersion, filterSet, influx, auditStatus)
	reports = append(reports, paths...)
	return reports, err
}

func classifyErrorOnLookup(err error) {
	if strings.Contains(err.Error(), "connect: connection refused") || strings.Contains(err.Error(), "net/http: TLS handshake timeout") {
		log.SetErrorCategory(log.ErrorService)
	}
}

func verifyFFProjectCompliance(ctx context.Context, config fortifyExecuteScanOptions, utils fortifyUtils, sys fortify.System, project *models.Project, projectVersion *models.ProjectVersion, filterSet *models.FilterSet, influx *fortifyExecuteScanInflux, auditStatus map[string]string) ([]piperutils.Path, error) {
	reports := []piperutils.Path{}
	// Generate report
	if config.Reporting {
		resultURL := []byte(fmt.Sprintf("%v/html/ssc/version/%v/fix/null/", config.ServerURL, projectVersion.ID))
		if err := os.WriteFile(fmt.Sprintf("%vtarget/%v-%v.%v", config.ModulePath, *project.Name, *projectVersion.Name, "txt"), resultURL, 0o700); err != nil {
			log.Entry().WithError(err).Error("failed to write file")
		}

		data, err := generateAndDownloadQGateReport(config, sys, project, projectVersion)
		if err != nil {
			return reports, err
		}
		if err := os.WriteFile(fmt.Sprintf("%vtarget/%v-%v.%v", config.ModulePath, *project.Name, *projectVersion.Name, config.ReportType), data, 0o700); err != nil {
			log.Entry().WithError(err).Warning("failed to write file")
		}
	}

	// Perform audit compliance checks
	issueFilterSelectorSet, err := sys.GetIssueFilterSelectorOfProjectVersionByName(projectVersion.ID, []string{"Analysis", "Folder", "Category"}, nil)
	if err != nil {
		return reports, errors.Wrapf(err, "failed to fetch project version issue filter selector for project version ID %v", projectVersion.ID)
	}
	log.Entry().Debugf("initial filter selector set: %v", issueFilterSelectorSet)

	spotChecksCountByCategory := []fortify.SpotChecksAuditCount{}
	numberOfViolations, issueGroups, err := analyseUnauditedIssues(config, sys, projectVersion, filterSet, issueFilterSelectorSet, influx, auditStatus, &spotChecksCountByCategory)
	if err != nil {
		return reports, errors.Wrap(err, "failed to analyze unaudited issues")
	}
	numberOfSuspiciousExploitable, issueGroupsSuspiciousExploitable := analyseSuspiciousExploitable(config, sys, projectVersion, filterSet, issueFilterSelectorSet, influx, auditStatus)
	numberOfViolations += numberOfSuspiciousExploitable
	issueGroups = append(issueGroups, issueGroupsSuspiciousExploitable...)

	log.Entry().Infof("Counted %v violations, details: %v", numberOfViolations, auditStatus)

	influx.fortify_data.fields.projectID = project.ID
	influx.fortify_data.fields.projectName = *project.Name
	influx.fortify_data.fields.projectVersion = *projectVersion.Name
	influx.fortify_data.fields.projectVersionID = projectVersion.ID
	influx.fortify_data.fields.violations = numberOfViolations

	fortifyReportingData := prepareReportData(influx)
	scanReport := fortify.CreateCustomReport(fortifyReportingData, issueGroups)
	paths, err := fortify.WriteCustomReports(scanReport)
	if err != nil {
		return reports, errors.Wrap(err, "failed to write custom reports")
	}
	reports = append(reports, paths...)

	log.Entry().Debug("Checking whether GitHub issue creation/update is active")
	log.Entry().Debugf("%v, %v, %v, %v, %v, %v", config.CreateResultIssue, numberOfViolations > 0, len(config.GithubToken) > 0, len(config.GithubAPIURL) > 0, len(config.Owner) > 0, len(config.Repository) > 0)
	if config.CreateResultIssue && numberOfViolations > 0 && len(config.GithubToken) > 0 && len(config.GithubAPIURL) > 0 && len(config.Owner) > 0 && len(config.Repository) > 0 {
		log.Entry().Debug("Creating/updating GitHub issue with scan results")
		gh := reporting.GitHub{
			Owner:         &config.Owner,
			Repository:    &config.Repository,
			Assignees:     &config.Assignees,
			IssueService:  utils.GetIssueService(),
			SearchService: utils.GetSearchService(),
		}
		if err := gh.UploadSingleReport(ctx, scanReport); err != nil {
			return reports, fmt.Errorf("failed to upload scan results into GitHub: %w", err)
		}
	}

	jsonReport := fortify.CreateJSONReport(fortifyReportingData, spotChecksCountByCategory, config.ServerURL)
	paths, err = fortify.WriteJSONReport(jsonReport)
	if err != nil {
		return reports, errors.Wrap(err, "failed to write json report")
	}
	reports = append(reports, paths...)

	if numberOfViolations > 0 {
		log.SetErrorCategory(log.ErrorCompliance)
		return reports, errors.New("fortify scan failed, the project is not compliant. For details check the archived report")
	}
	return reports, nil
}

func prepareReportData(influx *fortifyExecuteScanInflux) fortify.FortifyReportData {
	input := influx.fortify_data.fields
	output := fortify.FortifyReportData{}
	output.ProjectID = input.projectID
	output.ProjectName = input.projectName
	output.ProjectVersion = input.projectVersion
	output.AuditAllAudited = input.auditAllAudited
	output.AuditAllTotal = input.auditAllTotal
	output.CorporateAudited = input.corporateAudited
	output.CorporateTotal = input.corporateTotal
	output.SpotChecksAudited = input.spotChecksAudited
	output.SpotChecksGap = input.spotChecksGap
	output.SpotChecksTotal = input.spotChecksTotal
	output.Exploitable = input.exploitable
	output.Suppressed = input.suppressed
	output.Suspicious = input.suspicious
	output.ProjectVersionID = input.projectVersionID
	output.Violations = input.violations
	return output
}

func analyseUnauditedIssues(config fortifyExecuteScanOptions, sys fortify.System, projectVersion *models.ProjectVersion, filterSet *models.FilterSet, issueFilterSelectorSet *models.IssueFilterSelectorSet, influx *fortifyExecuteScanInflux, auditStatus map[string]string, spotChecksCountByCategory *[]fortify.SpotChecksAuditCount) (int, []*models.ProjectVersionIssueGroup, error) {
	log.Entry().Info("Analyzing unaudited issues")

	if config.SpotCheckMinimumUnit != "percentage" && config.SpotCheckMinimumUnit != "number" {
		return 0, nil, fmt.Errorf("Invalid spotCheckMinimumUnit. Please set it as 'percentage' or 'number'.")
	}

	reducedFilterSelectorSet := sys.ReduceIssueFilterSelectorSet(issueFilterSelectorSet, []string{"Folder"}, nil)
	fetchedIssueGroups, err := sys.GetProjectIssuesByIDAndFilterSetGroupedBySelector(projectVersion.ID, "", filterSet.GUID, reducedFilterSelectorSet)
	if err != nil {
		return 0, fetchedIssueGroups, errors.Wrapf(err, "failed to fetch project version issue groups with filter set %v and selector %v for project version ID %v", filterSet, issueFilterSelectorSet, projectVersion.ID)
	}
	overallViolations := 0
	for _, issueGroup := range fetchedIssueGroups {
		issueDelta, err := getIssueDeltaFor(config, sys, issueGroup, projectVersion.ID, filterSet, issueFilterSelectorSet, influx, auditStatus, spotChecksCountByCategory)
		if err != nil {
			return overallViolations, fetchedIssueGroups, errors.Wrap(err, "failed to get issue delta")
		}
		overallViolations += issueDelta
	}
	return overallViolations, fetchedIssueGroups, nil
}

func getIssueDeltaFor(config fortifyExecuteScanOptions, sys fortify.System, issueGroup *models.ProjectVersionIssueGroup, projectVersionID int64, filterSet *models.FilterSet, issueFilterSelectorSet *models.IssueFilterSelectorSet, influx *fortifyExecuteScanInflux, auditStatus map[string]string, spotChecksCountByCategory *[]fortify.SpotChecksAuditCount) (int, error) {
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
			return totalMinusAuditedDelta, fmt.Errorf("folder selector not found")
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
				return totalMinusAuditedDelta, errors.Wrapf(err, "failed to fetch project version issue groups with filter %v, filter set %v and selector %v for project version ID %v", filter, filterSet, issueFilterSelectorSet, projectVersionID)
			}
			totalMinusAuditedDelta += getSpotIssueCount(config, sys, fetchedIssueGroups, projectVersionID, filterSet, reducedFilterSelectorSet, influx, auditStatus, spotChecksCountByCategory)
		}
	}
	return totalMinusAuditedDelta, nil
}

func getSpotIssueCount(config fortifyExecuteScanOptions, sys fortify.System, spotCheckCategories []*models.ProjectVersionIssueGroup, projectVersionID int64, filterSet *models.FilterSet, issueFilterSelectorSet *models.IssueFilterSelectorSet, influx *fortifyExecuteScanInflux, auditStatus map[string]string, spotChecksCountByCategory *[]fortify.SpotChecksAuditCount) int {
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

		minSpotChecksPerCategory := getMinSpotChecksPerCategory(config, total)
		log.Entry().Debugf("Minimum spot checks for group %v is %v with audit count %v and total issue count %v", group, minSpotChecksPerCategory, audited, total)

		if ((total <= minSpotChecksPerCategory || minSpotChecksPerCategory < 0) && audited != total) || (total > minSpotChecksPerCategory && audited < minSpotChecksPerCategory) {
			currentDelta := minSpotChecksPerCategory - audited
			if minSpotChecksPerCategory < 0 || minSpotChecksPerCategory > total {
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
		*spotChecksCountByCategory = append(*spotChecksCountByCategory, fortify.SpotChecksAuditCount{Audited: audited, Total: total, Type: group})
	}

	influx.fortify_data.fields.spotChecksTotal = overallIssues
	influx.fortify_data.fields.spotChecksAudited = overallIssuesAudited
	influx.fortify_data.fields.spotChecksGap = overallDelta

	return overallDelta
}

func getMinSpotChecksPerCategory(config fortifyExecuteScanOptions, totalCount int) int {
	if config.SpotCheckMinimumUnit == "percentage" {
		spotCheckMinimumPercentageValue := int(math.Ceil(float64(config.SpotCheckMinimum) / 100.0 * float64(totalCount)))
		return getSpotChecksMinAsPerMaximum(config.SpotCheckMaximum, spotCheckMinimumPercentageValue)
	}

	return getSpotChecksMinAsPerMaximum(config.SpotCheckMaximum, config.SpotCheckMinimum)
}

func getSpotChecksMinAsPerMaximum(spotCheckMax int, spotCheckMin int) int {
	if spotCheckMax < 1 {
		return spotCheckMin
	}

	if spotCheckMin > spotCheckMax {
		return spotCheckMax
	}

	return spotCheckMin
}

func analyseSuspiciousExploitable(config fortifyExecuteScanOptions, sys fortify.System, projectVersion *models.ProjectVersion, filterSet *models.FilterSet, issueFilterSelectorSet *models.IssueFilterSelectorSet, influx *fortifyExecuteScanInflux, auditStatus map[string]string) (int, []*models.ProjectVersionIssueGroup) {
	log.Entry().Info("Analyzing suspicious and exploitable issues")
	reducedFilterSelectorSet := sys.ReduceIssueFilterSelectorSet(issueFilterSelectorSet, []string{"Analysis"}, []string{})
	fetchedGroups, err := sys.GetProjectIssuesByIDAndFilterSetGroupedBySelector(projectVersion.ID, "", filterSet.GUID, reducedFilterSelectorSet)
	if err != nil {
		log.Entry().WithError(err).Errorf("failed to get project issues")
	}

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

	return result, fetchedGroups
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
		return []byte{}, errors.Wrap(err, "failed to generate Q-Gate report")
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
			log.SetErrorCategory(log.ErrorService)
			return fmt.Errorf("terminating after %v since artifact for Project Version %v is still in status %v", timeout, projectVersionID, artifact.Status)
		}
		log.Entry().Infof("Most recent artifact uploaded on %v of Project Version %v is still in status %v...", artifact.UploadDate, projectVersionID, artifact.Status)
		time.Sleep(pollingDelay)
		return errProcessing
	}
	if "REQUIRE_AUTH" == artifact.Status {
		// verify no manual issue approval needed
		log.SetErrorCategory(log.ErrorCompliance)
		return fmt.Errorf("There are artifacts that require manual approval for Project Version %v, please visit Fortify SSC and approve them for processing\n%v/html/ssc/index.jsp#!/version/%v/artifacts?filterSet=%v", projectVersionID, config.ServerURL, projectVersionID, filterSet.GUID)
	}
	if "ERROR_PROCESSING" == artifact.Status {
		log.SetErrorCategory(log.ErrorService)
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
			return fmt.Errorf("failed to fetch artifacts of project version ID %v: %w", projectVersionID, err)
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

func executeTemplatedCommand(utils fortifyUtils, cmdTemplate []string, context map[string]string) error {
	for index, cmdTemplatePart := range cmdTemplate {
		result, err := piperutils.ExecuteTemplate(cmdTemplatePart, context)
		if err != nil {
			return errors.Wrapf(err, "failed to transform template for command fragment: %v", cmdTemplatePart)
		}
		cmdTemplate[index] = result
	}
	err := utils.RunExecutable(cmdTemplate[0], cmdTemplate[1:]...)
	if err != nil {
		return errors.Wrapf(err, "failed to execute command %v", cmdTemplate)
	}
	return nil
}

func autoresolvePipClasspath(executable string, parameters []string, file string, utils fortifyUtils) (string, error) {
	// redirect stdout and create cp file from command output
	outfile, err := os.Create(file)
	if err != nil {
		return "", errors.Wrap(err, "failed to create classpath file")
	}
	defer outfile.Close()
	utils.Stdout(outfile)
	err = utils.RunExecutable(executable, parameters...)
	if err != nil {
		return "", errors.Wrapf(err, "failed to run classpath autodetection command %v with parameters %v", executable, parameters)
	}
	utils.Stdout(log.Entry().Writer())
	return readClasspathFile(file), nil
}

func autoresolveMavenClasspath(config fortifyExecuteScanOptions, file string, utils fortifyUtils) (string, error) {
	if filepath.IsAbs(file) {
		log.Entry().Warnf("Passing an absolute path for -Dmdep.outputFile results in the classpath only for the last module in multi-module maven projects.")
	}
	defines := generateMavenFortifyDefines(&config, file)
	executeOptions := maven.ExecuteOptions{
		PomPath:             config.BuildDescriptorFile,
		ProjectSettingsFile: config.ProjectSettingsFile,
		GlobalSettingsFile:  config.GlobalSettingsFile,
		M2Path:              config.M2Path,
		Goals:               []string{"dependency:build-classpath", "package"},
		Defines:             defines,
		ReturnStdout:        false,
	}
	_, err := maven.Execute(&executeOptions, utils)
	if err != nil {
		log.Entry().WithError(err).Warnf("failed to determine classpath using Maven: %v", err)
	}
	return readAllClasspathFiles(file), nil
}

func autoresolveGradleClasspath(config fortifyExecuteScanOptions, file string, utils fortifyUtils) (string, error) {
	gradleOptions := &gradle.ExecuteOptions{
		Task:              "getClasspath",
		UseWrapper:        true,
		InitScriptContent: getClasspathScriptContent,
		ProjectProperties: map[string]string{"filename": file},
	}
	if _, err := gradle.Execute(gradleOptions, utils); err != nil {
		log.Entry().WithError(err).Warnf("failed to determine classpath using Gradle: %v", err)
	}
	return readAllClasspathFiles(file), nil
}

func generateMavenFortifyDefines(config *fortifyExecuteScanOptions, file string) []string {
	defines := []string{
		fmt.Sprintf("-Dmdep.outputFile=%v", file),
		// Parameter to indicate to maven build that the fortify step is the trigger, can be used for optimizations
		"-Dfortify",
		"-DincludeScope=compile",
		"-DskipTests",
		"-Dmaven.javadoc.skip=true",
		"--fail-at-end",
	}

	if len(config.BuildDescriptorExcludeList) > 0 {
		// From the documentation, these are file paths to a module's pom.xml.
		// For MTA projects, we support pom.xml files here and skip others.
		for _, exclude := range config.BuildDescriptorExcludeList {
			if !strings.HasSuffix(exclude, "pom.xml") {
				continue
			}
			exists, _ := piperutils.FileExists(exclude)
			if !exists {
				continue
			}
			moduleName := filepath.Dir(exclude)
			if moduleName != "" {
				defines = append(defines, "-pl", "!"+moduleName)
			}
		}
	}

	return defines
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
	data, err := os.ReadFile(file)
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

func triggerFortifyScan(config fortifyExecuteScanOptions, utils fortifyUtils, buildID, buildLabel, buildProject string) error {
	var err error
	// Do special Python related prep
	pipVersion := "pip3"
	if config.PythonVersion != "python3" {
		pipVersion = "pip2"
	}

	classpath := ""
	if config.BuildTool == "maven" {
		if config.AutodetectClasspath {
			classpath, err = autoresolveMavenClasspath(config, classpathFileName, utils)
			if err != nil {
				return err
			}
		}
		config.Translate, err = populateMavenGradleTranslate(&config, classpath)
		if err != nil {
			log.Entry().WithError(err).Warnf("failed to apply src ('%s') or exclude ('%s') parameter", config.Src, config.Exclude)
		}
	} else if config.BuildTool == "gradle" {
		if config.AutodetectClasspath {
			classpath, err = autoresolveGradleClasspath(config, classpathFileName, utils)
			if err != nil {
				return err
			}
		}
		config.Translate, err = populateMavenGradleTranslate(&config, classpath)
		if err != nil {
			log.Entry().WithError(err).Warnf("failed to apply src ('%s') or exclude ('%s') parameter", config.Src, config.Exclude)
		}
	} else if config.BuildTool == "pip" {
		if config.AutodetectClasspath {
			separator := getSeparator()
			script := fmt.Sprintf("import sys;p=sys.path;p.remove('');print('%v'.join(p))", separator)
			classpath, err = autoresolvePipClasspath(config.PythonVersion, []string{"-c", script}, classpathFileName, utils)
			if err != nil {
				return errors.Wrap(err, "failed to autoresolve pip classpath")
			}
		}
		// install the dev dependencies
		if len(config.PythonRequirementsFile) > 0 {
			context := map[string]string{}
			cmdTemplate := []string{pipVersion, "install", "--user", "-r", config.PythonRequirementsFile}
			cmdTemplate = append(cmdTemplate, tokenize(config.PythonRequirementsInstallSuffix)...)
			if err := executeTemplatedCommand(utils, cmdTemplate, context); err != nil {
				log.Entry().WithError(err).Error("failed to execute template command")
			}
		}

		if err := executeTemplatedCommand(utils, tokenize(config.PythonInstallCommand), map[string]string{"Pip": pipVersion}); err != nil {
			log.Entry().WithError(err).Error("failed to execute template command")
		}

		config.Translate, err = populatePipTranslate(&config, classpath)
		if err != nil {
			log.Entry().WithError(err).Warnf("failed to apply pythonAdditionalPath ('%s') or src ('%s') parameter", config.PythonAdditionalPath, config.Src)
		}

	} else {
		return fmt.Errorf("buildTool '%s' is not supported by this step", config.BuildTool)
	}

	err = translateProject(&config, utils, buildID, classpath)
	if err != nil {
		return err
	}

	return scanProject(&config, utils, buildID, buildLabel, buildProject)
}

func appendPythonVersionToTranslate(translateOptions map[string]interface{}, pythonVersion string) error {
	if pythonVersion == "python2" {
		translateOptions["pythonVersion"] = "2"
	} else if pythonVersion == "python3" {
		translateOptions["pythonVersion"] = "3"
	} else {
		return fmt.Errorf("Invalid pythonVersion '%s'. Possible values for pythonVersion are 'python2' and 'python3'. ", pythonVersion)
	}

	return nil
}

func populatePipTranslate(config *fortifyExecuteScanOptions, classpath string) (string, error) {
	if len(config.Translate) > 0 {
		return config.Translate, nil
	}

	var translateList []map[string]interface{}
	translateList = append(translateList, make(map[string]interface{}))
	separator := getSeparator()

	err := appendPythonVersionToTranslate(translateList[0], config.PythonVersion)
	if err != nil {
		return "", err
	}

	translateList[0]["pythonPath"] = classpath + separator +
		getSuppliedOrDefaultListAsString(config.PythonAdditionalPath, []string{}, separator)
	translateList[0]["src"] = getSuppliedOrDefaultListAsString(
		config.Src, []string{"./**/*"}, ":")
	translateList[0]["exclude"] = getSuppliedOrDefaultListAsString(
		config.Exclude, []string{"./**/tests/**/*", "./**/setup.py"}, separator)

	translateJSON, err := json.Marshal(translateList)

	return string(translateJSON), err
}

func populateMavenGradleTranslate(config *fortifyExecuteScanOptions, classpath string) (string, error) {
	if len(config.Translate) > 0 {
		return config.Translate, nil
	}

	var translateList []map[string]interface{}
	translateList = append(translateList, make(map[string]interface{}))
	translateList[0]["classpath"] = classpath

	setTranslateEntryIfNotEmpty(translateList[0], "src", ":", config.Src,
		[]string{"**/*.xml", "**/*.html", "**/*.jsp", "**/*.js", "**/src/main/resources/**/*", "**/src/main/java/**/*", "**/src/gen/java/cds/**/*", "**/target/main/java/**/*", "**/target/main/resources/**/*", "**/target/generated-sources/**/*"})

	setTranslateEntryIfNotEmpty(translateList[0], "exclude", getSeparator(), config.Exclude, []string{"**/src/test/**/*"})

	translateJSON, err := json.Marshal(translateList)

	return string(translateJSON), err
}

func translateProject(config *fortifyExecuteScanOptions, utils fortifyUtils, buildID, classpath string) error {
	var translateList []map[string]string
	json.Unmarshal([]byte(config.Translate), &translateList)
	log.Entry().Debugf("Translating with options: %v", translateList)
	for _, translate := range translateList {
		if len(classpath) > 0 {
			translate["autoClasspath"] = classpath
		}
		err := handleSingleTranslate(config, utils, buildID, translate)
		if err != nil {
			return err
		}
	}
	return nil
}

func handleSingleTranslate(config *fortifyExecuteScanOptions, command fortifyUtils, buildID string, t map[string]string) error {
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
			return errors.Wrapf(err, "failed to execute sourceanalyzer translate command with options %v", translateOptions)
		}
	} else {
		log.Entry().Debug("Skipping translate with nil value")
	}
	return nil
}

func scanProject(config *fortifyExecuteScanOptions, command fortifyUtils, buildID, buildLabel, buildProject string) error {
	scanOptions := []string{
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
	if len(config.AdditionalScanParameters) > 0 {
		scanOptions = append(scanOptions, config.AdditionalScanParameters...)
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
		return errors.Wrapf(err, "failed to execute sourceanalyzer scan command with scanOptions %v", scanOptions)
	}
	return nil
}

func determinePullRequestMerge(config fortifyExecuteScanOptions) (string, string) {
	author := ""
	// TODO provide parameter for trusted certs
	ctx, client, err := piperGithub.NewClientBuilder(config.GithubToken, config.GithubAPIURL).Build()
	if err == nil && ctx != nil && client != nil {
		prID, author, err := determinePullRequestMergeGithub(ctx, config, client.PullRequests)
		if err != nil {
			log.Entry().WithError(err).Warn("Failed to get PR metadata via GitHub client")
		} else {
			return prID, author
		}
	} else {
		log.Entry().WithError(err).Warn("Failed to instantiate GitHub client to get PR metadata")
	}

	log.Entry().Infof("Trying to determine PR ID in commit message: %v", config.CommitMessage)
	r, _ := regexp.Compile(config.PullRequestMessageRegex)
	matches := r.FindSubmatch([]byte(config.CommitMessage))
	if matches != nil && len(matches) > 1 {
		return string(matches[config.PullRequestMessageRegexGroup]), author
	}
	return "0", ""
}

func determinePullRequestMergeGithub(ctx context.Context, config fortifyExecuteScanOptions, pullRequestServiceInstance pullRequestService) (string, string, error) {
	number := "0"
	author := ""
	prList, _, err := pullRequestServiceInstance.ListPullRequestsWithCommit(ctx, config.Owner, config.Repository, config.CommitID, &github.ListOptions{})
	if err == nil && prList != nil && len(prList) > 0 {
		number = fmt.Sprintf("%v", prList[0].GetNumber())
		if prList[0].GetUser() != nil {
			author = prList[0].GetUser().GetLogin()
		}
		return number, author, nil
	}

	log.Entry().Infof("Unable to resolve PR via commit ID: %v", config.CommitID)
	return number, author, err
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

	case "maven", "gradle":
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
		if len(t["pythonVersion"]) > 0 {
			options = append(options, "-python-version", t["pythonVersion"])
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

func createToolRecordFortify(utils fortifyUtils, workspace string, config fortifyExecuteScanOptions, projectID int64, projectName string, projectVersionID int64, projectVersion string) (string, error) {
	record := toolrecord.New(utils, workspace, "fortify", config.ServerURL)
	// Project
	err := record.AddKeyData("project",
		strconv.FormatInt(projectID, 10),
		projectName,
		"")
	if err != nil {
		return "", err
	}
	// projectVersion
	projectVersionURL := config.ServerURL + "/html/ssc/version/" + strconv.FormatInt(projectVersionID, 10)
	err = record.AddKeyData("projectVersion",
		strconv.FormatInt(projectVersionID, 10),
		projectVersion,
		projectVersionURL)
	if err != nil {
		return "", err
	}
	err = record.Persist()
	if err != nil {
		return "", err
	}
	return record.GetFileName(), nil
}

func getProxyParams(proxyUrl string) (string, string) {
	if proxyUrl == "" {
		return "", ""
	}

	urlParams, err := url.Parse(proxyUrl)
	if err != nil {
		log.Entry().Warningf("Failed to parse proxy url %s", proxyUrl)
		return "", ""
	}
	return urlParams.Port(), urlParams.Hostname()
}
