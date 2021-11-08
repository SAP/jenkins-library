package cmd

import (
	"io"
	"os"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/bmatcuk/doublestar"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type checkStepActiveCommandOptions struct {
	openFile        func(s string, t map[string]string) (io.ReadCloser, error)
	stageConfigFile string
	stepName        string
	stageName       string
	v1Active        bool
}

var checkStepActiveOptions checkStepActiveCommandOptions

// CheckStepActiveCommand is the entry command for checking if a step is active in a defined stage
func CheckStepActiveCommand() *cobra.Command {
	checkStepActiveOptions.openFile = config.OpenPiperFile
	var checkStepActiveCmd = &cobra.Command{
		Use:   "checkIfStepActive",
		Short: "Checks if a step is active in a defined stage.",
		PreRun: func(cmd *cobra.Command, args []string) {
			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)
			initStageName(false)
			log.SetVerbose(GeneralConfig.Verbose)
			GeneralConfig.GitHubAccessTokens = ResolveAccessTokens(GeneralConfig.GitHubTokens)
		},
		Run: func(cmd *cobra.Command, _ []string) {
			utils := &piperutils.Files{}
			err := checkIfStepActive(utils)
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				log.Entry().WithError(err).Fatal("Checking for an active step failed")
			}
		},
	}
	addCheckStepActiveFlags(checkStepActiveCmd)
	return checkStepActiveCmd
}

func checkIfStepActive(utils piperutils.FileUtils) error {
	var pConfig config.Config

	// load project config and defaults
	projectConfig, err := initializeConfig(&pConfig)
	if err != nil {
		log.Entry().Errorf("Failed to load project config: %v", err)
	}

	stageConfigFile, err := checkStepActiveOptions.openFile(checkStepActiveOptions.stageConfigFile, GeneralConfig.GitHubAccessTokens)
	if err != nil {
		return errors.Wrapf(err, "config: open stage configuration file '%v' failed", checkStepActiveOptions.stageConfigFile)
	}
	defer stageConfigFile.Close()

	runSteps := map[string]map[string]bool{}
	runStages := map[string]bool{}

	// load and evaluate step conditions
	if checkStepActiveOptions.v1Active {
		runConfig := config.RunConfig{StageConfigFile: stageConfigFile}
		runConfigV1 := &config.RunConfigV1{RunConfig: runConfig}
		err = runConfigV1.InitRunConfigV1(projectConfig, nil, nil, nil, nil, utils)
		if err != nil {
			return err
		}
		runSteps = runConfigV1.RunSteps
		runStages = runConfigV1.RunStages
	} else {
		runConfig := &config.RunConfig{StageConfigFile: stageConfigFile}
		err = runConfig.InitRunConfig(projectConfig, nil, nil, nil, nil, doublestar.Glob, checkStepActiveOptions.openFile)
		if err != nil {
			return err
		}
		runSteps = runConfig.RunSteps
		runStages = runConfig.RunStages
	}

	log.Entry().Debugf("RunSteps: %v", runSteps)
	log.Entry().Debugf("RunStages: %v", runStages)

	if !runSteps[checkStepActiveOptions.stageName][checkStepActiveOptions.stepName] {
		return errors.Errorf("Step %s in stage %s is not active", checkStepActiveOptions.stepName, checkStepActiveOptions.stageName)
	}
	log.Entry().Infof("Step %s in stage %s is active", checkStepActiveOptions.stepName, checkStepActiveOptions.stageName)

	return nil
}

func addCheckStepActiveFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&checkStepActiveOptions.stageConfigFile, "stageConfig", ".resources/piper-stage-config.yml",
		"Default config of piper pipeline stages")
	cmd.Flags().StringVar(&checkStepActiveOptions.stepName, "step", "", "Name of the step being checked")
	cmd.Flags().StringVar(&checkStepActiveOptions.stageName, "stage", "", "Name of the stage in which contains the step being checked")
	cmd.Flags().BoolVar(&checkStepActiveOptions.v1Active, "useV1", false, "Use new CRD-style stage configuration")
	cmd.MarkFlagRequired("step")
	cmd.MarkFlagRequired("stage")
}

func initializeConfig(pConfig *config.Config) (*config.Config, error) {
	projectConfigFile := getProjectConfigFile(GeneralConfig.CustomConfig)
	customConfig, err := checkStepActiveOptions.openFile(projectConfigFile, GeneralConfig.GitHubAccessTokens)
	if err != nil {
		return nil, errors.Wrapf(err, "config: open configuration file '%v' failed", projectConfigFile)
	}
	defer customConfig.Close()

	defaultConfig := []io.ReadCloser{}
	for _, f := range GeneralConfig.DefaultConfig {
		fc, err := checkStepActiveOptions.openFile(f, GeneralConfig.GitHubAccessTokens)
		// only create error for non-default values
		if err != nil && f != ".pipeline/defaults.yaml" {
			return nil, errors.Wrapf(err, "config: getting defaults failed: '%v'", f)
		}
		if err == nil {
			defaultConfig = append(defaultConfig, fc)
		}
	}
	var flags map[string]interface{}
	stepAliase := []config.Alias{}
	filter := config.StepFilters{
		All:     []string{},
		General: []string{},
		Stages:  []string{},
		Steps:   []string{},
		Env:     []string{},
	}

	_, err = pConfig.GetStepConfig(flags, "", customConfig, defaultConfig, GeneralConfig.IgnoreCustomDefaults, filter, nil, nil, nil, "", "",
		stepAliase)
	if err != nil {
		return nil, errors.Wrap(err, "getting step config failed")
	}
	return pConfig, nil
}
