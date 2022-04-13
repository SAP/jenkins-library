package cmd

import (
	"fmt"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type isChangeInDevelopmentUtils interface {
	command.ExecRunner
	GetExitCode() int

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The isChangeInDevelopmentUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type isChangeInDevelopmentUtilsBundle struct {
	*command.Command

	// Embed more structs as necessary to implement methods or interfaces you add to isChangeInDevelopmentUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// isChangeInDevelopmentUtilsBundle and forward to the implementation of the dependency.
}

func newIsChangeInDevelopmentUtils() isChangeInDevelopmentUtils {
	utils := isChangeInDevelopmentUtilsBundle{
		Command: &command.Command{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func isChangeInDevelopment(config isChangeInDevelopmentOptions,
	telemetryData *telemetry.CustomData,
	commonPipelineEnvironment *isChangeInDevelopmentCommonPipelineEnvironment) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newIsChangeInDevelopmentUtils()

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runIsChangeInDevelopment(&config, telemetryData, utils, commonPipelineEnvironment)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runIsChangeInDevelopment(config *isChangeInDevelopmentOptions,
	telemetryData *telemetry.CustomData,
	utils isChangeInDevelopmentUtils,
	commonPipelineEnvironment *isChangeInDevelopmentCommonPipelineEnvironment) error {

	log.Entry().Infof("Checking change status for change '%s'", config.ChangeDocumentID)

	isInDevelopment, err := perform(config, utils)

	if err != nil {
		return err
	}

	commonPipelineEnvironment.custom.isChangeInDevelopment = isInDevelopment

	if isInDevelopment {
		log.Entry().Infof("Change '%s' is in status 'in development'.", config.ChangeDocumentID)
		return nil
	}
	if config.FailIfStatusIsNotInDevelopment {
		return fmt.Errorf("change '%s' is not in status 'in development'", config.ChangeDocumentID)
	}
	log.Entry().Warningf("Change '%s' is not in status 'in development'. Failing the step has been explicitly disabled.", config.ChangeDocumentID)
	return nil
}

func perform(config *isChangeInDevelopmentOptions, utils isChangeInDevelopmentUtils) (bool, error) {

	if len(config.CmClientOpts) > 0 {
		utils.AppendEnv([]string{fmt.Sprintf("CMCLIENT_OPTS=%s", strings.Join(config.CmClientOpts, " "))})
	}

	err := utils.RunExecutable("cmclient",
		"--endpoint", config.Endpoint,
		"--user", config.Username,
		"--password", config.Password,
		"is-change-in-development",
		"--change-id", config.ChangeDocumentID,
		"--return-code")

	exitCode := utils.GetExitCode()

	hint := "check log for details"
	if err != nil {
		hint = err.Error()
	}

	if exitCode == 0 {
		return true, nil
	} else if exitCode == 3 {
		return false, nil
	} else if exitCode == 2 {
		hint = "invalid credentials"
	}

	return false, fmt.Errorf("cannot retrieve change status: %s", hint)
}
