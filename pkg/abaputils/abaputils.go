package abaputils

import (
	"encoding/json"
	"regexp"

	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
)

/*
AbapUtils Struct
*/
type AbapUtils struct {
	Exec command.ExecRunner
}

/*
Communication for defining function used for communication
*/
type Communication interface {
	GetAbapCommunicationArrangementInfo(options AbapEnvironmentOptions, oDataURL string) (ConnectionDetailsHTTP, error)
}

// GetAbapCommunicationArrangementInfo function fetches the communcation arrangement information in SAP CP ABAP Environment
func (abaputils *AbapUtils) GetAbapCommunicationArrangementInfo(options AbapEnvironmentOptions, oDataURL string) (ConnectionDetailsHTTP, error) {
	c := abaputils.Exec
	var connectionDetails ConnectionDetailsHTTP
	var error error

	if options.Host != "" {
		// Host, User and Password are directly provided -> check for host schema (double https)
		match, err := regexp.MatchString(`^(https|HTTPS):\/\/.*`, options.Host)
		if err != nil {
			return connectionDetails, errors.Wrap(err, "Schema validation for host parameter failed. Check for https.")
		}
		var hostOdataURL = options.Host + oDataURL
		if match {
			connectionDetails.URL = hostOdataURL
		} else {
			connectionDetails.URL = "https://" + hostOdataURL
		}
		connectionDetails.User = options.Username
		connectionDetails.Password = options.Password
	} else {
		if options.CfAPIEndpoint == "" || options.CfOrg == "" || options.CfSpace == "" || options.CfServiceInstance == "" || options.CfServiceKeyName == "" {
			var err = errors.New("Parameters missing. Please provide EITHER the Host of the ABAP server OR the Cloud Foundry ApiEndpoint, Organization, Space, Service Instance and a corresponding Service Key for the Communication Scenario SAP_COM_0510")
			return connectionDetails, err
		}
		// Url, User and Password should be read from a cf service key
		var abapServiceKey, error = ReadServiceKeyAbapEnvironment(options, c)
		if error != nil {
			return connectionDetails, errors.Wrap(error, "Read service key failed")
		}
		connectionDetails.URL = abapServiceKey.URL + oDataURL
		connectionDetails.User = abapServiceKey.Abap.Username
		connectionDetails.Password = abapServiceKey.Abap.Password
	}
	return connectionDetails, error
}

// ReadServiceKeyAbapEnvironment from Cloud Foundry and returns it.
// Depending on user/developer requirements if he wants to perform further Cloud Foundry actions
func ReadServiceKeyAbapEnvironment(options AbapEnvironmentOptions, c command.ExecRunner) (AbapServiceKey, error) {

	var abapServiceKey AbapServiceKey
	var serviceKeyJSON string
	var err error

	cfconfig := cloudfoundry.ServiceKeyOptions{
		CfAPIEndpoint:     options.CfAPIEndpoint,
		CfOrg:             options.CfOrg,
		CfSpace:           options.CfSpace,
		CfServiceInstance: options.CfServiceInstance,
		CfServiceKeyName:  options.CfServiceKeyName,
		Username:          options.Username,
		Password:          options.Password,
	}

	cf := cloudfoundry.CFUtils{Exec: c}

	serviceKeyJSON, err = cf.ReadServiceKey(cfconfig)

	if err != nil {
		// Executing cfReadServiceKeyScript failed
		return abapServiceKey, err
	}

	// parse
	json.Unmarshal([]byte(serviceKeyJSON), &abapServiceKey)
	if abapServiceKey == (AbapServiceKey{}) {
		return abapServiceKey, errors.New("Parsing the service key failed")
	}

	log.Entry().Info("Service Key read successfully")
	return abapServiceKey, nil
}

/****************************************
 *	Structs for the A4C_A2G_GHA service *
 ****************************************/

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

// AbapLogs struct for ABAP logs
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

/*******************************
 *	Structs for specific steps *
 *******************************/

// AbapEnvironmentPullGitRepoOptions struct for the PullGitRepo piper step
type AbapEnvironmentPullGitRepoOptions struct {
	AbapEnvOptions  AbapEnvironmentOptions
	RepositoryNames []string `json:"repositoryNames,omitempty"`
}

// AbapEnvironmentRunATCCheckOptions struct for the RunATCCheck piper step
type AbapEnvironmentRunATCCheckOptions struct {
	AbapEnvOptions AbapEnvironmentOptions
	AtcConfig      string `json:"atcConfig,omitempty"`
}

/********************************
 *	Structs for ABAP in general *
 ********************************/

//AbapEnvironmentOptions contains cloud foundry fields and the host parameter for connections to ABAP Environment instances
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

// AbapMetadata contains the URI of metadata files
type AbapMetadata struct {
	URI string `json:"uri"`
}

// ConnectionDetailsHTTP contains fields for HTTP connections including the XCSRF token
type ConnectionDetailsHTTP struct {
	User       string `json:"user"`
	Password   string `json:"password"`
	URL        string `json:"url"`
	XCsrfToken string `json:"xcsrftoken"`
}

// AbapError contains the error code and the error message for ABAP errors
type AbapError struct {
	Code    string           `json:"code"`
	Message AbapErrorMessage `json:"message"`
}

// AbapErrorMessage contains the lanuage and value fields for ABAP errors
type AbapErrorMessage struct {
	Lang  string `json:"lang"`
	Value string `json:"value"`
}

// AbapServiceKey contains information about an ABAP service key
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
