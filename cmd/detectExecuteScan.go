package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/pkg/errors"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/toolrecord"
	"github.com/SAP/jenkins-library/pkg/versioning"
)

type detectUtils interface {
	Abs(path string) (string, error)
	FileExists(filename string) (bool, error)
	FileRemove(filename string) error
	Copy(src, dest string) (int64, error)
	FileRead(path string) ([]byte, error)
	FileWrite(path string, content []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	Chmod(path string, mode os.FileMode) error
	Glob(pattern string) (matches []string, err error)

	GetExitCode() int
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	SetDir(dir string)
	SetEnv(env []string)
	RunExecutable(e string, p ...string) error
	RunShell(shell, script string) error

	DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error
}

type detectUtilsBundle struct {
	*command.Command
	*piperutils.Files
	*piperhttp.Client
}

func newDetectUtils() detectUtils {
	utils := detectUtilsBundle{
		Command: &command.Command{
			ErrorCategoryMapping: map[string][]string{
				log.ErrorCompliance.String(): {
					"FAILURE_POLICY_VIOLATION - Detect found policy violations.",
				},
				log.ErrorConfiguration.String(): {
					"FAILURE_CONFIGURATION - Detect was unable to start due to issues with it's configuration.",
					"FAILURE_DETECTOR - Detect had one or more detector failures while extracting dependencies. Check that all projects build and your environment is configured correctly.",
					"FAILURE_SCAN - Detect was unable to run the signature scanner against your source. Check your configuration.",
				},
				log.ErrorInfrastructure.String(): {
					"FAILURE_PROXY_CONNECTIVITY - Detect was unable to use the configured proxy. Check your configuration and connection.",
					"FAILURE_BLACKDUCK_CONNECTIVITY - Detect was unable to connect to Black Duck. Check your configuration and connection.",
					"FAILURE_POLARIS_CONNECTIVITY - Detect was unable to connect to Polaris. Check your configuration and connection.",
				},
				log.ErrorService.String(): {
					"FAILURE_TIMEOUT - Detect could not wait for actions to be completed on Black Duck. Check your Black Duck server or increase your timeout.",
					"FAILURE_DETECTOR_REQUIRED - Detect did not run all of the required detectors. Fix detector issues or disable required detectors.",
					"FAILURE_BLACKDUCK_VERSION_NOT_SUPPORTED - Detect attempted an operation that was not supported by your version of Black Duck. Ensure your Black Duck is compatible with this version of detect.",
					"FAILURE_BLACKDUCK_FEATURE_ERROR - Detect encountered an error while attempting an operation on Black Duck. Ensure your Black Duck is compatible with this version of detect.",
					"FAILURE_GENERAL_ERROR - Detect encountered a known error, details of the error are provided.",
					"FAILURE_UNKNOWN_ERROR - Detect encountered an unknown error.",
				},
			},
		},
		Files:  &piperutils.Files{},
		Client: &piperhttp.Client{},
	}
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func detectExecuteScan(config detectExecuteScanOptions, _ *telemetry.CustomData) {
	utils := newDetectUtils()
	err := runDetect(config, utils)

	if err != nil {
		log.Entry().
			WithError(err).
			Fatal("failed to execute detect scan")
	}

	// create Toolrecord file
	toolRecordFileName, err := createToolRecordDetect("./", config)
	if err != nil {
		// do not fail until the framework is well established
		log.Entry().Warning("TR_DETECT: Failed to create toolrecord file "+toolRecordFileName, err)
	}
}

func runDetect(config detectExecuteScanOptions, utils detectUtils) error {
	// detect execution details, see https://synopsys.atlassian.net/wiki/spaces/INTDOCS/pages/88440888/Sample+Synopsys+Detect+Scan+Configuration+Scenarios+for+Black+Duck
	err := getDetectScript(config, utils)
	if err != nil {
		return fmt.Errorf("failed to download 'detect.sh' script: %w", err)
	}
	defer func() {
		err := utils.FileRemove("detect.sh")
		if err != nil {
			log.Entry().Warnf("failed to delete 'detect.sh' script: %v", err)
		}
	}()
	err = utils.Chmod("detect.sh", 0700)
	if err != nil {
		return err
	}

	if config.InstallArtifacts {
		err := maven.InstallMavenArtifacts(&maven.EvaluateOptions{
			M2Path:              config.M2Path,
			ProjectSettingsFile: config.ProjectSettingsFile,
			GlobalSettingsFile:  config.GlobalSettingsFile,
		}, utils)
		if err != nil {
			return err
		}
	}

	args := []string{"./detect.sh"}
	args, err = addDetectArgs(args, config, utils)
	if err != nil {
		return err
	}
	script := strings.Join(args, " ")

	envs := []string{"BLACKDUCK_SKIP_PHONE_HOME=true"}
	envs = append(envs, config.CustomEnvironmentVariables...)

	utils.SetDir(".")
	utils.SetEnv(envs)

	err = utils.RunShell("/bin/bash", script)
	if err == nil && piperutils.ContainsString(config.FailOn, "BLOCKER") {
		violations := struct {
			PolicyViolations int      `json:"policyViolations"`
			Reports          []string `json:"reports"`
		}{
			PolicyViolations: 0,
			Reports:          []string{},
		}

		if files, err := utils.Glob("**/*BlackDuck_RiskReport.pdf"); err == nil && len(files) > 0 {
			// there should only be one RiskReport thus only taking the first one
			_, reportFile := filepath.Split(files[0])
			violations.Reports = append(violations.Reports, reportFile)
		}

		violationContent, err := json.Marshal(violations)
		if err != nil {
			return fmt.Errorf("failed to marshal policy violation data: %w", err)
		}

		err = utils.FileWrite("blackduck-ip.json", violationContent, 0666)
		if err != nil {
			return fmt.Errorf("failed to write policy violation report: %w", err)
		}
	} else if err != nil {
		// Setting error category based on exit code
		mapErrorCategory(utils.GetExitCode())

		// Error code mapping with more human readable text
		// log.Entry().Errorf("[ERROR ERRORF] => %v", exitCodeMapping(utils.GetExitCode()))
		return errors.Wrapf(err, exitCodeMapping(utils.GetExitCode()))
	}

	return err
}

// Get proper error category
func mapErrorCategory(exitCodeKey int) {
	switch exitCodeKey {
	case 1:
		log.SetErrorCategory(log.ErrorInfrastructure)
	case 2:
		log.SetErrorCategory(log.ErrorService)
	case 3:
		log.SetErrorCategory(log.ErrorCompliance)
	case 4:
		log.SetErrorCategory(log.ErrorInfrastructure)
	case 5:
		log.SetErrorCategory(log.ErrorConfiguration)
	case 6:
		log.SetErrorCategory(log.ErrorConfiguration)
	case 7:
		log.SetErrorCategory(log.ErrorConfiguration)
	case 9:
		log.SetErrorCategory(log.ErrorService)
	case 10:
		log.SetErrorCategory(log.ErrorService)
	case 11:
		log.SetErrorCategory(log.ErrorService)
	case 12:
		log.SetErrorCategory(log.ErrorInfrastructure)
	case 99:
		log.SetErrorCategory(log.ErrorService)
	case 100:
		log.SetErrorCategory(log.ErrorUndefined)
	default:
		log.SetErrorCategory(log.ErrorUndefined)
	}
}

// Exit codes/error code mapping
func exitCodeMapping(exitCodeKey int) string {

	exitCodes := map[int]string{
		0:   "SUCCESS => Detect exited successfully.",
		1:   "FAILURE_BLACKDUCK_CONNECTIVITY => Detect was unable to connect to Black Duck. Check your configuration and connection.",
		2:   "FAILURE_TIMEOUT => Detect could not wait for actions to be completed on Black Duck. Check your Black Duck server or increase your timeout.",
		3:   "FAILURE_POLICY_VIOLATION => Detect found policy violations.",
		4:   "FAILURE_PROXY_CONNECTIVITY => Detect was unable to use the configured proxy. Check your configuration and connection.",
		5:   "FAILURE_DETECTOR => Detect had one or more detector failures while extracting dependencies. Check that all projects build and your environment is configured correctly.",
		6:   "FAILURE_SCAN => Detect was unable to run the signature scanner against your source. Check your configuration.",
		7:   "FAILURE_CONFIGURATION => Detect was unable to start because of a configuration issue. Check and fix your configuration.",
		9:   "FAILURE_DETECTOR_REQUIRED => Detect did not run all of the required detectors. Fix detector issues or disable required detectors.",
		10:  "FAILURE_BLACKDUCK_VERSION_NOT_SUPPORTED => Detect attempted an operation that was not supported by your version of Black Duck. Ensure your Black Duck is compatible with this version of detect.",
		11:  "FAILURE_BLACKDUCK_FEATURE_ERROR => Detect encountered an error while attempting an operation on Black Duck. Ensure your Black Duck is compatible with this version of detect.",
		12:  "FAILURE_POLARIS_CONNECTIVITY => Detect was unable to connect to Polaris. Check your configuration and connection.",
		99:  "FAILURE_GENERAL_ERROR => Detect encountered a known error, details of the error are provided.",
		100: "FAILURE_UNKNOWN_ERROR => Detect encountered an unknown error.",
	}

	if _, isKeyExists := exitCodes[exitCodeKey]; isKeyExists {
		return exitCodes[exitCodeKey]
	}

	return "[" + strconv.Itoa(exitCodeKey) + "]: Not known exit code key"
}

func getDetectScript(config detectExecuteScanOptions, utils detectUtils) error {
	if config.ScanOnChanges {
		return utils.DownloadFile("https://raw.githubusercontent.com/blackducksoftware/detect_rescan/master/detect_rescan.sh", "detect.sh", nil, nil)
	}
	return utils.DownloadFile("https://detect.synopsys.com/detect.sh", "detect.sh", nil, nil)
}

func addDetectArgs(args []string, config detectExecuteScanOptions, utils detectUtils) ([]string, error) {
	detectVersionName := config.CustomScanVersion
	if len(detectVersionName) > 0 {
		log.Entry().Infof("Using custom version: %v", detectVersionName)
	} else {
		detectVersionName = versioning.ApplyVersioningModel(config.VersioningModel, config.Version)
	}
	//Split on spaces, the scanPropeties, so that each property is available as a single string
	//instead of all properties being part of a single string
	config.ScanProperties = piperutils.SplitAndTrim(config.ScanProperties, " ")

	if config.ScanOnChanges {
		args = append(args, "--report")
		config.Unmap = false
	}

	if config.Unmap {
		if !piperutils.ContainsString(config.ScanProperties, "--detect.project.codelocation.unmap=true") {
			args = append(args, fmt.Sprintf("--detect.project.codelocation.unmap=true"))
		}
		config.ScanProperties, _ = piperutils.RemoveAll(config.ScanProperties, "--detect.project.codelocation.unmap=false")
	} else {
		//When unmap is set to false, any occurances of unmap=true from scanProperties must be removed
		config.ScanProperties, _ = piperutils.RemoveAll(config.ScanProperties, "--detect.project.codelocation.unmap=true")
	}

	args = append(args, config.ScanProperties...)

	args = append(args, fmt.Sprintf("--blackduck.url=%v", config.ServerURL))
	args = append(args, fmt.Sprintf("--blackduck.api.token=%v", config.Token))
	// ProjectNames, VersionName, GroupName etc can contain spaces and need to be escaped using double quotes in CLI
	// Hence the string need to be surrounded by \"
	args = append(args, fmt.Sprintf("\"--detect.project.name='%v'\"", config.ProjectName))
	args = append(args, fmt.Sprintf("\"--detect.project.version.name='%v'\"", detectVersionName))

	// Groups parameter is added only when there is atleast one non-empty groupname provided
	if len(config.Groups) > 0 && len(config.Groups[0]) > 0 {
		args = append(args, fmt.Sprintf("\"--detect.project.user.groups='%v'\"", strings.Join(config.Groups, ",")))
	}

	// Atleast 1, non-empty category to fail on must be provided
	if len(config.FailOn) > 0 && len(config.FailOn[0]) > 0 {
		args = append(args, fmt.Sprintf("--detect.policy.check.fail.on.severities=%v", strings.Join(config.FailOn, ",")))
	}

	codelocation := config.CodeLocation
	if len(codelocation) == 0 && len(config.ProjectName) > 0 {
		codelocation = fmt.Sprintf("%v/%v", config.ProjectName, detectVersionName)
	}
	args = append(args, fmt.Sprintf("\"--detect.code.location.name='%v'\"", codelocation))

	if len(config.ScanPaths) > 0 && len(config.ScanPaths[0]) > 0 {
		args = append(args, fmt.Sprintf("--detect.blackduck.signature.scanner.paths=%v", strings.Join(config.ScanPaths, ",")))
	}

	if len(config.DependencyPath) > 0 {
		args = append(args, fmt.Sprintf("--detect.source.path=%v", config.DependencyPath))
	} else {
		args = append(args, fmt.Sprintf("--detect.source.path='.'"))
	}

	if len(config.IncludedPackageManagers) > 0 {
		args = append(args, fmt.Sprintf("--detect.included.detector.types=%v", strings.ToUpper(strings.Join(config.IncludedPackageManagers, ","))))
	}

	if len(config.ExcludedPackageManagers) > 0 {
		args = append(args, fmt.Sprintf("--detect.excluded.detector.types=%v", strings.ToUpper(strings.Join(config.ExcludedPackageManagers, ","))))
	}

	if len(config.MavenExcludedScopes) > 0 {
		args = append(args, fmt.Sprintf("--detect.maven.excluded.scopes=%v", strings.ToLower(strings.Join(config.MavenExcludedScopes, ","))))
	}

	if len(config.DetectTools) > 0 {
		args = append(args, fmt.Sprintf("--detect.tools=%v", strings.Join(config.DetectTools, ",")))
	}

	mavenArgs, err := maven.DownloadAndGetMavenParameters(config.GlobalSettingsFile, config.ProjectSettingsFile, utils)
	if err != nil {
		return nil, err
	}

	if len(config.M2Path) > 0 {
		absolutePath, err := utils.Abs(config.M2Path)
		if err != nil {
			return nil, err
		}
		mavenArgs = append(mavenArgs, fmt.Sprintf("-Dmaven.repo.local=%v", absolutePath))
	}

	if len(mavenArgs) > 0 {
		args = append(args, fmt.Sprintf("\"--detect.maven.build.command='%v'\"", strings.Join(mavenArgs, " ")))
	}

	return args, nil
}

// create toolrecord file for detect
//
//
func createToolRecordDetect(workspace string, config detectExecuteScanOptions) (string, error) {
	record := toolrecord.New(workspace, "detectExecute", config.ServerURL)

	projectId := ""  // todo needs more research; according to synopsis documentation
	productURL := "" // relevant ids can be found in the logfile
	err := record.AddKeyData("project",
		projectId,
		config.ProjectName,
		productURL)
	if err != nil {
		return "", err
	}
	record.AddContext("DetectTools", config.DetectTools)
	err = record.Persist()
	if err != nil {
		return "", err
	}
	return record.GetFileName(), nil
}
