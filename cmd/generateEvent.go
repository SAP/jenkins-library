package cmd

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/command"
	piperConfig "github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type generateEventUtils interface {
	command.ExecRunner

	FileExists(filename string) (bool, error)

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The generateEventUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type generateEventUtilsBundle struct {
	*command.Command
	*piperutils.Files

	// Embed more structs as necessary to implement methods or interfaces you add to generateEventUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// generateEventUtilsBundle and forward to the implementation of the dependency.
}

func newGenerateEventUtils() generateEventUtils {
	utils := generateEventUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func generateEvent(config generateEventOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newGenerateEventUtils()

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runGenerateEvent(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runGenerateEvent(config *generateEventOptions, StetelemetryData *telemetry.CustomData, utils generateEventUtils) error {
	log.Entry().WithField("LogField", "Log field content").Info("This is just a demo for a simple step.")

	vaultCreds := piperConfig.VaultCredentials{
		AppRoleID:       GeneralConfig.VaultRoleID,
		AppRoleSecretID: GeneralConfig.VaultRoleSecretID,
		VaultToken:      GeneralConfig.VaultToken,
	}
	// GeneralConfig VaultServerURL and VaultNamespace are empty swicthing to stepConfig
	var vaultConfig = map[string]interface{}{
		"vaultServerUrl": config.VaultServerURL,
		"vaultNamespace": config.VaultNamespace,
	}

	stepConfig := piperConfig.StepConfig{
		Config: vaultConfig,
	}
	// Generating vault client
	vaultClient, err := piperConfig.GetVaultClientFromConfig(stepConfig, vaultCreds)
	if err != nil {
		log.Entry().WithError(err).Fatal("getting vault client failed")
	}
	// Getting oidc token and setting it in environment variable
	_, err = vaultClient.GetOidcTokenByValidation(GeneralConfig.HookConfig.OidcConfig.RoleID)
	if err != nil {
		log.Entry().WithError(err).Fatal("getting oidc token failed")
	}
	// Example of calling methods from external dependencies directly on utils:
	exists, err := utils.FileExists("file.txt")
	if err != nil {
		// It is good practice to set an error category.
		// Most likely you want to do this at the place where enough context is known.
		log.SetErrorCategory(log.ErrorConfiguration)
		// Always wrap non-descriptive errors to enrich them with context for when they appear in the log:
		return fmt.Errorf("failed to check for important file: %w", err)
	}
	if !exists {
		log.SetErrorCategory(log.ErrorConfiguration)
		return fmt.Errorf("cannot run without important file")
	}

	return nil
}
