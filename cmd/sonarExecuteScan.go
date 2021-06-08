package cmd

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar"
	"github.com/pkg/errors"

	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	FileUtils "github.com/SAP/jenkins-library/pkg/piperutils"
	SliceUtils "github.com/SAP/jenkins-library/pkg/piperutils"
	StepResults "github.com/SAP/jenkins-library/pkg/piperutils"
	SonarUtils "github.com/SAP/jenkins-library/pkg/sonar"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/versioning"
)

type sonarSettings struct {
	workingDir  string
	binary      string
	environment []string
	options     []string
}

func (s *sonarSettings) addEnvironment(element string) {
	s.environment = append(s.environment, element)
}

func (s *sonarSettings) addOption(element string) {
	s.options = append(s.options, element)
}

var (
	sonar sonarSettings

	execLookPath    = exec.LookPath
	fileUtilsExists = FileUtils.FileExists
	fileUtilsUnzip  = FileUtils.Unzip
	osRename        = os.Rename
	osStat          = os.Stat
	doublestarGlob  = doublestar.Glob
)

const (
	coverageReportPaths = "sonar.coverage.jacoco.xmlReportPaths="
	javaBinaries        = "sonar.java.binaries="
	javaLibraries       = "sonar.java.libraries="
	coverageExclusions  = "sonar.coverage.exclusions="
	pomXMLPattern       = "**/pom.xml"
)

func sonarExecuteScan(config sonarExecuteScanOptions, _ *telemetry.CustomData, influx *sonarExecuteScanInflux) {
	runner := command.Command{
		ErrorCategoryMapping: map[string][]string{
			log.ErrorConfiguration.String(): {
				"You must define the following mandatory properties for '*': *",
				"org.sonar.java.AnalysisException: Your project contains .java files, please provide compiled classes with sonar.java.binaries property, or exclude them from the analysis with sonar.exclusions property.",
				"ERROR: Invalid value for *",
				"java.lang.IllegalStateException: No files nor directories matching '*'",
			},
			log.ErrorInfrastructure.String(): {
				"ERROR: SonarQube server [*] can not be reached",
				"Caused by: java.net.SocketTimeoutException: timeout",
				"java.lang.IllegalStateException: Fail to request *",
				"java.lang.IllegalStateException: Fail to download plugin [*] into *",
			},
		},
	}
	// reroute command output to logging framework
	runner.Stdout(log.Writer())
	runner.Stderr(log.Writer())
	// client for downloading the sonar-scanner
	downloadClient := &piperhttp.Client{}
	downloadClient.SetOptions(piperhttp.ClientOptions{TransportTimeout: 20 * time.Second})
	// client for talking to the SonarQube API
	apiClient := &piperhttp.Client{}
	//TODO: implement certificate handling
	apiClient.SetOptions(piperhttp.ClientOptions{TransportSkipVerification: true})

	sonar = sonarSettings{
		workingDir:  "./",
		binary:      "sonar-scanner",
		environment: []string{},
		options:     []string{},
	}

	influx.step_data.fields.sonar = false
	if err := runSonar(config, downloadClient, &runner, apiClient, influx); err != nil {
		if log.GetErrorCategory() == log.ErrorUndefined && runner.GetExitCode() == 2 {
			// see https://github.com/SonarSource/sonar-scanner-cli/blob/adb67d645c3bcb9b46f29dea06ba082ebec9ba7a/src/main/java/org/sonarsource/scanner/cli/Exit.java#L25
			log.SetErrorCategory(log.ErrorConfiguration)
		}
		log.Entry().WithError(err).Fatal("Execution failed")
	}
	influx.step_data.fields.sonar = true
}

func runSonar(config sonarExecuteScanOptions, client piperhttp.Downloader, runner command.ExecRunner, apiClient SonarUtils.Sender, influx *sonarExecuteScanInflux) error {
	// Set config based on orchestrator-specific environment variables
	detectParametersFromCI(&config)

	if len(config.ServerURL) > 0 {
		sonar.addEnvironment("SONAR_HOST_URL=" + config.ServerURL)
	}
	if len(config.Token) == 0 {
		log.Entry().Warn("sonar token not set")
		// use token provided by sonar-scanner-jenkins plugin
		// https://github.com/SonarSource/sonar-scanner-jenkins/blob/441ef2f485884758b60767bed2ef8a1a0a7fc863/src/main/java/hudson/plugins/sonar/SonarBuildWrapper.java#L132
		if len(os.Getenv("SONAR_AUTH_TOKEN")) > 0 {
			log.Entry().Info("using token from env var SONAR_AUTH_TOKEN")
			config.Token = os.Getenv("SONAR_AUTH_TOKEN")
		}
	}
	if len(config.Token) > 0 {
		sonar.addEnvironment("SONAR_TOKEN=" + config.Token)
	}
	if len(config.Organization) > 0 {
		sonar.addOption("sonar.organization=" + config.Organization)
	}
	if len(config.Version) > 0 {
		version := config.CustomScanVersion
		if len(version) > 0 {
			log.Entry().Infof("Using custom version: %v", version)
		} else {
			version = versioning.ApplyVersioningModel(config.VersioningModel, config.Version)
		}
		sonar.addOption("sonar.projectVersion=" + version)
	}
	if len(config.ProjectKey) > 0 {
		sonar.addOption("sonar.projectKey=" + config.ProjectKey)
	}
	if len(config.M2Path) > 0 && config.InferJavaLibraries {
		sonar.addOption(javaLibraries + filepath.Join(config.M2Path, "**"))
	}
	if len(config.CoverageExclusions) > 0 && !isInOptions(config, coverageExclusions) {
		sonar.addOption(coverageExclusions + strings.Join(config.CoverageExclusions, ","))
	}
	if config.InferJavaBinaries && !isInOptions(config, javaBinaries) {
		addJavaBinaries()
	}
	if err := handlePullRequest(config); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return err
	}
	if err := loadSonarScanner(config.SonarScannerDownloadURL, client); err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		return err
	}
	if err := loadCertificates(config.CustomTLSCertificateLinks, client, runner); err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		return err
	}

	if len(config.Options) > 0 {
		sonar.options = append(sonar.options, config.Options...)
	}

	sonar.options = SliceUtils.PrefixIfNeeded(SliceUtils.Trim(sonar.options), "-D")

	log.Entry().
		WithField("command", sonar.binary).
		WithField("options", sonar.options).
		WithField("environment", sonar.environment).
		Debug("Executing sonar scan command")
	// execute scan
	runner.SetEnv(sonar.environment)
	err := runner.RunExecutable(sonar.binary, sonar.options...)
	if err != nil {
		return err
	}

	// as PRs are handled locally for legacy SonarQube systems, no measurements will be fetched.
	if len(config.ChangeID) > 0 && config.LegacyPRHandling {
		return nil
	}

	// load task results
	taskReport, err := SonarUtils.ReadTaskReport(sonar.workingDir)
	if err != nil {
		log.Entry().WithError(err).Warning("no scan report found")
		return nil
	}
	// write links JSON
	links := []StepResults.Path{
		{
			Target: taskReport.DashboardURL,
			Name:   "Sonar Web UI",
		},
	}
	StepResults.PersistReportsAndLinks("sonarExecuteScan", sonar.workingDir, nil, links)

	if len(config.Token) == 0 {
		log.Entry().Warn("no measurements are fetched due to missing credentials")
		return nil
	}
	taskService := SonarUtils.NewTaskService(taskReport.ServerURL, config.Token, taskReport.TaskID, apiClient)
	// wait for analysis task to complete
	err = taskService.WaitForTask()
	if err != nil {
		return err
	}
	// fetch number of issues by severity
	issueService := SonarUtils.NewIssuesService(taskReport.ServerURL, config.Token, taskReport.ProjectKey, config.Organization, config.BranchName, config.ChangeID, apiClient)
	influx.sonarqube_data.fields.blocker_issues, err = issueService.GetNumberOfBlockerIssues()
	if err != nil {
		return err
	}
	influx.sonarqube_data.fields.critical_issues, err = issueService.GetNumberOfCriticalIssues()
	if err != nil {
		return err
	}
	influx.sonarqube_data.fields.major_issues, err = issueService.GetNumberOfMajorIssues()
	if err != nil {
		return err
	}
	influx.sonarqube_data.fields.minor_issues, err = issueService.GetNumberOfMinorIssues()
	if err != nil {
		return err
	}
	influx.sonarqube_data.fields.info_issues, err = issueService.GetNumberOfInfoIssues()
	if err != nil {
		return err
	}
	log.Entry().Debugf("Influx values: %v", influx.sonarqube_data.fields)
	err = SonarUtils.WriteReport(SonarUtils.ReportData{
		ServerURL:    taskReport.ServerURL,
		ProjectKey:   taskReport.ProjectKey,
		TaskID:       taskReport.TaskID,
		ChangeID:     config.ChangeID,
		BranchName:   config.BranchName,
		Organization: config.Organization,
		NumberOfIssues: SonarUtils.Issues{
			Blocker:  influx.sonarqube_data.fields.blocker_issues,
			Critical: influx.sonarqube_data.fields.critical_issues,
			Major:    influx.sonarqube_data.fields.major_issues,
			Minor:    influx.sonarqube_data.fields.minor_issues,
			Info:     influx.sonarqube_data.fields.info_issues,
		},
	}, sonar.workingDir, ioutil.WriteFile)
	if err != nil {
		return err
	}
	return nil
}

// isInOptions returns true, if the given property is already provided in config.Options.
func isInOptions(config sonarExecuteScanOptions, property string) bool {
	property = strings.TrimSuffix(property, "=")
	return SliceUtils.ContainsStringPart(config.Options, property)
}

func addJavaBinaries() {
	pomFiles, err := doublestarGlob(pomXMLPattern)
	if err != nil {
		log.Entry().Warnf("failed to glob for pom modules: %v", err)
		return
	}
	var binaries []string

	var classesDirs = []string{"classes", "test-classes"}

	for _, pomFile := range pomFiles {
		module := filepath.Dir(pomFile)
		for _, classDir := range classesDirs {
			classesPath := filepath.Join(module, "target", classDir)
			_, err := osStat(classesPath)
			if err == nil {
				binaries = append(binaries, classesPath)
			}
		}
	}
	if len(binaries) > 0 {
		sonar.addOption(javaBinaries + strings.Join(binaries, ","))
	}
}

func handlePullRequest(config sonarExecuteScanOptions) error {
	if len(config.ChangeID) > 0 {
		if config.LegacyPRHandling {
			// see https://docs.sonarqube.org/display/PLUG/GitHub+Plugin
			sonar.addOption("sonar.analysis.mode=preview")
			sonar.addOption("sonar.github.pullRequest=" + config.ChangeID)
			if len(config.GithubAPIURL) > 0 {
				sonar.addOption("sonar.github.endpoint=" + config.GithubAPIURL)
			}
			if len(config.GithubToken) > 0 {
				sonar.addOption("sonar.github.oauth=" + config.GithubToken)
			}
			if len(config.Owner) > 0 && len(config.Repository) > 0 {
				sonar.addOption("sonar.github.repository=" + config.Owner + "/" + config.Repository)
			}
			if config.DisableInlineComments {
				sonar.addOption("sonar.github.disableInlineComments=" + strconv.FormatBool(config.DisableInlineComments))
			}
		} else {
			// see https://sonarcloud.io/documentation/analysis/pull-request/
			provider := strings.ToLower(config.PullRequestProvider)
			if provider == "github" {
				if len(config.Owner) > 0 && len(config.Repository) > 0 {
					sonar.addOption("sonar.pullrequest.github.repository=" + config.Owner + "/" + config.Repository)
				}
			} else {
				return errors.New("Pull-Request provider '" + provider + "' is not supported!")
			}
			sonar.addOption("sonar.pullrequest.key=" + config.ChangeID)
			sonar.addOption("sonar.pullrequest.base=" + config.ChangeTarget)
			sonar.addOption("sonar.pullrequest.branch=" + config.ChangeBranch)
			sonar.addOption("sonar.pullrequest.provider=" + provider)
		}
	} else if len(config.BranchName) > 0 {
		sonar.addOption("sonar.branch.name=" + config.BranchName)
	}
	return nil
}

func loadSonarScanner(url string, client piperhttp.Downloader) error {
	if scannerPath, err := execLookPath(sonar.binary); err == nil {
		// using existing sonar-scanner
		log.Entry().WithField("path", scannerPath).Debug("Using local sonar-scanner")
	} else if len(url) != 0 {
		// download sonar-scanner-cli into TEMP folder
		log.Entry().WithField("url", url).Debug("Downloading sonar-scanner")
		tmpFolder := getTempDir()
		defer os.RemoveAll(tmpFolder) // clean up
		archive := filepath.Join(tmpFolder, path.Base(url))
		if err := client.DownloadFile(url, archive, nil, nil); err != nil {
			return errors.Wrap(err, "Download of sonar-scanner failed")
		}
		// unzip sonar-scanner-cli
		log.Entry().WithField("source", archive).WithField("target", tmpFolder).Debug("Extracting sonar-scanner")
		if _, err := fileUtilsUnzip(archive, tmpFolder); err != nil {
			return errors.Wrap(err, "Extraction of sonar-scanner failed")
		}
		// move sonar-scanner-cli to .sonar-scanner/
		toolPath := ".sonar-scanner"
		foldername := strings.ReplaceAll(strings.ReplaceAll(archive, ".zip", ""), "cli-", "")
		log.Entry().WithField("source", foldername).WithField("target", toolPath).Debug("Moving sonar-scanner")
		if err := osRename(foldername, toolPath); err != nil {
			return errors.Wrap(err, "Moving of sonar-scanner failed")
		}
		// update binary path
		sonar.binary = filepath.Join(getWorkingDir(), toolPath, "bin", sonar.binary)
		log.Entry().Debug("Download completed")
	}
	return nil
}

func loadCertificates(certificateList []string, client piperhttp.Downloader, runner command.ExecRunner) error {
	trustStoreFile := filepath.Join(getWorkingDir(), ".certificates", "cacerts")

	if exists, _ := fileUtilsExists(trustStoreFile); exists {
		// use local existing trust store
		sonar.addEnvironment("SONAR_SCANNER_OPTS=-Djavax.net.ssl.trustStore=" + trustStoreFile + " -Djavax.net.ssl.trustStorePassword=changeit")
		log.Entry().WithField("trust store", trustStoreFile).Info("Using local trust store")
	} else
	//TODO: certificate loading is deactivated due to the missing JAVA keytool
	// see https://github.com/SAP/jenkins-library/issues/1072
	if os.Getenv("PIPER_SONAR_LOAD_CERTIFICATES") == "true" && len(certificateList) > 0 {
		// use local created trust store with downloaded certificates
		keytoolOptions := []string{
			"-import",
			"-noprompt",
			"-storepass", "changeit",
			"-keystore", trustStoreFile,
		}
		tmpFolder := getTempDir()
		defer os.RemoveAll(tmpFolder) // clean up

		for _, certificate := range certificateList {
			filename := path.Base(certificate) // decode?
			target := filepath.Join(tmpFolder, filename)

			log.Entry().WithField("source", certificate).WithField("target", target).Info("Downloading TLS certificate")
			// download certificate
			if err := client.DownloadFile(certificate, target, nil, nil); err != nil {
				return errors.Wrapf(err, "Download of TLS certificate failed")
			}
			options := append(keytoolOptions, "-file", target)
			options = append(options, "-alias", filename)
			// add certificate to keystore
			if err := runner.RunExecutable("keytool", options...); err != nil {
				return errors.Wrap(err, "Adding certificate to keystore failed")
			}
		}
		sonar.addEnvironment("SONAR_SCANNER_OPTS=-Djavax.net.ssl.trustStore=" + trustStoreFile + " -Djavax.net.ssl.trustStorePassword=changeit")
		log.Entry().WithField("trust store", trustStoreFile).Info("Using local trust store")
	} else {
		log.Entry().Debug("Download of TLS certificates skipped")
	}
	return nil
}

func getWorkingDir() string {
	workingDir, err := os.Getwd()
	if err != nil {
		log.Entry().WithError(err).WithField("path", workingDir).Debug("Retrieving of work directory failed")
	}
	return workingDir
}

func getTempDir() string {
	tmpFolder, err := ioutil.TempDir(".", "temp-")
	if err != nil {
		log.Entry().WithError(err).WithField("path", tmpFolder).Debug("Creating temp directory failed")
	}
	return tmpFolder
}

// Fetches parameters from environment variables and updates the options accordingly (only if not already set)
func detectParametersFromCI(options *sonarExecuteScanOptions) {
	provider, err := orchestrator.NewOrchestratorSpecificConfigProvider()
	if err != nil {
		log.Entry().WithError(err).Warning("Cannot infer config from CI environment")
		return
	}

	if provider.IsPullRequest() {
		config := provider.GetPullRequestConfig()
		if len(options.ChangeBranch) == 0 {
			log.Entry().Info("Infering parameter changeBranch from environment: " + config.Branch)
			options.ChangeBranch = config.Branch
		}
		if len(options.ChangeTarget) == 0 {
			log.Entry().Info("Infering parameter changeTarget from environment: " + config.Base)
			options.ChangeTarget = config.Base
		}
		if len(options.ChangeID) == 0 {
			log.Entry().Info("Infering parameter changeId from environment: " + config.Key)
			options.ChangeID = config.Key
		}
	} else {
		config := provider.GetBranchBuildConfig()

		if options.InferBranchName && len(options.BranchName) == 0 {
			log.Entry().Info("Infering parameter branchName from environment: " + config.Branch)
			options.BranchName = config.Branch
		}
	}
}
