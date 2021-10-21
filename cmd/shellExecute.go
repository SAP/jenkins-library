package cmd

import (
	"fmt"
	"github.com/hashicorp/vault/api"
	"os/exec"

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

	err := runShellExecute(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runShellExecute(config *shellExecuteOptions, telemetryData *telemetry.CustomData, utils shellExecuteUtils) error {
	// create vault client
	// try to retrieve existing credentials
	// if it's impossible - will add it
	vaultConfig := &vault.Config{
		Config: &api.Config{
			Address: config.VaultServerURL,
		},
		Namespace: config.VaultNamespace,
	}
	client, err := vault.NewClientWithAppRole(vaultConfig, GeneralConfig.VaultRoleID, GeneralConfig.VaultRoleSecretID)
	if err != nil {
		log.Entry().WithError(err).Fatal("could not create vault client")
	}
	defer client.MustRevokeToken()

	// check if all scripts are present
	for _, script := range config.Script {
		exists, err := utils.FileExists(script)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return fmt.Errorf("failed to check for defined script: %w", err)
		}
		if !exists {
			log.SetErrorCategory(log.ErrorConfiguration)
			return fmt.Errorf("the specified script could not be found: %w", err)
		}
	}

	// if all ok - try to run them one by one
	for _, script := range config.Script {
		_, err := exec.Command(script).Output()
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
				// build was failed
				return fmt.Errorf("build was failed: %w", err)
			default:
				// exit code 2 or >2 - build unstable
				return fmt.Errorf("build unstable or something went wrong: %w", err)
			}
		}
	}

	return nil
}
