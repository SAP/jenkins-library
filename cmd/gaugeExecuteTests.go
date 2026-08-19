package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"errors"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

var ErrorGaugeInstall error = errors.New("error installing gauge")
var ErrorGaugeRunnerInstall error = errors.New("error installing runner")
var ErrorGaugeRun error = errors.New("error running gauge")

type gaugeExecuteTestsUtils interface {
	FileExists(filename string) (bool, error)
	MkdirAll(path string, perm os.FileMode) error
	SetEnv([]string)
	RunExecutable(executable string, params ...string) error
	Stdout(io.Writer)
	Getenv(key string) string
}

type gaugeExecuteTestsUtilsBundle struct {
	*command.Command
	*piperutils.Files
}

func newGaugeExecuteTestsUtils() gaugeExecuteTestsUtils {
	utils := gaugeExecuteTestsUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func gaugeExecuteTests(config gaugeExecuteTestsOptions, telemetryData *telemetry.CustomData, influx *gaugeExecuteTestsInflux) {
	utils := newGaugeExecuteTestsUtils()

	influx.step_data.fields.gauge = false
	err := runGaugeExecuteTests(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
	influx.step_data.fields.gauge = true
}

func runGaugeExecuteTests(config *gaugeExecuteTestsOptions, telemetryData *telemetry.CustomData, utils gaugeExecuteTestsUtils) error {
	if config.InstallCommand != "" {
		err := installGauge(config.InstallCommand, utils)
		if err != nil {
			return err
		}
	}

	if config.LanguageRunner != "" {
		err := installLanguageRunner(config.LanguageRunner, utils)
		if err != nil {
			return err
		}
	}

	err := runGauge(config, utils)
	if err != nil {
		return fmt.Errorf("failed to run gauge: %w", err)
	}

	return nil
}

func installGauge(gaugeInstallCommand string, utils gaugeExecuteTestsUtils) error {
	installGaugeTokens := strings.Split(gaugeInstallCommand, " ")
	installGaugeTokens = append(installGaugeTokens, "--prefix=~/.npm-global")
	err := utils.RunExecutable(installGaugeTokens[0], installGaugeTokens[1:]...)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return fmt.Errorf("%s: %w", err.Error(), ErrorGaugeInstall)
	}

	return nil
}

func installLanguageRunner(languageRunner string, utils gaugeExecuteTestsUtils) error {
	installParams := []string{"install", languageRunner}
	gaugePath := filepath.FromSlash(filepath.Join(utils.Getenv("HOME"), "/.npm-global/bin/gauge"))
	err := utils.RunExecutable(gaugePath, installParams...)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return fmt.Errorf("%s: %w", err.Error(), ErrorGaugeRunnerInstall)
	}
	return nil

}

func runGauge(config *gaugeExecuteTestsOptions, utils gaugeExecuteTestsUtils) error {
	runCommandTokens := strings.Split(config.RunCommand, " ")
	if config.TestOptions != "" {
		runCommandTokens = append(runCommandTokens, strings.Split(config.TestOptions, " ")...)
	}
	gaugePath := filepath.FromSlash(filepath.Join(utils.Getenv("HOME"), "/.npm-global/bin/gauge"))
	err := utils.RunExecutable(gaugePath, runCommandTokens...)
	if err != nil {
		return fmt.Errorf("%s: %w", err.Error(), ErrorGaugeRun)
	}
	return nil
}

func (utils gaugeExecuteTestsUtilsBundle) Getenv(key string) string {
	return os.Getenv(key)
}
