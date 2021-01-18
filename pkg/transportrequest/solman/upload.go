package solman

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"reflect"
	"strings"
)

type FileSystem interface {
	FileExists(path string) (bool, error)
}

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
	ChangeDocumentId   string
	TransportRequestId string
	ApplicationID      string
	File               string
	CMOpts             []string
}

type Action interface {
	WithConnection(Connection)
	WithChangeDocumentId(string)
	WithTransportRequestId(string)
	WithApplicationID(string)
	WithFile(string)
	WithCMOpts([]string)
	Perform(fs FileSystem, command Exec) error
}

func (a *UploadAction) WithConnection(c Connection) {
	a.Connection = c
}
func (a *UploadAction) WithChangeDocumentId(id string) {
	a.ChangeDocumentId = id
}
func (a *UploadAction) WithTransportRequestId(id string) {
	a.TransportRequestId = id
}
func (a *UploadAction) WithApplicationID(id string) {
	a.ApplicationID = id
}
func (a *UploadAction) WithFile(f string) {
	a.File = f
}
func (a *UploadAction) WithCMOpts(opts []string) {
	a.CMOpts = opts
}

func (a *UploadAction) Perform(fs FileSystem, command Exec) error {

	missingParameters, err := FindEmptyStrings(*a)
	if err != nil {
		return fmt.Errorf("Cannot check that all required parameters are available for SOLMAN upload: %w", err)
	}
	notInitialized := len(missingParameters) != 0
	if notInitialized {
		return fmt.Errorf("Cannot perform artifact upload. The following parameters are not available %s", missingParameters)
	}

	exists, err := fs.FileExists(a.File)
	if err != nil {
		return fmt.Errorf("Cannot upload file: %w", err)
	}
	if !exists {
		return fmt.Errorf("File '%s' does not exist.", a.File)
	}

	command.SetEnv(a.CMOpts)

	err = command.RunExecutable("cmclient",
		"--endpoint", a.Connection.Endpoint,
		"--user", a.Connection.User,
		"--password", a.Connection.Password,
		"--backend-type", "SOLMAN",
		"upload-file-to-transport",
		"-cID", a.ChangeDocumentId,
		"-tID", a.TransportRequestId,
		a.ApplicationID, a.File)

	if err != nil {
		err = fmt.Errorf("Cannot upload '%s': %w", a.File, err)
	}

	exitCode := command.GetExitCode()

	if exitCode != 0 {
		err = fmt.Errorf("Cannot upload '%s': Upload command returned with exit code '%d'", a.File, exitCode)
	}
	return err
}

// the next functions should be in some kind of helper ...

// FindEmptyStrings finds empty strings in a struct.
// in case the stuct contains another struct, also this struct is checked.
func FindEmptyStrings(v interface{}) ([]string, error) {
	emptyStrings := []string{}
	if reflect.ValueOf(v).Kind() != reflect.Struct {
		return emptyStrings, fmt.Errorf("%v (%T) was not a stuct", v, v)
	}
	findEmptyStringsInternal(v, &emptyStrings, []string{})
	return emptyStrings, nil
}

func findEmptyStringsInternal(v interface{}, emptyStrings *[]string, prefix []string) (bool, error) {
	fields := reflect.TypeOf(v)
	values := reflect.ValueOf(v)
	for i := 0; i < fields.NumField(); i++ {
		switch values.Field(i).Kind() {
		case reflect.String:
			if len(values.Field(i).String()) == 0 {
				*emptyStrings = append(*emptyStrings, strings.Join(append(prefix, fields.Field(i).Name), "."))
			}
		case reflect.Struct:
			_, err := findEmptyStringsInternal(values.Field(i).Interface(), emptyStrings, append(prefix, fields.Field(i).Name))
			if err != nil {
				return false, err
			}
		case reflect.Int:
		case reflect.Int32:
		case reflect.Int64:
		case reflect.Bool:
		case reflect.Slice:
		default:
			return false, fmt.Errorf("Unexpected field: %v, value: %v", fields.Field(i), values.Field(i))
		}
	}
	return false, nil
}
