package cmd

import (
	"bytes"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
	"path/filepath"
	"strings"
	"text/template"
)

type executeNewmanUtils interface {
	Glob(pattern string) (matches []string, err error)

	RunShell(shell, script string) error
	RunExecutable(executable string, params ...string) error
}

type executeNewmanUtilsBundle struct {
	*command.Command
	*piperutils.Files

	// Embed more structs as necessary to implement methods or interfaces you add to executeNewmanUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// executeNewmanUtilsBundle and forward to the implementation of the dependency.
}

func newExecuteNewmanUtils() executeNewmanUtils {
	utils := executeNewmanUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func executeNewman(config executeNewmanOptions, _ *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newExecuteNewmanUtils()

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runExecuteNewman(&config, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runExecuteNewman(config *executeNewmanOptions, utils executeNewmanUtils) error {

	collectionList, err := utils.Glob(config.NewmanCollection)
	if err != nil {
		return errors.Wrapf(err, "Could not execute global search for '%v'", config.NewmanCollection)
	}

	if collectionList == nil {
		return fmt.Errorf("no collection found with pattern '%v'", config.NewmanCollection)
	} else {
		log.Entry().Infof("Found files '%v'", collectionList)
	}

	err = logVersions(utils)
	// TODO: should error in version logging cause failure?
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
		log.Entry().Debug(cmd) // TODO continue developing
	}

	return nil
}

func logVersions(utils executeNewmanUtils) error {
	//utils.SetDir(".") // TODO: Need this?
	//returnStatus: true // TODO: How to do this? If necessary at all.
	err := utils.RunExecutable("node", "--version")
	if err != nil {
		return errors.Wrap(err, "error installing newman")
	}

	//utils.SetDir(".") // TODO: Need this?
	//returnStatus: true // TODO: How to do this? If necessary at all.
	err = utils.RunExecutable("npm", "--version")
	if err != nil {
		return errors.Wrap(err, "error installing newman")
	}

	return nil
}

func installNewman(newmanInstallCommand string, utils executeNewmanUtils) error {
	args := []string{"NPM_CONFIG_PREFIX=~/.npm-global", newmanInstallCommand}
	script := strings.Join(args, " ")
	//utils.SetDir(".") // TODO: Need this?
	err := utils.RunShell("/bin/sh", script)
	if err != nil {
		return errors.Wrap(err, "error installing newman")
	}
	return nil
}

func resolveTemplate(config *executeNewmanOptions, collection string) (string, error) {
	collectionDisplayName := defineCollectionDisplayName(collection)

	type TemplateConfig struct {
		Config                interface{}
		CollectionDisplayName string
		// TODO: New field as structs cannot be extended in Go
		NewmanCollection string
	}

	templ, err := template.New("template").Parse(config.NewmanRunCommand)
	if err != nil {
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
		return "", errors.Wrap(err, "error on executing template")
	}
	cmd := buf.String()
	return cmd, nil
}

func defineCollectionDisplayName(collection string) string {
	replacedSeparators := strings.Replace(collection, string(filepath.Separator), "_", -1)
	return strings.Split(replacedSeparators, ".")[0]
}
