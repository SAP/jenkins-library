package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type newmanExecuteUtils interface {
	Glob(pattern string) (matches []string, err error)
	RunExecutable(executable string, params ...string) error
	Getenv(key string) string
}

type newmanExecuteUtilsBundle struct {
	*command.Command
	*piperutils.Files
}

func newNewmanExecuteUtils() newmanExecuteUtils {
	utils := newmanExecuteUtilsBundle{
		Command: &command.Command{
			ErrorCategoryMapping: map[string][]string{
				log.ErrorConfiguration.String(): {
					"ENOENT: no such file or directory",
				},
				log.ErrorTest.String(): {
					"AssertionError",
					"TypeError",
				},
			},
		},
		Files: &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func newmanExecute(config newmanExecuteOptions, _ *telemetry.CustomData, influx *newmanExecuteInflux) {
	utils := newNewmanExecuteUtils()

	influx.step_data.fields.newman = false
	err := runNewmanExecute(&config, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
	influx.step_data.fields.newman = true
}

func runNewmanExecute(config *newmanExecuteOptions, utils newmanExecuteUtils) error {
	if config.NewmanRunCommand != "" {
		log.Entry().Warn("found configuration for deprecated parameter newmanRunCommand, please use runOptions instead")
		log.Entry().Warn("setting runOptions to value of deprecated parameter newmanRunCommand")
		config.RunOptions = strings.Split(config.NewmanRunCommand, " ")
	}

	collectionList, err := utils.Glob(config.NewmanCollection)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return fmt.Errorf("Could not execute global search for '%v': %w", config.NewmanCollection, err)
	}

	if collectionList == nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return fmt.Errorf("no collection found with pattern '%v'", config.NewmanCollection)
	}
	log.Entry().Infof("Found the following newman collections: %v", collectionList)

	err = logVersions(utils)
	if err != nil {
		return err
	}

	err = installNewman(config.NewmanInstallCommand, utils)
	if err != nil {
		return err
	}

	// resolve environment and globals if not covered by templating
	options := resolveOptions(config)

	for _, collection := range collectionList {
		runOptions, err := resolveTemplate(config, collection)
		if err != nil {
			return err
		}

		commandSecrets := handleCfAppCredentials(config)

		runOptions = append(runOptions, options...)
		runOptions = append(runOptions, commandSecrets...)

		if !config.FailOnError {
			runOptions = append(runOptions, "--suppress-exit-code")
		}

		newmanPath := filepath.Join(utils.Getenv("HOME"), "/.npm-global/bin/newman")
		err = utils.RunExecutable(newmanPath, runOptions...)
		if err != nil {
			return fmt.Errorf("The execution of the newman tests failed, see the log for details.: %w", err)
		}
	}
	return nil
}

func logVersions(utils newmanExecuteUtils) error {
	err := utils.RunExecutable("node", "--version")
	if err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		return fmt.Errorf("error logging node version: %w", err)
	}
	err = utils.RunExecutable("npm", "--version")
	if err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		return fmt.Errorf("error logging npm version: %w", err)
	}
	return nil
}

func installNewman(newmanInstallCommand string, utils newmanExecuteUtils) error {
	installCommandTokens := strings.Split(newmanInstallCommand, " ")
	installCommandTokens = append(installCommandTokens, "--prefix=~/.npm-global")
	err := utils.RunExecutable(installCommandTokens[0], installCommandTokens[1:]...)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return fmt.Errorf("error installing newman: %w", err)
	}
	return nil
}

func resolveOptions(config *newmanExecuteOptions) []string {
	options := []string{}
	if config.NewmanEnvironment != "" && !contains(config.RunOptions, "{{.Config.NewmanEnvironment}}") {
		options = append(options, "--environment")
		options = append(options, config.NewmanEnvironment)
	}
	if config.NewmanGlobals != "" && !contains(config.RunOptions, "{{.Config.NewmanGlobals}}") {
		options = append(options, "--globals")
		options = append(options, config.NewmanGlobals)
	}
	return options
}

func resolveTemplate(config *newmanExecuteOptions, collection string) ([]string, error) {
	cmd := []string{}
	collectionDisplayName := defineCollectionDisplayName(collection)

	type TemplateConfig struct {
		Config                interface{}
		CollectionDisplayName string
		NewmanCollection      string
	}

	for _, runOption := range config.RunOptions {
		templ, err := template.New("template").Funcs(template.FuncMap{
			"getenv": func(varName string) string {
				return os.Getenv(varName)
			},
		}).Parse(runOption)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return nil, fmt.Errorf("could not parse newman command template: %w", err)
		}
		buf := new(bytes.Buffer)
		err = templ.Execute(buf, TemplateConfig{
			Config:                config,
			CollectionDisplayName: collectionDisplayName,
			NewmanCollection:      collection,
		})
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return nil, fmt.Errorf("error on executing template: %w", err)
		}
		cmd = append(cmd, buf.String())
	}

	return cmd, nil
}

func defineCollectionDisplayName(collection string) string {
	replacedSeparators := strings.Replace(collection, string(filepath.Separator), "_", -1)
	displayName := strings.Split(replacedSeparators, ".")
	if displayName[0] == "" && len(displayName) >= 2 {
		return displayName[1]
	}
	return displayName[0]
}

func handleCfAppCredentials(config *newmanExecuteOptions) []string {
	commandSecrets := []string{}
	if len(config.CfAppsWithSecrets) > 0 {
		for _, appName := range config.CfAppsWithSecrets {
			var clientID, clientSecret string
			clientID = os.Getenv("PIPER_NEWMANEXECUTE_" + appName + "_clientid")
			clientSecret = os.Getenv("PIPER_NEWMANEXECUTE_" + appName + "_clientsecret")
			if clientID != "" && clientSecret != "" {
				log.RegisterSecret(clientSecret)
				secretVar := fmt.Sprintf("--env-var %v_clientid=%v --env-var %v_clientsecret=%v", appName, clientID, appName, clientSecret)
				commandSecrets = append(commandSecrets, strings.Split(secretVar, " ")...)
				log.Entry().Infof("secrets found for app %v and forwarded to newman as --env-var parameter", appName)
			} else {
				log.Entry().Errorf("cannot fetch secrets from environment variables for app %v", appName)
			}
		}
	}
	return commandSecrets
}

func contains(slice []string, substr string) bool {
	for _, e := range slice {
		if strings.Contains(e, substr) {
			return true
		}
	}
	return false
}

func (utils newmanExecuteUtilsBundle) Getenv(key string) string {
	return os.Getenv(key)
}
