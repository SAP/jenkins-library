package cloudfoundry

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
)

// LoginCheck checks if user is logged in to Cloud Foundry with the receiver provided
// to the function call.
func (cf *CFUtils) LoginCheck(options LoginOptions) (bool, error) {
	// TODO: options are not used ?
	return cf.loggedIn, nil
}

// Login logs user in to Cloud Foundry via cf cli.
// Checks if user is logged in first, if not perform 'cf login' command with appropriate parameters
func (cf *CFUtils) Login(options LoginOptions) error {
	_c := cf.Exec

	if _c == nil {
		_c = &command.Command{}
	}

	if options.CfAPIEndpoint == "" || options.CfOrg == "" || options.CfSpace == "" || options.Username == "" || options.Password == "" {
		return fmt.Errorf("Failed to login to Cloud Foundry: %w", errors.New("Parameters missing. Please provide the Cloud Foundry Endpoint, Org, Space, Username and Password"))
	}

	if cf.loggedIn {
		return nil
	}

	log.Entry().Info("Logging in to Cloud Foundry")

	escapedUsername := escapeValuesForCLI(options.Username, getGOOS)
	escapedPassword := escapeValuesForCLI(options.Password, getGOOS)

	cfLoginScript := []string{
		"login",
		"-a", options.CfAPIEndpoint,
		"-o", options.CfOrg,
		"-s", options.CfSpace,
		"-u", escapedUsername,
		"-p", escapedPassword,
	}

	cfLoginScript = append(cfLoginScript, options.CfLoginOpts...)

	log.Entry().WithField("cfAPI:", options.CfAPIEndpoint).WithField("cfOrg", options.CfOrg).WithField("space", options.CfSpace).Info("Logging into Cloud Foundry..")

	err := _c.RunExecutable("cf", cfLoginScript...)
	if err != nil {
		return fmt.Errorf("Failed to login to Cloud Foundry: %w", err)
	}
	log.Entry().Info("Logged in successfully to Cloud Foundry..")
	cf.loggedIn = true
	return nil
}

func escapeValuesForCLI(password string, getGOOS func() string) string {
	switch getGOOS() {
	case "windows":
		return fmt.Sprintf("'%s'", strings.ReplaceAll(password, "'", `''`))
	case "linux", "darwin":
		return fmt.Sprintf("'%s'", strings.ReplaceAll(password, "'", `'\''`))
	default:
		return fmt.Sprintf("'%s'", strings.ReplaceAll(password, "'", `'\''`))
	}
}

// Logout logs User out of Cloud Foundry
// Logout can be perforned via 'cf logout' command regardless if user is logged in or not
func (cf *CFUtils) Logout() error {

	_c := cf.Exec

	if _c == nil {
		_c = &command.Command{}
	}

	var cfLogoutScript = "logout"

	log.Entry().Info("Logging out of Cloud Foundry")

	err := _c.RunExecutable("cf", cfLogoutScript)
	if err != nil {
		return fmt.Errorf("Failed to Logout of Cloud Foundry: %w", err)
	}
	log.Entry().Info("Logged out successfully")
	cf.loggedIn = false
	return nil
}

// LoginOptions for logging in to CF
type LoginOptions struct {
	CfAPIEndpoint string
	CfOrg         string
	CfSpace       string
	Username      string
	Password      string
	CfLoginOpts   []string
}

// CFUtils ...
type CFUtils struct {
	// In order to avoid clashes between parallel workflows requiring cf login/logout
	// this instance of command.ExecRunner can be configured accordingly by settings the
	// environment variables CF_HOME to distict directories.
	// In order to ensure plugins installed to the cf cli are found environment variables
	// CF_PLUGIN_HOME can be set accordingly.
	Exec     command.ExecRunner
	loggedIn bool
}

// AuthenticationUtils - interface for cloud foundry login and logout
type AuthenticationUtils interface {
	Login(options LoginOptions) error
	Logout() error
}

// CfUtilsMock - mock for CfUtils
type CfUtilsMock struct {
	LoginError  error
	LogoutError error
}

// Login mock implementation
func (cf *CfUtilsMock) Login(options LoginOptions) error {
	return cf.LoginError
}

// Logout mock implementation
func (cf *CfUtilsMock) Logout() error {
	return cf.LogoutError
}

// Cleanup for CfUtilsMock
func (cf *CfUtilsMock) Cleanup() {
	cf.LoginError = nil
	cf.LogoutError = nil
}

func getGOOS() string {
	return runtime.GOOS
}
