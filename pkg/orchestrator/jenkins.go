package orchestrator

import (
	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
	"io/ioutil"
)

type JenkinsConfigProvider struct{}

func (a *JenkinsConfigProvider) OrchestratorVersion() string {
	return getEnv("JENKINS_VERSION", "n/a")
}

func (a *JenkinsConfigProvider) OrchestratorType() string {
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

	client := &piperHttp.Client{}
	options := piperHttp.ClientOptions{
		Username: getEnv("PIPER_jenkinsUser", "N/A"),
		Password: getEnv("PIPER_jenkinsToken", "N/A"),
	}

	client.SetOptions(options)
	response, err := client.GetRequest(URL, nil, nil)
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

func (a *JenkinsConfigProvider) GetPipelineStartTime() string {
	log.Entry().Infof("GetPipelineStartTime() for Jenkins not yet implemented.")
	return "n/a"
}

func (j *JenkinsConfigProvider) GetJobName() string {
	return getEnv("JOB_NAME", "n/a")
}

func (j *JenkinsConfigProvider) getJenkinsHome() string {
	return getEnv("JENKINS_HOME", "n/a")
}

func (j *JenkinsConfigProvider) GetBuildNumber() string {
	return getEnv("BUILD_NUMBER", "n/a")
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
