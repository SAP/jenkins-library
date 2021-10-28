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

func (j *JenkinsConfigProvider) GetLog() ([]byte, error) {
	// Questions:
	// How to get the data from jenkins?
	// (a) Getting it from local file systems, difficulties with mounted volumes on Google Cloud
	// (b) Getting data via API ->
	//	* Problem of authentication, do we have credentials available in vault?
	//	* ...
	// How to get step specific data? As it is shown in Blue Ocean?

	URL := j.GetBuildUrl() + "consoleText"

	response, err := j.client.GetRequest(URL, nil, nil)
	if err != nil {
		return []byte{}, errors.Wrapf(err, "Could not read Jenkins logfile. %v", err)
	}
	if response.StatusCode != 200 {
		log.Entry().Errorf("Could not get log information from Jenkins. Returning with empty log.")
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

	myvar := responseInterface["timestamp"].(float64)
	timestamp := time.Unix(int64(myvar)/1000, 0)

	log.Entry().Debugf("Pipeline start time: %v", timestamp.String())
	defer response.Body.Close()
	return timestamp
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

func (j *JenkinsConfigProvider) GetBuildNumber() string {
	return getEnv("BUILD_NUMBER", "n/a")
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
