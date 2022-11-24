package cmd

import (
	"archive/zip"
	"context"
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
    
    "github.com/SAP/jenkins-library/pkg/checkmarxone"
	piperGithub "github.com/SAP/jenkins-library/pkg/github"
	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
    "github.com/SAP/jenkins-library/pkg/reporting"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/toolrecord"
	"github.com/bmatcuk/doublestar"
	"github.com/google/go-github/v45/github"
	"github.com/pkg/errors"
)

type checkmarxoneExecuteScanUtils interface {
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

type checkmarxoneExecuteScanHelper struct{}

type checkmarxoneExecuteScanUtilsBundle struct {
	workspace string
	issues    *github.IssuesService
	search    *github.SearchService
}

func (c *checkmarxoneExecuteScanUtilsBundle) PathMatch(pattern, name string) (bool, error) {
	return doublestar.PathMatch(pattern, name)
}

func (c *checkmarxoneExecuteScanUtilsBundle) GetWorkspace() string {
	return c.workspace
}

func (c *checkmarxoneExecuteScanUtilsBundle) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}

func (c *checkmarxoneExecuteScanUtilsBundle) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (c *checkmarxoneExecuteScanUtilsBundle) FileInfoHeader(fi os.FileInfo) (*zip.FileHeader, error) {
	return zip.FileInfoHeader(fi)
}

func (c *checkmarxoneExecuteScanUtilsBundle) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (c *checkmarxoneExecuteScanUtilsBundle) Open(name string) (*os.File, error) {
	return os.Open(name)
}

func (c *checkmarxoneExecuteScanUtilsBundle) CreateIssue(ghCreateIssueOptions *piperGithub.CreateIssueOptions) error {
	return piperGithub.CreateIssue(ghCreateIssueOptions)
}

func (c *checkmarxoneExecuteScanUtilsBundle) GetIssueService() *github.IssuesService {
	return c.issues
}

func (c *checkmarxoneExecuteScanUtilsBundle) GetSearchService() *github.SearchService {
	return c.search
}

func newCheckmarxOneExecuteScanUtilsBundle(workspace string, client *github.Client) checkmarxoneExecuteScanUtils {
	utils := checkmarxoneExecuteScanUtilsBundle{
		workspace: workspace,
	}
	if client != nil {
		utils.issues = client.Issues
		utils.search = client.Search
	}
	return &utils
}

func checkmarxoneExecuteScan(config checkmarxoneExecuteScanOptions, _ *telemetry.CustomData, influx *checkmarxoneExecuteScanInflux) {
	client := &piperHttp.Client{}
	options := piperHttp.ClientOptions{MaxRetries: config.MaxRetries}
	client.SetOptions(options)
	// TODO provide parameter for trusted certs
	ctx, ghClient, err := piperGithub.NewClient(config.GithubToken, config.GithubAPIURL, "", []string{})
	if err != nil {
		log.Entry().WithError(err).Warning("Failed to get GitHub client")
	}

	// Updated for Cx1: serverURL, iamURL, tenant, APIKey, client_id, client_secret string
	// This handles the authentication based on the provided configuration.
	// Priority is: First use the APIKey if present, otherwise use the ClientID + Secret

	sys, err := checkmarxone.NewSystemInstance(client, config.ServerURL, config.IamURL, config.Tenant, config.APIKey, config.ClientID, config.ClientSecret)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Failed to create Checkmarx One client talking to URLs %v and %v with tenant %v", config.ServerURL, config.IamURL, config.Tenant)
	}
	influx.step_data.fields.checkmarxone = false
	utils := newCheckmarxOneExecuteScanUtilsBundle("./", ghClient)

	cx1scanhelper := checkmarxoneExecuteScanHelper{}

    _, err = cx1scanhelper.RunScan(ctx, config, sys, influx, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to execute Checkmarx One scan.")
	}

	influx.step_data.fields.checkmarxone = true
}

// Updated for Cx1
func (cx1sh *checkmarxoneExecuteScanHelper) RunScan(ctx context.Context, config checkmarxoneExecuteScanOptions, sys checkmarxone.System, influx *checkmarxoneExecuteScanInflux, utils checkmarxoneExecuteScanUtils) (checkmarxone.Scan, error) {
    var scan checkmarxone.Scan
    var group checkmarxone.Group
    var err error

	if len(config.GroupID) > 0 {
        group, err = sys.GetGroupByID(config.GroupID)
		if err != nil {
			return scan, err
		}
    } else if len(config.GroupName) > 0 {
		group, err = sys.GetGroupByName(config.GroupName)
		if err != nil {
			return scan, err
		}
	}

    if len(config.ProjectName) == 0 {
        return scan, errors.New( "No project name set in the configuration" )
    }

	// get the Project, if it exists
	projects, err := sys.GetProjectsByNameAndGroup(config.ProjectName, group.GroupID)
	if err != nil {
		return scan, errors.Wrap(err, "error when trying to load project")
	}

	var project checkmarxone.Project

	if len(projects) == 0 {
		if len(group.Name) == 0 {
			return scan, errors.New("GroupName or GroupID is required to create a new project")
		}

		project, err = sys.CreateProject(config.ProjectName, []string{group.GroupID})
		if err != nil {
			return scan, errors.Wrap(err, "Failed to create new project")
		}

		// new project, set the defaults per pipeline config
		if len(config.Preset) != 0 {
			err = sys.SetProjectPreset(project.ProjectID, config.Preset, true)
            if err != nil {
                log.Entry().Errorf( "Unable to set preset for project %v to: %v. %s", project.ProjectID, config.Preset, err )
                return scan, err
            } else {
                log.Entry().Infof( "Project preset updated to %v", config.Preset )
            }
		}
		if len(config.LanguageMode) != 0 {
			err = sys.SetProjectLanguageMode(project.ProjectID, config.LanguageMode, true)
            if err != nil {
                log.Entry().Errorf( "Unable to set languageMode for project %v to: %v. %s", project.ProjectID, config.LanguageMode, err )
                return scan, err
            } else {
                log.Entry().Infof( "Project languageMode updated to %v", config.LanguageMode )
            }
		}

	} else if len(projects) > 1 {
		log.Entry().Warning("There were " + strconv.Itoa(len(projects)) + " found matching.")

		projectFound := false
		for _, p := range projects {
			if p.Name == config.ProjectName {
				log.Entry().Info("Exact project name match found")
				project = p
				projectFound = true
			}
		}

		if !projectFound {
			project = projects[0]
			log.Entry().Info("Exact project name match not found, selecting the first project in the list: " + project.Name)
		}
	} else {
        project = projects[0]
    }

	log.Entry().Infof("Project %v (ID %v)", project.ProjectID, project.Name)

	
    scan, err = cx1sh.uploadAndScan(ctx, config, sys, group, project, influx, utils)
    if err != nil {
        return scan, errors.Wrap(err, "scan, upload, and result validation returned an error")
    }

    // validation
    
    err = cx1sh.verifyCxProjectCompliance(ctx, config, sys, group, project, scan, influx, utils)
    if err != nil {
        log.SetErrorCategory(log.ErrorCompliance)
        return scan, errors.Wrapf(err, "project %v not compliant", project.Name)
    }

	return scan, nil
}

func (cx1sh *checkmarxoneExecuteScanHelper) uploadAndScan(ctx context.Context, config checkmarxoneExecuteScanOptions, sys checkmarxone.System, group checkmarxone.Group, project checkmarxone.Project, influx *checkmarxoneExecuteScanInflux, utils checkmarxoneExecuteScanUtils) (checkmarxone.Scan, error) {
	var scan checkmarxone.Scan
    previousScans, err := sys.GetLastScans(project.ProjectID, 20)
	if err != nil && config.VerifyOnly {
		log.Entry().Warnf("Cannot load scans for project %v, verification only mode aborted", project.Name)
	}

	if len(previousScans) > 0 && config.VerifyOnly {
		err := cx1sh.verifyCxProjectCompliance(ctx, config, sys, group, project, previousScans[0], influx, utils)
		if err != nil {
			log.SetErrorCategory(log.ErrorCompliance)
			return scan, errors.Wrapf(err, "project %v not compliant", project.Name)
		}
	} else {
		zipFile, err := cx1sh.zipWorkspaceFiles(config.FilterPattern, utils)
		if err != nil {
			return scan, errors.Wrap(err, "failed to zip workspace files")
		}

		uploadUri, err := sys.UploadProjectSourceCode(project.ProjectID, zipFile.Name())
		if err != nil {
			return scan, errors.Wrapf(err, "failed to upload source code for project %v", project.Name)
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
			return scan, errors.Wrapf(err, "invalid configuration value for fullScanCycle %v, must be a positive int", config.FullScanCycle)
		}

		if config.IsOptimizedAndScheduled {
			incremental = false
		} else if incremental && config.FullScansScheduled && fullScanCycle > 0 && (cx1sh.getNumCoherentIncrementalScans(previousScans)+1)%fullScanCycle == 0 {
			incremental = false
		}

		// TODO: need to define the engines somewhere also.
		sastConfig := checkmarxone.ScanConfiguration{}
		sastConfig.ScanType = "sast"

        sastConfig.Values = make( map[string]string, 0 )
		sastConfig.Values["incremental"] = strconv.FormatBool(incremental)

		if len(config.Preset) > 0 {
			sastConfig.Values["presetName"] = config.Preset
		}

		if len(config.LanguageMode) > 0 {
			sastConfig.Values["languageMode"] = config.LanguageMode
		}

		return cx1sh.triggerScan(ctx, config, sys, project, uploadUri, config.PullRequestName, []checkmarxone.ScanConfiguration{sastConfig}, influx, utils)
	}
	return scan, nil
}

func (cx1sh *checkmarxoneExecuteScanHelper) triggerScan(ctx context.Context, config checkmarxoneExecuteScanOptions, sys checkmarxone.System, project checkmarxone.Project, repoUrl string, branch string, settings []checkmarxone.ScanConfiguration, influx *checkmarxoneExecuteScanInflux, utils checkmarxoneExecuteScanUtils) (checkmarxone.Scan, error) {
	scan, err := sys.ScanProjectZip(project.ProjectID, repoUrl, branch, settings)

	if err != nil {
		return scan, errors.Wrapf(err, "cannot scan project %v", project.Name)
	}

	log.Entry().Debugf("Scanning project %v: %v ", project.Name, scan.ScanID)

	

	err = cx1sh.pollScanStatus(sys, scan)
    if err != nil {
        return scan, errors.Wrap(err, "polling scan status failed")
    }

    log.Entry().Debugln("Scan finished")
    // original verify step here
    return scan, nil	
}

func (cx1sh *checkmarxoneExecuteScanHelper) createReportName(workspace, reportFileNameTemplate string) string {
	regExpFileName := regexp.MustCompile(`[^\w\d]`)
	timeStamp, _ := time.Now().Local().MarshalText()
	return filepath.Join(workspace, fmt.Sprintf(reportFileNameTemplate, regExpFileName.ReplaceAllString(string(timeStamp), "_")))
}

func (cx1sh *checkmarxoneExecuteScanHelper) pollScanStatus(sys checkmarxone.System, scan checkmarxone.Scan) error {
	statusDetails := "Scan phase: New"
	pastStatusDetails := statusDetails
	log.Entry().Info(statusDetails)
    status := "New"
	for {
		scan_refresh, err := sys.GetScan(scan.ScanID)

        if err != nil {
            log.Entry().Errorf( "Error while polling scan %v: %s", scan.ScanID, err )
            return err;
        }

        status = scan_refresh.Status
        workflow, err := sys.GetScanWorkflow( scan.ScanID )
        if err != nil {
            log.Entry().Errorf( "Error while polling scan %v: %s", scan.ScanID, err )
            return err;
        }

        statusDetails = workflow[ len(workflow)-1 ].Info

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
		return fmt.Errorf("scan canceled via web interface")
	}
	if status == "Failed" {
		return fmt.Errorf("Checkmarx One scan failed with the following error: %v", statusDetails)
	} 
	return nil
}

func (cx1sh *checkmarxoneExecuteScanHelper) downloadAndSaveReport(sys checkmarxone.System, reportFileName string, scan checkmarxone.Scan, utils checkmarxoneExecuteScanUtils) error {
    report, err := cx1sh.generateAndDownloadReport(sys, scan, "pdf")
    if err != nil {
        return errors.Wrap(err, "failed to download the report")
    }
    log.Entry().Debugf("Saving report to file %v...", reportFileName)
    return utils.WriteFile(reportFileName, report, 0o700)
}


func (cx1sh *checkmarxoneExecuteScanHelper) generateAndDownloadReport(sys checkmarxone.System, scan checkmarxone.Scan, reportType string) ([]byte, error) {
	var finalStatus checkmarxone.ReportStatus

    report, err := sys.RequestNewReport(scan.ScanID, scan.ProjectID, scan.Branch, reportType)
    if err != nil {
        return []byte{}, errors.Wrap(err, "failed to request new report")
    }
    for {
        finalStatus, err = sys.GetReportStatus(report)
        if err != nil {
            return []byte{}, errors.Wrap(err, "failed to get report status")
        }
        
        if finalStatus.Status == "completed" {
            break
        }
        time.Sleep(10 * time.Second)
    }
    if finalStatus.Status == "completed" {
        return sys.DownloadReport(finalStatus.ReportURL)
    }
	return []byte{}, fmt.Errorf("unexpected status %v recieved", finalStatus.Status)
}

func (cx1sh *checkmarxoneExecuteScanHelper) getNumCoherentIncrementalScans(scans []checkmarxone.Scan) int {
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

func (cx1sh *checkmarxoneExecuteScanHelper) getDetailedResults(config checkmarxoneExecuteScanOptions, group checkmarxone.Group, project checkmarxone.Project, scan checkmarxone.Scan, scanmeta checkmarxone.ScanMetadata, results []checkmarxone.ScanResults, /*sys checkmarxone.System, reportFileName string, scanID int,*/ utils checkmarxoneExecuteScanUtils) (map[string]interface{}, error) {
	log.Entry().Infof( "Test - getDetailedResults entry with %d results", len( results ) )

    resultMap := map[string]interface{}{}
    resultMap["InitiatorName"] = scan.Initiator
    resultMap["Owner"] = "Cx1 Gap: no project owner" //xmlResult.Owner
    resultMap["ScanId"] = scan.ScanID
    resultMap["ProjectId"] = project.ProjectID
    resultMap["ProjectName"] = project.Name
    resultMap["Group"] = group.GroupID
    resultMap["GroupFullPathOnReportDate"] = group.Name
    resultMap["ScanStart"] = scan.CreatedAt

    scanCreated, err := time.Parse(time.RFC3339, scan.CreatedAt )
    if err != nil {
        log.Entry().Warningf( "Failed to parse string %v into time: %s", scan.CreatedAt, err )
        resultMap["ScanTime"] = "Error parsing scan.CreatedAt"
    } else {
        scanFinished, err := time.Parse(time.RFC3339, scan.UpdatedAt)
        if err != nil {
            log.Entry().Warningf( "Failed to parse string %v into time: %s", scan.UpdatedAt, err )
            resultMap["ScanTime"] = "Error parsing scan.UpdatedAt"
        } else {
            difference := scanFinished.Sub(scanCreated)
            resultMap["ScanTime"] = difference.String()
        }
    }

    
    resultMap["LinesOfCodeScanned"] = scanmeta.LOC
    resultMap["FilesScanned"] = scanmeta.FileCount
    resultMap["CheckmarxVersion"] = "Cx1 Gap: No API for this, bundle.js"
    
    if scanmeta.IsIncremental {
        resultMap["ScanType"] = "Incremental"
    } else {
        resultMap["ScanType"] = "Full"
    }

    resultMap["Preset"] = scanmeta.PresetName
    resultMap["DeepLink"] = fmt.Sprintf( "%v/projects/%v/overview?branch=%v", config.ServerURL, project.ProjectID, scan.Branch )
    resultMap["ReportCreationTime"] = time.Now().String()
    resultMap["High"] = map[string]int{}
    resultMap["Medium"] = map[string]int{}
    resultMap["Low"] = map[string]int{}
    resultMap["Information"] = map[string]int{}


	if len(results) > 0 {       
		for _, result := range results {			
            key := "Information"
            switch result.Severity {
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
            case "URGENT", "URGENT " :
                auditState = "Urgent"
            case "PROPOSED_NOT_EXPLOITABLE":
                auditState = "ProposedNotExploitable"
            case "TO_VERIFY":
            default:
                auditState = "ToVerify"
            }
            submap[auditState]++

            //TODO
            // Review the change below - how does .FalsePositive work in old SAST XML, original line:
            //if result.FalsePositive != "True" {
            if auditState != "NotExploitable" {
                submap["NotFalsePositive"]++
            }
			
		}

		// if the flag is switched on, build the list  of Low findings per query
		if config.VulnerabilityThresholdLowPerQuery {
			var lowPerQuery = map[string]map[string]int{}

            for _, result := range results {
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
                case "URGENT", "URGENT " :
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

func (cx1sh *checkmarxoneExecuteScanHelper) zipWorkspaceFiles(filterPattern string, utils checkmarxoneExecuteScanUtils) (*os.File, error) {
	zipFileName := filepath.Join(utils.GetWorkspace(), "workspace.zip")
	patterns := piperutils.Trim(strings.Split(filterPattern, ","))
	sort.Strings(patterns)
	zipFile, err := os.Create(zipFileName)
	if err != nil {
		return zipFile, errors.Wrap(err, "failed to create archive of project sources")
	}
	defer zipFile.Close()


	err = cx1sh.zipFolder(utils.GetWorkspace(), zipFile, patterns, utils)
	if err != nil {
		return nil, errors.Wrap(err, "failed to compact folder")
	}
	return zipFile, nil
}

func (cx1sh *checkmarxoneExecuteScanHelper) zipFolder(source string, zipFile io.Writer, patterns []string, utils checkmarxoneExecuteScanUtils) error {
	archive := zip.NewWriter(zipFile)
	defer archive.Close()

    log.Entry().Infof( "Zipping %v into workspace.zip", source )

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

		noMatch, err := cx1sh.isFileNotMatchingPattern(patterns, path, info, utils)
		if err != nil || noMatch {
			return err
		}

		/*header, err := utils.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
		}

		cx1sh.adaptHeader(info, header)

		writer, err := archive.CreateHeader(header) */

        fileName := strings.TrimPrefix(path, baseDir)
        writer, err := archive.Create( fileName )
		if err != nil {
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
	err = cx1sh.handleZeroFilesZipped(source, err, fileCount)
	return err
}

func (cx1sh *checkmarxoneExecuteScanHelper) adaptHeader(info os.FileInfo, header *zip.FileHeader) {
	if info.IsDir() {
		header.Name += "/"
	} else {
		header.Method = zip.Deflate
	}
}

func (cx1sh *checkmarxoneExecuteScanHelper) handleZeroFilesZipped(source string, err error, fileCount int) error {
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
func (cx1sh *checkmarxoneExecuteScanHelper) isFileNotMatchingPattern(patterns []string, path string, info os.FileInfo, utils checkmarxoneExecuteScanUtils) (bool, error) {
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

func (cx1sh *checkmarxoneExecuteScanHelper) createToolRecordCx(utils checkmarxoneExecuteScanUtils, workspace string, config checkmarxoneExecuteScanOptions, results map[string]interface{} ) (string, error) {
	record := toolrecord.New(utils, workspace, "checkmarx", config.ServerURL)
	// Todo GroupId - see run_scan()
	// record.AddKeyData("team", XXX, resultMap["Group"], "")
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

func (cx1sh *checkmarxoneExecuteScanHelper) verifyCxProjectCompliance(ctx context.Context, config checkmarxoneExecuteScanOptions, sys checkmarxone.System, group checkmarxone.Group, project checkmarxone.Project, scan checkmarxone.Scan, influx *checkmarxoneExecuteScanInflux, utils checkmarxoneExecuteScanUtils) error {
	log.Entry().Info( "Test - verifyCxProjectCompliance entry" )
    
    var reports []piperutils.Path
	if config.GeneratePdfReport {
		pdfReportName := cx1sh.createReportName(utils.GetWorkspace(), "Cx1_SASTReport_%v.pdf")
		err := cx1sh.downloadAndSaveReport(sys, pdfReportName, scan, utils)
		if err != nil {
			log.Entry().Warning("Report download failed - continue processing ...")
		} else {
			reports = append(reports, piperutils.Path{Target: pdfReportName, Mandatory: true})
		}
	} else {
		log.Entry().Debug("Report generation is disabled via configuration")
	}

    results, err := sys.GetScanResults( scan.ScanID )
    if err != nil {
		return errors.Wrap(err, "failed to get detailed results")
	}

    scanmeta, err := sys.GetScanMetadata( scan.ScanID )
    if err != nil {
        log.Entry().Warnf( "Unable to fetch scan metadata for scan %v: %s", scan.ScanID, err )
    }

    detailedResults, err := cx1sh.getDetailedResults(config, group, project, scan, scanmeta, results, utils)
	if err != nil {
		return errors.Wrap(err, "failed to get detailed results")
	}

    // unclear if this step is required, currently 
	/*xmlReportName := cx1sh.createReportName(utils.GetWorkspace(), "CxSASTResults_%v.xml")
	results, err := cx1sh.getDetailedResults(config, sys, xmlReportName, scanID, utils)
	if err != nil {
		return errors.Wrap(err, "failed to get detailed results")
	}
	reports = append(reports, piperutils.Path{Target: xmlReportName})
    */


    /* Also unclear if SARIF-style report should be created from XML, or wait for support in the platform.
	// generate sarif report
	if config.ConvertToSarif {
		log.Entry().Info("Calling conversion to SARIF function.")
		sarif, err := checkmarxone.ConvertCxxmlToSarif(sys, xmlReportName, scanID)
		if err != nil {
			return fmt.Errorf("failed to generate SARIF")
		}
		paths, err := checkmarxone.WriteSarif(sarif)
		if err != nil {
			return fmt.Errorf("failed to write sarif")
		}
		reports = append(reports, paths...)
	} 
    */

	// create toolrecord
	toolRecordFileName, err := cx1sh.createToolRecordCx(utils, utils.GetWorkspace(), config, detailedResults)
	if err != nil {
		// do not fail until the framework is well established
		log.Entry().Warning("TR_CHECKMARX: Failed to create toolrecord file ...", err)
	} else {
		reports = append(reports, piperutils.Path{Target: toolRecordFileName})
	}

    
    /*
	// create JSON report (regardless vulnerabilityThreshold enabled or not)
	jsonReport := checkmarxone.CreateJSONReport(results)
	paths, err := checkmarxone.WriteJSONReport(jsonReport)
	if err != nil {
		log.Entry().Warning("failed to write JSON report...", err)
	} else {
		// add JSON report to archiving list
		reports = append(reports, paths...)
	}
	
    */

    links := []piperutils.Path{{Target: detailedResults["DeepLink"].(string), Name: "Checkmarx One Web UI"}}

	insecure := false
	var insecureResults []string
	var neutralResults []string

	if config.VulnerabilityThresholdEnabled {
		insecure, insecureResults, neutralResults = cx1sh.enforceThresholds(config, detailedResults)
		scanReport := checkmarxone.CreateCustomReport(detailedResults, insecureResults, neutralResults)

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

		paths, err := checkmarxone.WriteCustomReports(scanReport, project.Name, project.ProjectID)
		if err != nil {
			// do not fail until we have a better idea to handle it
			log.Entry().Warning("failed to write HTML/MarkDown report file ...", err)
		} else {
			reports = append(reports, paths...)
		}
	}
    

	piperutils.PersistReportsAndLinks("checkmarxoneExecuteScan", utils.GetWorkspace(), utils, reports, links)
	
    cx1sh.reportToInflux(detailedResults, influx)

	if insecure {
		if config.VulnerabilityThresholdResult == "FAILURE" {
			log.SetErrorCategory(log.ErrorCompliance)
			return fmt.Errorf("the project is not compliant - see report for details")
		}
		log.Entry().Errorf("Checkmarx One scan result set to %v, some results are not meeting defined thresholds. For details see the archived report.", config.VulnerabilityThresholdResult)
	} else {
		log.Entry().Infoln("Checkmarx One scan finished successfully")
	}
	return nil
}

func (cx1sh *checkmarxoneExecuteScanHelper) enforceThresholds(config checkmarxoneExecuteScanOptions, results map[string]interface{}) (bool, []string, []string) {
    log.Entry().Info( "Test - enforceThresholds entry" )
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

func (cx1sh *checkmarxoneExecuteScanHelper) reportToInflux(results map[string]interface{}, influx *checkmarxoneExecuteScanInflux) {
    log.Entry().Info( "Test - reportToInflux entry" )
	
    influx.checkmarxone_data.fields.high_issues = results["High"].(map[string]int)["Issues"]
    influx.checkmarxone_data.fields.high_not_false_postive = results["High"].(map[string]int)["NotFalsePositive"]
    influx.checkmarxone_data.fields.high_not_exploitable = results["High"].(map[string]int)["NotExploitable"]
    influx.checkmarxone_data.fields.high_confirmed = results["High"].(map[string]int)["Confirmed"]
    influx.checkmarxone_data.fields.high_urgent = results["High"].(map[string]int)["Urgent"]
    influx.checkmarxone_data.fields.high_proposed_not_exploitable = results["High"].(map[string]int)["ProposedNotExploitable"]
    influx.checkmarxone_data.fields.high_to_verify = results["High"].(map[string]int)["ToVerify"]
    influx.checkmarxone_data.fields.medium_issues = results["Medium"].(map[string]int)["Issues"]
    influx.checkmarxone_data.fields.medium_not_false_postive = results["Medium"].(map[string]int)["NotFalsePositive"]
    influx.checkmarxone_data.fields.medium_not_exploitable = results["Medium"].(map[string]int)["NotExploitable"]
    influx.checkmarxone_data.fields.medium_confirmed = results["Medium"].(map[string]int)["Confirmed"]
    influx.checkmarxone_data.fields.medium_urgent = results["Medium"].(map[string]int)["Urgent"]
    influx.checkmarxone_data.fields.medium_proposed_not_exploitable = results["Medium"].(map[string]int)["ProposedNotExploitable"]
    influx.checkmarxone_data.fields.medium_to_verify = results["Medium"].(map[string]int)["ToVerify"]
    influx.checkmarxone_data.fields.low_issues = results["Low"].(map[string]int)["Issues"]
    influx.checkmarxone_data.fields.low_not_false_postive = results["Low"].(map[string]int)["NotFalsePositive"]
    influx.checkmarxone_data.fields.low_not_exploitable = results["Low"].(map[string]int)["NotExploitable"]
    influx.checkmarxone_data.fields.low_confirmed = results["Low"].(map[string]int)["Confirmed"]
    influx.checkmarxone_data.fields.low_urgent = results["Low"].(map[string]int)["Urgent"]
    influx.checkmarxone_data.fields.low_proposed_not_exploitable = results["Low"].(map[string]int)["ProposedNotExploitable"]
    influx.checkmarxone_data.fields.low_to_verify = results["Low"].(map[string]int)["ToVerify"]
    influx.checkmarxone_data.fields.information_issues = results["Information"].(map[string]int)["Issues"]
    influx.checkmarxone_data.fields.information_not_false_postive = results["Information"].(map[string]int)["NotFalsePositive"]
    influx.checkmarxone_data.fields.information_not_exploitable = results["Information"].(map[string]int)["NotExploitable"]
    influx.checkmarxone_data.fields.information_confirmed = results["Information"].(map[string]int)["Confirmed"]
    influx.checkmarxone_data.fields.information_urgent = results["Information"].(map[string]int)["Urgent"]
    influx.checkmarxone_data.fields.information_proposed_not_exploitable = results["Information"].(map[string]int)["ProposedNotExploitable"]
    influx.checkmarxone_data.fields.information_to_verify = results["Information"].(map[string]int)["ToVerify"]
    influx.checkmarxone_data.fields.initiator_name = results["InitiatorName"].(string)
    influx.checkmarxone_data.fields.owner = results["Owner"].(string)
    influx.checkmarxone_data.fields.scan_id = results["ScanId"].(string)
    influx.checkmarxone_data.fields.project_id = results["ProjectId"].(string)
    influx.checkmarxone_data.fields.projectName = results["ProjectName"].(string)
    influx.checkmarxone_data.fields.group = results["Group"].(string)
    influx.checkmarxone_data.fields.group_full_path_on_report_date = results["GroupFullPathOnReportDate"].(string)
    influx.checkmarxone_data.fields.scan_start = results["ScanStart"].(string)
    influx.checkmarxone_data.fields.scan_time = results["ScanTime"].(string)
    influx.checkmarxone_data.fields.lines_of_code_scanned = results["LinesOfCodeScanned"].(int)
    influx.checkmarxone_data.fields.files_scanned = results["FilesScanned"].(int)
    influx.checkmarxone_data.fields.checkmarxone_version = results["CheckmarxVersion"].(string)
    influx.checkmarxone_data.fields.scan_type = results["ScanType"].(string)
    influx.checkmarxone_data.fields.preset = results["Preset"].(string)
    influx.checkmarxone_data.fields.deep_link = results["DeepLink"].(string)
    influx.checkmarxone_data.fields.report_creation_time = results["ReportCreationTime"].(string)
}

