package solman

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
)

// FileSystem interface collecting everything which is file system
// related and needed in the context of a SOLMAN upload.
type FileSystem interface {
	FileExists(path string) (bool, error)
}

// Exec interface collecting everything which is execution related
// and needed in the context of a SOLMAN upload.
type Exec interface {
	command.ExecRunner
	GetExitCode() int
}

// Connection Everything wee need for connecting to CTS
type Connection struct {
	Endpoint string
	User     string
	Password string
}

// UploadAction Collects all the properties we need for the deployment
type UploadAction struct {
	Connection         Connection
	ChangeDocumentID   string
	TransportRequestID string
	ApplicationID      string
	File               string
	CMOpts             []string
}

// Action collects everything which is needed to perform a SOLMAN upload
type Action interface {
	WithConnection(Connection)
	WithChangeDocumentID(string)
	WithTransportRequestID(string)
	WithApplicationID(string)
	WithFile(string)
	WithCMOpts([]string)
	Perform(fs FileSystem, command Exec) error
}

// WithConnection specifies all the connection details which
// are required in order to connect so SOLMAN
func (a *UploadAction) WithConnection(c Connection) {
	a.Connection = c
}

// WithChangeDocumentID specifies the change document which
// receives the executable.
func (a *UploadAction) WithChangeDocumentID(id string) {
	a.ChangeDocumentID = id
}

// WithTransportRequestID specifies the transport request which
// receives the executable.
func (a *UploadAction) WithTransportRequestID(id string) {
	a.TransportRequestID = id
}

// WithApplicationID specifies the application ID.
func (a *UploadAction) WithApplicationID(id string) {
	a.ApplicationID = id
}

// WithFile specifies the executable which should be uploaded into
// a transport on SOLMAN
func (a *UploadAction) WithFile(f string) {
	a.File = f
}

// WithCMOpts sets additional options for calling the
// cm client tool. E.g. -D options. Useful for troubleshooting
func (a *UploadAction) WithCMOpts(opts []string) {
	a.CMOpts = opts
}

// Perform performs the SOLMAN upload
func (a *UploadAction) Perform(fs FileSystem, command Exec) error {

	log.Entry().Infof("Deploying artifact '%s' to '%s'.",
		a.File, a.Connection.Endpoint)

	exists, err := fs.FileExists(a.File)

	if !exists {
		err = fmt.Errorf("file '%s' does not exist", a.File)
	}

	if err == nil {
		command.SetEnv(a.CMOpts)

		err = command.RunExecutable("cmclient",
			"--endpoint", a.Connection.Endpoint,
			"--user", a.Connection.User,
			"--password", a.Connection.Password,
			"--backend-type", "SOLMAN",
			"upload-file-to-transport",
			"-cID", a.ChangeDocumentID,
			"-tID", a.TransportRequestID,
			a.ApplicationID, a.File)

		exitCode := command.GetExitCode()

		if exitCode != 0 {
			err = fmt.Errorf("Upload command returned with exit code '%d'", exitCode)
		}
	}

	if err == nil {
		log.Entry().Infof("Deployment succeeded, artifact: '%s', endpoint: '%s'",
			a.File, a.Connection.Endpoint)
	} else {
		log.Entry().WithError(err).Warnf("Deployment failed, artifact: '%s', endpoint: '%s'",
			a.File, a.Connection.Endpoint)
	}

	return errors.Wrapf(err, "cannot upload artifact '%s'", a.File)
}
