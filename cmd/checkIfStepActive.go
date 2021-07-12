package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type checkStepActiveCommandOptions struct {
	openFile        func(s string, t map[string]string) (io.ReadCloser, error)
	stageConfigFile string
	step            string
	stage           string
}

var checkStepActiveOptions checkStepActiveCommandOptions

// CheckStepCommand is the entry command for checking if a step is active in a defined stage
func CheckStepActiveCommand() *cobra.Command {

	checkStepActiveOptions.openFile = config.OpenPiperFile
	var checkStepActiveCmd = &cobra.Command{
		Use:   "checkIfStepActive",
		Short: "Checks is a step active in a defined stage.",
		PreRun: func(cmd *cobra.Command, args []string) {
			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)
			initStageName(false)
			GeneralConfig.GitHubAccessTokens = ResolveAccessTokens(GeneralConfig.GitHubTokens)
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

	var myConfig config.Config

	projectConfigFile := getProjectConfigFile(GeneralConfig.CustomConfig)

	customConfig, err := checkStepActiveOptions.openFile(projectConfigFile, GeneralConfig.GitHubAccessTokens)
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrapf(err, "config: open configuration file '%v' failed", projectConfigFile)
		}
		customConfig = nil
	}

	defaultConfig := []io.ReadCloser{}
	if err != nil {
		return errors.Wrap(err, "defaults: retrieving step defaults failed")
	}

	for _, f := range GeneralConfig.DefaultConfig {
		fc, err := checkStepActiveOptions.openFile(f, GeneralConfig.GitHubAccessTokens)
		// only create error for non-default values
		if err != nil && f != ".pipeline/defaults.yaml" {
			return errors.Wrapf(err, "config: getting defaults failed: '%v'", f)
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

	_, err = myConfig.GetStepConfig(flags, "", customConfig, defaultConfig, GeneralConfig.IgnoreCustomDefaults, filter, nil, nil, nil, "", "", stepAliase)
	if err != nil {
		return errors.Wrap(err, "getting step config failed")
	}

	// load and evaluate step conditions
	stageConditions := &config.RunConfig{ConditionFilePath: checkStepActiveOptions.stageConfigFile}
	err = stageConditions.InitRunConfig(&myConfig, myConfig.Stages, nil, nil, nil, nil, nil)
	if err != nil {
		return err
	}

	fmt.Println(stageConditions.DeactivateStageSteps)

	if stageConditions.DeactivateStageSteps[checkStepActiveOptions.stage][checkStepActiveOptions.step] {
		return errors.New("Step is not active")
	}

	return nil
}

func addCheckStepActiveFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&checkStepActiveOptions.stageConfigFile, "stageConfig", ".resources/piper-stage-config.yml", "Default config of piper pipeline stages")
	cmd.Flags().StringVar(&checkStepActiveOptions.step, "step", "", "")
	cmd.Flags().StringVar(&checkStepActiveOptions.stage, "stage", "", "")
}
