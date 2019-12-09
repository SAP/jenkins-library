package cmd

import (
	"fmt"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
)

func detectExecuteScan(myDetectExecuteScanOptions detectExecuteScanOptions) error {
	c := command.Command{}
	// reroute command output to loging framework
	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())
	runDetect(myDetectExecuteScanOptions, &c)
	return nil
}

func runDetect(myDetectExecuteScanOptions detectExecuteScanOptions, command shellRunner) error {
	// detect execution details, see https://synopsys.atlassian.net/wiki/spaces/INTDOCS/pages/88440888/Sample+Synopsys+Detect+Scan+Configuration+Scenarios+for+Black+Duck

	args := []string{"bash <(curl -s https://detect.synopsys.com/detect.sh)"}
	args = addDetectArgs(args, myDetectExecuteScanOptions)
	script := strings.Join(args, " ")

	command.Dir(".")

	err := command.RunShell("/bin/bash", script)
	if err != nil {
		log.Entry().
			WithError(err).
			WithField("command", myKarmaExecuteTestsOptions.InstallCommand).
			Fatal("failed to execute detect scan")
	}
	return nil
}

func addDetectArgs(args []string, myDetectExecuteScanOptions detectExecuteScanOptions) []string {

	args = append(args, myDetectExecuteScanOptions.ScanProperties...)

	args = append(args, fmt.Sprintf("--blackduck.url=%v", myDetectExecuteScanOptions.ServerURL))
	args = append(args, fmt.Sprintf("--blackduck.api.token=%v", myDetectExecuteScanOptions.APIToken))

	args = append(args, fmt.Sprintf("--detect.project.name=%v", myDetectExecuteScanOptions.ProjectName))
	args = append(args, fmt.Sprintf("--detect.project.version.name=%v", myDetectExecuteScanOptions.ProjectVersion))
	codeLocation := myDetectExecuteScanOptions.CodeLocation
	if len(codeLocation) == 0 {
		codeLocation = fmt.Sprintf("%v/%v", myDetectExecuteScanOptions.ProjectName, myDetectExecuteScanOptions.ProjectVersion)
	}
	args = append(args, fmt.Sprintf("--detect.code.location.name=%v", codeLocation))

	if sliceContains(myDetectExecuteScanOptions.Scanners, "signature") {
		args = append(args, fmt.Sprintf("--detect.blackduck.signature.scanner.paths=%v", strings.Join(myDetectExecuteScanOptions.ScanPaths, ",")))
	}

	if sliceContains(myDetectExecuteScanOptions.Scanners, "source") {
		args = append(args, fmt.Sprintf("--detect.source.path=%v", myDetectExecuteScanOptions.ScanPaths[0]))
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
