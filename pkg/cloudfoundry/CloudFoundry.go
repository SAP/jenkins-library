package cloudfoundry

import (
	"bytes"
	"encoding/json"
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

//ReadServiceKey ABAP Service Key from Cloud Foundry and returns it.
//Depending on user requirements if he wants to perform further Cloud Foundry actions the cfLogoutOption parameters gives the option to logout or not.
func ReadServiceKey(options ServiceKeyOptions, cfLogoutOption bool) (ServiceKey, error) {
	var abapServiceKey ServiceKey
	var err error

	//Logging into Cloud Foundry
	config := LoginOptions{
		CfAPIEndpoint: options.CfAPIEndpoint,
		CfOrg:         options.CfOrg,
		CfSpace:       options.CfSpace,
		Username:      options.Username,
		Password:      options.Password,
	}

	err = Login(config)
	var serviceKeyBytes bytes.Buffer
	c.Stdout(&serviceKeyBytes)
	if err == nil {
		//Reading Service Key
		log.Entry().WithField("cfServiceInstance", options.CfServiceInstance).WithField("cfServiceKey", options.CfServiceKey).Info("Read service key for service instance")

		cfReadServiceKeyScript := []string{"service-key", options.CfServiceInstance, options.CfServiceKey}

		err = c.RunExecutable("cf", cfReadServiceKeyScript...)
	}
	if err == nil {
		var serviceKeyJSON string

		if len(serviceKeyBytes.String()) > 0 {
			var lines []string = strings.Split(serviceKeyBytes.String(), "\n")
			serviceKeyJSON = strings.Join(lines[2:], "")
		}

		json.Unmarshal([]byte(serviceKeyJSON), &abapServiceKey)
		if abapServiceKey == (ServiceKey{}) {
			return abapServiceKey, errors.New("Parsing the service key failed")
		}

		log.Entry().Info("Service Key read successfully")
	}
	if err != nil {
		if cfLogoutOption == true {
			var logoutErr error
			logoutErr = Logout()
			if logoutErr != nil {
				return abapServiceKey, fmt.Errorf("Failed to Logout of Cloud Foundry: %w", err)
			}
		}
		return abapServiceKey, fmt.Errorf("Reading Service Key failed: %w", err)
	}

	//Logging out of CF
	if cfLogoutOption == true {
		var logoutErr error
		logoutErr = Logout()
		if logoutErr != nil {
			return abapServiceKey, fmt.Errorf("Failed to Logout of Cloud Foundry: %w", err)
		}
	}
	return abapServiceKey, nil
}

//ServiceKeyOptions for reading CF Service Key
type ServiceKeyOptions struct {
	CfAPIEndpoint     string
	CfOrg             string
	CfSpace           string
	CfServiceInstance string
	CfServiceKey      string
	Username          string
	Password          string
}

//LoginOptions for logging in to CF
type LoginOptions struct {
	CfAPIEndpoint string
	CfOrg         string
	CfSpace       string
	Username      string
	Password      string
}

//ServiceKey struct to parse CF Service Key
type ServiceKey struct {
	Abap     AbapConnection `json:"abap"`
	Binding  AbapBinding    `json:"binding"`
	Systemid string         `json:"systemid"`
	URL      string         `json:"url"`
}

//AbapConnection contains information about the ABAP connection for the ABAP endpoint
type AbapConnection struct {
	CommunicationArrangementID string `json:"communication_arrangement_id"`
	CommunicationScenarioID    string `json:"communication_scenario_id"`
	CommunicationSystemID      string `json:"communication_system_id"`
	Password                   string `json:"password"`
	Username                   string `json:"username"`
}

//AbapBinding contains information about service binding in Cloud Foundry
type AbapBinding struct {
	Env     string `json:"env"`
	ID      string `json:"id"`
	Type    string `json:"type"`
	Version string `json:"version"`
}
