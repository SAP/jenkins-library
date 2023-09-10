package abaputils

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
)

const failureMessageClonePull = "Could not pull the Repository / Software Component "
const numberOfEntriesPerPage = 100000
const logOutputStatusLength = 10
const logOutputTimestampLength = 29

// PollEntity periodically polls the pull/import entity to get the status. Check if the import is still running
func PollEntity(repositoryName string, connectionDetails ConnectionDetailsHTTP, client piperhttp.Sender, pollIntervall time.Duration) (string, error) {

	log.Entry().Info("Start polling the status...")
	var status string = "R"

	for {
		pullEntity, responseStatus, err := GetStatus(failureMessageClonePull+repositoryName, connectionDetails, client)
		if err != nil {
			return status, err
		}
		status = pullEntity.Status
		log.Entry().WithField("StatusCode", responseStatus).Info("Status: " + pullEntity.StatusDescription)
		if pullEntity.Status != "R" {

			PrintLogs(repositoryName, connectionDetails, client)
			break
		}
		time.Sleep(pollIntervall)
	}
	return status, nil
}

func PrintLogs(repositoryName string, connectionDetails ConnectionDetailsHTTP, client piperhttp.Sender) {
	connectionDetails.URL = connectionDetails.URL + "?$expand=to_Log_Overview"
	entity, _, err := GetStatus(failureMessageClonePull+repositoryName, connectionDetails, client)
	if err != nil || len(entity.ToLogOverview.Results) == 0 {
		// return if no logs are available
		return
	}

	// Sort logs
	sort.SliceStable(entity.ToLogOverview.Results, func(i, j int) bool {
		return entity.ToLogOverview.Results[i].Index < entity.ToLogOverview.Results[j].Index
	})

	printOverview(entity)

	// Print Details
	for _, logEntryForDetails := range entity.ToLogOverview.Results {
		printLog(logEntryForDetails, connectionDetails, client)
	}
	AddDefaultDashedLine()

	return
}

func printOverview(entity PullEntity) {

	logOutputPhaseLength, logOutputLineLength := calculateLenghts(entity)

	log.Entry().Infof("\n")

	printDashedLine(logOutputLineLength)

	log.Entry().Infof("| %-"+fmt.Sprint(logOutputPhaseLength)+"s | %"+fmt.Sprint(logOutputStatusLength)+"s | %-"+fmt.Sprint(logOutputTimestampLength)+"s |", "Phase", "Status", "Timestamp")

	printDashedLine(logOutputLineLength)

	for _, logEntry := range entity.ToLogOverview.Results {
		log.Entry().Infof("| %-"+fmt.Sprint(logOutputPhaseLength)+"s | %"+fmt.Sprint(logOutputStatusLength)+"s | %-"+fmt.Sprint(logOutputTimestampLength)+"s |", logEntry.Name, logEntry.Status, ConvertTime(logEntry.Timestamp))
	}
	printDashedLine(logOutputLineLength)
}

func calculateLenghts(entity PullEntity) (int, int) {
	phaseLength := 22
	for _, logEntry := range entity.ToLogOverview.Results {
		if l := len(logEntry.Name); l > phaseLength {
			phaseLength = l
		}
	}

	lineLength := 10 + phaseLength + logOutputStatusLength + logOutputTimestampLength
	return phaseLength, lineLength
}

func printDashedLine(i int) {
	log.Entry().Infof(strings.Repeat("-", i))
}

func printLog(logOverviewEntry LogResultsV2, connectionDetails ConnectionDetailsHTTP, client piperhttp.Sender) {

	page := 0

	printHeader(logOverviewEntry)

	for {
		connectionDetails.URL = logOverviewEntry.ToLogProtocol.Deferred.URI + getLogProtocolQuery(page)
		entity, err := GetProtocol(failureMessageClonePull, connectionDetails, client)

		printLogProtocolEntries(logOverviewEntry, entity)

		page += 1
		if allLogsHaveBeenPrinted(entity, page, err) {
			break
		}
	}

}

func printLogProtocolEntries(logEntry LogResultsV2, entity LogProtocolResults) {

	sort.SliceStable(entity.Results, func(i, j int) bool {
		return entity.Results[i].ProtocolLine < entity.Results[j].ProtocolLine
	})

	if logEntry.Status != `Success` {
		for _, entry := range entity.Results {
			log.Entry().Info(entry.Description)
		}

	} else {
		for _, entry := range entity.Results {
			log.Entry().Debug(entry.Description)
		}
	}
}

func allLogsHaveBeenPrinted(entity LogProtocolResults, page int, err error) bool {
	allPagesHaveBeenRead := false
	numberOfProtocols, errConversion := strconv.Atoi(entity.Count)
	if errConversion == nil {
		allPagesHaveBeenRead = numberOfProtocols <= page*numberOfEntriesPerPage
	}
	return (err != nil || allPagesHaveBeenRead || reflect.DeepEqual(entity.Results, LogProtocolResults{}))
}

func printHeader(logEntry LogResultsV2) {
	if logEntry.Status != `Success` {
		log.Entry().Infof("\n")
		AddDefaultDashedLine()
		log.Entry().Infof("%s (%v)", logEntry.Name, ConvertTime(logEntry.Timestamp))
		AddDefaultDashedLine()
	} else {
		log.Entry().Debugf("\n")
		AddDebugDashedLine()
		log.Entry().Debugf("%s (%v)", logEntry.Name, ConvertTime(logEntry.Timestamp))
		AddDebugDashedLine()
	}
}

func getLogProtocolQuery(page int) string {
	skip := page * numberOfEntriesPerPage
	top := numberOfEntriesPerPage

	return fmt.Sprintf("?$skip=%s&$top=%s&$inlinecount=allpages", fmt.Sprint(skip), fmt.Sprint(top))
}

func GetStatus(failureMessage string, connectionDetails ConnectionDetailsHTTP, client piperhttp.Sender) (body PullEntity, status string, err error) {
	resp, err := GetHTTPResponse("GET", connectionDetails, nil, client)
	if err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		err = HandleHTTPError(resp, err, failureMessage, connectionDetails)
		if resp != nil {
			status = resp.Status
		}
		return body, status, err
	}
	defer resp.Body.Close()

	// Parse response
	var abapResp map[string]*json.RawMessage
	bodyText, _ := io.ReadAll(resp.Body)

	marshallError := json.Unmarshal(bodyText, &abapResp)
	if marshallError != nil {
		return body, status, errors.Wrap(marshallError, "Could not parse response from the ABAP Environment system")
	}
	marshallError = json.Unmarshal(*abapResp["d"], &body)
	if marshallError != nil {
		return body, status, errors.Wrap(marshallError, "Could not parse response from the ABAP Environment system")
	}

	if reflect.DeepEqual(PullEntity{}, body) {
		log.Entry().WithField("StatusCode", resp.Status).Error(failureMessage)
		log.SetErrorCategory(log.ErrorInfrastructure)
		var err = errors.New("Request to ABAP System not successful")
		return body, resp.Status, err
	}
	return body, resp.Status, nil
}

func GetProtocol(failureMessage string, connectionDetails ConnectionDetailsHTTP, client piperhttp.Sender) (body LogProtocolResults, err error) {
	resp, err := GetHTTPResponse("GET", connectionDetails, nil, client)
	if err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		err = HandleHTTPError(resp, err, failureMessage, connectionDetails)
		return body, err
	}
	defer resp.Body.Close()

	// Parse response
	var abapResp map[string]*json.RawMessage
	bodyText, _ := io.ReadAll(resp.Body)

	marshallError := json.Unmarshal(bodyText, &abapResp)
	if marshallError != nil {
		return body, errors.Wrap(marshallError, "Could not parse response from the ABAP Environment system")
	}
	marshallError = json.Unmarshal(*abapResp["d"], &body)
	if marshallError != nil {
		return body, errors.Wrap(marshallError, "Could not parse response from the ABAP Environment system")
	}

	return body, nil
}

// GetRepositories for parsing  one or multiple branches and repositories from repositories file or branchName and repositoryName configuration
func GetRepositories(config *RepositoriesConfig, branchRequired bool) ([]Repository, error) {
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
	if config.RepositoryName != "" && !branchRequired {
		repositories = append(repositories, Repository{Name: config.RepositoryName, CommitID: config.CommitID})
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
	Metadata      AbapMetadata        `json:"__metadata"`
	Index         int                 `json:"log_index"`
	Name          string              `json:"log_name"`
	Status        string              `json:"type_of_found_issues"`
	Timestamp     string              `json:"timestamp"`
	ToLogProtocol LogProtocolDeferred `json:"to_Log_Protocol"`
}

type LogProtocolDeferred struct {
	Deferred URI `json:"__deferred"`
}

type URI struct {
	URI string `json:"uri"`
}

type LogProtocolResults struct {
	Results []LogProtocol `json:"results"`
	Count   string        `json:"__count"`
}

type LogProtocol struct {
	Metadata      AbapMetadata `json:"__metadata"`
	OverviewIndex int          `json:"log_index"`
	ProtocolLine  int          `json:"index_no"`
	Type          string       `json:"type"`
	Description   string       `json:"descr"`
	Timestamp     string       `json:"timestamp"`
}

// LogResults struct for Execution and Transport Log entities A4C_A2G_GHA_SC_LOG_EXE and A4C_A2G_GHA_SC_LOG_TP
type LogResults struct {
	Index       string `json:"index_no"`
	Type        string `json:"type"`
	Description string `json:"descr"`
	Timestamp   string `json:"timestamp"`
}

// RepositoriesConfig struct for parsing one or multiple branches and repositories configurations
type RepositoriesConfig struct {
	BranchName      string
	CommitID        string
	RepositoryName  string
	RepositoryNames []string
	Repositories    string
}

type EntitySetsForManageGitRepository struct {
	EntitySets []string `json:"EntitySets"`
}
