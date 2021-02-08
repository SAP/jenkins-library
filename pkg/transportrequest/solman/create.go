package solman

import (
	"bytes"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/config/validation"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
	"io"
	"strings"
)

// CreateAction Collects all the properties we need for creating a transport request
type CreateAction struct {
	Connection          Connection
	ChangeDocumentID    string
	DevelopmentSystemID string
	CMOpts              []string
}

// Create collects everything which is needed for creating a transport request
type Create interface {
	WithConnection(Connection)
	WithChangeDocumentID(string)
	WithDevelopmentSystemID(string)
	WithCMOpts([]string)
	Perform(command Exec) (string, error)
}

// WithConnection specifies all the connection details which
// are required in order to connect so SOLMAN
func (a *CreateAction) WithConnection(c Connection) {
	a.Connection = c
}

// WithChangeDocumentID specifies the change document for that
// the transport request is created.
func (a *CreateAction) WithChangeDocumentID(id string) {
	a.ChangeDocumentID = id
}

// WithDevelopmentSystemID specifies the development system ID.
func (a *CreateAction) WithDevelopmentSystemID(id string) {
	a.DevelopmentSystemID = id
}

// WithCMOpts sets additional options for calling the
// cm client tool. E.g. -D options. Useful for troubleshooting
func (a *CreateAction) WithCMOpts(opts []string) {
	a.CMOpts = opts
}

// Perform creates a new transport request
func (a *CreateAction) Perform(command Exec) (string, error) {

	log.Entry().Infof("Creating new transport request via '%s'.", a.Connection.Endpoint)

	missingParameters, err := validation.FindEmptyStringsInConfigStruct(*a)

	if err == nil {
		notInitialized := len(missingParameters) != 0
		if notInitialized {
			err = fmt.Errorf("the following parameters are not available %s", missingParameters)
		}
	}

	var transportRequestID string

	if err == nil {
		if len(a.CMOpts) > 0 {
			command.SetEnv([]string{fmt.Sprintf("CMCLIENT_OPTS=%s", strings.Join(a.CMOpts, " "))})
		}

		oldStdout := command.GetStdout()
		defer func() {
			command.Stdout(oldStdout)
		}()

		var cmClientStdout bytes.Buffer
		w := io.MultiWriter(&cmClientStdout, oldStdout)
		command.Stdout(w)

		err = command.RunExecutable("cmclient",
			"--endpoint", a.Connection.Endpoint,
			"--user", a.Connection.User,
			"--password", a.Connection.Password,
			"--backend-type", "SOLMAN",
			"create-transport",
			"-cID", a.ChangeDocumentID,
			"-dID", a.DevelopmentSystemID,
		)

		if err == nil {
			exitCode := command.GetExitCode()

			if exitCode != 0 {
				err = fmt.Errorf("Create transport request command returned with exit code '%d'", exitCode)
			} else {
				transportRequestID = strings.TrimSpace(cmClientStdout.String())
			}
		}
	}

	if err == nil {
		log.Entry().Infof("Created transport request '%s' at '%s'.", transportRequestID, a.Connection.Endpoint)
	} else {
		log.Entry().WithError(err).Warnf("Creating transport request at '%s' failed",
			a.Connection.Endpoint)
	}

	return transportRequestID, errors.Wrap(err, "cannot create transport request")
}
