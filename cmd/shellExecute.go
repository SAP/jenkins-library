package cmd

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"github.com/pkg/errors"

	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type shellExecuteUtils interface {
	command.ExecRunner
	piperutils.FileUtils
	piperhttp.Downloader
}

type shellExecuteUtilsBundle struct {
	*command.Command
	*piperutils.Files
	*piperhttp.Client
}

func newShellExecuteUtils() shellExecuteUtils {
	utils := shellExecuteUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
		Client:  &piperhttp.Client{},
	}
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func shellExecute(config shellExecuteOptions, telemetryData *telemetry.CustomData) {
	utils := newShellExecuteUtils()

	err := runShellExecute(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runShellExecute(config *shellExecuteOptions, telemetryData *telemetry.CustomData, utils shellExecuteUtils) error {
	// check input data
	// example for script: sources: ["./script.sh"]
	for _, source := range config.Sources {

		if strings.Contains(source, "https") {
			scriptName, err := downloadScripts(config, utils, source)
			if err != nil {
				return errors.Wrapf(err, "script download error")
			}
			source = scriptName
		}
		// check if the script is physically present
		exists, err := utils.FileExists(source)
		if err != nil {
			log.Entry().WithError(err).Error("failed to check for defined script")
			return fmt.Errorf("failed to check for defined script: %w", err)
		}
		if !exists {
			log.Entry().WithError(err).Errorf("the script '%v' could not be found: %v", source, err)
			return fmt.Errorf("the script '%v' could not be found", source)
		}
		log.Entry().Info("starting running script:", source)

		err = utils.RunExecutable(source)
		if err != nil {
			log.Entry().Errorln("starting running script:", source)
		}
		// handle exit code
		if ee, ok := err.(*exec.ExitError); ok {
			switch ee.ExitCode() {
			case 0:
				// success
				return nil
			case 1:
				return errors.Wrap(err, "an error occurred while executing the script")
			default:
				// exit code 2 or >2 - unstable
				return errors.Wrap(err, "script execution unstable or something went wrong")
			}
		} else if err != nil {
			return errors.Wrap(err, "script execution error occurred")
		}
	}

	return nil
}

func downloadScripts(config *shellExecuteOptions, utils shellExecuteUtils, url string) (string, error) {
	header := http.Header{}
	if len(config.GithubToken) > 0 {
		header = http.Header{"Authorization": []string{"Token " + config.GithubToken}}
	}

	log.Entry().Infof("downloading script : %v", url)
	fileNameParts := strings.Split(url, "/")
	fileName := fileNameParts[len(fileNameParts)-1]
	err := utils.DownloadFile(url, fileName, header, []*http.Cookie{})
	if err != nil {
		return "", errors.Wrapf(err, "unable to download script from %v", url)
	}
	log.Entry().Infof("downloaded script %v successfully", url)
	err = utils.Chmod(fileName, 0700)
	if err != nil {
		return "", fmt.Errorf("unable to change file permission for script '%v'", fileName)
	}

	return "./" + fileName, nil
}
