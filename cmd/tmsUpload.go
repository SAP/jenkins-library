package cmd

import (
	"encoding/json"

	"github.com/SAP/jenkins-library/pkg/command"
	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/tms"
)

type uaa struct {
	url          string
	clinetid     string
	clientsecret string
}

type serviceKey struct {
	uaa uaa
}

type tmsUploadUtils interface {
	command.ExecRunner

	FileExists(filename string) (bool, error)

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The tmsUploadUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type tmsUploadUtilsBundle struct {
	*command.Command
	*piperutils.Files

	// Embed more structs as necessary to implement methods or interfaces you add to tmsUploadUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// tmsUploadUtilsBundle and forward to the implementation of the dependency.
}

func newTmsUploadUtils() tmsUploadUtils {
	utils := tmsUploadUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func tmsUpload(config tmsUploadOptions, telemetryData *telemetry.CustomData, influx *tmsUploadInflux) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	// TODO: do we need it?
	/*
		utils := newTmsUploadUtils()
	*/

	client := &piperHttp.Client{}

	// TODO: any options to set for the client? (see checkmarxExecuteScan.go)

	// TODO: how to read credentials by their id from jenkins?
	var serviceKey serviceKey
	json.Unmarshal([]byte(config.ServiceKey), &serviceKey)

	communicationInstance, err := tms.NewCommunicationInstance(client, serviceKey.uaa.url, serviceKey.uaa.clinetid, serviceKey.uaa.clientsecret, GeneralConfig.Verbose)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Failed to prepare client for talking with TMS")
	}

	// TODO: understand, what does this influx part do
	influx.step_data.fields.tms = false

	// TODO: understand, why the variable is passed with asterisk
	if err := runTmsUpload(config, *communicationInstance); err != nil {
		log.Entry().WithError(err).Fatal("Failed to run tmsUpload step")
	}
	influx.step_data.fields.tms = true
}

func runTmsUpload(config tmsUploadOptions, sys tms.CommunicationInstance) error {
	/*
		log.Entry().WithField("LogField", "Log field content").Info("This is just a demo for a simple step.")

		// Example of calling methods from external dependencies directly on utils:
		exists, err := utils.FileExists("file.txt")
		if err != nil {
			// It is good practice to set an error category.
			// Most likely you want to do this at the place where enough context is known.
			log.SetErrorCategory(log.ErrorConfiguration)
			// Always wrap non-descriptive errors to enrich them with context for when they appear in the log:
			return fmt.Errorf("failed to check for important file: %w", err)
		}
		if !exists {
			log.SetErrorCategory(log.ErrorConfiguration)
			return fmt.Errorf("cannot run without important file")
		}
	*/

	return nil
}
