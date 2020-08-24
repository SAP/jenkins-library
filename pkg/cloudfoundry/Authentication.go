package cloudfoundry

import (
<<<<<<< HEAD
	"bytes"
	"errors"
	"fmt"
	"strings"
=======
	"errors"
	"fmt"
>>>>>>> 67feb87b800243c559aacd67191796e9f39bfeee

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
)

<<<<<<< HEAD
var c = command.Command{}

//LoginCheck checks if user is logged in to Cloud Foundry.
//If user is not logged in 'cf api' command will return string that contains 'User is not logged in' only if user is not logged in.
//If the returned string doesn't contain the substring 'User is not logged in' we know he is logged in.
func LoginCheck(options LoginOptions) (bool, error) {
	var err error

	if options.CfAPIEndpoint == "" {
		return false, errors.New("Cloud Foundry API endpoint parameter missing. Please provide the Cloud Foundry Endpoint")
	}

	//Check if logged in --> Cf api command responds with "not logged in" if positive
	var cfCheckLoginScript = []string{"api", options.CfAPIEndpoint}

	var cfLoginBytes bytes.Buffer
	c.Stdout(&cfLoginBytes)

	var result string

	err = c.RunExecutable("cf", cfCheckLoginScript...)

	if err != nil {
		return false, fmt.Errorf("Failed to check if logged in: %w", err)
	}

	result = cfLoginBytes.String()
	log.Entry().WithField("result: ", result).Info("Login check")

	//Logged in
	if strings.Contains(result, "Not logged in") == false {
		log.Entry().Info("Login check indicates you are already logged in to Cloud Foundry")
		return true, err
	}

	//Not logged in
	log.Entry().Info("Login check indicates you are not yet logged in to Cloud Foundry")
	return false, err
=======
//LoginCheck checks if user is logged in to Cloud Foundry with the receiver provided
//to the function call.
func (cf *CFUtils) LoginCheck(options LoginOptions) (bool, error) {
	return cf.loggedIn, nil
>>>>>>> 67feb87b800243c559aacd67191796e9f39bfeee
}

//Login logs user in to Cloud Foundry via cf cli.
//Checks if user is logged in first, if not perform 'cf login' command with appropriate parameters
<<<<<<< HEAD
func Login(options LoginOptions) error {

	var err error

=======
func (cf *CFUtils) Login(options LoginOptions) error {
	var err error

	_c := cf.Exec

	if _c == nil {
		_c = &command.Command{}
	}

>>>>>>> 67feb87b800243c559aacd67191796e9f39bfeee
	if options.CfAPIEndpoint == "" || options.CfOrg == "" || options.CfSpace == "" || options.Username == "" || options.Password == "" {
		return fmt.Errorf("Failed to login to Cloud Foundry: %w", errors.New("Parameters missing. Please provide the Cloud Foundry Endpoint, Org, Space, Username and Password"))
	}

	var loggedIn bool

<<<<<<< HEAD
	loggedIn, err = LoginCheck(options)
=======
	loggedIn, err = cf.LoginCheck(options)
>>>>>>> 67feb87b800243c559aacd67191796e9f39bfeee

	if loggedIn == true {
		return err
	}

	if err == nil {
		log.Entry().Info("Logging in to Cloud Foundry")

<<<<<<< HEAD
		var cfLoginScript = []string{"login", "-a", options.CfAPIEndpoint, "-o", options.CfOrg, "-s", options.CfSpace, "-u", options.Username, "-p", options.Password}

		log.Entry().WithField("cfAPI:", options.CfAPIEndpoint).WithField("cfOrg", options.CfOrg).WithField("space", options.CfSpace).Info("Logging into Cloud Foundry..")

		err = c.RunExecutable("cf", cfLoginScript...)
=======
		var cfLoginScript = append([]string{
			"login",
			"-a", options.CfAPIEndpoint,
			"-o", options.CfOrg,
			"-s", options.CfSpace,
			"-u", options.Username,
			"-p", options.Password,
		}, options.CfLoginOpts...)

		log.Entry().WithField("cfAPI:", options.CfAPIEndpoint).WithField("cfOrg", options.CfOrg).WithField("space", options.CfSpace).Info("Logging into Cloud Foundry..")

		err = _c.RunExecutable("cf", cfLoginScript...)
>>>>>>> 67feb87b800243c559aacd67191796e9f39bfeee
	}

	if err != nil {
		return fmt.Errorf("Failed to login to Cloud Foundry: %w", err)
	}
	log.Entry().Info("Logged in successfully to Cloud Foundry..")
<<<<<<< HEAD
=======
	cf.loggedIn = true
>>>>>>> 67feb87b800243c559aacd67191796e9f39bfeee
	return nil
}

//Logout logs User out of Cloud Foundry
//Logout can be perforned via 'cf logout' command regardless if user is logged in or not
<<<<<<< HEAD
func Logout() error {
=======
func (cf *CFUtils) Logout() error {

	_c := cf.Exec

	if _c == nil {
		_c = &command.Command{}
	}

>>>>>>> 67feb87b800243c559aacd67191796e9f39bfeee
	var cfLogoutScript = "logout"

	log.Entry().Info("Logging out of Cloud Foundry")

<<<<<<< HEAD
	err := c.RunExecutable("cf", cfLogoutScript)
=======
	err := _c.RunExecutable("cf", cfLogoutScript)
>>>>>>> 67feb87b800243c559aacd67191796e9f39bfeee
	if err != nil {
		return fmt.Errorf("Failed to Logout of Cloud Foundry: %w", err)
	}
	log.Entry().Info("Logged out successfully")
<<<<<<< HEAD
=======
	cf.loggedIn = false
>>>>>>> 67feb87b800243c559aacd67191796e9f39bfeee
	return nil
}

//LoginOptions for logging in to CF
type LoginOptions struct {
	CfAPIEndpoint string
	CfOrg         string
	CfSpace       string
	Username      string
	Password      string
<<<<<<< HEAD
=======
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
>>>>>>> 67feb87b800243c559aacd67191796e9f39bfeee
}
