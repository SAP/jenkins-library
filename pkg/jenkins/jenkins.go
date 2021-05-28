package jenkins

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/bndr/gojenkins"
)

// Jenkins is an interface to abstract gojenkins.Jenkins.
// mock generated with: mockery --name Jenkins --dir pkg/jenkins --output pkg/jenkins/mocks
type Jenkins interface {
	BuildJob(ctx context.Context, name string, options ...interface{}) (int64, error)
	GetBuildFromQueueID(ctx context.Context, queueid int64) (*gojenkins.Build, error)
}

// Instance connects to a Jenkins instance and returns a handler.
func Instance(ctx context.Context, client *http.Client, jenkinsURL, user, token string) (*gojenkins.Jenkins, error) {
	return gojenkins.
		CreateJenkins(client, jenkinsURL, user, token).
		Init(ctx)
}

// TriggerJob starts a build for a given job name.
func TriggerJob(ctx context.Context, jenkins Jenkins, jobName string, parameters map[string]string) (*gojenkins.Build, error) {
	// get job id
	jobID := strings.ReplaceAll(jobName, "/", "/job/")
	// start job
	queueID, startBuildErr := jenkins.BuildJob(ctx, jobID, parameters)
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
	return jenkins.GetBuildFromQueueID(ctx, queueID)
}
