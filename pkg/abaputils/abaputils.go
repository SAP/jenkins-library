package abaputils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
)

// AbapEnvironmentCommunicationArrangement Interface for all abap piper steps
// type AbapEnvironmentCommunicationArrangement interface {
// 	GetAbapCommunicationArrangementInfo(config AbapEnvironmentOptions, c execRunner) ConnectionDetailsHTTP
// 	ReadCfServiceKey(config AbapEnvironmentOptions, c execRunner) ServiceKey
// }

//ReadServiceKeyAbapEnvironment from Cloud Foundry and returns it.
//Depending on user/developer requirements if he wants to perform further Cloud Foundry actions the cfLogoutOption parameters gives the option to logout after reading ABAP communication arrangement or not.
func ReadServiceKeyAbapEnvironment(options ServiceKeyOptions, cfLogoutOption bool) (ServiceKey, error) {
	var abapServiceKey ServiceKey
	var err error

	//Logging into Cloud Foundry
	config := cloudfoundry.LoginOptions{
		CfAPIEndpoint: options.CfAPIEndpoint,
		CfOrg:         options.CfOrg,
		CfSpace:       options.CfSpace,
		Username:      options.Username,
		Password:      options.Password,
	}

	err = cloudfoundry.Login(config)
	var serviceKeyBytes bytes.Buffer

	var c = &command.Command{}
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
			logoutErr = cloudfoundry.Logout()
			if logoutErr != nil {
				return abapServiceKey, fmt.Errorf("Failed to Logout of Cloud Foundry: %w", err)
			}
		}
		return abapServiceKey, fmt.Errorf("Reading Service Key failed: %w", err)
	}

	//Logging out of CF
	if cfLogoutOption == true {
		var logoutErr error
		logoutErr = cloudfoundry.Logout()
		if logoutErr != nil {
			return abapServiceKey, fmt.Errorf("Failed to Logout of Cloud Foundry: %w", err)
		}
	}
	return abapServiceKey, nil
}

// GetAbapCommunicationArrangementInfo does ...
// func GetAbapCommunicationArrangementInfo(config AbapEnvironmentOptions, c execRunner) (ConnectionDetailsHTTP, error) {

// 	oDataServiceSapCom0510 := "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/"
// 	pullAction := "Pull"

// 	var connectionDetails ConnectionDetailsHTTP
// 	var error error

// 	if config.Host != "" {
// 		// Host, User and Password are directly provided
// 		connectionDetails.URL = "https://" + config.Host + oDataServiceSapCom0510 + pullAction //"/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull"
// 		connectionDetails.User = config.Username
// 		connectionDetails.Password = config.Password
// 	} else {
// 		if config.CfAPIEndpoint == "" || config.CfOrg == "" || config.CfSpace == "" || config.CfServiceInstance == "" || config.CfServiceKeyName == "" {
// 			var err = errors.New("Parameters missing. Please provide EITHER the Host of the ABAP server OR the Cloud Foundry ApiEndpoint, Organization, Space, Service Instance and a corresponding Service Key for the Communication Scenario SAP_COM_0510")
// 			return connectionDetails, err
// 		}
// 		// Url, User and Password should be read from a cf service key
// 		var abapServiceKey, error = ReadCfServiceKey(config, c)
// 		if error != nil {
// 			return connectionDetails, errors.Wrap(error, "Read service key failed")
// 		}
// 		connectionDetails.URL = abapServiceKey.URL + oDataServiceSapCom0510 + pullAction // "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull"
// 		connectionDetails.User = abapServiceKey.Abap.Username
// 		connectionDetails.Password = abapServiceKey.Abap.Password
// 	}
// 	return connectionDetails, error
// }

// // ReadCfServiceKey does ...
// func ReadCfServiceKey(config AbapEnvironmentOptions, c execRunner) (ServiceKey, error) { //  not needed

// 	var abapServiceKey ServiceKey

// 	c.Stderr(log.Writer())

// 	// Logging into the Cloud Foundry via CF CLI
// 	log.Entry().WithField("cfApiEndpoint", config.CfAPIEndpoint).WithField("cfSpace", config.CfSpace).WithField("cfOrg", config.CfOrg).WithField("User", config.Username).Info("Cloud Foundry parameters: ")
// 	cfLoginSlice := []string{"login", "-a", config.CfAPIEndpoint, "-u", config.Username, "-p", config.Password, "-o", config.CfOrg, "-s", config.CfSpace}
// 	errorRunExecutable := c.RunExecutable("cf", cfLoginSlice...)
// 	if errorRunExecutable != nil {
// 		log.Entry().Error("Login at cloud foundry failed.")
// 		return abapServiceKey, errorRunExecutable
// 	}

// 	// Reading the Service Key via CF CLI
// 	var serviceKeyBytes bytes.Buffer
// 	c.Stdout(&serviceKeyBytes)
// 	cfReadServiceKeySlice := []string{"service-key", config.CfServiceInstance, config.CfServiceKeyName}
// 	errorRunExecutable = c.RunExecutable("cf", cfReadServiceKeySlice...)
// 	var serviceKeyJSON string
// 	if len(serviceKeyBytes.String()) > 0 {
// 		var lines []string = strings.Split(serviceKeyBytes.String(), "\n")
// 		serviceKeyJSON = strings.Join(lines[2:], "")
// 	}
// 	if errorRunExecutable != nil {
// 		return abapServiceKey, errorRunExecutable
// 	}
// 	log.Entry().WithField("cfServiceInstance", config.CfServiceInstance).WithField("cfServiceKeyName", config.CfServiceKeyName).Info("Read service key for service instance")
// 	json.Unmarshal([]byte(serviceKeyJSON), &abapServiceKey)
// 	if abapServiceKey == (ServiceKey{}) {
// 		return abapServiceKey, errors.New("Parsing the service key failed")
// 	}
// 	return abapServiceKey, errorRunExecutable
// }

// ConvertTime does ...
// func ConvertTime(logTimeStamp string) time.Time {
// 	// The ABAP Environment system returns the date in the following format: /Date(1585576807000+0000)/
// 	seconds := strings.TrimPrefix(strings.TrimSuffix(logTimeStamp, "000+0000)/"), "/Date(")
// 	n, error := strconv.ParseInt(seconds, 10, 64)
// 	if error != nil {
// 		return time.Unix(0, 0).UTC()
// 	}
// 	t := time.Unix(n, 0).UTC()
// 	return t
// }

type SoftwareComponentEntity struct {
	Metadata       AbapMetadata `json:"__metadata"`
	UUID           string       `json:"uuid"`
	Name           string       `json:"sc_name"`
	Namespace      string       `json:"namepsace"`
	Status         string       `json:"status"`
	StatusDescr    string       `json:"status_descr"`
	ToExecutionLog AbapLogs     `json:"to_Execution_log"`
	ToTransportLog AbapLogs     `json:"to_Transport_log"`
}

type AbapMetadata struct {
	URI string `json:"uri"`
}

type AbapLogs struct {
	Results []LogResults `json:"results"`
}

type LogResults struct {
	Index       string `json:"index_no"`
	Type        string `json:"type"`
	Description string `json:"descr"`
	Timestamp   string `json:"timestamp"`
}

type ConnectionDetailsHTTP struct {
	User       string `json:"user"`
	Password   string `json:"password"`
	URL        string `json:"url"`
	XCsrfToken string `json:"xcsrftoken"`
}

type AbapError struct {
	Code    string           `json:"code"`
	Message AbapErrorMessage `json:"message"`
}

type AbapErrorMessage struct {
	Lang  string `json:"lang"`
	Value string `json:"value"`
}

type AbapEnvironmentOptions struct {
	Username          string `json:"username,omitempty"`
	Password          string `json:"password,omitempty"`
	Host              string `json:"host,omitempty"`
	CfAPIEndpoint     string `json:"cfApiEndpoint,omitempty"`
	CfOrg             string `json:"cfOrg,omitempty"`
	CfSpace           string `json:"cfSpace,omitempty"`
	CfServiceInstance string `json:"cfServiceInstance,omitempty"`
	CfServiceKeyName  string `json:"cfServiceKeyName,omitempty"`
}

// CF Structs

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
