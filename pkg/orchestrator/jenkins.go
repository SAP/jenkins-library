package orchestrator

import (
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
)

type JenkinsConfigProvider struct{}

func (j *JenkinsConfigProvider) GetLog() ([]byte, error) {
	filePath := j.getJenkinsHome() + "/jobs/" + j.GetJobName() + "/builds/" + j.GetBuildNumber() + "/log"
	log.Entry().Debugf("Reading Jenkins-Logfile from: %v", filePath)
	logFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not read Jenkins-Logfile from %s", filePath)
	}
	//logFileContent := new File("${JENKINS_HOME}/jobs/${JOB_NAME}/builds/${BUILD_NUMBER}/log").collect {it}
	return logFile, nil
}

func (j *JenkinsConfigProvider) GetJobName() string {
	return os.Getenv("JOB_NAME")
}

func (j *JenkinsConfigProvider) getJenkinsHome() string {
	return os.Getenv("JOB_NAME")
}

func (j *JenkinsConfigProvider) GetBuildNumber() string {
	return os.Getenv("BUILD_NUMBER")
}

func (j *JenkinsConfigProvider) GetBranch() string {
	return os.Getenv("GIT_BRANCH")
}

func (j *JenkinsConfigProvider) GetBuildUrl() string {
	return os.Getenv("BUILD_URL")
}

func (j *JenkinsConfigProvider) GetCommit() string {
	return os.Getenv("GIT_COMMIT")
}

func (j *JenkinsConfigProvider) GetRepoUrl() string {
	return os.Getenv("GIT_URL")
}

func (j *JenkinsConfigProvider) GetPullRequestConfig() PullRequestConfig {
	return PullRequestConfig{
		Branch: os.Getenv("CHANGE_BRANCH"),
		Base:   os.Getenv("CHANGE_TARGET"),
		Key:    os.Getenv("CHANGE_ID"),
	}
}

func (j *JenkinsConfigProvider) IsPullRequest() bool {
	return truthy("CHANGE_ID")
}

func isJenkins() bool {
	envVars := []string{"JENKINS_HOME", "JENKINS_URL"}
	return areIndicatingEnvVarsSet(envVars)
}
