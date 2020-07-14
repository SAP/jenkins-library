package cloudfoundry

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
)

//LoginCheck checks if user is logged in to Cloud Foundry.
//If user is not logged in 'cf api' command will return string that contains 'User is not logged in' only if user is not logged in.
//If the returned string doesn't contain the substring 'User is not logged in' we know he is logged in.
func (cf *CFUtils) LoginCheck(options LoginOptions) (bool, error) {
	var err error

	_c := cf.Exec

	if _c == nil {
		_c = &command.Command{}
	}

	if options.CfAPIEndpoint == "" {
		return false, errors.New("Cloud Foundry API endpoint parameter missing. Please provide the Cloud Foundry Endpoint")
	}

	//Check if logged in --> Cf api command responds with "not logged in" if positive
	var cfCheckLoginScript = append([]string{"api", options.CfAPIEndpoint}, options.CfAPIOpts...)

	defer func() {
		// We set it back to what is set from the generated stub. Of course this is not
		// fully accurate in case we create our own instance above (nothing handed in via
		// the receiver).
		// Would be better to remember the old stdout and set back to this.
		// But command.Command does not allow to get the currently set
		// stdout handler.
		// Reason for changing the output stream here: we need to parse the output
		// of the command issued here in order to check if we are already logged in.
		// This is expected to change soon to a boolean variable where we remember the
		// login state.
		_c.Stdout(log.Writer())
	}()

	var cfLoginBytes bytes.Buffer
	_c.Stdout(&cfLoginBytes)

	var result string

	err = _c.RunExecutable("cf", cfCheckLoginScript...)

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
}

//Login logs user in to Cloud Foundry via cf cli.
//Checks if user is logged in first, if not perform 'cf login' command with appropriate parameters
func (cf *CFUtils) Login(options LoginOptions) error {
	var err error

	_c := cf.Exec

	if _c == nil {
		_c = &command.Command{}
	}

	if options.CfAPIEndpoint == "" || options.CfOrg == "" || options.CfSpace == "" || options.Username == "" || options.Password == "" {
		return fmt.Errorf("Failed to login to Cloud Foundry: %w", errors.New("Parameters missing. Please provide the Cloud Foundry Endpoint, Org, Space, Username and Password"))
	}

	var loggedIn bool

	loggedIn, err = cf.LoginCheck(options)

	if loggedIn == true {
		return err
	}

	if err == nil {
		log.Entry().Info("Logging in to Cloud Foundry")

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
	}

	if err != nil {
		return fmt.Errorf("Failed to login to Cloud Foundry: %w", err)
	}
	log.Entry().Info("Logged in successfully to Cloud Foundry..")
	return nil
}

//Logout logs User out of Cloud Foundry
//Logout can be perforned via 'cf logout' command regardless if user is logged in or not
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
	return nil
}

//LoginOptions for logging in to CF
type LoginOptions struct {
	CfAPIEndpoint string
	CfOrg         string
	CfSpace       string
	Username      string
	Password      string
	CfAPIOpts     []string
	CfLoginOpts   []string
}

// CFUtils ...
type CFUtils struct {
	// In order to avoid clashes between parallel workflows requiring cf login/logout
	// this instance of command.ExecRunner can be configured accordingly by settings the
	// environment variables CF_HOME to distict directories.
	// In order to ensure plugins installed to the cf cli are found environment variables
	// CF_PLUGIN_HOME can be set accordingly.
	Exec command.ExecRunner
}
