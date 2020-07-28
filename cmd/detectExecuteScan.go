package cmd

import (
	"fmt"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"strings"

	sliceUtils "github.com/SAP/jenkins-library/pkg/piperutils"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/versioning"
)

func detectExecuteScan(config detectExecuteScanOptions, telemetryData *telemetry.CustomData) {
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	fileUtils := piperutils.Files{}
	httpClient := piperhttp.Client{}

	runDetect(config, &c, &fileUtils, &httpClient)
}

func runDetect(config detectExecuteScanOptions, command command.ShellRunner, fileUtils piperutils.FileUtils, httpClient piperhttp.Downloader) {
	// detect execution details, see https://synopsys.atlassian.net/wiki/spaces/INTDOCS/pages/88440888/Sample+Synopsys+Detect+Scan+Configuration+Scenarios+for+Black+Duck

	args := []string{"bash <(curl -s https://detect.synopsys.com/detect.sh)"}
	args = addDetectArgs(args, config)
	script := strings.Join(args, " ")

	envs := []string{"BLACKDUCK_SKIP_PHONE_HOME=true"}
	if len(config.M2Path) > 0 {
		absolutePath, err := fileUtils.Abs(config.M2Path)
		if err != nil {
			log.Entry().
				WithError(err).
				Fatal("failed to execute detect scan")
		}
		envs = append(envs, "MAVEN_OPTS=-Dmaven.repo.local=" + absolutePath)
	}

	command.SetDir(".")
	command.SetEnv(envs)

	err := command.RunShell("/bin/bash", script)
	if err != nil {
		log.Entry().
			WithError(err).
			Fatal("failed to execute detect scan")
	}
}

func addDetectArgs(args []string, config detectExecuteScanOptions) []string {

	coordinates := struct {
		Version string
	}{
		Version: config.Version,
	}

	_, detectVersionName := versioning.DetermineProjectCoordinates("", config.VersioningModel, coordinates)

	args = append(args, config.ScanProperties...)

	args = append(args, fmt.Sprintf("--blackduck.url=%v", config.ServerURL))
	args = append(args, fmt.Sprintf("--blackduck.api.token=%v", config.APIToken))

	args = append(args, fmt.Sprintf("--detect.project.name=%v", config.ProjectName))
	args = append(args, fmt.Sprintf("--detect.project.version.name=%v", detectVersionName))
	codeLocation := config.CodeLocation
	if len(codeLocation) == 0 && len(config.ProjectName) > 0 {
		codeLocation = fmt.Sprintf("%v/%v", config.ProjectName, detectVersionName)
	}
	args = append(args, fmt.Sprintf("--detect.code.location.name=%v", codeLocation))

	if sliceUtils.ContainsString(config.Scanners, "signature") {
		args = append(args, fmt.Sprintf("--detect.blackduck.signature.scanner.paths=%v", strings.Join(config.ScanPaths, ",")))
	}

	if sliceUtils.ContainsString(config.Scanners, "source") {
		args = append(args, fmt.Sprintf("--detect.source.path=%v", config.ScanPaths[0]))
	}
	return args
}
