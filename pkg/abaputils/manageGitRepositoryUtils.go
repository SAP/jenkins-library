package abaputils

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"errors"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

const numberOfEntriesPerPage = 100000
const logOutputStatusLength = 10
const logOutputTimestampLength = 29

// Specifies which output option is used for logs
type LogOutputManager struct {
	LogOutput    string
	PiperStep    string
	FileNameStep string
	StepReports  []piperutils.Path
}

func PersistArchiveLogsForPiperStep(logOutputManager *LogOutputManager) {
	fileUtils := piperutils.Files{}
	switch logOutputManager.PiperStep {
	case "clone":
		piperutils.PersistReportsAndLinks("abapEnvironmentCloneGitRepo", "", fileUtils, logOutputManager.StepReports, nil)
	case "pull":
		piperutils.PersistReportsAndLinks("abapEnvironmentPullGitRepo", "", fileUtils, logOutputManager.StepReports, nil)
	case "checkoutBranch":
		piperutils.PersistReportsAndLinks("abapEnvironmentCheckoutBranch", "", fileUtils, logOutputManager.StepReports, nil)
	default:
		log.Entry().Info("Cannot save log archive because no piper step was defined.")
	}
}

// PollEntity periodically polls the action entity to get the status. Check if the import is still running
func PollEntity(api SoftwareComponentApiInterface, pollIntervall time.Duration, logOutputManager *LogOutputManager) (string, error) {

	log.Entry().Info("Start polling the status...")
	var statusCode string = "R"
	var err error

	api.initialRequest()

	for {
		// pullEntity, responseStatus, err := api.GetStatus(failureMessageClonePull+repositoryName, connectionDetails, client)
		statusCode, err = api.GetAction()
		if err != nil {
			return statusCode, err
		}

		if statusCode != "R" && statusCode != "Q" {

			PrintLogs(api, logOutputManager)
			break
		}
		time.Sleep(pollIntervall)
	}
	return statusCode, nil
}

func PrintLogs(api SoftwareComponentApiInterface, logOutputManager *LogOutputManager) {

	// Get Execution Logs
	executionLogs, err := api.GetExecutionLog()
	if err == nil {
		printExecutionLogs(executionLogs)
	}

	results, _ := api.GetLogOverview()

	// Sort logs
	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Index < results[j].Index
	})

	printOverview(results, api)

	if logOutputManager.LogOutput == "ZIP" {
		// get zip file as byte array
		zipfile, err := api.GetLogArchive()
		// Saving logs in file and adding to piperutils to archive file
		if err == nil {
			fileName := "LogArchive-" + logOutputManager.FileNameStep + "-" + strings.Replace(api.getRepositoryName(), "/", "_", -1) + "-" + api.getUUID() + "_" + time.Now().Format("2006-01-02T15:04:05") + ".zip"

			err = os.WriteFile(fileName, zipfile, 0o644)

			if err == nil {
				log.Entry().Infof("Writing %s file was successful", fileName)
				logOutputManager.StepReports = append(logOutputManager.StepReports, piperutils.Path{Target: fileName, Name: "Log_Archive_" + api.getUUID(), Mandatory: true})
			}
		}

	} else {
		// Print Details
		if len(results) != 0 {
			for _, logEntryForDetails := range results {
				printLog(logEntryForDetails, api)
			}
		}
		AddDefaultDashedLine(1)
	}

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

func printOverview(results []LogResultsV2, api SoftwareComponentApiInterface) {

	if len(results) == 0 {
		return
	}

	logOutputPhaseLength, logOutputLineLength := calculateLenghts(results)

	log.Entry().Infof("\n")

	printDashedLine(logOutputLineLength)

	log.Entry().Infof("| %-"+fmt.Sprint(logOutputPhaseLength)+"s | %"+fmt.Sprint(logOutputStatusLength)+"s | %-"+fmt.Sprint(logOutputTimestampLength)+"s |", "Phase", "Status", "Timestamp")

	printDashedLine(logOutputLineLength)

	for _, logEntry := range results {
		log.Entry().Infof("| %-"+fmt.Sprint(logOutputPhaseLength)+"s | %"+fmt.Sprint(logOutputStatusLength)+"s | %-"+fmt.Sprint(logOutputTimestampLength)+"s |", logEntry.Name, logEntry.Status, api.ConvertTime(logEntry.Timestamp))
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
	log.Entry().Info(strings.Repeat("-", i))
}

func printLog(logOverviewEntry LogResultsV2, api SoftwareComponentApiInterface) {

	page := 0
	printHeader(logOverviewEntry, api)
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
			log.Entry().Infof("%s %s", entry.Type, entry.Description)
		}
	} else {
		for _, entry := range logProtocols {
			log.Entry().Debugf("%s %s", entry.Type, entry.Description)
		}
	}
}

func allLogsHaveBeenPrinted(protocols []LogProtocol, page int, count int, err error) bool {
	allPagesHaveBeenRead := count <= page*numberOfEntriesPerPage
	return (err != nil || allPagesHaveBeenRead || reflect.DeepEqual(protocols, []LogProtocol{}))
}

func printHeader(logEntry LogResultsV2, api SoftwareComponentApiInterface) {
	if logEntry.Status == `Error` {
		log.Entry().Infof("\n")
		AddDefaultDashedLine(1)
		log.Entry().Infof("%s (%v)", logEntry.Name, api.ConvertTime(logEntry.Timestamp))
		AddDefaultDashedLine(1)
	} else {
		log.Entry().Debugf("\n")
		AddDebugDashedLine()
		log.Entry().Debugf("%s (%v)", logEntry.Name, api.ConvertTime(logEntry.Timestamp))
		AddDebugDashedLine()
	}
}

// GetRepositories for parsing one or multiple branches and repositories from repositories file or branchName and repositoryName configuration
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

func (repo *Repository) GetRequestBodyForBYOGCredentials() (string, error) {
	var byogBodyString string

	if repo.ByogAuthMethod != "" {
		byogBodyString += `, "auth_method":"` + repo.ByogAuthMethod + `"`
	}
	if repo.ByogUsername != "" {
		byogBodyString += `, "username":"` + repo.ByogUsername + `"`
	} else {
		return "", fmt.Errorf("Failed to get BYOG credentials: %w", errors.New("Username for BYOG is missing, please provide git username to authenticate"))
	}
	if repo.ByogPassword != "" {
		byogBodyString += `, "password":"` + repo.ByogPassword + `"`
	} else {
		return "", fmt.Errorf("Failed to get BYOG credentials: %w", errors.New("Password/Token for BYOG is missing, please provide git password or token to authenticate"))
	}
	return byogBodyString, nil
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

func (repo *Repository) GetCloneRequestBody() (body string, err error) {
	if repo.CommitID != "" && repo.Tag != "" {
		log.Entry().WithField("Tag", repo.Tag).WithField("Commit ID", repo.CommitID).Info("The commit ID takes precedence over the tag")
	}
	requestBodyString := repo.GetRequestBodyForCommitOrTag()
	var byogBodyString = ""
	if repo.IsByog {
		byogBodyString, err = repo.GetRequestBodyForBYOGCredentials()
		if err != nil {
			return "", err
		}
	}

	body = `{"branch_name":"` + repo.Branch + `"` + requestBodyString + byogBodyString + `}`
	return body, nil
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
