package orchestrator

import (
	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
	"io/ioutil"
	"time"
)

type JenkinsConfigProvider struct {
	client  piperHttp.Client
	options piperHttp.ClientOptions
}

// InitOrchestratorProvider initializes the Jenkins orchestrator with credentials
func (j *JenkinsConfigProvider) InitOrchestratorProvider(settings *OrchestratorSettings) {
	j.client = piperHttp.Client{}
	j.options = piperHttp.ClientOptions{
		Username:         settings.JenkinsUser,
		Password:         settings.JenkinsToken,
		MaxRetries:       3,
		TransportTimeout: time.Second * 10,
	}
	j.client.SetOptions(j.options)
	log.Entry().Debug("Successfully initialized Jenkins config provider")
}

// OrchestratorVersion returns the orchestrator version currently running on
func (j *JenkinsConfigProvider) OrchestratorVersion() string {
	return getEnv("JENKINS_VERSION", "n/a")
}

// OrchestratorType returns the orchestrator type Jenkins
func (j *JenkinsConfigProvider) OrchestratorType() string {
	return "Jenkins"
}

func (j *JenkinsConfigProvider) getAPIInformation() map[string]interface{} {
	URL := j.GetBuildUrl() + "api/json"

	response, err := j.client.GetRequest(URL, nil, nil)
	if err != nil {
		log.Entry().WithError(err).Error("could not get api information from Jenkins")
		return map[string]interface{}{}
	}

	if response.StatusCode != 200 { //http.StatusNoContent
		log.Entry().Errorf("Response-Code is %v . \n Could not get timestamp from Jenkins. Setting timestamp to 1970.", response.StatusCode)
		return map[string]interface{}{}
	}
	var responseInterface map[string]interface{}
	err = piperHttp.ParseHTTPResponseBodyJSON(response, &responseInterface)
	if err != nil {
		log.Entry().Error(err)
		return map[string]interface{}{}
	}
	return responseInterface
}

// GetBuildStatus returns build status of the current job
func (j *JenkinsConfigProvider) GetBuildStatus() string {
	responseInterface := j.getAPIInformation()

	if val, ok := responseInterface["result"]; ok {
		// cases in ADO: succeeded, failed, canceled, none, partiallySucceeded
		switch result := val; result {
		case "SUCCESS":
			return "SUCCESS"
		case "ABORTED":
			return "ABORTED"
		default:
			// FAILURE, NOT_BUILT
			return "FAILURE"
		}
	}

	return "FAILURE"
}

// GetLog returns the logfile from the current job as byte object
func (j *JenkinsConfigProvider) GetLog() ([]byte, error) {
	URL := j.GetBuildUrl() + "consoleText"

	response, err := j.client.GetRequest(URL, nil, nil)
	if err != nil {
		return []byte{}, errors.Wrapf(err, "Could not read Jenkins logfile. %v", err)
	}
	if response.StatusCode != 200 {
		log.Entry().Error("Could not get log information from Jenkins. Returning with empty log.")
		return []byte{}, nil
	}
	defer response.Body.Close()

	logFile, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return []byte{}, errors.Wrapf(err, "could not read Jenkins logfile from request. %v", err)
	}

	return logFile, nil
}

// GetPipelineStartTime returns the pipeline start time in UTC
func (j *JenkinsConfigProvider) GetPipelineStartTime() time.Time {
	URL := j.GetBuildUrl() + "api/json"

	response, err := j.client.GetRequest(URL, nil, nil)
	if err != nil {
		log.Entry().Error(err)
	}

	if response.StatusCode != 200 { //http.StatusNoContent -> also empty log!
		log.Entry().Errorf("Response-Code is %v . \n Could not get timestamp from Jenkins. Setting timestamp to 1970.", response.StatusCode)
		return time.Time{}.UTC()
	}
	var responseInterface map[string]interface{}
	err = piperHttp.ParseHTTPResponseBodyJSON(response, &responseInterface)
	if err != nil {
		log.Entry().WithError(err).Infof("could not parse http response, returning 1970")
		return time.Time{}.UTC()
	}

	rawTimeStamp := responseInterface["timestamp"].(float64)
	timeStamp := time.Unix(int64(rawTimeStamp)/1000, 0)

	log.Entry().Debugf("Pipeline start time: %v", timeStamp.String())
	defer response.Body.Close()
	return timeStamp.UTC()
}

// GetJobName returns the job name of the current job e.g. foo/bar/BRANCH
func (j *JenkinsConfigProvider) GetJobName() string {
	return getEnv("JOB_NAME", "n/a")
}

// GetJobUrl returns the current job URL e.g. https://JAAS.URL/job/foo/job/bar/job/main
func (j *JenkinsConfigProvider) GetJobUrl() string {
	return getEnv("JOB_URL", "n/a")
}

func (j *JenkinsConfigProvider) getJenkinsHome() string {
	return getEnv("JENKINS_HOME", "n/a")
}

// GetBuildID returns the build ID of the current job, e.g. 1234
func (j *JenkinsConfigProvider) GetBuildID() string {
	return getEnv("BUILD_ID", "n/a")
}

// GetStageName returns the stage name the job is currently in, e.g. Promote
func (j *JenkinsConfigProvider) GetStageName() string {
	return getEnv("STAGE_NAME", "n/a")
}

// GetBranch returns the branch name, only works with the git plugin enabled
func (j *JenkinsConfigProvider) GetBranch() string {
	return getEnv("BRANCH_NAME", "n/a")
}

// GetBuildUrl returns the build url, e.g. https://JAAS.URL/job/foo/job/bar/job/main/1234/
func (j *JenkinsConfigProvider) GetBuildUrl() string {
	return getEnv("BUILD_URL", "n/a")
}

// GetCommit returns the commit SHA from the current build, only works with the git plugin enabled
func (j *JenkinsConfigProvider) GetCommit() string {
	return getEnv("GIT_COMMIT", "n/a")
}

// GetRepoUrl returns the repo URL of the current build, only works with the git plugin enabled
func (j *JenkinsConfigProvider) GetRepoUrl() string {
	return getEnv("GIT_URL", "n/a")
}

// GetPullRequestConfig returns the pull request config
func (j *JenkinsConfigProvider) GetPullRequestConfig() PullRequestConfig {
	return PullRequestConfig{
		Branch: getEnv("CHANGE_BRANCH", "n/a"),
		Base:   getEnv("CHANGE_TARGET", "n/a"),
		Key:    getEnv("CHANGE_ID", "n/a"),
	}
}

// IsPullRequest returns boolean indicating if current job is a PR
func (j *JenkinsConfigProvider) IsPullRequest() bool {
	return truthy("CHANGE_ID")
}

func isJenkins() bool {
	envVars := []string{"JENKINS_HOME", "JENKINS_URL"}
	return areIndicatingEnvVarsSet(envVars)
}
