package cmd

import (
	"archive/zip"
	"fmt"
	"os"
	"io"
	"strconv"
	"path/filepath"
	"strings"

	"github.com/SAP/jenkins-library/pkg/checkmarx"
	"github.com/pkg/errors"
)

func main() {
	sourceDir := os.Getenv("source_dir")
	cxServerURL := os.Getenv("cx_server_url")
	cxUsername := os.Getenv("cx_username")
	cxPassword := os.Getenv("cx_password")

	cxProjectName := os.Getenv("cx_project_name")
	cxTeamName := os.Getenv("cx_team_name")
	cxPresetName := os.Getenv("cx_preset_name")
	cxExcludeFolders := os.Getenv("cx_exclude_folders")
	cxExcludeFiles := os.Getenv("cx_exclude_files")
	cxEngineConfiguration := os.Getenv("cx_engine_configuration")

	cxHighThreshold := os.Getenv("cx_high_threshold")
	cxMediumThreshold := os.Getenv("cx_medium_threshold")
	cxLowThreshold := os.Getenv("cx_low_threshold")

	if len(sourceDir) == 0 {
		fmt.Println("SOURCE_DIR is empty")
		os.Exit(10)
	} else {
		fmt.Println("SOURCE_DIR : " + sourceDir)
	}

	if len(cxServerURL) == 0 {
		fmt.Println("CX_SERVER_URL path is empty")
		os.Exit(10)
	} else {
		fmt.Println("CX_SERVER_URL : " + cxServerURL)
		cxServerURL = cxServerURL + "/cxrestapi/"
	}

	if len(cxUsername) == 0 {
		fmt.Println("CX_USERNAME is empty")
		os.Exit(10)
	} else {
		fmt.Println("CX_USERNAME : " + cxUsername)
	}

	if len(cxPassword) == 0 {
		fmt.Println("CX_PASSWORD is empty")
		os.Exit(10)
	}

	if len(cxProjectName) == 0 {
		fmt.Println("CX_PROJECT_NAME is empty")
		os.Exit(10)
	} else {
		fmt.Println("CX_PROJECT_NAME : " + cxProjectName)
	}

	if len(cxTeamName) == 0 {
		fmt.Println("CX_TEAM_NAME is empty")
		os.Exit(10)
	} else {
		fmt.Println("CX_TEAM_NAME : " + cxTeamName)
	}

	if len(cxPresetName) == 0 {
		fmt.Println("CX_PRESET_NAME is empty")
		os.Exit(10)
	} else {
		fmt.Println("CX_PRESET_NAME : " + cxPresetName)
	}

	if len(cxExcludeFolders) == 0 {
		fmt.Println("CX_EXCLUDE_FOLDERS is empty")
		cxExcludeFolders = ""
	} else {
		fmt.Println("CX_EXCLUDE_FOLDERS : " + cxExcludeFolders)
	}

	if len(cxExcludeFiles) == 0 {
		fmt.Println("CX_EXCLUDE_FILES is empty")
		cxExcludeFiles = ""
	} else {
		fmt.Println("CX_EXCLUDE_FILES : " + cxExcludeFiles)
	}

	if len(cxEngineConfiguration) == 0 {
		fmt.Println("CX_ENGINE_CONFIGURATION is empty")
		cxEngineConfiguration = "1"
	} else {
		fmt.Println("CX_ENGINE_CONFIGURATION : " + cxEngineConfiguration)
	}

	if len(cxHighThreshold) == 0 {
		cxHighThreshold = "9999999999999999"
		fmt.Println("\nCX_HIGH_THRESHOLD is empty")
	} else {
		fmt.Println("\nCX_HIGH_THRESHOLD : " + cxHighThreshold)
	}

	if len(cxMediumThreshold) == 0 {
		cxMediumThreshold = "9999999999999999"
		fmt.Println("CX_MEDIUM_THRESHOLD is empty")
	} else {
		fmt.Println("CX_MEDIUM_THRESHOLD : " + cxMediumThreshold)
	}

	if len(cxLowThreshold) == 0 {
		cxLowThreshold = "9999999999999999"
		fmt.Println("CX_LOW_THRESHOLD is empty")
	} else {
		fmt.Println("CX_LOW_THRESHOLD : " + cxLowThreshold)
	}
	fmt.Println("")

	cmx, err := checkmarx.NewCheckmarx(cxServerURL, cxUsername, cxPassword)
	if err != nil {
		errors.Errorf("Failed to create Checkmarx client talking to URL %v with error %v", cxServerURL, err)
	}

	teams := cmx.GetTeams()
	if len(teams) > 0 {
		team := cmx.GetTeamByName(teams, cxTeamName)
		if team.FullName == cxTeamName {
			projects := cmx.GetProjects()
			project := cmx.GetProjectByName(projects, cxProjectName)
			if project.Name == cxProjectName {
				fmt.Println("Project exists...")
			} else {
				fmt.Println("Project does not exists...")
				projectCreated := cmx.CreateProject(cxProjectName, team.ID)
				if projectCreated {
					projects := cmx.GetProjects()
					project := cmx.GetProjectByName(projects, cxProjectName)
					fmt.Println("New Project Created : " + project.Name)
				} else {
					fmt.Println("Cannot create Project : " + cxProjectName)
					os.Exit(10)
				}
			}

			zipFolder(sourceDir, sourceDir+".zip")
			sourceCodeUploaded := cmx.UploadProjectSourceCode(project.ID, sourceDir+".zip")
			if sourceCodeUploaded {
				fmt.Println("Source code Uploaded for Project : " + cxProjectName)
				err := os.Remove(sourceDir + ".zip")
				if err != nil {
					fmt.Printf("Error : %s\n", err)
				}
				excludeSettingsUpdated := cmx.UpdateProjectExcludeSettings(project.ID, cxExcludeFolders, cxExcludeFiles)
				if excludeSettingsUpdated {
					fmt.Println("Exclude settings Updated for Project : " + cxProjectName)
				} else {
					fmt.Println("Cannot update exclude settings for Project : " + cxProjectName)
				}
				presets := cmx.GetPresets()
				preset := cmx.GetPresetByName(presets, cxPresetName)
				if preset.Name == cxPresetName {
					configurationUpdated := cmx.UpdateProjectConfiguration(project.ID, preset.ID, cxEngineConfiguration)
					if configurationUpdated {
						fmt.Println("Configuration Updated for Project : " + cxProjectName)
						projectIsScanning, scan := cmx.ScanProject(project.ID)
						if projectIsScanning {
							fmt.Println("Scanning : " + cxProjectName)
							status := "New"
							pastStatus := status
							fmt.Println("Scan phase : " + status)
							for true {
								status = cmx.GetScanStatus(scan.ID)
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
								results := cmx.GetResults(scan.ID)
								insecure := false
								cxHighThreshold, _ := strconv.Atoi(cxHighThreshold)
								if results.High > cxHighThreshold {
									insecure = true
								}
								cxMediumThreshold, _ := strconv.Atoi(cxMediumThreshold)
								if results.Medium > cxMediumThreshold {
									insecure = true
								}
								cxLowThreshold, _ := strconv.Atoi(cxLowThreshold)
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
							fmt.Println("Cannot scan Project : " + cxProjectName)
							os.Exit(10)
						}
					} else {
						fmt.Println("Cannot update Configuration for Project : " + cxProjectName)
						os.Exit(10)
					}
				} else {
					fmt.Println("Preset does not exists : " + cxPresetName)
					os.Exit(10)
				}
			} else {
				fmt.Println("Cannot upload source code for Project : " + cxProjectName)
				os.Exit(10)
			}
		} else {
			fmt.Println("Team does not exists : " + cxTeamName)
			os.Exit(10)
		}
	} else {
		fmt.Println("No existing teams")
		os.Exit(10)
	}
}

func zipFolder(source string, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
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