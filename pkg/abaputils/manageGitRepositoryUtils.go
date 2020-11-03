package abaputils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"sort"
	"time"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
)

// PollEntity periodically polls the pull/import entity to get the status. Check if the import is still running
func PollEntity(repositoryName string, connectionDetails ConnectionDetailsHTTP, client piperhttp.Sender, pollIntervall time.Duration) (string, error) {

	log.Entry().Info("Start polling the status...")
	var status string = "R"

	for {
		var resp, err = GetHTTPResponse("GET", connectionDetails, nil, client)
		if err != nil {
			err = HandleHTTPError(resp, err, "Could not pull the Repository / Software Component "+repositoryName, connectionDetails)
			return "", err
		}
		defer resp.Body.Close()

		// Parse response
		var abapResp map[string]*json.RawMessage
		var body PullEntity
		bodyText, _ := ioutil.ReadAll(resp.Body)

		json.Unmarshal(bodyText, &abapResp)
		json.Unmarshal(*abapResp["d"], &body)

		if reflect.DeepEqual(PullEntity{}, body) {
			log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", repositoryName).Error("Could not pull the Repository / Software Component")
			var err = errors.New("Request to ABAP System not successful")
			return "", err
		}

		status = body.Status
		log.Entry().WithField("StatusCode", resp.Status).Info("Pull Status: " + body.StatusDescription)
		if body.Status != "R" {
			if body.Status == "E" {
				PrintLogs(body, true)
			} else {
				PrintLogs(body, false)
			}
			break
		}
		time.Sleep(pollIntervall)
	}
	return status, nil
}

// PrintLogs sorts and formats the received transport and execution log of an import
func PrintLogs(entity PullEntity, errorOnSystem bool) {

	// Sort logs
	sort.SliceStable(entity.ToExecutionLog.Results, func(i, j int) bool {
		return entity.ToExecutionLog.Results[i].Index < entity.ToExecutionLog.Results[j].Index
	})

	sort.SliceStable(entity.ToTransportLog.Results, func(i, j int) bool {
		return entity.ToTransportLog.Results[i].Index < entity.ToTransportLog.Results[j].Index
	})

	// Show transport and execution log if either the action was erroenous on the system or the log level is set to "debug" (verbose = true)
	if errorOnSystem {
		log.Entry().Info("-------------------------")
		log.Entry().Info("Transport Log")
		log.Entry().Info("-------------------------")
		for _, logEntry := range entity.ToTransportLog.Results {

			log.Entry().WithField("Timestamp", ConvertTime(logEntry.Timestamp)).Info(logEntry.Description)
		}

		log.Entry().Info("-------------------------")
		log.Entry().Info("Execution Log")
		log.Entry().Info("-------------------------")
		for _, logEntry := range entity.ToExecutionLog.Results {
			log.Entry().WithField("Timestamp", ConvertTime(logEntry.Timestamp)).Info(logEntry.Description)
		}
		log.Entry().Info("-------------------------")
	} else {
		log.Entry().Debug("-------------------------")
		log.Entry().Debug("Transport Log")
		log.Entry().Debug("-------------------------")
		for _, logEntry := range entity.ToTransportLog.Results {

			log.Entry().WithField("Timestamp", ConvertTime(logEntry.Timestamp)).Debug(logEntry.Description)
		}

		log.Entry().Debug("-------------------------")
		log.Entry().Debug("Execution Log")
		log.Entry().Debug("-------------------------")
		for _, logEntry := range entity.ToExecutionLog.Results {
			log.Entry().WithField("Timestamp", ConvertTime(logEntry.Timestamp)).Debug(logEntry.Description)
		}
		log.Entry().Debug("-------------------------")
	}

}

//GetRepositories for parsing  one or multiple branches and repositories from repositories file or branchName and repositoryName configuration
func GetRepositories(config *RepositoriesConfig) ([]Repository, error) {
	var repositories = make([]Repository, 0)
	if reflect.DeepEqual(RepositoriesConfig{}, config) {
		return repositories, fmt.Errorf("Failed to read repository configuration: %w", errors.New("Eror in configuration, most likely you have entered empty or wrong configuration values. Please make sure that you have correctly specified them. For more information please read the User documentation"))
	}
	if config.RepositoryName == "" && config.BranchName == "" && config.Repositories == "" && len(config.RepositoryNames) == 0 {
		return repositories, fmt.Errorf("Failed to read repository configuration: %w", errors.New("You have not specified any repository configuration. Please make sure that you have correctly specified it. For more information please read the User documentation"))
	}
	if config.Repositories != "" {
		descriptor, err := ReadAddonDescriptor(config.Repositories)
		if err != nil {
			return repositories, err
		}
		err = CheckAddonDescriptorForRepositories(descriptor)
		if err != nil {
			return repositories, fmt.Errorf("Error in config file %v, %w", config.Repositories, err)
		}
		repositories = descriptor.Repositories
	}
	if config.RepositoryName != "" && config.BranchName != "" {
		repositories = append(repositories, Repository{Name: config.RepositoryName, Branch: config.BranchName})
	}
	if len(config.RepositoryNames) > 0 {
		for _, repository := range config.RepositoryNames {
			repositories = append(repositories, Repository{Name: repository})
		}
	}
	return repositories, nil
}

//GetCommitStrings for getting the commit_id property for the http request and a string for logging output
func GetCommitStrings(commitID string) (commitQuery string, commitString string) {
	if commitID != "" {
		commitQuery = `, "commit_id":"` + commitID + `"`
		commitString = ", commit '" + commitID + "'"
	}
	return commitQuery, commitString
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

// CloneEntity struct for the Clone entity A4C_A2G_GHA_SC_CLONE
type CloneEntity struct {
	Metadata          AbapMetadata `json:"__metadata"`
	ScName            string       `json:"sc_name"`
	BranchName        string       `json:"branch_name"`
	ImportType        string       `json:"import_type"`
	Namespace         string       `json:"namepsace"`
	Status            string       `json:"status"`
	StatusDescription string       `json:"status_descr"`
	StartedByUser     string       `json:"user_name"`
	StartTime         string       `json:"start_time"`
	ChangeTime        string       `json:"change_time"`
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

//RepositoriesConfig struct for parsing one or multiple branches and repositories configurations
type RepositoriesConfig struct {
	BranchName      string
	RepositoryName  string
	RepositoryNames []string
	Repositories    string
}
