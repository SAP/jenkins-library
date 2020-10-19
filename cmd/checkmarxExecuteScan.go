package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"encoding/xml"

	"github.com/SAP/jenkins-library/pkg/checkmarx"
	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/bmatcuk/doublestar"
	"github.com/pkg/errors"
)

func checkmarxExecuteScan(config checkmarxExecuteScanOptions, telemetryData *telemetry.CustomData, influx *checkmarxExecuteScanInflux) {
	client := &piperHttp.Client{}
	sys, err := checkmarx.NewSystemInstance(client, config.ServerURL, config.Username, config.Password)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Failed to create Checkmarx client talking to URL %v", config.ServerURL)
	}
	if err := runScan(config, sys, "./", influx); err != nil {
		log.Entry().WithError(err).Fatal("Failed to execute Checkmarx scan.")
	}
}

func runScan(config checkmarxExecuteScanOptions, sys checkmarx.System, workspace string, influx *checkmarxExecuteScanInflux) error {

	team, err := loadTeam(sys, config.TeamName, config.TeamID)
	if err != nil {
		return errors.Wrap(err, "failed to load team")
	}
	project, projectName, err := loadExistingProject(sys, config.ProjectName, config.PullRequestName, team.ID)
	if err != nil {
		return errors.Wrap(err, "error when trying to load project")
	}
	if project.Name == projectName {
		log.Entry().Infof("Project %v exists...", projectName)
		if len(config.Preset) > 0 {
			err := setPresetForProject(sys, project.ID, projectName, config.Preset, config.SourceEncoding)
			if err != nil {
				return errors.Wrapf(err, "failed to set preset %v for project %v", config.Preset, projectName)
			}
		}
	} else {
		log.Entry().Infof("Project %v does not exist, starting to create it...", projectName)
		project, err = createAndConfigureNewProject(sys, projectName, team.ID, config.Preset, config.SourceEncoding)
		if err != nil {
			return errors.Wrapf(err, "failed to create and configure new project %v", projectName)
		}
	}

	err = uploadAndScan(config, sys, project, workspace, influx)
	if err != nil {
		return errors.Wrap(err, "failed to run scan and upload result")
	}
	return nil
}

func loadTeam(sys checkmarx.System, teamName, teamID string) (checkmarx.Team, error) {
	teams := sys.GetTeams()
	team := checkmarx.Team{}
	if len(teams) > 0 {
		if len(teamName) > 0 {
			team = sys.FilterTeamByName(teams, teamName)
		} else {
			team = sys.FilterTeamByID(teams, teamID)
		}
	}
	if len(team.ID) == 0 {
		return team, fmt.Errorf("failed to identify team by teamName %v as well as by checkmarxGroupId %v", teamName, teamID)
	}
	return team, nil
}

func loadExistingProject(sys checkmarx.System, initialProjectName, pullRequestName, teamID string) (checkmarx.Project, string, error) {
	var project checkmarx.Project
	projectName := initialProjectName
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
		project = projects[0]
		log.Entry().Debugf("Loaded project with name %v", project.Name)
	}
	return project, projectName, nil
}

func zipWorkspaceFiles(workspace, filterPattern string) (*os.File, error) {
	zipFileName := filepath.Join(workspace, "workspace.zip")
	patterns := strings.Split(strings.ReplaceAll(strings.ReplaceAll(filterPattern, ", ", ","), " ,", ","), ",")
	sort.Strings(patterns)
	zipFile, err := os.Create(zipFileName)
	if err != nil {
		return zipFile, errors.Wrap(err, "failed to create archive of project sources")
	}
	defer zipFile.Close()
	zipFolder(workspace, zipFile, patterns)
	return zipFile, nil
}

func uploadAndScan(config checkmarxExecuteScanOptions, sys checkmarx.System, project checkmarx.Project, workspace string, influx *checkmarxExecuteScanInflux) error {
	previousScans, err := sys.GetScans(project.ID)
	if err != nil && config.VerifyOnly {
		log.Entry().Warnf("Cannot load scans for project %v, verification only mode aborted", project.Name)
	}
	if len(previousScans) > 0 && config.VerifyOnly {
		err := verifyCxProjectCompliance(config, sys, previousScans[0].ID, workspace, influx)
		if err != nil {
			log.SetErrorCategory(log.ErrorCompliance)
			return errors.Wrapf(err, "project %v not compliant", project.Name)
		}
	} else {
		zipFile, err := zipWorkspaceFiles(workspace, config.FilterPattern)
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

		if incremental && config.FullScansScheduled && fullScanCycle > 0 && (getNumCoherentIncrementalScans(sys, previousScans)+1)%fullScanCycle == 0 {
			incremental = false
		}

		return triggerScan(config, sys, project, workspace, incremental, influx)
	}
	return nil
}

func triggerScan(config checkmarxExecuteScanOptions, sys checkmarx.System, project checkmarx.Project, workspace string, incremental bool, influx *checkmarxExecuteScanInflux) error {
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
	return verifyCxProjectCompliance(config, sys, scan.ID, workspace, influx)
}

func verifyCxProjectCompliance(config checkmarxExecuteScanOptions, sys checkmarx.System, scanID int, workspace string, influx *checkmarxExecuteScanInflux) error {
	var reports []piperutils.Path
	if config.GeneratePdfReport {
		pdfReportName := createReportName(workspace, "CxSASTReport_%v.pdf")
		err := downloadAndSaveReport(sys, pdfReportName, scanID)
		if err != nil {
			log.Entry().Warning("Report download failed - continue processing ...")
		} else {
			reports = append(reports, piperutils.Path{Target: pdfReportName, Mandatory: true})
		}
	} else {
		log.Entry().Debug("Report generation is disabled via configuration")
	}

	xmlReportName := createReportName(workspace, "CxSASTResults_%v.xml")
	results, err := getDetailedResults(sys, xmlReportName, scanID)
	if err != nil {
		return errors.Wrap(err, "failed to get detailed results")
	}
	reports = append(reports, piperutils.Path{Target: xmlReportName})
	links := []piperutils.Path{{Target: results["DeepLink"].(string), Name: "Checkmarx Web UI"}}
	piperutils.PersistReportsAndLinks("checkmarxExecuteScan", workspace, reports, links)

	reportToInflux(results, influx)

	insecure := false
	if config.VulnerabilityThresholdEnabled {
		insecure = enforceThresholds(config, results)
	}

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
	for true {
		stepDetail := "..."
		stageDetail := "..."
		var detail checkmarx.ScanStatusDetail
		status, detail = sys.GetScanStatusAndDetail(scan.ID)
		if status == "Finished" || status == "Canceled" || status == "Failed" {
			break
		}
		if len(detail.Stage) > 0 {
			stageDetail = detail.Stage
		}
		if len(detail.Step) > 0 {
			stepDetail = detail.Step
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
		return fmt.Errorf("scan failed, please check the Checkmarx UI for details")
	}
	return nil
}

func reportToInflux(results map[string]interface{}, influx *checkmarxExecuteScanInflux) {
	influx.checkmarx_data.fields.high_issues = strconv.Itoa(results["High"].(map[string]int)["Issues"])
	influx.checkmarx_data.fields.high_not_false_positive = strconv.Itoa(results["High"].(map[string]int)["NotFalsePositive"])
	influx.checkmarx_data.fields.high_not_exploitable = strconv.Itoa(results["High"].(map[string]int)["NotExploitable"])
	influx.checkmarx_data.fields.high_confirmed = strconv.Itoa(results["High"].(map[string]int)["Confirmed"])
	influx.checkmarx_data.fields.high_urgent = strconv.Itoa(results["High"].(map[string]int)["Urgent"])
	influx.checkmarx_data.fields.high_proposed_not_exploitable = strconv.Itoa(results["High"].(map[string]int)["ProposedNotExploitable"])
	influx.checkmarx_data.fields.high_to_verify = strconv.Itoa(results["High"].(map[string]int)["ToVerify"])
	influx.checkmarx_data.fields.medium_issues = strconv.Itoa(results["Medium"].(map[string]int)["Issues"])
	influx.checkmarx_data.fields.medium_not_false_positive = strconv.Itoa(results["Medium"].(map[string]int)["NotFalsePositive"])
	influx.checkmarx_data.fields.medium_not_exploitable = strconv.Itoa(results["Medium"].(map[string]int)["NotExploitable"])
	influx.checkmarx_data.fields.medium_confirmed = strconv.Itoa(results["Medium"].(map[string]int)["Confirmed"])
	influx.checkmarx_data.fields.medium_urgent = strconv.Itoa(results["Medium"].(map[string]int)["Urgent"])
	influx.checkmarx_data.fields.medium_proposed_not_exploitable = strconv.Itoa(results["Medium"].(map[string]int)["ProposedNotExploitable"])
	influx.checkmarx_data.fields.medium_to_verify = strconv.Itoa(results["Medium"].(map[string]int)["ToVerify"])
	influx.checkmarx_data.fields.low_issues = strconv.Itoa(results["Low"].(map[string]int)["Issues"])
	influx.checkmarx_data.fields.low_not_false_positive = strconv.Itoa(results["Low"].(map[string]int)["NotFalsePositive"])
	influx.checkmarx_data.fields.low_not_exploitable = strconv.Itoa(results["Low"].(map[string]int)["NotExploitable"])
	influx.checkmarx_data.fields.low_confirmed = strconv.Itoa(results["Low"].(map[string]int)["Confirmed"])
	influx.checkmarx_data.fields.low_urgent = strconv.Itoa(results["Low"].(map[string]int)["Urgent"])
	influx.checkmarx_data.fields.low_proposed_not_exploitable = strconv.Itoa(results["Low"].(map[string]int)["ProposedNotExploitable"])
	influx.checkmarx_data.fields.low_to_verify = strconv.Itoa(results["Low"].(map[string]int)["ToVerify"])
	influx.checkmarx_data.fields.information_issues = strconv.Itoa(results["Information"].(map[string]int)["Issues"])
	influx.checkmarx_data.fields.information_not_false_positive = strconv.Itoa(results["Information"].(map[string]int)["NotFalsePositive"])
	influx.checkmarx_data.fields.information_not_exploitable = strconv.Itoa(results["Information"].(map[string]int)["NotExploitable"])
	influx.checkmarx_data.fields.information_confirmed = strconv.Itoa(results["Information"].(map[string]int)["Confirmed"])
	influx.checkmarx_data.fields.information_urgent = strconv.Itoa(results["Information"].(map[string]int)["Urgent"])
	influx.checkmarx_data.fields.information_proposed_not_exploitable = strconv.Itoa(results["Information"].(map[string]int)["ProposedNotExploitable"])
	influx.checkmarx_data.fields.information_to_verify = strconv.Itoa(results["Information"].(map[string]int)["ToVerify"])
	influx.checkmarx_data.fields.initiator_name = results["InitiatorName"].(string)
	influx.checkmarx_data.fields.owner = results["Owner"].(string)
	influx.checkmarx_data.fields.scan_id = results["ScanId"].(string)
	influx.checkmarx_data.fields.project_id = results["ProjectId"].(string)
	influx.checkmarx_data.fields.projectName = results["ProjectName"].(string)
	influx.checkmarx_data.fields.team = results["Team"].(string)
	influx.checkmarx_data.fields.team_full_path_on_report_date = results["TeamFullPathOnReportDate"].(string)
	influx.checkmarx_data.fields.scan_start = results["ScanStart"].(string)
	influx.checkmarx_data.fields.scan_time = results["ScanTime"].(string)
	influx.checkmarx_data.fields.lines_of_code_scanned = results["LinesOfCodeScanned"].(string)
	influx.checkmarx_data.fields.files_scanned = results["FilesScanned"].(string)
	influx.checkmarx_data.fields.checkmarx_version = results["CheckmarxVersion"].(string)
	influx.checkmarx_data.fields.scan_type = results["ScanType"].(string)
	influx.checkmarx_data.fields.preset = results["Preset"].(string)
	influx.checkmarx_data.fields.deep_link = results["DeepLink"].(string)
	influx.checkmarx_data.fields.report_creation_time = results["ReportCreationTime"].(string)
}

func downloadAndSaveReport(sys checkmarx.System, reportFileName string, scanID int) error {
	report, err := generateAndDownloadReport(sys, scanID, "PDF")
	if err != nil {
		return errors.Wrap(err, "failed to download the report")
	}
	log.Entry().Debugf("Saving report to file %v...", reportFileName)
	ioutil.WriteFile(reportFileName, report, 0700)
	return nil
}

func enforceThresholds(config checkmarxExecuteScanOptions, results map[string]interface{}) bool {
	insecure := false
	cxHighThreshold := config.VulnerabilityThresholdHigh
	cxMediumThreshold := config.VulnerabilityThresholdMedium
	cxLowThreshold := config.VulnerabilityThresholdLow
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
		if lowValue < cxLowThreshold {
			insecure = true
			lowViolation = fmt.Sprintf("<-- %v %v deviation", cxLowThreshold-lowValue, unit)
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

	log.Entry().Infoln("")
	log.Entry().Infof("High %v%v %v", highValue, unit, highViolation)
	log.Entry().Infof("Medium %v%v %v", mediumValue, unit, mediumViolation)
	log.Entry().Infof("Low %v%v %v", lowValue, unit, lowViolation)
	log.Entry().Infoln("")

	return insecure
}

func createAndConfigureNewProject(sys checkmarx.System, projectName, teamID, presetValue, engineConfiguration string) (checkmarx.Project, error) {
	if len(presetValue) == 0 {
		log.SetErrorCategory(log.ErrorConfiguration)
		return checkmarx.Project{}, fmt.Errorf("preset not specified, creation of project %v failed", projectName)
	}

	projectCreateResult, err := sys.CreateProject(projectName, teamID)
	if err != nil {
		return checkmarx.Project{}, errors.Wrapf(err, "cannot create project %v", projectName)
	}

	if err := setPresetForProject(sys, projectCreateResult.ID, projectName, presetValue, engineConfiguration); err != nil {
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
	presetID, err := strconv.Atoi(presetValue)
	var configuredPresetID int
	var configuredPresetName string
	if err != nil {
		preset = sys.FilterPresetByName(presets, presetValue)
		configuredPresetName = presetValue
	} else {
		preset = sys.FilterPresetByID(presets, presetID)
		configuredPresetID = presetID
	}

	if configuredPresetID > 0 && preset.ID == configuredPresetID || len(configuredPresetName) > 0 && preset.Name == configuredPresetName {
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
func setPresetForProject(sys checkmarx.System, projectID int, projectName, presetValue, engineConfiguration string) error {
	preset, err := loadPreset(sys, presetValue)
	if err != nil {
		return errors.Wrapf(err, "preset %v not found, configuration of project %v failed", presetValue, projectName)
	}

	err = sys.UpdateProjectConfiguration(projectID, preset.ID, engineConfiguration)
	if err != nil {
		return errors.Wrapf(err, "updating configuration of project %v failed", projectName)
	}
	log.Entry().Debugf("Configuration of project %v updated", projectName)
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

func getNumCoherentIncrementalScans(sys checkmarx.System, scans []checkmarx.ScanStatus) int {
	count := 0
	for _, scan := range scans {
		if !scan.IsIncremental {
			break
		}
		count++
	}
	return count
}

func getDetailedResults(sys checkmarx.System, reportFileName string, scanID int) (map[string]interface{}, error) {
	resultMap := map[string]interface{}{}
	data, err := generateAndDownloadReport(sys, scanID, "XML")
	if err != nil {
		return resultMap, errors.Wrap(err, "failed to download xml report")
	}
	if len(data) > 0 {
		ioutil.WriteFile(reportFileName, data, 0700)
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
	}
	return resultMap, nil
}

func zipFolder(source string, zipFile io.Writer, patterns []string) error {
	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	fileCount := 0
	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filterFileGlob(patterns, path, info) {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		fileCount++
		return err
	})
	log.Entry().Infof("Zipped %d files", fileCount)
	return err
}

func filterFileGlob(patterns []string, path string, info os.FileInfo) bool {
	for _, pattern := range patterns {
		negative := false
		if strings.HasPrefix(pattern, "!") {
			pattern = strings.TrimLeft(pattern, "!")
			negative = true
		}
		match, _ := doublestar.Match(pattern, path)
		if !info.IsDir() {
			if match && negative {
				return true
			} else if match && !negative {
				return false
			}
		} else {
			return false
		}
	}
	return true
}
