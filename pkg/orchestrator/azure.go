package orchestrator

import (
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
)

type azureDevopsConfigProvider struct {
	client         piperHttp.Client
	apiInformation map[string]interface{}
}

func newAzureDevopsConfigProvider() *azureDevopsConfigProvider {
	return &azureDevopsConfigProvider{}
}

// Configure initializes http client for AzureDevopsConfigProvider
func (a *azureDevopsConfigProvider) Configure(opts *Options) error {
	a.client.SetOptions(piperHttp.ClientOptions{
		Username:         "",
		Password:         opts.AzureToken,
		MaxRetries:       3,
		TransportTimeout: time.Second * 10,
	})

	log.Entry().Debug("Successfully initialized Azure config provider")
	return nil
}

// fetchAPIInformation fetches Azure API information of current build
func (a *azureDevopsConfigProvider) fetchAPIInformation() {
	// if apiInformation is empty fill it otherwise do nothing
	if len(a.apiInformation) == 0 {
		log.Entry().Debugf("apiInformation is empty, getting infos from API")
		URL := a.getSystemCollectionURI() + a.getTeamProjectID() + "/_apis/build/builds/" + a.getAzureBuildID() + "/"
		log.Entry().Debugf("API URL: %s", URL)
		response, err := a.client.GetRequest(URL, nil, nil)
		if err != nil {
			log.Entry().Error("failed to get HTTP response, returning empty API information", err)
			a.apiInformation = map[string]interface{}{}
			return
		} else if response.StatusCode != 200 { //http.StatusNoContent
			log.Entry().Errorf("response code is %v, could not get API information from AzureDevOps. Returning with empty interface.", response.StatusCode)
			a.apiInformation = map[string]interface{}{}
			return
		}

		err = piperHttp.ParseHTTPResponseBodyJSON(response, &a.apiInformation)
		if err != nil {
			log.Entry().Error("failed to parse HTTP response, returning with empty interface", err)
			a.apiInformation = map[string]interface{}{}
			return
		}
		log.Entry().Debugf("successfully retrieved apiInformation")
	} else {
		log.Entry().Debugf("apiInformation already set")
	}
}

func (a *azureDevopsConfigProvider) ChangeSets() []ChangeSet {
	log.Entry().Warn("ChangeSets for AzureDevOps not yet implemented")
	return []ChangeSet{}
}

// getSystemCollectionURI returns the URI of the TFS collection or Azure DevOps organization e.g. https://dev.azure.com/fabrikamfiber/
func (a *azureDevopsConfigProvider) getSystemCollectionURI() string {
	return getEnv("SYSTEM_COLLECTIONURI", "n/a")
}

// getTeamProjectID is the name of the project that contains this build e.g. 123a4567-ab1c-12a1-1234-123456ab7890
func (a *azureDevopsConfigProvider) getTeamProjectID() string {
	return getEnv("SYSTEM_TEAMPROJECTID", "n/a")
}

// getAzureBuildID returns the id of the build, e.g. 1234
func (a *azureDevopsConfigProvider) getAzureBuildID() string {
	// INFO: Private function only used for API requests, buildId for e.g. reporting
	// is GetBuildNumber to align with the UI of ADO
	return getEnv("BUILD_BUILDID", "n/a")
}

// JobName returns the pipeline job name, currently org/repo
func (a *azureDevopsConfigProvider) JobName() string {
	return getEnv("BUILD_REPOSITORY_NAME", "n/a")
}

// OrchestratorVersion returns the agent version on ADO
func (a *azureDevopsConfigProvider) OrchestratorVersion() string {
	return getEnv("AGENT_VERSION", "n/a")
}

// OrchestratorType returns the orchestrator name e.g. Azure/GitHubActions/Jenkins
func (a *azureDevopsConfigProvider) OrchestratorType() string {
	return OrchestratorTypeAzure
}

// BuildStatus returns status of the build. Return variables are aligned with Jenkins build statuses.
func (a *azureDevopsConfigProvider) BuildStatus() string {
	// cases to align with Jenkins: SUCCESS, FAILURE, NOT_BUILD, ABORTED
	switch buildStatus := getEnv("AGENT_JOBSTATUS", "FAILURE"); buildStatus {
	case "Succeeded":
		return BuildStatusSuccess
	case "Canceled":
		return BuildStatusAborted
	default:
		// Failed, SucceededWithIssues
		return BuildStatusFailure
	}
}

// FullLogs returns the whole logfile for the current pipeline run
func (a *azureDevopsConfigProvider) FullLogs() ([]byte, error) {
	URL := a.getSystemCollectionURI() + a.getTeamProjectID() + "/_apis/build/builds/" + a.getAzureBuildID() + "/logs"

	response, err := a.client.GetRequest(URL, nil, nil)

	if err != nil {
		log.Entry().Error("failed to get HTTP response: ", err)
		return []byte{}, err
	}
	if response.StatusCode != 200 { //http.StatusNoContent -> also empty log!
		log.Entry().Errorf("response-Code is %v, could not get log information from AzureDevOps, returning with empty log.", response.StatusCode)
		return []byte{}, nil
	}
	var responseInterface map[string]interface{}
	err = piperHttp.ParseHTTPResponseBodyJSON(response, &responseInterface)
	if err != nil {
		log.Entry().Error("failed to parse http response: ", err)
		return []byte{}, err
	}
	// check if response interface is empty or non-existent
	var logCount int
	if val, ok := responseInterface["count"]; ok {
		logCount = int(val.(float64))
	} else {
		log.Entry().Error("log count variable not found, returning empty log")
		return []byte{}, err
	}
	var logs []byte
	for i := 1; i <= logCount; i++ {
		counter := strconv.Itoa(i)
		logURL := URL + "/" + counter
		log.Entry().Debugf("Getting log no.: %d  from %v", i, logURL)
		response, err := a.client.GetRequest(logURL, nil, nil)
		if err != nil {
			log.Entry().Error("failed to get log", err)
			return []byte{}, err
		}
		if response.StatusCode != 200 { //http.StatusNoContent -> also empty log!
			log.Entry().Errorf("response code is %v, could not get log information from AzureDevOps ", response.StatusCode)
			return []byte{}, err
		}
		content, err := io.ReadAll(response.Body)
		if err != nil {
			log.Entry().Error("failed to parse http response", err)
			return []byte{}, err
		}
		logs = append(logs, content...)
	}

	return logs, nil
}

// PipelineStartTime returns the pipeline start time in UTC
func (a *azureDevopsConfigProvider) PipelineStartTime() time.Time {
	//"2022-03-18T07:30:31.1915758Z"
	a.fetchAPIInformation()
	if val, ok := a.apiInformation["startTime"]; ok {
		parsed, err := time.Parse(time.RFC3339, val.(string))
		if err != nil {
			log.Entry().Errorf("could not parse timestamp, %v", err)
			parsed = time.Time{}
		}
		return parsed.UTC()
	}
	return time.Time{}.UTC()
}

// BuildID returns the BuildNumber displayed in the ADO UI
func (a *azureDevopsConfigProvider) BuildID() string {
	// INFO: ADO has BUILD_ID and buildNumber, as buildNumber is used in the UI we return this value
	// for the buildID used only for API requests we have a private method getAzureBuildID
	// example: buildNumber: 20220318.16 buildId: 76443
	return getEnv("BUILD_BUILDNUMBER", "n/a")
}

// StageName returns the human-readable name given to a stage. e.g. "Promote" or "Init"
func (a *azureDevopsConfigProvider) StageName() string {
	return getEnv("SYSTEM_STAGEDISPLAYNAME", "n/a")
}

// Branch returns the source branch name, e.g. main
func (a *azureDevopsConfigProvider) Branch() string {
	tmp := getEnv("BUILD_SOURCEBRANCH", "n/a")
	return strings.TrimPrefix(tmp, "refs/heads/")
}

// GitReference return the git reference
func (a *azureDevopsConfigProvider) GitReference() string {
	return getEnv("BUILD_SOURCEBRANCH", "n/a")
}

// BuildURL returns the builds URL e.g. https://dev.azure.com/fabrikamfiber/your-repo-name/_build/results?buildId=1234
func (a *azureDevopsConfigProvider) BuildURL() string {
	return os.Getenv("SYSTEM_TEAMFOUNDATIONCOLLECTIONURI") + os.Getenv("SYSTEM_TEAMPROJECT") + "/" + os.Getenv("SYSTEM_DEFINITIONNAME") + "/_build/results?buildId=" + a.getAzureBuildID()
}

// JobURL returns tje current job url e.g. https://dev.azure.com/fabrikamfiber/your-repo-name/_build?definitionId=1234
func (a *azureDevopsConfigProvider) JobURL() string {
	// TODO: Check if this is the correct URL
	return os.Getenv("SYSTEM_TEAMFOUNDATIONCOLLECTIONURI") + os.Getenv("SYSTEM_TEAMPROJECT") + "/" + os.Getenv("SYSTEM_DEFINITIONNAME") + "/_build?definitionId=" + os.Getenv("SYSTEM_DEFINITIONID")
}

// CommitSHA returns commit SHA of current build
func (a *azureDevopsConfigProvider) CommitSHA() string {
	return getEnv("BUILD_SOURCEVERSION", "n/a")
}

// RepoURL returns current repo URL e.g. https://github.com/SAP/jenkins-library
func (a *azureDevopsConfigProvider) RepoURL() string {
	return getEnv("BUILD_REPOSITORY_URI", "n/a")
}

// BuildReason returns the build reason
func (a *azureDevopsConfigProvider) BuildReason() string {
	// https://docs.microsoft.com/en-us/azure/devops/pipelines/build/variables?view=azure-devops&tabs=yaml#build-variables-devops-services
	return getEnv("BUILD_REASON", "n/a")
}

// PullRequestConfig returns pull request configuration
func (a *azureDevopsConfigProvider) PullRequestConfig() PullRequestConfig {
	prKey := getEnv("SYSTEM_PULLREQUEST_PULLREQUESTID", "n/a")

	// This variable is populated for pull requests which have a different pull request ID and pull request number.
	// In this case the pull request ID will contain an internal numeric ID and the pull request number will be provided
	// as part of the 'SYSTEM_PULLREQUEST_PULLREQUESTNUMBER' environment variable.
	prNumber, prNumberEnvVarSet := os.LookupEnv("SYSTEM_PULLREQUEST_PULLREQUESTNUMBER")
	if prNumberEnvVarSet == true {
		prKey = prNumber
	}

	return PullRequestConfig{
		Branch: os.Getenv("SYSTEM_PULLREQUEST_SOURCEBRANCH"),
		Base:   os.Getenv("SYSTEM_PULLREQUEST_TARGETBRANCH"),
		Key:    prKey,
	}
}

// IsPullRequest indicates whether the current build is a PR
func (a *azureDevopsConfigProvider) IsPullRequest() bool {
	return getEnv("BUILD_REASON", "n/a") == "PullRequest"
}

func isAzure() bool {
	envVars := []string{"AZURE_HTTP_USER_AGENT"}
	return envVarsAreSet(envVars)
}
