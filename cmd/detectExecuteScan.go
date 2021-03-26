package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/maven"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
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

	utils.SetDir(".")
	utils.SetEnv(envs)

	return utils.RunShell("/bin/bash", script)
}

func getDetectScript(config detectExecuteScanOptions, utils detectUtils) error {
	if config.ScanOnChanges {
		return utils.DownloadFile("https://raw.githubusercontent.com/blackducksoftware/detect_rescan/master/detect_rescan.sh", "detect.sh", nil, nil)
	}
	return utils.DownloadFile("https://detect.synopsys.com/detect.sh", "detect.sh", nil, nil)
}

func addDetectArgs(args []string, config detectExecuteScanOptions, utils detectUtils) ([]string, error) {

	coordinates := versioning.Coordinates{
		Version: config.Version,
	}

	_, detectVersionName := versioning.DetermineProjectCoordinates("", config.VersioningModel, coordinates)

	//Split on spaces, the scanPropeties, so that each property is available as a single string
	//instead of all properties being part of a single string
	config.ScanProperties = piperutils.SplitAndTrim(config.ScanProperties, " ")

	if config.ScanOnChanges {
		args = append(args, "--report")
		config.Unmap = false
	}

	if config.Unmap {
		args = append(args, fmt.Sprintf("--detect.project.codelocation.unmap=true"))
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

	if len(config.CodeLocation) > 0 {
		args = append(args, fmt.Sprintf("\"--detect.code.location.name='%v'\"", config.CodeLocation))
	}

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
