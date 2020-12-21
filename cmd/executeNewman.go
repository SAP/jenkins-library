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

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The executeNewmanUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
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

func executeNewman(config executeNewmanOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newExecuteNewmanUtils()

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runExecuteNewman(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runExecuteNewman(config *executeNewmanOptions, telemetryData *telemetry.CustomData, utils executeNewmanUtils) error {

	collectionList, err := utils.Glob(config.NewmanCollection)
	if err != nil {
		return errors.Wrapf(err, "Could not execute global search for '%v'", config.NewmanCollection)
	}

	if collectionList == nil {
		return fmt.Errorf("no collection found with pattern '%v'", config.NewmanCollection)
	} else {
		log.Entry().Infof("Found files '%v'", collectionList)
	}

	for _, collection := range collectionList {
		collectionDisplayName := defineCollectionDisplayName(collection)

		cmd, err := resolveTemplate(config, collectionDisplayName)
		if err != nil {
			return err
		}

		log.Entry().Debug(cmd)

	}

	return nil
}

func resolveTemplate(config *executeNewmanOptions, collectionDisplayName string) (string, error) {
	type TemplateConfig struct {
		Config                string
		CollectionDisplayName string
	}

	templ, err := template.New("template").Parse(config.NewmanRunCommand)
	if err != nil {
		return "", errors.Wrap(err, "could not parse newman command template")
	}
	buf := new(bytes.Buffer)
	// TODO: Config and CollectionDisplayName must be capitalized <-> was small letter in groovy --> Templates must be adapted
	err = templ.Execute(buf, TemplateConfig{
		Config:                "", //config.plus([newmanCollection:collection]) // TODO
		CollectionDisplayName: collectionDisplayName,
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
