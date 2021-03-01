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
	"github.com/pkg/errors"
)

type newmanExecuteUtils interface {
	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The newmanExecuteUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
	Glob(pattern string) (matches []string, err error)

	RunExecutable(executable string, params ...string) error
	RunShell(shell, script string) error
	SetEnv(env []string)
	AppendEnv(env []string)
}

type newmanExecuteUtilsBundle struct {
	*command.Command
	*piperutils.Files

	// Embed more structs as necessary to implement methods or interfaces you add to newmanExecuteUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// newmanExecuteUtilsBundle and forward to the implementation of the dependency.
}

func newNewmanExecuteUtils() newmanExecuteUtils {
	utils := newmanExecuteUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func newmanExecute(config newmanExecuteOptions, _ *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newNewmanExecuteUtils()

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runNewmanExecute(&config, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
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
		return errors.Wrapf(err, "Could not execute global search for '%v'", config.NewmanCollection)
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

	// append environment and globals if not resolved by templating
	options := []string{}
	if config.NewmanEnvironment != "" && !contains(config.RunOptions, "{{.Config.NewmanEnvironment}}") {
		options = append(options, "--environment")
		options = append(options, config.NewmanEnvironment)
	}
	if config.NewmanGlobals != "" && !contains(config.RunOptions, "{{.Config.NewmanGlobals}}") {
		options = append(options, "--globals")
		options = append(options, config.NewmanGlobals)
	}

	for _, collection := range collectionList {
		runOptions := []string{}
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

		newmanPath := filepath.Join(os.Getenv("HOME"), "/.npm-global/bin/newman")
		err = utils.RunExecutable(newmanPath, runOptions...)
		if err != nil {
			log.SetErrorCategory(log.ErrorService)
			return errors.Wrap(err, "The execution of the newman tests failed, see the log for details.")
		}
	}
	return nil
}

func logVersions(utils newmanExecuteUtils) error {
	err := utils.RunExecutable("node", "--version")
	if err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		return errors.Wrap(err, "error logging node version")
	}
	err = utils.RunExecutable("npm", "--version")
	if err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		return errors.Wrap(err, "error logging npm version")
	}
	return nil
}

func installNewman(newmanInstallCommand string, utils newmanExecuteUtils) error {
	installCommandTokens := strings.Split(newmanInstallCommand, " ")
	installCommandTokens = append(installCommandTokens, "--prefix=~/.npm-global")
	err := utils.RunExecutable(installCommandTokens[0], installCommandTokens[1:]...)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Wrap(err, "error installing newman")
	}
	return nil
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
		templ, err := template.New("template").Parse(runOption)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return nil, errors.Wrap(err, "could not parse newman command template")
		}
		buf := new(bytes.Buffer)
		err = templ.Execute(buf, TemplateConfig{
			Config:                config,
			CollectionDisplayName: collectionDisplayName,
			NewmanCollection:      collection,
		})
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return nil, errors.Wrap(err, "error on executing template")
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
				commandSecrets = append(commandSecrets, secretVar)
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
