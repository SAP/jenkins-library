package cmd

import (
	"net/url"
	"os/exec"
	"strings"

	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"

	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/vault"
)

type shellExecuteUtils interface {
	command.ExecRunner
	FileExists(filename string) (bool, error)
}

type shellExecuteUtilsBundle struct {
	*vault.Client
	*command.Command
	*piperutils.Files
}

func newShellExecuteUtils() shellExecuteUtils {
	utils := shellExecuteUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func shellExecute(config shellExecuteOptions, telemetryData *telemetry.CustomData) {
	utils := newShellExecuteUtils()
	fileUtils := &piperutils.Files{}

	err := runShellExecute(&config, telemetryData, utils, fileUtils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runShellExecute(config *shellExecuteOptions, telemetryData *telemetry.CustomData, utils shellExecuteUtils, fileUtils piperutils.FileUtils) error {
	// create vault client
	// try to retrieve existing credentials
	// if it's impossible - will add it
	vaultConfig := &vault.Config{
		Config: &api.Config{
			Address: config.VaultServerURL,
		},
		Namespace: config.VaultNamespace,
	}
	_, err := vault.NewClientWithAppRole(vaultConfig, GeneralConfig.VaultRoleID, GeneralConfig.VaultRoleSecretID)
	if err != nil {
		log.Entry().Info("could not create vault client:", err)
	}

	// piper http client for downloading scripts
	httpClient := piperhttp.Client{}

	// scripts for running locally
	var e []string

	// check input data
	// example for script: sources: ["./script.sh"]
	for _, source := range config.Sources {
		// check it's a local script or remote
		_, err := url.ParseRequestURI(source)
		if err != nil {
			// err means that it's not a remote script
			// check if the script is physically present (for local scripts)
			exists, err := fileUtils.FileExists(source)
			if err != nil {
				log.Entry().WithError(err).Error("failed to check for defined script")
				return errors.Wrap(err, "failed to check for defined script")
			}
			if !exists {
				log.Entry().WithError(err).Error("the specified script could not be found")
				return errors.New("the specified script could not be found")
			}
			e = append(e, source)
		} else {
			// this block means that it's a remote script
			// so, need to download it before
			// get script name at first
			path := strings.Split(source, "/")
			err = httpClient.DownloadFile(source, path[len(path)-1], nil, nil)
			if err != nil {
				log.Entry().WithError(err).Errorf("the specified script could not be downloaded")
			}
			// make script executable
			exec.Command("/bin/sh", "chmod +x "+path[len(path)-1])

			e = append(e, path[len(path)-1])

		}
	}

	// if all ok - try to run them one by one
	for _, script := range e {
		log.Entry().Info("starting running script:", script)
		err = utils.RunExecutable(script)
		if err != nil {
			log.Entry().Errorln("starting running script:", script)
		}

		// if it's an exit error, then check the exit code
		// according to the requirements
		// 0 - success
		// 1 - fails the build (or > 2)
		// 2 - build unstable - unsupported now
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
