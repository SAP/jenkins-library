package cmd

import (
	"fmt"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func detectExecuteScan(config detectExecuteScanOptions, telemetryData *telemetry.CustomData) {
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())
	runDetect(config, &c)
}

func runDetect(config detectExecuteScanOptions, command shellRunner) {
	// detect execution details, see https://synopsys.atlassian.net/wiki/spaces/INTDOCS/pages/88440888/Sample+Synopsys+Detect+Scan+Configuration+Scenarios+for+Black+Duck

	args := []string{"bash <(curl -s https://detect.synopsys.com/detect.sh)"}
	args = addDetectArgs(args, config)
	script := strings.Join(args, " ")

	command.SetDir(".")

	err := command.RunShell("/bin/bash", script)
	if err != nil {
		log.Entry().
			WithError(err).
			Fatal("failed to execute detect scan")
	}
}

func addDetectArgs(args []string, config detectExecuteScanOptions) []string {

	args = append(args, config.ScanProperties...)

	args = append(args, fmt.Sprintf("--blackduck.url=%v", config.ServerURL))
	args = append(args, fmt.Sprintf("--blackduck.api.token=%v", config.APIToken))

	args = append(args, fmt.Sprintf("--detect.project.name=%v", config.ProjectName))
	args = append(args, fmt.Sprintf("--detect.project.version.name=%v", config.ProjectVersion))
	codeLocation := config.CodeLocation
	if len(codeLocation) == 0 && len(config.ProjectName) > 0 {
		codeLocation = fmt.Sprintf("%v/%v", config.ProjectName, config.ProjectVersion)
	}
	args = append(args, fmt.Sprintf("--detect.code.location.name=%v", codeLocation))

	if sliceContains(config.Scanners, "signature") {
		args = append(args, fmt.Sprintf("--detect.blackduck.signature.scanner.paths=%v", strings.Join(config.ScanPaths, ",")))
	}

	if sliceContains(config.Scanners, "source") {
		args = append(args, fmt.Sprintf("--detect.source.path=%v", config.ScanPaths[0]))
	}
	return args
}

func sliceContains(slice []string, find string) bool {
	for _, elem := range slice {
		if elem == find {
			return true
		}
	}
	return false
}
