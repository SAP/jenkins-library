package jenkins

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/bndr/gojenkins"
	"github.com/pkg/errors"
)

// Jenkins is an interface to abstract gojenkins.Jenkins.
type Jenkins interface {
	GetJobObj(ctx context.Context, name string) *gojenkins.Job
	BuildJob(ctx context.Context, name string, params map[string]string) (int64, error)
	GetBuildFromQueueID(ctx context.Context, job *gojenkins.Job, queueid int64) (*gojenkins.Build, error)
}

// Instance connects to a Jenkins instance and returns a handler.
func Instance(ctx context.Context, client *http.Client, jenkinsURL, user, token string) (*gojenkins.Jenkins, error) {
	return gojenkins.
		CreateJenkins(client, jenkinsURL, user, token).
		Init(ctx)
}

func GetJob(ctx context.Context, jenkins Jenkins, jobName string) (Job, error) {
	// get job id
	jobID := strings.ReplaceAll(jobName, "/", "/job/")
	// get job
	return &JobImpl{Job: jenkins.GetJobObj(ctx, jobID)}, nil
}

// TriggerJob starts a build for a given job name.
func TriggerJob(ctx context.Context, jenkins Jenkins, job Job, parameters map[string]string) (*gojenkins.Build, error) {
	// update job
	_, pollJobErr := job.Poll(ctx)
	if pollJobErr != nil {
		return nil, errors.Wrapf(pollJobErr, "failed to load job")
	}
	// start job
	queueID, startBuildErr := job.InvokeSimple(ctx, parameters)
	if startBuildErr != nil {
		return nil, startBuildErr
	}
	if queueID == 0 {
		// handle rare error case where queueID is not set
		// see https://github.com/bndr/gojenkins/issues/205
		// see https://github.com/bndr/gojenkins/pull/226
		return nil, fmt.Errorf("unable to queue build")
	}

	// get build
	return jenkins.GetBuildFromQueueID(ctx, job.GetJob(), queueID)
}
