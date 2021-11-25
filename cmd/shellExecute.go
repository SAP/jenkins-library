package cmd

import (
	"os/exec"

	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"

	"github.com/SAP/jenkins-library/pkg/command"
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

	// no need to create a vault client, just need to resolve variables for scripts
	_, err := vault.NewClientWithAppRole(vaultConfig, GeneralConfig.VaultRoleID, GeneralConfig.VaultRoleSecretID)
	if err != nil {
		log.Entry().Info("could not create vault client:", err)
	}

	// check input data
	// example for script: sources: ["./script.sh"]
	for _, source := range config.Sources {
		// check if the script is physically present
		exists, err := fileUtils.FileExists(source)
		if err != nil {
			log.Entry().WithError(err).Error("failed to check for defined script")
			return errors.Wrap(err, "failed to check for defined script")
		}
		if !exists {
			log.Entry().WithError(err).Error("the specified script could not be found")
			return errors.New("the specified script could not be found")
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
