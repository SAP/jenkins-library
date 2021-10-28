package cmd

import (
	"io"
	"os"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/bmatcuk/doublestar"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type checkStepActiveCommandOptions struct {
	openFile        func(s string, t map[string]string) (io.ReadCloser, error)
	stageConfigFile string
	stepName        string
}

var checkStepActiveOptions checkStepActiveCommandOptions

// CheckStepActiveCommand is the entry command for checking if a step is active in a defined stage
func CheckStepActiveCommand() *cobra.Command {
	checkStepActiveOptions.openFile = config.OpenPiperFile
	var checkStepActiveCmd = &cobra.Command{
		Use:   "checkIfStepActive",
		Short: "Checks if a step is active in a defined stage.",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)
			initStageName(false)
			if GeneralConfig.StageName == "" {
				return errors.New("required flag 'stageName' not set")
			}
			log.SetVerbose(GeneralConfig.Verbose)
			GeneralConfig.GitHubAccessTokens = ResolveAccessTokens(GeneralConfig.GitHubTokens)
			return nil
		},
		Run: func(cmd *cobra.Command, _ []string) {
			err := checkIfStepActive()
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				log.Entry().WithError(err).Fatal("Checking for an active step failed")
			}
		},
	}
	addCheckStepActiveFlags(checkStepActiveCmd)
	return checkStepActiveCmd
}

func checkIfStepActive() error {
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

	// load and evaluate step conditions
	stageConditions := &config.RunConfig{StageConfigFile: stageConfigFile}
	err = stageConditions.InitRunConfig(projectConfig, nil, nil, nil, nil, doublestar.Glob, checkStepActiveOptions.openFile)
	if err != nil {
		return err
	}

	log.Entry().Debugf("RunSteps: %v", stageConditions.RunSteps)

	stageName := GeneralConfig.StageName
	if !stageConditions.RunSteps[stageName][checkStepActiveOptions.stepName] {
		return errors.Errorf("Step %s in stage %s is not active", checkStepActiveOptions.stepName, stageName)
	}
	log.Entry().Infof("Step %s in stage %s is active", checkStepActiveOptions.stepName, stageName)

	return nil
}

func addCheckStepActiveFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&checkStepActiveOptions.stageConfigFile, "stageConfig", ".resources/piper-stage-config.yml",
		"Default config of piper pipeline stages")
	cmd.Flags().StringVar(&checkStepActiveOptions.stepName, "step", "", "Name of the step being checked")
	cmd.MarkFlagRequired("step")
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
