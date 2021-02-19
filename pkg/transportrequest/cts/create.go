package cts

import (
	"bytes"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/config/validation"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
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
	WithCMOpts([]string)
	Perform(command Exec) (string, error)
}

// CreateAction The implementation for creating a transport request
type CreateAction struct {
	Connection     Connection
	TransportType  string
	TargetSystemID string
	Description    string
	CMOpts         []string
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

// WithCMOpts sets additional options for calling the
// cm client tool. E.g. -D options. Useful for troubleshooting
func (c *CreateAction) WithCMOpts(opts []string) {
	c.CMOpts = opts
}

// Perform Creates the transport request
func (c *CreateAction) Perform(command Exec) (string, error) {

	log.Entry().Infof("Creating new transport request at '%s'", c.Connection.Endpoint)

	missingParameters, err := validation.FindEmptyStringsInConfigStruct(*c)

	if err == nil {
		// Connection.Client is not used in this case
		missingParameters, _ := piperutils.RemoveAll(missingParameters, "Connection.Client")
		notInitialized := len(missingParameters) != 0
		if notInitialized {
			err = fmt.Errorf("the following parameters are not available %s", missingParameters)
		}
	}

	var transportRequestID string

	if err == nil {
		if len(c.CMOpts) > 0 {
			command.SetEnv([]string{fmt.Sprintf("CMCLIENT_OPTS=%s", strings.Join(c.CMOpts, " "))})
		}

		oldStdout := command.GetStdout()
		defer func() {
			command.Stdout(oldStdout)
		}()

		var cmClientStdout bytes.Buffer
		w := io.MultiWriter(&cmClientStdout, oldStdout)
		command.Stdout(w)

		err = command.RunExecutable("cmclient",
			"--endpoint", c.Connection.Endpoint,
			"--user", c.Connection.User,
			"--password", c.Connection.Password,
			"-t", "CTS",
			"create-transport",
			"-tt", c.TransportType,
			"-ts", c.TargetSystemID,
			"-d", c.Description,
		)

		if err == nil {
			exitCode := command.GetExitCode()

			if exitCode != 0 {
				err = fmt.Errorf("Create transport request command returned with exit code '%d'", exitCode)
			} else {
				transportRequestID = strings.TrimSpace(cmClientStdout.String())
				if len(transportRequestID) == 0 {
					err = fmt.Errorf("No transport request id received.")
				}
			}
		}
	}

	if err == nil {
		log.Entry().Infof("Transport request '%s' has been created at '%s'", transportRequestID, c.Connection.Endpoint)
	} else {
		log.Entry().Warnf("Cannot create transport request at '%s': %s", c.Connection.Endpoint, err.Error())
	}

	return transportRequestID, errors.Wrapf(err, "cannot create transport request at '%s'", c.Connection.Endpoint)
}
