package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/fortify"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type translate interface {
	AppendToOptions([]string) []string
}

type aspTranslate struct {
	Src               string `json:"src"`
	Aspnetcore        string `json:"aspnetcore"`
	DotNetCoreVersion string `json:"dotNetCoreVersion"`
	Exclude           string `json:"exclude"`
	LibDirs           string `json:"libDirs"`
}

func (t *aspTranslate) AppendToOptions(options []string) []string {
	if len(t.Aspnetcore) > 0 {
		options = append(options, "-aspnetcore")
	}
	if len(t.DotNetCoreVersion) > 0 {
		options = append(options, "-dotnet-core-version", t.DotNetCoreVersion)
	}
	if len(t.Exclude) > 0 {
		options = append(options, "-exclude", t.Exclude)
	}
	if len(t.LibDirs) > 0 {
		options = append(options, "-libdirs", t.LibDirs)
	}
	return append(options, t.Src)
}

type javaTranslate struct {
	Classpath    string `json:"classpath"`
	Extdirs      string `json:"extdirs"`
	JavaBuildDir string `json:"javaBuildDir"`
	Source       string `json:"source"`
	Jdk          string `json:"jdk"`
	Sourcepath   string `json:"sourcepath"`
}

func (t *javaTranslate) AppendToOptions(options []string) []string {
	if len(t.Classpath) > 0 {
		options = append(options, "-cp", t.Classpath)
	}
	if len(t.Extdirs) > 0 {
		options = append(options, "-extdirs", t.Extdirs)
	}
	if len(t.JavaBuildDir) > 0 {
		options = append(options, "-java-build-dir", t.JavaBuildDir)
	}
	if len(t.Source) > 0 {
		options = append(options, "-source", t.Source)
	}
	if len(t.Jdk) > 0 {
		options = append(options, "-jdk", t.Jdk)
	}
	if len(t.Sourcepath) > 0 {
		options = append(options, "-sourcepath", t.Sourcepath)
	}
	return options
}

type pythonTranslate struct {
	PythonPath     string `json:"pythonPath"`
	PythonIncludes string `json:"pythonIncludes"`
	PythonExcludes string `json:"pythonExcludes"`
}

func (t *pythonTranslate) AppendToOptions(options []string) []string {
	if len(t.PythonPath) > 0 {
		options = append(options, "-python-path", t.PythonPath)
	}
	if len(t.PythonExcludes) > 0 {
		options = append(options, "-python-excludes", t.PythonExcludes)
	}
	if len(t.PythonIncludes) > 0 {
		options = append(options, "-python-includes", t.PythonIncludes)
	}
	return options
}

func fortifyExecuteScan(config fortifyExecuteScanOptions, telemetryData *telemetry.CustomData, influx *fortifyExecuteScanInflux) error {
	sys := fortify.NewSystemInstance(config.ServerURL, config.APIEndpoint, config.AuthToken, time.Second*30)
	c := command.Command{}
	// reroute command output to loging framework
	// also log stdout as Karma reports into it
	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())
	return runFortifyScan(config, sys, &c, telemetryData, influx)
}

func runFortifyScan(config fortifyExecuteScanOptions, sys fortify.System, command execRunner, telemetryData *telemetry.CustomData, influx *fortifyExecuteScanInflux) error {
	log.Entry().Debugf("Running Fortify scan against SSC at %v", config.ServerURL)
	gav, err := piperutils.GetMavenCoordinates(config.BuildDescriptorFile)
	if err != nil {
		log.Entry().Warnf("Unable to load project coordinates from descriptor %v: %v", config.BuildDescriptorFile, err)
	}
	fortifyProjectName, fortifyProjectVersion := determineProjectCoordinates(config, gav)
	project, err := sys.GetProjectByName(fortifyProjectName)
	if err != nil {
		log.Entry().Fatalf("Failed to load project %v: %v", fortifyProjectName, err)
	}
	projectVersion, err := sys.GetProjectVersionDetailsByProjectIDAndVersionName(project.ID, fortifyProjectVersion)
	if err != nil {
		log.Entry().Fatalf("Failed to load project version %v: %v", fortifyProjectVersion, err)
	}
	if len(config.PullRequestName) > 0 {
		fortifyProjectVersion = config.PullRequestName
		projectVersion, err := sys.LookupOrCreateProjectVersionDetailsForPullRequest(project.ID, projectVersion, fortifyProjectVersion)
		if err != nil {
			log.Entry().Fatalf("Failed to lookup / create project version for pull request %v: %v", fortifyProjectVersion, err)
		}
		log.Entry().Debugf("Looked up / created project version with ID %v for PR %v", projectVersion.ID, fortifyProjectVersion)
	} else {
		prID := determinePullRequestMerge(config)
		if len(prID) > 0 {
			log.Entry().Debugf("Determined PR identifier %v for merge check", prID)
			err = sys.MergeProjectVersionStateOfPRIntoMaster(config.FprDownloadEndpoint, config.FprUploadEndpoint, project.ID, projectVersion.ID, fmt.Sprintf("PR-%v", prID))
			if err != nil {
				log.Entry().Fatalf("Failed to merge project version state for pull request %v: %v", fortifyProjectVersion, err)
			}
		}
	}

	log.Entry().Debugf("Scanning and uploading to project %v with version %v and projectVersionId %v", fortifyProjectName, fortifyProjectVersion, projectVersion.ID)
	repoURL := strings.ReplaceAll(config.RepoURL, ".git", "")
	buildLabel := fmt.Sprintf("%v/commit/%v", repoURL, config.CommitID)

	// Create sourceanalyzer / maven command based on configuration
	if config.ScanType != "maven" {
		// Create and execute special maven command

	} else {
		buildID := uuid.New().String()
		command.Dir(config.ModulePath)
		os.MkdirAll(fmt.Sprintf("%v/%v", config.ModulePath, "target"), os.ModePerm)

		if config.UpdateRulePack {
			err := command.RunExecutable("fortifyupdate", "-acceptKey", "-acceptSSLCertificate", "-url", config.ServerURL)
			if err != nil {
				log.Entry().WithError(err).WithField("serverUrl", config.ServerURL).Fatal("Failed to update rule pack")
			}
			err = command.RunExecutable("fortifyupdate", "-acceptKey", "-acceptSSLCertificate", "-showInstalledRules")
			if err != nil {
				log.Entry().WithError(err).WithField("serverUrl", config.ServerURL).Fatal("Failed to fetch details of installed rule pack")
			}
		}

		triggerFortifyScan(config, command, buildID, buildLabel)
	}

	var reports []piperutils.Path
	reports = append(reports, piperutils.Path{Target: fmt.Sprintf("%vtarget/fortify-scan.*", config.ModulePath)})
	reports = append(reports, piperutils.Path{Target: fmt.Sprintf("%vtarget/*.fpr", config.ModulePath)})

	var message string
	if config.UploadResults {
		log.Entry().Debug("Uploading results")
		resultFilePath := fmt.Sprintf("%vtarget/result.fpr", config.ModulePath)
		err = sys.UploadResultFile(config.FprUploadEndpoint, resultFilePath, projectVersion.ID)
		message = fmt.Sprintf("Failed to upload result file %v to Fortify SSC at %v", resultFilePath, config.ServerURL)
	} else {
		log.Entry().Debug("Generating XML report")
		xmlReportName := "fortify_result.xml"
		err = command.RunExecutable("ReportGenerator", "-format", "xml", "-f", xmlReportName, "-source", fmt.Sprintf("%vtarget/result.fpr", config.ModulePath))
		message = fmt.Sprintf("Failed to generate XML report %v", xmlReportName)
		if err != nil {
			reports = append(reports, piperutils.Path{Target: fmt.Sprintf("%vfortify_result.xml", config.ModulePath)})
		}
	}
	piperutils.PersistReportsAndLinks("fortifyExecuteScan", config.ModulePath, reports, nil)
	if err != nil {
		log.Entry().Fatal(message)
	}

	// Fetch report

	// Perform audit compliance checks

	return nil
}

func triggerFortifyScan(config fortifyExecuteScanOptions, command execRunner, buildID, buildLabel string) {
	if config.ScanType == "pip" {
		// Do special Python related prep
		pipVersion := "pip3"
		if config.PythonVersion != "python3" {
			pipVersion = "pip2"
		}
		installCommand, err := piperutils.ExecuteTemplate(config.PythonInstallCommand, map[string]string{"pip": pipVersion})
		if err != nil {
			log.Entry().WithError(err).Fatalf("Failed to execute template for PythonInstallCommand: %v", config.PythonInstallCommand)
		}
		installCommandTokens := tokenize(installCommand)
		err = command.RunExecutable(installCommandTokens[0], installCommandTokens[1:]...)
		if err != nil {
			log.Entry().WithError(err).WithField("command", config.PythonInstallCommand).Fatal("Failed to execute python install command")
		}

		if len(config.Translate) == 0 {
			buf := new(bytes.Buffer)
			command.Stdout(buf)
			err := command.RunExecutable(config.PythonVersion, "-c", "import sys;p=sys.path;p.remove('');print(';'.join(p))")
			command.Stdout(log.Entry().Writer())

			config.Translate = `[{"pythonPath": `
			if err == nil {
				config.Translate += strings.TrimSpace(buf.String())
				config.Translate += `;`
			}
			config.Translate += config.PythonAdditionalPath
			config.Translate += `, "pythonIncludes": `
			config.Translate += config.PythonIncludes
			config.Translate += `, "pythonExcludes": `
			config.Translate += config.PythonExcludes
			config.Translate += `}]`
		}
	}

	translateProject(config, command, buildID)

	scanProject(config, command, buildID, buildLabel)
}

func translateProject(config fortifyExecuteScanOptions, command execRunner, buildID string) {
	var translateList []translate
	json.Unmarshal([]byte(config.Translate), &translateList)
	for _, translate := range translateList {
		translateOptions := []string{
			"-verbose",
			"-64",
			config.Memory,
			fmt.Sprintf("-b %v", buildID),
		}
		translateOptions = translate.AppendToOptions(translateOptions)
		err := command.RunExecutable("sourceanalyzer", translateOptions...)
		if err != nil {
			log.Entry().WithError(err).WithField("translateOptions", translateOptions).Fatal("failed to execute sourceanalyzer translate command")
		}
	}
}

func scanProject(config fortifyExecuteScanOptions, command execRunner, buildID, buildLabel string) {
	var scanOptions = []string{
		"-show-build-warnings",
		"-verbose",
		"-64",
		config.Memory,
		fmt.Sprintf("-b %v", buildID),
		"-scan",
	}
	if config.QuickScan {
		scanOptions = append(scanOptions, "-quick")
	}
	if len(buildLabel) > 0 {
		scanOptions = append(scanOptions, fmt.Sprintf("-build-label %v", buildLabel))
	}
	scanOptions = append(scanOptions, "-logfile target/fortify-scan.log")
	scanOptions = append(scanOptions, "-f target/result.fpr")

	err := command.RunExecutable("sourceanalyzer", scanOptions...)
	if err != nil {
		log.Entry().WithError(err).WithField("scanOptions", scanOptions).Fatal("failed to execute sourceanalyzer scan command")
	}
}

func determinePullRequestMerge(config fortifyExecuteScanOptions) string {
	log.Entry().Debugf("Retrieved commit message %v", config.CommitMessage)
	r, _ := regexp.Compile(config.PullRequestMessageRegex)
	matches := r.FindSubmatch([]byte(config.CommitMessage))
	if matches != nil && len(matches) > 1 {
		return string(matches[config.PullRequestMessageRegexGroup])
	}
	return ""
}

func determineProjectCoordinates(config fortifyExecuteScanOptions, gav *piperutils.MavenDescriptor) (string, string) {
	projectName, err := piperutils.ExecuteTemplate(config.ProjectName, *gav)
	if err != nil {
		log.Entry().Warnf("Unable to resolve fortify project name %v", err)
	}
	projectVersion, err := piperutils.ExecuteTemplate(config.ProjectVersion, *gav)
	if err != nil {
		log.Entry().Warnf("Unable to resolve fortify project version %v", err)
	}
	return projectName, projectVersion
}
