package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"errors"

	"github.com/spf13/cobra"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

type checkStepActiveCommandOptions struct {
	openFile        func(s string, t map[string]string) (io.ReadCloser, error)
	fileExists      func(filename string) (bool, error)
	stageConfigFile string
	stepName        string
	stageName       string
	v1Active        bool
	stageOutputFile string
	stepOutputFile  string
}

var checkStepActiveOptions checkStepActiveCommandOptions

// CheckStepActiveCommand is the entry command for checking if a step is active in a defined stage
func CheckStepActiveCommand() *cobra.Command {
	checkStepActiveOptions.openFile = config.OpenPiperFile
	checkStepActiveOptions.fileExists = piperutils.FileExists
	var checkStepActiveCmd = &cobra.Command{
		Use:   "checkIfStepActive",
		Short: "Checks if a step is active in a defined stage.",
		PreRun: func(cmd *cobra.Command, _ []string) {
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
	// make the stageName the leading parameter
	if len(checkStepActiveOptions.stageName) == 0 && GeneralConfig.StageName != "" {
		checkStepActiveOptions.stageName = GeneralConfig.StageName
	}
	if checkStepActiveOptions.stageName == "" {
		return errors.New("stage name must not be empty")
	}
	if checkStepActiveOptions.v1Active {
		log.Entry().Warning("Please do not use --useV1 flag since it is deprecated and will be removed in future releases")
	}
	var pConfig config.Config

	// load project config and defaults
	projectConfig, err := initializeConfig(&pConfig)
	if err != nil {
		log.Entry().Errorf("Failed to load project config: %v", err)
		return fmt.Errorf("Failed to load project config failed: %w", err)
	}

	stageConfigFile, err := checkStepActiveOptions.openFile(checkStepActiveOptions.stageConfigFile, GeneralConfig.GitHubAccessTokens)
	if err != nil {
		return fmt.Errorf("config: open stage configuration file '%v' failed: %w", checkStepActiveOptions.stageConfigFile, err)
	}
	defer stageConfigFile.Close()

	// load and evaluate step conditions
	runConfig := config.RunConfig{StageConfigFile: stageConfigFile}
	runConfigV1 := &config.RunConfigV1{RunConfig: runConfig}
	err = runConfigV1.InitRunConfigV1(projectConfig, utils, GeneralConfig.EnvRootPath)
	if err != nil {
		return err
	}
	runSteps := runConfigV1.RunSteps
	runStages := runConfigV1.RunStages

	log.Entry().Debugf("RunSteps: %v", runSteps)
	log.Entry().Debugf("RunStages: %v", runStages)

	if len(checkStepActiveOptions.stageOutputFile) > 0 || len(checkStepActiveOptions.stepOutputFile) > 0 {
		if len(checkStepActiveOptions.stageOutputFile) > 0 {
			result, err := json.Marshal(runStages)
			if err != nil {
				return fmt.Errorf("error marshalling json: %w", err)
			}
			log.Entry().Infof("Writing stage condition file %v", checkStepActiveOptions.stageOutputFile)
			err = utils.FileWrite(checkStepActiveOptions.stageOutputFile, result, 0666)
			if err != nil {
				return fmt.Errorf("error writing file '%v': %w", checkStepActiveOptions.stageOutputFile, err)
			}
		}

		if len(checkStepActiveOptions.stepOutputFile) > 0 {
			result, err := json.Marshal(runSteps)
			if err != nil {
				return fmt.Errorf("error marshalling json: %w", err)
			}
			log.Entry().Infof("Writing step condition file %v", checkStepActiveOptions.stepOutputFile)
			err = utils.FileWrite(checkStepActiveOptions.stepOutputFile, result, 0666)
			if err != nil {
				return fmt.Errorf("error writing file '%v': %w", checkStepActiveOptions.stepOutputFile, err)
			}
		}

		// do not perform a check if output files are written
		return nil
	}

	if !runSteps[checkStepActiveOptions.stageName][checkStepActiveOptions.stepName] {
		return fmt.Errorf("Step %s in stage %s is not active", checkStepActiveOptions.stepName, checkStepActiveOptions.stageName)
	}
	log.Entry().Infof("Step %s in stage %s is active", checkStepActiveOptions.stepName, checkStepActiveOptions.stageName)

	return nil
}

func addCheckStepActiveFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&checkStepActiveOptions.stageConfigFile, "stageConfig", ".resources/piper-stage-config.yml",
		"Default config of piper pipeline stages")
	cmd.Flags().StringVar(&checkStepActiveOptions.stepName, "step", "", "Name of the step being checked")
	cmd.Flags().StringVar(&checkStepActiveOptions.stageName, "stage", "", "Name of the stage in which contains the step being checked")
	cmd.Flags().BoolVar(&checkStepActiveOptions.v1Active, "useV1", false, "Use new CRD-style stage configuration (deprecated)")
	cmd.Flags().StringVar(&checkStepActiveOptions.stageOutputFile, "stageOutputFile", "", "Defines a file path. If set, the stage output will be written to the defined file")
	cmd.Flags().StringVar(&checkStepActiveOptions.stepOutputFile, "stepOutputFile", "", "Defines a file path. If set, the step output will be written to the defined file")
	_ = cmd.MarkFlagRequired("step")
}

func initializeConfig(pConfig *config.Config) (*config.Config, error) {
	projectConfigFile := getProjectConfigFile(GeneralConfig.CustomConfig)
	var customConfig io.ReadCloser
	var err error
	//accept that config file cannot be loaded as its not mandatory here
	if exists, err := checkStepActiveOptions.fileExists(projectConfigFile); exists {
		log.Entry().Infof("Project config: '%s'", projectConfigFile)
		customConfig, err = checkStepActiveOptions.openFile(projectConfigFile, GeneralConfig.GitHubAccessTokens)
		if err != nil {
			return nil, fmt.Errorf("config: open configuration file '%v' failed: %w", projectConfigFile, err)
		}
		defer customConfig.Close()
	} else {
		log.Entry().Infof("Project config: NONE ('%s' does not exist)", projectConfigFile)
	}

	defaultConfig := []io.ReadCloser{}
	for _, f := range GeneralConfig.DefaultConfig {
		fc, err := checkStepActiveOptions.openFile(f, GeneralConfig.GitHubAccessTokens)
		// only create error for non-default values
		if err != nil && f != ".pipeline/defaults.yaml" {
			return nil, fmt.Errorf("config: getting defaults failed: '%v': %w", f, err)
		}
		if err == nil {
			defaultConfig = append(defaultConfig, fc)
		}
	}
	var flags map[string]interface{}
	filter := config.StepFilters{
		All:     []string{},
		General: []string{},
		Stages:  []string{},
		Steps:   []string{},
		Env:     []string{},
	}

	_, err = pConfig.GetStepConfig(flags, "", customConfig, defaultConfig, GeneralConfig.IgnoreCustomDefaults, filter, config.StepData{}, nil, "", "")
	if err != nil {
		return nil, fmt.Errorf("getting step config failed: %w", err)
	}
	return pConfig, nil
}
