package cloudfoundry

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"strings"
)

//ReadServiceKeyAbapEnvironment from Cloud Foundry and returns it.
//Depending on user/developer requirements if he wants to perform further Cloud Foundry actions the cfLogoutOption parameters gives the option to logout after reading ABAP communication arrangement or not.
func (cf *CFUtils) ReadServiceKeyAbapEnvironment(options ServiceKeyOptions, cfLogoutOption bool) (ServiceKey, error) {

	_c := cf.Exec

	if _c == nil {
		_c = &command.Command{}
	}

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

	err = cf.Login(config)
	var serviceKeyBytes bytes.Buffer
	_c.Stdout(&serviceKeyBytes)
	if err == nil {
		//Reading Service Key
		log.Entry().WithField("cfServiceInstance", options.CfServiceInstance).WithField("cfServiceKey", options.CfServiceKey).Info("Read service key for service instance")

		cfReadServiceKeyScript := []string{"service-key", options.CfServiceInstance, options.CfServiceKey}

		err = _c.RunExecutable("cf", cfReadServiceKeyScript...)
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
			logoutErr = cf.Logout()
			if logoutErr != nil {
				return abapServiceKey, fmt.Errorf("Failed to Logout of Cloud Foundry: %w", err)
			}
		}
		return abapServiceKey, fmt.Errorf("Reading Service Key failed: %w", err)
	}

	//Logging out of CF
	if cfLogoutOption == true {
		var logoutErr error
		logoutErr = cf.Logout()
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
