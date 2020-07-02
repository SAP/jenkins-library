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

// ReadServiceKeyAbapEnvironment from Cloud Foundry and returns it.
// Depending on user/developer requirements if he wants to perform further Cloud Foundry actions
// the cfLogoutOption parameters gives the option to logout after reading ABAP communication arrangement or not.
func ReadServiceKeyAbapEnvironment(options AbapEnvironmentOptions, cfLogoutOption bool) (AbapServiceKey, error) {
	var abapServiceKey AbapServiceKey
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

	// Command
	var c = &command.Command{}

	c.Stdout(&serviceKeyBytes)
	if err == nil {
		// Reading Service Key
		log.Entry().WithField("cfServiceInstance", options.CfServiceInstance).WithField("cfServiceKey", options.CfServiceKeyName).Info("Read service key for service instance")

		cfReadServiceKeyScript := []string{"service-key", options.CfServiceInstance, options.CfServiceKeyName}

		err = c.RunExecutable("cf", cfReadServiceKeyScript...)
	}
	if err == nil {
		var serviceKeyJSON string

		if len(serviceKeyBytes.String()) > 0 {
			var lines []string = strings.Split(serviceKeyBytes.String(), "\n")
			serviceKeyJSON = strings.Join(lines[2:], "")
		}

		json.Unmarshal([]byte(serviceKeyJSON), &abapServiceKey)
		if abapServiceKey == (AbapServiceKey{}) {
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

	// Logging out of CF
	if cfLogoutOption == true {
		var logoutErr error
		logoutErr = cloudfoundry.Logout()
		if logoutErr != nil {
			return abapServiceKey, fmt.Errorf("Failed to Logout of Cloud Foundry: %w", err)
		}
	}
	return abapServiceKey, nil
}

// GetAbapCommunicationArrangementInfo function fetches the communcation arrangement information for scenario 0510 of SAP CP ABAP Environment
// Therefore the MANAGE_GIT_REPOSITORY OData service is used
func GetAbapCommunicationArrangementInfo(config AbapEnvironmentOptions, c command.Command, cfLoginOption bool) (ConnectionDetailsHTTP, error) {

	oDataServiceSapCom0510 := "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/"
	pullAction := "Pull"

	var connectionDetails ConnectionDetailsHTTP
	var error error

	if config.Host != "" {
		// Host, User and Password are directly provided
		connectionDetails.URL = "https://" + config.Host + oDataServiceSapCom0510 + pullAction //"/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull"
		connectionDetails.User = config.Username
		connectionDetails.Password = config.Password
	} else {
		if config.CfAPIEndpoint == "" || config.CfOrg == "" || config.CfSpace == "" || config.CfServiceInstance == "" || config.CfServiceKeyName == "" {
			var err = errors.New("Parameters missing. Please provide EITHER the Host of the ABAP server OR the Cloud Foundry ApiEndpoint, Organization, Space, Service Instance and a corresponding Service Key for the Communication Scenario SAP_COM_0510")
			return connectionDetails, err
		}
		// Url, User and Password should be read from a cf service key
		var abapServiceKey, error = ReadServiceKeyAbapEnvironment(config, cfLoginOption)
		if error != nil {
			return connectionDetails, errors.Wrap(error, "Read service key failed")
		}
		connectionDetails.URL = abapServiceKey.URL + oDataServiceSapCom0510 + pullAction // "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull"
		connectionDetails.User = abapServiceKey.Abap.Username
		connectionDetails.Password = abapServiceKey.Abap.Password
	}
	return connectionDetails, error
}

/****************************************
 *	Structs for the A4C_A2G_GHA service *
 ****************************************/

// SoftwareComponentEntity struct for the root entity SoftwareComponent A4C_A2G_GHA_SC
type SoftwareComponentEntity struct {
	Metadata            AbapMetadata `json:"__metadata"`
	Namespace           string       `json:"namepsace"`
	Name                string       `json:"sc_name"`
	Description         string       `json:"descr"`
	Type                string       `json:"sc_type"`
	TypeDescription     string       `json:"sc_type_descr"`
	LastImportID        string       `json:"imp_id"`
	ActiveBranch        string       `json:"active_branch"`
	AvailableOnInstance bool         `json:"avail_on_inst"`
	NewVersionAvailable bool         `json:"new_vers_avail"`
	CreatedBy           string       `json:"created_by"`
	CreatedAt           string       `json:"created_at"`
	ChangedBy           string       `json:"changed_by"`
	ChangedAt           string       `json:"changed_at"`
	RelativeChangedAt   string       `json:"relative_changed_at"`
	ToImport            PullEntity   `json:"to_Import"`
	ToBranch            BranchEntity `json:"to_Branch"`
}

// PullEntity struct for the Pull/Import entity A4C_A2G_GHA_SC_IMP
type PullEntity struct {
	Metadata          AbapMetadata `json:"__metadata"`
	UUID              string       `json:"uuid"`
	Namespace         string       `json:"namepsace"`
	ScName            string       `json:"sc_name"`
	ImportType        string       `json:"import_type"`
	BranchName        string       `json:"branch_name"`
	StartedByUser     string       `json:"user_name"`
	Status            string       `json:"status"`
	StatusDescription string       `json:"status_descr"`
	CommitID          string       `json:"commit_id"`
	StartTime         string       `json:"start_time"`
	ChangeTime        string       `json:"change_time"`
	ToExecutionLog    AbapLogs     `json:"to_Execution_log"`
	ToTransportLog    AbapLogs     `json:"to_Transport_log"`
}

// BranchEntity struct for the Branch entity A4C_A2G_GHA_SC_BRANCH
type BranchEntity struct {
	Metadata      AbapMetadata `json:"__metadata"`
	ScName        string       `json:"sc_name"`
	Namespace     string       `json:"namepsace"`
	BranchName    string       `json:"branch_name"`
	ParentBranch  string       `json:"derived_from"`
	CreatedBy     string       `json:"created_by"`
	CreatedOn     string       `json:"created_on"`
	IsActive      bool         `json:"is_active"`
	CommitID      string       `json:"commit_id"`
	CommitMessage string       `json:"commit_message"`
	LastCommitBy  string       `json:"last_commit_by"`
	LastCommitOn  string       `json:"last_commit_on"`
}

// AbapLogs struct for LogResults
type AbapLogs struct {
	Results []LogResults `json:"results"`
}

// LogResults struct for Execution and Transport Log entities A4C_A2G_GHA_SC_LOG_EXE and A4C_A2G_GHA_SC_LOG_TP
type LogResults struct {
	Index       string `json:"index_no"`
	Type        string `json:"type"`
	Description string `json:"descr"`
	Timestamp   string `json:"timestamp"`
}

/********************************
 *	Structs for ABAP in general *
 ********************************/

// AbapMetadata struct
type AbapMetadata struct {
	URI string `json:"uri"`
}

// ConnectionDetailsHTTP struct
type ConnectionDetailsHTTP struct {
	User       string `json:"user"`
	Password   string `json:"password"`
	URL        string `json:"url"`
	XCsrfToken string `json:"xcsrftoken"`
}

// AbapError struct
type AbapError struct {
	Code    string           `json:"code"`
	Message AbapErrorMessage `json:"message"`
}

// AbapErrorMessage struct
type AbapErrorMessage struct {
	Lang  string `json:"lang"`
	Value string `json:"value"`
}

// AbapEnvironmentOptions struct
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

// AbapServiceKey struct
type AbapServiceKey struct {
	SapCloudService    string         `json:"sap.cloud.service"`
	URL                string         `json:"url"`
	SystemID           string         `json:"systemid"`
	Abap               AbapConnection `json:"abap"`
	Binding            AbapBinding    `json:"binding"`
	PreserveHostHeader bool           `json:"preserve_host_header"`
}

// AbapConnection contains information about the ABAP connection for the ABAP endpoint
type AbapConnection struct {
	Username                         string `json:"username"`
	Password                         string `json:"password"`
	CommunicationScenarioID          string `json:"communication_scenario_id"`
	CommunicationArrangementID       string `json:"communication_arrangement_id"`
	CommunicationSystemID            string `json:"communication_system_id"`
	CommunicationInboundUserID       string `json:"communication_inbound_user_id"`
	CommunicationInboundUserAuthMode string `json:"communication_inbound_user_auth_mode"`
}

// AbapBinding contains information about service binding in Cloud Foundry
type AbapBinding struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Env     string `json:"env"`
}
