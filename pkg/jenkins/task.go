package jenkins

import (
	"fmt"
	"time"

	"github.com/bndr/gojenkins"
)

// Task is an interface to abstract gojenkins.Task.
type Task interface {
	Poll() (int, error)
	BuildNumber() (int64, error)
	HasStarted() bool
	WaitToStart(pollInterval time.Duration) (int64, error)
}

// TaskImpl is a wrapper struct for gojenkins.Task that respects the Task interface.
type TaskImpl struct {
	Task *gojenkins.Task
}

// Poll refers to the gojenkins.Task.Poll function.
func (t *TaskImpl) Poll() (int, error) {
	return t.Task.Poll()
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
func (t *TaskImpl) WaitToStart(pollInterval time.Duration) (int64, error) {
	for retry := 0; retry < 15; {
		if t.HasStarted() {
			return t.BuildNumber()
		}
		time.Sleep(pollInterval)
		t.Poll()
		retry++
	}
	return 0, fmt.Errorf("build did not start in a reasonable amount of time")
}
