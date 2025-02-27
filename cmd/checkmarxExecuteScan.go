package cmd

import (
	"archive/zip"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/checkmarx"
	piperGithub "github.com/SAP/jenkins-library/pkg/github"
	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/reporting"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/toolrecord"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/pkg/errors"

	"github.com/google/go-github/v68/github"
)

type checkmarxExecuteScanUtils interface {
	FileInfoHeader(fi os.FileInfo) (*zip.FileHeader, error)
	Stat(name string) (os.FileInfo, error)
	Open(name string) (*os.File, error)
	WriteFile(filename string, data []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	PathMatch(pattern, name string) (bool, error)
	GetWorkspace() string
	GetIssueService() *github.IssuesService
	GetSearchService() *github.SearchService
}

type checkmarxExecuteScanUtilsBundle struct {
	workspace string
	issues    *github.IssuesService
	search    *github.SearchService
}

func (c *checkmarxExecuteScanUtilsBundle) PathMatch(pattern, name string) (bool, error) {
	return doublestar.PathMatch(pattern, name)
}

func (c *checkmarxExecuteScanUtilsBundle) GetWorkspace() string {
	return c.workspace
}

func (c *checkmarxExecuteScanUtilsBundle) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}

func (c *checkmarxExecuteScanUtilsBundle) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (c *checkmarxExecuteScanUtilsBundle) FileInfoHeader(fi os.FileInfo) (*zip.FileHeader, error) {
	return zip.FileInfoHeader(fi)
}

func (c *checkmarxExecuteScanUtilsBundle) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (c *checkmarxExecuteScanUtilsBundle) Open(name string) (*os.File, error) {
	return os.Open(name)
}

func (c *checkmarxExecuteScanUtilsBundle) CreateIssue(ghCreateIssueOptions *piperGithub.CreateIssueOptions) error {
	_, err := piperGithub.CreateIssue(ghCreateIssueOptions)
	return err
}

func (c *checkmarxExecuteScanUtilsBundle) GetIssueService() *github.IssuesService {
	return c.issues
}

func (c *checkmarxExecuteScanUtilsBundle) GetSearchService() *github.SearchService {
	return c.search
}

func newCheckmarxExecuteScanUtilsBundle(workspace string, client *github.Client) checkmarxExecuteScanUtils {
	utils := checkmarxExecuteScanUtilsBundle{
		workspace: workspace,
	}
	if client != nil {
		utils.issues = client.Issues
		utils.search = client.Search
	}
	return &utils
}

func checkmarxExecuteScan(config checkmarxExecuteScanOptions, _ *telemetry.CustomData, influx *checkmarxExecuteScanInflux) {
	client := &piperHttp.Client{}
	options := piperHttp.ClientOptions{MaxRetries: config.MaxRetries}
	client.SetOptions(options)
	// TODO provide parameter for trusted certs
	ctx, ghClient, err := piperGithub.NewClientBuilder(config.GithubToken, config.GithubAPIURL).Build()
	if err != nil {
		log.Entry().WithError(err).Warning("Failed to get GitHub client")
	}
	sys, err := checkmarx.NewSystemInstance(client, config.ServerURL, config.Username, config.Password)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Failed to create Checkmarx client talking to URL %v", config.ServerURL)
	}
	influx.step_data.fields.checkmarx = false
	utils := newCheckmarxExecuteScanUtilsBundle("./", ghClient)
	if err := runScan(ctx, config, sys, influx, utils); err != nil {
		log.Entry().WithError(err).Fatal("Failed to execute Checkmarx scan.")
	}
	influx.step_data.fields.checkmarx = true
}

func runScan(ctx context.Context, config checkmarxExecuteScanOptions, sys checkmarx.System, influx *checkmarxExecuteScanInflux, utils checkmarxExecuteScanUtils) error {
	teamID := config.TeamID
	if len(teamID) == 0 {
		readTeamID, err := loadTeamIDByTeamName(config, sys, teamID)
		if err != nil {
			return err
		}
		teamID = readTeamID
	}
	project, projectName, err := loadExistingProject(sys, config.ProjectName, config.PullRequestName, teamID)
	if err != nil {
		return errors.Wrap(err, "error when trying to load project")
	}
	if strings.EqualFold(project.Name, projectName) { // case insensitive string comparison
		err = presetExistingProject(config, sys, projectName, project)
		if err != nil {
			return err
		}
	} else {
		if len(teamID) == 0 {
			return errors.Wrap(err, "TeamName or TeamID is required to create a new project")
		}
		project, err = createNewProject(config, sys, projectName, teamID)
		if err != nil {
			return err
		}
	}

	err = uploadAndScan(ctx, config, sys, project, influx, utils)
	if err != nil {
		return errors.Wrap(err, "scan, upload, and result validation returned an error")
	}
	return nil
}

func loadTeamIDByTeamName(config checkmarxExecuteScanOptions, sys checkmarx.System, teamID string) (string, error) {
	team, err := loadTeam(sys, config.TeamName)
	if err != nil {
		return "", errors.Wrap(err, "failed to load team")
	}
	teamIDBytes, _ := team.ID.MarshalJSON()
	err = json.Unmarshal(teamIDBytes, &teamID)
	if err != nil {
		var teamIDInt int
		err = json.Unmarshal(teamIDBytes, &teamIDInt)
		if err != nil {
			return "", errors.Wrap(err, "failed to unmarshall team.ID")
		}
		teamID = strconv.Itoa(teamIDInt)
	}
	return teamID, nil
}

func createNewProject(config checkmarxExecuteScanOptions, sys checkmarx.System, projectName string, teamID string) (checkmarx.Project, error) {
	log.Entry().Infof("Project %v does not exist, starting to create it...", projectName)
	presetID, _ := strconv.Atoi(config.Preset)
	project, err := createAndConfigureNewProject(sys, projectName, teamID, presetID, config.Preset, config.EngineConfigurationID)
	if err != nil {
		return checkmarx.Project{}, errors.Wrapf(err, "failed to create and configure new project %v", projectName)
	}
	return project, nil
}

func presetExistingProject(config checkmarxExecuteScanOptions, sys checkmarx.System, projectName string, project checkmarx.Project) error {
	log.Entry().Infof("Project %v exists...", projectName)
	if len(config.Preset) > 0 {
		presetID, _ := strconv.Atoi(config.Preset)
		err := setPresetForProject(sys, project.ID, presetID, projectName, config.Preset, config.EngineConfigurationID)
		if err != nil {
			return errors.Wrapf(err, "failed to set preset %v for project %v", config.Preset, projectName)
		}
	}
	return nil
}

func loadTeam(sys checkmarx.System, teamName string) (checkmarx.Team, error) {
	teams := sys.GetTeams()
	team := checkmarx.Team{}
	var err error
	if len(teams) > 0 && len(teamName) > 0 {
		team, err = sys.FilterTeamByName(teams, teamName)
	}
	if err != nil {
		return team, fmt.Errorf("failed to identify team by teamName %v", teamName)
	} else {
		return team, nil
	}
}

func loadExistingProject(sys checkmarx.System, initialProjectName, pullRequestName, teamID string) (checkmarx.Project, string, error) {
	var project checkmarx.Project
	projectName := initialProjectName
	if len(initialProjectName) == 0 {
		return project, projectName, errors.New("You need to provide the Checkmarx project name, projectName parameter is mandatory")
	}
	if len(pullRequestName) > 0 {
		projectName = fmt.Sprintf("%v_%v", initialProjectName, pullRequestName)
		projects, err := sys.GetProjectsByNameAndTeam(projectName, teamID)
		if err != nil || len(projects) == 0 {
			projects, err = sys.GetProjectsByNameAndTeam(initialProjectName, teamID)
			if err != nil {
				return project, projectName, errors.Wrap(err, "failed getting projects")
			}
			if len(projects) == 0 {
				return checkmarx.Project{}, projectName, nil
			}
			branchProject, err := sys.GetProjectByID(sys.CreateBranch(projects[0].ID, projectName))
			if err != nil {
				return project, projectName, fmt.Errorf("failed to create branch %v for project %v", projectName, initialProjectName)
			}
			project = branchProject
		} else {
			project = projects[0]
			log.Entry().Debugf("Loaded project with name %v", project.Name)
		}
	} else {
		projects, err := sys.GetProjectsByNameAndTeam(projectName, teamID)
		if err != nil {
			return project, projectName, errors.Wrap(err, "failed getting projects")
		}
		if len(projects) == 0 {
			return checkmarx.Project{}, projectName, nil
		}
		if len(projects) == 1 {
			project = projects[0]
		} else {
			for _, current_project := range projects {
				if projectName == current_project.Name {
					project = current_project
					break
				}
			}
			if len(project.Name) == 0 {
				return project, projectName, errors.New("Cannot find project " + projectName + ". You need to provide the teamName parameter if you want a new project to be created.")
			}
		}
		log.Entry().Debugf("Loaded project with name %v", project.Name)
	}
	return project, projectName, nil
}

func zipWorkspaceFiles(filterPattern string, utils checkmarxExecuteScanUtils) (*os.File, error) {
	zipFileName := filepath.Join(utils.GetWorkspace(), "workspace.zip")
	patterns := piperutils.Trim(strings.Split(filterPattern, ","))
	sort.Strings(patterns)
	zipFile, err := os.Create(zipFileName)
	if err != nil {
		return zipFile, errors.Wrap(err, "failed to create archive of project sources")
	}
	defer zipFile.Close()
	err = zipFolder(utils.GetWorkspace(), zipFile, patterns, utils)
	if err != nil {
		return nil, errors.Wrap(err, "failed to compact folder")
	}
	return zipFile, nil
}

func uploadAndScan(ctx context.Context, config checkmarxExecuteScanOptions, sys checkmarx.System, project checkmarx.Project, influx *checkmarxExecuteScanInflux, utils checkmarxExecuteScanUtils) error {
	previousScans, err := sys.GetScans(project.ID)
	if err != nil && config.VerifyOnly {
		log.Entry().Warnf("Cannot load scans for project %v, verification only mode aborted", project.Name)
	}
	if len(previousScans) > 0 && config.VerifyOnly {
		err := verifyCxProjectCompliance(ctx, config, sys, previousScans[0].ID, influx, utils)
		if err != nil {
			log.SetErrorCategory(log.ErrorCompliance)
			return errors.Wrapf(err, "project %v not compliant", project.Name)
		}
	} else {
		zipFile, err := zipWorkspaceFiles(config.FilterPattern, utils)
		if err != nil {
			return errors.Wrap(err, "failed to zip workspace files")
		}
		err = sys.UploadProjectSourceCode(project.ID, zipFile.Name())
		if err != nil {
			return errors.Wrapf(err, "failed to upload source code for project %v", project.Name)
		}

		log.Entry().Debugf("Source code uploaded for project %v", project.Name)
		err = os.Remove(zipFile.Name())
		if err != nil {
			log.Entry().WithError(err).Warnf("Failed to delete zipped source code for project %v", project.Name)
		}

		incremental := config.Incremental
		fullScanCycle, err := strconv.Atoi(config.FullScanCycle)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.Wrapf(err, "invalid configuration value for fullScanCycle %v, must be a positive int", config.FullScanCycle)
		}

		if config.IsOptimizedAndScheduled {
			incremental = false
		} else if incremental && config.FullScansScheduled && fullScanCycle > 0 && (getNumCoherentIncrementalScans(previousScans)+1)%fullScanCycle == 0 {
			incremental = false
		} else if incremental && isLastScanFailedIncremental(previousScans) { // if the last incremental scan failed, trigger a full scan instead
			incremental = false
			log.Entry().Infof("Last incremental scan for project %v failed, triggering full scan instead", project.Name)
		}

		return triggerScan(ctx, config, sys, project, incremental, influx, utils)
	}
	return nil
}

func isLastScanFailedIncremental(scans []checkmarx.ScanStatus) bool {
	if len(scans) == 0 {
		return false
	}

	scan := scans[0]
	if scan.IsIncremental && scan.Status.Name == "Failed" {
		return true
	} else {
		return false
	}
}

func triggerScan(ctx context.Context, config checkmarxExecuteScanOptions, sys checkmarx.System, project checkmarx.Project, incremental bool, influx *checkmarxExecuteScanInflux, utils checkmarxExecuteScanUtils) error {
	scan, err := sys.ScanProject(project.ID, incremental, true, !config.AvoidDuplicateProjectScans)
	if err != nil {
		return errors.Wrapf(err, "cannot scan project %v", project.Name)
	}

	log.Entry().Debugf("Scanning project %v ", project.Name)
	err = pollScanStatus(sys, scan)
	if err != nil {
		return errors.Wrap(err, "polling scan status failed")
	}

	log.Entry().Debugln("Scan finished")
	return verifyCxProjectCompliance(ctx, config, sys, scan.ID, influx, utils)
}

func verifyCxProjectCompliance(ctx context.Context, config checkmarxExecuteScanOptions, sys checkmarx.System, scanID int, influx *checkmarxExecuteScanInflux, utils checkmarxExecuteScanUtils) error {
	var reports []piperutils.Path
	if config.GeneratePdfReport {
		pdfReportName := createReportName(utils.GetWorkspace(), "CxSASTReport_%v.pdf")
		err := downloadAndSaveReport(sys, pdfReportName, scanID, utils)
		if err != nil {
			log.Entry().Warning("Report download failed - continue processing ...")
		} else {
			reports = append(reports, piperutils.Path{Target: pdfReportName, Mandatory: true})
		}
	} else {
		log.Entry().Debug("Report generation is disabled via configuration")
	}

	xmlReportName := createReportName(utils.GetWorkspace(), "CxSASTResults_%v.xml")
	results, err := getDetailedResults(config, sys, xmlReportName, scanID, utils)
	if err != nil {
		return errors.Wrap(err, "failed to get detailed results")
	}
	reports = append(reports, piperutils.Path{Target: xmlReportName})

	// generate sarif report
	if config.ConvertToSarif {
		log.Entry().Info("Calling conversion to SARIF function.")
		sarif, err := checkmarx.ConvertCxxmlToSarif(sys, xmlReportName, scanID)
		if err != nil {
			return fmt.Errorf("failed to generate SARIF")
		}
		paths, err := checkmarx.WriteSarif(sarif)
		if err != nil {
			return fmt.Errorf("failed to write sarif")
		}
		reports = append(reports, paths...)
	}

	// create toolrecord
	toolRecordFileName, err := createToolRecordCx(utils, utils.GetWorkspace(), config, results)
	if err != nil {
		// do not fail until the framework is well established
		log.Entry().Warning("TR_CHECKMARX: Failed to create toolrecord file ...", err)
	} else {
		reports = append(reports, piperutils.Path{Target: toolRecordFileName})
	}

	// create JSON report (regardless vulnerabilityThreshold enabled or not)
	jsonReport := checkmarx.CreateJSONReport(results)
	paths, err := checkmarx.WriteJSONReport(jsonReport)
	if err != nil {
		log.Entry().Warning("failed to write JSON report...", err)
	} else {
		// add JSON report to archiving list
		reports = append(reports, paths...)
	}
	links := []piperutils.Path{{Target: results["DeepLink"].(string), Name: "Checkmarx Web UI"}}

	insecure := false
	var insecureResults []string
	var neutralResults []string

	if config.VulnerabilityThresholdEnabled {
		insecure, insecureResults, neutralResults = enforceThresholds(config, results)
		scanReport := checkmarx.CreateCustomReport(results, insecureResults, neutralResults)

		if insecure && config.CreateResultIssue && len(config.GithubToken) > 0 && len(config.GithubAPIURL) > 0 && len(config.Owner) > 0 && len(config.Repository) > 0 {
			log.Entry().Debug("Creating/updating GitHub issue with check results")
			gh := reporting.GitHub{
				Owner:         &config.Owner,
				Repository:    &config.Repository,
				Assignees:     &config.Assignees,
				IssueService:  utils.GetIssueService(),
				SearchService: utils.GetSearchService(),
			}
			if err := gh.UploadSingleReport(ctx, scanReport); err != nil {
				return fmt.Errorf("failed to upload scan results into GitHub: %w", err)
			}
		}

		paths, err := checkmarx.WriteCustomReports(scanReport, fmt.Sprint(results["ProjectName"]), fmt.Sprint(results["ProjectID"]))
		if err != nil {
			// do not fail until we have a better idea to handle it
			log.Entry().Warning("failed to write HTML/MarkDown report file ...", err)
		} else {
			reports = append(reports, paths...)
		}
	}

	piperutils.PersistReportsAndLinks("checkmarxExecuteScan", utils.GetWorkspace(), utils, reports, links)
	reportToInflux(results, influx)

	if insecure {
		if config.VulnerabilityThresholdResult == "FAILURE" {
			log.SetErrorCategory(log.ErrorCompliance)
			return fmt.Errorf("the project is not compliant - see report for details")
		}
		log.Entry().Errorf("Checkmarx scan result set to %v, some results are not meeting defined thresholds. For details see the archived report.", config.VulnerabilityThresholdResult)
	} else {
		log.Entry().Infoln("Checkmarx scan finished successfully")
	}
	return nil
}

func createReportName(workspace, reportFileNameTemplate string) string {
	regExpFileName := regexp.MustCompile(`[^\w\d]`)
	timeStamp, _ := time.Now().Local().MarshalText()
	return filepath.Join(workspace, fmt.Sprintf(reportFileNameTemplate, regExpFileName.ReplaceAllString(string(timeStamp), "_")))
}

func pollScanStatus(sys checkmarx.System, scan checkmarx.Scan) error {
	status := "Scan phase: New"
	pastStatus := status
	log.Entry().Info(status)
	stepDetail := "..."
	stageDetail := "..."
	for {
		var detail checkmarx.ScanStatusDetail
		status, detail = sys.GetScanStatusAndDetail(scan.ID)
		if len(detail.Stage) > 0 {
			stageDetail = detail.Stage
		}
		if len(detail.Step) > 0 {
			stepDetail = detail.Step
		}
		if status == "Finished" || status == "Canceled" || status == "Failed" {
			break
		}

		status = fmt.Sprintf("Scan phase: %v (%v / %v)", status, stageDetail, stepDetail)
		if pastStatus != status {
			log.Entry().Info(status)
			pastStatus = status
		}
		log.Entry().Debug("Polling for status: sleeping...")
		time.Sleep(10 * time.Second)
	}
	if status == "Canceled" {
		log.SetErrorCategory(log.ErrorCustom)
		return fmt.Errorf("scan canceled via web interface")
	}
	if status == "Failed" {
		if strings.Contains(stageDetail, "<ErrorCode>17033</ErrorCode>") { // Translate a cryptic XML error into a human-readable message
			stageDetail = "Failed to start scanning due to one of following reasons: source folder is empty, all source files are of an unsupported language or file format"
		}
		return fmt.Errorf("Checkmarx scan failed with the following error: %v", stageDetail)
	}
	return nil
}

func reportToInflux(results map[string]interface{}, influx *checkmarxExecuteScanInflux) {
	influx.checkmarx_data.fields.high_issues = results["High"].(map[string]int)["Issues"]
	influx.checkmarx_data.fields.high_not_false_positive = results["High"].(map[string]int)["NotFalsePositive"]
	influx.checkmarx_data.fields.high_not_exploitable = results["High"].(map[string]int)["NotExploitable"]
	influx.checkmarx_data.fields.high_confirmed = results["High"].(map[string]int)["Confirmed"]
	influx.checkmarx_data.fields.high_urgent = results["High"].(map[string]int)["Urgent"]
	influx.checkmarx_data.fields.high_proposed_not_exploitable = results["High"].(map[string]int)["ProposedNotExploitable"]
	influx.checkmarx_data.fields.high_to_verify = results["High"].(map[string]int)["ToVerify"]
	influx.checkmarx_data.fields.medium_issues = results["Medium"].(map[string]int)["Issues"]
	influx.checkmarx_data.fields.medium_not_false_positive = results["Medium"].(map[string]int)["NotFalsePositive"]
	influx.checkmarx_data.fields.medium_not_exploitable = results["Medium"].(map[string]int)["NotExploitable"]
	influx.checkmarx_data.fields.medium_confirmed = results["Medium"].(map[string]int)["Confirmed"]
	influx.checkmarx_data.fields.medium_urgent = results["Medium"].(map[string]int)["Urgent"]
	influx.checkmarx_data.fields.medium_proposed_not_exploitable = results["Medium"].(map[string]int)["ProposedNotExploitable"]
	influx.checkmarx_data.fields.medium_to_verify = results["Medium"].(map[string]int)["ToVerify"]
	influx.checkmarx_data.fields.low_issues = results["Low"].(map[string]int)["Issues"]
	influx.checkmarx_data.fields.low_not_false_positive = results["Low"].(map[string]int)["NotFalsePositive"]
	influx.checkmarx_data.fields.low_not_exploitable = results["Low"].(map[string]int)["NotExploitable"]
	influx.checkmarx_data.fields.low_confirmed = results["Low"].(map[string]int)["Confirmed"]
	influx.checkmarx_data.fields.low_urgent = results["Low"].(map[string]int)["Urgent"]
	influx.checkmarx_data.fields.low_proposed_not_exploitable = results["Low"].(map[string]int)["ProposedNotExploitable"]
	influx.checkmarx_data.fields.low_to_verify = results["Low"].(map[string]int)["ToVerify"]
	influx.checkmarx_data.fields.information_issues = results["Information"].(map[string]int)["Issues"]
	influx.checkmarx_data.fields.information_not_false_positive = results["Information"].(map[string]int)["NotFalsePositive"]
	influx.checkmarx_data.fields.information_not_exploitable = results["Information"].(map[string]int)["NotExploitable"]
	influx.checkmarx_data.fields.information_confirmed = results["Information"].(map[string]int)["Confirmed"]
	influx.checkmarx_data.fields.information_urgent = results["Information"].(map[string]int)["Urgent"]
	influx.checkmarx_data.fields.information_proposed_not_exploitable = results["Information"].(map[string]int)["ProposedNotExploitable"]
	influx.checkmarx_data.fields.information_to_verify = results["Information"].(map[string]int)["ToVerify"]
	influx.checkmarx_data.fields.initiator_name = results["InitiatorName"].(string)
	influx.checkmarx_data.fields.owner = results["Owner"].(string)
	influx.checkmarx_data.fields.scan_id = results["ScanId"].(string)
	influx.checkmarx_data.fields.project_id = results["ProjectId"].(string)
	influx.checkmarx_data.fields.projectName = results["ProjectName"].(string)
	influx.checkmarx_data.fields.team = results["Team"].(string)
	influx.checkmarx_data.fields.team_full_path_on_report_date = results["TeamFullPathOnReportDate"].(string)
	influx.checkmarx_data.fields.scan_start = results["ScanStart"].(string)
	influx.checkmarx_data.fields.scan_time = results["ScanTime"].(string)
	influx.checkmarx_data.fields.lines_of_code_scanned = results["LinesOfCodeScanned"].(int)
	influx.checkmarx_data.fields.files_scanned = results["FilesScanned"].(int)
	influx.checkmarx_data.fields.checkmarx_version = results["CheckmarxVersion"].(string)
	influx.checkmarx_data.fields.scan_type = results["ScanType"].(string)
	influx.checkmarx_data.fields.preset = results["Preset"].(string)
	influx.checkmarx_data.fields.deep_link = results["DeepLink"].(string)
	influx.checkmarx_data.fields.report_creation_time = results["ReportCreationTime"].(string)
}

func downloadAndSaveReport(sys checkmarx.System, reportFileName string, scanID int, utils checkmarxExecuteScanUtils) error {
	report, err := generateAndDownloadReport(sys, scanID, "PDF")
	if err != nil {
		return errors.Wrap(err, "failed to download the report")
	}
	log.Entry().Debugf("Saving report to file %v...", reportFileName)
	return utils.WriteFile(reportFileName, report, 0o700)
}

func enforceThresholds(config checkmarxExecuteScanOptions, results map[string]interface{}) (bool, []string, []string) {
	neutralResults := []string{}
	insecureResults := []string{}
	insecure := false
	cxHighThreshold := config.VulnerabilityThresholdHigh
	cxMediumThreshold := config.VulnerabilityThresholdMedium
	cxLowThreshold := config.VulnerabilityThresholdLow
	cxLowThresholdPerQuery := config.VulnerabilityThresholdLowPerQuery
	cxLowThresholdPerQueryMax := config.VulnerabilityThresholdLowPerQueryMax
	highValue := results["High"].(map[string]int)["NotFalsePositive"]
	mediumValue := results["Medium"].(map[string]int)["NotFalsePositive"]
	lowValue := results["Low"].(map[string]int)["NotFalsePositive"]
	var unit string
	highViolation := ""
	mediumViolation := ""
	lowViolation := ""
	if config.VulnerabilityThresholdUnit == "percentage" {
		unit = "%"
		highAudited := results["High"].(map[string]int)["Issues"] - results["High"].(map[string]int)["NotFalsePositive"]
		highOverall := results["High"].(map[string]int)["Issues"]
		if highOverall == 0 {
			highAudited = 1
			highOverall = 1
		}
		mediumAudited := results["Medium"].(map[string]int)["Issues"] - results["Medium"].(map[string]int)["NotFalsePositive"]
		mediumOverall := results["Medium"].(map[string]int)["Issues"]
		if mediumOverall == 0 {
			mediumAudited = 1
			mediumOverall = 1
		}
		lowAudited := results["Low"].(map[string]int)["Confirmed"] + results["Low"].(map[string]int)["NotExploitable"]
		lowOverall := results["Low"].(map[string]int)["Issues"]
		if lowOverall == 0 {
			lowAudited = 1
			lowOverall = 1
		}
		highValue = int(float32(highAudited) / float32(highOverall) * 100.0)
		mediumValue = int(float32(mediumAudited) / float32(mediumOverall) * 100.0)
		lowValue = int(float32(lowAudited) / float32(lowOverall) * 100.0)

		if highValue < cxHighThreshold {
			insecure = true
			highViolation = fmt.Sprintf("<-- %v %v deviation", cxHighThreshold-highValue, unit)
		}
		if mediumValue < cxMediumThreshold {
			insecure = true
			mediumViolation = fmt.Sprintf("<-- %v %v deviation", cxMediumThreshold-mediumValue, unit)
		}
		// if the flag is switched on, calculate the Low findings threshold per query
		if cxLowThresholdPerQuery {
			lowPerQueryMap := results["LowPerQuery"].(map[string]map[string]int)
			if lowPerQueryMap != nil {
				for lowQuery, resultsLowQuery := range lowPerQueryMap {
					lowAuditedPerQuery := resultsLowQuery["Confirmed"] + resultsLowQuery["NotExploitable"]
					lowOverallPerQuery := resultsLowQuery["Issues"]
					lowAuditedRequiredPerQuery := int(math.Ceil(float64(lowOverallPerQuery) * float64(cxLowThreshold) / 100.0))
					if lowAuditedPerQuery < lowAuditedRequiredPerQuery && lowAuditedPerQuery < cxLowThresholdPerQueryMax {
						insecure = true
						msgSeperator := "|"
						if lowViolation == "" {
							msgSeperator = "<--"
						}
						lowViolation += fmt.Sprintf(" %v query: %v, audited: %v, required: %v ", msgSeperator, lowQuery, lowAuditedPerQuery, lowAuditedRequiredPerQuery)
					}
				}
			}
		} else { // calculate the Low findings threshold in total
			if lowValue < cxLowThreshold {
				insecure = true
				lowViolation = fmt.Sprintf("<-- %v %v deviation", cxLowThreshold-lowValue, unit)
			}
		}

	}
	if config.VulnerabilityThresholdUnit == "absolute" {
		unit = " findings"
		if highValue > cxHighThreshold {
			insecure = true
			highViolation = fmt.Sprintf("<-- %v%v deviation", highValue-cxHighThreshold, unit)
		}
		if mediumValue > cxMediumThreshold {
			insecure = true
			mediumViolation = fmt.Sprintf("<-- %v%v deviation", mediumValue-cxMediumThreshold, unit)
		}
		if lowValue > cxLowThreshold {
			insecure = true
			lowViolation = fmt.Sprintf("<-- %v%v deviation", lowValue-cxLowThreshold, unit)
		}
	}

	highText := fmt.Sprintf("High %v%v %v", highValue, unit, highViolation)
	mediumText := fmt.Sprintf("Medium %v%v %v", mediumValue, unit, mediumViolation)
	lowText := fmt.Sprintf("Low %v%v %v", lowValue, unit, lowViolation)
	if len(highViolation) > 0 {
		insecureResults = append(insecureResults, highText)
		log.Entry().Error(highText)
	} else {
		neutralResults = append(neutralResults, highText)
		log.Entry().Info(highText)
	}
	if len(mediumViolation) > 0 {
		insecureResults = append(insecureResults, mediumText)
		log.Entry().Error(mediumText)
	} else {
		neutralResults = append(neutralResults, mediumText)
		log.Entry().Info(mediumText)
	}
	if len(lowViolation) > 0 {
		insecureResults = append(insecureResults, lowText)
		log.Entry().Error(lowText)
	} else {
		neutralResults = append(neutralResults, lowText)
		log.Entry().Info(lowText)
	}

	return insecure, insecureResults, neutralResults
}

func createAndConfigureNewProject(sys checkmarx.System, projectName, teamID string, presetIDValue int, presetValue, engineConfiguration string) (checkmarx.Project, error) {
	if len(presetValue) == 0 {
		log.SetErrorCategory(log.ErrorConfiguration)
		return checkmarx.Project{}, fmt.Errorf("preset not specified, creation of project %v failed", projectName)
	}

	projectCreateResult, err := sys.CreateProject(projectName, teamID)
	if err != nil {
		return checkmarx.Project{}, errors.Wrapf(err, "cannot create project %v", projectName)
	}

	if err := setPresetForProject(sys, projectCreateResult.ID, presetIDValue, projectName, presetValue, engineConfiguration); err != nil {
		return checkmarx.Project{}, errors.Wrapf(err, "failed to set preset %v for project", presetValue)
	}

	projects, err := sys.GetProjectsByNameAndTeam(projectName, teamID)
	if err != nil || len(projects) == 0 {
		return checkmarx.Project{}, errors.Wrapf(err, "failed to load newly created project %v", projectName)
	}
	log.Entry().Debugf("New Project %v created", projectName)
	log.Entry().Debugf("Projects: %v", projects)
	return projects[0], nil
}

// loadPreset finds a checkmarx.Preset that has either the ID or Name given by presetValue.
// presetValue is not expected to be empty.
func loadPreset(sys checkmarx.System, presetValue string) (checkmarx.Preset, error) {
	presets := sys.GetPresets()
	var preset checkmarx.Preset
	var configuredPresetName string
	preset = sys.FilterPresetByName(presets, presetValue)
	configuredPresetName = presetValue
	if len(configuredPresetName) > 0 && preset.Name == configuredPresetName {
		log.Entry().Infof("Loaded preset %v", preset.Name)
		return preset, nil
	}
	log.Entry().Infof("Preset '%s' not found. Available presets are:", presetValue)
	for _, prs := range presets {
		log.Entry().Infof("preset id: %v, name: '%v'", prs.ID, prs.Name)
	}
	return checkmarx.Preset{}, fmt.Errorf("preset %v not found", preset.Name)
}

// setPresetForProject is only called when it has already been established that the preset needs to be set.
// It will exit via the logging framework in case the preset could be found, or the project could not be updated.
func setPresetForProject(sys checkmarx.System, projectID, presetIDValue int, projectName, presetValue, engineConfiguration string) error {
	presetID := presetIDValue
	if presetID <= 0 {
		preset, err := loadPreset(sys, presetValue)
		if err != nil {
			return errors.Wrapf(err, "preset %v not found, configuration of project %v failed", presetValue, projectName)
		}
		presetID = preset.ID
	}
	err := sys.UpdateProjectConfiguration(projectID, presetID, engineConfiguration)
	if err != nil {
		return errors.Wrapf(err, "updating configuration of project %v failed", projectName)
	}
	return nil
}

func generateAndDownloadReport(sys checkmarx.System, scanID int, reportType string) ([]byte, error) {
	report, err := sys.RequestNewReport(scanID, reportType)
	if err != nil {
		return []byte{}, errors.Wrap(err, "failed to request new report")
	}
	finalStatus := 1
	for {
		reportStatus, err := sys.GetReportStatus(report.ReportID)
		if err != nil {
			return []byte{}, errors.Wrap(err, "failed to get report status")
		}
		finalStatus = reportStatus.Status.ID
		if finalStatus != 1 {
			break
		}
		time.Sleep(10 * time.Second)
	}
	if finalStatus == 2 {
		return sys.DownloadReport(report.ReportID)
	}
	return []byte{}, fmt.Errorf("unexpected status %v recieved", finalStatus)
}

func getNumCoherentIncrementalScans(scans []checkmarx.ScanStatus) int {
	count := 0
	for _, scan := range scans {
		if !scan.IsIncremental {
			break
		}
		count++
	}
	return count
}

func getDetailedResults(config checkmarxExecuteScanOptions, sys checkmarx.System, reportFileName string, scanID int, utils checkmarxExecuteScanUtils) (map[string]interface{}, error) {
	resultMap := map[string]interface{}{}
	data, err := generateAndDownloadReport(sys, scanID, "XML")
	if err != nil {
		return resultMap, errors.Wrap(err, "failed to download xml report")
	}
	if len(data) > 0 {
		err = utils.WriteFile(reportFileName, data, 0o700)
		if err != nil {
			return resultMap, errors.Wrap(err, "failed to write file")
		}
		var xmlResult checkmarx.DetailedResult
		err := xml.Unmarshal(data, &xmlResult)
		if err != nil {
			return resultMap, errors.Wrapf(err, "failed to unmarshal XML report for scan %v", scanID)
		}
		resultMap["InitiatorName"] = xmlResult.InitiatorName
		resultMap["Owner"] = xmlResult.Owner
		resultMap["ScanId"] = xmlResult.ScanID
		resultMap["ProjectId"] = xmlResult.ProjectID
		resultMap["ProjectName"] = xmlResult.ProjectName
		resultMap["Team"] = xmlResult.Team
		resultMap["TeamFullPathOnReportDate"] = xmlResult.TeamFullPathOnReportDate
		resultMap["ScanStart"] = xmlResult.ScanStart
		resultMap["ScanTime"] = xmlResult.ScanTime
		resultMap["LinesOfCodeScanned"] = xmlResult.LinesOfCodeScanned
		resultMap["FilesScanned"] = xmlResult.FilesScanned
		resultMap["CheckmarxVersion"] = xmlResult.CheckmarxVersion
		resultMap["ScanType"] = xmlResult.ScanType
		resultMap["Preset"] = xmlResult.Preset
		resultMap["DeepLink"] = xmlResult.DeepLink
		resultMap["ReportCreationTime"] = xmlResult.ReportCreationTime
		resultMap["High"] = map[string]int{}
		resultMap["Medium"] = map[string]int{}
		resultMap["Low"] = map[string]int{}
		resultMap["Information"] = map[string]int{}
		for _, query := range xmlResult.Queries {
			for _, result := range query.Results {
				key := result.Severity
				var submap map[string]int
				if resultMap[key] == nil {
					submap = map[string]int{}
					resultMap[key] = submap
				} else {
					submap = resultMap[key].(map[string]int)
				}
				submap["Issues"]++

				auditState := "ToVerify"
				switch result.State {
				case "1":
					auditState = "NotExploitable"
				case "2":
					auditState = "Confirmed"
				case "3":
					auditState = "Urgent"
				case "4":
					auditState = "ProposedNotExploitable"
				case "0":
				default:
					auditState = "ToVerify"
				}
				submap[auditState]++

				if result.FalsePositive != "True" {
					submap["NotFalsePositive"]++
				}
			}
		}

		// if the flag is switched on, build the list  of Low findings per query
		if config.VulnerabilityThresholdLowPerQuery {
			var lowPerQuery = map[string]map[string]int{}
			for _, query := range xmlResult.Queries {
				for _, result := range query.Results {
					if result.Severity != "Low" {
						continue
					}
					key := query.Name
					var submap map[string]int
					if lowPerQuery[key] == nil {
						submap = map[string]int{}
						lowPerQuery[key] = submap
					} else {
						submap = lowPerQuery[key]
					}
					submap["Issues"]++
					auditState := "ToVerify"
					switch result.State {
					case "1":
						auditState = "NotExploitable"
						break
					case "2":
						auditState = "Confirmed"
						break
					case "3":
						auditState = "Urgent"
						break
					case "4":
						auditState = "ProposedNotExploitable"
						break
					case "0":
					default:
						auditState = "ToVerify"
						break
					}
					submap[auditState]++

					if result.FalsePositive != "True" {
						submap["NotFalsePositive"]++
					}
				}
			}
			resultMap["LowPerQuery"] = lowPerQuery
		}
	}
	return resultMap, nil
}

func zipFolder(source string, zipFile io.Writer, patterns []string, utils checkmarxExecuteScanUtils) error {
	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	info, err := utils.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	fileCount := 0
	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		noMatch, err := isFileNotMatchingPattern(patterns, path, info, utils)
		if err != nil || noMatch {
			return err
		}

		header, err := utils.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
		}

		adaptHeader(info, header)

		writer, err := archive.CreateHeader(header)
		if err != nil || info.IsDir() {
			return err
		}

		file, err := utils.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		fileCount++
		return err
	})
	log.Entry().Infof("Zipped %d files", fileCount)
	err = handleZeroFilesZipped(source, err, fileCount)
	return err
}

func adaptHeader(info os.FileInfo, header *zip.FileHeader) {
	if info.IsDir() {
		header.Name += "/"
	} else {
		header.Method = zip.Deflate
	}
}

func handleZeroFilesZipped(source string, err error, fileCount int) error {
	if err == nil && fileCount == 0 {
		log.SetErrorCategory(log.ErrorConfiguration)
		err = fmt.Errorf("filterPattern matched no files or workspace directory '%s' was empty", source)
	}
	return err
}

// isFileNotMatchingPattern checks if file path does not match one of the patterns.
// If it matches a negative pattern (starting with '!') then true is returned.
//
// If it is a directory, false is returned.
// If no patterns are provided, false is returned.
func isFileNotMatchingPattern(patterns []string, path string, info os.FileInfo, utils checkmarxExecuteScanUtils) (bool, error) {
	if len(patterns) == 0 || info.IsDir() {
		return false, nil
	}

	for _, pattern := range patterns {
		negative := false
		if strings.HasPrefix(pattern, "!") {
			pattern = strings.TrimLeft(pattern, "!")
			negative = true
		}
		match, err := utils.PathMatch(pattern, path)
		if err != nil {
			return false, errors.Wrapf(err, "Pattern %v could not get executed", pattern)
		}

		if match {
			return negative, nil
		}
	}
	return true, nil
}

func createToolRecordCx(utils checkmarxExecuteScanUtils, workspace string, config checkmarxExecuteScanOptions, results map[string]interface{}) (string, error) {
	record := toolrecord.New(utils, workspace, "checkmarx", config.ServerURL)
	// Todo TeamId - see run_scan()
	// record.AddKeyData("team", XXX, resultMap["Team"], "")
	// Project
	err := record.AddKeyData("project",
		results["ProjectId"].(string),
		results["ProjectName"].(string),
		"")
	if err != nil {
		return "", err
	}
	// Scan
	err = record.AddKeyData("scanid",
		results["ScanId"].(string),
		results["ScanId"].(string),
		results["DeepLink"].(string))
	if err != nil {
		return "", err
	}
	err = record.Persist()
	if err != nil {
		return "", err
	}
	return record.GetFileName(), nil
}
