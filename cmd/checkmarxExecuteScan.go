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
)

func checkmarxExecuteScan(config checkmarxExecuteScanOptions, telemetryData *telemetry.CustomData, influx *checkmarxExecuteScanInflux) error {
	client := &piperHttp.Client{}
	sys, err := checkmarx.NewSystemInstance(client, config.ServerURL, config.Username, config.Password)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Failed to create Checkmarx client talking to URL %v", config.ServerURL)
	}
	runScan(config, sys, "./", influx)
	return nil
}

func runScan(config checkmarxExecuteScanOptions, sys checkmarx.System, workspace string, influx *checkmarxExecuteScanInflux) {

	team := loadTeam(sys, config.TeamName, config.TeamID)
	projectName := config.ProjectName

	project := loadExistingProject(sys, config.ProjectName, config.PullRequestName, team.ID)
	if project.Name == projectName {
		log.Entry().Debugf("Project %v exists...", projectName)
	} else {
		log.Entry().Debugf("Project %v does not exist, starting to create it...", projectName)
		project = createAndConfigureNewProject(sys, projectName, team.ID, config.Preset, config.SourceEncoding)
	}

	uploadAndScan(config, sys, project, workspace, influx)
}

func loadTeam(sys checkmarx.System, teamName, teamID string) checkmarx.Team {
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
		log.Entry().Fatalf("Failed to identify team by teamName %v as well as by checkmarxGroupId %v", teamName, teamID)
	}
	return team
}

func loadExistingProject(sys checkmarx.System, initialProjectName, pullRequestName, teamID string) checkmarx.Project {
	var project checkmarx.Project
	projectName := initialProjectName
	if len(pullRequestName) > 0 {
		projectName = fmt.Sprintf("%v_%v", initialProjectName, pullRequestName)
		projects := sys.GetProjectsByNameAndTeam(projectName, teamID)
		if len(projects) == 0 {
			projects = sys.GetProjectsByNameAndTeam(initialProjectName, teamID)
			if len(projects) > 0 {
				ok, branchProject := sys.GetProjectByID(sys.CreateBranch(projects[0].ID, projectName))
				if !ok {
					log.Entry().Fatalf("Failed to create branch %v for project %v", projectName, initialProjectName)
				}
				project = branchProject
			}
		}
	} else {
		projects := sys.GetProjectsByNameAndTeam(projectName, teamID)
		if len(projects) > 0 {
			project = projects[0]
			log.Entry().Debugf("Loaded project with name %v", project.Name)
		}
	}
	return project
}

func zipWorkspaceFiles(workspace, filterPattern string) *os.File {
	zipFileName := filepath.Join(workspace, "workspace.zip")
	patterns := strings.Split(filterPattern, ",")
	sort.Strings(patterns)
	zipFile, err := os.Create(zipFileName)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to create archive of project sources")
	}
	defer zipFile.Close()
	zipFolder(workspace, zipFile, patterns)
	return zipFile
}

func uploadAndScan(config checkmarxExecuteScanOptions, sys checkmarx.System, project checkmarx.Project, workspace string, influx *checkmarxExecuteScanInflux) {
	zipFile := zipWorkspaceFiles(workspace, config.FilterPattern)
	sourceCodeUploaded := sys.UploadProjectSourceCode(project.ID, zipFile.Name())
	if sourceCodeUploaded {
		log.Entry().Debugf("Source code uploaded for project %v", project.Name)
		zipFile.Close()
		err := os.Remove(zipFile.Name())
		if err != nil {
			log.Entry().WithError(err).Warnf("Failed to delete zipped source code for project %v", project.Name)
		}

		incremental := config.Incremental
		fullScanCycle, err := strconv.Atoi(config.FullScanCycle)
		if err != nil {
			log.Entry().WithError(err).Fatalf("Invalid configuration value for fullScanCycle %v, must be a positive int", config.FullScanCycle)
		}
		if incremental && config.FullScansScheduled && fullScanCycle > 0 && (getNumCoherentIncrementalScans(sys, project.ID)+1)%fullScanCycle == 0 {
			incremental = false
		}

		triggerScan(config, sys, project, workspace, incremental, influx)
	} else {
		log.Entry().Fatalf("Cannot upload source code for project %v", project.Name)
	}
}

func triggerScan(config checkmarxExecuteScanOptions, sys checkmarx.System, project checkmarx.Project, workspace string, incremental bool, influx *checkmarxExecuteScanInflux) {
	projectIsScanning, scan := sys.ScanProject(project.ID, incremental, false, !config.AvoidDuplicateProjectScans)
	if projectIsScanning {
		log.Entry().Debugf("Scanning project %v ", project.Name)
		pollScanStatus(sys, scan)

		log.Entry().Debugln("Scan finished")

		var reports []piperutils.Path
		if config.GeneratePdfReport {
			pdfReportName := createReportName(workspace, "CxSASTReport_%v.pdf")
			ok := downloadAndSaveReport(sys, pdfReportName, scan)
			if ok {
				reports = append(reports, piperutils.Path{Target: pdfReportName, Mandatory: true})
			}
		} else {
			log.Entry().Debug("Report generation is disabled via configuration")
		}

		xmlReportName := createReportName(workspace, "CxSASTResults_%v.xml")
		results := getDetailedResults(sys, xmlReportName, scan.ID)
		reports = append(reports, piperutils.Path{Target: xmlReportName})
		links := []piperutils.Path{piperutils.Path{Target: results["DeepLink"].(string), Name: "Checkmarx Web UI"}}
		piperutils.PersistReportsAndLinks("checkmarxExecuteScan", workspace, reports, links)

		reportToInflux(results, influx)

		insecure := false
		if config.VulnerabilityThresholdEnabled {
			insecure = enforceThresholds(config, results)
		}

		if insecure {
			if config.VulnerabilityThresholdResult == "FAILURE" {
				log.Entry().Fatalln("Checkmarx scan failed, the project is not compliant. For details see the archived report.")
			}
			log.Entry().Errorf("Checkmarx scan result set to %v, some results are not meeting defined thresholds. For details see the archived report.", config.VulnerabilityThresholdResult)
		} else {
			log.Entry().Infoln("Checkmarx scan finished")
		}
	} else {
		log.Entry().Fatalf("Cannot scan project %v", project.Name)
	}
}

func createReportName(workspace, reportFileNameTemplate string) string {
	regExpFileName := regexp.MustCompile(`[^\w\d]`)
	timeStamp, _ := time.Now().Local().MarshalText()
	return filepath.Join(workspace, fmt.Sprintf(reportFileNameTemplate, regExpFileName.ReplaceAllString(string(timeStamp), "_")))
}

func pollScanStatus(sys checkmarx.System, scan checkmarx.Scan) {
	status := "Scan phase: New"
	pastStatus := status
	log.Entry().Info(status)
	for true {
		stepDetail := "..."
		stageDetail := "..."
		status, detail := sys.GetScanStatusAndDetail(scan.ID)
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
		log.Entry().Fatalln("Scan canceled via web interface")
	}
	if status == "Failed" {
		log.Entry().Fatalln("Scan failed, please check the Checkmarx UI for details")
	}
}

func reportToInflux(results map[string]interface{}, influx *checkmarxExecuteScanInflux) {
	influx.checkmarx_data.fields.high_issues = strconv.Itoa(results["High"].(map[string]int)["Issues"])
	influx.checkmarx_data.fields.high_not_false_postive = strconv.Itoa(results["High"].(map[string]int)["NotFalsePositive"])
	influx.checkmarx_data.fields.high_not_exploitable = strconv.Itoa(results["High"].(map[string]int)["NotExploitable"])
	influx.checkmarx_data.fields.high_confirmed = strconv.Itoa(results["High"].(map[string]int)["Confirmed"])
	influx.checkmarx_data.fields.high_urgent = strconv.Itoa(results["High"].(map[string]int)["Urgent"])
	influx.checkmarx_data.fields.high_proposed_not_exploitable = strconv.Itoa(results["High"].(map[string]int)["ProposedNotExploitable"])
	influx.checkmarx_data.fields.high_to_verify = strconv.Itoa(results["High"].(map[string]int)["ToVerify"])
	influx.checkmarx_data.fields.medium_issues = strconv.Itoa(results["Medium"].(map[string]int)["Issues"])
	influx.checkmarx_data.fields.medium_not_false_postive = strconv.Itoa(results["Medium"].(map[string]int)["NotFalsePositive"])
	influx.checkmarx_data.fields.medium_not_exploitable = strconv.Itoa(results["Medium"].(map[string]int)["NotExploitable"])
	influx.checkmarx_data.fields.medium_confirmed = strconv.Itoa(results["Medium"].(map[string]int)["Confirmed"])
	influx.checkmarx_data.fields.medium_urgent = strconv.Itoa(results["Medium"].(map[string]int)["Urgent"])
	influx.checkmarx_data.fields.medium_proposed_not_exploitable = strconv.Itoa(results["Medium"].(map[string]int)["ProposedNotExploitable"])
	influx.checkmarx_data.fields.medium_to_verify = strconv.Itoa(results["Medium"].(map[string]int)["ToVerify"])
	influx.checkmarx_data.fields.low_issues = strconv.Itoa(results["Low"].(map[string]int)["Issues"])
	influx.checkmarx_data.fields.low_not_false_postive = strconv.Itoa(results["Low"].(map[string]int)["NotFalsePositive"])
	influx.checkmarx_data.fields.low_not_exploitable = strconv.Itoa(results["Low"].(map[string]int)["NotExploitable"])
	influx.checkmarx_data.fields.low_confirmed = strconv.Itoa(results["Low"].(map[string]int)["Confirmed"])
	influx.checkmarx_data.fields.low_urgent = strconv.Itoa(results["Low"].(map[string]int)["Urgent"])
	influx.checkmarx_data.fields.low_proposed_not_exploitable = strconv.Itoa(results["Low"].(map[string]int)["ProposedNotExploitable"])
	influx.checkmarx_data.fields.low_to_verify = strconv.Itoa(results["Low"].(map[string]int)["ToVerify"])
	influx.checkmarx_data.fields.information_issues = strconv.Itoa(results["Information"].(map[string]int)["Issues"])
	influx.checkmarx_data.fields.information_not_false_postive = strconv.Itoa(results["Information"].(map[string]int)["NotFalsePositive"])
	influx.checkmarx_data.fields.information_not_exploitable = strconv.Itoa(results["Information"].(map[string]int)["NotExploitable"])
	influx.checkmarx_data.fields.information_confirmed = strconv.Itoa(results["Information"].(map[string]int)["Confirmed"])
	influx.checkmarx_data.fields.information_urgent = strconv.Itoa(results["Information"].(map[string]int)["Urgent"])
	influx.checkmarx_data.fields.information_proposed_not_exploitable = strconv.Itoa(results["Information"].(map[string]int)["ProposedNotExploitable"])
	influx.checkmarx_data.fields.information_to_verify = strconv.Itoa(results["Information"].(map[string]int)["ToVerify"])
	influx.checkmarx_data.fields.initiator_name = results["InitiatorName"].(string)
	influx.checkmarx_data.fields.owner = results["Owner"].(string)
	influx.checkmarx_data.fields.scan_id = results["ScanId"].(string)
	influx.checkmarx_data.fields.project_id = results["ProjectId"].(string)
	influx.checkmarx_data.fields.project_name = results["ProjectName"].(string)
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

func downloadAndSaveReport(sys checkmarx.System, reportFileName string, scan checkmarx.Scan) bool {
	ok, report := generateAndDownloadReport(sys, scan.ID, "PDF")
	if ok {
		log.Entry().Debugf("Saving report to file %v...", reportFileName)
		ioutil.WriteFile(reportFileName, report, 0700)
		return true
	}
	log.Entry().Debugf("Failed to fetch report %v from backend...", reportFileName)
	return false
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
		unit = "findings"
		if highValue > cxHighThreshold {
			insecure = true
			highViolation = fmt.Sprintf("<-- %v %v deviation", highValue-cxHighThreshold, unit)
		}
		if mediumValue > cxMediumThreshold {
			insecure = true
			mediumViolation = fmt.Sprintf("<-- %v %v deviation", mediumValue-cxMediumThreshold, unit)
		}
		if lowValue > cxLowThreshold {
			insecure = true
			lowViolation = fmt.Sprintf("<-- %v %v deviation", lowValue-cxLowThreshold, unit)
		}
	}

	log.Entry().Infoln("")
	log.Entry().Infof("High %v%v %v", highValue, unit, highViolation)
	log.Entry().Infof("Medium %v%v %v", mediumValue, unit, mediumViolation)
	log.Entry().Infof("Low %v%v %v", lowValue, unit, lowViolation)
	log.Entry().Infoln("")

	return insecure
}

func createAndConfigureNewProject(sys checkmarx.System, projectName, teamID, presetValue, engineConfiguration string) checkmarx.Project {
	ok, projectCreateResult := sys.CreateProject(projectName, teamID)
	if ok {
		if len(presetValue) > 0 {
			ok, preset := loadPreset(sys, presetValue)
			if ok {
				configurationUpdated := sys.UpdateProjectConfiguration(projectCreateResult.ID, preset.ID, engineConfiguration)
				if configurationUpdated {
					log.Entry().Debugf("Configuration of project %v updated", projectName)
				} else {
					log.Entry().Fatalf("Updating configuration of project %v failed", projectName)
				}
			} else {
				log.Entry().Fatalf("Preset %v not found, creation of project %v failed", presetValue, projectName)
			}
		} else {
			log.Entry().Fatalf("Preset not specified, creation of project %v failed", projectName)
		}
		projects := sys.GetProjectsByNameAndTeam(projectName, teamID)
		if len(projects) > 0 {
			log.Entry().Debugf("New Project %v created", projectName)
			return projects[0]
		}
		log.Entry().Fatalf("Failed to load newly created project %v", projectName)
	}
	log.Entry().Fatalf("Cannot create project %v", projectName)
	return checkmarx.Project{}
}

func loadPreset(sys checkmarx.System, presetValue string) (bool, checkmarx.Preset) {
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
		log.Entry().Debugf("Loaded preset %v", preset.Name)
		return true, preset
	}
	return false, checkmarx.Preset{}
}

func generateAndDownloadReport(sys checkmarx.System, scanID int, reportType string) (bool, []byte) {
	success, report := sys.RequestNewReport(scanID, reportType)
	if success {
		finalStatus := 1
		for {
			finalStatus = sys.GetReportStatus(report.ReportID).Status.ID
			if finalStatus != 1 {
				break
			}
			time.Sleep(10 * time.Second)
		}
		if finalStatus == 2 {
			return sys.DownloadReport(report.ReportID)
		}
	}
	return false, []byte{}
}

func getNumCoherentIncrementalScans(sys checkmarx.System, projectID int) int {
	ok, scans := sys.GetScans(projectID)
	count := 0
	if ok {
		for _, scan := range scans {
			if !scan.IsIncremental {
				break
			}
			count++
		}
	}
	return count
}

func getDetailedResults(sys checkmarx.System, reportFileName string, scanID int) map[string]interface{} {
	resultMap := map[string]interface{}{}
	ok, data := generateAndDownloadReport(sys, scanID, "XML")
	if ok && len(data) > 0 {
		ioutil.WriteFile(reportFileName, data, 0700)
		var xmlResult checkmarx.DetailedResult
		err := xml.Unmarshal(data, &xmlResult)
		if err != nil {
			log.Entry().Fatalf("Failed to unmarshal XML report for scan %v: %s", scanID, err)
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
	return resultMap
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
		return err
	})

	return err
}

func filterFileGlob(patterns []string, path string, info os.FileInfo) bool {
	for index := 0; index < len(patterns); index++ {
		pattern := patterns[index]
		negative := false
		if strings.Index(pattern, "!") == 0 {
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
