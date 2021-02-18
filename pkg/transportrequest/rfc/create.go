package rfc

import (
	"bytes"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/config/validation"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
	"io"
	"strings"
)

// Exec Everything we need for interacting with a proccess
type Exec interface {
	command.ExecRunner
	GetExitCode() int
}

// Create Provides all the functions we need for creating a transport request
type Create interface {
	WithConnection(Connection)
	WithTransportType(string)
	WithTargetSystemID(string)
	WithDescription(string)
	Perform(command Exec) (string, error)
}

// CreateAction The implementation for creating a transport request
type CreateAction struct {
	Connection     Connection
	TransportType  string
	TargetSystemID string
	Description    string
}

// WithConnection Set the connection details
func (c *CreateAction) WithConnection(con Connection) {
	c.Connection = con
}

// WithTransportType Sets the transport type
func (c *CreateAction) WithTransportType(t string) {
	c.TransportType = t
}

// WithTargetSystemID Sets the target system
func (c *CreateAction) WithTargetSystemID(t string) {
	c.TargetSystemID = t
}

// WithDescription Sets the description
func (c *CreateAction) WithDescription(d string) {
	c.Description = d
}

// Perform Creates the transport request
func (c *CreateAction) Perform(command Exec) (string, error) {

	log.Entry().Infof("Creating new transport request at '%s'", c.Connection.Endpoint)

	missingParameters, err := validation.FindEmptyStringsInConfigStruct(*c)

	if err == nil {
		notInitialized := len(missingParameters) != 0
		if notInitialized {
			err = fmt.Errorf("the following parameters are not available %s", missingParameters)
		}
	}

	var transportRequestID string

	if err == nil {

			command.SetEnv(
				[]string{
					"ABAP_DEVELOPMENT_SERVER" + "=" + c.Connection.Endpoint,
					"ABAP_DEVELOPMENT_USER" + "=" + c.Connection.User,
					"ABAP_DEVELOPMENT_PASSWORD" + "=" + c.Connection.Password,
					"TRANSPORT_DESCRIPTION" + "=" + c.Description,
					"ABAP_DEVELOPMENT_INSTANCE" + "=" + c.Connection.Instance,
					"ABAP_DEVELOPMENT_CLIENT" + "=" + c.Connection.Client,
					//VERBOSE: verbose, TODO: how to handle the verbose flag?
				},
			)

		oldStdout := command.GetStdout()
		defer func() {
			command.Stdout(oldStdout)
		}()

		var cmClientStdout bytes.Buffer
		w := io.MultiWriter(&cmClientStdout, oldStdout)
		command.Stdout(w)

		err = command.RunExecutable("cts", "createTransportRequest")

		if err == nil {
			exitCode := command.GetExitCode()

			if exitCode != 0 {
				err = fmt.Errorf("Create transport request command returned with exit code '%d'", exitCode)
			} else {
				stdoutContent := cmClientStdout.String()
				var c *gabs.Container
				c, err = gabs.ParseJSON([]byte(strings.TrimSpace(stdoutContent)))
				c = c.Path("REQUESTID")
				if c == nil {
					log.Entry().Warnf("Invalid response received: '%s'. Maybe unexpected format", stdoutContent)
				} else {
					transportRequestID = c.Data().(string)
				}

				if len(transportRequestID) == 0 {
					err = fmt.Errorf("No transport request id received.")
				}
			}
		}
	}

	if err != nil {
		log.Entry().Warnf("Cannot create transport request at '%s': %s", c.Connection.Endpoint, err.Error())
	} else {
		log.Entry().Infof("Transport request '%s' has been created at '%s'", transportRequestID, c.Connection.Endpoint)
	}

	return transportRequestID, errors.Wrapf(err, "Cannot create transport request at '%s'", c.Connection.Endpoint)
}
