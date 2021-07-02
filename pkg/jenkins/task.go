package jenkins

import (
	"context"
	"fmt"
	"time"

	"github.com/bndr/gojenkins"
)

// Task is an interface to abstract gojenkins.Task.
// mock generated with: mockery --name Task --dir pkg/jenkins --output pkg/jenkins/mocks
type Task interface {
	Poll(context.Context) (int, error)
	BuildNumber() (int64, error)
	HasStarted() bool
	WaitToStart(ctx context.Context, pollInterval time.Duration) (int64, error)
}

// TaskImpl is a wrapper struct for gojenkins.Task that respects the Task interface.
type TaskImpl struct {
	Task *gojenkins.Task
}

// Poll refers to the gojenkins.Task.Poll function.
func (t *TaskImpl) Poll(ctx context.Context) (int, error) {
	return t.Task.Poll(ctx)
}

// HasStarted checks if the wrapped gojenkins.Task has started by checking the assigned executable URL.
func (t *TaskImpl) HasStarted() bool {
	return t.Task.Raw.Executable.URL != ""
}

// BuildNumber returns the assigned build number or an error if the build has not yet started.
func (t *TaskImpl) BuildNumber() (int64, error) {
	if !t.HasStarted() {
		return 0, fmt.Errorf("build did not start yet")
	}
	return t.Task.Raw.Executable.Number, nil
}

// WaitToStart waits till the build has started.
func (t *TaskImpl) WaitToStart(ctx context.Context, pollInterval time.Duration) (int64, error) {
	for retry := 0; retry < 15; {
		if t.HasStarted() {
			return t.BuildNumber()
		}
		time.Sleep(pollInterval)
		t.Poll(ctx)
		retry++
	}
	return 0, fmt.Errorf("build did not start in a reasonable amount of time")
}
