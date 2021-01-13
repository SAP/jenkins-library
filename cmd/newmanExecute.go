package cmd

import (
	"bytes"
	"fmt"
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

	RunShell(shell, script string) error
	RunExecutable(executable string, params ...string) error
	SetEnv(env []string)
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
	collectionList, err := utils.Glob(config.NewmanCollection)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Wrapf(err, "Could not execute global search for '%v'", config.NewmanCollection)
	}

	if collectionList == nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return fmt.Errorf("no collection found with pattern '%v'", config.NewmanCollection)
	} else {
		log.Entry().Infof("Found files '%v'", collectionList)
	}

	err = logVersions(utils)
	if err != nil {
		return err
	}

	err = installNewman(config.NewmanInstallCommand, utils)
	if err != nil {
		return err
	}

	for _, collection := range collectionList {
		cmd, err := resolveTemplate(config, collection)
		if err != nil {
			return err
		}

		commandSecrets := ""
		hasSecrets := len(config.CfAppsWithSecrets) > 0
		if hasSecrets {
			//	CloudFoundry cfUtils = new CloudFoundry(script); // TODO: ???
			for _, appName := range config.CfAppsWithSecrets {
				var clientId, clientSecret string
				// def xsuaaCredentials = cfUtils.getXsuaaCredentials(config.cloudFoundry.apiEndpoint, // TODO: ???
				// config.cloudFoundry.org,
				// config.cloudFoundry.space,
				// config.cloudFoundry.credentialsId,
				// appName,
				// config.verbose ? true : false ) //to avoid config.verbose as "null" if undefined in yaml and since function parameter boolean

				commandSecrets += " --env-var " + appName + "_clientid=" + clientId + " --env-var " + appName + "_clientsecret=" + clientSecret
				// TODO: How to do echo in golang?
				// echo "Exposing client id and secret for ${appName}: as ${appName}_clientid and ${appName}_clientsecret to newman"
			}
		}

		if !config.FailOnError {
			cmd += " --suppress-exit-code"
		}

		if hasSecrets {
			log.Entry().Info("PATH=$PATH:~/.npm-global/bin newman " + cmd + " **env/secrets**")

			//utils.SetDir(".") // TODO: Need this?
			err := utils.RunShell("/bin/sh", "set +x") // TODO: Does this help for later call?
			if err != nil {
				log.SetErrorCategory(log.ErrorService)
				//TODO: other message here
				return errors.Wrap(err, "The execution of the newman tests failed, see the log for details.")
			}
		}

		installCommandTokens := strings.Split(cmd, " ")
		//utils.SetEnv([]string{"NPM_CONFIG_PREFIX=/home/node/.npm-global",
		//	"PATH=$PATH:.\\node_modules\\.bin"})
		err = utils.RunExecutable(installCommandTokens[0], installCommandTokens[1:]...)

		if err != nil {
			log.SetErrorCategory(log.ErrorService)
			return errors.Wrap(err, "The execution of the newman tests failed, see the log for details.")
		}
	}

	return nil
}

func logVersions(utils newmanExecuteUtils) error {
	//utils.SetDir(".") // TODO: Need this?
	err := utils.RunExecutable("node", "--version")
	if err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		return errors.Wrap(err, "error logging node version")
	}

	//utils.SetDir(".") // TODO: Need this?
	err = utils.RunExecutable("npm", "--version")
	if err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		return errors.Wrap(err, "error logging npm version")
	}

	return nil
}

func installNewman(newmanInstallCommand string, utils newmanExecuteUtils) error {
	//args := []string{ /*"NPM_CONFIG_PREFIX=~/.npm-global", */ newmanInstallCommand}
	//script := strings.Join(args, " ")
	////utils.SetDir(".") // TODO: Need this?
	//err := utils.RunExecutable("npm", script)

	installCommandTokens := strings.Split(newmanInstallCommand, " ")
	//utils.SetEnv([]string{"NPM_CONFIG_PREFIX=/home/node/.npm-global"})
	err := utils.RunExecutable(installCommandTokens[0], installCommandTokens[1:]...)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Wrap(err, "error installing newman")
	}
	return nil
}

func resolveTemplate(config *newmanExecuteOptions, collection string) (string, error) {
	collectionDisplayName := defineCollectionDisplayName(collection)

	type TemplateConfig struct {
		Config                interface{}
		CollectionDisplayName string
		NewmanCollection      string
		// TODO: New field as structs cannot be extended in Go
	}

	templ, err := template.New("template").Parse(config.NewmanRunCommand)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return "", errors.Wrap(err, "could not parse newman command template")
	}
	buf := new(bytes.Buffer)
	// TODO: Config and CollectionDisplayName must be capitalized <-> was small letter in groovy --> Templates must be adapted
	err = templ.Execute(buf, TemplateConfig{
		Config:                config,
		CollectionDisplayName: collectionDisplayName,
		NewmanCollection:      collection,
	})
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return "", errors.Wrap(err, "error on executing template")
	}
	cmd := buf.String()
	return cmd, nil
}

func defineCollectionDisplayName(collection string) string {
	replacedSeparators := strings.Replace(collection, string(filepath.Separator), "_", -1)
	return strings.Split(replacedSeparators, ".")[0]
}
