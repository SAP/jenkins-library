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

	"github.com/bmatcuk/doublestar/v4"
	"github.com/google/go-github/v68/github"
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
	ctx      context.Context
	config   checkmarxOneExecuteScanOptions
	sys      checkmarxOne.System
	influx   *checkmarxOneExecuteScanInflux
	utils    checkmarxOneExecuteScanUtils
	Project  *checkmarxOne.Project
	Group    *checkmarxOne.Group
	App      *checkmarxOne.Application
	ScanSAST bool
	ScanIAC  bool
	reports  []piperutils.Path
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

	// if this is an IaC scan, load the help links from json
	if cx1sh.ScanIAC {
		if config.IacHelpLinks == "" {
			log.Entry().Infof("No iacHelpLinks parameter provided - will use the backend API. Please consider providing this parameter to reduce server load.")
		} else {
			if err = cx1sh.sys.LoadIACHelpLinks(config.IacHelpLinks); err != nil {
				log.Entry().WithError(err).Errorf("failed to load IAC help links from %s, will use the backend API.", config.IacHelpLinks)
			}
		}
	}

	if len(cx1sh.config.ProjectID) == 0 {
		cx1sh.Project, err = cx1sh.GetProjectByName()
		if err != nil && err.Error() != "project not found" {
			return fmt.Errorf("failed to get project: %s", err)
		}
	} else {
		cx1sh.Project, err = cx1sh.GetProjectByID(cx1sh.config.ProjectID)
		if err != nil {
			return fmt.Errorf("failed to get project by ID: %s", err)
		}
	}

	if len(config.GroupName) > 0 {
		cx1sh.Group, err = cx1sh.GetGroup() // used when creating a project and when generating a SARIF report
		if err != nil {
			log.Entry().WithError(err).Warnf("failed to get group")
		}
	}

	if cx1sh.Project == nil {
		if len(config.ApplicationID) > 0 {
			cx1sh.App, err = cx1sh.GetApplicationByID(config.ApplicationID)
			if err != nil {
				return fmt.Errorf("failed to get application by ID: %v", err)
			}
		} else if len(config.ApplicationName) > 0 {
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

	fullScanCycle, err := strconv.Atoi(cx1sh.config.FullScanCycle)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return fmt.Errorf("invalid configuration value for fullScanCycle %v, must be a positive int", cx1sh.config.FullScanCycle)
	}
	branch, isPR, baseBranch := cx1sh.GetScanBranch()
	scans, err := cx1sh.GetLastScans(fullScanCycle+1, branch)
	if err != nil {
		log.Entry().WithError(err).Warnf("failed to get last 10 scans")
	}

	if config.VerifyOnly {
		if len(scans) > 0 {
			return cx1sh.CheckScanCompliance(&scans[0])
		} else {
			return fmt.Errorf("Cannot load scans for project %v, verification only mode aborted", cx1sh.Project.Name)
		}
	}

	err = cx1sh.SetProjectPresetsAndFilters()
	if err != nil {
		return fmt.Errorf("failed to set configuration: %s", err)
	}

	// update project's tags
	if (len(config.ProjectTags)) > 0 {
		err = cx1sh.UpdateProjectTags()
		if err != nil {
			log.Entry().WithError(err).Warnf("failed to tags the project: %s", err)
		}
	}

	incremental, fullScanExists, contiguousIncrScansCurrentBranch, err := cx1sh.IncrementalOrFull(scans) // requires: scan list
	if err != nil {
		return fmt.Errorf("failed to determine incremental or full scan configuration: %s", err)
	}

	if config.Incremental && cx1sh.ScanSAST {
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
	// user requested an incremental scan on a branch, and the project has a Primary Branch set, not in PR context and no full scan on the branch
	if config.Incremental && !isPR && !fullScanExists && cx1sh.Project.MainBranch != "" && cx1sh.Project.MainBranch != branch {
		scansMainBranch, err := cx1sh.GetLastScans(fullScanCycle+1, cx1sh.Project.MainBranch)
		if err != nil {
			return fmt.Errorf("failed to get scans from primary branch %v: %s", cx1sh.Project.MainBranch, err)
		}
		// We check if the main branch is eligible for an incremental scan
		incrementalMainBranch, _, contiguousIncrScansMainBranch, err := cx1sh.IncrementalOrFull(scansMainBranch)
		if err != nil {
			return fmt.Errorf("failed to determine incremental or full scan configuration: %s", err)
		}
		log.Entry().Debugf("Main branch %v incremental scan eligibility: %t", cx1sh.Project.MainBranch, incrementalMainBranch)
		if contiguousIncrScansMainBranch+contiguousIncrScansCurrentBranch+1 >= fullScanCycle { // contiguous incremental scans on main branch and current branch must not exceed fullScanCycle
			incrementalMainBranch = false
		}
		log.Entry().Debugf("Main branch + current branch incremental scan eligibility: %t", incrementalMainBranch)
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
	scan, err = cx1sh.PollScanStatus(scan)
	if err != nil {
		return fmt.Errorf("failed while polling scan status: %s", err)
	}

	err = cx1sh.CheckScanCompliance(scan)
	if err != nil {
		return err
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
	sastScan := false
	iacScan := false
	for _, engine := range config.Engines {
		if strings.EqualFold(engine, "sast") {
			sastScan = true
		}
		if strings.EqualFold(engine, "iac") {
			iacScan = true
		}
	}

	if !sastScan && !iacScan {
		return checkmarxOneExecuteScanHelper{}, fmt.Errorf("at least one scan engine must be set in the engines configuration (sast iac)")
	}

	return checkmarxOneExecuteScanHelper{ctx, config, sys, influx, utils, nil, nil, nil, sastScan, iacScan, []piperutils.Path{}}, nil
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
	if c.ScanSAST && len(c.config.SastPreset) == 0 {
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
		// don't allow creation of project at tenant level
		return nil, fmt.Errorf("No application found in config, project cannot be created")
		//project, err = c.sys.CreateProject(c.config.ProjectName, groupIDs)
	}

	if err != nil {
		return nil, fmt.Errorf("Error when trying to create project: %s", err)
	}
	log.Entry().Infof("Project %v created", project.ProjectID)

	// new project, set the defaults per pipeline config
	if c.ScanSAST {
		err = c.sys.SetProjectSASTPreset(project.ProjectID, c.config.SastPreset, true)
		if err != nil {
			return nil, fmt.Errorf("Unable to set SAST preset for project %v to %v: %s", project.ProjectID, c.config.SastPreset, err)
		}
		log.Entry().Infof("Project SAST preset updated to %v", c.config.SastPreset)

		// TODO: set sast defaults even for non-sast scan?
		if len(c.config.LanguageMode) != 0 {
			err = c.sys.SetProjectLanguageMode(project.ProjectID, c.config.LanguageMode, true)
			if err != nil {

				return nil, fmt.Errorf("Unable to set SAST languageMode for project %v to %v: %s", project.ProjectID, c.config.LanguageMode, err)
			}
			log.Entry().Infof("Project languageMode updated to %v", c.config.LanguageMode)
		}
	}

	if c.ScanIAC {
		if c.config.IacPreset != "" {
			err = c.sys.SetProjectIACPreset(project.ProjectID, c.config.IacPreset, true)

			if err != nil {
				return nil, fmt.Errorf("Unable to set IAC preset for project %v to %v: %s", project.ProjectID, c.config.IacPreset, err)
			}
			log.Entry().Infof("Project IAC preset updated to %v", c.config.IacPreset)
		}
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

func (c *checkmarxOneExecuteScanHelper) SetProjectPresetsAndFilters() error {
	projectConf, err := c.sys.GetProjectConfiguration(c.Project.ProjectID)

	if err != nil {
		return fmt.Errorf("failed to retrieve current project configuration: %s", err)
	}

	currentSASTPreset := ""
	currentSASTFilter := ""
	currentIACPreset := ""
	currentIACFilter := ""
	currentLanguageMode := "multi" // piper default
	for _, conf := range projectConf {
		switch conf.Key {
		case checkmarxOne.ConfigurationKeys.SAST.PresetName:
			currentSASTPreset = conf.Value
		case checkmarxOne.ConfigurationKeys.SAST.LanguageMode:
			currentLanguageMode = conf.Value
		case checkmarxOne.ConfigurationKeys.SAST.FileFilter:
			currentSASTFilter = conf.Value
		case checkmarxOne.ConfigurationKeys.IAC.FileFilter:
			currentIACFilter = conf.Value
		case checkmarxOne.ConfigurationKeys.IAC.PresetID:
			iacPresetName, err := c.sys.GetIACPresetNameByID(conf.Value)
			if err != nil {
				return err
			}
			currentIACPreset = iacPresetName
		}

	}

	if c.ScanSAST {
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

		if c.config.SastPreset == "" {
			if currentSASTPreset == "" {
				return fmt.Errorf("must specify the SAST preset in either the pipeline yaml or in the CheckmarxOne project configuration")
			} else {
				log.Entry().Infof("Pipeline yaml does not specify a SAST preset, will use project configuration (%v).", currentSASTPreset)
			}
			c.config.SastPreset = currentSASTPreset
		} else if currentSASTPreset != c.config.SastPreset {
			log.Entry().Infof("Project configured SAST preset (%v) does not match pipeline yaml (%v) - updating project configuration.", currentSASTPreset, c.config.SastPreset)
			c.sys.SetProjectSASTPreset(c.Project.ProjectID, c.config.SastPreset, true)

			if c.config.Incremental {
				log.Entry().Warn("Changing project settings requires a full scan to take effect - switching from incremental to full")
				c.config.Incremental = false
			}
		}

		filterStr := currentSASTFilter
		if filterStr == "" {
			filterStr = "no filter"
		}

		if c.config.SastFilterPattern == "" {
			log.Entry().Infof("Pipeline yaml does not specify a SAST file filter, will use project configuration (%v).", filterStr)
			c.config.SastFilterPattern = currentSASTFilter
		} else if currentSASTFilter != c.config.SastFilterPattern {
			log.Entry().Infof("Project configured SAST file filter (%v) does not match pipeline yaml (%v) - updating project configuration.", currentSASTFilter, c.config.SastFilterPattern)
			c.sys.SetProjectSASTFileFilter(c.Project.ProjectID, c.config.SastFilterPattern, true)

			if c.config.Incremental {
				log.Entry().Warn("Changing project settings requires a full scan to take effect - switching from incremental to full")
				c.config.Incremental = false
			}
		}
	}

	if c.ScanIAC {
		presetStr := currentIACPreset
		if presetStr == "" {
			presetStr = "all checks"
		}
		if c.config.IacPreset == "" {
			if currentIACPreset == "" {
				//return fmt.Errorf("must specify the IAC preset in either the pipeline yaml or in the CheckmarxOne project configuration")
				log.Entry().Infof("No IAC preset is set - using default (%s)", checkmarxOne.IACDefaultBlankPreset)
			} else {
				log.Entry().Infof("Pipeline yaml does not specify a IAC preset, will use project configuration (%v).", presetStr)
			}
			c.config.IacPreset = currentIACPreset
		} else if currentIACPreset != c.config.IacPreset {
			log.Entry().Infof("Project configured IAC preset (%v) does not match pipeline yaml (%v) - updating project configuration.", currentIACPreset, c.config.IacPreset)
			c.sys.SetProjectIACPreset(c.Project.ProjectID, c.config.IacPreset, true)
		}

		filterStr := currentIACFilter
		if filterStr == "" {
			filterStr = "no filter"
		}
		if c.config.IacFilterPattern == "" {
			log.Entry().Infof("Pipeline yaml does not specify a IAC file filter, will use project configuration (%v).", filterStr)
			c.config.IacFilterPattern = currentIACFilter
		} else if currentIACFilter != c.config.IacFilterPattern {
			log.Entry().Infof("Project configured IAC file filter (%v) does not match pipeline yaml (%v) - updating project configuration.", currentIACFilter, c.config.IacFilterPattern)
			c.sys.SetProjectIACFileFilter(c.Project.ProjectID, c.config.IacFilterPattern, true)
		}
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

func (c *checkmarxOneExecuteScanHelper) IncrementalOrFull(scans []checkmarxOne.Scan) (bool, bool, int, error) {
	incremental := c.config.Incremental
	fullScanExists := false
	fullScanCycle, err := strconv.Atoi(c.config.FullScanCycle)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return false, false, 0, fmt.Errorf("invalid configuration value for fullScanCycle %v, must be a positive int", c.config.FullScanCycle)
	}

	if len(scans) == 0 {
		return false, false, 0, nil // no scans exist, so we need to do a full scan
	}

	var scanIds []string
	for _, scan := range scans {
		scanIds = append(scanIds, scan.ScanID)
	}

	scanMetadatas, err := c.sys.GetScanSASTMetadatas(scanIds)
	if err != nil {
		return false, false, 0, fmt.Errorf("failed to fetch metadata for scans: %w", err)
	}

	contiguousIncrementalScans := 0
	for _, scanMetadata := range scanMetadatas {
		if scanMetadata.IsIncremental {
			contiguousIncrementalScans++
		} else {
			fullScanExists = true
			break
		}
	}

	if c.config.IsOptimizedAndScheduled {
		incremental = false
	} else if incremental && c.config.FullScansScheduled && fullScanCycle > 0 && (contiguousIncrementalScans+1) >= fullScanCycle {
		incremental = false
	}

	return incremental, fullScanExists, contiguousIncrementalScans, nil
}

const defaultZipFilterPattern = `!**/node_modules/**, !**/.xmake/**, !**/*_test.go, !**/vendor/**/*.go, **/*.html, **/*.xml, **/*.go, **/*.py, **/*.js, **/*.rb, **/*.scala, **/*.ts`

func (c *checkmarxOneExecuteScanHelper) ZipFiles() (*os.File, error) {
	if c.ScanIAC && c.config.FilterPattern == defaultZipFilterPattern {
		log.Entry().Warn("Zip file filter pattern is SAST-specific, but IaC scan is enabled. Verify that the IaC files are included in the filterPattern, otherwise no files will be scanned.")
	}
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
	cicdOrch := orchestrator.GetOrchestratorConfigProvider(nil)
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

	if branch == "" {
		branch = ".unknown"
		log.Entry().Info("No branch name found, using the cxone default '.unknown' as branch name")
	}

	baseBranch := cicdOrch.PullRequestConfig().Base
	isPR := cicdOrch.IsPullRequest()
	log.Entry().Debugf("CxOne scan branch was automatically set to : %v", branch)
	return branch, isPR, baseBranch
}

func (c *checkmarxOneExecuteScanHelper) CreateScanRequest(incremental bool, uploadLink string, baseBranch string) (*checkmarxOne.Scan, error) {
	configs := []checkmarxOne.ScanConfiguration{}

	branch, _, _ := c.GetScanBranch()
	generalConfigString := fmt.Sprintf("Cx1 Branch name %v", branch)
	configStrings := []string{generalConfigString}

	if c.ScanSAST {
		sastConfigString := "SAST: "
		sastConfig := checkmarxOne.ScanConfiguration{
			ScanType: "sast",

			Values: make(map[string]string, 0)}
		sastConfig.Values["incremental"] = strconv.FormatBool(incremental)
		sastConfig.Values["presetName"] = c.config.SastPreset // always set, either coming from config or coming from Cx1 configuration
		if incremental && len(baseBranch) > 0 {               // base the incremental scan on the specified base branch
			sastConfig.Values["baseBranch"] = baseBranch
			sastConfigString = fmt.Sprintf("baseBranch: %v, ", baseBranch)
		}
		sastConfigString = fmt.Sprintf("%vincremental %v, preset %v", sastConfigString, strconv.FormatBool(incremental), c.config.SastPreset)

		if len(c.config.LanguageMode) > 0 {
			sastConfig.Values["languageMode"] = c.config.LanguageMode
			sastConfigString = sastConfigString + fmt.Sprintf(", languageMode %v", c.config.LanguageMode)
		}

		if c.config.SastFilterPattern != "" {
			sastConfigString += fmt.Sprintf(", file filter <%s>", c.config.SastFilterPattern)
		} else {
			sastConfigString += ", no files filtered"
		}

		configs = append(configs, sastConfig)
		configStrings = append(configStrings, sastConfigString)
	}

	if c.ScanIAC {
		iacConfigString := "IAC: "
		iacConfig := checkmarxOne.ScanConfiguration{
			ScanType: "kics",

			Values: make(map[string]string, 0)}
		presetId, err := c.sys.GetIACPresetIDByName(c.config.IacPreset)
		if err != nil {
			return nil, err
		}
		if presetId != "" {
			iacConfig.Values["presetId"] = presetId
		}
		iacConfigString += fmt.Sprintf("preset %s", c.config.IacPreset)

		if c.config.IacFilterPattern != "" {
			iacConfigString += fmt.Sprintf(", file filter <%s>", c.config.IacFilterPattern)
		} else {
			iacConfigString += ", no files filtered"
		}

		configs = append(configs, iacConfig)
		configStrings = append(configStrings, iacConfigString)
	}

	log.Entry().Infof("Will run a scan with the following configuration: %s", strings.Join(configStrings, "; "))

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

func (c *checkmarxOneExecuteScanHelper) PollScanStatus(scan *checkmarxOne.Scan) (*checkmarxOne.Scan, error) {
	statusDetails := "Scan phase: New"
	pastStatusDetails := statusDetails
	log.Entry().Info(statusDetails)
	status := "New"
	var scan_refresh checkmarxOne.Scan
	var err error
	for {
		scan_refresh, err = c.sys.GetScan(scan.ScanID)

		if err != nil {
			return nil, fmt.Errorf("Error while polling scan %v: %s", scan.ScanID, err)
		}

		status = scan_refresh.Status
		workflow, err := c.sys.GetScanWorkflow(scan.ScanID)
		if err != nil {
			return nil, fmt.Errorf("Error while getting workflow for scan %v: %s", scan.ScanID, err)
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
		return nil, fmt.Errorf("Scan %v canceled via web interface", scan.ScanID)
	}
	if status == "Failed" {
		return nil, fmt.Errorf("Checkmarx One scan failed with the following error: %v", statusDetails)
	}
	return &scan_refresh, nil
}

type gitComment struct {
	criticalSeverityString, highSeverityString, mediumSeverityString, lowSeverityString, criticalComplianceCheckString, highComplianceCheckString, mediumComplianceCheckString, lowComplianceCheckString string
}

func (g *gitComment) String() string {
	return fmt.Sprintf(`Severity | Number of unaudited findings
--- | ---
:bangbang: Critical | %s
:red_circle: High | %s
:orange_circle: Medium | %s
:yellow_circle: Low | %s`, g.criticalSeverityString, g.highSeverityString, g.mediumSeverityString, g.lowSeverityString)
}
func (g *gitComment) Parse(findings *[]checkmarxOne.Finding, config *checkmarxOneExecuteScanOptions) {
	for _, finding := range *findings {
		switch finding.ClassificationName {
		case "Critical":
			// TODO: check if config threshold unit is percent or absolute number
			if *finding.Audited < int(math.Ceil((float64(config.VulnerabilityThresholdCritical)/100.0)*float64(finding.Total))) {
				g.criticalComplianceCheckString = ":x:"
			} else {
				g.criticalComplianceCheckString = ":white_check_mark:"
			}
			if finding.Confirmed > 0 {
				g.criticalSeverityString = fmt.Sprintf("%s %d (%d confirmed)", g.criticalComplianceCheckString, finding.Total-*finding.Audited, finding.Confirmed)
			} else {
				g.criticalSeverityString = fmt.Sprintf("%s %d", g.criticalComplianceCheckString, finding.Total-*finding.Audited)
			}
		case "High":
			if *finding.Audited < int(math.Ceil((float64(config.VulnerabilityThresholdHigh)/100.0)*float64(finding.Total))) {
				g.highComplianceCheckString = ":x:"
			} else {
				g.highComplianceCheckString = ":white_check_mark:"
			}
			if finding.Confirmed > 0 {
				g.highSeverityString = fmt.Sprintf("%s %d (%d confirmed)", g.highComplianceCheckString, finding.Total-*finding.Audited, finding.Confirmed)
			} else {
				g.highSeverityString = fmt.Sprintf("%s %d", g.highComplianceCheckString, finding.Total-*finding.Audited)
			}
		case "Medium":
			if *finding.Audited < int(math.Ceil((float64(config.VulnerabilityThresholdMedium)/100.0)*float64(finding.Total))) {
				g.mediumComplianceCheckString = ":x:"
			} else {
				g.mediumComplianceCheckString = ":white_check_mark:"
			}
			if finding.Confirmed > 0 {
				g.mediumSeverityString = fmt.Sprintf("%s %d (%d confirmed)", g.mediumComplianceCheckString, finding.Total-*finding.Audited, finding.Confirmed)
			} else {
				g.mediumSeverityString = fmt.Sprintf("%s %d", g.mediumComplianceCheckString, finding.Total-*finding.Audited)
			}
		case "Low":
			if finding.LowPerQuery != nil {
				lowPerQuery := *finding.LowPerQuery
				if config.VulnerabilityThresholdLowPerQuery {
					sort.Slice(lowPerQuery, func(i, j int) bool {
						iRequired := min(int(math.Ceil(float64(lowPerQuery[i].Total)*float64(config.VulnerabilityThresholdLow)/100.0)), config.VulnerabilityThresholdLowPerQueryMax)
						jRequired := min(int(math.Ceil(float64(lowPerQuery[j].Total)*float64(config.VulnerabilityThresholdLow)/100.0)), config.VulnerabilityThresholdLowPerQueryMax)
						iFailing := lowPerQuery[i].Audited < iRequired
						jFailing := lowPerQuery[j].Audited < jRequired
						if iFailing != jFailing {
							return iFailing // failing entries first
						}
						return lowPerQuery[i].QueryName < lowPerQuery[j].QueryName
					})
				}
				for _, lowFinding := range lowPerQuery {
					if config.VulnerabilityThresholdLowPerQuery {
						confirmedLowString := ""
						if lowFinding.Confirmed > 0 {
							confirmedLowString = fmt.Sprintf(", of which %d confirmed", lowFinding.Confirmed)
						}
						lowAuditedRequiredPerQuery := min(int(math.Ceil(float64(lowFinding.Total)*float64(config.VulnerabilityThresholdLow)/100.0)), config.VulnerabilityThresholdLowPerQueryMax)
						if lowFinding.Audited < lowAuditedRequiredPerQuery {
							g.lowComplianceCheckString = ":x:"
						} else {
							g.lowComplianceCheckString = ":white_check_mark:"
						}
						g.lowSeverityString = fmt.Sprintf("%s%s %d %s (%d audited / %d required%s) <br>", g.lowSeverityString, g.lowComplianceCheckString, lowFinding.Total-lowFinding.Audited, lowFinding.QueryName, lowFinding.Audited, lowAuditedRequiredPerQuery, confirmedLowString)
					} else {
						g.lowSeverityString = fmt.Sprintf("%s%s %d %s<br>", g.lowSeverityString, g.lowComplianceCheckString, lowFinding.Total-lowFinding.Audited, lowFinding.QueryName)
					}
				}

				if g.lowSeverityString == "" { // no findings at all
					g.lowSeverityString = ":white_check_mark: 0"
				}
			} else {
				if *finding.Audited < int(math.Ceil((float64(config.VulnerabilityThresholdLow)/100.0)*float64(finding.Total))) {
					g.lowComplianceCheckString = ":x:"
				} else {
					g.lowComplianceCheckString = ":white_check_mark:"
				}
				if finding.Confirmed > 0 {
					g.lowSeverityString = fmt.Sprintf("%s %d (%d confirmed)", g.lowComplianceCheckString, finding.Total-*finding.Audited, finding.Confirmed)
				} else {
					g.lowSeverityString = fmt.Sprintf("%s %d", g.lowComplianceCheckString, finding.Total-*finding.Audited)
				}
			}
		}
	}
}

func (c *checkmarxOneExecuteScanHelper) PostScanSummaryInPullRequest(detailedResults *map[string]any, insecure bool) error {
	cicdOrch := orchestrator.GetOrchestratorConfigProvider(nil)
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
				log.Entry().Debugf("Found repository %s and owner %s from orchestrator", repository, owner)
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

		var scanId, deepLink string
		var sastScan, iacScan string
		if c.ScanSAST {
			var sast_status gitComment
			sastScanReportOverview := checkmarxOne.CreateJSONHeaderReport(detailedResults, "sast")
			deepLink = sastScanReportOverview.DeepLink
			scanId = sastScanReportOverview.ScanID
			sast_status.Parse(sastScanReportOverview.Findings, &c.config)
			sastTable := sast_status.String()

			sastScan = fmt.Sprintf(`**SAST Scan type**: %s
**SAST Scan Preset**: %s
**SAST Results**
%s

`, strings.ToLower(sastScanReportOverview.ScanType), sastScanReportOverview.Preset, sastTable)

		}
		if c.ScanIAC {
			var iac_status gitComment
			iacScanReportOverview := checkmarxOne.CreateJSONHeaderReport(detailedResults, "iac")
			if deepLink == "" {
				deepLink = iacScanReportOverview.DeepLink
			}
			iac_status.Parse(iacScanReportOverview.Findings, &c.config)
			iacTable := iac_status.String()

			iacScan = fmt.Sprintf(`**IAC Preset**: %s
**IAC Results**
%s

`, iacScanReportOverview.Preset, iacTable)
		}
		var scanIcon string
		if insecure {
			scanIcon = ":x:"
		} else {
			scanIcon = ":white_check_mark:"
		}
		comment := &github.IssueComment{
			Body: new(fmt.Sprintf(`<!-- Piper CxOne Scan Summary -->
# %s CheckmarxOne scan completed 
**Project**: %s
**ScanId**: %s
%s%s

[Go to the scan results](%s)
		`, scanIcon, c.Project.Name, scanId, sastScan, iacScan, deepLink)),
		}
		pullRequestNumber, err := strconv.Atoi(pullRequestId)
		if err != nil {
			return fmt.Errorf("failed to parse int from pull request name %s: %s", c.config.PullRequestName, err)
		}
		// Check if comment already exists, delete old one to avoid multiple comments
		// search for watermark <!-- Piper CxOne Scan Summary -->
		comments, _, err := ghIssues.ListComments(c.ctx, owner, repository, pullRequestNumber, &github.IssueListCommentsOptions{})
		if err != nil {
			log.Entry().Errorf("failed to list GitHub issue comments: %s", err)
		} else {
			for _, existingComment := range comments {
				if strings.Contains(*existingComment.Body, "<!-- Piper CxOne Scan Summary -->") {
					_, err := ghIssues.DeleteComment(c.ctx, owner, repository, existingComment.GetID())
					if err != nil {
						log.Entry().Errorf("failed to delete old GitHub issue comment: %s", err)
					}
					log.Entry().Infof("Deleted old GitHub issue comment for project %v", c.Project.Name)
					break
				}
			}
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

func (c *checkmarxOneExecuteScanHelper) CheckScanCompliance(scan *checkmarxOne.Scan) error {
	results, err := c.ParseResults(scan) // incl report-gen
	if err != nil {
		return fmt.Errorf("failed to get scan results: %s", err)
	}
	err = c.CheckCompliance(scan, &results)
	if err != nil {
		log.SetErrorCategory(log.ErrorCompliance)
		return fmt.Errorf("project %v not compliant: %s", c.Project.Name, err)
	}
	return nil
}

func (c *checkmarxOneExecuteScanHelper) CheckCompliance(scan *checkmarxOne.Scan, detailedResults *map[string]any) error {
	links := []piperutils.Path{{Target: (*detailedResults)["DeepLink"].(string), Name: "Checkmarx One Web UI"}}
	insecure := false
	var insecureResults []string
	var neutralResults []string

	if c.config.VulnerabilityThresholdEnabled || c.config.IacVulnerabilityThresholdEnabled {
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

func (c *checkmarxOneExecuteScanHelper) GetReportPDF(scan *checkmarxOne.Scan, engines []string) error {
	if len(engines) == 0 {
		return fmt.Errorf("cannot generate a report for 0 engines")
	}
	if c.config.GeneratePdfReport {
		pdfReportName := c.createReportName(c.utils.GetWorkspace(), "Cx1_"+strings.ToUpper(engines[0])+"Report_%v.pdf")
		err := c.downloadAndSaveReport(pdfReportName, scan, "pdf", engines)
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

func (c *checkmarxOneExecuteScanHelper) GetReportSASTSARIF(scan *checkmarxOne.Scan, scanmeta *checkmarxOne.ScanMetadata, results *[]checkmarxOne.ScanResult) error {
	if c.config.ConvertToSarif {
		if scanmeta.SAST != nil {
			log.Entry().Info("Calling SAST JSON conversion to SARIF function.")
			sarif, err := checkmarxOne.ConvertCxSASTJSONToSarif(c.sys, c.config.ServerURL, results, scan)
			if err != nil {
				return fmt.Errorf("Failed to generate SARIF: %s", err)
			}
			paths, err := checkmarxOne.WriteSASTSarif(sarif)
			if err != nil {
				return fmt.Errorf("Failed to write SARIF: %s", err)
			}
			c.reports = append(c.reports, paths...)
		}
	}
	return nil
}

func (c *checkmarxOneExecuteScanHelper) GetReportIACSARIF(scan *checkmarxOne.Scan, scanmeta *checkmarxOne.ScanMetadata, results *[]checkmarxOne.ScanResult) error {
	if c.config.ConvertToSarif {
		if scanmeta.IAC != nil {
			log.Entry().Info("Calling IAC JSON conversion to SARIF function.")
			sarif, err := checkmarxOne.ConvertCxIACJSONToSarif(c.sys, c.config.ServerURL, results, scan)
			if err != nil {
				return fmt.Errorf("Failed to generate SARIF: %s", err)
			}
			paths, err := checkmarxOne.WriteIACSarif(sarif)
			if err != nil {
				return fmt.Errorf("Failed to write SARIF: %s", err)
			}
			c.reports = append(c.reports, paths...)
		}
	}
	return nil
}

func (c *checkmarxOneExecuteScanHelper) GetReportJSON(scan *checkmarxOne.Scan, engines []string) error {
	if len(engines) == 0 {
		return fmt.Errorf("cannot generate a report for 0 engines")
	}
	jsonReportName := c.createReportName(c.utils.GetWorkspace(), "Cx1_"+strings.ToUpper(engines[0])+"Report_%v.json")
	err := c.downloadAndSaveReport(jsonReportName, scan, "json", engines)
	if err != nil {
		return fmt.Errorf("Report download failed: %s", err)
	} else {
		c.reports = append(c.reports, piperutils.Path{Target: jsonReportName, Mandatory: true})
	}
	return nil
}

func (c *checkmarxOneExecuteScanHelper) GetHeaderReportJSON(detailedResults *map[string]any) error {
	// This is for the SAP-piper-format short-form JSON report
	if c.ScanSAST {
		jsonReport := checkmarxOne.CreateJSONHeaderReport(detailedResults, "sast")
		paths, err := checkmarxOne.WriteJSONHeaderReport(jsonReport, "sast")
		if err != nil {
			return fmt.Errorf("Failed to write JSON header report: %s", err)
		} else {
			// add JSON report to archiving list
			c.reports = append(c.reports, paths...)
		}
	}

	if c.ScanIAC {
		jsonReport := checkmarxOne.CreateJSONHeaderReport(detailedResults, "iac")
		paths, err := checkmarxOne.WriteJSONHeaderReport(jsonReport, "iac")
		if err != nil {
			return fmt.Errorf("Failed to write JSON header report: %s", err)
		} else {
			// add JSON report to archiving list
			c.reports = append(c.reports, paths...)
		}
	}
	return nil
}

func (c *checkmarxOneExecuteScanHelper) ParseResults(scan *checkmarxOne.Scan) (map[string]any, error) {
	var detailedResults map[string]any

	scanmeta, err := c.sys.GetScanMetadata(scan)
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

	if c.ScanSAST {
		err = c.GetReportJSON(scan, []string{"sast"})
		if err != nil {
			log.Entry().WithError(err).Warnf("Failed to get JSON SAST report")
		}
		err = c.GetReportPDF(scan, []string{"sast"})
		if err != nil {
			log.Entry().WithError(err).Warnf("Failed to get PDF SAST report")
		}
		err = c.GetReportSASTSARIF(scan, &scanmeta, &results)
		if err != nil {
			log.Entry().WithError(err).Warnf("Failed to get SARIF SAST report")
		}
	}

	if c.ScanIAC {
		err = c.GetReportJSON(scan, []string{"iac"})
		if err != nil {
			log.Entry().WithError(err).Warnf("Failed to get JSON IAC report")
		}
		err = c.GetReportPDF(scan, []string{"iac"})
		if err != nil {
			log.Entry().WithError(err).Warnf("Failed to get PDF IAC report")
		}
		err = c.GetReportIACSARIF(scan, &scanmeta, &results)
		if err != nil {
			log.Entry().WithError(err).Warnf("Failed to get SARIF IAC report")
		}
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

func (c *checkmarxOneExecuteScanHelper) downloadAndSaveReport(reportFileName string, scan *checkmarxOne.Scan, reportType string, engines []string) error {
	report, err := c.generateAndDownloadReport(scan, reportType, engines)
	if err != nil {
		return fmt.Errorf("failed to download the report: %w", err)
	}
	log.Entry().Debugf("Saving report to file %v...", reportFileName)
	return c.utils.WriteFile(reportFileName, report, 0o700)
}

func (c *checkmarxOneExecuteScanHelper) generateAndDownloadReport(scan *checkmarxOne.Scan, reportType string, engines []string) ([]byte, error) {
	var finalStatus checkmarxOne.ReportStatus

	report, err := c.sys.RequestNewReport(scan.ScanID, scan.ProjectID, scan.Branch, reportType, engines)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to request new report: %w", err)
	}
	for {
		finalStatus, err = c.sys.GetReportStatus(report)
		if err != nil {
			return []byte{}, fmt.Errorf("failed to get report status: %w", err)
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

func (c *checkmarxOneExecuteScanHelper) getDetailedResults(scan *checkmarxOne.Scan, scanmeta *checkmarxOne.ScanMetadata, results *[]checkmarxOne.ScanResult) (map[string]any, error) {
	// this converts the JSON format results from Cx1 into the "resultMap" structure used in other parts of this step (influx etc)

	resultMap := map[string]any{}
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

	version, err := c.sys.GetVersion()
	if err != nil {
		resultMap["ToolVersion"] = "Error fetching current version"
	} else {
		resultMap["ToolVersion"] = "CxOne: " + version.CxOne
		resultMap["SASTVersion"] = "SAST: " + version.SAST
		resultMap["IACVersion"] = "IAC: " + version.KICS
	}

	if scanmeta.SAST != nil {
		resultMap["SastPreset"] = scanmeta.SAST.PresetName
		if !scanmeta.SAST.IsIncremental {
			resultMap["ScanType"] = "Full"
		} else {
			resultMap["ScanType"] = "Incremental"
		}

		resultMap["LinesOfCodeScanned"] = scanmeta.SAST.LOC
		resultMap["FilesScanned"] = scanmeta.SAST.FileCount
	} else {
		resultMap["ScanType"] = "Full"
		resultMap["SastPreset"] = "n/a"

		resultMap["LinesOfCodeScanned"] = 0
		resultMap["FilesScanned"] = 0
	}

	if scanmeta.IAC != nil {
		resultMap["IacPreset"] = scanmeta.IAC.PresetName
		resultMap["IacLinesOfCodeScanned"] = scanmeta.IAC.IACLOC
		resultMap["IacFilesScanned"] = scanmeta.IAC.FileCount
	} else {
		resultMap["IacPreset"] = "n/a"
		resultMap["IacLinesOfCodeScanned"] = 0
		resultMap["IacFilesScanned"] = 0
	}
	resultMap["DeepLink"] = fmt.Sprintf("%v/projects/%v/overview?branch=%v", c.config.ServerURL, c.Project.ProjectID, url.QueryEscape(scan.Branch))
	resultMap["ReportCreationTime"] = time.Now().String()
	resultMap["Critical"] = map[string]int{}
	resultMap["High"] = map[string]int{}
	resultMap["Medium"] = map[string]int{}
	resultMap["Low"] = map[string]int{}
	resultMap["Information"] = map[string]int{}

	resultMap["IACCritical"] = map[string]int{}
	resultMap["IACHigh"] = map[string]int{}
	resultMap["IACMedium"] = map[string]int{}
	resultMap["IACLow"] = map[string]int{}
	resultMap["IACInformation"] = map[string]int{}

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

			if strings.EqualFold(result.Type, "kics") {
				key = "IAC" + key
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
		// this only covers SAST findings
		if c.config.VulnerabilityThresholdLowPerQuery {
			var lowPerQuery = map[string]map[string]int{}

			for _, result := range *results {
				if !strings.EqualFold(result.Type, "sast") {
					continue
				}
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

			lowPerQuery = map[string]map[string]int{}

			for _, result := range *results {
				if !strings.EqualFold(result.Type, "kics") {
					continue
				}
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

			resultMap["IACLowPerQuery"] = lowPerQuery
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
		return zipFile, fmt.Errorf("failed to create archive of project sources: %w", err)
	}
	defer zipFile.Close()

	err = c.zipFolder(utils.GetWorkspace(), zipFile, patterns, zipFileName, utils)
	if err != nil {
		return nil, fmt.Errorf("failed to compact folder: %w", err)
	}
	return zipFile, nil
}

func (c *checkmarxOneExecuteScanHelper) zipFolder(source string, zipFile io.Writer, patterns []string, zipFileName string, utils checkmarxOneExecuteScanUtils) error {
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

	// resolve the output archive's absolute path so it can be skipped during the walk,
	// otherwise the archive would recursively include itself and inflate indefinitely
	absZipFileName, absZipErr := filepath.Abs(zipFileName)
	if absZipErr != nil {
		absZipFileName = filepath.Clean(zipFileName)
	}

	fileCount := 0
	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.Mode().IsRegular() || info.Size() == 0 {
			return nil
		}

		// skip the output archive itself to avoid recursively zipping it into itself
		absPath, absErr := filepath.Abs(path)
		if absErr != nil {
			absPath = filepath.Clean(path)
		}
		if absPath == absZipFileName {
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
			return false, fmt.Errorf("Pattern %v could not get executed: %w", pattern, err)
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
			return false, fmt.Errorf("Pattern %v could not get executed: %w", pattern, err)
		}

		if match { // match with an exclude pattern, the file is excluded
			return true, nil
		}
	}

	return false, nil
}

func (c *checkmarxOneExecuteScanHelper) createToolRecordCx(results *map[string]any) (string, error) {
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

func (c *checkmarxOneExecuteScanHelper) enforceThresholds(results *map[string]any) (bool, []string, []string) {
	neutralResults := []string{}
	insecureResults := []string{}
	insecure := false
	if c.ScanSAST && c.config.VulnerabilityThresholdEnabled {
		insecure, neutralResults, insecureResults = c.enforceThresholdsPerEngine("SAST", results)
	}
	if c.ScanIAC {
		if c.config.VulnerabilityThresholdEnabled {
			insecure2, neutralResults2, insecureResults2 := c.enforceThresholdsPerEngine("IAC", results)
			if c.config.IacVulnerabilityThresholdEnabled {
				insecure = insecure || insecure2
				neutralResults = append(neutralResults, neutralResults2...)
				insecureResults = append(insecureResults, insecureResults2...)
			} else {
				log.Entry().Infof("Skipping IAC threshold enforcement, IacVulnerabilityThresholdEnabled is set to false")
			}
		} else {
			log.Entry().Warnf("IacVulnerabilityThresholdEnabled is set to true, but VulnerabilityThresholdEnabled is set to false. IAC threshold enforcement will be skipped.")
		}
	}
	return insecure, neutralResults, insecureResults
}

func (c *checkmarxOneExecuteScanHelper) enforceThresholdsPerEngine(engine string, results *map[string]any) (bool, []string, []string) {
	pre := ""
	if strings.EqualFold(engine, "iac") {
		pre = "IAC"
	}
	neutralResults := []string{}
	insecureResults := []string{}
	insecure := false

	cxCriticalThreshold := c.config.VulnerabilityThresholdCritical
	cxHighThreshold := c.config.VulnerabilityThresholdHigh
	cxMediumThreshold := c.config.VulnerabilityThresholdMedium
	cxLowThreshold := c.config.VulnerabilityThresholdLow
	cxLowThresholdPerQuery := c.config.VulnerabilityThresholdLowPerQuery
	cxLowThresholdPerQueryMax := c.config.VulnerabilityThresholdLowPerQueryMax
	// findings are audited if they are in state Confirmed, Urgent or NotExploitable
	criticalValue := (*results)[pre+"Critical"].(map[string]int)["ToVerify"] + (*results)[pre+"Critical"].(map[string]int)["ProposedNotExploitable"]
	confirmedCriticalValue := (*results)[pre+"Critical"].(map[string]int)["Confirmed"] + (*results)[pre+"Critical"].(map[string]int)["Urgent"]
	highValue := (*results)[pre+"High"].(map[string]int)["ToVerify"] + (*results)[pre+"High"].(map[string]int)["ProposedNotExploitable"]
	confirmedHighValue := (*results)[pre+"High"].(map[string]int)["Confirmed"] + (*results)[pre+"High"].(map[string]int)["Urgent"]
	mediumValue := (*results)[pre+"Medium"].(map[string]int)["ToVerify"] + (*results)[pre+"Medium"].(map[string]int)["ProposedNotExploitable"]
	confirmedMediumValue := (*results)[pre+"Medium"].(map[string]int)["Confirmed"] + (*results)[pre+"Medium"].(map[string]int)["Urgent"]
	lowValue := (*results)[pre+"Low"].(map[string]int)["ToVerify"] + (*results)[pre+"Low"].(map[string]int)["ProposedNotExploitable"]
	confirmedLowValue := (*results)[pre+"Low"].(map[string]int)["Confirmed"] + (*results)[pre+"Low"].(map[string]int)["Urgent"]
	var unit string
	criticalViolation := ""
	highViolation := ""
	mediumViolation := ""
	lowViolation := ""
	if c.config.VulnerabilityThresholdUnit == "percentage" {
		unit = "%"
		criticalAudited := (*results)[pre+"Critical"].(map[string]int)["NotExploitable"] + (*results)[pre+"Critical"].(map[string]int)["Confirmed"] + (*results)[pre+"Critical"].(map[string]int)["Urgent"]
		criticalOverall := (*results)[pre+"Critical"].(map[string]int)["Issues"]
		if criticalOverall == 0 {
			criticalAudited = 1
			criticalOverall = 1
		}
		highAudited := (*results)[pre+"High"].(map[string]int)["NotExploitable"] + (*results)[pre+"High"].(map[string]int)["Confirmed"] + (*results)[pre+"High"].(map[string]int)["Urgent"]
		highOverall := (*results)[pre+"High"].(map[string]int)["Issues"]
		if highOverall == 0 {
			highAudited = 1
			highOverall = 1
		}
		mediumAudited := (*results)[pre+"Medium"].(map[string]int)["NotExploitable"] + (*results)[pre+"Medium"].(map[string]int)["Confirmed"] + (*results)[pre+"Medium"].(map[string]int)["Urgent"]
		mediumOverall := (*results)[pre+"Medium"].(map[string]int)["Issues"]
		if mediumOverall == 0 {
			mediumAudited = 1
			mediumOverall = 1
		}
		lowAudited := (*results)[pre+"Low"].(map[string]int)["Confirmed"] + (*results)[pre+"Low"].(map[string]int)["NotExploitable"] + (*results)[pre+"Low"].(map[string]int)["Urgent"]
		lowOverall := (*results)[pre+"Low"].(map[string]int)["Issues"]
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
			if (*results)[pre+"LowPerQuery"] != nil {
				lowPerQueryMap := (*results)[pre+"LowPerQuery"].(map[string]map[string]int)

				for lowQuery, resultsLowQuery := range lowPerQueryMap {
					lowAuditedPerQuery := resultsLowQuery["Confirmed"] + resultsLowQuery["NotExploitable"] + resultsLowQuery["Urgent"]
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

	var confirmedCriticalString, confirmedHighString, confirmedMediumString, confirmedLowString string
	if confirmedCriticalValue > 0 {
		confirmedCriticalString = fmt.Sprintf(" (of which %v confirmed)", confirmedCriticalValue)
	}
	if confirmedHighValue > 0 {
		confirmedHighString = fmt.Sprintf(" (of which %v confirmed)", confirmedHighValue)
	}
	if confirmedMediumValue > 0 {
		confirmedMediumString = fmt.Sprintf(" (of which %v confirmed)", confirmedMediumValue)
	}
	if confirmedLowValue > 0 {
		confirmedLowString = fmt.Sprintf(" (of which %v confirmed)", confirmedLowValue)
	}
	criticalText := fmt.Sprintf("Critical %v%v %v %v", criticalValue, unit, confirmedCriticalString, criticalViolation)
	highText := fmt.Sprintf("High %v%v %v %v", highValue, unit, confirmedHighString, highViolation)
	mediumText := fmt.Sprintf("Medium %v%v %v %v", mediumValue, unit, confirmedMediumString, mediumViolation)
	lowText := fmt.Sprintf("Low %v%v %v %v", lowValue, unit, confirmedLowString, lowViolation)
	log.Entry().Info(engine + " Result auditing status per severity:")
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

func (c *checkmarxOneExecuteScanHelper) reportToInflux(results *map[string]any) {
	getCount := func(severity, key string) int {
		count := 0

		if m, ok := (*results)[severity]; ok {
			if m, ok := m.(map[string]int); ok {
				count += m[key]
			}
		}
		if m, ok := (*results)["IAC"+severity]; ok {
			if m, ok := m.(map[string]int); ok {
				count += m[key]
			}
		}
		return count
	}

	c.influx.checkmarxOne_data.fields.critical_issues = getCount("Critical", "Issues")
	c.influx.checkmarxOne_data.fields.critical_not_false_postive = getCount("Critical", "NotFalsePositive")
	c.influx.checkmarxOne_data.fields.critical_not_exploitable = getCount("Critical", "NotExploitable")
	c.influx.checkmarxOne_data.fields.critical_confirmed = getCount("Critical", "Confirmed")
	c.influx.checkmarxOne_data.fields.critical_urgent = getCount("Critical", "Urgent")
	c.influx.checkmarxOne_data.fields.critical_proposed_not_exploitable = getCount("Critical", "ProposedNotExploitable")
	c.influx.checkmarxOne_data.fields.critical_to_verify = getCount("Critical", "ToVerify")

	c.influx.checkmarxOne_data.fields.high_issues = getCount("High", "Issues")
	c.influx.checkmarxOne_data.fields.high_not_false_postive = getCount("High", "NotFalsePositive")
	c.influx.checkmarxOne_data.fields.high_not_exploitable = getCount("High", "NotExploitable")
	c.influx.checkmarxOne_data.fields.high_confirmed = getCount("High", "Confirmed")
	c.influx.checkmarxOne_data.fields.high_urgent = getCount("High", "Urgent")
	c.influx.checkmarxOne_data.fields.high_proposed_not_exploitable = getCount("High", "ProposedNotExploitable")
	c.influx.checkmarxOne_data.fields.high_to_verify = getCount("High", "ToVerify")
	c.influx.checkmarxOne_data.fields.medium_issues = getCount("Medium", "Issues")
	c.influx.checkmarxOne_data.fields.medium_not_false_postive = getCount("Medium", "NotFalsePositive")
	c.influx.checkmarxOne_data.fields.medium_not_exploitable = getCount("Medium", "NotExploitable")
	c.influx.checkmarxOne_data.fields.medium_confirmed = getCount("Medium", "Confirmed")
	c.influx.checkmarxOne_data.fields.medium_urgent = getCount("Medium", "Urgent")
	c.influx.checkmarxOne_data.fields.medium_proposed_not_exploitable = getCount("Medium", "ProposedNotExploitable")
	c.influx.checkmarxOne_data.fields.medium_to_verify = getCount("Medium", "ToVerify")
	c.influx.checkmarxOne_data.fields.low_issues = getCount("Low", "Issues")
	c.influx.checkmarxOne_data.fields.low_not_false_postive = getCount("Low", "NotFalsePositive")
	c.influx.checkmarxOne_data.fields.low_not_exploitable = getCount("Low", "NotExploitable")
	c.influx.checkmarxOne_data.fields.low_confirmed = getCount("Low", "Confirmed")
	c.influx.checkmarxOne_data.fields.low_urgent = getCount("Low", "Urgent")
	c.influx.checkmarxOne_data.fields.low_proposed_not_exploitable = getCount("Low", "ProposedNotExploitable")
	c.influx.checkmarxOne_data.fields.low_to_verify = getCount("Low", "ToVerify")
	c.influx.checkmarxOne_data.fields.information_issues = getCount("Information", "Issues")
	c.influx.checkmarxOne_data.fields.information_not_false_postive = getCount("Information", "NotFalsePositive")
	c.influx.checkmarxOne_data.fields.information_not_exploitable = getCount("Information", "NotExploitable")
	c.influx.checkmarxOne_data.fields.information_confirmed = getCount("Information", "Confirmed")
	c.influx.checkmarxOne_data.fields.information_urgent = getCount("Information", "Urgent")
	c.influx.checkmarxOne_data.fields.information_proposed_not_exploitable = getCount("Information", "ProposedNotExploitable")
	c.influx.checkmarxOne_data.fields.information_to_verify = getCount("Information", "ToVerify")

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
	c.influx.checkmarxOne_data.fields.tool_version = fmt.Sprintf("%s, %s, %s", (*results)["ToolVersion"], (*results)["SASTVersion"], (*results)["IACVersion"])

	c.influx.checkmarxOne_data.fields.scan_type = (*results)["ScanType"].(string)
	c.influx.checkmarxOne_data.fields.preset = (*results)["SastPreset"].(string)
	c.influx.checkmarxOne_data.fields.iac_preset = (*results)["IacPreset"].(string)
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
