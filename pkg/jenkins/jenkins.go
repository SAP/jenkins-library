package jenkins

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bndr/gojenkins"
)

// Jenkins is an interface to abstract gojenkins.Jenkins.
type Jenkins interface {
	BuildJob(name string, options ...interface{}) (int64, error)
	GetQueueItem(id int64) (*gojenkins.Task, error)
	GetBuild(jobName string, number int64) (*gojenkins.Build, error)
}

// Instance connects to a Jenkins instance and returns a handler.
func Instance(client *http.Client, jenkinsURL, user, token string) (*gojenkins.Jenkins, error) {
	return gojenkins.
		CreateJenkins(client, jenkinsURL, user, token).
		Init()
}

// TriggerJob starts a build for a given job name.
func TriggerJob(jenkins Jenkins, jobName string, parameters map[string]string) (*gojenkins.Task, error) {
	// get job id
	jobID := strings.ReplaceAll(jobName, "/", "/job/")
	// start job
	queueID, startBuildErr := jenkins.BuildJob(jobID, parameters)
	if startBuildErr != nil {
		return nil, startBuildErr
	}
	if queueID == 0 {
		// handle rare error case where queueID is not set
		// see https://github.com/bndr/gojenkins/issues/205
		return nil, fmt.Errorf("Unable to queue build")
	}
	// get task
	return jenkins.GetQueueItem(queueID)
}

// WaitForBuildToStart waits till a build is started.
func WaitForBuildToStart(jenkins Jenkins, jobName string, taskWrapper Task, pollInterval time.Duration) (*gojenkins.Build, error) {
	// wait for job to start
	buildNumber, taskTimedOutErr := taskWrapper.WaitToStart(pollInterval)
	if taskTimedOutErr != nil {
		return nil, taskTimedOutErr
	}
	// get job id
	jobID := strings.ReplaceAll(jobName, "/", "/job/")
	// get build
	return jenkins.GetBuild(jobID, buildNumber)
}
