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
		pullEntity, responseStatus, err := GetPullStatus(repositoryName, connectionDetails, client)
		if err != nil {
			return status, err
		}
		status = pullEntity.Status
		log.Entry().WithField("StatusCode", responseStatus).Info("Pull Status: " + pullEntity.StatusDescription)
		if pullEntity.Status != "R" {

			if pullEntity.Status == "E" {
				log.SetErrorCategory(log.ErrorUndefined)
				PrintLegacyLogs(repositoryName, connectionDetails, client, true)
			} else {
				PrintLegacyLogs(repositoryName, connectionDetails, client, false)
			}
			break
		}
		time.Sleep(pollIntervall)
	}
	return status, nil
}

func GetPullStatus(repositoryName string, connectionDetails ConnectionDetailsHTTP, client piperhttp.Sender) (body PullEntity, status string, err error) {
	resp, err := GetHTTPResponse("GET", connectionDetails, nil, client)
	if err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		err = HandleHTTPError(resp, err, "Could not pull the Repository / Software Component "+repositoryName, connectionDetails)
		return body, resp.Status, err
	}
	defer resp.Body.Close()

	// Parse response
	var abapResp map[string]*json.RawMessage
	bodyText, _ := ioutil.ReadAll(resp.Body)

	json.Unmarshal(bodyText, &abapResp)
	json.Unmarshal(*abapResp["d"], &body)

	if reflect.DeepEqual(PullEntity{}, body) {
		log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", repositoryName).Error("Could not pull the Repository / Software Component")
		log.SetErrorCategory(log.ErrorInfrastructure)
		var err = errors.New("Request to ABAP System not successful")
		return body, resp.Status, err
	}
	return body, resp.Status, nil
}

func PrintLogs(repositoryName string, connectionDetails ConnectionDetailsHTTP, client piperhttp.Sender, errorOnSystem bool) (LogDoesNotExist error) {
	connectionDetails.URL = connectionDetails.URL + "?$expand=to_Log_Overview"
	entity, _, err := GetPullStatus(repositoryName, connectionDetails, client)
	if err != nil {
		return err
	}

	// Sort logs
	sort.SliceStable(entity.ToExecutionLog.Results, func(i, j int) bool {
		return entity.ToExecutionLog.Results[i].Index < entity.ToExecutionLog.Results[j].Index
	})

	return nil
}

// PrintLegacyLogs sorts and formats the received transport and execution log of an import; Deprecated with SAP BTP, ABAP Environment release 2205
func PrintLegacyLogs(repositoryName string, connectionDetails ConnectionDetailsHTTP, client piperhttp.Sender, errorOnSystem bool) {

	connectionDetails.URL = connectionDetails.URL + "?$expand=to_Transport_Log,to_Execution_Log"
	entity, _, err := GetPullStatus(repositoryName, connectionDetails, client)
	if err != nil {
		return
	}
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
		log.SetErrorCategory(log.ErrorConfiguration)
		return repositories, fmt.Errorf("Failed to read repository configuration: %w", errors.New("Eror in configuration, most likely you have entered empty or wrong configuration values. Please make sure that you have correctly specified them. For more information please read the User documentation"))
	}
	if config.RepositoryName == "" && config.BranchName == "" && config.Repositories == "" && len(config.RepositoryNames) == 0 {
		log.SetErrorCategory(log.ErrorConfiguration)
		return repositories, fmt.Errorf("Failed to read repository configuration: %w", errors.New("You have not specified any repository configuration. Please make sure that you have correctly specified it. For more information please read the User documentation"))
	}
	if config.Repositories != "" {
		descriptor, err := ReadAddonDescriptor(config.Repositories)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return repositories, err
		}
		err = CheckAddonDescriptorForRepositories(descriptor)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
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

func (repo *Repository) GetRequestBodyForCommitOrTag() (requestBodyString string) {
	if repo.CommitID != "" {
		requestBodyString = `, "commit_id":"` + repo.CommitID + `"`
	} else if repo.Tag != "" {
		requestBodyString = `, "tag_name":"` + repo.Tag + `"`
	}
	return requestBodyString
}

func (repo *Repository) GetLogStringForCommitOrTag() (logString string) {
	if repo.CommitID != "" {
		logString = ", commit '" + repo.CommitID + "'"
	} else if repo.Tag != "" {
		logString = ", tag '" + repo.Tag + "'"
	}
	return logString
}

func (repo *Repository) GetCloneRequestBody() (body string) {
	if repo.CommitID != "" && repo.Tag != "" {
		log.Entry().WithField("Tag", repo.Tag).WithField("Commit ID", repo.CommitID).Info("The commit ID takes precedence over the tag")
	}
	requestBodyString := repo.GetRequestBodyForCommitOrTag()
	body = `{"sc_name":"` + repo.Name + `", "branch_name":"` + repo.Branch + `"` + requestBodyString + `}`
	return body
}

func (repo *Repository) GetCloneLogString() (logString string) {
	commitOrTag := repo.GetLogStringForCommitOrTag()
	logString = "repository / software component '" + repo.Name + "', branch '" + repo.Branch + "'" + commitOrTag
	return logString
}

func (repo *Repository) GetPullRequestBody() (body string) {
	if repo.CommitID != "" && repo.Tag != "" {
		log.Entry().WithField("Tag", repo.Tag).WithField("Commit ID", repo.CommitID).Info("The commit ID takes precedence over the tag")
	}
	requestBodyString := repo.GetRequestBodyForCommitOrTag()
	body = `{"sc_name":"` + repo.Name + `"` + requestBodyString + `}`
	return body
}

func (repo *Repository) GetPullLogString() (logString string) {
	commitOrTag := repo.GetLogStringForCommitOrTag()
	logString = "repository / software component '" + repo.Name + "'" + commitOrTag
	return logString
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
	ToLogOverview     AbapLogsV2   `json:"to_Log_Overview"`
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
	UUID              string       `json:"uuid"`
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

type AbapLogsV2 struct {
	Results []LogResultsV2 `json:"results"`
}

type LogResultsV2 struct {
	Metadata AbapMetadata `json:"__metadata"`
	Index    int          `json:"log_index"`
	Name     string       `json:"log_name"`
	Status   string       `json:"type_of_found_issues"`
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
