package jenkins

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bndr/gojenkins"
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
		return nil, fmt.Errorf("failed to load job: %w", pollJobErr)
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
	return getBuildFromQueueID(ctx, jenkins, job.GetJob(), queueID)
}

// getBuildFromQueueID replaces the broken gojenkins.Jenkins.GetBuildFromQueueID.
// gojenkins.Job.GetBuild strips the server host from job.Raw.URL using
// strings.Replace, which silently no-ops when the Jenkins API returns a different
// host than the one used at connect time (e.g. after a DNS migration), producing a
// malformed concatenated URL. This implementation constructs the Build directly from
// the executable URL's path, bypassing that broken logic.
func getBuildFromQueueID(ctx context.Context, j Jenkins, job *gojenkins.Job, queueID int64) (*gojenkins.Build, error) {
	rawJenkins, ok := j.(*gojenkins.Jenkins)
	if !ok {
		// Non-gojenkins implementation (e.g. test mock): fall back to the standard path.
		return j.GetBuildFromQueueID(ctx, job, queueID)
	}

	task, err := rawJenkins.GetQueueItem(ctx, queueID)
	if err != nil {
		return nil, err
	}
	for task.Raw.Executable.Number == 0 {
		time.Sleep(1000 * time.Millisecond)
		if _, err = task.Poll(ctx); err != nil {
			return nil, err
		}
	}

	parsedURL, err := url.Parse(task.Raw.Executable.URL)
	if err != nil || parsedURL.Path == "" {
		return nil, fmt.Errorf("unexpected build URL from queue item '%s'", task.Raw.Executable.URL)
	}

	build := &gojenkins.Build{
		Jenkins: rawJenkins,
		Raw:     new(gojenkins.BuildResponse),
		Depth:   1,
		Base:    parsedURL.Path,
	}
	status, err := build.Poll(ctx)
	if err != nil {
		return nil, err
	}
	if status != 200 {
		return nil, fmt.Errorf("unexpected HTTP status %d fetching build at path '%s'", status, parsedURL.Path)
	}
	return build, nil
}
