package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/checkmarx"
	"github.com/bmatcuk/doublestar"
	"github.com/pkg/errors"
)

func checkmarxExecuteScan(myCheckmarxExecuteScanOptions checkmarxExecuteScanOptions) error {
	client := &piperHttp.Client{}
	sys, err := checkmarx.NewSystem(client, myCheckmarxExecuteScanOptions.CheckmarxServerURL, myCheckmarxExecuteScanOptions.Username, myCheckmarxExecuteScanOptions.Password)
	if err != nil {
		errors.Errorf("Failed to create Checkmarx client talking to URL %v with error %v", myCheckmarxExecuteScanOptions.CheckmarxServerURL, err)
	}

	projects := sys.GetProjects()
	project := sys.GetProjectByName(projects, myCheckmarxExecuteScanOptions.CheckmarxProject)
	if project.Name == myCheckmarxExecuteScanOptions.CheckmarxProject {
		fmt.Println("Project exists...")
	} else {
		teams := sys.GetTeams()
		team := checkmarx.Team{}
		if len(teams) > 1 {
			team = sys.GetTeamByName(teams, myCheckmarxExecuteScanOptions.TeamName)
		}
		if len(team.ID) == 0 {
			team = teams[0]
		}
		fmt.Println("Project does not exists...")
		projectCreated := sys.CreateProject(myCheckmarxExecuteScanOptions.CheckmarxProject, team.ID)
		if projectCreated {
			if len(myCheckmarxExecuteScanOptions.Preset) > 0 {
				presets := sys.GetPresets()
				preset := sys.GetPresetByName(presets, myCheckmarxExecuteScanOptions.Preset)
				if preset.Name == myCheckmarxExecuteScanOptions.Preset {
					configurationUpdated := sys.UpdateProjectConfiguration(project.ID, preset.ID, myCheckmarxExecuteScanOptions.EngineConfiguration)
					if configurationUpdated {
						fmt.Println("Configuration of project updated: " + project.Name)
					} else {
						fmt.Println("Updating project configuration failed: " + project.Name)
						os.Exit(10)
					}
				} else {
					fmt.Println("Preset not found, project creation failed: " + project.Name)
					os.Exit(10)
				}
			} else {
				fmt.Println("Preset not specified, project creation failed: " + project.Name)
				os.Exit(10)
			}
			projects := sys.GetProjects()
			project := sys.GetProjectByName(projects, myCheckmarxExecuteScanOptions.CheckmarxProject)
			fmt.Println("New Project Created : " + project.Name)
		} else {
			fmt.Println("Cannot create Project : " + myCheckmarxExecuteScanOptions.CheckmarxProject)
			os.Exit(10)
		}
	}

	zipFileName := "workspace.zip"
	patterns := strings.Split(myCheckmarxExecuteScanOptions.FilterPattern, ",")
	sort.Strings(patterns)
	zipFile, err := os.Create(zipFileName)
	if err != nil {
		return err
	}
	defer zipFile.Close()
	zipFolder("./", zipFile, patterns)
	sourceCodeUploaded := sys.UploadProjectSourceCode(project.ID, zipFileName)
	if sourceCodeUploaded {
		fmt.Println("Source code Uploaded for Project : " + myCheckmarxExecuteScanOptions.CheckmarxProject)
		err := os.Remove(zipFileName)
		if err != nil {
			fmt.Printf("Error : %s\n", err)
		}

		projectIsScanning, scan := sys.ScanProject(project.ID)
		if projectIsScanning {
			fmt.Println("Scanning : " + myCheckmarxExecuteScanOptions.CheckmarxProject)
			status := "New"
			pastStatus := status
			fmt.Println("Scan phase : " + status)
			for true {
				status = sys.GetScanStatus(scan.ID)
				if status == "Finished" || status == "Canceled" {
					break
				}
				if pastStatus != status {
					fmt.Println("Scan phase : " + status)
					pastStatus = status
				}
			}
			if status == "Canceled" {
				fmt.Println("Scan Canceled via Web Interface")
				os.Exit(10)
			} else {
				fmt.Println("Scan Finished")
				results := sys.GetResults(scan.ID)
				insecure := false
				cxHighThreshold, _ := strconv.Atoi(myCheckmarxExecuteScanOptions.VulnerabilityThresholdHigh)
				if results.High > cxHighThreshold {
					insecure = true
				}
				cxMediumThreshold, _ := strconv.Atoi(myCheckmarxExecuteScanOptions.VulnerabilityThresholdMedium)
				if results.Medium > cxMediumThreshold {
					insecure = true
				}
				cxLowThreshold, _ := strconv.Atoi(myCheckmarxExecuteScanOptions.VulnerabilityThresholdMedium)
				if results.Low > cxLowThreshold {
					insecure = true
				}
				if insecure {
					fmt.Println("Insecure Application !")
					fmt.Println("")
					fmt.Println("High : " + strconv.Itoa(results.High))
					fmt.Println("Medium : " + strconv.Itoa(results.Medium))
					fmt.Println("Low : " + strconv.Itoa(results.Low))
					fmt.Println("")
					fmt.Println("Step Finished")
					os.Exit(10)
				} else {
					fmt.Println("Application Secured !")
					fmt.Println("")
					fmt.Println("High : " + strconv.Itoa(results.High))
					fmt.Println("Medium : " + strconv.Itoa(results.Medium))
					fmt.Println("Low : " + strconv.Itoa(results.Low))
					fmt.Println("")
					fmt.Println("Step Finished")
				}
			}
		} else {
			fmt.Println("Cannot scan Project : " + myCheckmarxExecuteScanOptions.CheckmarxProject)
			os.Exit(10)
		}
	} else {
		fmt.Println("Cannot upload source code for Project : " + myCheckmarxExecuteScanOptions.CheckmarxProject)
		os.Exit(10)
	}
	return nil
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
