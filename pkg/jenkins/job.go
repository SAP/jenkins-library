package jenkins

import (
	"context"

	"github.com/bndr/gojenkins"
)

// Task is an interface to abstract gojenkins.Task.
type Job interface {
	Poll(context.Context) (int, error)
	InvokeSimple(ctx context.Context, params map[string]string) (int64, error)
	GetJob() *gojenkins.Job
}

// JobImpl is a wrapper struct for gojenkins.Task that respects the Task interface.
type JobImpl struct {
	Job *gojenkins.Job
}

// Poll refers to the gojenkins.Job.Poll function.
func (t *JobImpl) Poll(ctx context.Context) (int, error) {
	return t.Job.Poll(ctx)
}

// InvokeSimple refers to the gojenkins.Job.InvokeSimple function.
func (t *JobImpl) InvokeSimple(ctx context.Context, params map[string]string) (int64, error) {
	return t.Job.InvokeSimple(ctx, params)
}

// GetJob returns wrapped gojenkins.Job.
func (t *JobImpl) GetJob() *gojenkins.Job {
	return t.Job
}
