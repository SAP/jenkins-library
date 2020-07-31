package cmd

import (
	"fmt"
	"os"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/maven"
	"strings"
	"path/filepath"

	sliceUtils "github.com/SAP/jenkins-library/pkg/piperutils"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/versioning"
)

type utilsBundleDetectMaven struct {
	*piperhttp.Client
	*piperutils.Files
}

func newUtils() *utilsBundleDetectMaven {
	return &utilsBundleDetectMaven{
		Client: &piperhttp.Client{},
		Files:  &piperutils.Files{},
	}
}

func detectExecuteScan(config detectExecuteScanOptions, telemetryData *telemetry.CustomData) {
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	fileUtils := piperutils.Files{}
	httpClient := piperhttp.Client{}

	err := runDetect(config, &c, &fileUtils, &httpClient)

	if err != nil {
		log.Entry().
			WithError(err).
			Fatal("failed to execute detect scan")
	}
}

func runDetect(config detectExecuteScanOptions, command command.ShellRunner, fileUtils piperutils.FileUtils, httpClient piperhttp.Downloader) error {
	// detect execution details, see https://synopsys.atlassian.net/wiki/spaces/INTDOCS/pages/88440888/Sample+Synopsys+Detect+Scan+Configuration+Scenarios+for+Black+Duck
	httpClient.DownloadFile("https://detect.synopsys.com/detect.sh", "detect.sh", nil, nil)
	err := fileUtils.Chmod("detect.sh", 0700)
	if err != nil {
		return err
	}
	args := []string{"./detect.sh"}
	args, err = addDetectArgsAndBuild(args, config, fileUtils, httpClient)
	if err != nil {
		return err
	}
	script := strings.Join(args, " ")

	envs := []string{"BLACKDUCK_SKIP_PHONE_HOME=true"}

	command.SetDir(".")
	command.SetEnv(envs)

	return command.RunShell("/bin/bash", script)
}

func addDetectArgsAndBuild(args []string, config detectExecuteScanOptions, fileUtils piperutils.FileUtils, httpClient piperhttp.Downloader) ([]string, error) {

	coordinates := struct {
		Version string
	}{
		Version: config.Version,
	}

	_, detectVersionName := versioning.DetermineProjectCoordinates("", config.VersioningModel, coordinates)

	args = append(args, config.ScanProperties...)

	args = append(args, fmt.Sprintf("--blackduck.url=%v", config.ServerURL))
	args = append(args, fmt.Sprintf("--blackduck.api.token=%v", config.APIToken))
	// ProjectNames, VersionName, GroupName etc can contain spaces and need to be escaped using double quotes in CLI
	// Hence the string need to be surrounded by \"
	args = append(args, fmt.Sprintf("--detect.project.name=\\\"%v\\\"", config.ProjectName))
	args = append(args, fmt.Sprintf("--detect.project.version.name=\\\"%v\\\"", detectVersionName))

	// Groups parameter is added only when there is atleast one non-empty groupname provided
	if len(config.Groups) > 0 && len(config.Groups[0]) > 0 {
		args = append(args, fmt.Sprintf("--detect.project.user.groups=\\\"%v\\\"", strings.Join(config.Groups, "\\\",\\\"")))
	}

	// Atleast 1, non-empty category to fail on must be provided
	if len(config.FailOn) > 0 && len(config.FailOn[0]) > 0 {
		args = append(args, fmt.Sprintf("--detect.policy.check.fail.on.severities=%v", strings.Join(config.FailOn, ",")))
	}

	codeLocation := config.CodeLocation
	if len(codeLocation) == 0 && len(config.ProjectName) > 0 {
		codeLocation = fmt.Sprintf("%v/%v", config.ProjectName, detectVersionName)
	}
	args = append(args, fmt.Sprintf("--detect.code.location.name=\\\"%v\\\"", codeLocation))

	if sliceUtils.ContainsString(config.Scanners, "signature") {
		args = append(args, fmt.Sprintf("--detect.blackduck.signature.scanner.paths=%v", strings.Join(config.ScanPaths, ",")))
	}

	if sliceUtils.ContainsString(config.Scanners, "source") {
		args = append(args, fmt.Sprintf("--detect.source.path=%v", config.ScanPaths[0]))
	}

	if (!config.BuildCode){
		mavenArgs, err := maven.DownloadAndGetMavenParameters(config.GlobalSettingsFile, config.ProjectSettingsFile, fileUtils, httpClient)

		if err != nil {
			return nil, err
		}

		if len(config.M2Path) > 0 {
			absolutePath, err := fileUtils.Abs(config.M2Path)
			if err != nil {
				return nil, err
			}
			mavenArgs = append(mavenArgs, fmt.Sprintf("-Dmaven.repo.local=%v", absolutePath))
		}

		if len(mavenArgs) > 0 {
			args = append(args, fmt.Sprintf("\"--detect.maven.build.command='%v'\"", strings.Join(mavenArgs, " ")))
		}
	} else {
		c1 := command.Command{} 
		switch config.BuildTool {
			case "maven" : 
				pomFiles, err := newUtils().Glob(filepath.Join("**", "pom.xml"))
				if err != nil {
					log.Entry().WithError(err).Warn("no pom xml found")
				}
			
				_, found := findElement(pomFiles, "pom.xml")
				if found {
				args = append(args, fmt.Sprintf("-detect.maven.build.command=\"clean install\""))
				} else {
					log.Entry().Info("no parent pom xml")
					localMavenBuild(fileUtils, config, &c1, args)
				}
			default : 
				pomFiles, err := newUtils().Glob(filepath.Join("**", "pom.xml"))
				if err != nil {
					log.Entry().WithError(err).Warn("no pom xml found")
				}
				localMavenBuild(fileUtils, config, &c1, pomFiles)
		}
	}
	return args, nil
}

func localMavenBuild(fileUtils piperutils.FileUtils, config detectExecuteScanOptions, command command.ExecRunner, pomFiles []string) {
	/* pomFiles, err := newUtils().Glob(filepath.Join("**", "pom.xml"))
	if err != nil {
		log.Entry().WithError(err).Warn("no pom xml found")
	}
    
    _, found := findElement(pomFiles, "pom.xml")
    if found {
        return append(args, fmt.Sprintf("-detect.maven.build.command=\"clean install\""))
    } else { */
		for _, pomFile := range pomFiles {
			if strings.Count(pomFile, string(os.PathSeparator)) == 1 {
				executeCleanOptions := maven.ExecuteOptions{
					PomPath:             pomFile,
					ProjectSettingsFile: config.ProjectSettingsFile,
					GlobalSettingsFile:  config.GlobalSettingsFile,
					M2Path:              config.M2Path,
					Goals:               []string{"clean"},
					Defines:             []string{"-DskipTests=true"},
					ReturnStdout: 		 true,
				}
				_, errClean := maven.Execute(&executeCleanOptions, command)
				if errClean != nil {
					log.Entry().WithError(errClean).Warn("failed to clean : ", pomFile)
				}
				executeInstallOptions := maven.ExecuteOptions{
					PomPath:             pomFile,
					ProjectSettingsFile: config.ProjectSettingsFile,
					GlobalSettingsFile:  config.GlobalSettingsFile,
					M2Path:              config.M2Path,
					Goals:               []string{"install"},
					Defines:             []string{"-DskipTests=true"},
					ReturnStdout: 		 true,
				}
				_, errInstall := maven.Execute(&executeInstallOptions, command)
				if errInstall != nil {
					log.Entry().WithError(errInstall).Warn("failed to clean : ", pomFile)
				}				
			}
		}
	//}
	// if pom.xml : parent POM : maven intall  args = append(args, fmt.Sprintf("\"--detect.maven.build.command='%v'\"", "clean install"))

	// 
	// A/pom.xml : MTA : maven install



	/* if config.M2Path != "" {
		config.M2Path, err = filepath.Abs(config.M2Path)
		if err != nil {
			log.Entry().WithError(err).Warn("absolute file path error for pom")
		}
	}

	for _, pomFile := range pomFiles {
		executeCleanOptions := maven.ExecuteOptions{
			PomPath:             pomFile,
			ProjectSettingsFile: config.ProjectSettingsFile,
			GlobalSettingsFile:  config.GlobalSettingsFile,
			M2Path:              config.M2Path,
			Goals:               []string{"clean"},
			Defines:             []string{"-DskipTests=true"},
			ReturnStdout: 		 true,
		}
		_, errClean := maven.Execute(&executeCleanOptions, command)
		if errClean != nil {
			log.Entry().WithError(err).Warn("failed to clean : ", pomFile)
		}
		executeInstallOptions := maven.ExecuteOptions{
			PomPath:             pomFile,
			ProjectSettingsFile: config.ProjectSettingsFile,
			GlobalSettingsFile:  config.GlobalSettingsFile,
			M2Path:              config.M2Path,
			Goals:               []string{"install"},
			Defines:             []string{"-DskipTests=true"},
			ReturnStdout: 		 true,
		}
		_, errInstall := maven.Execute(&executeInstallOptions, command)
		if errInstall != nil {
			log.Entry().WithError(err).Warn("failed to clean : ", pomFile)
		}
	} */
}

func findElement(slice []string, val string) (int, bool) {
		for i, item := range slice {
			if item == val {
				return i, true
			}
		}
		return -1, false
}
