package cloudfoundry

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
)

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
}

//Login logs user in to Cloud Foundry via cf cli.
//Checks if user is logged in first, if not perform 'cf login' command with appropriate parameters
func Login(options LoginOptions) error {

	var err error

	if options.CfAPIEndpoint == "" || options.CfOrg == "" || options.CfSpace == "" || options.Username == "" || options.Password == "" {
		return fmt.Errorf("Failed to login to Cloud Foundry: %w", errors.New("Parameters missing. Please provide the Cloud Foundry Endpoint, Org, Space, Username and Password"))
	}

	var loggedIn bool

	loggedIn, err = LoginCheck(options)

	if loggedIn == true {
		return err
	}

	if err == nil {
		log.Entry().Info("Logging in to Cloud Foundry")

		var cfLoginScript = []string{"login", "-a", options.CfAPIEndpoint, "-o", options.CfOrg, "-s", options.CfSpace, "-u", options.Username, "-p", options.Password}

		log.Entry().WithField("cfAPI:", options.CfAPIEndpoint).WithField("cfOrg", options.CfOrg).WithField("space", options.CfSpace).Info("Logging into Cloud Foundry..")

		err = c.RunExecutable("cf", cfLoginScript...)
	}

	if err != nil {
		return fmt.Errorf("Failed to login to Cloud Foundry: %w", err)
	}
	log.Entry().Info("Logged in successfully to Cloud Foundry..")
	return nil
}

//Logout logs User out of Cloud Foundry
//Logout can be perforned via 'cf logout' command regardless if user is logged in or not
func Logout() error {
	var cfLogoutScript = "logout"

	log.Entry().Info("Logging out of Cloud Foundry")

	err := c.RunExecutable("cf", cfLogoutScript)
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
}
