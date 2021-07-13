package rfc

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/config/validation"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
	"strconv"
)

const (
	eq = "="
)

// Exec Everything we need for calling an executable
type Exec interface {
	command.ExecRunner
	GetExitCode() int
}

// Upload ...
type Upload interface {
	Perform(Exec) error
	WithConnection(Connection)
	WithConfiguration(UploadConfig)
	WithApplication(Application)
	WithTransportRequestID(string)
	WithApplicationURL(string)
}

// Connection Everything we need for connecting to the ABAP system
type Connection struct {
	// The endpoint in for form <protocol>://<host>:<port>, no path
	Endpoint string
	// The ABAP client, like e.g. "001"
	Client string
	// The ABAP instance, like e.g. "DEV",  "QA"
	Instance string
	User     string
	Password string
}

// Application Application specific properties
type Application struct {
	Name        string
	Description string
	AbapPackage string
}

// UploadConfig additional configuration properties
type UploadConfig struct {
	AcceptUnixStyleEndOfLine bool
	CodePage                 string
	FailUploadOnWarning      bool
	Verbose                  bool
}

// UploadAction Collects all the properties we need for the deployment
type UploadAction struct {
	Connection         Connection
	Application        Application
	Configuration      UploadConfig
	TransportRequestID string
	ApplicationURL     string
}

// WithApplicationURL The location of the deployable
func (action *UploadAction) WithApplicationURL(z string) {
	action.ApplicationURL = z
}

// WithTransportRequestID The transport request ID for the upload
func (action *UploadAction) WithTransportRequestID(t string) {
	action.TransportRequestID = t
}

// WithApplication Everything we need to know about the application
func (action *UploadAction) WithApplication(a Application) {
	action.Application = a
}

// WithConfiguration Everything we need to know in order to perform the upload
func (action *UploadAction) WithConfiguration(c UploadConfig) {
	action.Configuration = c
}

// WithConnection Everything we need to know about the connection
func (action *UploadAction) WithConnection(c Connection) {
	action.Connection = c
}

// Perform Performs the upload
func (action *UploadAction) Perform(command Exec) error {

	log.Entry().Infof("Deploying artifact '%s' to '%s', client: '%s', instance: '%s'.",
		action.ApplicationURL, action.Connection.Endpoint, action.Connection.Client, action.Connection.Instance,
	)

	parametersWithMissingValues, err := validation.FindEmptyStringsInConfigStruct(*action)
	if err != nil {
		return fmt.Errorf("invalid configuration parameters detected. SOLMAN upload parameter may be missing : %w", err)
	}
	if len(parametersWithMissingValues) != 0 {
		return fmt.Errorf("cannot perform artifact upload. The following parameters are not available %s", parametersWithMissingValues)
	}

	command.SetEnv([]string{
		"ABAP_DEVELOPMENT_SERVER" + eq + action.Connection.Endpoint,
		"ABAP_DEVELOPMENT_USER" + eq + action.Connection.User,
		"ABAP_DEVELOPMENT_PASSWORD" + eq + action.Connection.Password,
		"ABAP_DEVELOPMENT_INSTANCE" + eq + action.Connection.Instance,
		"ABAP_DEVELOPMENT_CLIENT" + eq + action.Connection.Client,
		"ABAP_APPLICATION_NAME" + eq + action.Application.Name,
		"ABAP_APPLICATION_DESC" + eq + action.Application.Description,
		"ABAP_PACKAGE" + eq + action.Application.AbapPackage,
		"ZIP_FILE_URL" + eq + action.ApplicationURL,
		"CODE_PAGE" + eq + action.Configuration.CodePage,
		"ABAP_ACCEPT_UNIX_STYLE_EOL" + eq + toAbapBool(action.Configuration.AcceptUnixStyleEndOfLine),
		"FAIL_UPLOAD_ON_WARNING" + eq + strconv.FormatBool(action.Configuration.FailUploadOnWarning),
		"VERBOSE" + eq + strconv.FormatBool(action.Configuration.Verbose),
	})

	err = command.RunExecutable("cts", fmt.Sprintf("uploadToABAP:%s", action.TransportRequestID))

	exitCode := command.GetExitCode()
	if exitCode != 0 {
		message := fmt.Sprintf("upload command returned with exit code '%d'", exitCode)
		if err != nil {
			// Using the wrapping here is to some extend an abuse, since it is not really
			// error chaining (the other error is not necessaryly a "predecessor" of this one).
			// But it is a pragmatic approach for not loosing information for trouble shooting. There
			// is no possibility to have something like suppressed errors.
			err = errors.Wrap(err, message)
		} else {
			err = errors.New(message)
		}
	}

	if err == nil {
		log.Entry().Infof("Deploying artifact '%s' to '%s', client: '%s', instance: '%s' succeeded.",
			action.ApplicationURL, action.Connection.Endpoint, action.Connection.Client, action.Connection.Instance,
		)
	} else {
		log.Entry().Warnf("Deploying artifact '%s' to '%s', client: '%s', instance: '%s' failed.",
			action.ApplicationURL, action.Connection.Endpoint, action.Connection.Client, action.Connection.Instance,
		)
	}
	return errors.Wrap(err, "cannot upload artifact")
}

func toAbapBool(b bool) string {
	abapBool := "-"
	if b {
		abapBool = "X"
	}
	return abapBool
}
