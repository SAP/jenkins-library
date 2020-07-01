package abaputils

import (
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func getAbapCommunicationArrangementInfo(config abapEnvironmentPullGitRepoOptions, c execRunner) (connectionDetailsHTTP, error) {

	var connectionDetails connectionDetailsHTTP
	var error error

	if config.Host != "" {
		// Host, User and Password are directly provided
		connectionDetails.URL = "https://" + config.Host + "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull"
		connectionDetails.User = config.Username
		connectionDetails.Password = config.Password
	} else {
		if config.CfAPIEndpoint == "" || config.CfOrg == "" || config.CfSpace == "" || config.CfServiceInstance == "" || config.CfServiceKeyName == "" {
			var err = errors.New("Parameters missing. Please provide EITHER the Host of the ABAP server OR the Cloud Foundry ApiEndpoint, Organization, Space, Service Instance and a corresponding Service Key for the Communication Scenario SAP_COM_0510")
			return connectionDetails, err
		}
		// Url, User and Password should be read from a cf service key
		var abapServiceKey, error = readCfServiceKey(config, c)
		if error != nil {
			return connectionDetails, errors.Wrap(error, "Read service key failed")
		}
		connectionDetails.URL = abapServiceKey.URL + "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull"
		connectionDetails.User = abapServiceKey.Abap.Username
		connectionDetails.Password = abapServiceKey.Abap.Password
	}
	return connectionDetails, error
}

func ConvertTime(logTimeStamp string) time.Time {
	// The ABAP Environment system returns the date in the following format: /Date(1585576807000+0000)/
	seconds := strings.TrimPrefix(strings.TrimSuffix(logTimeStamp, "000+0000)/"), "/Date(")
	n, error := strconv.ParseInt(seconds, 10, 64)
	if error != nil {
		return time.Unix(0, 0).UTC()
	}
	t := time.Unix(n, 0).UTC()
	return t
}

type AbapEntity struct {
	Metadata       AbapMetadata `json:"__metadata"`
	UUID           string       `json:"uuid"`
	ScName         string       `json:"sc_name"`
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

type ServiceKey struct {
	Abap     AbapConnection `json:"abap"`
	Binding  AbapBinding    `json:"binding"`
	Systemid string         `json:"systemid"`
	URL      string         `json:"url"`
}

type Deferred struct {
	URI string `json:"uri"`
}

type AbapConnection struct {
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
