package cmd

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	checkmarxOne "github.com/SAP/jenkins-library/pkg/checkmarxone"
	piperGithub "github.com/SAP/jenkins-library/pkg/github"
	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/reporting"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/toolrecord"
	"github.com/bmatcuk/doublestar"
	"github.com/google/go-github/v68/github"
	"github.com/pkg/errors"
)

type checkmarxOneExecuteScanUtils interface {
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

type checkmarxOneExecuteScanHelper struct {
	ctx     context.Context
	config  checkmarxOneExecuteScanOptions
	sys     checkmarxOne.System
	influx  *checkmarxOneExecuteScanInflux
	utils   checkmarxOneExecuteScanUtils
	Project *checkmarxOne.Project
	Group   *checkmarxOne.Group
	App     *checkmarxOne.Application
	reports []piperutils.Path
}

type checkmarxOneExecuteScanUtilsBundle struct {
	workspace string
	issues    *github.IssuesService
	search    *github.SearchService
}

func checkmarxOneExecuteScan(config checkmarxOneExecuteScanOptions, _ *telemetry.CustomData, influx *checkmarxOneExecuteScanInflux) {
	// TODO: Setup connection with Splunk, influxDB?
	cx1sh, err := Authenticate(config, influx)
	if err != nil {
		log.Entry().WithError(err).Fatalf("failed to create Cx1 client: %s", err)
	}

	err = runStep(config, influx, &cx1sh)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Failed to run CheckmarxOne scan.")
	}
	influx.step_data.fields.checkmarxOne = true
}

func runStep(config checkmarxOneExecuteScanOptions, influx *checkmarxOneExecuteScanInflux, cx1sh *checkmarxOneExecuteScanHelper) error {
	err := error(nil)
	cx1sh.Project, err = cx1sh.GetProjectByName()
	if err != nil && err.Error() != "project not found" {
		return fmt.Errorf("failed to get project: %s", err)
	}

	if len(config.GroupName) > 0 {
		cx1sh.Group, err = cx1sh.GetGroup() // used when creating a project and when generating a SARIF report
		if err != nil {
			log.Entry().WithError(err).Warnf("failed to get group")
		}
	}

	if cx1sh.Project == nil {
		if len(config.ApplicationName) > 0 {
			cx1sh.App, err = cx1sh.GetApplication() // read application name from piper config (optional) and get ID from CxONE API
			if err != nil {
				return fmt.Errorf("failed to get application: %v", err)
			}
		}
		cx1sh.Project, err = cx1sh.CreateProject() // requires groups, repoUrl, mainBranch, origin, tags, criticality
		if err != nil {
			return fmt.Errorf("failed to create project: %s", err)
		}
	} else {
		cx1sh.Project, err = cx1sh.GetProjectByID(cx1sh.Project.ProjectID)
		if err != nil {
			return fmt.Errorf("failed to get project by ID: %s", err)
		} else {
			if len(cx1sh.Project.Applications) > 0 {
				appId := cx1sh.Project.Applications[0]
				cx1sh.App, err = cx1sh.GetApplicationByID(cx1sh.Project.Applications[0])
				if err != nil {
					return fmt.Errorf("failed to retrieve information for project's assigned application %v", appId)
				}
			}
		}
	}

	err = cx1sh.SetProjectPreset()
	if err != nil {
		return fmt.Errorf("failed to set preset: %s", err)
	}

	// update project's tags
	if (len(config.ProjectTags)) > 0 {
		err = cx1sh.UpdateProjectTags()
		if err != nil {
			log.Entry().WithError(err).Warnf("failed to tags the project: %s", err)
		}
	}

	branch, isPR, baseBranch := cx1sh.GetScanBranch()
	scans, err := cx1sh.GetLastScans(10, branch)
	if err != nil {
		log.Entry().WithError(err).Warnf("failed to get last 10 scans")
	}

	if config.VerifyOnly {
		if len(scans) > 0 {
			results, err := cx1sh.ParseResults(&scans[0]) // incl report-gen
			if err != nil {
				return fmt.Errorf("failed to get scan results: %s", err)
			}

			err = cx1sh.CheckCompliance(&scans[0], &results)
			if err != nil {
				log.SetErrorCategory(log.ErrorCompliance)
				return fmt.Errorf("project %v not compliant: %s", cx1sh.Project.Name, err)
			}

			return nil
		} else {
			log.Entry().Warnf("Cannot load scans for project %v, verification only mode aborted", cx1sh.Project.Name)
		}
	}

	incremental, err := cx1sh.IncrementalOrFull(scans) // requires: scan list
	if err != nil {
		return fmt.Errorf("failed to determine incremental or full scan configuration: %s", err)
	}

	if config.Incremental {
		log.Entry().Info("If you change your file filter pattern it is recommended to run a Full scan instead of an incremental, to ensure full code coverage.")
	}

	zipFile, err := cx1sh.ZipFiles()
	if err != nil {
		return fmt.Errorf("failed to create zip file: %s", err)
	}

	uploadLink, err := cx1sh.UploadScanContent(zipFile) // POST /api/uploads + PUT /{uploadLink}
	if err != nil {
		return fmt.Errorf("failed to get upload URL: %s", err)
	}

	// TODO : The step structure should allow to enable different scanners: SAST, KICKS, SCA
	var scan *checkmarxOne.Scan
	// We search for a full scan in history:
	fullScanExists := false
	for _, histScan := range scans {
		isIncr, _ := histScan.IsIncremental()
		if isIncr == false {
			fullScanExists = true
			break
		}
	}
	// user requested an incremental scan on a branch, and the project has a Primary Branch set, not in PR context and no full scan on the branch
	if config.Incremental && !isPR && !fullScanExists && cx1sh.Project.MainBranch != "" && cx1sh.Project.MainBranch != branch {
		scansMainBranch, err := cx1sh.GetLastScans(10, cx1sh.Project.MainBranch)
		if err != nil {
			return fmt.Errorf("failed to determine incremental or full scan configuration: %s", err)
		}
		// We check if the main branch is eligible for an incremental scan
		incrementalMainBranch, err := cx1sh.IncrementalOrFull(scansMainBranch)
		log.Entry().Debugf("Main branch %v incremental scan eligibility: %t", cx1sh.Project.MainBranch, incrementalMainBranch)
		scan, err = cx1sh.CreateScanRequest(incrementalMainBranch, uploadLink, cx1sh.Project.MainBranch) // this will create a full scan on the current branch if the main branch is not eligible for an incremental scan
	} else if config.Incremental && isPR && len(baseBranch) > 0 && baseBranch != "n/a" { // running in a PR context, and we have a base branch for the incremental scan
		// in a PR context we always want to do an incremental scan
		// The scan will be based on the PR's target branch (baseBranch) if there is no full scan on the PR branch
		if fullScanExists {
			log.Entry().Debugf("A full scan exists on the PR branch %v, so the incremental scan will be based on it", branch)
			scan, err = cx1sh.CreateScanRequest(true, uploadLink, "")
		} else {
			log.Entry().Debugf("There is no full scan on the PR branch %v, so the incremental scan will be based on branch %v", branch, baseBranch)
			scan, err = cx1sh.CreateScanRequest(true, uploadLink, baseBranch)
		}
	} else {
		scan, err = cx1sh.CreateScanRequest(incremental, uploadLink, "")
	}

	if err != nil {
		return fmt.Errorf("failed to create scan: %s", err)
	}

	// TODO: how to provide other scan parameters like engineConfiguration?
	// TODO: potential to persist file exclusions for git?
	err = cx1sh.PollScanStatus(scan)
	if err != nil {
		return fmt.Errorf("failed while polling scan status: %s", err)
	}

	results, err := cx1sh.ParseResults(scan) // incl report-gen
	if err != nil {
		return fmt.Errorf("failed to get scan results: %s", err)
	}
	err = cx1sh.CheckCompliance(scan, &results)
	if err != nil {
		log.SetErrorCategory(log.ErrorCompliance)
		return fmt.Errorf("project %v not compliant: %s", cx1sh.Project.Name, err)
	}
	// TODO: upload logs to Splunk, influxDB?
	return nil

}

func Authenticate(config checkmarxOneExecuteScanOptions, influx *checkmarxOneExecuteScanInflux) (checkmarxOneExecuteScanHelper, error) {
	client := &piperHttp.Client{}

	ctx, ghClient, err := piperGithub.NewClientBuilder(config.GithubToken, config.GithubAPIURL).Build()
	if err != nil {
		log.Entry().WithError(err).Warning("Failed to get GitHub client")
	}
	sys, err := checkmarxOne.NewSystemInstance(client, config.ServerURL, config.IamURL, config.Tenant, config.APIKey, config.ClientID, config.ClientSecret)
	if err != nil {
		return checkmarxOneExecuteScanHelper{}, fmt.Errorf("failed to create Checkmarx One client talking to URLs %v and %v with tenant %v: %s", config.ServerURL, config.IamURL, config.Tenant, err)
	}
	influx.step_data.fields.checkmarxOne = false

	utils := newcheckmarxOneExecuteScanUtilsBundle("./", ghClient)

	return checkmarxOneExecuteScanHelper{ctx, config, sys, influx, utils, nil, nil, nil, []piperutils.Path{}}, nil
}

func (c *checkmarxOneExecuteScanHelper) GetProjectByName() (*checkmarxOne.Project, error) {
	if len(c.config.ProjectName) == 0 {
		log.Entry().Fatalf("No project name set in the configuration")
	}

	// get the Project, if it exists
	projects, err := c.sys.GetProjectsByName(c.config.ProjectName)
	if err != nil {
		return nil, fmt.Errorf("error when trying to load project: %s", err)
	}

	for _, p := range projects {
		if p.Name == c.config.ProjectName {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("project not found")
}

func (c *checkmarxOneExecuteScanHelper) GetProjectByID(projectId string) (*checkmarxOne.Project, error) {
	project, err := c.sys.GetProjectByID(projectId)
	return &project, err
}

func (c *checkmarxOneExecuteScanHelper) GetGroup() (*checkmarxOne.Group, error) {
	if len(c.config.GroupName) > 0 {
		group, err := c.sys.GetGroupByName(c.config.GroupName)
		if err != nil {
			return nil, fmt.Errorf("Failed to get Checkmarx One group by Name %v: %s", c.config.GroupName, err)
		}
		return &group, nil
	}
	return nil, fmt.Errorf("No group name specified in configuration")
}

func (c *checkmarxOneExecuteScanHelper) GetApplication() (*checkmarxOne.Application, error) {
	if len(c.config.ApplicationName) > 0 {
		app, err := c.sys.GetApplicationByName(c.config.ApplicationName)
		if err != nil {
			return nil, fmt.Errorf("Failed to get Checkmarx One application by Name %v: %s", c.config.ApplicationName, err)
		}

		return &app, nil
	}
	return nil, fmt.Errorf("No application name specified in configuration")
}

func (c *checkmarxOneExecuteScanHelper) GetApplicationByID(applicationId string) (*checkmarxOne.Application, error) {
	app, err := c.sys.GetApplicationByID(applicationId)
	if err != nil {
		return nil, fmt.Errorf("Failed to get Checkmarx One application by Name %v: %s", c.config.ApplicationName, err)
	}

	return &app, nil
}

func (c *checkmarxOneExecuteScanHelper) CreateProject() (*checkmarxOne.Project, error) {
	if len(c.config.Preset) == 0 {
		return nil, fmt.Errorf("Preset is required to create a project")
	}

	var project checkmarxOne.Project
	var err error
	var groupIDs []string = []string{}
	if c.Group != nil {
		groupIDs = []string{c.Group.GroupID}
	}

	if c.App != nil {
		project, err = c.sys.CreateProjectInApplication(c.config.ProjectName, c.App.ApplicationID, groupIDs)
	} else {
		project, err = c.sys.CreateProject(c.config.ProjectName, groupIDs)
	}

	if err != nil {
		return nil, fmt.Errorf("Error when trying to create project: %s", err)
	}
	log.Entry().Infof("Project %v created", project.ProjectID)

	// new project, set the defaults per pipeline config
	err = c.sys.SetProjectPreset(project.ProjectID, c.config.Preset, true)
	if err != nil {
		return nil, fmt.Errorf("Unable to set preset for project %v to %v: %s", project.ProjectID, c.config.Preset, err)
	}
	log.Entry().Infof("Project preset updated to %v", c.config.Preset)

	if len(c.config.LanguageMode) != 0 {
		err = c.sys.SetProjectLanguageMode(project.ProjectID, c.config.LanguageMode, true)
		if err != nil {

			return nil, fmt.Errorf("Unable to set languageMode for project %v to %v: %s", project.ProjectID, c.config.LanguageMode, err)
		}
		log.Entry().Infof("Project languageMode updated to %v", c.config.LanguageMode)
	}

	return &project, nil
}

func (c *checkmarxOneExecuteScanHelper) UpdateProjectTags() error {
	if len(c.config.ProjectTags) > 0 {
		tags := make(map[string]string, 0)
		err := json.Unmarshal([]byte(c.config.ProjectTags), &tags)
		if err != nil {
			log.Entry().Infof("Failed to parse the project tags: %v", c.config.ProjectTags)
			return err
		}
		// merge new tags to the existing ones
		maps.Copy(c.Project.Tags, tags)

		return c.sys.UpdateProject(c.Project)
	}

	return nil
}

func (c *checkmarxOneExecuteScanHelper) SetProjectPreset() error {
	projectConf, err := c.sys.GetProjectConfiguration(c.Project.ProjectID)

	if err != nil {
		return fmt.Errorf("Failed to retrieve current project configuration: %s", err)
	}

	currentPreset := ""
	currentLanguageMode := "multi" // piper default
	for _, conf := range projectConf {
		if conf.Key == "scan.config.sast.presetName" {
			currentPreset = conf.Value
		}
		if conf.Key == "scan.config.sast.languageMode" {
			currentLanguageMode = conf.Value
		}
	}

	if c.config.LanguageMode == "" || strings.EqualFold(c.config.LanguageMode, "multi") { // default multi if blank
		if currentLanguageMode != "multi" {
			log.Entry().Info("Pipeline yaml requests multi-language scan - updating project configuration")
			c.sys.SetProjectLanguageMode(c.Project.ProjectID, "multi", true)

			if c.config.Incremental {
				log.Entry().Warn("Pipeline yaml requests incremental scan, but switching from 'primary' to 'multi' language mode requires a full scan - switching from incremental to full")
				c.config.Incremental = false
			}
		}
	} else { // primary language mode
		if currentLanguageMode != "primary" {
			log.Entry().Info("Pipeline yaml requests primary-language scan - updating project configuration")
			c.sys.SetProjectLanguageMode(c.Project.ProjectID, "primary", true)
			// no need to switch incremental to full here (multi-language scan includes single-language scan coverage)
		}
	}

	if c.config.Preset == "" {
		if currentPreset == "" {
			return fmt.Errorf("must specify the preset in either the pipeline yaml or in the CheckmarxOne project configuration")
		} else {
			log.Entry().Infof("Pipeline yaml does not specify a preset, will use project configuration (%v).", currentPreset)
		}
		c.config.Preset = currentPreset
	} else if currentPreset != c.config.Preset {
		log.Entry().Infof("Project configured preset (%v) does not match pipeline yaml (%v) - updating project configuration.", currentPreset, c.config.Preset)
		c.sys.SetProjectPreset(c.Project.ProjectID, c.config.Preset, true)

		if c.config.Incremental {
			log.Entry().Warn("Changing project settings requires a full scan to take effect - switching from incremental to full")
			c.config.Incremental = false
		}
	} else {
		log.Entry().Infof("Project is already configured to use pipeline preset %v", currentPreset)
	}
	return nil
}

func (c *checkmarxOneExecuteScanHelper) GetLastScans(count int, branch string) ([]checkmarxOne.Scan, error) {
	scans, err := c.sys.GetLastScansByStatus(c.Project.ProjectID, branch, count, []string{"Completed"})
	if err != nil {
		return []checkmarxOne.Scan{}, fmt.Errorf("Failed to get last %d Completed scans for project %v: %s", count, c.Project.ProjectID, err)
	}
	return scans, nil
}

func (c *checkmarxOneExecuteScanHelper) IncrementalOrFull(scans []checkmarxOne.Scan) (bool, error) {
	incremental := c.config.Incremental
	fullScanCycle, err := strconv.Atoi(c.config.FullScanCycle)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return false, fmt.Errorf("invalid configuration value for fullScanCycle %v, must be a positive int", c.config.FullScanCycle)
	}

	coherentIncrementalScans := c.getNumCoherentIncrementalScans(scans)

	if c.config.IsOptimizedAndScheduled {
		incremental = false
	} else if incremental && c.config.FullScansScheduled && fullScanCycle > 0 && (coherentIncrementalScans+1) >= fullScanCycle {
		incremental = false
	}

	return incremental, nil
}

func (c *checkmarxOneExecuteScanHelper) ZipFiles() (*os.File, error) {
	zipFile, err := c.zipWorkspaceFiles(c.config.FilterPattern, c.utils)
	if err != nil {
		return nil, fmt.Errorf("Failed to zip workspace files")
	}
	return zipFile, nil
}

func (c *checkmarxOneExecuteScanHelper) UploadScanContent(zipFile *os.File) (string, error) {
	uploadUri, err := c.sys.UploadProjectSourceCode(c.Project.ProjectID, zipFile.Name())
	if err != nil {
		return "", fmt.Errorf("Failed to upload source code for project %v: %s", c.Project.ProjectID, err)
	}

	log.Entry().Debugf("Source code uploaded for project %v", c.Project.Name)
	err = os.Remove(zipFile.Name())
	if err != nil {
		log.Entry().WithError(err).Warnf("Failed to delete zipped source code for project %v", c.Project.Name)
	}
	return uploadUri, nil
}

func (c *checkmarxOneExecuteScanHelper) GetScanBranch() (string, bool, string) {
	branch := c.config.Branch
	cicdOrch, err := orchestrator.GetOrchestratorConfigProvider(nil)
	if err != nil {
		log.Entry().Warn("Could not identify orchestrator")
	}
	if len(branch) == 0 && len(c.config.GitBranch) > 0 && c.config.GitBranch != "n/a" {
		branch = c.config.GitBranch
	} else if len(branch) == 0 && (len(c.config.GitBranch) == 0 || c.config.GitBranch == "n/a") { // use the branch from the orchestrator by default
		cicdBranch := cicdOrch.Branch()
		if cicdBranch != "n/a" {
			branch = cicdBranch
		} else {
			log.Entry().Info("Could not retrieve branch name from orchestrator")
		}
	}
	if len(c.config.PullRequestName) > 0 {
		branch = fmt.Sprintf("%v-%v", c.config.PullRequestName, branch)
	} else if cicdOrch.IsPullRequest() && cicdOrch.PullRequestConfig().Branch != "n/a" {
		branch = fmt.Sprintf("PR%v-%v", cicdOrch.PullRequestConfig().Key, cicdOrch.PullRequestConfig().Branch)
	}

	baseBranch := cicdOrch.PullRequestConfig().Base
	isPR := cicdOrch.IsPullRequest()
	log.Entry().Infof("CxOne scan branch was automatically set to : %v", branch)
	return branch, isPR, baseBranch
}

func (c *checkmarxOneExecuteScanHelper) CreateScanRequest(incremental bool, uploadLink string, baseBranch string) (*checkmarxOne.Scan, error) {
	sastConfigString := ""
	sastConfig := checkmarxOne.ScanConfiguration{}
	sastConfig.ScanType = "sast"

	sastConfig.Values = make(map[string]string, 0)
	sastConfig.Values["incremental"] = strconv.FormatBool(incremental)
	sastConfig.Values["presetName"] = c.config.Preset // always set, either coming from config or coming from Cx1 configuration
	if incremental && len(baseBranch) > 0 {           // base the incremental scan on the specified base branch
		sastConfig.Values["baseBranch"] = baseBranch
		sastConfigString = fmt.Sprintf("baseBranch: %v, ", baseBranch)
	}
	sastConfigString = fmt.Sprintf("%vincremental %v, preset %v", sastConfigString, strconv.FormatBool(incremental), c.config.Preset)

	if len(c.config.LanguageMode) > 0 {
		sastConfig.Values["languageMode"] = c.config.LanguageMode
		sastConfigString = sastConfigString + fmt.Sprintf(", languageMode %v", c.config.LanguageMode)
	}

	branch, _, _ := c.GetScanBranch()

	sastConfigString = fmt.Sprintf("Cx1 Branch name %v, ", branch) + sastConfigString

	log.Entry().Infof("Will run a scan with the following configuration: %v", sastConfigString)

	configs := []checkmarxOne.ScanConfiguration{sastConfig}

	// add scan's tags
	tags := make(map[string]string, 0)
	if len(c.config.ScanTags) > 0 {
		err := json.Unmarshal([]byte(c.config.ScanTags), &tags)
		if err != nil {
			log.Entry().WithError(err).Warnf("Failed to parse the scan tags: %v", c.config.ScanTags)
		}
	}

	// add more engines
	scan, err := c.sys.ScanProjectZip(c.Project.ProjectID, uploadLink, branch, configs, tags)

	if err != nil {
		return nil, fmt.Errorf("Failed to run scan on project %v: %s", c.Project.Name, err)
	}

	log.Entry().Debugf("Scanning project %v: %v ", c.Project.Name, scan.ScanID)

	return &scan, nil
}

func (c *checkmarxOneExecuteScanHelper) PollScanStatus(scan *checkmarxOne.Scan) error {
	statusDetails := "Scan phase: New"
	pastStatusDetails := statusDetails
	log.Entry().Info(statusDetails)
	status := "New"
	for {
		scan_refresh, err := c.sys.GetScan(scan.ScanID)

		if err != nil {
			return fmt.Errorf("Error while polling scan %v: %s", scan.ScanID, err)
		}

		status = scan_refresh.Status
		workflow, err := c.sys.GetScanWorkflow(scan.ScanID)
		if err != nil {
			return fmt.Errorf("Error while getting workflow for scan %v: %s", scan.ScanID, err)
		}

		statusDetails = workflow[len(workflow)-1].Info

		if pastStatusDetails != statusDetails {
			log.Entry().Info(statusDetails)
			pastStatusDetails = statusDetails
		}

		if status == "Completed" || status == "Canceled" || status == "Failed" {
			break
		}

		if pastStatusDetails != statusDetails {
			log.Entry().Info(statusDetails)
			pastStatusDetails = statusDetails
		}

		log.Entry().Debug("Polling for status: sleeping...")

		time.Sleep(10 * time.Second)
	}
	if status == "Canceled" {
		log.SetErrorCategory(log.ErrorCustom)
		return fmt.Errorf("Scan %v canceled via web interface", scan.ScanID)
	}
	if status == "Failed" {
		return fmt.Errorf("Checkmarx One scan failed with the following error: %v", statusDetails)
	}
	return nil
}

func (c *checkmarxOneExecuteScanHelper) PostScanSummaryInPullRequest(detailedResults *map[string]interface{}, insecure bool) error {
	cicdOrch, err := orchestrator.GetOrchestratorConfigProvider(nil)
	if err != nil {
		return fmt.Errorf("Failed to get orchestrator config provider: %s", err)
	}
	isPullRequest := cicdOrch.IsPullRequest()
	pullRequestId := cicdOrch.PullRequestConfig().Key
	var owner, repository string
	if len(c.config.Repository) == 0 || len(c.config.Owner) == 0 {
		log.Entry().Debug("No repository or owner configured, trying to get it from orchestrator")
		repoUrl := cicdOrch.RepoURL()
		if repoUrl != "n/a" {
			parsedURL, err := url.Parse(repoUrl)
			if err != nil {
				return fmt.Errorf("failed to parse repository URL %s: %s", repoUrl, err)
			}
			pathParts := strings.Split(strings.TrimSuffix(parsedURL.Path, ".git"), "/")
			if len(pathParts) >= 2 {
				if len(c.config.Owner) == 0 {
					owner = pathParts[len(pathParts)-2]
				}
				if len(c.config.Repository) == 0 {
					repository = pathParts[len(pathParts)-1]
				}
				log.Entry().Infof("Found repository %s and owner %s from orchestrator", c.config.Repository, c.config.Owner)
			} else {
				return fmt.Errorf("failed to extract owner and repository from URL %s", repoUrl)
			}
		} else {
			log.Entry().Debug("Could not retrieve repository URL from orchestrator")
		}
	} else {
		owner = c.config.Owner
		repository = c.config.Repository
		log.Entry().Debug("Using Owner and Repository from configuration: " + owner + "/" + repository)
	}
	log.Entry().Debugf("Parameters for PR summary: ScanSummaryInPullRequest: %t, isPullRequest: %t, pullRequestId: %s, PullRequestName: %s, GithubAPIURL: %s, GithubToken: %s, Owner: %s, Repository: %s", c.config.ScanSummaryInPullRequest, isPullRequest, pullRequestId, c.config.PullRequestName, c.config.GithubAPIURL, c.config.GithubToken, owner, repository)
	if c.config.ScanSummaryInPullRequest && isPullRequest && pullRequestId != "n/a" && len(c.config.GithubToken) > 0 && len(c.config.GithubAPIURL) > 0 && len(owner) > 0 && len(repository) > 0 {
		ghIssues := c.utils.GetIssueService()
		log.Entry().Debugf("Creating/updating GitHub issue with check results with PR: %s, GithubAPIURL: %s, Owner: %s, Repository: %s", c.config.PullRequestName, c.config.GithubAPIURL, owner, repository)
		scanReportOverview := checkmarxOne.CreateJSONHeaderReport(detailedResults)
		var criticalSeverityString, highSeverityString, mediumSeverityString, lowSeverityString string
		for _, finding := range *scanReportOverview.Findings {
			switch finding.ClassificationName {
			case "Critical":
				criticalSeverityString = fmt.Sprintf("%d", finding.Total-*finding.Audited)
			case "High":
				highSeverityString = fmt.Sprintf("%d", finding.Total-*finding.Audited)
			case "Medium":
				mediumSeverityString = fmt.Sprintf("%d", finding.Total-*finding.Audited)
			case "Low":
				if finding.LowPerQuery != nil {
					for _, lowFinding := range *finding.LowPerQuery {
						if c.config.VulnerabilityThresholdLowPerQuery {
							lowAuditedRequiredPerQuery := min(int(math.Ceil(float64(lowFinding.Total)*float64(c.config.VulnerabilityThresholdLow)/100.0)), c.config.VulnerabilityThresholdLowPerQueryMax)
							lowSeverityString = fmt.Sprintf("%s%d %s (%d audits / %d required) <br>", lowSeverityString, lowFinding.Total-lowFinding.Audited, lowFinding.QueryName, lowFinding.Audited, lowAuditedRequiredPerQuery)
						} else {
							lowSeverityString = fmt.Sprintf("%s%d %s<br>", lowSeverityString, lowFinding.Total-lowFinding.Audited, lowFinding.QueryName)
						}
					}
				}
			}
		}
		var scanIcon string
		if insecure {
			scanIcon = ":x:"
		} else {
			scanIcon = ":white_check_mark:"
		}
		comment := &github.IssueComment{
			Body: github.Ptr(fmt.Sprintf(`# %s Checkmarx %s scan completed 
**Project**: %s
**ScanId**: %s
**Preset**: %s
Severity | Number of open findings
--- | ---
:bangbang: Critical | %s
:red_circle: High | %s
:orange_circle: Medium | %s
:yellow_circle: Low | %s

[Go to the scan results](%s)
		`, scanIcon, strings.ToLower(scanReportOverview.ScanType), c.Project.Name, scanReportOverview.ScanID, scanReportOverview.Preset, criticalSeverityString, highSeverityString, mediumSeverityString, lowSeverityString, scanReportOverview.DeepLink)),
		}
		pullRequestNumber, err := strconv.Atoi(pullRequestId)
		if err != nil {
			return fmt.Errorf("failed to parse int from pull request name %s: %s", c.config.PullRequestName, err)
		}
		_, _, err = ghIssues.CreateComment(c.ctx, owner, repository, pullRequestNumber, comment)
		if err != nil {
			return fmt.Errorf("failed to create GitHub issue comment: %s", err)
		}
		log.Entry().Infof("Created GitHub issue comment for project %v", c.Project.Name)
	} else {
		log.Entry().Debug("Skipping GitHub issue comment creation, no pull request or GitHub configuration provided")
	}
	return nil
}

func (c *checkmarxOneExecuteScanHelper) CheckCompliance(scan *checkmarxOne.Scan, detailedResults *map[string]interface{}) error {
	links := []piperutils.Path{{Target: (*detailedResults)["DeepLink"].(string), Name: "Checkmarx One Web UI"}}
	insecure := false
	var insecureResults []string
	var neutralResults []string

	if c.config.VulnerabilityThresholdEnabled {
		insecure, insecureResults, neutralResults = c.enforceThresholds(detailedResults)
		scanReport := checkmarxOne.CreateCustomReport(detailedResults, insecureResults, neutralResults)

		// Create scan summary comment in PR
		if c.config.ScanSummaryInPullRequest {
			err := c.PostScanSummaryInPullRequest(detailedResults, insecure)
			if err != nil {
				log.Entry().Errorf("failed to post scan summary in pull request: %s", err)
			}
		}

		if insecure && c.config.CreateResultIssue && len(c.config.GithubToken) > 0 && len(c.config.GithubAPIURL) > 0 && len(c.config.Owner) > 0 && len(c.config.Repository) > 0 {
			log.Entry().Debug("Creating/updating GitHub issue with check results")
			gh := reporting.GitHub{
				Owner:         &c.config.Owner,
				Repository:    &c.config.Repository,
				Assignees:     &c.config.Assignees,
				IssueService:  c.utils.GetIssueService(),
				SearchService: c.utils.GetSearchService(),
			}
			if err := gh.UploadSingleReport(c.ctx, scanReport); err != nil {
				return fmt.Errorf("failed to upload scan results into GitHub: %s", err)
			}
		}

		paths, err := checkmarxOne.WriteCustomReports(scanReport, c.Project.Name, c.Project.ProjectID)
		if err != nil {
			// do not fail until we have a better idea to handle it
			log.Entry().Warning("failed to write HTML/MarkDown report file ...", err)
		} else {
			c.reports = append(c.reports, paths...)
		}
	}

	piperutils.PersistReportsAndLinks("checkmarxOneExecuteScan", c.utils.GetWorkspace(), c.utils, c.reports, links)

	c.reportToInflux(detailedResults)

	if insecure {
		if c.config.VulnerabilityThresholdResult == "FAILURE" {
			log.SetErrorCategory(log.ErrorCompliance)
			return fmt.Errorf("the project is not compliant - see report for details")
		}
		log.Entry().Errorf("Checkmarx One scan result set to %v, some results are not meeting defined thresholds. For details see the archived report.", c.config.VulnerabilityThresholdResult)
	} else {
		log.Entry().Infoln("Checkmarx One scan finished successfully")
	}
	return nil
}

func (c *checkmarxOneExecuteScanHelper) GetReportPDF(scan *checkmarxOne.Scan) error {
	if c.config.GeneratePdfReport {
		pdfReportName := c.createReportName(c.utils.GetWorkspace(), "Cx1_SASTReport_%v.pdf")
		err := c.downloadAndSaveReport(pdfReportName, scan, "pdf")
		if err != nil {
			return fmt.Errorf("Report download failed: %s", err)
		} else {
			c.reports = append(c.reports, piperutils.Path{Target: pdfReportName, Mandatory: true})
		}
	} else {
		log.Entry().Debug("Report generation is disabled via configuration")
	}

	return nil
}

func (c *checkmarxOneExecuteScanHelper) GetReportSARIF(scan *checkmarxOne.Scan, scanmeta *checkmarxOne.ScanMetadata, results *[]checkmarxOne.ScanResult) error {
	if c.config.ConvertToSarif {
		log.Entry().Info("Calling conversion to SARIF function.")
		sarif, err := checkmarxOne.ConvertCxJSONToSarif(c.sys, c.config.ServerURL, results, scanmeta, scan)
		if err != nil {
			return fmt.Errorf("Failed to generate SARIF: %s", err)
		}
		paths, err := checkmarxOne.WriteSarif(sarif)
		if err != nil {
			return fmt.Errorf("Failed to write SARIF: %s", err)
		}
		c.reports = append(c.reports, paths...)
	}
	return nil
}

func (c *checkmarxOneExecuteScanHelper) GetReportJSON(scan *checkmarxOne.Scan) error {
	jsonReportName := c.createReportName(c.utils.GetWorkspace(), "Cx1_SASTReport_%v.json")
	err := c.downloadAndSaveReport(jsonReportName, scan, "json")
	if err != nil {
		return fmt.Errorf("Report download failed: %s", err)
	} else {
		c.reports = append(c.reports, piperutils.Path{Target: jsonReportName, Mandatory: true})
	}
	return nil
}

func (c *checkmarxOneExecuteScanHelper) GetHeaderReportJSON(detailedResults *map[string]interface{}) error {
	// This is for the SAP-piper-format short-form JSON report
	jsonReport := checkmarxOne.CreateJSONHeaderReport(detailedResults)
	paths, err := checkmarxOne.WriteJSONHeaderReport(jsonReport)
	if err != nil {
		return fmt.Errorf("Failed to write JSON header report: %s", err)
	} else {
		// add JSON report to archiving list
		c.reports = append(c.reports, paths...)
	}
	return nil
}

func (c *checkmarxOneExecuteScanHelper) ParseResults(scan *checkmarxOne.Scan) (map[string]interface{}, error) {
	var detailedResults map[string]interface{}

	scanmeta, err := c.sys.GetScanMetadata(scan.ScanID)
	if err != nil {
		return detailedResults, fmt.Errorf("Unable to fetch scan metadata for scan %v: %s", scan.ScanID, err)
	}

	totalResultCount := uint64(0)

	scansummary, err := c.sys.GetScanSummary(scan.ScanID)
	if err != nil {
		/* TODO: scansummary throws a 404 for 0-result scans, once the bug is fixed put this code back. */
		// return detailedResults, fmt.Errorf("Unable to fetch scan summary for scan %v: %s", scan.ScanID, err)
	} else {
		totalResultCount = scansummary.TotalCount()
	}

	results, err := c.sys.GetScanResults(scan.ScanID, totalResultCount)
	if err != nil {
		return detailedResults, fmt.Errorf("Unable to fetch scan results for scan %v: %s", scan.ScanID, err)
	}

	detailedResults, err = c.getDetailedResults(scan, &scanmeta, &results)
	if err != nil {
		return detailedResults, fmt.Errorf("Unable to fetch detailed results for scan %v: %s", scan.ScanID, err)
	}

	err = c.GetReportJSON(scan)
	if err != nil {
		log.Entry().WithError(err).Warnf("Failed to get JSON report")
	}
	err = c.GetReportPDF(scan)
	if err != nil {
		log.Entry().WithError(err).Warnf("Failed to get PDF report")
	}
	err = c.GetReportSARIF(scan, &scanmeta, &results)
	if err != nil {
		log.Entry().WithError(err).Warnf("Failed to get SARIF report")
	}
	err = c.GetHeaderReportJSON(&detailedResults)
	if err != nil {
		log.Entry().WithError(err).Warnf("Failed to generate JSON Header report")
	}

	// create toolrecord
	toolRecordFileName, err := c.createToolRecordCx(&detailedResults)
	if err != nil {
		// do not fail until the framework is well established
		log.Entry().Warning("TR_CHECKMARXONE: Failed to create toolrecord file ...", err)
	} else {
		c.reports = append(c.reports, piperutils.Path{Target: toolRecordFileName})
	}

	return detailedResults, nil
}

func (c *checkmarxOneExecuteScanHelper) createReportName(workspace, reportFileNameTemplate string) string {
	regExpFileName := regexp.MustCompile(`[^\w\d]`)
	timeStamp, _ := time.Now().Local().MarshalText()
	return filepath.Join(workspace, fmt.Sprintf(reportFileNameTemplate, regExpFileName.ReplaceAllString(string(timeStamp), "_")))
}

func (c *checkmarxOneExecuteScanHelper) downloadAndSaveReport(reportFileName string, scan *checkmarxOne.Scan, reportType string) error {
	report, err := c.generateAndDownloadReport(scan, reportType)
	if err != nil {
		return errors.Wrap(err, "failed to download the report")
	}
	log.Entry().Debugf("Saving report to file %v...", reportFileName)
	return c.utils.WriteFile(reportFileName, report, 0o700)
}

func (c *checkmarxOneExecuteScanHelper) generateAndDownloadReport(scan *checkmarxOne.Scan, reportType string) ([]byte, error) {
	var finalStatus checkmarxOne.ReportStatus

	report, err := c.sys.RequestNewReport(scan.ScanID, scan.ProjectID, scan.Branch, reportType)
	if err != nil {
		return []byte{}, errors.Wrap(err, "failed to request new report")
	}
	for {
		finalStatus, err = c.sys.GetReportStatus(report)
		if err != nil {
			return []byte{}, errors.Wrap(err, "failed to get report status")
		}

		if finalStatus.Status == "completed" {
			break
		} else if finalStatus.Status == "failed" {
			return []byte{}, fmt.Errorf("report generation failed")
		}
		time.Sleep(10 * time.Second)
	}
	if finalStatus.Status == "completed" {
		return c.sys.DownloadReport(finalStatus.ReportURL)
	}

	return []byte{}, fmt.Errorf("unexpected status %v recieved", finalStatus.Status)
}

func (c *checkmarxOneExecuteScanHelper) getNumCoherentIncrementalScans(scans []checkmarxOne.Scan) int {
	count := 0
	for _, scan := range scans {
		inc, err := scan.IsIncremental()
		if !inc && err == nil {
			break
		}
		count++
	}
	return count
}

func (c *checkmarxOneExecuteScanHelper) getDetailedResults(scan *checkmarxOne.Scan, scanmeta *checkmarxOne.ScanMetadata, results *[]checkmarxOne.ScanResult) (map[string]interface{}, error) {
	// this converts the JSON format results from Cx1 into the "resultMap" structure used in other parts of this step (influx etc)

	resultMap := map[string]interface{}{}
	resultMap["InitiatorName"] = scan.Initiator
	resultMap["Owner"] = "Cx1 Gap: no project owner" // TODO: check for functionality
	resultMap["ScanId"] = scan.ScanID
	resultMap["ProjectId"] = c.Project.ProjectID
	resultMap["ProjectName"] = c.Project.Name

	resultMap["Group"] = ""
	resultMap["GroupFullPathOnReportDate"] = ""

	if c.App != nil {
		resultMap["Application"] = c.App.ApplicationID
		resultMap["ApplicationFullPathOnReportDate"] = c.App.Name
	} else {
		resultMap["Application"] = ""
		resultMap["ApplicationFullPathOnReportDate"] = ""
	}

	resultMap["ScanStart"] = scan.CreatedAt

	scanCreated, err := time.Parse(time.RFC3339, scan.CreatedAt)
	if err != nil {
		log.Entry().Warningf("Failed to parse string %v into time: %s", scan.CreatedAt, err)
		resultMap["ScanTime"] = "Error parsing scan.CreatedAt"
	} else {
		scanFinished, err := time.Parse(time.RFC3339, scan.UpdatedAt)
		if err != nil {
			log.Entry().Warningf("Failed to parse string %v into time: %s", scan.UpdatedAt, err)
			resultMap["ScanTime"] = "Error parsing scan.UpdatedAt"
		} else {
			difference := scanFinished.Sub(scanCreated)
			resultMap["ScanTime"] = difference.String()
		}
	}

	resultMap["LinesOfCodeScanned"] = scanmeta.LOC
	resultMap["FilesScanned"] = scanmeta.FileCount

	version, err := c.sys.GetVersion()
	if err != nil {
		resultMap["ToolVersion"] = "Error fetching current version"
	} else {
		resultMap["ToolVersion"] = fmt.Sprintf("CxOne: %v, SAST: %v, KICS: %v", version.CxOne, version.SAST, version.KICS)
	}

	if scanmeta.IsIncremental {
		resultMap["ScanType"] = "Incremental"
	} else {
		resultMap["ScanType"] = "Full"
	}

	resultMap["Preset"] = scanmeta.PresetName
	resultMap["DeepLink"] = fmt.Sprintf("%v/projects/%v/overview?branch=%v", c.config.ServerURL, c.Project.ProjectID, url.QueryEscape(scan.Branch))
	resultMap["ReportCreationTime"] = time.Now().String()
	resultMap["Critical"] = map[string]int{}
	resultMap["High"] = map[string]int{}
	resultMap["Medium"] = map[string]int{}
	resultMap["Low"] = map[string]int{}
	resultMap["Information"] = map[string]int{}

	if len(*results) > 0 {
		for _, result := range *results {
			key := "Information"
			switch result.Severity {
			case "CRITICAL":
				key = "Critical"
			case "HIGH":
				key = "High"
			case "MEDIUM":
				key = "Medium"
			case "LOW":
				key = "Low"
			case "INFORMATION":
			default:
				key = "Information"
			}

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
			case "NOT_EXPLOITABLE":
				auditState = "NotExploitable"
			case "CONFIRMED":
				auditState = "Confirmed"
			case "URGENT", "URGENT ":
				auditState = "Urgent"
			case "PROPOSED_NOT_EXPLOITABLE":
				auditState = "ProposedNotExploitable"
			case "TO_VERIFY":
			default:
				auditState = "ToVerify"
			}
			submap[auditState]++

			if auditState != "NotExploitable" {
				submap["NotFalsePositive"]++
			}

		}

		// if the flag is switched on, build the list  of Low findings per query
		if c.config.VulnerabilityThresholdLowPerQuery {
			var lowPerQuery = map[string]map[string]int{}

			for _, result := range *results {
				if result.Severity != "LOW" {
					continue
				}
				key := result.Data.QueryName
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
				case "NOT_EXPLOITABLE":
					auditState = "NotExploitable"
				case "CONFIRMED":
					auditState = "Confirmed"
				case "URGENT", "URGENT ":
					auditState = "Urgent"
				case "PROPOSED_NOT_EXPLOITABLE":
					auditState = "ProposedNotExploitable"
				case "TO_VERIFY":
				default:
					auditState = "ToVerify"
				}
				submap[auditState]++

				if auditState != "NotExploitable" {
					submap["NotFalsePositive"]++
				}
			}

			resultMap["LowPerQuery"] = lowPerQuery
		}
	}
	return resultMap, nil
}

func (c *checkmarxOneExecuteScanHelper) zipWorkspaceFiles(filterPattern string, utils checkmarxOneExecuteScanUtils) (*os.File, error) {
	zipFileName := filepath.Join(utils.GetWorkspace(), "workspace.zip")
	log.Entry().Infof("Zipping files using filter: %v", filterPattern)
	patterns := piperutils.Trim(strings.Split(filterPattern, ","))
	sort.Strings(patterns)
	zipFile, err := os.Create(zipFileName)
	if err != nil {
		return zipFile, errors.Wrap(err, "failed to create archive of project sources")
	}
	defer zipFile.Close()

	err = c.zipFolder(utils.GetWorkspace(), zipFile, patterns, utils)
	if err != nil {
		return nil, errors.Wrap(err, "failed to compact folder")
	}
	return zipFile, nil
}

func (c *checkmarxOneExecuteScanHelper) zipFolder(source string, zipFile io.Writer, patterns []string, utils checkmarxOneExecuteScanUtils) error {
	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	log.Entry().Infof("Zipping %v into workspace.zip", source)

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

		if !info.Mode().IsRegular() || info.Size() == 0 {
			return nil
		}

		fileName := strings.TrimPrefix(path, baseDir)
		noMatch, err := c.isFileNotMatchingPattern(patterns, path, info, utils)
		if err != nil || noMatch {
			if noMatch {
				log.Entry().Debugf("Excluded %s", fileName)
			}
			return err
		}

		writer, err := archive.Create(fileName)
		if err != nil {
			return err
		}

		file, err := utils.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		if err == nil {
			log.Entry().Debugf("Zipped %s", fileName)
		}
		fileCount++
		return err
	})
	log.Entry().Infof("Zipped %d files", fileCount)
	err = c.handleZeroFilesZipped(source, err, fileCount)
	return err
}

func (c *checkmarxOneExecuteScanHelper) adaptHeader(info os.FileInfo, header *zip.FileHeader) {
	if info.IsDir() {
		header.Name += "/"
	} else {
		header.Method = zip.Deflate
	}
}

func (c *checkmarxOneExecuteScanHelper) handleZeroFilesZipped(source string, err error, fileCount int) error {
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
func (c *checkmarxOneExecuteScanHelper) isFileNotMatchingPattern(patterns []string, path string, info os.FileInfo, utils checkmarxOneExecuteScanUtils) (bool, error) {
	if len(patterns) == 0 || info.IsDir() {
		return false, nil
	}

	// Check if it is matched by at least one include pattern
	includeMatch := false
	for _, pattern := range patterns {
		if strings.HasPrefix(pattern, "!") {
			continue
		}
		match, err := utils.PathMatch(pattern, path)
		if err != nil {
			return false, errors.Wrapf(err, "Pattern %v could not get executed", pattern)
		}
		if match {
			includeMatch = true
			break
		}
	}

	if !includeMatch {
		return true, nil // if there is no include pattern matching, the file is necessarily excluded
	}

	// Check if it is matched by at least one exclude pattern
	for _, pattern := range patterns {
		if strings.HasPrefix(pattern, "!") {
			pattern = strings.TrimLeft(pattern, "!")
		} else {
			continue
		}
		match, err := utils.PathMatch(pattern, path)
		if err != nil {
			return false, errors.Wrapf(err, "Pattern %v could not get executed", pattern)
		}

		if match { // match with an exclude pattern, the file is excluded
			return true, nil
		}
	}

	return false, nil
}

func (c *checkmarxOneExecuteScanHelper) createToolRecordCx(results *map[string]interface{}) (string, error) {
	workspace := c.utils.GetWorkspace()
	record := toolrecord.New(c.utils, workspace, "checkmarxOne", c.config.ServerURL)

	// Project
	err := record.AddKeyData("project",
		(*results)["ProjectId"].(string),
		(*results)["ProjectName"].(string),
		"")
	if err != nil {
		return "", err
	}
	// Scan
	err = record.AddKeyData("scanid",
		(*results)["ScanId"].(string),
		(*results)["ScanId"].(string),
		(*results)["DeepLink"].(string))
	if err != nil {
		return "", err
	}
	err = record.Persist()
	if err != nil {
		return "", err
	}
	return record.GetFileName(), nil
}

func (c *checkmarxOneExecuteScanHelper) enforceThresholds(results *map[string]interface{}) (bool, []string, []string) {
	neutralResults := []string{}
	insecureResults := []string{}
	insecure := false

	cxCriticalThreshold := c.config.VulnerabilityThresholdCritical
	cxHighThreshold := c.config.VulnerabilityThresholdHigh
	cxMediumThreshold := c.config.VulnerabilityThresholdMedium
	cxLowThreshold := c.config.VulnerabilityThresholdLow
	cxLowThresholdPerQuery := c.config.VulnerabilityThresholdLowPerQuery
	cxLowThresholdPerQueryMax := c.config.VulnerabilityThresholdLowPerQueryMax
	criticalValue := (*results)["Critical"].(map[string]int)["NotFalsePositive"]
	highValue := (*results)["High"].(map[string]int)["NotFalsePositive"]
	mediumValue := (*results)["Medium"].(map[string]int)["NotFalsePositive"]
	lowValue := (*results)["Low"].(map[string]int)["NotFalsePositive"]
	var unit string
	criticalViolation := ""
	highViolation := ""
	mediumViolation := ""
	lowViolation := ""
	if c.config.VulnerabilityThresholdUnit == "percentage" {
		unit = "%"
		criticalAudited := (*results)["Critical"].(map[string]int)["Issues"] - (*results)["Critical"].(map[string]int)["NotFalsePositive"]
		criticalOverall := (*results)["Critical"].(map[string]int)["Issues"]
		if criticalOverall == 0 {
			criticalAudited = 1
			criticalOverall = 1
		}
		highAudited := (*results)["High"].(map[string]int)["Issues"] - (*results)["High"].(map[string]int)["NotFalsePositive"]
		highOverall := (*results)["High"].(map[string]int)["Issues"]
		if highOverall == 0 {
			highAudited = 1
			highOverall = 1
		}
		mediumAudited := (*results)["Medium"].(map[string]int)["Issues"] - (*results)["Medium"].(map[string]int)["NotFalsePositive"]
		mediumOverall := (*results)["Medium"].(map[string]int)["Issues"]
		if mediumOverall == 0 {
			mediumAudited = 1
			mediumOverall = 1
		}
		lowAudited := (*results)["Low"].(map[string]int)["Confirmed"] + (*results)["Low"].(map[string]int)["NotExploitable"]
		lowOverall := (*results)["Low"].(map[string]int)["Issues"]
		if lowOverall == 0 {
			lowAudited = 1
			lowOverall = 1
		}
		criticalValue = int(float32(criticalAudited) / float32(criticalOverall) * 100.0)
		highValue = int(float32(highAudited) / float32(highOverall) * 100.0)
		mediumValue = int(float32(mediumAudited) / float32(mediumOverall) * 100.0)
		lowValue = int(float32(lowAudited) / float32(lowOverall) * 100.0)

		if criticalValue < cxCriticalThreshold {
			insecure = true
			criticalViolation = fmt.Sprintf("<-- %v %v deviation", cxCriticalThreshold-criticalValue, unit)
		}
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
			if (*results)["LowPerQuery"] != nil {
				lowPerQueryMap := (*results)["LowPerQuery"].(map[string]map[string]int)

				for lowQuery, resultsLowQuery := range lowPerQueryMap {
					lowAuditedPerQuery := resultsLowQuery["Confirmed"] + resultsLowQuery["NotExploitable"]
					lowOverallPerQuery := resultsLowQuery["Issues"]
					lowAuditedRequiredPerQuery := min(int(math.Ceil(float64(lowOverallPerQuery)*float64(cxLowThreshold)/100.0)), cxLowThresholdPerQueryMax)
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
	if c.config.VulnerabilityThresholdUnit == "absolute" {
		unit = " findings"
		if criticalValue > cxCriticalThreshold {
			insecure = true
			criticalViolation = fmt.Sprintf("<-- %v%v deviation", criticalValue-cxCriticalThreshold, unit)
		}
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

	criticalText := fmt.Sprintf("Critical %v%v %v", criticalValue, unit, criticalViolation)
	highText := fmt.Sprintf("High %v%v %v", highValue, unit, highViolation)
	mediumText := fmt.Sprintf("Medium %v%v %v", mediumValue, unit, mediumViolation)
	lowText := fmt.Sprintf("Low %v%v %v", lowValue, unit, lowViolation)
	if len(criticalViolation) > 0 {
		insecureResults = append(insecureResults, criticalText)
		log.Entry().Error(criticalText)
	} else {
		neutralResults = append(neutralResults, criticalText)
		log.Entry().Info(criticalText)
	}
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

func (c *checkmarxOneExecuteScanHelper) reportToInflux(results *map[string]interface{}) {
	c.influx.checkmarxOne_data.fields.critical_issues = (*results)["Critical"].(map[string]int)["Issues"]
	c.influx.checkmarxOne_data.fields.critical_not_false_postive = (*results)["Critical"].(map[string]int)["NotFalsePositive"]
	c.influx.checkmarxOne_data.fields.critical_not_exploitable = (*results)["Critical"].(map[string]int)["NotExploitable"]
	c.influx.checkmarxOne_data.fields.critical_confirmed = (*results)["Critical"].(map[string]int)["Confirmed"]
	c.influx.checkmarxOne_data.fields.critical_urgent = (*results)["Critical"].(map[string]int)["Urgent"]
	c.influx.checkmarxOne_data.fields.critical_proposed_not_exploitable = (*results)["Critical"].(map[string]int)["ProposedNotExploitable"]
	c.influx.checkmarxOne_data.fields.critical_to_verify = (*results)["Critical"].(map[string]int)["ToVerify"]

	c.influx.checkmarxOne_data.fields.high_issues = (*results)["High"].(map[string]int)["Issues"]
	c.influx.checkmarxOne_data.fields.high_not_false_postive = (*results)["High"].(map[string]int)["NotFalsePositive"]
	c.influx.checkmarxOne_data.fields.high_not_exploitable = (*results)["High"].(map[string]int)["NotExploitable"]
	c.influx.checkmarxOne_data.fields.high_confirmed = (*results)["High"].(map[string]int)["Confirmed"]
	c.influx.checkmarxOne_data.fields.high_urgent = (*results)["High"].(map[string]int)["Urgent"]
	c.influx.checkmarxOne_data.fields.high_proposed_not_exploitable = (*results)["High"].(map[string]int)["ProposedNotExploitable"]
	c.influx.checkmarxOne_data.fields.high_to_verify = (*results)["High"].(map[string]int)["ToVerify"]
	c.influx.checkmarxOne_data.fields.medium_issues = (*results)["Medium"].(map[string]int)["Issues"]
	c.influx.checkmarxOne_data.fields.medium_not_false_postive = (*results)["Medium"].(map[string]int)["NotFalsePositive"]
	c.influx.checkmarxOne_data.fields.medium_not_exploitable = (*results)["Medium"].(map[string]int)["NotExploitable"]
	c.influx.checkmarxOne_data.fields.medium_confirmed = (*results)["Medium"].(map[string]int)["Confirmed"]
	c.influx.checkmarxOne_data.fields.medium_urgent = (*results)["Medium"].(map[string]int)["Urgent"]
	c.influx.checkmarxOne_data.fields.medium_proposed_not_exploitable = (*results)["Medium"].(map[string]int)["ProposedNotExploitable"]
	c.influx.checkmarxOne_data.fields.medium_to_verify = (*results)["Medium"].(map[string]int)["ToVerify"]
	c.influx.checkmarxOne_data.fields.low_issues = (*results)["Low"].(map[string]int)["Issues"]
	c.influx.checkmarxOne_data.fields.low_not_false_postive = (*results)["Low"].(map[string]int)["NotFalsePositive"]
	c.influx.checkmarxOne_data.fields.low_not_exploitable = (*results)["Low"].(map[string]int)["NotExploitable"]
	c.influx.checkmarxOne_data.fields.low_confirmed = (*results)["Low"].(map[string]int)["Confirmed"]
	c.influx.checkmarxOne_data.fields.low_urgent = (*results)["Low"].(map[string]int)["Urgent"]
	c.influx.checkmarxOne_data.fields.low_proposed_not_exploitable = (*results)["Low"].(map[string]int)["ProposedNotExploitable"]
	c.influx.checkmarxOne_data.fields.low_to_verify = (*results)["Low"].(map[string]int)["ToVerify"]
	c.influx.checkmarxOne_data.fields.information_issues = (*results)["Information"].(map[string]int)["Issues"]
	c.influx.checkmarxOne_data.fields.information_not_false_postive = (*results)["Information"].(map[string]int)["NotFalsePositive"]
	c.influx.checkmarxOne_data.fields.information_not_exploitable = (*results)["Information"].(map[string]int)["NotExploitable"]
	c.influx.checkmarxOne_data.fields.information_confirmed = (*results)["Information"].(map[string]int)["Confirmed"]
	c.influx.checkmarxOne_data.fields.information_urgent = (*results)["Information"].(map[string]int)["Urgent"]
	c.influx.checkmarxOne_data.fields.information_proposed_not_exploitable = (*results)["Information"].(map[string]int)["ProposedNotExploitable"]
	c.influx.checkmarxOne_data.fields.information_to_verify = (*results)["Information"].(map[string]int)["ToVerify"]
	c.influx.checkmarxOne_data.fields.initiator_name = (*results)["InitiatorName"].(string)
	c.influx.checkmarxOne_data.fields.owner = (*results)["Owner"].(string)
	c.influx.checkmarxOne_data.fields.scan_id = (*results)["ScanId"].(string)
	c.influx.checkmarxOne_data.fields.project_id = (*results)["ProjectId"].(string)
	c.influx.checkmarxOne_data.fields.projectName = (*results)["ProjectName"].(string)
	c.influx.checkmarxOne_data.fields.group = (*results)["Group"].(string)
	c.influx.checkmarxOne_data.fields.group_full_path_on_report_date = (*results)["GroupFullPathOnReportDate"].(string)
	c.influx.checkmarxOne_data.fields.scan_start = (*results)["ScanStart"].(string)
	c.influx.checkmarxOne_data.fields.scan_time = (*results)["ScanTime"].(string)
	c.influx.checkmarxOne_data.fields.lines_of_code_scanned = (*results)["LinesOfCodeScanned"].(int)
	c.influx.checkmarxOne_data.fields.files_scanned = (*results)["FilesScanned"].(int)
	c.influx.checkmarxOne_data.fields.tool_version = (*results)["ToolVersion"].(string)

	c.influx.checkmarxOne_data.fields.scan_type = (*results)["ScanType"].(string)
	c.influx.checkmarxOne_data.fields.preset = (*results)["Preset"].(string)
	c.influx.checkmarxOne_data.fields.deep_link = (*results)["DeepLink"].(string)
	c.influx.checkmarxOne_data.fields.report_creation_time = (*results)["ReportCreationTime"].(string)
}

// Utils Bundle
// various utilities to set up or work with the workspace and prepare data to send to Cx1

func (c *checkmarxOneExecuteScanUtilsBundle) PathMatch(pattern, name string) (bool, error) {
	return doublestar.PathMatch(pattern, name)
}

func (c *checkmarxOneExecuteScanUtilsBundle) GetWorkspace() string {
	return c.workspace
}

func (c *checkmarxOneExecuteScanUtilsBundle) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}

func (c *checkmarxOneExecuteScanUtilsBundle) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (c *checkmarxOneExecuteScanUtilsBundle) FileInfoHeader(fi os.FileInfo) (*zip.FileHeader, error) {
	return zip.FileInfoHeader(fi)
}

func (c *checkmarxOneExecuteScanUtilsBundle) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (c *checkmarxOneExecuteScanUtilsBundle) Open(name string) (*os.File, error) {
	return os.Open(name)
}

func (c *checkmarxOneExecuteScanUtilsBundle) CreateIssue(ghCreateIssueOptions *piperGithub.CreateIssueOptions) error {
	_, err := piperGithub.CreateIssue(ghCreateIssueOptions)
	return err
}

func (c *checkmarxOneExecuteScanUtilsBundle) GetIssueService() *github.IssuesService {
	return c.issues
}

func (c *checkmarxOneExecuteScanUtilsBundle) GetSearchService() *github.SearchService {
	return c.search
}

func newcheckmarxOneExecuteScanUtilsBundle(workspace string, client *github.Client) checkmarxOneExecuteScanUtils {
	utils := checkmarxOneExecuteScanUtilsBundle{
		workspace: workspace,
	}
	if client != nil {
		utils.issues = client.Issues
		utils.search = client.Search
	}
	return &utils
}
