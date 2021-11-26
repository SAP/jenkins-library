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

func (j *JenkinsConfigProvider) InitOrchestratorProvider(settings *OrchestratorSettings) {
	j.client = piperHttp.Client{}
	j.options = piperHttp.ClientOptions{
		Username: settings.JenkinsUser,
		Password: settings.JenkinsToken,
	}
	j.client.SetOptions(j.options)
	log.Entry().Debug("Successfully initialized Jenkins config provider")
}

func (j *JenkinsConfigProvider) OrchestratorVersion() string {
	return getEnv("JENKINS_VERSION", "n/a")
}

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

// GetBuildInformation
func (j *JenkinsConfigProvider) GetBuildStatus() string {
	responseInterface := j.getAPIInformation()

	if val, ok := responseInterface["result"]; ok {
		// cases in ADO: succeeded, failed, canceled, none, partiallySucceeded
		switch result := responseInterface["result"]; result {
		case "SUCCESS":
			return "SUCCESS"
		case "ABORTED":
			return "ABORTED"
		default:
			// FAILURE, NOT_BUILT
			return "FAILURE"
		}
		return val.(string)
	}

	return "FAILURE"
}

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

func (j *JenkinsConfigProvider) GetPipelineStartTime() time.Time {
	URL := j.GetBuildUrl() + "api/json"

	response, err := j.client.GetRequest(URL, nil, nil)
	if err != nil {
		log.Entry().Error(err)
	}

	if response.StatusCode != 200 { //http.StatusNoContent -> also empty log!
		log.Entry().Errorf("Response-Code is %v . \n Could not get timestamp from Jenkins. Setting timestamp to 1970.", response.StatusCode)
		return time.Unix(1, 0)
	}
	var responseInterface map[string]interface{}
	err = piperHttp.ParseHTTPResponseBodyJSON(response, &responseInterface)
	if err != nil {
		log.Entry().Error(err)
	}

	rawTimeStamp := responseInterface["timestamp"].(float64)
	timeStamp := time.Unix(int64(rawTimeStamp)/1000, 0)

	log.Entry().Debugf("Pipeline start time: %v", timeStamp.String())
	defer response.Body.Close()
	return timeStamp
}

func (j *JenkinsConfigProvider) GetJobName() string {
	return getEnv("JOB_NAME", "n/a")
}

func (j *JenkinsConfigProvider) GetJobUrl() string {
	return getEnv("JOB_URL", "n/a")
}

func (j *JenkinsConfigProvider) getJenkinsHome() string {
	return getEnv("JENKINS_HOME", "n/a")
}

func (j *JenkinsConfigProvider) GetBuildId() string {
	return getEnv("BUILD_ID", "n/a")
}

func (a *JenkinsConfigProvider) GetStageName() string {
	return getEnv("STAGE_NAME", "n/a")
}

func (j *JenkinsConfigProvider) GetBranch() string {
	return getEnv("GIT_BRANCH", "n/a")
}

func (j *JenkinsConfigProvider) GetBuildUrl() string {
	return getEnv("BUILD_URL", "n/a")
}

func (j *JenkinsConfigProvider) GetCommit() string {
	return getEnv("GIT_COMMIT", "n/a")
}

func (j *JenkinsConfigProvider) GetRepoUrl() string {
	return getEnv("GIT_URL", "n/a")
}

func (j *JenkinsConfigProvider) GetPullRequestConfig() PullRequestConfig {
	return PullRequestConfig{
		Branch: getEnv("CHANGE_BRANCH", "n/a"),
		Base:   getEnv("CHANGE_TARGET", "n/a"),
		Key:    getEnv("CHANGE_ID", "n/a"),
	}
}

func (j *JenkinsConfigProvider) IsPullRequest() bool {
	return truthy("CHANGE_ID")
}

func isJenkins() bool {
	envVars := []string{"JENKINS_HOME", "JENKINS_URL"}
	return areIndicatingEnvVarsSet(envVars)
}
