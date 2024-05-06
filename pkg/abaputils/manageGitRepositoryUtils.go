package abaputils

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
)

const numberOfEntriesPerPage = 100000
const logOutputStatusLength = 10
const logOutputTimestampLength = 29

// PollEntity periodically polls the action entity to get the status. Check if the import is still running
func PollEntity(api SoftwareComponentApiInterface, pollIntervall time.Duration) (string, error) {

	log.Entry().Info("Start polling the status...")
	var statusCode string = "R"
	var err error

	for {
		// pullEntity, responseStatus, err := api.GetStatus(failureMessageClonePull+repositoryName, connectionDetails, client)
		statusCode, err = api.GetAction()
		if err != nil {
			return statusCode, err
		}

		if statusCode != "R" && statusCode != "Q" {

			PrintLogs(api)
			break
		}
		time.Sleep(pollIntervall)
	}
	return statusCode, nil
}

func PrintLogs(api SoftwareComponentApiInterface) {

	// Get Execution Logs
	executionLogs, err := api.GetExecutionLog()
	if err == nil {
		printExecutionLogs(executionLogs)
	}

	results, err := api.GetLogOverview()
	if err != nil || len(results) == 0 {
		// return if no logs are available
		return
	}

	// Sort logs
	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Index < results[j].Index
	})

	printOverview(results)

	// Print Details
	for _, logEntryForDetails := range results {
		printLog(logEntryForDetails, api)
	}
	AddDefaultDashedLine(1)

	return
}

func printExecutionLogs(executionLogs ExecutionLog) {
	log.Entry().Infof("\n")
	AddDefaultDashedLine(1)
	log.Entry().Infof("Execution Logs")
	AddDefaultDashedLine(1)
	for _, entry := range executionLogs.Value {
		log.Entry().Infof("%7s - %s", entry.Type, entry.Descr)
	}
	AddDefaultDashedLine(1)
}

func printOverview(results []LogResultsV2) {

	logOutputPhaseLength, logOutputLineLength := calculateLenghts(results)

	log.Entry().Infof("\n")

	printDashedLine(logOutputLineLength)

	log.Entry().Infof("| %-"+fmt.Sprint(logOutputPhaseLength)+"s | %"+fmt.Sprint(logOutputStatusLength)+"s | %-"+fmt.Sprint(logOutputTimestampLength)+"s |", "Phase", "Status", "Timestamp")

	printDashedLine(logOutputLineLength)

	for _, logEntry := range results {
		log.Entry().Infof("| %-"+fmt.Sprint(logOutputPhaseLength)+"s | %"+fmt.Sprint(logOutputStatusLength)+"s | %-"+fmt.Sprint(logOutputTimestampLength)+"s |", logEntry.Name, logEntry.Status, ConvertTime(logEntry.Timestamp))
	}
	printDashedLine(logOutputLineLength)
}

func calculateLenghts(results []LogResultsV2) (int, int) {
	phaseLength := 22
	for _, logEntry := range results {
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

func printLog(logOverviewEntry LogResultsV2, api SoftwareComponentApiInterface) {

	page := 0
	printHeader(logOverviewEntry)
	for {
		logProtocols, count, err := api.GetLogProtocol(logOverviewEntry, page)
		printLogProtocolEntries(logOverviewEntry, logProtocols)
		page += 1
		if allLogsHaveBeenPrinted(logProtocols, page, count, err) {
			break
		}
	}
}

func printLogProtocolEntries(logEntry LogResultsV2, logProtocols []LogProtocol) {

	sort.SliceStable(logProtocols, func(i, j int) bool {
		return logProtocols[i].ProtocolLine < logProtocols[j].ProtocolLine
	})
	if logEntry.Status == `Error` {
		for _, entry := range logProtocols {
			log.Entry().Info(entry.Description)
		}
	} else {
		for _, entry := range logProtocols {
			log.Entry().Debug(entry.Description)
		}
	}
}

func allLogsHaveBeenPrinted(protocols []LogProtocol, page int, count int, err error) bool {
	allPagesHaveBeenRead := count <= page*numberOfEntriesPerPage
	return (err != nil || allPagesHaveBeenRead || reflect.DeepEqual(protocols, []LogProtocol{}))
}

func printHeader(logEntry LogResultsV2) {
	if logEntry.Status != `Success` {
		log.Entry().Infof("\n")
		AddDefaultDashedLine(1)
		log.Entry().Infof("%s (%v)", logEntry.Name, ConvertTime(logEntry.Timestamp))
		AddDefaultDashedLine(1)
	} else {
		log.Entry().Debugf("\n")
		AddDebugDashedLine()
		log.Entry().Debugf("%s (%v)", logEntry.Name, ConvertTime(logEntry.Timestamp))
		AddDebugDashedLine()
	}
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

func (repo *Repository) GetCloneRequestBodyWithSWC() (body string) {
	if repo.CommitID != "" && repo.Tag != "" {
		log.Entry().WithField("Tag", repo.Tag).WithField("Commit ID", repo.CommitID).Info("The commit ID takes precedence over the tag")
	}
	requestBodyString := repo.GetRequestBodyForCommitOrTag()
	body = `{"sc_name":"` + repo.Name + `", "branch_name":"` + repo.Branch + `"` + requestBodyString + `}`
	return body
}

func (repo *Repository) GetCloneRequestBody() (body string) {
	if repo.CommitID != "" && repo.Tag != "" {
		log.Entry().WithField("Tag", repo.Tag).WithField("Commit ID", repo.CommitID).Info("The commit ID takes precedence over the tag")
	}
	requestBodyString := repo.GetRequestBodyForCommitOrTag()
	body = `{"branch_name":"` + repo.Branch + `"` + requestBodyString + `}`
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

func (repo *Repository) GetPullActionRequestBody() (body string) {
	return `{` + `"commit_id":"` + repo.CommitID + `", ` + `"tag_name":"` + repo.Tag + `"` + `}`
}

func (repo *Repository) GetPullLogString() (logString string) {
	commitOrTag := repo.GetLogStringForCommitOrTag()
	logString = "repository / software component '" + repo.Name + "'" + commitOrTag
	return logString
}
