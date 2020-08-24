package cloudfoundry

import (
	"bytes"
<<<<<<< HEAD
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
)

//ReadServiceKeyAbapEnvironment from Cloud Foundry and returns it.
//Depending on user/developer requirements if he wants to perform further Cloud Foundry actions the cfLogoutOption parameters gives the option to logout after reading ABAP communication arrangement or not.
func ReadServiceKeyAbapEnvironment(options ServiceKeyOptions, cfLogoutOption bool) (ServiceKey, error) {
	var abapServiceKey ServiceKey
	var err error

	//Logging into Cloud Foundry
	config := LoginOptions{
=======
	"fmt"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
)

// ReadServiceKey reads a cloud foundry service key based on provided service instance and service key name parameters
func (cf *CFUtils) ReadServiceKey(options ServiceKeyOptions) (string, error) {

	_c := cf.Exec

	if _c == nil {
		_c = &command.Command{}
	}
	cfconfig := LoginOptions{
>>>>>>> 67feb87b800243c559aacd67191796e9f39bfeee
		CfAPIEndpoint: options.CfAPIEndpoint,
		CfOrg:         options.CfOrg,
		CfSpace:       options.CfSpace,
		Username:      options.Username,
		Password:      options.Password,
	}
<<<<<<< HEAD

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
=======
	err := cf.Login(cfconfig)

	if err != nil {
		// error while trying to run cf login
		return "", fmt.Errorf("Login to Cloud Foundry failed: %w", err)
	}
	var serviceKeyBytes bytes.Buffer
	_c.Stdout(&serviceKeyBytes)

	// we are logged in --> read service key
	log.Entry().WithField("cfServiceInstance", options.CfServiceInstance).WithField("cfServiceKey", options.CfServiceKeyName).Info("Read service key for service instance")
	cfReadServiceKeyScript := []string{"service-key", options.CfServiceInstance, options.CfServiceKeyName}
	err = _c.RunExecutable("cf", cfReadServiceKeyScript...)

	if err != nil {
		// error while reading service key
		return "", fmt.Errorf("Reading service key failed: %w", err)
	}

	// parse and return service key
	var serviceKeyJSON string
	if len(serviceKeyBytes.String()) > 0 {
		var lines []string = strings.Split(serviceKeyBytes.String(), "\n")
		serviceKeyJSON = strings.Join(lines[2:], "")
	}

	err = cf.Logout()
	if err != nil {
		return serviceKeyJSON, fmt.Errorf("Logout of Cloud Foundry failed: %w", err)
	}

	return serviceKeyJSON, err
>>>>>>> 67feb87b800243c559aacd67191796e9f39bfeee
}

//ServiceKeyOptions for reading CF Service Key
type ServiceKeyOptions struct {
	CfAPIEndpoint     string
	CfOrg             string
	CfSpace           string
	CfServiceInstance string
<<<<<<< HEAD
	CfServiceKey      string
	Username          string
	Password          string
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
=======
	CfServiceKeyName  string
	Username          string
	Password          string
}
>>>>>>> 67feb87b800243c559aacd67191796e9f39bfeee
