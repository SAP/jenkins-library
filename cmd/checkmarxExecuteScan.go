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
	"github.com/bmatcuk/doublestar"
)

func checkmarxExecuteScan(config checkmarxExecuteScanOptions, influx *checkmarxExecuteScanInflux) error {
	client := &piperHttp.Client{}
	sys, err := checkmarx.NewSystemInstance(client, config.CheckmarxServerURL, config.Username, config.Password)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Failed to create Checkmarx client talking to URL %v", config.CheckmarxServerURL)
	}
	return runScan(config, sys, "./", influx)
}

func runScan(config checkmarxExecuteScanOptions, sys checkmarx.System, workspace string, influx *checkmarxExecuteScanInflux) error {
	projectName := config.CheckmarxProject

	teams := sys.GetTeams()
	team := checkmarx.Team{}
	if len(teams) > 0 {
		if len(config.TeamName) > 0 {
			team = sys.GetTeamByName(teams, config.TeamName)
		} else {
			team = sys.GetTeamByID(teams, config.CheckmarxGroupID)
		}
	}
	if len(team.ID) == 0 {
		log.Entry().Fatalf("Failed to identify team by teamName %v as well as by checkmarxGroupId %v", config.TeamName, config.CheckmarxGroupID)
	}

	projects := sys.GetProjects(team.ID)
	var project checkmarx.Project
	if len(config.PullRequestName) > 0 {
		projectName = fmt.Sprintf("%v_%v", config.CheckmarxProject, config.PullRequestName)
		project = sys.GetProjectByName(projects, projectName)
		if project.Name != projectName {
			project = sys.GetProjectByName(projects, config.CheckmarxProject)
			if project.ID != 0 {
				ok, branchProject := sys.GetProjectByID(sys.CreateBranch(project.ID, projectName))
				if !ok {
					log.Entry().Fatalf("Failed to create branch %v for project %v", projectName, config.CheckmarxProject)
				}
				project = branchProject
			}
		}
	} else {
		project = sys.GetProjectByName(projects, projectName)
	}

	if project.Name == projectName {
		log.Entry().Debugf("Project %v exists...", projectName)
	} else {
		log.Entry().Debugf("Project %v does not exist, starting to create it...", projectName)
		ok, projectCreateResult := sys.CreateProject(projectName, team.ID)
		if ok {
			if len(config.Preset) > 0 {
				presets := sys.GetPresets()
				var preset checkmarx.Preset
				presetID, err := strconv.Atoi(config.Preset)
				var configuredPresetID int
				var configuredPresetName string
				if err != nil {
					preset = sys.GetPresetByName(presets, config.Preset)
					configuredPresetName = config.Preset
				} else {
					preset = sys.GetPresetByID(presets, presetID)
					configuredPresetID = presetID
				}

				if configuredPresetID > 0 && preset.ID == configuredPresetID || len(configuredPresetName) > 0 && preset.Name == configuredPresetName {
					configurationUpdated := sys.UpdateProjectConfiguration(projectCreateResult.ID, preset.ID, config.EngineConfiguration)
					if configurationUpdated {
						log.Entry().Debugf("Configuration of project %v updated", project.Name)
					} else {
						log.Entry().Fatalf("Updating configuration of project %v failed", project.Name)
					}
				} else {
					log.Entry().Fatalf("Preset %v not found, creation of project %v failed", config.Preset, project.Name)
				}
			} else {
				log.Entry().Fatalf("Preset not specified, creation of project %v failed", project.Name)
			}
			projects := sys.GetProjects(team.ID)
			project := sys.GetProjectByName(projects, projectName)
			log.Entry().Debugf("New Project %v created", project.Name)
		} else {
			log.Entry().Fatalf("Cannot create project %v", config.CheckmarxProject)
		}
	}

	zipFileName := filepath.Join(workspace, "workspace.zip")
	patterns := strings.Split(config.FilterPattern, ",")
	sort.Strings(patterns)
	zipFile, err := os.Create(zipFileName)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to create archive of project sources")
	}
	defer zipFile.Close()
	zipFolder(workspace, zipFile, patterns)
	sourceCodeUploaded := sys.UploadProjectSourceCode(project.ID, zipFileName)
	if sourceCodeUploaded {
		log.Entry().Debugf("Source code uploaded for project %v", projectName)
		zipFile.Close()
		err := os.Remove(zipFileName)
		if err != nil {
			log.Entry().WithError(err).Warnf("Failed to delete zipped source code for project %v", projectName)
		}

		incremental := config.Incremental
		fullScanCycle, err := strconv.Atoi(config.FullScanCycle)
		if err != nil {
			log.Entry().WithError(err).Fatalf("Invalid configuration value for fullScanCycle %v, must be a positive int", config.FullScanCycle)
		}
		if incremental && config.FullScansScheduled && fullScanCycle > 0 && (getNumCoherentIncrementalScans(sys, project.ID)+1)%fullScanCycle == 0 {
			incremental = false
		}

		projectIsScanning, scan := sys.ScanProject(project.ID, incremental, false, false)
		if projectIsScanning {
			log.Entry().Debugf("Scanning project %v ", projectName)
			status := "New"
			pastStatus := status
			log.Entry().Debugf("Scan phase %v", status)
			for true {
				status = sys.GetScanStatus(scan.ID)
				if status != "Scanning" && (status == "Finished" || status == "Canceled" || status == "Failed") {
					break
				}
				if pastStatus != status {
					log.Entry().Debugf("Scan phase %v ", status)
					pastStatus = status
				}
				time.Sleep(10 * time.Second)
			}
			if status == "Canceled" {
				log.Entry().Fatalln("Scan canceled via web interface")
			}
			if status == "Failed" {
				log.Entry().Fatalln("Scan failed, please check the Checkmarx UI for details")
			} else {
				log.Entry().Debugln("Scan finished")

				if config.GeneratePdfReport {
					regExpFileName := regexp.MustCompile(`[^\w\d]`)
					timeStamp, _ := time.Now().Local().MarshalText()
					reportFileName := filepath.Join(workspace, fmt.Sprintf("CxSASTReport_%v.pdf", regExpFileName.ReplaceAllString(string(timeStamp), "_")))
					ok, report := generateAndDownloadReport(sys, scan.ID, "PDF")
					if ok {
						log.Entry().Debugf("Saving report to file %v...", reportFileName)
						ioutil.WriteFile(reportFileName, report, 0700)
					} else {
						log.Entry().Debugf("Failed to fetch report %v from backend...", reportFileName)
					}
				} else {
					log.Entry().Debug("Report generation is disabled via configuration")
				}

				results := getDetailedResults(sys, scan.ID)
				insecure := false

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

				if config.VulnerabilityThresholdEnabled {
					cxHighThreshold, _ := strconv.Atoi(config.VulnerabilityThresholdHigh)
					cxMediumThreshold, _ := strconv.Atoi(config.VulnerabilityThresholdMedium)
					cxLowThreshold, _ := strconv.Atoi(config.VulnerabilityThresholdMedium)
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
						lowAudited := results["Low"].(map[string]int)["Issues"] - results["Low"].(map[string]int)["NotFalsePositive"]
						lowOverall := results["Low"].(map[string]int)["Issues"]
						if lowOverall == 0 {
							lowAudited = 1
							lowOverall = 1
						}
						highValue = highAudited / highOverall * 100
						mediumValue = mediumAudited / mediumOverall * 100
						lowValue = lowAudited / lowOverall * 100

						if highValue < cxHighThreshold {
							insecure = true
							highViolation = "<--"
						}
						if mediumValue < cxMediumThreshold {
							insecure = true
							mediumViolation = "<--"
						}
						if lowValue < cxLowThreshold {
							insecure = true
							lowViolation = "<--"
						}
					}
					if config.VulnerabilityThresholdUnit == "absolute" {
						unit = ""
						if highValue > cxHighThreshold {
							insecure = true
							highViolation = "<--"
						}
						if mediumValue > cxMediumThreshold {
							insecure = true
							mediumViolation = "<--"
						}
						if lowValue > cxLowThreshold {
							insecure = true
							lowViolation = "<--"
						}
					}

					log.Entry().Errorln("")
					log.Entry().Errorf("High %v%v %v", highValue, unit, highViolation)
					log.Entry().Errorf("Medium %v%v %v", mediumValue, unit, mediumViolation)
					log.Entry().Errorf("Low %v%v %v", lowValue, unit, lowViolation)
					log.Entry().Errorln("")
				}

				if insecure {
					if config.VulnerabilityThresholdResult == "FAILURE" {
						log.Entry().Fatalln("Checkmarx scan failed, the project is not compliant. For details see the archived report.")
					} else {
						log.Entry().Errorf("Checkmarx scan result set to %v, some results are not meeting defined thresholds. For details see the archived report.", config.VulnerabilityThresholdResult)
					}
				} else {
					log.Entry().Infoln("Checkmarx scan finished")
				}
			}
		} else {
			log.Entry().Fatalf("Cannot scan project %v", projectName)
		}
	} else {
		log.Entry().Fatalf("Cannot upload source code for project %v", projectName)
	}
	return nil
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

func getDetailedResults(sys checkmarx.System, scanID int) map[string]interface{} {
	resultMap := map[string]interface{}{}
	ok, data := generateAndDownloadReport(sys, scanID, "XML")
	if ok && len(data) > 0 {
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
