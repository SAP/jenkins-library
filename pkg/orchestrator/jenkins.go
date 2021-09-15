package orchestrator

import (
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

	filePath := j.getJenkinsHome() + "/jobs/" + j.GetJobName() + "/builds/" + j.GetBuildNumber() + "/log"

	logFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not read Jenkins-Logfile from %s", filePath)
	}
	log.Entry().Debugf("Successful read Jenkins-Logfile from: %v", filePath)
	//logFileContent := new File("${JENKINS_HOME}/jobs/${JOB_NAME}/builds/${BUILD_NUMBER}/log").collect {it}
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
