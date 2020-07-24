package abaputils

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"sort"
	"time"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
)

// PollEntity periodically polls the pull/import entity to get the status. Check if the import is still running
func PollEntity(repositoryName string, connectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender, pollIntervall time.Duration) (string, error) {

	log.Entry().Info("Start polling the status...")
	var status string = "R"

	for {
		var resp, err = getHTTPResponse("GET", connectionDetails, nil, client)
		if err != nil {
			err = handleHTTPError(resp, err, "Could not pull the Repository / Software Component "+repositoryName, connectionDetails)
			return "", err
		}
		defer resp.Body.Close()

		// Parse response
		var body abaputils.PullEntity
		bodyText, _ := ioutil.ReadAll(resp.Body)
		var abapResp map[string]*json.RawMessage
		json.Unmarshal(bodyText, &abapResp)
		json.Unmarshal(*abapResp["d"], &body)
		if reflect.DeepEqual(abaputils.PullEntity{}, body) {
			log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", repositoryName).Error("Could not pull the Repository / Software Component")
			var err = errors.New("Request to ABAP System not successful")
			return "", err
		}
		status = body.Status
		log.Entry().WithField("StatusCode", resp.Status).Info("Pull Status: " + body.StatusDescription)
		if body.Status != "R" {
			printLogs(body)
			break
		}
		time.Sleep(pollIntervall)
	}

	return status, nil
}

// PrintLogs sorts and formats the received transport and execution log of an import
func PrintLogs(entity abaputils.PullEntity) {

	// Sort logs
	sort.SliceStable(entity.ToExecutionLog.Results, func(i, j int) bool {
		return entity.ToExecutionLog.Results[i].Index < entity.ToExecutionLog.Results[j].Index
	})

	sort.SliceStable(entity.ToTransportLog.Results, func(i, j int) bool {
		return entity.ToTransportLog.Results[i].Index < entity.ToTransportLog.Results[j].Index
	})

	log.Entry().Info("-------------------------")
	log.Entry().Info("Transport Log")
	log.Entry().Info("-------------------------")
	for _, logEntry := range entity.ToTransportLog.Results {

		log.Entry().WithField("Timestamp", convertTime(logEntry.Timestamp)).Info(logEntry.Description)
	}

	log.Entry().Info("-------------------------")
	log.Entry().Info("Execution Log")
	log.Entry().Info("-------------------------")
	for _, logEntry := range entity.ToExecutionLog.Results {
		log.Entry().WithField("Timestamp", convertTime(logEntry.Timestamp)).Info(logEntry.Description)
	}
	log.Entry().Info("-------------------------")
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
