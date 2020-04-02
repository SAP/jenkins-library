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

func LoginCheck(options CloudFoundryLoginOptions) (bool, error) {

	var err error

	//Check if logged in --> Cf api command responds with "not logged in" if positive
	var cfCheckLoginScript = []string{"api"}

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

func Login(options CloudFoundryLoginOptions) error {
	var err error

	var loggedIn bool

	loggedIn, err = LoginCheck(options)
	if err != nil {
		return fmt.Errorf("Failed to check if logged in: %w", err)
	}

	if loggedIn == true {
		return err
	}

	log.Entry().Info("Logging in to Cloud Foundry")

	var cfLoginScript = []string{"login", "-a", options.CfAPIEndpoint, "-o", options.CfOrg, "-s", options.CfSpace, "-u", options.Username, "-p", options.Password}

	log.Entry().WithField("cfAPI:", options.CfAPIEndpoint).WithField("cfOrg", options.CfOrg).WithField("space", options.CfSpace).Info("Logging into Cloud Foundry..")

	err = c.RunExecutable("cf", cfLoginScript...)

	if err != nil {
		return fmt.Errorf("Failed to login to Cloud Foundry: %w", err)
	}
	log.Entry().Info("Logged in successfully to Cloud Foundry..")
	return nil
}

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

func ReadServiceKey(options CloudFoundryReadServiceKeyOptions, cfLogoutOption bool) (ServiceKey, error) {

	var abapServiceKey ServiceKey
	var err error

	//Logging into Cloud Foundry
	config := CloudFoundryLoginOptions{
		CfAPIEndpoint: options.CfAPIEndpoint,
		CfOrg:         options.CfOrg,
		CfSpace:       options.CfSpace,
		Username:      options.Username,
		Password:      options.Password,
	}

	err = Login(config)
	if err != nil {
		return abapServiceKey, fmt.Errorf("Login at Cloud Foundry failed: %w", err)
	}

	//Reading Service Key
	var serviceKeyBytes bytes.Buffer
	c.Stdout(&serviceKeyBytes)

	log.Entry().WithField("cfServiceInstance", options.CfServiceInstance).WithField("cfServiceKey", options.CfServiceKey).Info("Read service key for service instance")

	cfReadServiceKeyScript := []string{"service-key", options.CfServiceInstance, options.CfServiceKey}

	err = c.RunExecutable("cf", cfReadServiceKeyScript...)
	if err != nil {
		return abapServiceKey, fmt.Errorf("Reading Service Key failed: %w", err)
	}
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

type CloudFoundryReadServiceKeyOptions struct {
	CfAPIEndpoint     string
	CfOrg             string
	CfSpace           string
	CfServiceInstance string
	CfServiceKey      string
	Username          string
	Password          string
}

type CloudFoundryLoginOptions struct {
	CfAPIEndpoint string
	CfOrg         string
	CfSpace       string
	Username      string
	Password      string
}

type ServiceKey struct {
	Abap     AbapConenction `json:"abap"`
	Binding  AbapBinding    `json:"binding"`
	Systemid string         `json:"systemid"`
	URL      string         `json:"url"`
}

type AbapConenction struct {
	CommunicationArrangementID string `json:"communication_arrangement_id"`
	CommunicationScenarioID    string `json:"communication_scenario_id"`
	CommunicationSystemID      string `json:"communication_system_id"`
	Password                   string `json:"password"`
	Username                   string `json:"username"`
}

type AbapBinding struct {
	Env     string `json:"env"`
	ID      string `json:"id"`
	Type    string `json:"type"`
	Version string `json:"version"`
}

type ServiceKeyOptions struct {
	Username          string `json:"username,omitempty"`
	Password          string `json:"password,omitempty"`
	RepositoryName    string `json:"repositoryName,omitempty"`
	Host              string `json:"host,omitempty"`
	CfAPIEndpoint     string `json:"cfApiEndpoint,omitempty"`
	CfOrg             string `json:"cfOrg,omitempty"`
	CfSpace           string `json:"cfSpace,omitempty"`
	CfServiceInstance string `json:"cfServiceInstance,omitempty"`
	CfServiceKey      string `json:"cfServiceKey,omitempty"`
}
